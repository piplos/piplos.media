import { env } from '$env/dynamic/public';
import { building, dev } from '$app/environment';
import type { RequestEvent } from '@sveltejs/kit';

/** Origin only — e.g. https://api.piplos.media (no trailing slash). */
export function normalizeApiOrigin(raw: string): string {
	return raw.trim().replace(/\/+$/, '');
}

const DEV_ORIGIN = 'http://localhost:3001';
const PROD_ORIGIN = 'https://api.piplos.media';

export type ApiRequestContext = Pick<RequestEvent, 'platform'>;

/** Базовый URL API: Worker runtime → build env → fallback.
 *  В dev эмуляция platform.env берёт прод-URL из wrangler.toml [vars] — пропускаем её. */
export function getApiBaseUrl(ctx?: ApiRequestContext): string {
	if (!building && !dev && ctx?.platform) {
		const runtime = ctx.platform.env?.PUBLIC_API_URL?.trim();
		if (runtime) return normalizeApiOrigin(runtime);
	}

	const configured = env.PUBLIC_API_URL?.trim();
	if (configured) return normalizeApiOrigin(configured);

	return dev ? DEV_ORIGIN : PROD_ORIGIN;
}

/** Versioned API base — e.g. https://api.piplos.media/v1 */
export function getApiV1(ctx?: ApiRequestContext): string {
	return `${getApiBaseUrl(ctx)}/v1`;
}

/** Файлы архива хранятся относительными путями (/uploads/...) — разворачивает их в URL API. */
export function resolveUploadUrl(path: string, ctx?: ApiRequestContext): string {
	const raw = path.startsWith('/uploads/') ? getApiBaseUrl(ctx) + path : path;
	return preferUploadWebp(raw);
}

/** PNG/JPEG → sibling .webp for /uploads/ assets (generated on API upload). */
export function preferUploadWebp(url: string): string {
	if (!url || !url.includes('/uploads/')) return url;
	return url.replace(/\.(png|jpe?g)(?=$|[?#])/i, '.webp');
}

/** Prefer WebP inside a single attribute value (supports srcset candidate lists). */
function preferUploadWebpInAttr(value: string): string {
	if (!value.includes('/uploads/')) return value;
	if (!value.includes(',')) return preferUploadWebp(value);
	return value
		.split(',')
		.map((part) => {
			const trimmed = part.trim();
			const space = trimmed.search(/\s/);
			if (space < 0) return preferUploadWebp(trimmed);
			return preferUploadWebp(trimmed.slice(0, space)) + trimmed.slice(space);
		})
		.join(', ');
}

/** Переписывает src/href/srcset на /uploads/... внутри готового HTML на абсолютные URL API. */
export function resolveUploadUrlsInHtml(html: string, ctx?: ApiRequestContext): string {
	if (!html.includes('/uploads/')) return html;
	const base = getApiBaseUrl(ctx);
	return html
		.replaceAll('src="/uploads/', `src="${base}/uploads/`)
		.replaceAll('href="/uploads/', `href="${base}/uploads/`)
		.replaceAll('srcset="/uploads/', `srcset="${base}/uploads/`)
		.replace(
			/(src|href|srcset)="([^"]*\/uploads\/[^"]+)"/gi,
			(_m, attr: string, value: string) => `${attr}="${preferUploadWebpInAttr(value)}"`
		);
}
