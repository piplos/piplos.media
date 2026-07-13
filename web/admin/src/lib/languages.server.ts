import { isRedirect, type RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import type { Language } from '$lib/types';

/** Загружает системные языки; при ошибке возвращает en/ru по умолчанию. */
export async function loadLanguages(event: RequestEvent): Promise<Language[]> {
	const fallback: Language[] = [
		{ code: 'en', name: 'English', is_default: true, enabled: true, sort_order: 0 },
		{ code: 'ru', name: 'Русский', is_default: false, enabled: true, sort_order: 1 }
	];
	try {
		const res = await fetchWithAuth(event, '/api/v1/languages');
		if (!res.ok) return fallback;
		const data = (await res.json()) as { languages: Language[] };
		return data.languages?.length ? data.languages : fallback;
	} catch (e) {
		if (isRedirect(e)) throw e;
		return fallback;
	}
}
