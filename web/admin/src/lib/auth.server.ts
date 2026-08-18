/**
 * Константы и утилиты авторизации (только сервер).
 * Единое место для имён cookie и TTL.
 */
export const COOKIE_ACCESS_TOKEN = 'admin_access_token';
export const COOKIE_REFRESH_TOKEN = 'admin_refresh_token';

/** Совпадает с JWT_EXPIRATION_MINUTES (15 min). */
export const ACCESS_TOKEN_MAX_AGE = 60 * 15;

/** Совпадает с JWT_REFRESH_EXPIRATION_HOURS (7 d). */
export const REFRESH_TOKEN_MAX_AGE = 60 * 60 * 24 * 7;

/** Проверяет, истёк ли access token (JWT) по полю exp. */
export function isAccessTokenExpired(token: string, leewaySeconds = 60): boolean {
	try {
		const parts = token.split('.');
		if (parts.length !== 3) return true;
		const payload = JSON.parse(Buffer.from(parts[1]!, 'base64url').toString('utf8')) as {
			exp?: number;
		};
		if (typeof payload.exp !== 'number') return true;
		return payload.exp * 1000 < Date.now() + leewaySeconds * 1000;
	} catch {
		return true;
	}
}

export function accessCookieOptions(secure: boolean) {
	return {
		path: '/' as const,
		httpOnly: true,
		secure,
		sameSite: 'lax' as const,
		maxAge: ACCESS_TOKEN_MAX_AGE
	};
}

export function refreshCookieOptions(secure: boolean) {
	return {
		path: '/' as const,
		httpOnly: true,
		secure,
		sameSite: 'lax' as const,
		maxAge: REFRESH_TOKEN_MAX_AGE
	};
}
