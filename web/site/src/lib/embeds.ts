/** Переменные {{projects …}} / {{services …}} в HTML статей и услуг.
 *  Токен вставляется в админке, API рендерит Markdown → HTML как обычный текст,
 *  сайт разворачивает токены в блоки с выборкой, количеством и дизайном. */

export type EmbedLayout = 'cards' | 'list' | 'compact';

export interface EmbedParams {
	kind: 'projects' | 'services';
	featured: boolean;
	category: string;
	tags: string[];
	slugs: string[];
	limit: number;
	layout: EmbedLayout;
}

export type BodySegment =
	| { type: 'html'; html: string }
	| { type: 'embed'; params: EmbedParams };

/** Минимальные поля проекта для клиентской выборки по токену. */
export interface EmbedProjectLike {
	id: string;
	featured: boolean;
	category: string;
	categories: string[];
	tags: string[];
	sort_order: number;
	year: number;
}

/** Минимальные поля услуги для клиентской выборки по токену. */
export interface EmbedServiceLike {
	slug: string;
	tags?: string[];
	sort_order: number;
}

/** Query shape для дозированного GET /public/projects (совместим с ProjectsQuery). */
export type EmbedProjectsQuery = {
	lang?: string;
	featured?: boolean;
	category?: string;
	tags?: string[];
	slugs?: string[];
	limit: number;
	mode: 'summary';
};

/** Query shape для дозированного GET /public/services (совместим с ServicesQuery). */
export type EmbedServicesQuery = {
	lang?: string;
	tags?: string[];
	slugs?: string[];
	limit: number;
	mode: 'summary';
};

/** Токен, при необходимости обёрнутый в свой <p> (goldmark ставит его вокруг строки). */
const TOKEN_RE = /(?:<p>\s*)?\{\{\s*(projects|services)\b([^{}]*)\}\}(?:\s*<\/p>)?/gi;

function decodeEntities(raw: string): string {
	return raw
		.replaceAll('&quot;', '"')
		.replaceAll('&#34;', '"')
		.replaceAll('&amp;', '&');
}

function splitList(value: string): string[] {
	return value
		.split(',')
		.map((s) => s.trim())
		.filter(Boolean);
}

