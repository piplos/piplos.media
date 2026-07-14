import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { loadServices } from '$lib/lists.server';
import { newProjectHref } from '../_projects';

export const load: PageServerLoad = async (event) => {
	const services = await loadServices(event);
	throw redirect(301, newProjectHref('', services));
};
