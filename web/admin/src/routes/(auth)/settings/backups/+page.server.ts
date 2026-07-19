import { isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, type SettingsListResponse } from '$lib/settings.server';
import { apiFormAction, parseBackup, postJSON } from './_backups.server';
import type { BackupArchive, BackupStatus } from './_types';

interface BackupsListResponse {
	archives: BackupArchive[];
	status: BackupStatus;
}

const emptyStatus: BackupStatus = { running: false };

export const load: PageServerLoad = async (event) => {
	try {
		// Настройки нужны только для значений по умолчанию в форме создания.
		const [settingsRes, backupsRes] = await Promise.all([
			fetchWithAuth(event, '/v1/settings'),
			fetchWithAuth(event, '/v1/backups')
		]);
		let rawBackup = '';
		if (settingsRes.ok) {
			const settings = (await settingsRes.json()) as SettingsListResponse;
			rawBackup = getByKey(settings.settings ?? [], 'BACKUP');
		}
		const backup = parseBackup(rawBackup);
		if (!backupsRes.ok) {
			const data = (await backupsRes.json().catch(() => ({}))) as { message?: string };
			return {
				backup,
				archives: [] as BackupArchive[],
				status: emptyStatus,
				error: data.message ?? apiLoadErrorMessage(backupsRes, 'Ошибка загрузки списка бекапов')
			};
		}
		const data = (await backupsRes.json()) as BackupsListResponse;
		return {
			backup,
			archives: data.archives ?? [],
			status: data.status ?? emptyStatus,
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return {
			backup: parseBackup(''),
			archives: [] as BackupArchive[],
			status: emptyStatus,
			error: 'API недоступен'
		};
	}
};

export const actions: Actions = {
	runBackup: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() =>
				postJSON(event, '/v1/backups', {
					type: fd.get('type')?.toString() ?? '',
					storage: fd.get('storage')?.toString() ?? ''
				}),
			'Не удалось запустить бекап'
		);
	},

	restoreBackup: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() =>
				postJSON(event, '/v1/backups/restore', {
					storage: fd.get('storage')?.toString() ?? '',
					name: fd.get('name')?.toString() ?? ''
				}),
			'Не удалось запустить восстановление'
		);
	},

	deleteBackup: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() =>
				postJSON(event, '/v1/backups/delete', {
					storage: fd.get('storage')?.toString() ?? '',
					name: fd.get('name')?.toString() ?? ''
				}),
			'Не удалось удалить архив'
		);
	}
};
