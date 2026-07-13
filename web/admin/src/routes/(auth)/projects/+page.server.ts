import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { loadServices } from '$lib/lists.server';
import { toggleProjectPublished } from '$lib/toggle.server';
import type { Project, Service } from '$lib/types';

type ProjectLayoutGroup = { group_id: string; ids: string[] };

export const load: PageServerLoad = async (event) => {
	try {
		const [res, services] = await Promise.all([
			fetchWithAuth(event, '/api/v1/projects'),
			loadServices(event)
		]);
		if (!res.ok) {
			return {
				projects: [],
				services,
				error: apiLoadErrorMessage(res, 'Ошибка загрузки проектов')
			};
		}
		const data = (await res.json()) as { projects: Project[] };
		return { projects: data.projects ?? [], services, error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { projects: [], services: await loadServices(event), error: 'API недоступен' };
	}
};

export const actions: Actions = {
	togglePublished: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });
		return toggleProjectPublished(event, id);
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/api/v1/projects/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить проект' });
		return { ok: true };
	},
	reorder: async (event) => {
		const layoutRaw = (await event.request.formData()).get('layout')?.toString() ?? '';
		let groups: ProjectLayoutGroup[];
		try {
			groups = JSON.parse(layoutRaw) as ProjectLayoutGroup[];
			if (
				!Array.isArray(groups) ||
				groups.some(
					(group) =>
						typeof group !== 'object' ||
						typeof group.group_id !== 'string' ||
						!group.group_id ||
						!Array.isArray(group.ids) ||
						group.ids.some((id) => typeof id !== 'string' || !id)
				)
			) {
				throw new Error('invalid');
			}
		} catch {
			return fail(400, { error: 'Некорректная раскладка проектов' });
		}

		const services = await loadServices(event);
		const serviceSlugs = new Set(services.map((service: Service) => service.slug));
		if (groups.some((group) => !serviceSlugs.has(group.group_id))) {
			return fail(400, { error: 'Группа должна соответствовать услуге из списка' });
		}

		const res = await fetchWithAuth(event, '/api/v1/projects/reorder', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ groups })
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, {
				error: data.message ?? res.statusText ?? 'Не удалось сохранить порядок'
			});
		}
		return { ok: true };
	}
};
