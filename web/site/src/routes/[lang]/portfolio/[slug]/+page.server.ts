import { error, redirect } from '@sveltejs/kit';
import { fetchPortfolioProject } from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
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

	const [project, seo] = await Promise.all([
		fetchPortfolioProject(params.slug, fetch, params.lang, { platform }),
		fetchSEOPage(`/portfolio/${params.slug}`, fetch, { platform })
	]);
	if (!project) throw error(404, 'Project not found');
	return { project, seo };
};
