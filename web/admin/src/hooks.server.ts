import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { ensureValidSession, fetchCurrentUser, redirectToLogin } from '$lib/api.server';
import { canAccessRoute, isPublicPath, STAFF_ROLES } from '$lib/permissions';

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	if (!isPublicPath(pathname)) {
		const hasSession = await ensureValidSession(event);
		if (!hasSession) {
			redirectToLogin(event);
		}

		const user = await fetchCurrentUser(event);
		if (!user || !STAFF_ROLES.includes(user.role)) {
			redirectToLogin(event);
		}

		if (!canAccessRoute(user.role, pathname)) {
			throw redirect(303, '/?error=forbidden');
		}

		event.locals.user = user;
	} else {
		event.locals.accessToken = null;
		event.locals.refreshToken = null;
		event.locals.user = null;

		const hasSession = await ensureValidSession(event);
		if (hasSession) {
			const user = await fetchCurrentUser(event);
			if (user && STAFF_ROLES.includes(user.role)) {
				event.locals.user = user;
			}
		}
	}

	if (pathname === '/login' && event.locals.user && STAFF_ROLES.includes(event.locals.user.role)) {
		throw redirect(303, '/');
	}

	return resolve(event);
};
