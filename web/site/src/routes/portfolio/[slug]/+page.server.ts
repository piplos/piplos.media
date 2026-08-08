import { loadPortfolioProjects, portfolioProjectEntries } from '$lib/portfolio-api';
import type { EntryGenerator } from './$types';

/** Только slug-и для prerender — без тяжёлых полей solution/result. */
export const entries: EntryGenerator = async () =>
	portfolioProjectEntries(await loadPortfolioProjects(fetch, { mode: 'summary' }));
