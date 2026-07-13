import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { stackPayload } from '$lib/content.server';
import { loadServices } from '$lib/lists.server';
import { toggleStackPublished } from '$lib/toggle.server';
import type { Service, StackItem } from '$lib/types';

type StackLayoutGroup = { group_id: string; ids: string[] };

function firstServiceSlug(services: Service[]): string | null {
	const sorted = [...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug));
	return sorted[0]?.slug ?? null;
}

function nextSortOrder(stack: StackItem[], groupId: string): number {
	const inGroup = stack.filter((item) => item.group_id === groupId);
	if (!inGroup.length) return 0;
	return Math.max(...inGroup.map((item) => item.sort_order)) + 1;
}

export const load: PageServerLoad = async (event) => {
	try {
		const [res, services] = await Promise.all([
			fetchWithAuth(event, '/api/v1/stack'),
			loadServices(event)
		]);
		if (!res.ok) {
			return {
				stack: [],
				services,
				error: apiLoadErrorMessage(res, 'Ошибка загрузки стека')
			};
		}
		const data = (await res.json()) as { stack: StackItem[] };
		return { stack: data.stack ?? [], services, error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { stack: [], services: await loadServices(event), error: 'API недоступен' };
	}
};

export const actions: Actions = {
	togglePublished: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });
		return toggleStackPublished(event, id);
	},
	save: async (event) => {
		const fd = await event.request.formData();
		const id = fd.get('id')?.toString() ?? '';
		const payload = stackPayload(fd);
		if (!payload.slug || !payload.label) {
			return fail(400, { error: 'Укажите slug и название' });
		}

		const services = await loadServices(event);
		const defaultGroup = firstServiceSlug(services);
		if (!defaultGroup) {
			return fail(400, { error: 'Сначала создайте услугу в разделе «Услуги»' });
		}

		const stackRes = await fetchWithAuth(event, '/api/v1/stack');
		const stackData = stackRes.ok
			? ((await stackRes.json()) as { stack: StackItem[] })
			: { stack: [] as StackItem[] };
		const stack = stackData.stack ?? [];
		const existing = id ? stack.find((item) => item.id === id) : undefined;

		const body = {
			...payload,
			group_id: existing?.group_id ?? defaultGroup,
			sort_order: existing?.sort_order ?? nextSortOrder(stack, defaultGroup)
		};

		const res = await fetchWithAuth(event, id ? `/api/v1/stack/${id}` : '/api/v1/stack', {
			method: id ? 'PUT' : 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить технологию' });
		}
		return { ok: true };
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/api/v1/stack/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить технологию' });
		return { ok: true };
	},
	reorder: async (event) => {
		const layoutRaw = (await event.request.formData()).get('layout')?.toString() ?? '';
		let groups: StackLayoutGroup[];
		try {
			groups = JSON.parse(layoutRaw) as StackLayoutGroup[];
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
			return fail(400, { error: 'Некорректная раскладка стека' });
		}

		const services = await loadServices(event);
		const serviceSlugs = new Set(services.map((service) => service.slug));
		if (groups.some((group) => !serviceSlugs.has(group.group_id))) {
			return fail(400, { error: 'Группа должна соответствовать услуге из списка' });
		}

		const res = await fetchWithAuth(event, '/api/v1/stack/reorder', {
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
