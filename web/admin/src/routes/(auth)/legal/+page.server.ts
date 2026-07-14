import { error, fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { legalPayload } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import type { LegalPage } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/legal');
		if (!res.ok) {
			return {
				pages: [] as LegalPage[],
				languages: await loadLanguages(event),
				error: apiLoadErrorMessage(res, 'Ошибка загрузки legal-документов')
			};
		}
		const data = (await res.json()) as { pages: LegalPage[] };
		return {
			pages: data.pages ?? [],
			languages: await loadLanguages(event),
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { pages: [], languages: await loadLanguages(event), error: 'API недоступен' };
	}
};

export const actions: Actions = {};
