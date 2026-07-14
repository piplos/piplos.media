import { error, fail, isRedirect, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { projectPayload, projectSaveBody, projectSeoPayload, shouldSaveProjectSeo } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadServices, loadStack } from '$lib/lists.server';
import { loadSeoByPath, projectSeoPath } from '$lib/seo.server';
import type { Project } from '$lib/types';
import { projectHref, serviceUrlSlug } from '../../_projects';

export const load: PageServerLoad = async (event) => {
	const services = await loadServices(event);
	const serviceSlugs = new Set(services.map((s) => s.slug));
	const { serviceSlug, projectSlug } = event.params;

	try {
		const res = await fetchWithAuth(event, `/v1/projects/${projectSlug}`);
		if (res.status === 404) throw error(404, 'Проект не найден');
		if (!res.ok) throw error(res.status, 'Ошибка загрузки проекта');
		const data = (await res.json()) as { project: Project };
		const project = data.project;
		const canonicalServiceSlug = serviceUrlSlug(project.category, serviceSlugs);
		if (serviceSlug !== canonicalServiceSlug) {
			throw redirect(301, `/projects/${canonicalServiceSlug}/${project.slug}`);
		}

		const seoPath = projectSeoPath(project.slug);
		return {
			project,
			serviceSlug: canonicalServiceSlug,
			seo: await loadSeoByPath(event, seoPath),
			languages: await loadLanguages(event),
			services,
			stack: await loadStack(event)
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		throw e;
	}
};

export const actions: Actions = {
	save: async (event) => {
		const services = await loadServices(event);
		const serviceSlugs = new Set(services.map((s) => s.slug));
		const fd = await event.request.formData();
		const payload = projectPayload(fd);
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });
		if (!payload.category) return fail(400, { error: 'Выберите группу (услугу)' });

		const existingRes = await fetchWithAuth(event, `/v1/projects/${event.params.projectSlug}`);
		if (!existingRes.ok) {
			return fail(existingRes.status, { error: 'Проект не найден' });
		}
		const existing = ((await existingRes.json()) as { project: Project }).project;

		const body = projectSaveBody(payload, { existing, services });
		if (!body) {
			return fail(400, { error: 'Выберите услугу (группу) или создайте её в разделе «Услуги»' });
		}

		const res = await fetchWithAuth(event, `/v1/projects/${event.params.projectSlug}`, {
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
				seo.id ? `/v1/seo/${seo.id}` : '/v1/seo',
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

		const nextServiceSlug = serviceUrlSlug(payload.category, serviceSlugs);
		throw redirect(303, `/projects/${nextServiceSlug}/${payload.slug}`);
	}
};
