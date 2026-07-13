import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { refreshTokens, redirectToLogin } from '$lib/api.server';
import {
	ALLOWED_ROLES,
	COOKIE_ACCESS_TOKEN,
	COOKIE_REFRESH_TOKEN,
	COOKIE_USER,
	isAccessTokenExpired,
	isAdminOnlyPath,
	ROLE_ADMIN
} from '$lib/auth.server';

export const handle: Handle = async ({ event, resolve }) => {
	const accessToken = event.cookies.get(COOKIE_ACCESS_TOKEN);
	const refreshToken = event.cookies.get(COOKIE_REFRESH_TOKEN);
	const userJson = event.cookies.get(COOKIE_USER);

	if (accessToken && userJson) {
		try {
			event.locals.user = JSON.parse(userJson) as App.Locals['user'];
			event.locals.accessToken = accessToken;
			event.locals.refreshToken = refreshToken ?? null;
		} catch {
			event.locals.accessToken = null;
			event.locals.refreshToken = null;
			event.locals.user = null;
		}
	} else {
		event.locals.accessToken = null;
		event.locals.refreshToken = null;
		event.locals.user = null;
	}

	const { pathname } = event.url;

	// Защита маршрутов: доступ только для admin/manager.
	if (pathname !== '/login' && !pathname.startsWith('/logout')) {
		if (!event.locals.user) {
			redirectToLogin(event);
		}
		if (!ALLOWED_ROLES.includes(event.locals.user.role)) {
			throw redirect(303, '/login?error=forbidden');
		}
		// Manager не имеет доступа к пользователям и настройкам.
		if (event.locals.user.role !== ROLE_ADMIN && isAdminOnlyPath(pathname)) {
			throw redirect(303, '/?error=forbidden');
		}
		if (event.locals.accessToken && isAccessTokenExpired(event.locals.accessToken)) {
			try {
				const refreshed = await refreshTokens(event);
				if (refreshed === null) redirectToLogin(event);
			} catch {
				redirectToLogin(event);
			}
		}
	}

	// Авторизован и открыл /login — в админку.
	if (pathname === '/login' && event.locals.user && ALLOWED_ROLES.includes(event.locals.user.role)) {
		throw redirect(303, '/');
	}

	return resolve(event);
};
