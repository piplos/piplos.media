import type { Actions, PageServerLoad } from './$types';
import { updateSetting } from '$lib/settings.server';
import {
	apiFormAction,
	loadSettingValue,
	parseS3,
	postJSON,
	s3PayloadFromForm
} from '../_backups.server';

export const load: PageServerLoad = async (event) => {
	const { value: s3, error } = await loadSettingValue(event, 'S3', parseS3);
	return { s3, error };
};

export const actions: Actions = {
	updateS3: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() => updateSetting(event, 'S3', JSON.stringify(s3PayloadFromForm(fd))),
			'Не удалось сохранить настройки S3'
		);
	},

	testS3: async (event) => {
		const fd = await event.request.formData();
		return apiFormAction(
			() =>
				postJSON(event, '/v1/settings/test', {
					key: 'S3',
					value: JSON.stringify(s3PayloadFromForm(fd))
				}),
			'Ошибка подключения к S3',
			'testError'
		);
	}
};
