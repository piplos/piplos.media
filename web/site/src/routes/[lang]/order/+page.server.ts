import { loadPortfolioProjects } from '$lib/portfolio-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
	return { projects: await loadPortfolioProjects(fetch, { lang: params.lang }) };
};
