import { error } from '@sveltejs/kit';
import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import {
	fetchPortfolioProject,
	loadPortfolioProjects,
	portfolioProjectEntries
} from '$lib/portfolio-api';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = async () => {
	const projects = await loadPortfolioProjects();
	return SUPPORTED_LANGS.flatMap((lang) =>
		portfolioProjectEntries(projects).map((entry) => ({ lang, slug: entry.slug }))
	);
};

export const load: PageServerLoad = async ({ params, fetch }) => {
	const project = await fetchPortfolioProject(params.slug, fetch, params.lang);
	if (!project) throw error(404, 'Project not found');
	return { project };
};
