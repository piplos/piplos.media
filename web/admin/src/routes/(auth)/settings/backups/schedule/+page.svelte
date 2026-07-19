<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import { storageOptions, typeOptions } from '../_options';
	import type { BackupSettingsForm } from '../_types';

	let { data } = $props();

	const backup = $derived(data.backup as BackupSettingsForm);

	let savingBackup = $state(false);

	// Writable derived: значения из load, редактируемые локально; сбрасываются при invalidateAll.
	let enabled = $derived(backup.enabled);
	let backupType = $derived(backup.type as string);
	let intervalHours = $derived(backup.intervalHours);
	let keep = $derived(backup.keep);
	let backupStorage = $derived(backup.storage as string);
</script>

<svelte:head>
	<title>Расписание — Бекапы — Настройки — Piplos Admin</title>
</svelte:head>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	<Card padding="sm">
		<h2 class="section-title">Автоматические бекапы</h2>

		<form
			method="POST"
			action="?/updateBackup"
			use:enhance={() => {
				savingBackup = true;
				return async ({ result }) => {
					savingBackup = false;
					if (result.type === 'success') {
						toast.success('Настройки бекапов сохранены');
						await invalidateAll();
					} else if (result.type === 'failure') {
						toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
					}
				};
			}}
		>
			<div class="backup-form">
				<label class="check">
					<input type="checkbox" bind:checked={enabled} />
					Делать бекапы автоматически по расписанию
				</label>

				<div class="fields-row">
					<FormField label="Что копировать" id="backup-type">
						<Select id="backup-type" bind:value={backupType} options={typeOptions} />
					</FormField>
					<FormField label="Хранилище" id="backup-storage">
						<Select id="backup-storage" bind:value={backupStorage} options={storageOptions} />
					</FormField>
					<FormField label="Интервал (часов)" id="backup-interval" hint="От 1 до 720">
						<Input
							id="backup-interval"
							type="number"
							bind:value={intervalHours}
							min="1"
							max="720"
						/>
					</FormField>
					<FormField label="Хранить копий" id="backup-keep" hint="0 — без ограничения">
						<Input id="backup-keep" type="number" bind:value={keep} min="0" max="100" />
					</FormField>
				</div>
			</div>

			<input type="hidden" name="enabled" value={enabled ? 'true' : 'false'} />
			<input type="hidden" name="type" value={backupType} />
			<input type="hidden" name="storage" value={backupStorage} />
			<input type="hidden" name="intervalHours" value={intervalHours} />
			<input type="hidden" name="keep" value={keep} />

			<div class="actions-row">
				<Button type="submit" loading={savingBackup}>Сохранить</Button>
			</div>
		</form>
	</Card>
{/if}

<style>
	.text-muted {
		color: #71717a;
		font-size: 0.875rem;
	}
	.section-title {
		margin: 0 0 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
	}
	.backup-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.fields-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: flex-start;
	}
	.fields-row > :global(*) {
		flex: 1;
		min-width: 8rem;
	}
	.check {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}
	.actions-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		margin-top: 1.25rem;
	}
</style>
