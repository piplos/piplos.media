import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, MASKED, updateSetting, type SettingsListResponse } from '$lib/settings.server';

export interface SmtpSettings {
	host: string;
	port: string;
	username: string;
	usernameMasked: boolean;
	password: string;
	passwordMasked: boolean;
	from: string;
	timeoutSeconds: string;
}

const defaults: SmtpSettings = {
	host: '',
	port: '587',
	username: '',
	usernameMasked: false,
	password: '',
	passwordMasked: false,
	from: '',
	timeoutSeconds: '30'
};

function parseSmtp(raw: string): SmtpSettings {
	if (!raw) return { ...defaults };
	try {
		const obj = JSON.parse(raw) as Record<string, unknown>;
		const username = (obj.username as string) ?? '';
		const password = (obj.password as string) ?? '';
		return {
			host: (obj.host as string) ?? '',
			port: String(obj.port ?? 587),
			username: username === MASKED ? '' : username,
			usernameMasked: username === MASKED,
			password: password === MASKED ? '' : password,
			passwordMasked: password === MASKED,
			from: (obj.from as string) ?? '',
			timeoutSeconds: String(obj.timeout_seconds ?? 30)
		};
	} catch {
		return { ...defaults };
	}
}

function smtpPayloadFromForm(fd: FormData): Record<string, unknown> {
	const port = Math.max(1, Math.min(65535, parseInt(fd.get('port')?.toString() ?? '587', 10) || 587));
	const timeoutSeconds = Math.max(
		1,
		Math.min(300, parseInt(fd.get('timeoutSeconds')?.toString() ?? '30', 10) || 30)
	);
	const usernameDirty = fd.get('usernameDirty') === 'true';
	const passwordDirty = fd.get('passwordDirty') === 'true';
	return {
		host: fd.get('host')?.toString().trim() ?? '',
		port,
		// Немаскированные поля не трогали — backend сохранит текущие секреты.
		username: usernameDirty ? (fd.get('username')?.toString().trim() ?? '') : MASKED,
		password: passwordDirty ? (fd.get('password')?.toString().trim() ?? '') : MASKED,
		from: fd.get('from')?.toString().trim() ?? '',
		timeout_seconds: timeoutSeconds
	};
}

export const load: PageServerLoad = async (event) => {
	try {
		const res = await fetchWithAuth(event, '/v1/settings');
		if (!res.ok) {
			return { smtp: { ...defaults }, error: apiLoadErrorMessage(res, 'Ошибка загрузки настроек') };
		}
		const data = (await res.json()) as SettingsListResponse;
		return { smtp: parseSmtp(getByKey(data.settings ?? [], 'SMTP')), error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { smtp: { ...defaults }, error: 'API недоступен' };
	}
};

export const actions: Actions = {
	updateSmtp: async (event) => {
		try {
			const fd = await event.request.formData();
			const res = await updateSetting(event, 'SMTP', JSON.stringify(smtpPayloadFromForm(fd)));
			if (!res.ok) {
				const data = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, { error: data.message ?? 'Не удалось сохранить настройки SMTP' });
			}
			return { ok: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { error: 'Не удалось сохранить настройки SMTP' });
		}
	},

	testSmtp: async (event) => {
		try {
			const fd = await event.request.formData();
			const res = await fetchWithAuth(event, '/v1/settings/test', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ key: 'SMTP', value: JSON.stringify(smtpPayloadFromForm(fd)) })
			});
			if (!res.ok) {
				const data = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, { testError: data.message ?? 'Ошибка подключения к SMTP' });
			}
			return { testOk: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { testError: 'Ошибка запроса к API' });
		}
	}
};
