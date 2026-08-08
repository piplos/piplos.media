import { getApiV1 } from '$lib/api';
import { fetchPortfolioProject } from '$lib/portfolio-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, fetch, platform }) => {
	const ctx = { platform };
	const fromId = url.searchParams.get('from')?.trim();
	const project = fromId
		? await fetchPortfolioProject(fromId, fetch, params.lang, ctx)
		: null;

	return {
		/** Один проект для префилла (?from=slug); полный список здесь не нужен. */
		prefillProject: project,
		apiV1: getApiV1(ctx)
	};
};
