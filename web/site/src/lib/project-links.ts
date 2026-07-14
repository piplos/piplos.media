export type ProjectLinkKind = 'website' | 'google_play' | 'app_store';

export type ProjectLink = {
	url: string;
	label: string;
	kind: ProjectLinkKind;
};

const WEBSITE_PARAGRAPH_RE =
	/<p[^>]*>\s*<strong>(?:Сайт проекта|Project website):?<\/strong>\s*<a\s+[^>]*href="([^"]+)"[^>]*>([^<]*)<\/a>\s*<\/p>/gi;

const WEBSITE_MARKDOWN_RE =
	/\n?\s*\*\*(?:Сайт проекта|Project website):\*\*\s*\[([^\]]*)\]\(([^)]+)\)\s*/gi;

const STORE_LINK_RE =
	/<a\s+[^>]*href="(https?:\/\/(?:play\.google\.com|apps\.apple\.com)[^"]+)"[^>]*>([^<]*)<\/a>/gi;

const STORE_MARKDOWN_RE =
	/\[([^\]]*)\]\((https?:\/\/(?:play\.google\.com|apps\.apple\.com)[^)]+)\)/gi;

function kindFromUrl(url: string): ProjectLinkKind {
	if (url.includes('play.google.com')) return 'google_play';
	if (url.includes('apps.apple.com')) return 'app_store';
	return 'website';
}

function displayLabel(url: string, kind: ProjectLinkKind, anchorText: string): string {
	const text = anchorText.trim();
	if (kind === 'google_play') return 'Google Play';
	if (kind === 'app_store') return 'App Store';
	if (text) return text;
	try {
		return new URL(url).hostname.replace(/^www\./, '');
	} catch {
		return url;
	}
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

function sortLinks(links: ProjectLink[]): ProjectLink[] {
	const order: Record<ProjectLinkKind, number> = { website: 0, app_store: 1, google_play: 2 };
	return [...links].sort((a, b) => order[a.kind] - order[b.kind] || a.label.localeCompare(b.label));
}

/** Извлекает ссылки на сайт проекта и сторы; убирает их из HTML/markdown solution. */
export function extractAndStripProjectLinks(html: string): { links: ProjectLink[]; html: string } {
	if (!html?.trim()) return { links: [], html: html ?? '' };

	const links: ProjectLink[] = [];
	const seen = new Set<string>();
	let out = html;

	out = out.replace(WEBSITE_PARAGRAPH_RE, (_, url: string, text: string) => {
		addLink(links, seen, url, text);
		return '';
	});

	out = out.replace(WEBSITE_MARKDOWN_RE, (_, text: string, url: string) => {
		addLink(links, seen, url, text);
		return '';
	});

	out = out.replace(STORE_LINK_RE, (_, url: string, text: string) => {
		addLink(links, seen, url, text);
		return text.trim();
	});

	out = out.replace(STORE_MARKDOWN_RE, (_, text: string, url: string) => {
		addLink(links, seen, url, text);
		return text.trim();
	});

	out = out.replace(/\n{3,}/g, '\n\n').trim();

	return { links: sortLinks(links), html: out };
}
