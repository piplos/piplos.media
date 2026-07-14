<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { fade, scale } from 'svelte/transition';
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import { nameDialogState, resolveNameDialog } from '$lib/name-dialog.svelte';

	let inputValue = $state('');
	let error = $state('');

	$effect(() => {
		if (nameDialogState.open) {
			inputValue = nameDialogState.value;
			error = '';
			queueMicrotask(() => document.getElementById('name-dialog-input')?.focus());
		}
	});

	function handleOpenChange(open: boolean) {
		if (!open && nameDialogState.open) resolveNameDialog(null);
	}

	function submit() {
		const name = inputValue.trim();
		if (!name) {
			error = 'Введите название';
			document.getElementById('name-dialog-input')?.focus();
			return;
		}
		resolveNameDialog(name);
	}
</script>

<Dialog.Root open={nameDialogState.open} onOpenChange={handleOpenChange}>
	<Dialog.Portal>
		<Dialog.Overlay forceMount>
			{#snippet child({ props: overlayProps, open })}
				{#if open}
					<div
						{...overlayProps}
						class="name-overlay"
						in:fade={{ duration: 150 }}
						out:fade={{ duration: 120 }}
					></div>
				{/if}
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content forceMount>
			{#snippet child({ props: contentProps, open })}
				{#if open}
					<div
						{...contentProps}
						class="name-panel"
						in:scale={{ duration: 150, start: 0.96 }}
						out:fade={{ duration: 120 }}
					>
						<Dialog.Title class="name-title">{nameDialogState.title}</Dialog.Title>
						<form
							class="name-form"
							onsubmit={(e) => {
								e.preventDefault();
								submit();
							}}
						>
							<FormField label={nameDialogState.label} id="name-dialog-input" error={error}>
								<Input
									id="name-dialog-input"
									bind:value={inputValue}
									placeholder={nameDialogState.placeholder}
									error={!!error}
									required
								/>
							</FormField>
							<div class="name-actions">
								<Button type="button" variant="secondary" onclick={() => resolveNameDialog(null)}>
									{nameDialogState.cancelLabel}
								</Button>
								<Button type="submit">{nameDialogState.confirmLabel}</Button>
							</div>
						</form>
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.name-overlay {
		position: fixed;
		inset: 0;
		z-index: 80;
		background: rgba(24, 24, 27, 0.45);
		backdrop-filter: blur(2px);
		-webkit-backdrop-filter: blur(2px);
	}
	.name-panel {
		position: fixed;
		top: 50%;
		left: 50%;
		z-index: 90;
		width: min(24rem, calc(100vw - 2rem));
		padding: 1.25rem;
		background: #fff;
		border-radius: 12px;
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.18);
		outline: none;
		transform: translate(-50%, -50%);
	}
	:global(.name-title) {
		margin: 0;
		font-size: 1.0625rem;
		font-weight: 600;
		color: #18181b;
	}
	.name-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 1rem;
	}
	.name-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}
</style>
