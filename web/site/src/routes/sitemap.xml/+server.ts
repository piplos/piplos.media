import { fetchArticles } from '$lib/articles-api';
import { fetchPortfolioProjects } from '$lib/portfolio-api';
import { fetchServices } from '$lib/services-api';
import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import { SITE } from '$lib/site';
import type { RequestHandler } from './$types';

const STATIC_PATHS = ['', '/portfolio', '/articles', '/order', '/legal/privacy', '/legal/terms', '/legal/cookies'];

/** Динамическая карта сайта: все страницы в обеих локалях с hreflang-альтернативами. */
export const GET: RequestHandler = async ({ fetch, platform }) => {
	const [projects, services, articles] = await Promise.all([
		fetchPortfolioProjects(fetch, {}, { platform }),
		fetchServices(fetch, undefined, { platform }),
		fetchArticles(fetch, undefined, { platform })
	]);

	const paths = [
		...STATIC_PATHS,
		...services.map((s) => `/services/${s.slug}`),
		...projects.map((p) => `/portfolio/${p.slug}`),
		...articles.map((a) => `/articles/${a.slug}`)
	];

	const urls = paths
		.map((path) => {
			const alternates = [
				...SUPPORTED_LANGS.map(
					(lang) =>
						`<xhtml:link rel="alternate" hreflang="${lang}" href="${SITE.url}/${lang}${path}"/>`
				),
				`<xhtml:link rel="alternate" hreflang="x-default" href="${SITE.url}/en${path}"/>`
			].join('');
			return SUPPORTED_LANGS.map(
				(lang) => `<url><loc>${SITE.url}/${lang}${path}</loc>${alternates}</url>`
			).join('');
		})
		.join('');

	const xml = `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">${urls}</urlset>`;

	return new Response(xml, {
		headers: {
			'Content-Type': 'application/xml; charset=utf-8',
			'Cache-Control': 'public, max-age=3600'
		}
	});
};
