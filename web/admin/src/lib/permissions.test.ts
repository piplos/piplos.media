import { describe, expect, it } from 'vitest';
import {
	canAccessRoute,
	getRequiredGroup,
	GROUP_ADMIN,
	GROUP_STAFF,
	isPublicPath,
	ROLE_ADMIN,
	ROLE_MANAGER
} from './permissions';

describe('permissions', () => {
	it('treats login and logout as public', () => {
		expect(isPublicPath('/login')).toBe(true);
		expect(isPublicPath('/logout')).toBe(true);
		expect(isPublicPath('/projects')).toBe(false);
	});

	it('requires admin group for settings routes', () => {
		expect(getRequiredGroup('/settings')).toBe(GROUP_ADMIN);
		expect(getRequiredGroup('/settings/users')).toBe(GROUP_ADMIN);
		expect(getRequiredGroup('/projects')).toBe(GROUP_STAFF);
	});

	it('blocks manager from settings', () => {
		expect(canAccessRoute(ROLE_MANAGER, '/settings')).toBe(false);
		expect(canAccessRoute(ROLE_MANAGER, '/settings/users')).toBe(false);
		expect(canAccessRoute(ROLE_MANAGER, '/projects')).toBe(true);
	});

	it('allows admin everywhere staff can go and in settings', () => {
		expect(canAccessRoute(ROLE_ADMIN, '/settings')).toBe(true);
		expect(canAccessRoute(ROLE_ADMIN, '/projects')).toBe(true);
	});
});
