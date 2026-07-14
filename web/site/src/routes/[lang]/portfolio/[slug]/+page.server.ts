import { error } from '@sveltejs/kit';
import { fetchPortfolioProject } from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const [project, seo] = await Promise.all([
		fetchPortfolioProject(params.slug, fetch, params.lang, { platform }),
		fetchSEOPage(`/portfolio/${params.slug}`, fetch, { platform })
	]);
	if (!project) throw error(404, 'Project not found');
	return { project, seo };
};
