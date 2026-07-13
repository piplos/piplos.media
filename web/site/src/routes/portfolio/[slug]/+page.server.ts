import { loadPortfolioProjects, portfolioProjectEntries } from '$lib/portfolio-api';
import type { EntryGenerator } from './$types';

export const entries: EntryGenerator = async () =>
	portfolioProjectEntries(await loadPortfolioProjects());
