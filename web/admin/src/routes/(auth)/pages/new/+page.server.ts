import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { pagePayload } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadStack } from '$lib/lists.server';
import { saveSeoFromForm } from '$lib/seo.server';
import type { Page } from '$lib/types';

export const load: PageServerLoad = async (event) => ({
	languages: await loadLanguages(event),
	stack: await loadStack(event)
});

export const actions: Actions = {
	save: async (event) => {
		const fd = await event.request.formData();
		const payload = pagePayload(fd);
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });

		const res = await fetchWithAuth(event, '/v1/pages', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось создать страницу' });
		}
		const data = (await res.json()) as { page: Page };

		// SEO сохраняем после создания (путь /articles/{slug}); ошибка SEO не блокирует редирект.
		await saveSeoFromForm(event, fd);

		throw redirect(303, `/pages/${data.page.id}`);
	}
};
