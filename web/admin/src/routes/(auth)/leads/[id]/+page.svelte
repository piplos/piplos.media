<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Card from '$lib/components/Card.svelte';
	import { formatDate } from '$lib/format';
	import { LEAD_STATUS_LABELS, type LeadStatus } from '$lib/types';

	let { data } = $props();
	const lead = $derived(data.lead);

	const statuses: LeadStatus[] = ['new', 'in_progress', 'done', 'spam'];

	const infoRows = $derived([
		{ label: 'Имя', value: `${lead.first_name} ${lead.last_name}`.trim() },
		{ label: 'Email', value: lead.email },
		{ label: 'Компания', value: lead.company || '—' },
		{ label: 'Телефон', value: lead.phone || '—' },
		{ label: 'Как нашли', value: lead.how_found || '—' },
		{ label: 'Язык сайта', value: lead.lang },
		{ label: 'Бюджет', value: lead.budget ? `${lead.budget.toLocaleString('ru-RU')} ${lead.currency}` : '—' },
		{ label: 'Сроки', value: lead.timeline || '—' },
		{ label: 'Стадия', value: lead.stage || '—' },
		{ label: 'Создана', value: formatDate(lead.created_at) }
	]);
</script>

<svelte:head>
	<title>Заявка: {lead.project_name || 'Без названия'} — Piplos Admin</title>
</svelte:head>

<AdminPage title={lead.project_name || 'Заявка без названия'}>
	{#snippet actions()}
		<a href="/leads" class="back-link">← К списку заявок</a>
	{/snippet}

	<div class="lead-status-row">
		<span class="text-muted">Статус:</span>
		{#each statuses as s (s)}
			<form
				method="POST"
				action="?/setStatus"
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
					class="status-btn"
					class:status-btn--active={lead.status === s}
					disabled={lead.status === s}
				>
					{LEAD_STATUS_LABELS[s]}
				</button>
			</form>
		{/each}
	</div>

	<div class="lead-grid">
		<Card padding="sm">
			<h2 class="card-title">Проект</h2>
			<div class="lead-types">
				{#each lead.types as t (t)}
					<Badge variant="neutral">{t}</Badge>
				{/each}
			</div>
			{#if lead.description}
				<h3 class="block-label">Описание</h3>
				<p class="block-text">{lead.description}</p>
			{/if}
			{#if lead.stack}
				<h3 class="block-label">Желаемый стек</h3>
				<p class="block-text">{lead.stack}</p>
			{/if}
			{#if lead.references}
				<h3 class="block-label">Референсы</h3>
				<p class="block-text">{lead.references}</p>
			{/if}
			{#if lead.notes}
				<h3 class="block-label">Дополнительно</h3>
				<p class="block-text">{lead.notes}</p>
			{/if}
		</Card>

		<Card padding="sm">
			<h2 class="card-title">Контакт и параметры</h2>
			<dl class="info-list">
				{#each infoRows as row (row.label)}
					<div class="info-row">
						<dt>{row.label}</dt>
						<dd>{row.value}</dd>
					</div>
				{/each}
			</dl>
		</Card>
	</div>
</AdminPage>

<style>
	.back-link {
		font-size: 0.875rem;
		color: #71717a;
		text-decoration: none;
	}
	.back-link:hover {
		color: #2563eb;
	}
	.lead-status-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.status-btn {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #71717a;
		background: #fff;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s, border-color 0.15s;
	}
	.status-btn:hover:not(:disabled) {
		color: #1a1a1a;
		border-color: #9ca3af;
	}
	.status-btn--active {
		color: #fff;
		background: #111;
		border-color: #111;
		cursor: default;
	}
	.lead-grid {
		display: grid;
		grid-template-columns: 3fr 2fr;
		gap: 1rem;
		align-items: start;
	}
	@media (max-width: 768px) {
		.lead-grid {
			grid-template-columns: 1fr;
		}
	}
	.card-title {
		margin: 0 0 0.75rem;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
	}
	.lead-types {
		display: flex;
		gap: 0.375rem;
		flex-wrap: wrap;
		margin-bottom: 0.75rem;
	}
	.block-label {
		margin: 1rem 0 0.25rem;
		font-size: 0.8125rem;
		font-weight: 600;
		color: #52525b;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.block-text {
		margin: 0;
		font-size: 0.9375rem;
		color: #444;
		white-space: pre-wrap;
	}
	.info-list {
		margin: 0;
		display: flex;
		flex-direction: column;
	}
	.info-row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.5rem 0;
		border-bottom: 1px solid #f4f4f5;
	}
	.info-row:last-child {
		border-bottom: none;
	}
	.info-row dt {
		font-size: 0.875rem;
		color: #71717a;
	}
	.info-row dd {
		margin: 0;
		font-size: 0.875rem;
		color: #18181b;
		font-weight: 500;
		text-align: right;
		overflow-wrap: anywhere;
	}
</style>
