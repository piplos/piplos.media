import { isRedirect, type RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import type { SEOPage } from '$lib/types';

export async function loadSeoByPath(event: RequestEvent, path: string): Promise<SEOPage | null> {
	try {
		const res = await fetchWithAuth(event, '/api/v1/seo');
		if (!res.ok) return null;
		const data = (await res.json()) as { pages: SEOPage[] };
		return (data.pages ?? []).find((page) => page.path === path) ?? null;
	} catch (e) {
		if (isRedirect(e)) throw e;
		return null;
	}
}

export function projectSeoPath(projectSlug: string): string {
	return `/portfolio/${projectSlug}`;
}
