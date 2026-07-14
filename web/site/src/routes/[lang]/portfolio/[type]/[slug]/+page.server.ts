import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

/** Старый сайт использовал URL вида /ru/portfolio/{type}/{slug} —
 *  постоянный редирект на новую структуру /ru/portfolio/{slug}. */
const LEGACY_TYPES = new Set(['sites', 'soft', 'landing', 'app', 'smm']);

export const load: PageServerLoad = ({ params }) => {
	if (!LEGACY_TYPES.has(params.type)) throw error(404, 'Not found');
	redirect(301, `/${params.lang}/portfolio/${params.slug}`);
};
