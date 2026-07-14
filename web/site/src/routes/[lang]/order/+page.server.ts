import { getApiV1 } from '$lib/api';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	return {
		projects: await loadPortfolioProjects(fetch, { lang: params.lang }, ctx),
		apiV1: getApiV1(ctx)
	};
};
