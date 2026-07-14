import { createLeadsLoad, leadsActions } from '../_leads.server';

export const load = createLeadsLoad('new');
export const actions = leadsActions;
