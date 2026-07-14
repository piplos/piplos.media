import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

/** Редирект со старых URL `/lists/stack`. */
export const load: PageServerLoad = () => {
	throw redirect(301, '/stack');
};
