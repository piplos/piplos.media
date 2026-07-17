import { error, fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { pagePayload } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadStack } from '$lib/lists.server';
import { articleSeoPath, loadSeoByPath, saveSeoFromForm } from '$lib/seo.server';
import type { Page } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, `/v1/pages/${event.params.id}`);
		if (res.status === 404) throw error(404, 'Страница не найдена');
		if (!res.ok) throw error(res.status, 'Ошибка загрузки страницы');
		const data = (await res.json()) as { page: Page };
		return {
			page: data.page,
			seo: await loadSeoByPath(event, articleSeoPath(data.page.slug)),
			languages: await loadLanguages(event),
			stack: await loadStack(event)
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		throw e;
	}
};

export const actions: Actions = {
	save: async (event) => {
		const fd = await event.request.formData();
		const payload = pagePayload(fd);
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });

		const res = await fetchWithAuth(event, `/v1/pages/${event.params.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить страницу' });
		}

		const seoError = await saveSeoFromForm(event, fd);
		if (seoError) {
			return fail(seoError.status, {
				error: seoError.message || 'Страница сохранена, но не удалось сохранить SEO'
			});
		}
		return { ok: true };
	}
};
