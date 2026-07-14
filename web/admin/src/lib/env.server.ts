import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

type ApiEnvContext = Pick<RequestEvent, 'platform'>;

/** Origin only — e.g. https://api.piplos.media (no trailing slash). */
export function normalizeApiOrigin(raw: string): string {
	return raw.trim().replace(/\/+$/, '');
}

const DEV_ORIGIN = 'http://localhost:3001';

/** Базовый URL API backend'а (Go, Fiber). */
export function getApiBaseUrl(ctx?: ApiEnvContext): string {
	const platformUrl = ctx?.platform?.env?.ADMIN_API_URL?.trim();
	if (platformUrl) return normalizeApiOrigin(platformUrl);

	const configured = env.ADMIN_API_URL?.trim();
	if (configured) return normalizeApiOrigin(configured);

	return DEV_ORIGIN;
}

/** Versioned path prefix on the API host. */
export const API_V1_PREFIX = '/v1';
