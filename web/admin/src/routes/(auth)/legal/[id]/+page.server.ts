import { error, fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { legalPayload } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import type { LegalPage } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, `/api/v1/legal/${event.params.id}`);
		if (res.status === 404) throw error(404, 'Документ не найден');
		if (!res.ok) throw error(res.status, 'Ошибка загрузки документа');
		const data = (await res.json()) as { page: LegalPage };
		return { page: data.page, languages: await loadLanguages(event) };
	} catch (e) {
		if (isRedirect(e)) throw e;
		throw e;
	}
};

export const actions: Actions = {
	save: async (event) => {
		const payload = legalPayload(await event.request.formData());
		const res = await fetchWithAuth(event, `/api/v1/legal/${event.params.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить документ' });
		}
		return { ok: true };
	}
};
