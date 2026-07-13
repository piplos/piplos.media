import { browser } from '$app/environment';
import en from '$lib/i18n/en.json';
import ru from '$lib/i18n/ru.json';

export type Lang = 'en' | 'ru';

const STORAGE_KEY = 'piplos-lang';

const translations: Record<Lang, typeof en> = { en, ru: ru as typeof en };

function isLang(value: unknown): value is Lang {
	return value === 'en' || value === 'ru';
}

function resolve(lang: Lang, key: string): unknown {
	let val: unknown = translations[lang];
	for (const k of key.split('.')) {
		if (val == null || typeof val !== 'object') return undefined;
		val = (val as Record<string, unknown>)[k];
	}
	return val;
}

function createLangStore() {
	let lang = $state<Lang>('en');

	function set(l: Lang) {
		lang = l;
		if (browser) {
			localStorage.setItem(STORAGE_KEY, l);
			document.documentElement.setAttribute('lang', l);
		}
	}

	function init() {
		if (!browser) return;
		const saved = localStorage.getItem(STORAGE_KEY);
		const browserLang = navigator.language.startsWith('ru') ? 'ru' : 'en';
		set(isLang(saved) ? saved : browserLang);
	}

	function toggle() {
		set(lang === 'en' ? 'ru' : 'en');
	}

	function t(key: string, params?: Record<string, string | number>): string {
		const val = resolve(lang, key);
		if (typeof val !== 'string') return key;
		if (!params) return val;
		return val.replace(/\{(\w+)\}/g, (match, name) =>
			name in params ? String(params[name]) : match
		);
	}

	function get<T>(key: string): T | undefined {
		return resolve(lang, key) as T | undefined;
	}

	return {
		get value() { return lang; },
		init,
		set,
		toggle,
		t,
		get
	};
}

export const langStore = createLangStore();
