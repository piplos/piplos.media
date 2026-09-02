import type { PageServerLoad } from './$types';
import { createPagesLoad, pagesActions } from './_pages.server';

export const load: PageServerLoad = createPagesLoad('articles');
export const actions = pagesActions;
