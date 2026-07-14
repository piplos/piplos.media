import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { fetchWithAuth, apiLoadErrorMessage } from '$lib/api.server';
import type { FileListing } from '$lib/files';

export const load: PageServerLoad = async (event) => {
	const path = event.url.searchParams.get('path') ?? '';
	const res = await fetchWithAuth(event, `/v1/files?path=${encodeURIComponent(path)}`);
	if (res.status === 404) throw error(404, 'Папка не найдена');
	if (!res.ok) throw error(res.status, apiLoadErrorMessage(res, 'Ошибка загрузки файлов'));
	return { listing: (await res.json()) as FileListing };
};
