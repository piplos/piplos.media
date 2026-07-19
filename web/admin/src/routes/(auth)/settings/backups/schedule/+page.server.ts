import type { Actions, PageServerLoad } from './$types';
import { updateSetting } from '$lib/settings.server';
import { apiFormAction, backupPayloadFromForm, loadSettingValue, parseBackup } from '../_backups.server';

export const load: PageServerLoad = async (event) => {
	const { value: backup, error } = await loadSettingValue(event, 'BACKUP', parseBackup);
	return { backup, error };
};

export const actions: Actions = {
	updateBackup: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() => updateSetting(event, 'BACKUP', JSON.stringify(backupPayloadFromForm(fd))),
			'Не удалось сохранить настройки'
		);
	}
};
