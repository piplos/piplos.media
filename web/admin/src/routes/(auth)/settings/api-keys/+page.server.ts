import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import type { APIKey } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/api-keys');
		if (!res.ok) {
			return { apiKeys: [], error: apiLoadErrorMessage(res, 'Ошибка загрузки ключей') };
		}
		const data = (await res.json()) as { api_keys: APIKey[] };
		return { apiKeys: data.api_keys ?? [], error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { apiKeys: [], error: 'API недоступен' };
	}
};

export const actions: Actions = {
	// Создание ключа. Сырой ключ показывается один раз — возвращаем его
	// в результате экшена, страница показывает его в диалоге.
	create: async (event) => {
		const fd = await event.request.formData();
		const name = fd.get('name')?.toString().trim() ?? '';
		if (!name) return fail(400, { error: 'Укажите название ключа' });
		// API считает длину в байтах (len в Go) — проверяем так же, чтобы
		// кириллические названия не падали на сервере с английской ошибкой.
		if (new TextEncoder().encode(name).length > 100) {
			return fail(400, { error: 'Название — максимум 100 символов' });
		}

		const res = await fetchWithAuth(event, '/v1/api-keys', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось создать ключ' });
		}
		const data = (await res.json().catch(() => ({}))) as { key?: string };
		if (!data.key) return fail(502, { error: 'Не удалось создать ключ' });
		return { createdKey: data.key };
	},
	revoke: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/api-keys/${id}/revoke`, { method: 'POST' });
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось отозвать ключ' });
		}
		return { ok: true };
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/api-keys/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось удалить ключ' });
		}
		return { ok: true };
	}
};
