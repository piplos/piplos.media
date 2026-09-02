import type { PageServerLoad } from './$types';
import { createPagesLoad, pagesActions } from '../_pages.server';

export const load: PageServerLoad = createPagesLoad('legal');
export const actions = pagesActions;
