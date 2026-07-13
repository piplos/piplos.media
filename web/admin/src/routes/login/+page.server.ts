import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import {
	ALLOWED_ROLES,
	COOKIE_ACCESS_TOKEN,
	COOKIE_REFRESH_TOKEN,
	COOKIE_USER,
	cookieOptions
} from '$lib/auth.server';
import { getApiBaseUrl } from '$lib/env.server';

interface LoginApiResponse {
	access_token: string;
	refresh_token: string;
	user: { id: string; email: string; full_name: string; role: string };
}

function getErrorMessage(data: unknown): string {
	if (data && typeof data === 'object') {
		const d = data as Record<string, unknown>;
		if (typeof d.message === 'string') return d.message;
	}
	return 'Неверный email или пароль';
}

export const load: PageServerLoad = async ({ url }) => {
	const urlError = url.searchParams.get('error');
	return { urlError: urlError === 'forbidden' ? 'Недостаточно прав доступа' : null };
};

export const actions: Actions = {
	login: async ({ request, cookies, url, fetch }) => {
		const formData = await request.formData();
		const email = (formData.get('email') as string)?.trim() ?? '';
		const password = (formData.get('password') as string) ?? '';

		if (!email || !password) {
			return fail(400, { email, error: 'Введите email и пароль' });
		}

		const base = getApiBaseUrl();
		let res: Response;
		let data: unknown;
		try {
			res = await fetch(`${base}/api/v1/auth/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password })
			});
			data = await res.json().catch(() => ({}));
		} catch {
			return fail(502, { email, error: 'Сервер авторизации недоступен' });
		}

		if (!res.ok) {
			return fail(res.status === 401 ? 401 : 400, { email, error: getErrorMessage(data) });
		}

		const payload = data as LoginApiResponse;
		if (!payload.user || !ALLOWED_ROLES.includes(payload.user.role)) {
			return fail(403, { email, error: 'Недостаточно прав доступа' });
		}

		const userPayload = {
			id: payload.user.id,
			email: payload.user.email,
			full_name: payload.user.full_name,
			role: payload.user.role
		};
		const opts = cookieOptions(url.protocol === 'https:');
		cookies.set(COOKIE_ACCESS_TOKEN, payload.access_token, opts);
		cookies.set(COOKIE_REFRESH_TOKEN, payload.refresh_token, opts);
		cookies.set(COOKIE_USER, JSON.stringify(userPayload), opts);

		const redirectTo = url.searchParams.get('redirectTo');
		const safe =
			redirectTo &&
			redirectTo.startsWith('/') &&
			!redirectTo.startsWith('//') &&
			redirectTo.length <= 2048;
		throw redirect(303, safe ? redirectTo : '/');
	}
};
