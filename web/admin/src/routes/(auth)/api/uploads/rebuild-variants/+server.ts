import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { fetchWithAuth } from '$lib/api.server';

/** Прокси статуса пересборки WebP (Bearer в httpOnly cookie). */
export const GET: RequestHandler = async (event) => {
	const res = await fetchWithAuth(event, '/v1/uploads/rebuild-variants');
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};
