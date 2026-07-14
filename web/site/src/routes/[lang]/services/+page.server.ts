import { redirect } from '@sveltejs/kit';
import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = () => SUPPORTED_LANGS.map((lang) => ({ lang }));

/** Список услуг — секция #services на главной. */
export const load: PageServerLoad = ({ params }) => {
	throw redirect(308, `/${params.lang}#services`);
};
