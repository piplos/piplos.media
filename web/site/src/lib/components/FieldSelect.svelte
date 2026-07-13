<script lang="ts">
	type Option = {
		value: string;
		label: string;
	};

	type Props = {
		id: string;
		name: string;
		placeholder: string;
		options: Option[];
		value?: string;
		variant?: 'default' | 'inline';
		onchange?: () => void;
		ariaLabel?: string;
	};

	let {
		id,
		name,
		placeholder,
		options,
		value = $bindable(''),
		variant = 'default',
		onchange,
		ariaLabel
	}: Props = $props();

	let open = $state(false);
	let rootEl = $state<HTMLDivElement | null>(null);

	let selectedLabel = $derived(
		options.find((option) => option.value === value)?.label ?? placeholder
	);

	let isPlaceholder = $derived(!value);

	function toggle(e?: Event) {
		e?.stopPropagation();
		open = !open;
	}

	function commit(next: string) {
		if (next === value) return;
		value = next;
		onchange?.();
	}

	function select(next: string) {
		commit(next);
		open = false;
	}

	function moveSelection(delta: number) {
		const idx = options.findIndex((option) => option.value === value);
		const next = options[Math.min(Math.max(idx + delta, 0), options.length - 1)];
		if (next) commit(next.value);
	}

	function onTriggerKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			return;
		}

		if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
			e.preventDefault();
			open = true;
			moveSelection(e.key === 'ArrowDown' ? 1 : -1);
			return;
		}

		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			toggle();
		}
	}

	$effect(() => {
		if (!open) return;

		function onDocumentClick(e: MouseEvent) {
			if (!rootEl?.contains(e.target as Node)) open = false;
		}

		function onDocumentKeydown(e: KeyboardEvent) {
			if (e.key === 'Escape') open = false;
		}

		document.addEventListener('click', onDocumentClick);
		document.addEventListener('keydown', onDocumentKeydown);

		return () => {
			document.removeEventListener('click', onDocumentClick);
			document.removeEventListener('keydown', onDocumentKeydown);
		};
	});
</script>

<div class="field-select-root" class:field-select-root--inline={variant === 'inline'} bind:this={rootEl}>
	<input type="hidden" {name} {value} />
	<button
		type="button"
		{id}
		class="field-select-trigger"
		class:field-select-trigger--inline={variant === 'inline'}
		class:is-placeholder={isPlaceholder}
		class:is-open={open}
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-controls="{id}-listbox"
		aria-label={ariaLabel}
		onclick={toggle}
		onkeydown={onTriggerKeydown}
	>
		{selectedLabel}
	</button>

	{#if open}
		<ul
			id="{id}-listbox"
			class="field-select-menu"
			class:field-select-menu--inline={variant === 'inline'}
			role="listbox"
			aria-labelledby={id}
		>
			{#each options as option (option.value)}
				<li role="presentation">
					<button
						type="button"
						role="option"
						class="field-select-option"
						class:selected={value === option.value}
						aria-selected={value === option.value}
						onclick={() => select(option.value)}
					>
						{option.label}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
