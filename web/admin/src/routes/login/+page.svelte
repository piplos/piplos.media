<script lang="ts">
	import { enhance } from '$app/forms';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import toast from 'svelte-french-toast';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	let submitting = $state(false);
	let emailValue = $state('');
	let lastError = $state<string | null>(null);

	$effect(() => {
		if (form?.email !== undefined) emailValue = form.email;
	});
	const error = $derived(form?.error ?? data.urlError ?? null);
	$effect(() => {
		if (error != null && error !== lastError) {
			lastError = error;
			toast.error(error);
		}
		if (error == null) lastError = null;
	});
</script>

<svelte:head>
	<title>Вход — Piplos Admin</title>
</svelte:head>

<div class="login-wrap">
	<div class="card-login">
		<Card>
			{#if error}
				<div class="login-error" role="alert">{error}</div>
			{/if}
			<form
				method="POST"
				action="?/login"
				class="form"
				use:enhance={() => {
					submitting = true;
					return async ({ update }) => {
						await update();
						submitting = false;
					};
				}}
			>
				<FormField label="Email" id="email">
					<Input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={emailValue}
						placeholder="name@example.com"
						error={!!error}
						disabled={submitting}
					/>
				</FormField>
				<FormField label="Пароль" id="password">
					<Input
						id="password"
						name="password"
						type="password"
						autocomplete="current-password"
						required
						placeholder="••••••••"
						error={!!error}
						disabled={submitting}
					/>
				</FormField>
				<div class="submit-wrap">
					<Button type="submit" loading={submitting} fullWidth>Войти</Button>
				</div>
			</form>
		</Card>
	</div>
</div>

<style>
	.login-wrap {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #fafafa;
		padding: 1.5rem;
	}
	.card-login {
		width: 100%;
		max-width: 24rem;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 1.375rem;
	}
	.submit-wrap {
		margin-top: 0.25rem;
	}
	.login-error {
		margin-bottom: 1rem;
		padding: 0.75rem 1rem;
		background: #fef2f2;
		color: #b91c1c;
		border-radius: 0.5rem;
		font-size: 0.875rem;
	}
</style>
