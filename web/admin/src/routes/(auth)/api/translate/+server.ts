import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { fetchWithAuth } from '$lib/api.server';

/** Прокси к API AI-перевода: браузер не имеет Bearer-токена (он в httpOnly cookie). */
export const POST: RequestHandler = async (event) => {
	const body = await event.request.text();
	const res = await fetchWithAuth(event, '/api/v1/translate', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body
	});
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};
