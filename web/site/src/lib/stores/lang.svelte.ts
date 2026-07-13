import { browser } from '$app/environment';
import en from '$lib/i18n/en.json';
import ru from '$lib/i18n/ru.json';

export type Lang = 'en' | 'ru';

const translations: Record<Lang, typeof en> = { en, ru: ru as typeof en };

function resolve(lang: Lang, key: string): unknown {
	const keys = key.split('.');
	let val: unknown = translations[lang];
	for (const k of keys) {
		if (val == null || typeof val !== 'object') return undefined;
		val = (val as Record<string, unknown>)[k];
	}
	return val;
}

function createLangStore() {
	let lang = $state<Lang>('en');

	function applyLang(l: Lang) {
		lang = l;
		if (browser) {
			localStorage.setItem('piplos-lang', l);
			document.documentElement.setAttribute('lang', l);
		}
	}

	function init() {
		if (!browser) return;
		const saved = localStorage.getItem('piplos-lang') as Lang | null;
		const browserLang = navigator.language.startsWith('ru') ? 'ru' : 'en';
		applyLang(saved ?? browserLang);
	}

	function set(l: Lang) {
		applyLang(l);
	}

	function toggle() {
		set(lang === 'en' ? 'ru' : 'en');
	}

	function t(key: string): string {
		const val = resolve(lang, key);
		return typeof val === 'string' ? val : key;
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
