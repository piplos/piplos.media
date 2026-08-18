/**
 * BFF-сессия и запросы к API. Единственная точка refresh — ensureValidSession.
 */
import { redirect } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	COOKIE_ACCESS_TOKEN,
	COOKIE_REFRESH_TOKEN,
	accessCookieOptions,
	isAccessTokenExpired,
	refreshCookieOptions
} from '$lib/auth.server';
import { getApiBaseUrl } from '$lib/env.server';
import type { AdminUser } from '$lib/types';

interface AuthTokenResponse {
	access_token: string;
	refresh_token: string;
	user?: AdminUser;
}

/** Единственная точка обновления access token (hooks вызывают до loaders). */
export async function ensureValidSession(event: RequestEvent): Promise<boolean> {
	const secure = event.url.protocol === 'https:';
	const accessToken = event.cookies.get(COOKIE_ACCESS_TOKEN) ?? null;
	const refreshToken = event.cookies.get(COOKIE_REFRESH_TOKEN) ?? null;

	event.locals.accessToken = accessToken;
	event.locals.refreshToken = refreshToken;

	if (accessToken && !isAccessTokenExpired(accessToken)) {
		return true;
	}

	if (!refreshToken) {
		return false;
	}

	const res = await event.fetch(`${getApiBaseUrl(event)}/v1/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ refresh_token: refreshToken })
	});
	if (!res.ok) return false;

	const data = (await res.json().catch(() => null)) as AuthTokenResponse | null;
	if (data?.access_token === undefined || data?.refresh_token === undefined) return false;

	event.cookies.set(COOKIE_ACCESS_TOKEN, data.access_token, accessCookieOptions(secure));
	event.cookies.set(COOKIE_REFRESH_TOKEN, data.refresh_token, refreshCookieOptions(secure));
	event.locals.accessToken = data.access_token;
	event.locals.refreshToken = data.refresh_token;

	return true;
}

/** Загружает актуальный профиль из /v1/auth/me (источник правды для role, notify_leads). */
export async function fetchCurrentUser(event: RequestEvent): Promise<AdminUser | null> {
	const token = event.locals.accessToken;
	if (!token) return null;

	const res = await event.fetch(`${getApiBaseUrl(event)}/v1/auth/me`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	if (!res.ok) return null;

	const data = (await res.json().catch(() => null)) as { user?: AdminUser } | null;
	return data?.user ?? null;
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
	// Удаляем legacy-cookie после миграции с 3-cookie модели.
	event.cookies.delete('admin_user', { path: '/' });
	const pathAndSearch = event.url.pathname + event.url.search;
	const redirectTo =
		event.url.pathname !== '/login' && pathAndSearch.trim() !== ''
			? `/login?redirectTo=${encodeURIComponent(pathAndSearch)}`
			: '/login';
	throw redirect(303, redirectTo);
}

/** Запрос к API с Bearer из locals (refresh выполняется в hooks до loaders). */
export async function fetchWithAuth(
	event: RequestEvent,
	path: string,
	init?: RequestInit
): Promise<Response> {
	const url = path.startsWith('http')
		? path
		: `${getApiBaseUrl(event)}${path.startsWith('/') ? path : `/${path}`}`;

	const token = event.locals.accessToken;
	if (token === null || token === '') {
		redirectToLogin(event);
	}

	let res: Response;
	try {
		res = await event.fetch(url, {
			...init,
			headers: { ...init?.headers, Authorization: `Bearer ${token}` }
		});
	} catch {
		return apiUnavailable();
	}

	if (res.status === 401) redirectToLogin(event);
	return res;
}
