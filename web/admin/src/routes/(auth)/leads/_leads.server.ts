import { fail, isRedirect, type Actions, type RequestEvent, type ServerLoad } from '@sveltejs/kit';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import type { Lead, LeadStatus } from '$lib/types';
import { LEAD_STATUSES } from './_leads';

const PAGE_SIZE = 50;

const COUNT_KEYS = ['', ...LEAD_STATUSES] as const;

async function fetchLeadTotal(event: RequestEvent, status: string): Promise<number> {
	const q = new URLSearchParams({ limit: '1', offset: '0' });
	if (status) q.set('status', status);
	const res = await fetchWithAuth(event, `/v1/leads?${q}`);
	if (!res.ok) return 0;
	const data = (await res.json()) as { total?: number };
	return data.total ?? 0;
}

export function createLeadsLoad(status: string): ServerLoad {
	return async (event: RequestEvent) => {
		const page = Math.max(1, Number(event.url.searchParams.get('page') ?? '1') || 1);
		const q = new URLSearchParams();
		if (status) q.set('status', status);
		q.set('limit', String(PAGE_SIZE));
		q.set('offset', String((page - 1) * PAGE_SIZE));

		const emptyCounts = { '': 0, new: 0, in_progress: 0, done: 0, spam: 0 };

		try {
			const [listRes, ...countTotals] = await Promise.all([
				fetchWithAuth(event, `/v1/leads?${q}`),
				...COUNT_KEYS.map((s) => fetchLeadTotal(event, s))
			]);

			const counts = Object.fromEntries(
				COUNT_KEYS.map((s, i) => [s, countTotals[i]])
			) as Record<(typeof COUNT_KEYS)[number], number>;

			if (!listRes.ok) {
				return {
					leads: [],
					total: 0,
					counts,
					status: (status || '') as '' | LeadStatus,
					page,
					error: apiLoadErrorMessage(listRes, 'Ошибка загрузки заявок')
				};
			}
			const data = (await listRes.json()) as { leads: Lead[]; total: number };
			return {
				leads: data.leads ?? [],
				total: data.total ?? 0,
				counts,
				status: (status || '') as '' | LeadStatus,
				page,
				error: null
			};
		} catch (e) {
			if (isRedirect(e)) throw e;
			return {
				leads: [],
				total: 0,
				counts: emptyCounts,
				status: (status || '') as '' | LeadStatus,
				page,
				error: 'API недоступен'
			};
		}
	};
}

export const leadsActions: Actions = {
	setStatus: async (event) => {
		const fd = await event.request.formData();
		const id = fd.get('id')?.toString();
		const status = fd.get('status')?.toString();
		if (!id || !status) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/leads/${id}/status`, {
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

		const res = await fetchWithAuth(event, `/v1/leads/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить заявку' });
		return { ok: true };
	}
};
