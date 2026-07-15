import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, updateSetting, type SettingsListResponse } from '$lib/settings.server';

export interface LeadEmailTemplate {
	subject: string;
	body: string;
}

const defaults: LeadEmailTemplate = { subject: '', body: '' };

function parseTemplate(raw: string): LeadEmailTemplate {
	if (!raw) return { ...defaults };
	try {
		const obj = JSON.parse(raw) as Record<string, unknown>;
		return {
			subject: (obj.subject as string) ?? '',
			body: (obj.body as string) ?? ''
		};
	} catch {
		return { ...defaults };
	}
}

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/settings');
		if (!res.ok) {
			return { template: { ...defaults }, error: apiLoadErrorMessage(res, 'Ошибка загрузки настроек') };
		}
		const data = (await res.json()) as SettingsListResponse;
		return {
			template: parseTemplate(getByKey(data.settings ?? [], 'LEAD_EMAIL_TEMPLATE')),
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { template: { ...defaults }, error: 'API недоступен' };
	}
};

export const actions: Actions = {
	updateTemplate: async (event) => {
		try {
			const fd = await event.request.formData();
			const payload: LeadEmailTemplate = {
				subject: fd.get('subject')?.toString() ?? '',
				body: fd.get('body')?.toString() ?? ''
			};
			const res = await updateSetting(event, 'LEAD_EMAIL_TEMPLATE', JSON.stringify(payload));
			if (!res.ok) {
				const data = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, { error: data.message ?? 'Не удалось сохранить шаблон' });
			}
			return { ok: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { error: 'Не удалось сохранить шаблон' });
		}
	}
};
