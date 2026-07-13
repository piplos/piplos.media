<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import type { SmtpSettings } from './+page.server';

	let { data } = $props();

	const smtp = $derived(data.smtp as SmtpSettings);

	let saving = $state(false);
	let testing = $state(false);

	// Writable derived: значения из load, редактируемые локально; сбрасываются при invalidateAll.
	let host = $derived(smtp.host);
	let port = $derived(smtp.port);
	let username = $derived(smtp.username);
	let password = $derived(smtp.password);
	let from = $derived(smtp.from);
	let timeoutSeconds = $derived(smtp.timeoutSeconds);

	// Секрет не менялся -> отправляем "****", backend сохранит текущее значение.
	const usernameDirty = $derived(username !== smtp.username);
	const passwordDirty = $derived(password !== smtp.password);

	const usernamePlaceholder = $derived(smtp.usernameMasked ? '•••••••• (задан)' : '');
	const passwordPlaceholder = $derived(smtp.passwordMasked ? '•••••••• (задан)' : '');
</script>

<svelte:head>
	<title>SMTP — Настройки — Piplos Admin</title>
</svelte:head>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	<Card padding="sm">
		<h2 class="section-title">SMTP</h2>
		<p class="section-hint">
			Почтовый сервер для исходящих писем. Логин и пароль хранятся в БД в зашифрованном виде.
		</p>

		<div class="smtp-form">
			<div class="smtp-fields-row">
				<div class="smtp-field-host">
					<FormField label="Хост" id="smtp-host">
						<Input id="smtp-host" name="host" bind:value={host} placeholder="smtp.example.com" />
					</FormField>
				</div>
				<div class="smtp-field-narrow">
					<FormField label="Порт" id="smtp-port">
						<Input id="smtp-port" name="port" type="number" bind:value={port} />
					</FormField>
				</div>
				<div class="smtp-field-narrow">
					<FormField label="Таймаут (сек)" id="smtp-timeout">
						<Input id="smtp-timeout" name="timeoutSeconds" type="number" bind:value={timeoutSeconds} />
					</FormField>
				</div>
			</div>

			<div class="smtp-fields-row">
				<FormField label="Логин" id="smtp-username">
					<Input
						id="smtp-username"
						name="username"
						bind:value={username}
						placeholder={usernamePlaceholder}
						autocomplete="off"
					/>
				</FormField>
				<FormField label="Пароль" id="smtp-password">
					<Input
						id="smtp-password"
						name="password"
						type="password"
						bind:value={password}
						placeholder={passwordPlaceholder}
						autocomplete="off"
					/>
				</FormField>
			</div>

			<FormField label="От (From)" id="smtp-from">
				<Input id="smtp-from" name="from" bind:value={from} placeholder="noreply@piplos.media" />
			</FormField>
		</div>

		<div class="smtp-actions-row">
			<form
				method="POST"
				action="?/updateSmtp"
				use:enhance={() => {
					saving = true;
					return async ({ result }) => {
						saving = false;
						if (result.type === 'success') {
							toast.success('Настройки SMTP сохранены');
							await invalidateAll();
						} else if (result.type === 'failure') {
							toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
						}
					};
				}}
			>
				<input type="hidden" name="usernameDirty" value={usernameDirty ? 'true' : 'false'} />
				<input type="hidden" name="passwordDirty" value={passwordDirty ? 'true' : 'false'} />
				<input type="hidden" name="host" value={host} />
				<input type="hidden" name="port" value={port} />
				<input type="hidden" name="username" value={username} />
				<input type="hidden" name="password" value={password} />
				<input type="hidden" name="from" value={from} />
				<input type="hidden" name="timeoutSeconds" value={timeoutSeconds} />
				<Button type="submit" loading={saving}>Сохранить</Button>
			</form>
			<form
				method="POST"
				action="?/testSmtp"
				class="smtp-test-form"
				use:enhance={() => {
					testing = true;
					return async ({ result }) => {
						testing = false;
						if (result.type === 'success') {
							toast.success('Подключение к SMTP успешно');
						} else if (result.type === 'failure') {
							toast.error((result.data?.testError as string) ?? 'Ошибка подключения к SMTP');
						}
					};
				}}
			>
				<input type="hidden" name="usernameDirty" value={usernameDirty ? 'true' : 'false'} />
				<input type="hidden" name="passwordDirty" value={passwordDirty ? 'true' : 'false'} />
				<input type="hidden" name="host" value={host} />
				<input type="hidden" name="port" value={port} />
				<input type="hidden" name="username" value={username} />
				<input type="hidden" name="password" value={password} />
				<input type="hidden" name="from" value={from} />
				<input type="hidden" name="timeoutSeconds" value={timeoutSeconds} />
				<Button type="submit" variant="secondary" loading={testing}>Тестировать</Button>
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
	.smtp-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.smtp-fields-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: flex-end;
	}
	.smtp-fields-row > :global(*) {
		flex: 1;
		min-width: 0;
	}
	.smtp-field-host {
		flex: 1;
		min-width: 8rem;
	}
	.smtp-field-narrow {
		flex: 0 0 15%;
		min-width: 4rem;
	}
	.smtp-actions-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		margin-top: 1.25rem;
	}
	.smtp-test-form {
		margin-left: auto;
	}
</style>
