import {
	fail,
	isRedirect,
	redirect,
	type Actions,
	type RequestEvent,
	type ServerLoad
} from '@sveltejs/kit';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { loadLanguages } from '$lib/languages.server';
import { togglePagePublished } from '$lib/toggle.server';
import type { LegalPage, Page } from '$lib/types';

export type PagesSection = 'articles' | 'legal';

/** Статей на одной странице списка. */
const PAGE_SIZE = 20;

export function createPagesLoad(section: PagesSection): ServerLoad {
	return async (event: RequestEvent) => {
		// Math.trunc отсекает дробные ?page=1.5 — иначе offset считается от дробной страницы.
		const page = Math.max(1, Math.trunc(Number(event.url.searchParams.get('page') ?? '1')) || 1);
		const q = new URLSearchParams({ limit: String(PAGE_SIZE), offset: String((page - 1) * PAGE_SIZE) });
		try {
			const [pagesRes, legalRes] = await Promise.all([
				fetchWithAuth(event, `/v1/pages?${q}`),
				fetchWithAuth(event, '/v1/legal')
			]);
			if (!pagesRes.ok || !legalRes.ok) {
				const bad = !pagesRes.ok ? pagesRes : legalRes;
				return {
					section,
					pages: [] as Page[],
					legalPages: [] as LegalPage[],
					languages: await loadLanguages(event),
					page,
					total: 0,
					totalPages: 1,
					error: apiLoadErrorMessage(bad, 'Ошибка загрузки страниц')
				};
			}
			const pagesData = (await pagesRes.json()) as { pages: Page[]; total: number };
			const legalData = (await legalRes.json()) as { pages: LegalPage[] };
			const total = pagesData.total ?? 0;
			const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
			// После удаления последней статьи на последней странице она опустела —
			// возвращаем на последнюю существующую страницу.
			if (section === 'articles' && page > totalPages) {
				redirect(307, totalPages > 1 ? `/pages?page=${totalPages}` : '/pages');
			}
			return {
				section,
				pages: pagesData.pages ?? [],
				legalPages: legalData.pages ?? [],
				languages: await loadLanguages(event),
				page,
				total,
				totalPages,
				error: null
			};
		} catch (e) {
			if (isRedirect(e)) throw e;
			return {
				section,
				pages: [] as Page[],
				legalPages: [] as LegalPage[],
				languages: await loadLanguages(event),
				page,
				total: 0,
				totalPages: 1,
				error: 'API недоступен'
			};
		}
	};
}

export const pagesActions: Actions = {
	togglePublished: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });
		return togglePagePublished(event, id);
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/pages/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить страницу' });
		return { ok: true };
	}
};
