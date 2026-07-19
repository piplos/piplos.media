import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { fetchWithAuth } from '$lib/api.server';

/** Прокси бекапов: статус для поллинга и скачивание архива (Bearer в httpOnly cookie). */
export const GET: RequestHandler = async (event) => {
	const action = event.url.searchParams.get('action') ?? 'status';

	if (action === 'download') {
		const storage = event.url.searchParams.get('storage') ?? '';
		const name = event.url.searchParams.get('name') ?? '';
		const res = await fetchWithAuth(
			event,
			`/v1/backups/download?storage=${encodeURIComponent(storage)}&name=${encodeURIComponent(name)}`
		);
		if (!res.ok) {
			const data = await res.json().catch(() => ({ message: 'Не удалось скачать архив' }));
			return json(data, { status: res.status });
		}
		return new Response(res.body, {
			status: res.status,
			headers: {
				'Content-Type': res.headers.get('Content-Type') ?? 'application/gzip',
				'Content-Disposition':
					res.headers.get('Content-Disposition') ?? `attachment; filename="${name}"`
			}
		});
	}

	const res = await fetchWithAuth(event, '/v1/backups/status');
	const data = await res.json().catch(() => ({ message: 'Некорректный ответ API' }));
	return json(data, { status: res.status });
};
