import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getApiBaseUrl } from '$lib/env.server';

/** Файлы архива хранятся относительными путями (/uploads/...) —
 *  превью в админке редиректим на файловый сервер API. */
export const GET: RequestHandler = (event) => {
	redirect(302, `${getApiBaseUrl(event)}/uploads/${event.params.path}`);
};
