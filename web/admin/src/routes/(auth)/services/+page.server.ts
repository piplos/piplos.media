import type { PageServerLoad } from './$types';
import { createServicesListLoad, servicesActions } from './_services.server';

export const load: PageServerLoad = createServicesListLoad();
export const actions = servicesActions;
