import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

/** Редирект со старых URL `/projects/group/{slug}`. */
export const load: PageServerLoad = ({ params }) => {
	throw redirect(301, `/projects/${params.slug}`);
};
