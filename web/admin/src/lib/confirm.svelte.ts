type ConfirmOptions = {
	title?: string;
	message: string;
	confirmLabel?: string;
	cancelLabel?: string;
};

export const confirmState = $state({
	open: false,
	title: 'Подтверждение',
	message: '',
	confirmLabel: 'Удалить',
	cancelLabel: 'Отмена'
});

let resolveFn: ((value: boolean) => void) | null = null;

export function confirmAction(options: string | ConfirmOptions): Promise<boolean> {
	const opts = typeof options === 'string' ? { message: options } : options;
	return new Promise((resolve) => {
		confirmState.title = opts.title ?? 'Подтверждение';
		confirmState.message = opts.message;
		confirmState.confirmLabel = opts.confirmLabel ?? 'Удалить';
		confirmState.cancelLabel = opts.cancelLabel ?? 'Отмена';
		resolveFn = resolve;
		confirmState.open = true;
	});
}

export function resolveConfirm(value: boolean) {
	resolveFn?.(value);
	resolveFn = null;
	confirmState.open = false;
}
