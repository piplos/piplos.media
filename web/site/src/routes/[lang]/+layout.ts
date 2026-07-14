import { error } from '@sveltejs/kit';
import { isLang } from '$lib/i18n/routing';
import type { LayoutLoad } from './$types';

export const prerender = false;

export const load: LayoutLoad = ({ params, data }) => {
	if (!isLang(params.lang)) throw error(404, 'Not found');
	return {
		lang: params.lang,
		footerServices: data.footerServices ?? []
	};
};
