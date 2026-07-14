import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import type { AdminUser } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/users');
		if (!res.ok) {
			return { users: [], error: apiLoadErrorMessage(res, 'Ошибка загрузки пользователей') };
		}
		const data = (await res.json()) as { users: AdminUser[] };
		return { users: data.users ?? [], error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { users: [], error: 'API недоступен' };
	}
};

export const actions: Actions = {
	save: async (event) => {
		const fd = await event.request.formData();
		const id = fd.get('id')?.toString() ?? '';
		const payload = {
			email: fd.get('email')?.toString().trim() ?? '',
			password: fd.get('password')?.toString() ?? '',
			full_name: fd.get('full_name')?.toString().trim() ?? '',
			role: fd.get('role')?.toString() ?? 'manager',
			is_active: fd.get('is_active') === 'on'
		};
		if (!id && !payload.email) return fail(400, { error: 'Укажите email' });
		if (!id && payload.password.length < 8) {
			return fail(400, { error: 'Пароль — минимум 8 символов' });
		}

		const res = await fetchWithAuth(event, id ? `/v1/users/${id}` : '/v1/users', {
			method: id ? 'PUT' : 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось сохранить пользователя' });
		}
		return { ok: true };
	},
	delete: async (event) => {
		const id = (await event.request.formData()).get('id')?.toString();
		if (!id) return fail(400, { error: 'Некорректный запрос' });

		const res = await fetchWithAuth(event, `/v1/users/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { error: data.message ?? 'Не удалось удалить пользователя' });
		}
		return { ok: true };
	}
};
