import { error } from '@sveltejs/kit';
import { resolveUploadUrlsInHtml } from '$lib/api';
import { fetchArticle } from '$lib/articles-api';
import {
	loadProjectsForEmbedHtml,
	loadRelatedProjectsForTags
} from '$lib/portfolio-api';
import { fetchSEOPage } from '$lib/seo-api';
import { loadServicesForEmbedHtml } from '$lib/services-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	const [article, seo] = await Promise.all([
		fetchArticle(params.slug, fetch, params.lang, ctx),
		fetchSEOPage(`/articles/${params.slug}`, fetch, ctx)
	]);
	if (!article) throw error(404, 'Article not found');

	for (const locale of Object.values(article.translations)) {
		if (locale.body) locale.body = resolveUploadUrlsInHtml(locale.body, ctx);
	}

	// Токены могут быть только в одном языке; collectEmbedItems дедупит одинаковые params.
	const bodyHtml = Object.values(article.translations)
		.map((locale) => locale.body ?? '')
		.join('\n');

	const [related, projects, services] = await Promise.all([
		loadRelatedProjectsForTags(article.tags ?? [], fetch, { lang: params.lang, limit: 3 }, ctx),
		loadProjectsForEmbedHtml(bodyHtml, fetch, { lang: params.lang }, ctx),
		loadServicesForEmbedHtml(bodyHtml, fetch, { lang: params.lang }, ctx)
	]);

	return { article, related, projects, services, seo };
};
