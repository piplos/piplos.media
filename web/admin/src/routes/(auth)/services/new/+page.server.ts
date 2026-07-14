import type { PageServerLoad } from './$types';
import { createServiceNewLoad, servicesActions } from '../_services.server';

export const load: PageServerLoad = createServiceNewLoad();
export const actions = servicesActions;
