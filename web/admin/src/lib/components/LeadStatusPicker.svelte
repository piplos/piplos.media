<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import {
		LEAD_STATUS_LABELS,
		LEAD_STATUS_ORDER,
		LEAD_STATUS_VARIANTS,
		type LeadStatus
	} from '$lib/types';

	interface Props {
		status: LeadStatus;
		action?: string;
	}

	let { status, action = '?/setStatus' }: Props = $props();
</script>

<div class="lead-status-picker" role="group" aria-label="Статус заявки">
	<div class="lead-status-options">
		{#each LEAD_STATUS_ORDER as s (s)}
			<form
				method="POST"
				{action}
				class="admin-action-form"
				use:enhance={() => {
					return async ({ result }) => {
						if (result.type === 'success') {
							toast.success('Статус обновлён');
							await invalidateAll();
						} else {
							toast.error('Не удалось обновить статус');
						}
					};
				}}
			>
				<input type="hidden" name="status" value={s} />
				<button
					type="submit"
					class="lead-status-btn lead-status-btn--{LEAD_STATUS_VARIANTS[s]}"
					class:lead-status-btn--active={status === s}
					disabled={status === s}
					aria-pressed={status === s}
				>
					{LEAD_STATUS_LABELS[s]}
				</button>
			</form>
		{/each}
	</div>
</div>

<style>
	.lead-status-picker {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.lead-status-options {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}
	.lead-status-btn {
		padding: 0.25rem 0.625rem;
		font-size: 0.75rem;
		font-weight: 500;
		line-height: 1.25;
		border: 1px solid transparent;
		border-radius: 9999px;
		cursor: pointer;
		transition:
			opacity 0.15s,
			box-shadow 0.15s;
	}
	.lead-status-btn:hover:not(:disabled) {
		opacity: 0.9;
	}
	.lead-status-btn--active {
		cursor: default;
		box-shadow: 0 0 0 2px #fff, 0 0 0 3px currentColor;
	}
	.lead-status-btn--info {
		background: rgba(59, 130, 246, 0.15);
		color: #1d4ed8;
	}
	.lead-status-btn--info.lead-status-btn--active {
		background: rgba(59, 130, 246, 0.28);
	}
	.lead-status-btn--warning {
		background: rgba(234, 179, 8, 0.2);
		color: #a16207;
	}
	.lead-status-btn--warning.lead-status-btn--active {
		background: rgba(234, 179, 8, 0.34);
	}
	.lead-status-btn--success {
		background: rgba(22, 163, 74, 0.15);
		color: #15803d;
	}
	.lead-status-btn--success.lead-status-btn--active {
		background: rgba(22, 163, 74, 0.28);
	}
	.lead-status-btn--danger {
		background: rgba(239, 68, 68, 0.15);
		color: #b91c1c;
	}
	.lead-status-btn--danger.lead-status-btn--active {
		background: rgba(239, 68, 68, 0.28);
	}
</style>
