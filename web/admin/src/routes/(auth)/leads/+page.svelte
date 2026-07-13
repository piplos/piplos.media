<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto, invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import Badge from '$lib/components/Badge.svelte';
	import LeadStatusToggleBadge from '$lib/components/LeadStatusToggleBadge.svelte';
	import { formatDate } from '$lib/format';
	import { type LeadStatus } from '$lib/types';

	let { data } = $props();

	const statusVariants: Record<LeadStatus, string> = {
		new: 'info',
		in_progress: 'warning',
		done: 'success',
		spam: 'danger'
	};

	const filters: { value: string; label: string }[] = [
		{ value: '', label: 'Все' },
		{ value: 'new', label: 'Новые' },
		{ value: 'in_progress', label: 'В работе' },
		{ value: 'done', label: 'Завершённые' },
		{ value: 'spam', label: 'Спам' }
	];

	function applyFilter(value: string) {
		goto(value ? `/leads?status=${value}` : '/leads');
	}

	function budgetLabel(budget: number, currency: string): string {
		if (!budget) return '—';
		return `${budget.toLocaleString('ru-RU')} ${currency}`;
	}
</script>

<svelte:head>
	<title>Заявки — Piplos Admin</title>
</svelte:head>

<AdminPage title="Заявки">
	<div class="filter-row" role="group" aria-label="Фильтр по статусу">
		{#each filters as f (f.value)}
			<button
				type="button"
				class="filter-btn"
				class:filter-btn--active={data.status === f.value}
				onclick={() => applyFilter(f.value)}
			>
				{f.label}
			</button>
		{/each}
		<span class="text-muted total-label">Всего: {data.total}</span>
	</div>

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
									variant={statusVariants[lead.status]}
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
</AdminPage>

<style>
	.filter-row {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem;
		background: #f4f4f5;
		border-radius: 10px;
		width: fit-content;
	}
	.filter-btn {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.filter-btn:hover {
		color: #1a1a1a;
	}
	.filter-btn--active {
		color: #1a1a1a;
		background: #fff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
	.total-label {
		margin-left: 0.75rem;
		margin-right: 0.5rem;
	}
	:global(.type-badge) {
		margin-right: 0.25rem;
	}
</style>
