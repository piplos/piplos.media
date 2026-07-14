import { error } from '@sveltejs/kit';
import { fetchPortfolioProject } from '$lib/portfolio-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const project = await fetchPortfolioProject(params.slug, fetch, params.lang, { platform });
	if (!project) throw error(404, 'Project not found');
	return { project };
};
