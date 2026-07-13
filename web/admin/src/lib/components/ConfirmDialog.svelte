<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { fade, scale } from 'svelte/transition';
	import Button from '$lib/components/Button.svelte';
	import { confirmState, resolveConfirm } from '$lib/confirm.svelte';

	function handleOpenChange(open: boolean) {
		if (!open && confirmState.open) resolveConfirm(false);
	}
</script>

<Dialog.Root open={confirmState.open} onOpenChange={handleOpenChange}>
	<Dialog.Portal>
		<Dialog.Overlay forceMount>
			{#snippet child({ props: overlayProps, open })}
				{#if open}
					<div
						{...overlayProps}
						class="confirm-overlay"
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
						class="confirm-panel"
						in:scale={{ duration: 150, start: 0.96 }}
						out:fade={{ duration: 120 }}
					>
						<Dialog.Title class="confirm-title">{confirmState.title}</Dialog.Title>
						<Dialog.Description class="confirm-message">{confirmState.message}</Dialog.Description>
						<div class="confirm-actions">
							<Button variant="secondary" onclick={() => resolveConfirm(false)}>
								{confirmState.cancelLabel}
							</Button>
							<Button variant="danger" onclick={() => resolveConfirm(true)}>
								{confirmState.confirmLabel}
							</Button>
						</div>
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.confirm-overlay {
		position: fixed;
		inset: 0;
		z-index: 80;
		background: rgba(24, 24, 27, 0.45);
		backdrop-filter: blur(2px);
		-webkit-backdrop-filter: blur(2px);
	}
	.confirm-panel {
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
	:global(.confirm-title) {
		margin: 0;
		font-size: 1.0625rem;
		font-weight: 600;
		color: #18181b;
	}
	:global(.confirm-message) {
		margin: 0.5rem 0 0;
		font-size: 0.875rem;
		line-height: 1.5;
		color: #52525b;
	}
	.confirm-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 1.25rem;
	}
</style>
