import DOMPurify from 'isomorphic-dompurify';

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

/** Плоский текст из HTML (превью, проверка заполненности). */
export function htmlToPlainText(html: string): string {
	return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
}
