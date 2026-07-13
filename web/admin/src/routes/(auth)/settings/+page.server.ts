import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import type { Language } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/api/v1/languages');
		if (!res.ok) {
			return { languages: [], error: apiLoadErrorMessage(res, 'Ошибка загрузки языков') };
		}
		const data = (await res.json()) as { languages: Language[] };
		return { languages: data.languages ?? [], error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { languages: [], error: 'API недоступен' };
	}
};

export const actions: Actions = {
	saveLanguage: async (event) => {
		const fd = await event.request.formData();
		const payload = {
			code: fd.get('code')?.toString().trim().toLowerCase() ?? '',
			name: fd.get('name')?.toString().trim() ?? '',
			is_default: fd.get('is_default') === 'on',
			enabled: fd.get('enabled') === 'on',
			sort_order: Number.parseInt(fd.get('sort_order')?.toString() ?? '0', 10) || 0
		};
		if (payload.code.length < 2 || !payload.name) {
			return fail(400, { error: 'Код (2-5 символов) и название обязательны' });
		}
		const res = await fetchWithAuth(event, '/api/v1/languages', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить язык' });
		}
		return { ok: true };
	},
	deleteLanguage: async (event) => {
		const code = (await event.request.formData()).get('code')?.toString();
		if (!code) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/api/v1/languages/${code}`, {
			method: 'DELETE'
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось удалить язык' });
		}
		return { ok: true };
	}
};
