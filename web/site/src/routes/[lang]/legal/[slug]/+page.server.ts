import { error } from '@sveltejs/kit';
import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import { LEGAL_SLUGS, fetchLegalPages, isLegalSlug } from '$lib/legal-api';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = () =>
	SUPPORTED_LANGS.flatMap((lang) => LEGAL_SLUGS.map((slug) => ({ lang, slug })));

export const load: PageServerLoad = async ({ params, fetch }) => {
	if (!isLegalSlug(params.slug)) throw error(404, 'Not found');
	return { slug: params.slug, legalPages: await fetchLegalPages(fetch) };
};
