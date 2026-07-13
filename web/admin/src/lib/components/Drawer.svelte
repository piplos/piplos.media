<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { fade, fly } from 'svelte/transition';

	interface Props {
		open?: boolean;
		title: string;
		children?: import('svelte').Snippet;
	}
	let { open = $bindable(false), title, children }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay forceMount>
			{#snippet child({ props: overlayProps, open: isOpen })}
				{#if isOpen}
					<div
						{...overlayProps}
						class="drawer-overlay"
						in:fade={{ duration: 200 }}
						out:fade={{ duration: 150 }}
					></div>
				{/if}
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content forceMount>
			{#snippet child({ props: contentProps, open: isOpen })}
				{#if isOpen}
					<div
						{...contentProps}
						class="drawer-panel"
						in:fly={{ duration: 250, x: 480, opacity: 1 }}
						out:fly={{ duration: 200, x: 480, opacity: 1 }}
					>
						<header class="drawer-head">
							<Dialog.Title class="drawer-title">{title}</Dialog.Title>
							<Dialog.Close class="drawer-close" aria-label="Закрыть">
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
									<line x1="18" y1="6" x2="6" y2="18" />
									<line x1="6" y1="6" x2="18" y2="18" />
								</svg>
							</Dialog.Close>
						</header>
						<div class="drawer-body">
							{@render children?.()}
						</div>
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.drawer-overlay {
		position: fixed;
		inset: 0;
		z-index: 60;
		background: rgba(24, 24, 27, 0.35);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
	}
	.drawer-panel {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		z-index: 70;
		width: min(34rem, 100vw);
		display: flex;
		flex-direction: column;
		background: #fff;
		box-shadow: -12px 0 40px rgba(0, 0, 0, 0.15);
		outline: none;
	}
	.drawer-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid #e5e7eb;
		flex-shrink: 0;
	}
	:global(.drawer-title) {
		margin: 0;
		font-size: 1.0625rem;
		font-weight: 600;
		color: #18181b;
	}
	:global(.drawer-close) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		padding: 0;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	:global(.drawer-close:hover) {
		background: #f4f4f5;
		color: #18181b;
	}
	.drawer-body {
		flex: 1;
		overflow-y: auto;
		padding: 1.25rem;
	}
</style>
