import en from '$lib/i18n/en.json';
import ru from '$lib/i18n/ru.json';
import {
	cookiesEn,
	cookiesRu,
	privacyEn,
	privacyRu,
	termsEn,
	termsRu
} from '$lib/i18n/legal';
import { isLang, persistLang } from '$lib/i18n/routing';

export type Lang = 'en' | 'ru';

type Messages = typeof en & {
	privacy: typeof privacyEn;
	terms: typeof termsEn;
	cookies: typeof cookiesEn;
};

const translations: Record<Lang, Messages> = {
	en: { ...en, privacy: privacyEn, terms: termsEn, cookies: cookiesEn },
	ru: { ...ru, privacy: privacyRu, terms: termsRu, cookies: cookiesRu } as unknown as Messages
};

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
		persistLang(l);
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
		set,
		t,
		get
	};
}

export const langStore = createLangStore();
