import { error, redirect } from '@sveltejs/kit';
import { resolveUploadUrlsInHtml } from '$lib/api';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
import { loadServicePageItem } from '$lib/services-api';
import type { PageServerLoad } from './$types';

/** Слаги услуг старого сайта → новые (301 для сохранения SEO-веса). */
const LEGACY_SLUGS: Record<string, string> = {
	'web-application-development': 'web',
	'custom-software-development': 'backend',
	'mobile-app-development': 'mobile',
	'maintenance-and-support': 'devops',
	'quality-assurance-and-testing': 'backend',
	'ui-ux-design-services': 'web'
};

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const legacy = LEGACY_SLUGS[params.slug];
	if (legacy) throw redirect(301, `/${params.lang}/services/${legacy}`);

	const ctx = { platform };
	const [service, projects, seo] = await Promise.all([
		loadServicePageItem(params.slug, fetch, params.lang, ctx),
		loadPortfolioProjects(fetch, { lang: params.lang }, ctx),
		fetchSEOPage(`/services/${params.slug}`, fetch, ctx)
	]);
	if (!service) throw error(404, 'Service not found');

	service.body = resolveUploadUrlsInHtml(service.body, ctx);

	const related = projects
		.filter(
			(project) =>
				project.category === service.slug || project.categories.includes(service.slug)
		)
		.slice(0, 3);

	return { service, related, seo };
};
