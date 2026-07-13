<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Badge from '$lib/components/Badge.svelte';
	import { LEAD_STATUS_LABELS, nextLeadStatus, type LeadStatus } from '$lib/types';

	interface Props {
		id: string;
		status: LeadStatus;
		variant: string;
	}
	let { id, status, variant }: Props = $props();

	const next = $derived(nextLeadStatus(status));
</script>

<form
	method="POST"
	action="?/setStatus"
	class="status-toggle-form"
	use:enhance={() => {
		return async ({ result }) => {
			if (result.type === 'success') await invalidateAll();
		};
	}}
>
	<input type="hidden" name="id" value={id} />
	<input type="hidden" name="status" value={next} />
	<button
		type="submit"
		class="status-toggle-btn"
		title="Следующий статус: {LEAD_STATUS_LABELS[next]}"
		aria-label="Сменить статус на «{LEAD_STATUS_LABELS[next]}»"
	>
		<Badge {variant} pill>
			{LEAD_STATUS_LABELS[status]}
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
