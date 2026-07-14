import { error, fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import type { Lead } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, `/v1/leads/${event.params.id}`);
		if (res.status === 404) throw error(404, 'Заявка не найдена');
		if (!res.ok) throw error(res.status, 'Ошибка загрузки заявки');
		const data = (await res.json()) as { lead: Lead };
		return { lead: data.lead };
	} catch (e) {
		if (isRedirect(e)) throw e;
		throw e;
	}
};

export const actions: Actions = {
	setStatus: async (event) => {
		const status = (await event.request.formData()).get('status')?.toString();
		if (!status) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/leads/${event.params.id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status })
		});
		if (!res.ok) return fail(res.status, { error: 'Не удалось обновить статус' });
		return { ok: true };
	}
};
