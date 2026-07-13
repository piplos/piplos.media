/**
 * Запросы к API с Bearer-токеном. При 401 — refresh и повтор; при неудаче — редирект на /login.
 */
import { redirect, isRedirect } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	COOKIE_ACCESS_TOKEN,
	COOKIE_REFRESH_TOKEN,
	COOKIE_USER,
	cookieOptions
} from '$lib/auth.server';
import { getApiBaseUrl } from '$lib/env.server';

interface RefreshResponse {
	access_token: string;
	refresh_token: string;
}

export async function refreshTokens(
	event: RequestEvent
): Promise<{ accessToken: string; refreshToken: string } | null> {
	const refreshToken = event.locals.refreshToken;
	if (refreshToken === null || refreshToken === '') return null;

	const base = getApiBaseUrl();
	const res = await event.fetch(`${base}/api/v1/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ refresh_token: refreshToken })
	});
	if (!res.ok) return null;

	const data = (await res.json().catch(() => null)) as RefreshResponse | null;
	if (data?.access_token === undefined || data?.refresh_token === undefined) return null;

	const opts = cookieOptions(event.url.protocol === 'https:');
	event.cookies.set(COOKIE_ACCESS_TOKEN, data.access_token, opts);
	event.cookies.set(COOKIE_REFRESH_TOKEN, data.refresh_token, opts);
	event.locals.accessToken = data.access_token;
	event.locals.refreshToken = data.refresh_token;

	return { accessToken: data.access_token, refreshToken: data.refresh_token };
}

/** Сообщение об ошибке для load-функций. */
export function apiLoadErrorMessage(res: Response, fallback: string): string {
	return res.status === 401 ? 'Сессия истекла' : fallback;
}

function apiUnavailable(): Response {
	return new Response(JSON.stringify({ message: 'API недоступен' }), {
		status: 503,
		headers: { 'Content-Type': 'application/json' }
	});
}

export function redirectToLogin(event: RequestEvent): never {
	event.cookies.delete(COOKIE_ACCESS_TOKEN, { path: '/' });
	event.cookies.delete(COOKIE_REFRESH_TOKEN, { path: '/' });
	event.cookies.delete(COOKIE_USER, { path: '/' });
	const pathAndSearch = event.url.pathname + event.url.search;
	const redirectTo =
		event.url.pathname !== '/login' && pathAndSearch.trim() !== ''
			? `/login?redirectTo=${encodeURIComponent(pathAndSearch)}`
			: '/login';
	throw redirect(303, redirectTo);
}

/** Запрос к API с авторизацией; 401 → refresh + retry; при неудаче — на /login. */
export async function fetchWithAuth(
	event: RequestEvent,
	path: string,
	init?: RequestInit
): Promise<Response> {
	const base = getApiBaseUrl();
	const url = path.startsWith('http') ? path : `${base}${path.startsWith('/') ? '' : '/'}${path}`;

	let token = event.locals.accessToken;
	if (token === null || token === '') {
		try {
			const refreshed = await refreshTokens(event);
			if (refreshed === null) redirectToLogin(event);
			token = refreshed.accessToken;
		} catch (e) {
			if (isRedirect(e)) throw e;
			return apiUnavailable();
		}
	}

	let res: Response;
	try {
		res = await event.fetch(url, {
			...init,
			headers: { ...init?.headers, Authorization: `Bearer ${token}` }
		});
	} catch (e) {
		if (isRedirect(e)) throw e;
		return apiUnavailable();
	}

	if (res.status === 401) {
		try {
			const refreshed = await refreshTokens(event);
			if (refreshed === null) redirectToLogin(event);
			res = await event.fetch(url, {
				...init,
				headers: { ...init?.headers, Authorization: `Bearer ${refreshed.accessToken}` }
			});
		} catch (e) {
			if (isRedirect(e)) throw e;
			return apiUnavailable();
		}
	}

	if (res.status === 401) redirectToLogin(event);
	return res;
}
