<script lang="ts">
	interface Props {
		id: string;
		name?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		error?: boolean;
		required?: boolean;
		rows?: number;
		class?: string;
	}
	let {
		id,
		name,
		value = $bindable(''),
		placeholder,
		disabled = false,
		error = false,
		required = false,
		rows = 4,
		class: className = ''
	}: Props = $props();
</script>

<textarea
	{id}
	{name}
	bind:value
	{placeholder}
	{disabled}
	{required}
	{rows}
	class="textarea {error ? 'textarea-error' : ''} {className}"
	aria-invalid={error}
></textarea>

<style>
	.textarea {
		width: 100%;
		min-height: 6rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5;
		font-family: inherit;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		background: #fff;
		color: #111;
		box-sizing: border-box;
		resize: vertical;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.textarea::placeholder {
		color: #9ca3af;
	}
	.textarea:hover:not(:disabled) {
		border-color: #9ca3af;
	}
	.textarea:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	.textarea:disabled {
		background: #f4f4f5;
		cursor: not-allowed;
		opacity: 1;
	}
	.textarea-error {
		border-color: #dc2626;
	}
	.textarea-error:focus {
		border-color: #dc2626;
		box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.2);
	}
</style>
