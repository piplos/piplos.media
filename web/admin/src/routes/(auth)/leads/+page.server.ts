import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import type { Lead } from '$lib/types';

const PAGE_SIZE = 50;

export const load: PageServerLoad = async (event) => {
	const status = event.url.searchParams.get('status') ?? '';
	const page = Math.max(1, Number(event.url.searchParams.get('page') ?? '1') || 1);
	const q = new URLSearchParams();
	if (status) q.set('status', status);
	q.set('limit', String(PAGE_SIZE));
	q.set('offset', String((page - 1) * PAGE_SIZE));

	try {
		const res = await fetchWithAuth(event, `/api/v1/leads?${q.toString()}`);
		if (!res.ok) {
			return { leads: [], total: 0, status, page, error: apiLoadErrorMessage(res, 'Ошибка загрузки заявок') };
		}
		const data = (await res.json()) as { leads: Lead[]; total: number };
		return { leads: data.leads ?? [], total: data.total ?? 0, status, page, error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { leads: [], total: 0, status, page, error: 'API недоступен' };
	}
};

export const actions: Actions = {
	setStatus: async (event) => {
		const fd = await event.request.formData();
		const id = fd.get('id')?.toString();
		const status = fd.get('status')?.toString();
		if (!id || !status) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/api/v1/leads/${id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status })
		});
		if (!res.ok) return fail(res.status, { error: 'Не удалось обновить статус' });
		return { ok: true };
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/api/v1/leads/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить заявку' });
		return { ok: true };
	}
};
