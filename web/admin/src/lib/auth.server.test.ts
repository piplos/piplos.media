import { describe, expect, it } from 'vitest';
import { isAccessTokenExpired } from './auth.server';

function makeJwt(exp: number): string {
	const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
	const payload = Buffer.from(JSON.stringify({ exp })).toString('base64url');
	return `${header}.${payload}.signature`;
}

describe('isAccessTokenExpired', () => {
	it('returns false for a token expiring in the future', () => {
		const exp = Math.floor(Date.now() / 1000) + 3600;
		expect(isAccessTokenExpired(makeJwt(exp))).toBe(false);
	});

	it('returns true for an expired token', () => {
		const exp = Math.floor(Date.now() / 1000) - 3600;
		expect(isAccessTokenExpired(makeJwt(exp))).toBe(true);
	});

	it('returns true for malformed tokens', () => {
		expect(isAccessTokenExpired('not-a-jwt')).toBe(true);
		expect(isAccessTokenExpired('')).toBe(true);
	});
});
