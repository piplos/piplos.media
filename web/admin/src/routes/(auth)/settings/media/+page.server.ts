import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';

export interface RebuildVariantsStatus {
	running: boolean;
	ok: number;
	failed: number;
	error?: string;
	started_at?: string;
	finished_at?: string;
}

const emptyStatus: RebuildVariantsStatus = { running: false, ok: 0, failed: 0 };

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/uploads/rebuild-variants');
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return {
				status: emptyStatus,
				error: data.message ?? apiLoadErrorMessage(res, 'Не удалось загрузить статус WebP')
			};
		}
		const data = (await res.json()) as { status?: RebuildVariantsStatus };
		return { status: data.status ?? emptyStatus, error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { status: emptyStatus, error: 'API недоступен' };
	}
};

export const actions: Actions = {
	rebuild: async (event) => {
		try {
			const res = await fetchWithAuth(event, '/v1/uploads/rebuild-variants', {
				method: 'POST'
			});
			if (!res.ok) {
				const data = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, {
					error: data.message ?? 'Не удалось запустить пересборку WebP'
				});
			}
			const data = (await res.json().catch(() => ({}))) as { status?: RebuildVariantsStatus };
			return { ok: true, status: data.status ?? emptyStatus };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { error: 'API недоступен' });
		}
	}
};
