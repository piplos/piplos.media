import { isRedirect, type RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import { seoPayload, shouldSaveSeo } from '$lib/content.server';
import type { SEOPage } from '$lib/types';

export async function loadSeoByPath(event: RequestEvent, path: string): Promise<SEOPage | null> {
	try {
		const res = await fetchWithAuth(event, '/v1/seo');
		if (!res.ok) return null;
		const data = (await res.json()) as { pages: SEOPage[] };
		return (data.pages ?? []).find((page) => page.path === path) ?? null;
	} catch (e) {
		if (isRedirect(e)) throw e;
		return null;
	}
}

/** Сохраняет SEO из скрытых полей формы (seo_id/seo_path/seo_translations).
 *  Возвращает ошибку API или null (успех либо нечего сохранять). */
export async function saveSeoFromForm(
	event: RequestEvent,
	fd: FormData
): Promise<{ status: number; message: string } | null> {
	const seo = seoPayload(fd);
	if (!seo || !shouldSaveSeo(seo)) return null;

	const res = await fetchWithAuth(event, seo.id ? `/v1/seo/${seo.id}` : '/v1/seo', {
		method: seo.id ? 'PUT' : 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ path: seo.path, translations: seo.translations })
	});
	if (!res.ok) {
		const data = (await res.json().catch(() => ({}))) as { message?: string };
		return { status: res.status, message: data.message ?? '' };
	}
	return null;
}

export function projectSeoPath(projectSlug: string): string {
	return `/portfolio/${projectSlug}`;
}

export function serviceSeoPath(serviceSlug: string): string {
	return `/services/${serviceSlug}`;
}

export function articleSeoPath(pageSlug: string): string {
	return `/articles/${pageSlug}`;
}
