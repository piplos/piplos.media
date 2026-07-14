import { API_URL } from '$lib/api';

export interface LegalSection {
	title: string;
	body: string;
}

export interface LegalLocale {
	label: string;
	title: string;
	last_updated: string;
	sections: LegalSection[];
}

export interface LegalPage {
	id: string;
	slug: string;
	path: string;
	sort_order: number;
	translations: Record<string, LegalLocale>;
}

export const LEGAL_SLUGS = ['privacy', 'terms', 'cookies'] as const;

export type LegalSlug = (typeof LEGAL_SLUGS)[number];

export function isLegalSlug(value: string): value is LegalSlug {
	return (LEGAL_SLUGS as readonly string[]).includes(value);
}

type FetchFn = typeof fetch;

export async function fetchLegalPages(fetchFn: FetchFn = fetch, lang?: string): Promise<LegalPage[]> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${API_URL}/api/v1/public/legal${qs}`);
		if (!res.ok) return [];
		const data = (await res.json()) as { pages: LegalPage[] };
		return (data.pages ?? []).sort((a, b) => a.sort_order - b.sort_order);
	} catch {
		return [];
	}
}

/** Один правовой документ по slug или null (lang — только этот перевод). */
export async function fetchLegalPage(
	slug: LegalSlug,
	fetchFn: FetchFn = fetch,
	lang?: string
): Promise<LegalPage | null> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${API_URL}/api/v1/public/legal/${slug}${qs}`);
		if (!res.ok) return null;
		const data = (await res.json()) as { page: LegalPage };
		return data.page ?? null;
	} catch {
		return null;
	}
}
