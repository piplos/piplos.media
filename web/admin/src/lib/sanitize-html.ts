import DOMPurify from 'isomorphic-dompurify';

const ALLOWED_TAGS = [
	'p',
	'br',
	'strong',
	'em',
	'b',
	'i',
	'ul',
	'ol',
	'li',
	'a',
	'h2',
	'h3',
	'img',
	'blockquote'
];

const ALLOWED_ATTR = ['href', 'src', 'alt', 'title', 'target', 'rel'];

/** Безопасный HTML для контента кейсов (solution и т.п.). */
export function sanitizeCaseHtml(html: string): string {
	const trimmed = html.trim();
	if (!trimmed) return '';

	return DOMPurify.sanitize(trimmed, {
		ALLOWED_TAGS,
		ALLOWED_ATTR,
		ALLOW_DATA_ATTR: false
	});
}
