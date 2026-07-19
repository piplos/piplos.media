import { fail, isRedirect, type RequestEvent } from '@sveltejs/kit';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, MASKED, type SettingsListResponse } from '$lib/settings.server';
import type { BackupSettingsForm, S3SettingsForm } from './_types';

/** Загружает одну композитную настройку; parse('') используется как fallback. */
export async function loadSettingValue<T>(
	event: RequestEvent,
	key: string,
	parse: (raw: string) => T
): Promise<{ value: T; error: string | null }> {
	try {
		const res = await fetchWithAuth(event, '/v1/settings');
		if (!res.ok) {
			return { value: parse(''), error: apiLoadErrorMessage(res, 'Ошибка загрузки настроек') };
		}
		const settings = (await res.json()) as SettingsListResponse;
		return { value: parse(getByKey(settings.settings ?? [], key)), error: null };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { value: parse(''), error: 'API недоступен' };
	}
}

/** POST с JSON-телом к Go API (Bearer из httpOnly cookie). */
export function postJSON(event: RequestEvent, path: string, body: unknown): Promise<Response> {
	return fetchWithAuth(event, path, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** Приводит запрос к API к результату form action: {ok} либо fail с текстом ошибки. */
export async function apiFormAction(
	request: () => Promise<Response>,
	errorMessage: string,
	errorKey = 'error'
) {
	try {
		const res = await request();
		if (!res.ok) {
			const data = (await res.json().catch(() => ({}))) as { message?: string };
			return fail(res.status, { [errorKey]: data.message ?? errorMessage });
		}
		return { ok: true };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return fail(500, { [errorKey]: errorMessage });
	}
}

const backupDefaults: BackupSettingsForm = {
	enabled: false,
	type: 'full',
	intervalHours: '24',
	keep: '7',
	storage: 'local'
};

const s3Defaults: S3SettingsForm = {
	endpoint: '',
	region: 'auto',
	bucket: '',
	accessKeyId: '',
	accessKeyIdMasked: false,
	secretAccessKey: '',
	secretAccessKeyMasked: false,
	usePathStyle: false
};

export function parseBackup(raw: string): BackupSettingsForm {
	if (!raw) return { ...backupDefaults };
	try {
		const obj = JSON.parse(raw) as Record<string, unknown>;
		const type = obj.type as string;
		const storage = obj.storage as string;
		return {
			enabled: obj.enabled === true,
			type: type === 'db' || type === 'files' ? type : 'full',
			intervalHours: String(obj.interval_hours ?? 24),
			keep: String(obj.keep ?? 7),
			storage: storage === 's3' ? 's3' : 'local'
		};
	} catch {
		return { ...backupDefaults };
	}
}

export function parseS3(raw: string): S3SettingsForm {
	if (!raw) return { ...s3Defaults };
	try {
		const obj = JSON.parse(raw) as Record<string, unknown>;
		const accessKeyId = (obj.access_key_id as string) ?? '';
		const secretAccessKey = (obj.secret_access_key as string) ?? '';
		return {
			endpoint: (obj.endpoint as string) ?? '',
			region: (obj.region as string) ?? 'auto',
			bucket: (obj.bucket as string) ?? '',
			accessKeyId: accessKeyId === MASKED ? '' : accessKeyId,
			accessKeyIdMasked: accessKeyId === MASKED,
			secretAccessKey: secretAccessKey === MASKED ? '' : secretAccessKey,
			secretAccessKeyMasked: secretAccessKey === MASKED,
			usePathStyle: obj.use_path_style === true
		};
	} catch {
		return { ...s3Defaults };
	}
}

export function backupPayloadFromForm(fd: FormData): Record<string, unknown> {
	const intervalHours = Math.max(
		1,
		Math.min(720, parseInt(fd.get('intervalHours')?.toString() ?? '24', 10) || 24)
	);
	const keep = Math.max(0, Math.min(100, parseInt(fd.get('keep')?.toString() ?? '7', 10) || 0));
	const type = fd.get('type')?.toString() ?? 'full';
	const storage = fd.get('storage')?.toString() ?? 'local';
	return {
		enabled: fd.get('enabled') === 'true',
		type: type === 'db' || type === 'files' ? type : 'full',
		interval_hours: intervalHours,
		keep,
		storage: storage === 's3' ? 's3' : 'local'
	};
}

export function s3PayloadFromForm(fd: FormData): Record<string, unknown> {
	const accessKeyDirty = fd.get('accessKeyDirty') === 'true';
	const secretKeyDirty = fd.get('secretKeyDirty') === 'true';
	return {
		endpoint: fd.get('endpoint')?.toString().trim() ?? '',
		region: fd.get('region')?.toString().trim() || 'auto',
		bucket: fd.get('bucket')?.toString().trim() ?? '',
		// Немаскированные поля не трогали — backend сохранит текущие секреты.
		access_key_id: accessKeyDirty ? (fd.get('accessKeyId')?.toString().trim() ?? '') : MASKED,
		secret_access_key: secretKeyDirty
			? (fd.get('secretAccessKey')?.toString().trim() ?? '')
			: MASKED,
		use_path_style: fd.get('usePathStyle') === 'true'
	};
}
