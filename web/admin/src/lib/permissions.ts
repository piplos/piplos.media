/**
 * Матрица прав UI — синхронизирована с internal/auth/permissions.go.
 * При изменении backend-матрицы обновить этот файл (Phase 4: codegen).
 */

export const GROUP_STAFF = 'staff';
export const GROUP_ADMIN = 'admin';

export const ROLE_ADMIN = 'admin';
export const ROLE_MANAGER = 'manager';

export const STAFF_ROLES = [ROLE_ADMIN, ROLE_MANAGER] as const;
export const ADMIN_ROLES = [ROLE_ADMIN] as const;

export type StaffRole = (typeof STAFF_ROLES)[number];
export type AdminRole = (typeof ADMIN_ROLES)[number];

/** UI-маршруты и требуемая группа (см. docs/plans/auth-refactor.md). */
const UI_ROUTE_GROUPS: { prefix: string; group: string }[] = [
	{ prefix: '/settings', group: GROUP_ADMIN }
];

/** Публичные маршруты без авторизации. */
export function isPublicPath(pathname: string): boolean {
	return pathname === '/login' || pathname === '/logout' || pathname.startsWith('/logout/');
}

export function getRequiredGroup(pathname: string): string | null {
	if (isPublicPath(pathname)) return null;
	for (const { prefix, group } of UI_ROUTE_GROUPS) {
		if (pathname === prefix || pathname.startsWith(prefix + '/')) return group;
	}
	return GROUP_STAFF;
}

export function hasRole(role: string, allowed: readonly string[]): boolean {
	return allowed.includes(role);
}

export function canAccessRoute(role: string, pathname: string): boolean {
	const group = getRequiredGroup(pathname);
	if (group === null) return true;
	if (group === GROUP_ADMIN) return hasRole(role, ADMIN_ROLES);
	if (group === GROUP_STAFF) return hasRole(role, STAFF_ROLES);
	return false;
}

export function isAdminRole(role: string | undefined | null): boolean {
	return role === ROLE_ADMIN;
}
