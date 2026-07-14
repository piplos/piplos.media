import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

/** Страница «О компании» старого сайта — секция #about на главной (301). */
export const load: PageServerLoad = ({ params }) => {
	throw redirect(301, `/${params.lang}#about`);
};
