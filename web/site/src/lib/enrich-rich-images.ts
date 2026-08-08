/** Добавляет loading/decoding к <img> в rich HTML (solution, статьи, услуги). */
export function enrichRichImages(html: string): string {
	if (!html?.trim()) return html ?? '';

	return html.replace(/<img\b([^>]*)>/gi, (_full, rawAttrs: string) => {
		let attrs = rawAttrs;
		if (!/\bloading\s*=/i.test(attrs)) {
			attrs += ' loading="lazy"';
		}
		if (!/\bdecoding\s*=/i.test(attrs)) {
			attrs += ' decoding="async"';
		}
		return `<img${attrs}>`;
	});
}
