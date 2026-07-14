import { getApiV1, type ApiRequestContext } from '$lib/api';

export interface SEOLocale {
	title?: string;
	description?: string;
	og_title?: string;
	og_description?: string;
	og_image?: string;
}

export type SEOTranslations = Record<string, SEOLocale>;

/** SEO-запись страницы по её path (например, /portfolio/aifront) или null. */
export async function fetchSEOPage(
	path: string,
	fetchFn: typeof fetch = fetch,
	ctx?: ApiRequestContext
): Promise<SEOTranslations | null> {
	try {
		const res = await fetchFn(`${getApiV1(ctx)}/public/seo?path=${encodeURIComponent(path)}`);
		if (!res.ok) return null;
		const data = (await res.json()) as { page: { translations?: SEOTranslations } | null };
		return data.page?.translations ?? null;
	} catch {
		return null;
	}
}
