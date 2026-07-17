import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { loadLanguages } from '$lib/languages.server';
import { togglePagePublished } from '$lib/toggle.server';
import type { LegalPage, Page } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const [pagesRes, legalRes] = await Promise.all([
			fetchWithAuth(event, '/v1/pages'),
			fetchWithAuth(event, '/v1/legal')
		]);
		if (!pagesRes.ok || !legalRes.ok) {
			const bad = !pagesRes.ok ? pagesRes : legalRes;
			return {
				pages: [] as Page[],
				legalPages: [] as LegalPage[],
				languages: await loadLanguages(event),
				error: apiLoadErrorMessage(bad, 'Ошибка загрузки страниц')
			};
		}
		const pagesData = (await pagesRes.json()) as { pages: Page[] };
		const legalData = (await legalRes.json()) as { pages: LegalPage[] };
		return {
			pages: pagesData.pages ?? [],
			legalPages: legalData.pages ?? [],
			languages: await loadLanguages(event),
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return {
			pages: [] as Page[],
			legalPages: [] as LegalPage[],
			languages: await loadLanguages(event),
			error: 'API недоступен'
		};
	}
};

export const actions: Actions = {
	togglePublished: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });
		return togglePagePublished(event, id);
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/pages/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить страницу' });
		return { ok: true };
	}
};
