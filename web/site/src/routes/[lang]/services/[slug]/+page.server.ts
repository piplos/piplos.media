import { error } from '@sveltejs/kit';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import { loadServicePageItem } from '$lib/services-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	const service = await loadServicePageItem(params.slug, fetch, params.lang, ctx);
	if (!service) throw error(404, 'Service not found');

	const projects = await loadPortfolioProjects(fetch, { lang: params.lang }, ctx);
	const related = projects
		.filter(
			(project) =>
				project.category === service.slug || project.categories.includes(service.slug)
		)
		.slice(0, 3);

	return { service, related };
};
