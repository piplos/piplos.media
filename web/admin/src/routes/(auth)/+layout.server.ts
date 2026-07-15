import { isRedirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';

export const load: LayoutServerLoad = async (event) => {
	// Актуальный флаг уведомлений о заявках — из API (меняется админом, cookie устаревает).
	let notifyLeads: boolean | null = null;
	try {
		const res = await fetchWithAuth(event, '/v1/auth/me');
		if (res.ok) {
			const data = (await res.json().catch(() => null)) as {
				user?: { notify_leads?: boolean };
			} | null;
			notifyLeads = data?.user?.notify_leads ?? null;
		}
	} catch (e) {
		if (isRedirect(e)) throw e;
		// API недоступен — статус просто не показываем.
	}
	return { user: event.locals.user, notifyLeads };
};
