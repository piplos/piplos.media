import { describe, expect, it, vi } from 'vitest';
import type { RequestEvent } from '@sveltejs/kit';
import { ensureValidSession } from './api.server';
import { COOKIE_ACCESS_TOKEN, COOKIE_REFRESH_TOKEN } from './auth.server';

function makeJwt(exp: number): string {
	const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
	const payload = Buffer.from(JSON.stringify({ exp })).toString('base64url');
	return `${header}.${payload}.signature`;
}

function mockEvent(opts: {
	access?: string | null;
	refresh?: string | null;
	fetchImpl?: (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
}): RequestEvent {
	const cookies = new Map<string, string>();
	if (opts.access) cookies.set(COOKIE_ACCESS_TOKEN, opts.access);
	if (opts.refresh) cookies.set(COOKIE_REFRESH_TOKEN, opts.refresh);

	return {
		url: new URL('http://localhost:5174/projects'),
		cookies: {
			get: (name: string) => cookies.get(name) ?? undefined,
			set: (name: string, value: string) => {
				cookies.set(name, value);
			},
			delete: vi.fn(),
			serialize: vi.fn()
		},
		fetch: vi.fn(opts.fetchImpl ?? (async () => new Response('{}', { status: 500 }))),
		locals: {},
		getClientAddress: () => '127.0.0.1',
		isDataRequest: false,
		isSubRequest: false,
		platform: undefined,
		route: { id: '/(auth)/projects' },
		request: new Request('http://localhost:5174/projects'),
		setHeaders: vi.fn(),
		depends: vi.fn(),
		parent: async () => ({})
	} as unknown as RequestEvent;
}

describe('ensureValidSession', () => {
	it('returns true when access token is still valid', async () => {
		const exp = Math.floor(Date.now() / 1000) + 3600;
		const event = mockEvent({ access: makeJwt(exp), refresh: 'refresh-token' });
		const ok = await ensureValidSession(event);
		expect(ok).toBe(true);
		expect(event.fetch).not.toHaveBeenCalled();
		expect(event.locals.accessToken).toBe(makeJwt(exp));
	});

	it('returns false when no tokens are present', async () => {
		const event = mockEvent({});
		const ok = await ensureValidSession(event);
		expect(ok).toBe(false);
	});

	it('refreshes when access is expired but refresh cookie exists', async () => {
		const expired = makeJwt(Math.floor(Date.now() / 1000) - 3600);
		const fetchImpl = async (_input: RequestInfo | URL, init?: RequestInit) => {
			expect(init?.method).toBe('POST');
			return new Response(
				JSON.stringify({ access_token: 'new-access', refresh_token: 'new-refresh' }),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		};
		const event = mockEvent({ access: expired, refresh: 'old-refresh', fetchImpl });
		const ok = await ensureValidSession(event);
		expect(ok).toBe(true);
		expect(event.locals.accessToken).toBe('new-access');
		expect(event.locals.refreshToken).toBe('new-refresh');
	});

	it('returns false when refresh endpoint fails', async () => {
		const expired = makeJwt(Math.floor(Date.now() / 1000) - 3600);
		const event = mockEvent({
			access: expired,
			refresh: 'old-refresh',
			fetchImpl: async () => new Response('', { status: 401 })
		});
		const ok = await ensureValidSession(event);
		expect(ok).toBe(false);
	});
});
