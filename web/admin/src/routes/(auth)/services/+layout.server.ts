import type { LayoutServerLoad } from './$types';
import { loadServicesPageData } from './_services.server';

export const load: LayoutServerLoad = async (event) => {
	const data = await loadServicesPageData(event);
	return {
		services: data.services,
		languages: data.languages,
		stack: data.stack,
		error: data.error
	};
};
