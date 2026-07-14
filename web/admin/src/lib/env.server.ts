import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

/** Origin only — e.g. https://api.piplos.media (no /api suffix). */
export function normalizeApiOrigin(raw: string): string {
	return raw.trim().replace(/\/+$/, '').replace(/\/api$/, '');
}

const DEV_ORIGIN = 'http://localhost:3001';

/** Базовый URL API backend'а (Go, Fiber). */
export function getApiBaseUrl(event?: Pick<RequestEvent, 'platform'>): string {
	const configured = env.ADMIN_API_URL?.trim();
	if (configured) return normalizeApiOrigin(configured);

	const platformUrl = event?.platform?.env?.ADMIN_API_URL?.trim();
	if (platformUrl) return normalizeApiOrigin(platformUrl);

	return DEV_ORIGIN;
}

/** Versioned path prefix on the API host. */
export const API_V1_PREFIX = '/v1';
