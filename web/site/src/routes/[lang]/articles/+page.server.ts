import { fetchArticles } from '$lib/articles-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const articles = await fetchArticles(fetch, params.lang, { platform });
	return { articles };
};
