import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	return {
		user: locals.user,
		notifyLeads: locals.user?.notify_leads ?? null
	};
};
