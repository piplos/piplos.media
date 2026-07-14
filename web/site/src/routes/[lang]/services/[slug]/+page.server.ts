import { error } from '@sveltejs/kit';
import { DEFAULT_LANG, SUPPORTED_LANGS } from '$lib/i18n/routing';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import {
	loadServicePageItem,
	loadServicePageItems,
	servicePageEntries
} from '$lib/services-api';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = async () => {
	const services = await loadServicePageItems(fetch, DEFAULT_LANG);
	return SUPPORTED_LANGS.flatMap((lang) =>
		servicePageEntries(services).map((entry) => ({ lang, slug: entry.slug }))
	);
};

export const load: PageServerLoad = async ({ params, fetch }) => {
	const service = await loadServicePageItem(params.slug, fetch, params.lang);
	if (!service) throw error(404, 'Service not found');

	const projects = await loadPortfolioProjects(fetch, { lang: params.lang });
	const related = projects
		.filter(
			(project) =>
				project.category === service.slug || project.categories.includes(service.slug)
		)
		.slice(0, 3);

	return { service, related };
};
