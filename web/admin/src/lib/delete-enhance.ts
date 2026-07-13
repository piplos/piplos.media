import type { SubmitFunction } from '@sveltejs/kit';
import { confirmAction } from './confirm.svelte';

type DeleteEnhanceOptions = {
	message: string;
	onSuccess: () => void | Promise<void>;
	onError?: (message: string) => void;
};

export function deleteEnhance({ message, onSuccess, onError }: DeleteEnhanceOptions): SubmitFunction {
	return async ({ cancel }) => {
		if (!(await confirmAction(message))) {
			cancel();
			return;
		}
		return async ({ result }) => {
			if (result.type === 'success') {
				await onSuccess();
			} else {
				const errorMessage =
					result.type === 'failure'
						? ((result.data?.error as string) ?? 'Не удалось удалить')
						: 'Не удалось удалить';
				onError?.(errorMessage);
			}
		};
	};
}
