import { loadPortfolioProjects } from '$lib/portfolio-api';
import { fetchServices } from '$lib/services-api';
import { fetchStackItems, toStackDisplayItems } from '$lib/stack-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	const [stackFromApi, servicesFromApi, projects] = await Promise.all([
		fetchStackItems(fetch, ctx),
		// Карточки услуг: title/description; body не нужен.
		fetchServices(fetch, { lang: params.lang, mode: 'summary' }, ctx),
		// Блок «избранное» на главной — ровно 3 карточки.
		loadPortfolioProjects(
			fetch,
			{ lang: params.lang, featured: true, limit: 3, mode: 'summary' },
			ctx
		)
	]);

	return {
		stackItems: toStackDisplayItems(stackFromApi),
		services: servicesFromApi,
		projects
	};
};
