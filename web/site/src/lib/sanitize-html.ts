import { browser } from '$app/environment';

const ALLOWED_TAGS = [
	'p',
	'br',
	'strong',
	'em',
	'b',
	'i',
	'del',
	'ul',
	'ol',
	'li',
	'a',
	'h2',
	'h3',
	'h4',
	'img',
	'blockquote',
	'code',
	'pre',
	'hr',
	'table',
	'thead',
	'tbody',
	'tr',
	'th',
	'td'
];

const ALLOWED_ATTR = ['href', 'src', 'alt', 'title', 'target', 'rel', 'class'];

const PURIFY_CONFIG = {
	ALLOWED_TAGS,
	ALLOWED_ATTR,
	ALLOW_DATA_ATTR: false as const
};

type Purify = {
	sanitize: (html: string, config: typeof PURIFY_CONFIG) => string;
};

let purify: Purify | null = null;
let purifyPromise: Promise<Purify> | null = null;

function loadPurify(): Promise<Purify> {
	if (!browser) {
		return Promise.resolve({ sanitize: (html) => html.trim() });
	}
	if (purify) return Promise.resolve(purify);
	if (!purifyPromise) {
		purifyPromise = import('isomorphic-dompurify').then((mod) => {
			purify = mod.default;
			return purify;
		});
	}
	return purifyPromise;
}

/** SSR-safe: pass-through on server; sanitizes in browser once DOMPurify is loaded. */
export function sanitizeCaseHtml(html: string): string {
	const trimmed = html.trim();
	if (!trimmed) return '';
	if (!browser) return trimmed;
	if (purify) return purify.sanitize(trimmed, PURIFY_CONFIG);
	void loadPurify();
	return trimmed;
}

/** Always sanitizes in browser; pass-through on server. */
export async function sanitizeCaseHtmlAsync(html: string): Promise<string> {
	const trimmed = html.trim();
	if (!trimmed) return '';
	if (!browser) return trimmed;
	const p = await loadPurify();
	return p.sanitize(trimmed, PURIFY_CONFIG);
}

/** Плоский текст из HTML (превью, проверка заполненности). */
export function htmlToPlainText(html: string): string {
	return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
}
