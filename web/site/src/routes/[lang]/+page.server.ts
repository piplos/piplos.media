import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import { fetchServices } from '$lib/services-api';
import { fetchStackItems, toStackDisplayItems } from '$lib/stack-api';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = () => SUPPORTED_LANGS.map((lang) => ({ lang }));

export const load: PageServerLoad = async ({ params, fetch }) => {
	const [stackFromApi, servicesFromApi, projects] = await Promise.all([
		fetchStackItems(fetch),
		fetchServices(fetch, params.lang),
		loadPortfolioProjects(fetch, { lang: params.lang, featured: true })
	]);

	return {
		stackItems: toStackDisplayItems(stackFromApi),
		services: servicesFromApi,
		projects
	};
};
