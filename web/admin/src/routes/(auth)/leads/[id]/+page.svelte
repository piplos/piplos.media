<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Card from '$lib/components/Card.svelte';
	import LeadStatusPicker from '$lib/components/LeadStatusPicker.svelte';
	import { formatDate } from '$lib/format';
	import type { LeadStatus } from '$lib/types';

	let { data } = $props();
	const lead = $derived(data.lead);

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
	const leadTitle = $derived(lead.project_name || 'Заявка без названия');
</script>

<svelte:head>
	<title>Заявка: {lead.project_name || 'Без названия'} — Piplos Admin</title>
</svelte:head>

<AdminPage
	title={leadTitle}
	breadcrumbs={[
		{ label: 'Заявки', href: '/leads' },
		{ label: leadTitle }
	]}
>
	<div class="lead-grid">
		<Card padding="sm">
			<h2 class="admin-section-title">Проект</h2>
			<div class="lead-types">
				{#each lead.types as t (t)}
					<Badge variant="neutral">{t}</Badge>
				{/each}
			</div>
			{#if lead.description}
				<h3 class="admin-block-label admin-block-label--spaced">Описание</h3>
				<p class="lead-block-text">{lead.description}</p>
			{/if}
			{#if lead.stack}
				<h3 class="admin-block-label admin-block-label--spaced">Желаемый стек</h3>
				<p class="lead-block-text">{lead.stack}</p>
			{/if}
			{#if lead.references}
				<h3 class="admin-block-label admin-block-label--spaced">Референсы</h3>
				<p class="lead-block-text">{lead.references}</p>
			{/if}
			{#if lead.notes}
				<h3 class="admin-block-label admin-block-label--spaced">Дополнительно</h3>
				<p class="lead-block-text">{lead.notes}</p>
			{/if}
		</Card>

		<Card padding="sm">
			<LeadStatusPicker status={lead.status as LeadStatus} />
			<h2 class="admin-section-title admin-section-title--spaced">Контакт и параметры</h2>
			<dl class="info-list">
				{#each infoRows as row (row.label)}
					<div class="info-row">
						<dt class="admin-field-label">{row.label}</dt>
						<dd>{row.value}</dd>
					</div>
				{/each}
			</dl>
		</Card>
	</div>
</AdminPage>

<style>
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
	.lead-types {
		display: flex;
		gap: 0.375rem;
		flex-wrap: wrap;
		margin-bottom: 0.75rem;
	}
	.lead-block-text {
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
	.info-row dd {
		margin: 0;
		font-size: 0.875rem;
		color: #18181b;
		font-weight: 500;
		text-align: right;
		overflow-wrap: anywhere;
	}
</style>
