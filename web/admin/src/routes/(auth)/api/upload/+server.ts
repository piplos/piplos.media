import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { fetchWithAuth } from '$lib/api.server';

/** Прокси загрузки медиа: Bearer в httpOnly cookie, multipart пересылается в Go API. */
export const POST: RequestHandler = async (event) => {
	const form = await event.request.formData();
	const file = form.get('file');
	if (!(file instanceof File)) {
		return json({ message: 'file is required' }, { status: 400 });
	}

	const body = new FormData();
	body.append('file', file, file.name);
	const path = form.get('path');
	if (typeof path === 'string' && path) body.append('path', path);
	const name = form.get('name');
	if (typeof name === 'string' && name) body.append('name', name);

	const res = await fetchWithAuth(event, '/v1/uploads', {
		method: 'POST',
		body
	});
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};
