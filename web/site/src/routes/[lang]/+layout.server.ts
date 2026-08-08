import { loadServicePageItems } from '$lib/services-api';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ params, fetch, platform }) => {
	// Футеру нужны только slug + title — без body HTML на каждой странице.
	const services = await loadServicePageItems(fetch, params.lang, { platform }, { mode: 'summary' });
	return {
		footerServices: services.map((service) => ({
			slug: service.slug,
			title: service.title
		}))
	};
};
