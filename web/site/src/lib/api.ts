import { env } from '$env/dynamic/public';
import { building, dev } from '$app/environment';
import type { RequestEvent } from '@sveltejs/kit';

/** Origin only — e.g. https://api.piplos.media (no /api suffix). */
export function normalizeApiOrigin(raw: string): string {
	return raw.trim().replace(/\/+$/, '').replace(/\/api$/, '');
}

const DEV_ORIGIN = 'http://localhost:3001';
const PROD_ORIGIN = 'https://api.piplos.media';

export type ApiRequestContext = Pick<RequestEvent, 'platform'>;

/** Базовый URL API: build env → Worker runtime → fallback. */
export function getApiBaseUrl(ctx?: ApiRequestContext): string {
	const configured = env.PUBLIC_API_URL?.trim();
	if (configured) return normalizeApiOrigin(configured);

	if (!building && ctx?.platform) {
		const runtime = ctx.platform.env?.PUBLIC_API_URL?.trim();
		if (runtime) return normalizeApiOrigin(runtime);
	}

	return dev ? DEV_ORIGIN : PROD_ORIGIN;
}

/** Versioned API base — e.g. https://api.piplos.media/v1 */
export function getApiV1(ctx?: ApiRequestContext): string {
	return `${getApiBaseUrl(ctx)}/v1`;
}

/** Файлы архива хранятся относительными путями (/uploads/...) — разворачивает их в URL API. */
export function resolveUploadUrl(path: string, ctx?: ApiRequestContext): string {
	return path.startsWith('/uploads/') ? getApiBaseUrl(ctx) + path : path;
}

/** Переписывает src/href на /uploads/... внутри готового HTML на абсолютные URL API. */
export function resolveUploadUrlsInHtml(html: string, ctx?: ApiRequestContext): string {
	if (!html.includes('/uploads/')) return html;
	const base = getApiBaseUrl(ctx);
	return html
		.replaceAll('src="/uploads/', `src="${base}/uploads/`)
		.replaceAll('href="/uploads/', `href="${base}/uploads/`);
}
