import { page } from '$app/state';
import { isLang, localizePath } from '$lib/i18n/routing';

/** Локализованный href с учётом текущего `page.params.lang`. */
export function l(path: string): string {
	const lang = page.params.lang;
	return isLang(lang) ? localizePath(path, lang) : path;
}
