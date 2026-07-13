<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Badge from '$lib/components/Badge.svelte';

	interface Props {
		id: string;
		published: boolean;
		publishedLabel?: string;
		draftLabel?: string;
	}
	let {
		id,
		published,
		publishedLabel = 'Опубликован',
		draftLabel = 'Черновик'
	}: Props = $props();
</script>

<form
	method="POST"
	action="?/togglePublished"
	class="status-toggle-form"
	use:enhance={() => {
		return async ({ result }) => {
			if (result.type === 'success') await invalidateAll();
		};
	}}
>
	<input type="hidden" name="id" value={id} />
	<button
		type="submit"
		class="status-toggle-btn"
		title={published ? 'Снять с публикации' : 'Опубликовать'}
		aria-label={published ? 'Снять с публикации' : 'Опубликовать'}
	>
		<Badge variant={published ? 'success' : 'neutral'} pill>
			{published ? publishedLabel : draftLabel}
		</Badge>
	</button>
</form>

<style>
	.status-toggle-form {
		display: inline;
		margin: 0;
		padding: 0;
	}
	.status-toggle-btn {
		display: inline-flex;
		margin: 0;
		padding: 0;
		border: none;
		background: none;
		cursor: pointer;
		border-radius: 9999px;
		transition: opacity 0.15s;
	}
	.status-toggle-btn:hover {
		opacity: 0.8;
	}
	.status-toggle-btn:focus-visible {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}
</style>
