import type { ProjectLink, ProjectLinkKind, Translations } from '$lib/types';

export const PROJECT_LINK_KIND_OPTIONS: { value: ProjectLinkKind; label: string }[] = [
	{ value: 'website', label: 'Сайт' },
	{ value: 'app_store', label: 'App Store' },
	{ value: 'google_play', label: 'Google Play' }
];

const KIND_LABEL: Record<ProjectLinkKind, string> = Object.fromEntries(
	PROJECT_LINK_KIND_OPTIONS.map((opt) => [opt.value, opt.label])
) as Record<ProjectLinkKind, string>;

/** Подписи store-типов — автоподставляются при смене URL. */
const AUTO_LABELS = new Set(
	PROJECT_LINK_KIND_OPTIONS.filter((opt) => opt.value !== 'website').map((opt) => opt.label)
);

const WEBSITE_PARAGRAPH_RE =
	/<p[^>]*>\s*<strong>(?:Сайт проекта|Project website):?<\/strong>\s*<a\s+[^>]*href="([^"]+)"[^>]*>([^<]*)<\/a>\s*<\/p>/gi;

const WEBSITE_MARKDOWN_RE =
	/\n?\s*\*\*(?:Сайт проекта|Project website):\*\*\s*\[([^\]]*)\]\(([^)]+)\)\s*/gi;

const STORE_LINK_RE =
	/<a\s+[^>]*href="(https?:\/\/(?:play\.google\.com|apps\.apple\.com)[^"]+)"[^>]*>([^<]*)<\/a>/gi;

const STORE_MARKDOWN_RE =
	/\[([^\]]*)\]\((https?:\/\/(?:play\.google\.com|apps\.apple\.com)[^)]+)\)/gi;

export function isProjectLinkKind(value: string): value is ProjectLinkKind {
	return value === 'website' || value === 'google_play' || value === 'app_store';
}

export function kindFromUrl(url: string): ProjectLinkKind {
	if (url.includes('play.google.com')) return 'google_play';
	if (url.includes('apps.apple.com')) return 'app_store';
	return 'website';
}

export function kindLabel(kind: ProjectLinkKind): string {
	return KIND_LABEL[kind];
}

function displayLabel(url: string, kind: ProjectLinkKind, anchorText: string): string {
	if (kind === 'google_play' || kind === 'app_store') return kindLabel(kind);
	const text = anchorText.trim();
	if (text) return text;
	try {
		return new URL(url).hostname.replace(/^www\./, '');
	} catch {
		return url;
	}
}

/** Пустая / store / равная URL — можно заменить при автоопределении kind. */
export function isAutoManagedLabel(link: ProjectLink): boolean {
	const label = link.label.trim();
	return !label || label === link.url || AUTO_LABELS.has(label);
}

/** Обновляет URL и kind; подпись — только если она «автоматическая». */
export function withUpdatedUrl(link: ProjectLink, url: string): ProjectLink {
	const kind = kindFromUrl(url);
	return {
		...link,
		url,
		kind,
		label: isAutoManagedLabel(link) ? displayLabel(url, kind, '') : link.label
	};
}

function addLink(links: ProjectLink[], seen: Set<string>, url: string, anchorText: string) {
	const normalized = url.trim();
	if (!normalized || seen.has(normalized)) return;
	seen.add(normalized);
	const kind = kindFromUrl(normalized);
	links.push({
		url: normalized,
		label: displayLabel(normalized, kind, anchorText),
		kind
	});
}

type CaptureOrder = 'urlText' | 'textUrl';

function collectMatches(
	html: string,
	re: RegExp,
	order: CaptureOrder,
	links: ProjectLink[],
	seen: Set<string>
) {
	html.replace(re, (_match, a: string, b: string) => {
		const url = order === 'urlText' ? a : b;
		const text = order === 'urlText' ? b : a;
		addLink(links, seen, url, text);
		return '';
	});
}

/** Извлекает ссылки из HTML/markdown solution (как на публичном сайте). */
export function extractLinksFromSolution(html: string): ProjectLink[] {
	if (!html?.trim()) return [];

	const links: ProjectLink[] = [];
	const seen = new Set<string>();

	collectMatches(html, WEBSITE_PARAGRAPH_RE, 'urlText', links, seen);
	collectMatches(html, WEBSITE_MARKDOWN_RE, 'textUrl', links, seen);
	collectMatches(html, STORE_LINK_RE, 'urlText', links, seen);
	collectMatches(html, STORE_MARKDOWN_RE, 'textUrl', links, seen);

	return links;
}

/** Собирает ссылки из переводов solution, если в проекте ещё нет структурированных. */
export function seedLinksFromTranslations(
	links: ProjectLink[] | undefined,
	translations: Translations | undefined
): ProjectLink[] {
	if (links?.length) return links.map((link) => ({ ...link }));
	const seen = new Set<string>();
	const out: ProjectLink[] = [];
	for (const locale of Object.values(translations ?? {})) {
		for (const link of extractLinksFromSolution(locale?.solution ?? '')) {
			if (seen.has(link.url)) continue;
			seen.add(link.url);
			out.push(link);
		}
	}
	return out;
}

export function normalizeProjectLinks(links: ProjectLink[]): ProjectLink[] {
	const seen = new Set<string>();
	const out: ProjectLink[] = [];
	for (const link of links) {
		const url = link.url.trim();
		if (!url || seen.has(url)) continue;
		seen.add(url);
		const kind = isProjectLinkKind(link.kind) ? link.kind : kindFromUrl(url);
		const label = link.label.trim() || displayLabel(url, kind, '');
		out.push({ url, label, kind });
	}
	return out;
}

/** Начальные ссылки формы: DB или однократный seed из solution. */
export function initialProjectLinks(
	links: ProjectLink[] | undefined,
	translations: Translations | undefined
): ProjectLink[] {
	return normalizeProjectLinks(seedLinksFromTranslations(links, translations));
}

export function emptyProjectLink(): ProjectLink {
	return { url: '', label: '', kind: 'website' };
}
