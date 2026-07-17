import { error } from '@sveltejs/kit';
import { resolveUploadUrlsInHtml } from '$lib/api';
import { fetchArticle } from '$lib/articles-api';
import { loadPortfolioProjects } from '$lib/portfolio-api';
import type { PortfolioProject } from '$lib/portfolio';
import { fetchSEOPage } from '$lib/seo-api';
import type { PageServerLoad } from './$types';

/** До 3 связанных проектов: пересечение по стеку статьи, иначе — сквозной порядок портфолио. */
function relatedProjects(projects: PortfolioProject[], articleTags: string[]): PortfolioProject[] {
	const tags = new Set(articleTags.filter(Boolean));
	const matched =
		tags.size > 0
			? projects.filter((project) => project.tags.some((tag) => tags.has(tag)))
			: [];
	const pool = matched.length > 0 ? matched : projects;
	return pool.slice(0, 3);
}

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	const ctx = { platform };
	const [article, projects, seo] = await Promise.all([
		fetchArticle(params.slug, fetch, params.lang, ctx),
		loadPortfolioProjects(fetch, { lang: params.lang }, ctx),
		fetchSEOPage(`/articles/${params.slug}`, fetch, ctx)
	]);
	if (!article) throw error(404, 'Article not found');

	for (const locale of Object.values(article.translations)) {
		if (locale.body) locale.body = resolveUploadUrlsInHtml(locale.body, ctx);
	}

	const related = relatedProjects(projects, article.tags ?? []);

	return { article, related, seo };
};
