<script lang="ts">
	import { Label } from 'bits-ui';

	interface Props {
		label: string;
		id: string;
		error?: string;
		hint?: string;
		children: import('svelte').Snippet;
	}
	let { label, id, error, hint, children }: Props = $props();
</script>

<div class="field">
	<div class="field-label">
		<Label.Root for={id}>{label}</Label.Root>
	</div>
	{@render children()}
	{#if error}
		<p class="error-msg" id="{id}-error" role="alert">{error}</p>
	{:else if hint}
		<p class="hint-msg" id="{id}-hint">{hint}</p>
	{/if}
</div>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.field-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}
	.error-msg {
		margin: 0;
		font-size: 0.8125rem;
		color: #b91c1c;
	}
	.hint-msg {
		margin: 0;
		font-size: 0.8125rem;
		color: #71717a;
	}
</style>
