import { env } from '$env/dynamic/public';
import { dev } from '$app/environment';

/** Origin only — e.g. https://api.piplos.media (no /api suffix). */
export function normalizeApiOrigin(raw: string): string {
	return raw.trim().replace(/\/+$/, '').replace(/\/api$/, '');
}

const DEV_ORIGIN = 'http://localhost:3001';
const PROD_ORIGIN = 'https://api.piplos.media';

function resolveApiOrigin(): string {
	const configured = env.PUBLIC_API_URL?.trim();
	if (configured) return normalizeApiOrigin(configured);
	return dev ? DEV_ORIGIN : PROD_ORIGIN;
}

// Static build: PUBLIC_API_URL фиксируется при сборке (Cloudflare Pages).
export const API_URL = resolveApiOrigin();

/** Versioned API base — e.g. https://api.piplos.media/v1 */
export const API_V1 = `${API_URL}/v1`;
