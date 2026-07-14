import { error, fail, isRedirect, redirect } from '@sveltejs/kit';
import type { Actions, RequestEvent, ServerLoad } from '@sveltejs/kit';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { servicePayload } from '$lib/content.server';
import { loadLanguages } from '$lib/languages.server';
import { loadServices, loadStack } from '$lib/lists.server';
import { toggleServicePublished } from '$lib/toggle.server';
import type { Service } from '$lib/types';
import { findServiceBySlug } from './_services';

export async function loadServicesPageData(event: RequestEvent) {
	try {
		const res = await fetchWithAuth(event, '/v1/services');
		if (!res.ok) {
			return {
				services: [] as Service[],
				languages: await loadLanguages(event),
				stack: await loadStack(event),
				error: apiLoadErrorMessage(res, 'Ошибка загрузки услуг')
			};
		}
		const data = (await res.json()) as { services: Service[] };
		return {
			services: data.services ?? [],
			languages: await loadLanguages(event),
			stack: await loadStack(event),
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return {
			services: [] as Service[],
			languages: await loadLanguages(event),
			stack: await loadStack(event),
			error: 'API недоступен'
		};
	}
}

export const createServicesListLoad = (): ServerLoad => async () => ({
	activeSlug: '' as const
});

export const createServiceEditLoad = (): ServerLoad => async (event) => {
	const parentData = await event.parent();
	const service = findServiceBySlug(parentData.services, event.params.slug ?? '');
	if (!service) throw error(404, 'Услуга не найдена');
	return { service, activeSlug: service.slug };
};

export const createServiceNewLoad = (): ServerLoad => async () => ({
	activeSlug: '__new__' as const
});

export const servicesActions: Actions = {
	togglePublished: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });
		return toggleServicePublished(event, id);
	},
	save: async (event) => {
		const fd = await event.request.formData();
		const id = fd.get('id')?.toString() ?? '';
		const payload = servicePayload(fd);
		if (!payload.slug) return fail(400, { error: 'Укажите slug' });

		const res = await fetchWithAuth(event, id ? `/v1/services/${id}` : '/v1/services', {
			method: id ? 'PUT' : 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить услугу' });
		}
		throw redirect(303, `/services/${payload.slug}`);
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/services/${id}`, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'Не удалось удалить услугу' });
		throw redirect(303, '/services');
	},
	reorder: async (event) => {
		const orderRaw = (await event.request.formData()).get('order')?.toString() ?? '';
		let ids: string[];
		try {
			ids = JSON.parse(orderRaw) as string[];
			if (!Array.isArray(ids) || ids.some((id) => typeof id !== 'string' || !id)) {
				throw new Error('invalid');
			}
		} catch {
			return fail(400, { error: 'Некорректный порядок сортировки' });
		}

		const res = await fetchWithAuth(event, '/v1/services/reorder', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids })
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
