<script lang="ts">
	import { page } from '$app/stores';
	import type { Snippet } from 'svelte';
	import AdminBreadcrumbs from '$lib/components/AdminBreadcrumbs.svelte';
	import { setTabLayoutContext } from '$lib/tab-layout.svelte';
	import { settingsBreadcrumbs } from './_settings';

	let { children } = $props();

	let pageActions = $state<Snippet | null>(null);

	setTabLayoutContext({
		setActions(actions) {
			pageActions = actions;
		}
	});

	const pathname = $derived($page.url.pathname);

	const tabs = [
		{ href: '/settings', label: 'Общие', icon: 'general' },
		{ href: '/settings/ai', label: 'AI-переводчик', icon: 'ai' },
		{ href: '/settings/smtp', label: 'SMTP', icon: 'mail' },
		{ href: '/settings/users', label: 'Пользователи', icon: 'users' }
	];

	function isActive(href: string) {
		if (href === '/settings') return pathname === '/settings';
		return pathname === href || pathname.startsWith(href + '/');
	}

	const aiSubTabs = [
		{ href: '/settings/ai', label: 'Провайдеры' },
		{ href: '/settings/ai/translation', label: 'Перевод' }
	];

	const showAiSidebar = $derived(
		pathname === '/settings/ai' || pathname.startsWith('/settings/ai/')
	);

	function isAiSubActive(href: string) {
		if (href === '/settings/ai') return pathname === '/settings/ai';
		return pathname === href || pathname.startsWith(href + '/');
	}

	const breadcrumbs = $derived(settingsBreadcrumbs(pathname));
</script>

<div class="admin-page settings-page">
	<AdminBreadcrumbs items={breadcrumbs} />

	<nav class="settings-tabs-horizontal" aria-label="Разделы настроек">
		<div class="settings-tabs-list">
			{#each tabs as tab (tab.href)}
				<a
					href={tab.href}
					class="settings-tab-link"
					class:settings-tab-link--active={isActive(tab.href)}
				>
					{#if tab.icon === 'general'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<circle cx="12" cy="12" r="3" />
							<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
						</svg>
					{:else if tab.icon === 'ai'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z" />
						</svg>
					{:else if tab.icon === 'mail'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
							<polyline points="22,6 12,13 2,6" />
						</svg>
					{:else if tab.icon === 'users'}
						<svg class="settings-tab-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
							<circle cx="9" cy="7" r="4" />
							<path d="M23 21v-2a4 4 0 0 0-3-3.87" />
							<path d="M16 3.13a4 4 0 0 1 0 7.75" />
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

	{#if showAiSidebar}
		<div class="admin-sidebar-row">
			<nav class="admin-sidebar-nav" aria-label="Подразделы AI">
				{#each aiSubTabs as tab (tab.href)}
					<a href={tab.href} class:active={isAiSubActive(tab.href)}>
						{tab.label}
					</a>
				{/each}
			</nav>
			<div class="admin-sidebar-content admin-sidebar-content--no-box">
				{@render children()}
			</div>
		</div>
	{:else}
		<div class="settings-content">
			{@render children()}
		</div>
	{/if}
</div>

<style>
	.admin-page.settings-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
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
	.settings-content {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
</style>
