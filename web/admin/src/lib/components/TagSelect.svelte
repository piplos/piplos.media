<script lang="ts">
	interface Option {
		value: string;
		label: string;
	}

	interface Props {
		id: string;
		name?: string;
		options: Option[];
		/** Выбранные значения (для single — массив из 0..1 элементов). */
		values?: string[];
		single?: boolean;
		placeholder?: string;
	}
	let {
		id,
		name,
		options,
		values = $bindable([]),
		single = false,
		placeholder = 'Выберите из списка'
	}: Props = $props();

	let open = $state(false);
	let search = $state('');
	let root: HTMLDivElement | undefined = $state(undefined);
	let searchInput: HTMLInputElement | undefined = $state(undefined);

	const labelOf = (value: string) => options.find((o) => o.value === value)?.label ?? value;

	const available = $derived(
		options.filter(
			(o) =>
				!values.includes(o.value) &&
				(search === '' || o.label.toLowerCase().includes(search.toLowerCase()))
		)
	);

	function add(value: string) {
		values = single ? [value] : [...values, value];
		search = '';
		if (single) {
			open = false;
		} else {
			searchInput?.focus();
		}
	}

	function onMenuPointerDown(e: PointerEvent) {
		e.preventDefault();
	}

	function selectOption(value: string) {
		add(value);
		if (!single) {
			open = true;
			searchInput?.focus();
		}
	}

	function remove(value: string) {
		values = values.filter((v) => v !== value);
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			return;
		}
		if (e.key === 'Enter') {
			e.preventDefault();
			if (available.length > 0) add(available[0].value);
			return;
		}
		if (e.key === 'Backspace' && search === '' && values.length > 0) {
			remove(values[values.length - 1]);
		}
	}

	function onFocusOut(e: FocusEvent) {
		if (root && e.relatedTarget instanceof Node && root.contains(e.relatedTarget)) return;
		open = false;
		search = '';
	}
</script>

<div class="tag-select" class:tag-select--open={open} bind:this={root} onfocusout={onFocusOut}>
	{#if name}
		<input type="hidden" {name} value={values.join(', ')} />
	{/if}
	<div
		class="tag-select-box"
		class:tag-select-box--open={open}
		role="presentation"
		onclick={() => {
			open = true;
			searchInput?.focus();
		}}
	>
		{#each values as value (value)}
			<span class="tag-chip">
				{labelOf(value)}
				<button
					type="button"
					class="tag-chip-remove"
					aria-label="Убрать {labelOf(value)}"
					onclick={(e) => {
						e.stopPropagation();
						remove(value);
					}}
				>
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true">
						<line x1="18" y1="6" x2="6" y2="18" />
						<line x1="6" y1="6" x2="18" y2="18" />
					</svg>
				</button>
			</span>
		{/each}
		<input
			{id}
			bind:this={searchInput}
			class="tag-select-input"
			type="text"
			bind:value={search}
			placeholder={values.length === 0 ? placeholder : ''}
			role="combobox"
			aria-expanded={open}
			aria-controls="{id}-listbox"
			aria-autocomplete="list"
			autocomplete="off"
			onfocus={() => (open = true)}
			onkeydown={onKeydown}
		/>
	</div>
	{#if open}
		<div
			class="tag-select-menu"
			id="{id}-listbox"
			role="listbox"
			aria-label="Доступные значения"
			onmousedown={onMenuPointerDown}
		>
			{#if available.length === 0}
				<div class="tag-select-empty">
					{options.length === values.length ? 'Все значения выбраны' : 'Ничего не найдено'}
				</div>
			{:else}
				{#each available as option (option.value)}
					<button
						type="button"
						class="tag-select-option"
						role="option"
						aria-selected="false"
						onmousedown={onMenuPointerDown}
						onclick={() => selectOption(option.value)}
					>
						{option.label}
					</button>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.tag-select {
		position: relative;
	}
	.tag-select--open {
		z-index: 40;
	}
	.tag-select-box {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.375rem;
		width: 100%;
		min-height: 2.25rem;
		padding: 0.25rem 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		background: #fff;
		box-sizing: border-box;
		cursor: text;
		transition:
			border-color 0.15s,
			box-shadow 0.15s;
	}
	.tag-select-box:hover {
		border-color: #9ca3af;
	}
	.tag-select-box--open,
	.tag-select-box:focus-within {
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	.tag-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.125rem 0.375rem 0.125rem 0.5rem;
		font-size: 0.8125rem;
		line-height: 1.25rem;
		color: #18181b;
		background: #f4f4f5;
		border-radius: 6px;
		white-space: nowrap;
	}
	.tag-chip-remove {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1rem;
		height: 1rem;
		padding: 0;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.tag-chip-remove:hover {
		color: #b91c1c;
		background: #e4e4e7;
	}
	.tag-select-input {
		flex: 1;
		min-width: 6rem;
		padding: 0.125rem 0.25rem;
		font-size: 0.875rem;
		line-height: 1.5;
		color: #111;
		background: transparent;
		border: none;
		outline: none;
	}
	.tag-select-input::placeholder {
		color: #9ca3af;
	}
	.tag-select-menu {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		z-index: 50;
		max-height: 14rem;
		overflow-y: auto;
		padding: 0.25rem;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
	}
	.tag-select-option {
		display: block;
		width: 100%;
		padding: 0.375rem 0.625rem;
		font-size: 0.875rem;
		color: #374151;
		text-align: left;
		background: transparent;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.15s;
	}
	.tag-select-option:hover,
	.tag-select-option:focus-visible {
		background: #f4f4f5;
		outline: none;
	}
	.tag-select-empty {
		padding: 0.5rem 0.625rem;
		font-size: 0.8125rem;
		color: #9ca3af;
	}
</style>
