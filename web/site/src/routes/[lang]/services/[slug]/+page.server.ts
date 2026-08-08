import { error, redirect } from '@sveltejs/kit';
import { resolveUploadUrlsInHtml } from '$lib/api';
import {
	loadProjectsForEmbedHtml,
	loadRelatedProjectsForCategory
} from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
import { loadServicePageItem, loadServicesForEmbedHtml } from '$lib/services-api';
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
	const [service, seo] = await Promise.all([
		loadServicePageItem(params.slug, fetch, params.lang, ctx),
		fetchSEOPage(`/services/${params.slug}`, fetch, ctx)
	]);
	if (!service) throw error(404, 'Service not found');

	service.body = resolveUploadUrlsInHtml(service.body, ctx);

	const [related, projects, services] = await Promise.all([
		loadRelatedProjectsForCategory(service.slug, fetch, { lang: params.lang, limit: 3 }, ctx),
		loadProjectsForEmbedHtml(service.body, fetch, { lang: params.lang }, ctx),
		loadServicesForEmbedHtml(service.body, fetch, { lang: params.lang }, ctx)
	]);

	return { service, related, projects, services, seo };
};
