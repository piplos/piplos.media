import { loadServicePageItems } from '$lib/services-api';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ params, fetch }) => {
	const services = await loadServicePageItems(fetch, params.lang);
	return {
		footerServices: services.map((service) => ({
			slug: service.slug,
			title: service.title
		}))
	};
};
