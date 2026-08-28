<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import TabPageActions from '$lib/components/TabPageActions.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import { formatDate } from '$lib/format';

	let { data } = $props();

	let createOpen = $state(false);
	let newName = $state('');
	// Сырой ключ показывается один раз после создания; в API больше не возвращается.
	let createdKey = $state('');
	let submitting = $state(false);

	function openCreate() {
		newName = '';
		createdKey = '';
		createOpen = true;
	}

	// Копирование по образцу страницы «Файлы»: toast-подтверждение.
	async function copyKey() {
		try {
			await navigator.clipboard.writeText(createdKey);
			toast.success('Ключ скопирован');
		} catch {
			toast.error('Не удалось скопировать');
		}
	}

	const revokeEnhance = deleteEnhance({
		message: 'Отозвать ключ? Агенты с этим ключом потеряют доступ.',
		confirmLabel: 'Отозвать',
		onSuccess: async () => {
			toast.success('Ключ отозван');
			await invalidateAll();
		},
		onError: (message) => toast.error(message)
	});

	const deleteEnhance_ = deleteEnhance({
		message: 'Удалить ключ навсегда?',
		onSuccess: async () => {
			toast.success('Ключ удалён');
			await invalidateAll();
		},
		onError: (message) => toast.error(message)
	});
</script>

<svelte:head>
	<title>API-ключи — Настройки — Piplos Admin</title>
</svelte:head>

<TabPageActions>
	<Button variant="success" onclick={openCreate}>+ Новый ключ</Button>
</TabPageActions>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else if !data.apiKeys.length}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">
			Ключей пока нет. Создайте ключ для внешнего агента (Manus, n8n и др.).
			Описание эндпоинтов — в docs/api/agent-api.md.
		</p>
	</div>
{:else}
	<div class="admin-table-wrap">
		<table class="chart-table">
			<thead>
				<tr>
					<th>Название</th>
					<th>Ключ</th>
					<th>Статус</th>
					<th>Последнее использование</th>
					<th>Создан</th>
					<th class="admin-table-cell-actions"></th>
				</tr>
			</thead>
			<tbody>
				{#each data.apiKeys as key (key.id)}
					<tr class:row-blocked={Boolean(key.revoked_at)}>
						<td class="chart-cell-main">{key.name}</td>
						<td class="chart-cell-muted"><code>{key.key_prefix}…</code></td>
						<td>
							{#if key.revoked_at}
								<Badge variant="danger" pill>Отозван</Badge>
							{:else}
								<Badge variant="success" pill>Активен</Badge>
							{/if}
						</td>
						<td class="chart-cell-muted">{key.last_used_at ? formatDate(key.last_used_at) : '—'}</td>
						<td class="chart-cell-muted">{formatDate(key.created_at)}</td>
						<td class="admin-table-cell-actions">
							<div class="admin-actions-wrap">
								{#if !key.revoked_at}
									<form
										method="POST"
										action="?/revoke"
										class="admin-action-form"
										use:enhance={revokeEnhance}
									>
										<input type="hidden" name="id" value={key.id} />
										<button type="submit" class="admin-action-btn" title="Отозвать" aria-label="Отозвать ключ {key.name}">
											<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
												<rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
												<path d="M7 11V7a5 5 0 0 1 9.9-1" />
											</svg>
										</button>
									</form>
								{/if}
								<form
									method="POST"
									action="?/delete"
									class="admin-action-form"
									use:enhance={deleteEnhance_}
								>
									<input type="hidden" name="id" value={key.id} />
									<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить ключ {key.name}">
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

<Drawer bind:open={createOpen} title={createdKey ? 'Ключ создан' : 'Новый API-ключ'}>
	{#if createdKey}
		<!-- Шаг 2: сырой ключ — единственный показ -->
		<p class="text-muted">
			Скопируйте ключ сейчас — он больше нигде не будет показан.
			В системе хранится только его хеш.
		</p>
		<div class="copy-field key-reveal">
			<input type="text" readonly value={createdKey} aria-label="API-ключ" />
			<Button variant="secondary" onclick={copyKey}>Копировать</Button>
		</div>
		<div class="form-actions">
			<Button type="button" onclick={() => (createOpen = false)}>Готово</Button>
		</div>
	{:else}
		<!-- Шаг 1: имя ключа -->
		<form
			method="POST"
			action="?/create"
			class="drawer-form"
			use:enhance={() => {
				submitting = true;
				return async ({ result }) => {
					submitting = false;
					if (result.type === 'success') {
						createdKey = (result.data?.createdKey as string) ?? '';
						await invalidateAll();
					} else if (result.type === 'failure') {
						toast.error((result.data?.error as string) ?? 'Не удалось создать ключ');
					}
				};
			}}
		>
			<FormField label="Название" id="api-key-name" hint="Например, имя агента: manus">
				<Input
					id="api-key-name"
					name="name"
					bind:value={newName}
					placeholder="manus"
					required
				/>
			</FormField>
			<div class="form-actions">
				<Button type="button" variant="secondary" onclick={() => (createOpen = false)}>Отмена</Button>
				<Button type="submit" loading={submitting}>Создать</Button>
			</div>
		</form>
	{/if}
</Drawer>

<style>
	/* Только специфика страницы: отступ поля показа ключа внутри Drawer.
	   Общие стили (подсветка row-blocked, поле copy-field) — в app.css. */
	.key-reveal {
		margin-top: 1rem;
	}
	.drawer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		padding-top: 0.25rem;
	}
</style>
