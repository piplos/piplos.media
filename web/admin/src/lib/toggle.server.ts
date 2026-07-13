/** Переключение published / статусов из списков админки. */
import { fail, type ActionFailure } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import type { Project, Service, StackItem } from '$lib/types';

type ToggleResult = ActionFailure<{ error: string }> | { ok: true };

export async function toggleProjectPublished(event: RequestEvent, slug: string): Promise<ToggleResult> {
	const getRes = await fetchWithAuth(event, `/api/v1/projects/${slug}`);
	if (!getRes.ok) return fail(getRes.status, { error: 'Проект не найден' });

	const { project } = (await getRes.json()) as { project: Project };
	const res = await fetchWithAuth(event, `/api/v1/projects/${slug}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			slug: project.slug,
			category: project.category,
			categories: project.categories?.length ? project.categories : [project.category],
			tags: project.tags ?? [],
			year: project.year,
			featured: project.featured,
			published: !project.published,
			sort_order: project.sort_order,
			translations: project.translations ?? {}
		})
	});
	if (!res.ok) {
		const data = (await res.json().catch(() => ({}))) as { message?: string };
		return fail(res.status, { error: data.message ?? 'Не удалось обновить статус' });
	}
	return { ok: true };
}

export async function toggleServicePublished(event: RequestEvent, id: string): Promise<ToggleResult> {
	const listRes = await fetchWithAuth(event, '/api/v1/services');
	if (!listRes.ok) return fail(listRes.status, { error: 'Не удалось загрузить услуги' });

	const { services } = (await listRes.json()) as { services: Service[] };
	const service = services.find((item) => item.id === id);
	if (!service) return fail(404, { error: 'Услуга не найдена' });

	const res = await fetchWithAuth(event, `/api/v1/services/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			slug: service.slug,
			icon: service.icon,
			tags: service.tags ?? [],
			published: !service.published,
			sort_order: service.sort_order,
			translations: service.translations ?? {}
		})
	});
	if (!res.ok) {
		const data = (await res.json().catch(() => ({}))) as { message?: string };
		return fail(res.status, { error: data.message ?? 'Не удалось обновить статус' });
	}
	return { ok: true };
}

export async function toggleStackPublished(event: RequestEvent, id: string): Promise<ToggleResult> {
	const listRes = await fetchWithAuth(event, '/api/v1/stack');
	if (!listRes.ok) return fail(listRes.status, { error: 'Не удалось загрузить стек' });

	const { stack } = (await listRes.json()) as { stack: StackItem[] };
	const item = stack.find((entry) => entry.id === id);
	if (!item) return fail(404, { error: 'Технология не найдена' });

	const res = await fetchWithAuth(event, `/api/v1/stack/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			slug: item.slug,
			label: item.label,
			group_id: item.group_id,
			published: !item.published,
			sort_order: item.sort_order
		})
	});
	if (!res.ok) {
		const data = (await res.json().catch(() => ({}))) as { message?: string };
		return fail(res.status, { error: data.message ?? 'Не удалось обновить статус' });
	}
	return { ok: true };
}
