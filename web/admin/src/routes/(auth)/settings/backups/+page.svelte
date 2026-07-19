<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { SubmitFunction } from '@sveltejs/kit';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Select from '$lib/components/Select.svelte';
	import TabPageActions from '$lib/components/TabPageActions.svelte';
	import { confirmAction } from '$lib/confirm.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import { formatBytes, formatDate } from '$lib/format';
	import { storageLabel, storageOptions, typeBadgeVariant, typeLabel, typeOptions } from './_options';
	import type { BackupSettingsForm, BackupStatus, BackupStorage } from './_types';

	let { data } = $props();

	const backup = $derived(data.backup as BackupSettingsForm);

	let drawerOpen = $state(false);
	let starting = $state(false);
	let restoring = $state(false);

	// Разовый бекап: тип и хранилище по умолчанию берутся из настроек расписания.
	let runType = $state('full');
	let runStorage = $state('local');

	const status = $derived(data.status as BackupStatus);
	const running = $derived(status.running);

	function openCreate() {
		runType = backup.type as string;
		runStorage = backup.storage as string;
		drawerOpen = true;
	}

	// Поллинг статуса, пока идёт бекап/восстановление: по завершении
	// показываем результат и перезагружаем данные страницы.
	$effect(() => {
		if (!running) return;
		const timer = setInterval(async () => {
			try {
				const res = await fetch('/api/backups?action=status');
				if (!res.ok) return;
				const payload = (await res.json()) as { status?: BackupStatus };
				const next = payload.status;
				if (!next || next.running) return;
				clearInterval(timer);
				const last = next.last;
				if (last?.ok) {
					toast.success(last.op === 'restore' ? 'Восстановление завершено' : 'Бекап создан');
				} else if (last) {
					toast.error(last.error ?? 'Операция завершилась с ошибкой');
				}
				await invalidateAll();
			} catch {
				// Сетевые сбои поллинга игнорируем — следующий тик повторит запрос.
			}
		}, 3000);
		return () => clearInterval(timer);
	});

	function restoreEnhance(name: string, storage: BackupStorage): SubmitFunction {
		return async ({ cancel }) => {
			const ok = await confirmAction({
				title: 'Восстановление из бекапа',
				message: `Восстановить данные из «${name}» (${storageLabel(storage)})? Текущие данные будут перезаписаны содержимым архива.`,
				confirmLabel: 'Восстановить'
			});
			if (!ok) {
				cancel();
				return;
			}
			restoring = true;
			return async ({ result }) => {
				restoring = false;
				if (result.type === 'success') {
					toast.success('Восстановление запущено');
					await invalidateAll();
				} else if (result.type === 'failure') {
					toast.error((result.data?.error as string) ?? 'Не удалось запустить восстановление');
				}
			};
		};
	}
</script>

<svelte:head>
	<title>Архивы — Бекапы — Настройки — Piplos Admin</title>
</svelte:head>

<TabPageActions>
	<Button variant="success" onclick={openCreate} disabled={running}>+ Создать бекап</Button>
</TabPageActions>

