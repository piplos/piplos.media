import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { createLeadsLoad, leadsActions } from './_leads.server';
import { isLeadStatus, STATUS_TO_SLUG } from './_leads';

export const load: PageServerLoad = async (event) => {
	const legacy = event.url.searchParams.get('status');
	if (legacy !== null) {
		if (!legacy || !isLeadStatus(legacy)) redirect(301, '/leads');
		redirect(301, `/leads/${STATUS_TO_SLUG[legacy]}`);
	}
	return createLeadsLoad('')(event);
};

export const actions = leadsActions;
