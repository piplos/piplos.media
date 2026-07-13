import { error } from '@sveltejs/kit';
import { isLang } from '$lib/i18n/routing';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = ({ params }) => {
	if (!isLang(params.lang)) throw error(404, 'Not found');
	return { lang: params.lang };
};
