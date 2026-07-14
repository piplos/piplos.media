import { loadPortfolioProjects } from '$lib/portfolio-api';
import { fetchServices } from '$lib/services-api';
import { fetchStackItems, toStackDisplayItems } from '$lib/stack-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	const [stackFromApi, servicesFromApi, projects] = await Promise.all([
		fetchStackItems(fetch, ctx),
		fetchServices(fetch, params.lang, ctx),
		loadPortfolioProjects(fetch, { lang: params.lang, featured: true }, ctx)
	]);

	return {
		stackItems: toStackDisplayItems(stackFromApi),
		services: servicesFromApi,
		projects
	};
};
