import type { PageServerLoad } from './$types';
import { createServiceEditLoad, servicesActions } from '../_services.server';

export const load: PageServerLoad = createServiceEditLoad();
export const actions = servicesActions;
