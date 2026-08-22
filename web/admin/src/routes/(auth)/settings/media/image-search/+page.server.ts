import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';

export interface ImageRefUsage {
	entity: string;
	id: string;
	label: string;
	field: string;
	/** Ссылка на редактирование сущности; пусто, если отдельной страницы нет (SEO). */
	href?: string;
}

export interface ImageFileRef {
	path: string;
	webp_path: string;
	webp_exists: boolean;
	usages: ImageRefUsage[];
}

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/media/image-references');
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return {
				files: [] as ImageFileRef[],
				error: data.message ?? apiLoadErrorMessage(res, 'Не удалось выполнить поиск PNG/JPG')
			};
		}
		const data = (await res.json()) as { files?: ImageFileRef[] };
		return { files: data.files ?? [], error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { files: [] as ImageFileRef[], error: 'API недоступен' };
	}
};

const replaceActionError = (resStatus: number, data: unknown, fallback: string) =>
	fail(resStatus, {
		error:
			((data as { message?: string; error?: string }).message ??
				(data as { error?: string }).error) ||
		fallback
	});

export const actions: Actions = {
	replace: async (event) => {
		const formData = await event.request.formData();
		const path = String(formData.get('path') ?? '');
		if (!path) {
			return fail(400, { error: 'Не указан файл' });
		}
		try {
			const res = await fetchWithAuth(event, '/v1/media/image-references/replace', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ path })
			});
			const data = (await res.json().catch(() => ({}))) as { updated?: number };
			if (!res.ok) return replaceActionError(res.status, data, 'Не удалось заменить ссылки');
			return { ok: true, updated: data.updated ?? 0 };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { error: 'API недоступен' });
		}
	},
	replaceAll: async (event) => {
		try {
			const res = await fetchWithAuth(event, '/v1/media/image-references/replace-all', {
				method: 'POST'
			});
			const data = (await res.json().catch(() => ({}))) as { updated?: number; files?: number };
			if (!res.ok) return replaceActionError(res.status, data, 'Не удалось заменить ссылки');
			return { ok: true, updated: data.updated ?? 0, files: data.files ?? 0 };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { error: 'API недоступен' });
		}
	}
};
