<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import TabPageActions from '$lib/components/TabPageActions.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import { formatDate } from '$lib/format';
	import type { AdminUser } from '$lib/types';
	import UserForm from './UserForm.svelte';

	let { data } = $props();

	let drawerOpen = $state(false);
	let editing = $state<AdminUser | null>(null);

	function openCreate() {
		editing = null;
		drawerOpen = true;
	}

	function openEdit(user: AdminUser) {
		editing = user;
		drawerOpen = true;
	}
</script>

<svelte:head>
	<title>Пользователи — Настройки — Piplos Admin</title>
</svelte:head>

<TabPageActions>
	<Button variant="success" onclick={openCreate}>+ Новый пользователь</Button>
</TabPageActions>

{#if data.error}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">{data.error}</p>
		</div>
{:else if !data.users.length}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">Пользователей пока нет. Добавьте первого.</p>
	</div>
{:else}
	<div class="admin-table-wrap">
			<table class="chart-table">
				<thead>
					<tr>
						<th>Email</th>
						<th>Имя</th>
						<th>Роль</th>
						<th>Создан</th>
						<th class="admin-table-cell-actions"></th>
					</tr>
				</thead>
				<tbody>
					{#each data.users as user (user.id)}
						<tr class:row-blocked={!user.is_active}>
							<td class="chart-cell-main">
								<button type="button" class="admin-text-link row-link" onclick={() => openEdit(user)}>
									{user.email}
								</button>
								{#if !user.is_active}
									<Badge variant="danger" pill>Заблокирован</Badge>
								{/if}
							</td>
							<td>{user.full_name || '—'}</td>
							<td>
								{#if user.role === 'admin'}
									<Badge variant="warning" pill>Администратор</Badge>
								{:else}
									<Badge variant="info" pill>Менеджер</Badge>
								{/if}
							</td>
							<td class="chart-cell-muted">{formatDate(user.created_at)}</td>
							<td class="admin-table-cell-actions">
								<div class="admin-actions-wrap">
									<button
										type="button"
										class="admin-action-btn"
										title="Редактировать"
										aria-label="Редактировать пользователя"
										onclick={() => openEdit(user)}
									>
										<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
											<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
											<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
										</svg>
									</button>
									<form
										method="POST"
										action="?/delete"
										class="admin-action-form"
										use:enhance={deleteEnhance({
											message: `Удалить пользователя ${user.email}?`,
											onSuccess: async () => {
												toast.success('Пользователь удалён');
												if (editing?.id === user.id) drawerOpen = false;
												await invalidateAll();
											},
											onError: (message) => toast.error(message)
										})}
									>
										<input type="hidden" name="id" value={user.id} />
										<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить пользователя">
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

<Drawer
	bind:open={drawerOpen}
	title={editing ? `Пользователь: ${editing.email}` : 'Новый пользователь'}
>
	{#key editing?.id ?? 'new'}
		<UserForm user={editing ?? {}} onSaved={() => (drawerOpen = false)} />
	{/key}
</Drawer>

<style>
	.row-link {
		padding: 0;
		font-size: inherit;
		background: none;
		border: none;
		cursor: pointer;
		margin-right: 0.375rem;
	}
	:global(.chart-table) tbody tr.row-blocked,
	:global(.chart-table) tbody tr.row-blocked td {
		background: rgba(220, 38, 38, 0.06);
	}
</style>
