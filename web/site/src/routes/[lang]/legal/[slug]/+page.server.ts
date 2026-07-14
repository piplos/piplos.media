import { error } from '@sveltejs/kit';
import { LEGAL_SLUGS, fetchLegalPage, isLegalSlug } from '$lib/legal-api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, platform }) => {
	if (!isLegalSlug(params.slug)) throw error(404, 'Not found');
	// Точечный запрос: один документ и только текущий язык
	// (null — fallback на статический контент внутри resolveLegalDocument).
	const page = await fetchLegalPage(params.slug, fetch, params.lang, { platform });
	return { slug: params.slug, legalPages: page ? [page] : [] };
};
