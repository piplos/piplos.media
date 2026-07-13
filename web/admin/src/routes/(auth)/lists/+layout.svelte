<script lang="ts">
	import { page } from '$app/stores';
	import type { Snippet } from 'svelte';
	import { setTabLayoutContext } from '$lib/tab-layout.svelte';

	let { children } = $props();

	let pageActions = $state<Snippet | null>(null);

	setTabLayoutContext({
		setActions(actions) {
			pageActions = actions;
		}
	});

	const pathname = $derived($page.url.pathname);

	const tabs = [
		{ href: '/lists/services', label: 'Услуги', icon: 'services' },
		{ href: '/lists/stack', label: 'Стек', icon: 'stack' }
	];

	function isActive(href: string) {
		return pathname === href || pathname.startsWith(href + '/');
	}
</script>

<div class="admin-page lists-page">
	<h1 class="admin-page-title">Списки</h1>

	<nav class="settings-tabs-horizontal" aria-label="Разделы списков">
		<div class="settings-tabs-list">
			{#each tabs as tab (tab.href)}
				<a
					href={tab.href}
					class="settings-tab-link"
					class:settings-tab-link--active={isActive(tab.href)}
				>
					{#if tab.icon === 'services'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<rect x="3" y="3" width="7" height="7" rx="1" />
							<rect x="14" y="3" width="7" height="7" rx="1" />
							<rect x="14" y="14" width="7" height="7" rx="1" />
							<rect x="3" y="14" width="7" height="7" rx="1" />
						</svg>
					{:else if tab.icon === 'stack'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<polygon points="12 2 22 8.5 12 15 2 8.5 12 2" />
							<polyline points="2 15.5 12 22 22 15.5" />
						</svg>
					{/if}
					<span class="settings-tab-label">{tab.label}</span>
				</a>
			{/each}
		</div>
		{#if pageActions}
			<div class="settings-tabs-actions">
				{@render pageActions()}
			</div>
		{/if}
	</nav>

	<div class="lists-content">
		{@render children()}
	</div>
</div>

<style>
	.admin-page.lists-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.admin-page.lists-page .admin-page-title {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 700;
		color: #1a1a1a;
	}
	.settings-tabs-horizontal {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.25rem;
		border-radius: 10px;
		background: #f4f4f5;
		width: 100%;
		box-sizing: border-box;
	}
	.settings-tabs-list {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.25rem;
	}
	.settings-tabs-actions {
		flex-shrink: 0;
		margin-left: auto;
		padding-right: 0.25rem;
	}
	.settings-tab-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.875rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #71717a;
		background: transparent;
		border-radius: 8px;
		text-decoration: none;
		transition:
			color 0.15s,
			background 0.15s;
	}
	.settings-tab-icon {
		flex-shrink: 0;
		opacity: 0.7;
	}
	.settings-tab-link--active .settings-tab-icon {
		opacity: 0.9;
	}
	.settings-tab-link:hover {
		color: #1a1a1a;
	}
	.settings-tab-link--active {
		color: #1a1a1a;
		background: #fff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
	.settings-tab-link:focus,
	.settings-tab-link:focus-visible {
		outline: none;
	}
	.lists-content {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
</style>
