import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

/** Страницы «Команда» и вакансий старого сайта (/team, /team/vacancy-*) —
 *  ближайший аналог: секция #about на главной (301). */
export const load: PageServerLoad = ({ params }) => {
	throw redirect(301, `/${params.lang}#about`);
};
