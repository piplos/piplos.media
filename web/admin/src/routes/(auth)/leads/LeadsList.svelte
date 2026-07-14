<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import Badge from '$lib/components/Badge.svelte';
	import LeadStatusToggleBadge from '$lib/components/LeadStatusToggleBadge.svelte';
	import { formatDate } from '$lib/format';
	import { LEAD_STATUS_VARIANTS, type LeadStatus } from '$lib/types';
	import { LEAD_FILTERS } from './_leads';
	import { leadsBreadcrumbs } from './_leads';

	let { data } = $props();

	function budgetLabel(budget: number, currency: string): string {
		if (!budget) return '—';
		return `${budget.toLocaleString('ru-RU')} ${currency}`;
	}
	const breadcrumbs = $derived(leadsBreadcrumbs(data.status));
</script>

<svelte:head>
	<title>Заявки — Piplos Admin</title>
</svelte:head>

<AdminPage title="Заявки" breadcrumbs={breadcrumbs}>
	<div class="admin-sidebar-row">
		<nav class="admin-sidebar-nav" aria-label="Фильтр по статусу">
			{#each LEAD_FILTERS as f (f.value)}
				<a href={f.href} class="sidebar-link" class:active={data.status === f.value}>
					<span class="sidebar-label">{f.label}</span>
					<span class="sidebar-count">{data.counts[f.value]}</span>
				</a>
			{/each}
		</nav>

		<div class="admin-sidebar-content admin-sidebar-content--no-box">
			{#if data.error}
				<div class="admin-table-wrap admin-table-wrap--empty">
					<p class="text-muted">{data.error}</p>
				</div>
			{:else if !data.leads.length}
				<div class="admin-table-wrap admin-table-wrap--empty">
					<p class="text-muted">Заявок пока нет.</p>
				</div>
			{:else}
				<div class="admin-table-wrap">
					<table class="chart-table">
						<thead>
							<tr>
								<th>Проект</th>
								<th>Контакт</th>
								<th>Тип</th>
								<th>Бюджет</th>
								<th>Статус</th>
								<th>Дата</th>
								<th class="admin-table-cell-actions"></th>
							</tr>
						</thead>
						<tbody>
							{#each data.leads as lead (lead.id)}
								<tr>
									<td class="chart-cell-main">
										<a href="/leads/{lead.id}" class="admin-text-link">
											{lead.project_name || 'Без названия'}
										</a>
									</td>
									<td>
										{lead.first_name}
										{lead.last_name}
										<div class="chart-cell-muted">{lead.email}</div>
									</td>
									<td>
										{#each lead.types as t (t)}
											<Badge variant="neutral" class="type-badge">{t}</Badge>
										{/each}
									</td>
									<td class="chart-cell-muted">{budgetLabel(lead.budget, lead.currency)}</td>
									<td>
										<LeadStatusToggleBadge
											id={lead.id}
											status={lead.status}
											variant={LEAD_STATUS_VARIANTS[lead.status as LeadStatus]}
										/>
									</td>
									<td class="chart-cell-muted">{formatDate(lead.created_at)}</td>
									<td class="admin-table-cell-actions">
										<div class="admin-actions-wrap">
											<a href="/leads/{lead.id}" class="admin-action-btn" title="Открыть" aria-label="Открыть заявку">
												<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
													<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
													<circle cx="12" cy="12" r="3" />
												</svg>
											</a>
											<form
												method="POST"
												action="?/delete"
												class="admin-action-form"
												use:enhance={deleteEnhance({
													message: 'Удалить заявку?',
													onSuccess: async () => {
														toast.success('Заявка удалена');
														await invalidateAll();
													},
													onError: () => toast.error('Не удалось удалить заявку')
												})}
											>
												<input type="hidden" name="id" value={lead.id} />
												<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить заявку">
													<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
														<polyline points="3 6 5 6 21 6" />
														<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
													</svg>
												</button>
											</form>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</AdminPage>

<style>
	:global(.admin-sidebar-nav) .sidebar-link {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.sidebar-label {
		min-width: 0;
	}
	.sidebar-count {
		flex-shrink: 0;
		min-width: 1.375rem;
		height: 1.375rem;
		padding: 0 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		line-height: 1.375rem;
		text-align: center;
		color: #71717a;
		background: #e5e7eb;
		border-radius: 6px;
		box-sizing: border-box;
	}
	:global(.admin-sidebar-nav) a.active .sidebar-count {
		color: #374151;
		background: #f4f4f5;
	}
	:global(.type-badge) {
		margin-right: 0.25rem;
	}
</style>
