<script lang="ts">
	import Input from '$lib/components/Input.svelte';
	import { slugify } from '$lib/slug';

	interface Props {
		id: string;
		name?: string;
		value?: string;
		placeholder?: string;
		required?: boolean;
		/** Текст-источник для генерации slug (обычно название). */
		source?: string;
	}
	let { id, name, value = $bindable(''), placeholder, required = false, source = '' }: Props = $props();

	const generated = $derived(slugify(source));
</script>

<div class="slug-field">
	<Input {id} {name} bind:value {placeholder} {required} class="slug-input" />
	<button
		type="button"
		class="slug-generate"
		title="Сгенерировать из названия"
		aria-label="Сгенерировать из названия"
		disabled={!generated}
		onclick={() => (value = generated)}
	>
		<svg
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			aria-hidden="true"
		>
			<path
				d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"
			/>
		</svg>
	</button>
</div>

<style>
	.slug-field {
		position: relative;
	}
	.slug-field :global(.slug-input) {
		padding-right: 2.5rem;
	}
	.slug-generate {
		position: absolute;
		top: 50%;
		right: 0.375rem;
		transform: translateY(-50%);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.75rem;
		height: 1.75rem;
		padding: 0;
		border: none;
		border-radius: 6px;
		background: transparent;
		color: #71717a;
		cursor: pointer;
		transition:
			color 0.15s,
			background 0.15s;
	}
	.slug-generate:hover:not(:disabled) {
		color: #18181b;
		background: #f4f4f5;
	}
	.slug-generate:disabled {
		color: #d4d4d8;
		cursor: not-allowed;
	}
	.slug-generate:focus-visible {
		outline: 2px solid #18181b;
		outline-offset: 1px;
	}
</style>
