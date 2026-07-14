import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { categoryFromUrlSlug, projectHref } from '../_projects';
import { createProjectsLoad, projectsActions } from '../_projects.server';
import { fetchWithAuth } from '$lib/api.server';
import { loadServices } from '$lib/lists.server';
import type { Project } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	const services = await loadServices(event);
	const serviceSlugs = new Set(services.map((s) => s.slug));
	const paramSlug = event.params.serviceSlug;
	const category = categoryFromUrlSlug(paramSlug, serviceSlugs);

	if (category === null) {
		const res = await fetchWithAuth(event, `/v1/projects/${paramSlug}`);
		if (res.ok) {
			const data = (await res.json()) as { project: Project };
			throw redirect(301, projectHref(data.project, services));
		}
		throw error(404, 'Группа не найдена');
	}

	return createProjectsLoad(category)(event);
};

export const actions = projectsActions;