{#if data.error && !data.archives?.length}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	{#if running}
		<div class="status-banner">
			<span class="status-spinner" aria-hidden="true"></span>
			{status.op === 'restore' ? 'Идёт восстановление из бекапа…' : 'Идёт создание бекапа…'}
		</div>
	{:else if status.last && !status.last.ok}
		<div class="status-banner status-banner--error">
			Последняя операция завершилась с ошибкой: {status.last.error}
		</div>
	{/if}

	{#if !data.archives?.length}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">Бекапов пока нет. Создайте первый.</p>
		</div>
	{:else}
		<div class="admin-table-wrap">
			<table class="chart-table">
				<thead>
					<tr>
						<th>Архив</th>
						<th>Тип</th>
						<th>Хранилище</th>
						<th>Размер</th>
						<th>Создан</th>
						<th class="admin-table-cell-actions"></th>
					</tr>
				</thead>
				<tbody>
					{#each data.archives as archive (archive.storage + '/' + archive.name)}
						<tr>
							<td class="chart-cell-main archive-name">{archive.name}</td>
							<td>
								<Badge variant={typeBadgeVariant(archive.type)} pill>{typeLabel(archive.type)}</Badge>
							</td>
							<td>
								<Badge variant={archive.storage === 's3' ? 'info' : 'neutral'} pill>
									{storageLabel(archive.storage)}
								</Badge>
							</td>
							<td class="chart-cell-muted">{formatBytes(archive.size)}</td>
							<td class="chart-cell-muted">{formatDate(archive.mod_time)}</td>
							<td class="admin-table-cell-actions">
								<div class="admin-actions-wrap">
									<a
										class="admin-action-btn"
										href={`/api/backups?action=download&storage=${encodeURIComponent(archive.storage)}&name=${encodeURIComponent(archive.name)}`}
										download={archive.name}
										title="Скачать"
										aria-label="Скачать архив"
									>
										<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
											<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
											<polyline points="7 10 12 15 17 10" />
											<line x1="12" y1="15" x2="12" y2="3" />
										</svg>
									</a>
									<form
										method="POST"
										action="?/restoreBackup"
										class="admin-action-form"
										use:enhance={restoreEnhance(archive.name, archive.storage)}
									>
										<input type="hidden" name="name" value={archive.name} />
										<input type="hidden" name="storage" value={archive.storage} />
										<button
											type="submit"
											class="admin-action-btn"
											title="Восстановить"
											aria-label="Восстановить из архива"
											disabled={running || restoring}
										>
											<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
												<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
												<path d="M3 3v5h5" />
											</svg>
										</button>
									</form>
									<form
										method="POST"
										action="?/deleteBackup"
										class="admin-action-form"
										use:enhance={deleteEnhance({
											message: `Удалить архив ${archive.name}?`,
											onSuccess: async () => {
												toast.success('Архив удалён');
												await invalidateAll();
											},
											onError: (message) => toast.error(message)
										})}
									>
										<input type="hidden" name="name" value={archive.name} />
										<input type="hidden" name="storage" value={archive.storage} />
										<button
											type="submit"
											class="admin-action-btn"
											title="Удалить"
											aria-label="Удалить архив"
										>
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
{/if}

<Drawer bind:open={drawerOpen} title="Создать бекап">
	<form
		method="POST"
		action="?/runBackup"
		class="drawer-form"
		use:enhance={() => {
			starting = true;
			return async ({ result }) => {
				starting = false;
				if (result.type === 'success') {
					drawerOpen = false;
					toast.success('Бекап запущен');
					await invalidateAll();
				} else if (result.type === 'failure') {
					toast.error((result.data?.error as string) ?? 'Не удалось запустить бекап');
				}
			};
		}}
	>
		<FormField label="Тип" id="backup-type">
			<Select id="backup-type" bind:value={runType} options={typeOptions} ariaLabel="Тип бекапа" />
		</FormField>
		<FormField label="Хранилище" id="backup-storage">
			<Select
				id="backup-storage"
				bind:value={runStorage}
				options={storageOptions}
				ariaLabel="Хранилище"
			/>
		</FormField>
		<input type="hidden" name="type" value={runType} />
		<input type="hidden" name="storage" value={runStorage} />
		<div class="form-actions">
			<Button type="submit" variant="success" loading={starting} disabled={running} fullWidth>
				Создать
			</Button>
		</div>
	</form>
</Drawer>

<style>
	.text-muted {
		color: #71717a;
		font-size: 0.875rem;
	}
	.drawer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.form-actions {
		padding-top: 0.25rem;
	}
	.status-banner {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 0.875rem;
		margin-bottom: 1rem;
		font-size: 0.875rem;
		color: #1d4ed8;
		background: rgba(59, 130, 246, 0.1);
		border-radius: 8px;
	}
	.status-banner--error {
		color: #b91c1c;
		background: rgba(239, 68, 68, 0.1);
	}
	.status-spinner {
		display: inline-block;
		width: 0.875rem;
		height: 0.875rem;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		animation: backup-spin 0.6s linear infinite;
	}
	@keyframes backup-spin {
		to {
			transform: rotate(360deg);
		}
	}
	.archive-name {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.8125rem;
	}
	.admin-action-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
