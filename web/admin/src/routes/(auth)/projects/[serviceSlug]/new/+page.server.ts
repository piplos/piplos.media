import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { fetchWithAuth } from '$lib/api.server';
import { projectPayload, projectSaveBody } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadServices, loadStack } from '$lib/lists.server';
import type { Project } from '$lib/types';
import { ORPHAN_GROUP_ID, categoryFromUrlSlug, serviceUrlSlug } from '../../_projects';

export const load: PageServerLoad = async (event) => {
	const services = await loadServices(event);
	const serviceSlugs = new Set(services.map((s) => s.slug));
	const category = categoryFromUrlSlug(event.params.serviceSlug, serviceSlugs);
	if (category === null || category === ORPHAN_GROUP_ID) {
		throw error(404, 'Группа не найдена');
	}
	return {
		serviceSlug: category,
		languages: await loadLanguages(event),
		services,
		stack: await loadStack(event)
	};
};

export const actions: Actions = {
	save: async (event) => {
		const services = await loadServices(event);
		const serviceSlugs = new Set(services.map((s) => s.slug));
		const payload = projectPayload(await event.request.formData());
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });
		if (!payload.category) return fail(400, { error: 'Выберите группу (услугу)' });

		const projectsRes = await fetchWithAuth(event, '/v1/projects');
		const projects = projectsRes.ok
			? (((await projectsRes.json()) as { projects: Project[] }).projects ?? [])
			: [];

		const body = projectSaveBody(payload, { services, projects });
		if (!body) {
			return fail(400, { error: 'Выберите услугу (группу) или создайте её в разделе «Услуги»' });
		}

		const res = await fetchWithAuth(event, '/v1/projects', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось создать проект' });
		}

		const nextServiceSlug = serviceUrlSlug(payload.category, serviceSlugs);
		throw redirect(303, `/projects/${nextServiceSlug}/${payload.slug}`);
	}
};
