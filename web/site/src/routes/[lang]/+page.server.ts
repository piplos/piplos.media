import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import { fetchServices } from '$lib/services-api';
import { fetchStackItems, toStackDisplayItems } from '$lib/stack-api';
import { STACK_ITEMS } from '$lib/constants/stack';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = () => SUPPORTED_LANGS.map((lang) => ({ lang }));

export const load: PageServerLoad = async ({ params, fetch }) => {
	// Главной нужны только featured-проекты и перевод текущего языка.
	const [stackFromApi, servicesFromApi, projects] = await Promise.all([
		fetchStackItems(fetch),
		fetchServices(fetch, params.lang),
		loadPortfolioProjects(fetch, { lang: params.lang, featured: true })
	]);

	const stackItems =
		stackFromApi.length > 0
			? toStackDisplayItems(stackFromApi)
			: STACK_ITEMS.map((item) => ({ slug: item.id, label: item.label }));

	return { stackItems, services: servicesFromApi, projects };
};
