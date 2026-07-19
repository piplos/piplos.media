<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import type { S3SettingsForm } from '../_types';

	let { data } = $props();

	const s3 = $derived(data.s3 as S3SettingsForm);

	let savingS3 = $state(false);
	let testingS3 = $state(false);

	// Writable derived: значения из load, редактируемые локально; сбрасываются при invalidateAll.
	let endpoint = $derived(s3.endpoint);
	let region = $derived(s3.region);
	let bucket = $derived(s3.bucket);
	let accessKeyId = $derived(s3.accessKeyId);
	let secretAccessKey = $derived(s3.secretAccessKey);
	let usePathStyle = $derived(s3.usePathStyle);

	// Секрет не менялся -> отправляем "****", backend сохранит текущее значение.
	const accessKeyDirty = $derived(accessKeyId !== s3.accessKeyId);
	const secretKeyDirty = $derived(secretAccessKey !== s3.secretAccessKey);
	const accessKeyPlaceholder = $derived(s3.accessKeyIdMasked ? '•••••••• (задан)' : '');
	const secretKeyPlaceholder = $derived(s3.secretAccessKeyMasked ? '•••••••• (задан)' : '');
</script>

<svelte:head>
	<title>S3 — Бекапы — Настройки — Piplos Admin</title>
</svelte:head>

<!-- Обе формы (сохранить и тестировать) отправляют один и тот же набор полей. -->
{#snippet s3FormFields()}
	<input type="hidden" name="accessKeyDirty" value={accessKeyDirty ? 'true' : 'false'} />
	<input type="hidden" name="secretKeyDirty" value={secretKeyDirty ? 'true' : 'false'} />
	<input type="hidden" name="endpoint" value={endpoint} />
	<input type="hidden" name="region" value={region} />
	<input type="hidden" name="bucket" value={bucket} />
	<input type="hidden" name="accessKeyId" value={accessKeyId} />
	<input type="hidden" name="secretAccessKey" value={secretAccessKey} />
	<input type="hidden" name="usePathStyle" value={usePathStyle ? 'true' : 'false'} />
{/snippet}

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	<Card padding="sm">
		<h2 class="section-title">S3-хранилище (Cloudflare R2)</h2>
		<p class="section-hint">
			Общее подключение к S3-совместимому хранилищу. Для Cloudflare R2 endpoint имеет вид
			https://&lt;account_id&gt;.r2.cloudflarestorage.com
		</p>

		<div class="backup-form">
			<div class="fields-row">
				<div class="field-wide">
					<FormField label="Endpoint" id="s3-endpoint">
						<Input
							id="s3-endpoint"
							bind:value={endpoint}
							placeholder="https://<account_id>.r2.cloudflarestorage.com"
						/>
					</FormField>
				</div>
				<div class="field-narrow">
					<FormField label="Регион" id="s3-region">
						<Input id="s3-region" bind:value={region} placeholder="auto" />
					</FormField>
				</div>
				<div class="field-narrow">
					<FormField label="Bucket" id="s3-bucket">
						<Input id="s3-bucket" bind:value={bucket} placeholder="piplos-backups" />
					</FormField>
				</div>
			</div>

			<div class="fields-row">
				<FormField label="Access Key ID" id="s3-access-key">
					<Input
						id="s3-access-key"
						bind:value={accessKeyId}
						placeholder={accessKeyPlaceholder}
						autocomplete="off"
					/>
				</FormField>
				<FormField label="Secret Access Key" id="s3-secret-key">
					<Input
						id="s3-secret-key"
						type="password"
						bind:value={secretAccessKey}
						placeholder={secretKeyPlaceholder}
						autocomplete="off"
					/>
				</FormField>
			</div>

			<label class="check">
				<input type="checkbox" bind:checked={usePathStyle} />
				Path-style адресация (MinIO и некоторые провайдеры; для R2 не требуется)
			</label>
		</div>

		<div class="actions-row">
			<form
				method="POST"
				action="?/updateS3"
				use:enhance={() => {
					savingS3 = true;
					return async ({ result }) => {
						savingS3 = false;
						if (result.type === 'success') {
							toast.success('Настройки S3 сохранены');
							await invalidateAll();
						} else if (result.type === 'failure') {
							toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
						}
					};
				}}
			>
				{@render s3FormFields()}
				<Button type="submit" loading={savingS3}>Сохранить</Button>
			</form>
			<form
				method="POST"
				action="?/testS3"
				class="test-form"
				use:enhance={() => {
					testingS3 = true;
					return async ({ result }) => {
						testingS3 = false;
						if (result.type === 'success') {
							toast.success('Подключение к S3 успешно');
						} else if (result.type === 'failure') {
							toast.error((result.data?.testError as string) ?? 'Ошибка подключения к S3');
						}
					};
				}}
			>
				{@render s3FormFields()}
				<Button type="submit" variant="secondary" loading={testingS3}>Тестировать</Button>
			</form>
		</div>
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
	.section-hint {
		margin: 0 0 1rem;
		font-size: 0.8125rem;
		color: #71717a;
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
	.field-wide {
		flex: 2;
		min-width: 14rem;
	}
	.field-narrow {
		flex: 1;
		min-width: 6rem;
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
	.test-form {
		margin-left: auto;
	}
</style>
