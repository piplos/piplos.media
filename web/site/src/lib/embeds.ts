/** Переменные {{projects …}} / {{services …}} в HTML статей и услуг.
 *  Токен вставляется в админке, API рендерит Markdown → HTML как обычный текст,
 *  сайт разворачивает токены в блоки с выборкой, количеством и дизайном. */

import type { PortfolioProject } from '$lib/portfolio';
import type { ServiceItem } from '$lib/services-api';

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

/** Выборка проектов по параметрам токена. Вход — опубликованные проекты
 *  в сквозном порядке портфолио (loadPortfolioProjects). */
export function selectProjects(all: PortfolioProject[], p: EmbedParams): PortfolioProject[] {
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

/** Выборка услуг по параметрам токена. Вход — опубликованные услуги (fetchServices). */
export function selectServices(all: ServiceItem[], p: EmbedParams): ServiceItem[] {
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
