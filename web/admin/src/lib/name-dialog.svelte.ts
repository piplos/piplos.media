type NameDialogOptions = {
	title: string;
	label?: string;
	value?: string;
	placeholder?: string;
	confirmLabel?: string;
	cancelLabel?: string;
};

export const nameDialogState = $state({
	open: false,
	title: '',
	label: 'Название',
	value: '',
	placeholder: '',
	confirmLabel: 'Сохранить',
	cancelLabel: 'Отмена'
});

let resolveFn: ((value: string | null) => void) | null = null;

export function promptName(options: NameDialogOptions): Promise<string | null> {
	return new Promise((resolve) => {
		nameDialogState.title = options.title;
		nameDialogState.label = options.label ?? 'Название';
		nameDialogState.value = options.value ?? '';
		nameDialogState.placeholder = options.placeholder ?? '';
		nameDialogState.confirmLabel = options.confirmLabel ?? 'Сохранить';
		nameDialogState.cancelLabel = options.cancelLabel ?? 'Отмена';
		resolveFn = resolve;
		nameDialogState.open = true;
	});
}

export function resolveNameDialog(value: string | null) {
	resolveFn?.(value);
	resolveFn = null;
	nameDialogState.open = false;
}
