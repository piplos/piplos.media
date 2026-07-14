import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { fetchWithAuth } from '$lib/api.server';

/** Прокси файлового архива: Bearer в httpOnly cookie, запросы пересылаются в Go API. */
export const GET: RequestHandler = async (event) => {
	const path = event.url.searchParams.get('path') ?? '';
	const res = await fetchWithAuth(event, `/v1/files?path=${encodeURIComponent(path)}`);
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};

const ACTIONS = new Set(['folders', 'rename', 'move', 'delete']);

export const POST: RequestHandler = async (event) => {
	const action = event.url.searchParams.get('action') ?? '';
	if (!ACTIONS.has(action)) {
		return json({ message: 'Неизвестное действие' }, { status: 400 });
	}
	const body = await event.request.text();
	const res = await fetchWithAuth(event, `/v1/files/${action}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body
	});
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};
