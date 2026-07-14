import { createLeadsLoad, leadsActions } from '../_leads.server';

export const load = createLeadsLoad('in_progress');
export const actions = leadsActions;
