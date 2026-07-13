<script lang="ts">
	interface Option {
		value: string | number;
		label?: string;
	}
	interface Props {
		value?: string;
		options: Option[];
		name?: string;
		id?: string;
		disabled?: boolean;
		ariaLabel?: string;
		class?: string;
		onchange?: (e: Event) => void;
	}
	let {
		value = $bindable(''),
		options = [],
		name,
		id,
		disabled = false,
		ariaLabel,
		class: className = '',
		onchange
	}: Props = $props();

	function handleChange(e: Event): void {
		const target = e.currentTarget as HTMLSelectElement;
		value = target.value;
		onchange?.(e);
	}
</script>

<select
	{id}
	{name}
	{disabled}
	aria-label={ariaLabel}
	class="select {className}"
	{value}
	onchange={handleChange}
>
	{#each options as opt (String(opt.value))}
		<option value={String(opt.value)}>{opt.label ?? String(opt.value)}</option>
	{/each}
</select>

<style>
	.select {
		width: 100%;
		min-height: 2.25rem;
		padding: 0.375rem 2.25rem 0.375rem 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		background-color: #fff;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.5rem center;
		color: #111;
		box-sizing: border-box;
		appearance: none;
		cursor: pointer;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.select:hover {
		border-color: #9ca3af;
	}
	.select:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	.select:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}
</style>