export function parseEmbedParams(kind: 'projects' | 'services', raw: string): EmbedParams {
	const params = new Map<string, string>();
	const flags = new Set<string>();
	for (const m of decodeEntities(raw).matchAll(/(\w+)(?:=("[^"]*"|\S+))?/g)) {
		const key = m[1].toLowerCase();
		if (m[2] === undefined) flags.add(key);
		else params.set(key, m[2].replace(/^"|"$/g, ''));
	}
	const layoutRaw = params.get('layout') ?? 'cards';
	const layout: EmbedLayout =
		layoutRaw === 'list' || layoutRaw === 'compact' ? layoutRaw : 'cards';
	const limit = Math.min(24, Math.max(1, Math.trunc(Number(params.get('limit'))) || 3));
	return {
		kind,
		featured: flags.has('featured'),
		category: params.get('category') ?? '',
		tags: splitList(params.get('tags') ?? ''),
		slugs: splitList(params.get('slugs') ?? ''),
		limit,
		layout
	};
}

/** Делит готовый HTML на сегменты: обычный HTML и embed-блоки. */
export function splitBodySegments(html: string): BodySegment[] {
	const segments: BodySegment[] = [];
	let last = 0;
	for (const m of html.matchAll(TOKEN_RE)) {
		const idx = m.index ?? 0;
		const before = html.slice(last, idx).trim();
		if (before) segments.push({ type: 'html', html: before });
		segments.push({
			type: 'embed',
			params: parseEmbedParams(m[1].toLowerCase() as 'projects' | 'services', m[2] ?? '')
		});
		last = idx + m[0].length;
	}
	const rest = html.slice(last).trim();
	if (rest) segments.push({ type: 'html', html: rest });
	return segments;
}

/** Выборка проектов по параметрам токена. Вход — batch с API (уже дозированный). */
export function selectProjects<T extends EmbedProjectLike>(all: T[], p: EmbedParams): T[] {
	let items = all;
	if (p.slugs.length) {
		const byId = new Map(items.map((item) => [item.id, item]));
		items = p.slugs.flatMap((slug) => byId.get(slug) ?? []);
	} else {
		if (p.featured) items = items.filter((item) => item.featured);
		if (p.category) {
			items = items
				.filter(
					(item) => item.category === p.category || item.categories.includes(p.category)
				)
				.toSorted((a, b) => a.sort_order - b.sort_order || b.year - a.year);
		}
		if (p.tags.length) {
			const wanted = new Set(p.tags.map((t) => t.toLowerCase()));
			items = items.filter((item) => item.tags.some((tag) => wanted.has(tag.toLowerCase())));
		}
	}
	return items.slice(0, p.limit);
}

/** Выборка услуг по параметрам токена. */
export function selectServices<T extends EmbedServiceLike>(all: T[], p: EmbedParams): T[] {
	let items = [...all].sort((a, b) => a.sort_order - b.sort_order);
	if (p.slugs.length) {
		const bySlug = new Map(items.map((item) => [item.slug, item]));
		items = p.slugs.flatMap((slug) => bySlug.get(slug) ?? []);
	} else if (p.tags.length) {
		const wanted = new Set(p.tags.map((t) => t.toLowerCase()));
		items = items.filter((item) => (item.tags ?? []).some((tag) => wanted.has(tag.toLowerCase())));
	}
	return items.slice(0, p.limit);
}

/** Параметры одного `{{projects …}}` → query public API (дозированная выборка). */
export function embedParamsToProjectsQuery(
	p: EmbedParams,
	base: { lang?: string } = {}
): EmbedProjectsQuery {
	const query: EmbedProjectsQuery = { lang: base.lang, limit: p.limit, mode: 'summary' };
	if (p.slugs.length) {
		query.slugs = p.slugs;
		return query;
	}
	if (p.featured) query.featured = true;
	if (p.category) query.category = p.category;
	if (p.tags.length) query.tags = p.tags;
	return query;
}

/** Параметры одного `{{services …}}` → query public API. */
export function embedParamsToServicesQuery(
	p: EmbedParams,
	base: { lang?: string } = {}
): EmbedServicesQuery {
	const query: EmbedServicesQuery = { lang: base.lang, limit: p.limit, mode: 'summary' };
	if (p.slugs.length) {
		query.slugs = p.slugs;
		return query;
	}
	if (p.tags.length) query.tags = p.tags;
	return query;
}

/** Все `{{projects …}}` / `{{services …}}` сегменты из HTML. */
export function listEmbedParams(html: string, kind: 'projects' | 'services'): EmbedParams[] {
	if (!html?.trim()) return [];
	const out: EmbedParams[] = [];
	for (const segment of splitBodySegments(html)) {
		if (segment.type === 'embed' && segment.params.kind === kind) {
			out.push(segment.params);
		}
	}
	return out;
}

/** Ключ выборки токена (без layout — на состав карточек не влияет). */
export function embedSelectionKey(p: EmbedParams): string {
	return [
		p.kind,
		p.featured ? '1' : '0',
		p.category,
		p.tags.join(','),
		p.slugs.join(','),
		String(p.limit)
	].join('|');
}

/** Выборки по ключу токена: один API-batch на уникальные params (без UNION-pollution). */
export async function collectEmbedItems<T>(
	html: string,
	kind: 'projects' | 'services',
	loadBatch: (params: EmbedParams) => Promise<T[]>,
	select: (batch: T[], params: EmbedParams) => T[]
): Promise<Record<string, T[]>> {
	const embeds = listEmbedParams(html, kind);
	if (embeds.length === 0) return {};

	// Одинаковые токены (в т.ч. дубли en/ru) — один запрос.
	const unique = new Map<string, EmbedParams>();
	for (const params of embeds) {
		unique.set(embedSelectionKey(params), params);
	}

	const entries = [...unique.entries()];
	const batches = await Promise.all(entries.map(([, params]) => loadBatch(params)));
	const byKey: Record<string, T[]> = {};
	for (let i = 0; i < entries.length; i++) {
		const [key, params] = entries[i];
		byKey[key] = select(batches[i] ?? [], params);
	}
	return byKey;
}
