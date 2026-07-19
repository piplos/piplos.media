/** Переменные (embed-токены) в Markdown: {{projects …}} / {{services …}}.
 *  Токен сохраняется в тексте как есть; сайт разворачивает его в блок
 *  проектов/услуг с выбранной выборкой, количеством и дизайном. */

export type EmbedKind = 'projects' | 'services';
export type EmbedLayout = 'cards' | 'list' | 'compact';
export type EmbedSelection = 'auto' | 'featured' | 'category' | 'tags' | 'slugs';

export interface EmbedDraft {
	kind: EmbedKind;
	selection: EmbedSelection;
	/** Значение выборки: slug группы, теги или slug-и через запятую. */
	value: string;
	limit: number;
	layout: EmbedLayout;
}

export const EMBED_KIND_OPTIONS = [
	{ value: 'projects', label: 'Проекты' },
	{ value: 'services', label: 'Услуги' }
] as const;

export const EMBED_LAYOUT_OPTIONS = [
	{ value: 'cards', label: 'Карточки' },
	{ value: 'list', label: 'Список' },
	{ value: 'compact', label: 'Компакт (ссылки)' }
] as const;

export const EMBED_SELECTION_OPTIONS: Record<
	EmbedKind,
	{ value: EmbedSelection; label: string }[]
> = {
	projects: [
		{ value: 'auto', label: 'Автоматически (порядок портфолио)' },
		{ value: 'featured', label: 'Только избранные (featured)' },
		{ value: 'category', label: 'Из группы услуги (slug услуги)' },
		{ value: 'tags', label: 'По тегам стека' },
		{ value: 'slugs', label: 'Конкретные проекты (slug-и)' }
	],
	services: [
		{ value: 'auto', label: 'Автоматически (порядок услуг)' },
		{ value: 'tags', label: 'По тегам стека' },
		{ value: 'slugs', label: 'Конкретные услуги (slug-и)' }
	]
};

function quoteIfNeeded(value: string): string {
	return /\s/.test(value) ? `"${value}"` : value;
}

/** Собирает токен вида {{projects limit=3 layout=cards featured}}. */
export function buildEmbedToken(draft: EmbedDraft): string {
	const parts: string[] = [draft.kind];
	const value = draft.value.trim().replace(/"/g, '');
	switch (draft.selection) {
		case 'featured':
			parts.push('featured');
			break;
		case 'category':
			if (value) parts.push(`category=${quoteIfNeeded(value)}`);
			break;
		case 'tags':
			if (value) parts.push(`tags=${quoteIfNeeded(value)}`);
			break;
		case 'slugs':
			if (value) parts.push(`slugs=${quoteIfNeeded(value)}`);
			break;
	}
	const limit = Math.min(24, Math.max(1, Math.trunc(draft.limit) || 3));
	parts.push(`limit=${limit}`);
	if (draft.layout !== 'cards') parts.push(`layout=${draft.layout}`);
	return `{{${parts.join(' ')}}}`;
}

const TOKEN_RE = /\{\{\s*(projects|services)\b([^{}]*)\}\}/g;

const LAYOUT_LABELS: Record<string, string> = {
	cards: 'карточки',
	list: 'список',
	compact: 'компакт'
};

/** Короткое описание токена для превью: «Проекты · 3 · карточки · featured». */
export function describeEmbedToken(kind: string, rawParams: string): string {
	const decoded = rawParams.replaceAll('&quot;', '"').replaceAll('&#34;', '"');
	const params = new Map<string, string>();
	const flags = new Set<string>();
	for (const m of decoded.matchAll(/(\w+)(?:=("[^"]*"|\S+))?/g)) {
		const key = m[1].toLowerCase();
		if (m[2] === undefined) flags.add(key);
		else params.set(key, m[2].replace(/^"|"$/g, ''));
	}
	const bits: string[] = [kind === 'projects' ? 'Проекты' : 'Услуги'];
	bits.push(params.get('limit') ?? '3');
	bits.push(LAYOUT_LABELS[params.get('layout') ?? 'cards'] ?? 'карточки');
	if (flags.has('featured')) bits.push('избранные');
	if (params.has('category')) bits.push(`группа: ${params.get('category')}`);
	if (params.has('tags')) bits.push(`теги: ${params.get('tags')}`);
	if (params.has('slugs')) bits.push(`slug: ${params.get('slugs')}`);
	return bits.join(' · ');
}

/** Заменяет токены в готовом HTML превью на визуальные плашки. */
export function embedTokensToPreviewChips(html: string): string {
	return html.replace(TOKEN_RE, (_, kind: string, rawParams: string) => {
		const label = describeEmbedToken(kind, rawParams);
		return `<span class="mde-embed-chip" title="Блок подставится на сайте">▦ ${label}</span>`;
	});
}
