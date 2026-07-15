import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import { SITE } from '$lib/site';
import type { RequestHandler } from './$types';

/** Дополнительные пути для Disallow — добавляйте сюда вручную при необходимости. */
const EXTRA_DISALLOW_PATHS = [
	// '/admin/',
	// '/api/private/',
] as const;

/** robots.txt: legal-страницы не индексируем (noindex в HTML + Disallow здесь). */
export const GET: RequestHandler = () => {
	const disallowPaths = [
		...SUPPORTED_LANGS.map((lang) => `/${lang}/legal/`),
		...EXTRA_DISALLOW_PATHS
	];

	const disallowRules = disallowPaths.map((path) => `Disallow: ${path}`).join('\n');

	const body = `# https://piplos.media robots.txt
User-agent: *
${disallowRules}

Sitemap: ${SITE.url}/sitemap.xml
`;

	return new Response(`${body.trim()}\n`, {
		headers: {
			'Content-Type': 'text/plain; charset=utf-8',
			'Cache-Control': 'public, max-age=3600'
		}
	});
};
