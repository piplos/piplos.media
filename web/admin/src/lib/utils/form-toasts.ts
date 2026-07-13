import toast from 'svelte-french-toast';

export interface FormToastConfig {
	[key: string]: { success?: string; error?: string };
}

export function createFormToastHandler(config: FormToastConfig) {
	const seen = new Map<string, string>();

	return function handleFormToasts(form: Record<string, unknown> | null) {
		if (form == null) return;

		for (const [key, msgs] of Object.entries(config)) {
			const successField = `${key}Success`;
			const errorField = `${key}Error`;

			if (form[successField] === true) {
				const msg = msgs.success ?? 'Готово';
				const dedup = `s:${key}:${msg}`;
				if (seen.get(key) !== dedup) {
					seen.set(key, dedup);
					toast.success(msg);
				}
			} else if (form[errorField] != null && form[errorField] !== '') {
				const msg =
					(typeof form[errorField] === 'string' ? form[errorField] : msgs.error) ?? 'Ошибка';
				const dedup = `e:${key}:${msg}`;
				if (seen.get(key) !== dedup) {
					seen.set(key, dedup);
					toast.error(msg);
				}
			}
		}
	};
}
