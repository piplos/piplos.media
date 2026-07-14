import { createLeadsLoad, leadsActions } from '../_leads.server';

export const load = createLeadsLoad('spam');
export const actions = leadsActions;
