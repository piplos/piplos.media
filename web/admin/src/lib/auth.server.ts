/**
 * Константы и утилиты авторизации (только сервер).
 * Единое место для имён cookie и ролей.
 */
export const COOKIE_ACCESS_TOKEN = 'admin_access_token';
export const COOKIE_REFRESH_TOKEN = 'admin_refresh_token';
export const COOKIE_USER = 'admin_user';

export const ROLE_ADMIN = 'admin';
export const ROLE_MANAGER = 'manager';
export const ALLOWED_ROLES = [ROLE_ADMIN, ROLE_MANAGER];

/** Разделы только для админа: manager не видит настройки (включая пользователей). */
export const ADMIN_ONLY_PATHS = ['/settings'];

export function isAdminOnlyPath(pathname: string): boolean {
	return ADMIN_ONLY_PATHS.some((p) => pathname === p || pathname.startsWith(p + '/'));
}

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

export function cookieOptions(secure: boolean) {
	return {
		path: '/' as const,
		httpOnly: true,
		secure,
		sameSite: 'lax' as const,
		maxAge: 60 * 60 * 24 * 7
	};
}
