import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { COOKIE_ACCESS_TOKEN, COOKIE_REFRESH_TOKEN } from '$lib/auth.server';
import { API_V1_PREFIX, getApiBaseUrl } from '$lib/env.server';

export const actions: Actions = {
	default: async ({ cookies, fetch, locals, platform }) => {
		const accessToken = locals.accessToken ?? cookies.get(COOKIE_ACCESS_TOKEN) ?? null;
		const refreshToken = locals.refreshToken ?? cookies.get(COOKIE_REFRESH_TOKEN) ?? null;

		if (accessToken || refreshToken) {
			try {
				await fetch(`${getApiBaseUrl({ platform })}${API_V1_PREFIX}/auth/logout`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
						...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {})
					},
					body: JSON.stringify({ refresh_token: refreshToken ?? '' })
				});
			} catch {
				// Серверная ревокация best-effort; cookies всё равно очищаем.
			}
		}

		cookies.delete(COOKIE_ACCESS_TOKEN, { path: '/' });
		cookies.delete(COOKIE_REFRESH_TOKEN, { path: '/' });
		cookies.delete('admin_user', { path: '/' });
		throw redirect(303, '/login');
	}
};
