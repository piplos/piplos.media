import { error, fail, isRedirect, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { projectPayload, projectSeoPayload, shouldSaveProjectSeo } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadServices, loadStack } from '$lib/lists.server';
import { loadSeoByPath, projectSeoPath } from '$lib/seo.server';
import type { Project } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, `/api/v1/projects/${event.params.slug}`);
		if (res.status === 404) throw error(404, 'Проект не найден');
		if (!res.ok) throw error(res.status, 'Ошибка загрузки проекта');
		const data = (await res.json()) as { project: Project };
		const seoPath = projectSeoPath(data.project.slug);
		return {
			project: data.project,
			seo: await loadSeoByPath(event, seoPath),
			languages: await loadLanguages(event),
			services: await loadServices(event),
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
		const payload = projectPayload(fd);
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });
		if (!payload.category) return fail(400, { error: 'Выберите группу (услугу)' });

		const existingRes = await fetchWithAuth(event, `/api/v1/projects/${event.params.slug}`);
		if (!existingRes.ok) {
			return fail(existingRes.status, { error: 'Проект не найден' });
		}
		const existing = ((await existingRes.json()) as { project: Project }).project;

		const body = projectSaveBody(payload, {
			existing,
			services: await loadServices(event)
		});
		if (!body) {
			return fail(400, { error: 'Выберите услугу (группу) или создайте её в «Списки → Услуги»' });
		}

		const res = await fetchWithAuth(event, `/api/v1/projects/${event.params.slug}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить проект' });
		}

		const seo = projectSeoPayload(fd);
		if (seo && shouldSaveProjectSeo(seo)) {
			const seoRes = await fetchWithAuth(
				event,
				seo.id ? `/api/v1/seo/${seo.id}` : '/api/v1/seo',
				{
					method: seo.id ? 'PUT' : 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ path: seo.path, translations: seo.translations })
				}
			);
			if (!seoRes.ok) {
				const data = (await seoRes.json().catch(() => ({}))) as { message?: string };
				return fail(seoRes.status, { error: data.message ?? 'Проект сохранён, но не удалось сохранить SEO' });
			}
		}

		throw redirect(303, `/projects/${payload.slug}`);
	}
};
