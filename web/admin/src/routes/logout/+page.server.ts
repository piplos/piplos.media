import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { COOKIE_ACCESS_TOKEN, COOKIE_REFRESH_TOKEN, COOKIE_USER } from '$lib/auth.server';

export const actions: Actions = {
	default: async ({ cookies }) => {
		cookies.delete(COOKIE_ACCESS_TOKEN, { path: '/' });
		cookies.delete(COOKIE_REFRESH_TOKEN, { path: '/' });
		cookies.delete(COOKIE_USER, { path: '/' });
		throw redirect(303, '/login');
	}
};
