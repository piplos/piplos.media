import { error, redirect } from '@sveltejs/kit';
import { fetchPortfolioProject, loadProjectsForEmbedHtml } from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
import { loadServicesForEmbedHtml } from '$lib/services-api';
import type { PageServerLoad } from './$types';

/** Разделы портфолио старого сайта (/portfolio/{type}) → фильтр нового списка (301). */
const LEGACY_TYPE_FILTERS: Record<string, string> = {
	sites: 'web',
	landing: 'web',
	smm: 'web',
	app: 'mobile',
	soft: 'backend'
};

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const legacyFilter = LEGACY_TYPE_FILTERS[params.slug];
	if (legacyFilter) throw redirect(301, `/${params.lang}/portfolio?filter=${legacyFilter}`);

	const ctx = { platform };
	const [project, seo] = await Promise.all([
		fetchPortfolioProject(params.slug, fetch, params.lang, ctx),
		fetchSEOPage(`/portfolio/${params.slug}`, fetch, ctx)
	]);
	if (!project) throw error(404, 'Project not found');

	// Токены могут быть только в одном языке; collectEmbedItems дедупит запросы.
	const solutionHtml = `${project.en.solution}\n${project.ru.solution}`;

	const [projects, services] = await Promise.all([
		loadProjectsForEmbedHtml(solutionHtml, fetch, { lang: params.lang }, ctx),
		loadServicesForEmbedHtml(solutionHtml, fetch, { lang: params.lang }, ctx)
	]);

	return { project, projects, services, seo };
};
