import { browser } from '$app/environment';
import type { Lang } from '$lib/stores/lang.svelte';

export const SUPPORTED_LANGS = ['en', 'ru'] as const satisfies readonly Lang[];
export const DEFAULT_LANG: Lang = 'en';
const LANG_STORAGE_KEY = 'piplos-lang';

export function isLang(value: unknown): value is Lang {
	return value === 'en' || value === 'ru';
}

/** Язык, явно выбранный пользователем ранее (localStorage), либо null. */
export function getSavedLang(): Lang | null {
	if (!browser) return null;
	const saved = localStorage.getItem(LANG_STORAGE_KEY);
	return isLang(saved) ? saved : null;
}

/** Язык браузера: ru → ru, иначе язык по умолчанию. */
export function detectBrowserLang(): Lang {
	if (!browser) return DEFAULT_LANG;
	return navigator.language.toLowerCase().startsWith('ru') ? 'ru' : DEFAULT_LANG;
}

/** Сохранённый выбор пользователя, иначе язык браузера (для входа на сайт). */
export function resolveInitialLang(): Lang {
	return getSavedLang() ?? detectBrowserLang();
}

/** Сохраняет ЯВНЫЙ выбор пользователя. Не вызывать при синхронизации с URL. */
export function persistLang(lang: Lang) {
	if (!browser) return;
	localStorage.setItem(LANG_STORAGE_KEY, lang);
}

export function applyDocumentLang(lang: Lang) {
	if (!browser) return;
	document.documentElement.setAttribute('lang', lang);
}

/** Убирает префикс `/en` или `/ru` из pathname. */
export function delocalizePath(pathname: string): string {
	const segments = pathname.split('/').filter(Boolean);
	if (segments.length > 0 && isLang(segments[0])) {
		const rest = segments.slice(1).join('/');
		return rest ? `/${rest}` : '/';
	}
	return pathname || '/';
}

/** Добавляет префикс языка: `/portfolio` → `/ru/portfolio`, `/#stack` → `/ru#stack`. */
export function localizePath(path: string, lang: Lang): string {
	const hashIndex = path.indexOf('#');
	const hash = hashIndex >= 0 ? path.slice(hashIndex) : '';
	const pathOnly = hashIndex >= 0 ? path.slice(0, hashIndex) : path;
	const base = delocalizePath(pathOnly || '/');

	if (base === '/') return `/${lang}${hash}`;
	return `/${lang}${base}${hash}`;
}

/** Тот же путь с другим языком (сохраняет query и hash). */
export function switchLangHref(
	pathname: string,
	search: string,
	hash: string,
	nextLang: Lang
): string {
	const base = delocalizePath(pathname);
	return localizePath(base, nextLang) + search + hash;
}
