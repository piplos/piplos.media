<script lang="ts">
	import { page } from '$app/stores';
	import { DropdownMenu } from 'bits-ui';
	import { fly } from 'svelte/transition';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import NameDialog from '$lib/components/NameDialog.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();
	let logoutForm: HTMLFormElement | undefined = $state(undefined);

	const email = $derived(data.user?.email ?? '');
	const initial = $derived(email ? email[0].toUpperCase() : '?');
	const isAdmin = $derived(data.user?.role === 'admin');

	const navLinks = [
		{ href: '/leads', label: 'Заявки' },
		{ href: '/projects', label: 'Проекты' },
		{ href: '/services', label: 'Услуги' },
		{ href: '/stack', label: 'Стек' },
		{ href: '/files', label: 'Файлы' },
		{ href: '/pages', label: 'Страницы' }
	];

	const pathname = $derived($page.url.pathname);
	function isActive(href: string) {
		return pathname === href || pathname.startsWith(href + '/');
	}
	const isSettingsActive = $derived(isActive('/settings'));

	const GEAR_ICON_PATH = 'M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z';

	function handleLogout() {
		logoutForm?.requestSubmit();
	}
</script>

<svelte:head>
	<title>Piplos Admin</title>
</svelte:head>

<form
	id="logout-form"
	method="POST"
	action="/logout"
	class="hidden"
	bind:this={logoutForm}
	aria-hidden="true"
></form>

<div class="admin-layout">
	<header class="admin-header">
		<div class="admin-header-inner">
			<div class="admin-header-left">
				<Logo href="/" label="Piplos Media — главная" iconOnly />
				<nav class="admin-nav" aria-label="Основная навигация">
					{#each navLinks as link (link.href)}
						<a
							href={link.href}
							class="admin-nav-link"
							class:admin-nav-link--active={isActive(link.href)}
						>
							{link.label}
						</a>
					{/each}
				</nav>
			</div>
			<div class="admin-header-right">
				{#if isAdmin}
					<a
						href="/settings"
						class="admin-settings-icon"
						class:admin-settings-icon--active={isSettingsActive}
						aria-label="Настройки"
					>
						<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<circle cx="12" cy="12" r="3" />
							<path d={GEAR_ICON_PATH} />
						</svg>
					</a>
				{/if}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="user-trigger" aria-label="Меню аккаунта">
						<span class="user-avatar" aria-hidden="true">{initial}</span>
						<span class="user-trigger-email">{email}</span>
						<svg class="user-chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<polyline points="6 9 12 15 18 9" />
						</svg>
					</DropdownMenu.Trigger>
					<DropdownMenu.Portal>
						<DropdownMenu.Content side="bottom" sideOffset={8} align="end" forceMount>
							{#snippet child({ props: contentProps, wrapperProps, open })}
								{#if open}
									<div {...wrapperProps}>
										<div
											{...contentProps}
											class="user-content"
											in:fly={{ duration: 200, y: 6 }}
											out:fly={{ duration: 150, y: 6 }}
										>
											<div class="user-info">
												<div class="user-info-email">{email}</div>
												<div class="user-info-role">{data.user?.role === 'admin' ? 'Администратор' : 'Менеджер'}</div>
												{#if data.notifyLeads !== null}
													<div class="user-info-notify">
														<span
															class="user-notify-dot"
															class:user-notify-dot--on={data.notifyLeads}
															aria-hidden="true"
														></span>
														Письма о заявках: {data.notifyLeads ? 'включены' : 'выключены'}
													</div>
												{/if}
											</div>
											<DropdownMenu.Separator class="user-sep" />
											<DropdownMenu.Item
												class="user-item user-item-logout"
												textValue="Выйти"
												onSelect={handleLogout}
											>
												<svg class="user-item-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
													<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
													<polyline points="16 17 21 12 16 7" />
													<line x1="21" y1="12" x2="9" y2="12" />
												</svg>
												Выйти
											</DropdownMenu.Item>
										</div>
									</div>
								{/if}
							{/snippet}
						</DropdownMenu.Content>
					</DropdownMenu.Portal>
				</DropdownMenu.Root>
			</div>
		</div>
	</header>
	<main class="admin-main">
		{@render children()}
	</main>
</div>

<ConfirmDialog />
<NameDialog />

<style>
	.hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		border: 0;
	}
	.admin-layout {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		background: #fafafa;
	}
	.admin-header {
		background: #fff;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
		position: sticky;
		top: 0;
		z-index: 10;
	}
	.admin-header-inner {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0.75rem 1.5rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}
	.admin-header-left {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}
	.admin-nav {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.admin-nav-link {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #71717a;
		text-decoration: none;
		border-radius: 6px;
		transition: color 0.15s, background 0.15s;
	}
	.admin-nav-link:hover {
		color: #111;
		background: #d1d5db;
	}
	.admin-nav-link--active {
		color: #111;
		background: #e5e7eb;
	}
	.admin-header-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.admin-settings-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2.25rem;
		height: 2.25rem;
		color: #374151;
		border: none;
		border-radius: 10px;
		background: transparent;
		transition: background 0.15s, color 0.15s;
	}
	.admin-settings-icon:hover {
		background: #d1d5db;
		color: #111;
	}
	.admin-settings-icon--active {
		background: #e5e7eb;
		color: #111;
	}
	:global(.user-trigger) {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		min-height: 2.25rem;
		padding: 0.25rem 0.5rem 0.25rem 0.25rem;
		background: transparent;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		cursor: pointer;
		font-size: 0.875rem;
		color: #374151;
		transition: background 0.15s, border-color 0.15s;
	}
	:global(.user-trigger:hover) {
		background: #f9fafb;
		border-color: #d1d5db;
	}
	.user-avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.75rem;
		height: 1.75rem;
		border-radius: 7px;
		background: #111;
		color: #fff;
		font-size: 0.75rem;
		font-weight: 600;
	}
	.user-trigger-email {
		max-width: 12rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.user-chevron {
		flex-shrink: 0;
		opacity: 0.6;
	}
	:global(.user-content) {
		z-index: 100;
		min-width: 14rem;
		max-width: 20rem;
		padding: 0.25rem;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
		outline: none;
	}
	.user-info {
		padding: 0.5rem 0.75rem;
	}
	.user-info-email {
		font-size: 0.875rem;
		font-weight: 500;
		color: #18181b;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.user-info-role {
		font-size: 0.75rem;
		color: #71717a;
		margin-top: 0.125rem;
	}
	.user-info-notify {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		color: #71717a;
		margin-top: 0.375rem;
	}
	.user-notify-dot {
		flex-shrink: 0;
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: #d4d4d8;
	}
	.user-notify-dot--on {
		background: #16a34a;
	}
	:global(.user-sep) {
		height: 1px;
		background: #e5e7eb;
		margin: 0.25rem 0;
	}
	:global(.user-item) {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		width: 100%;
		padding: 0.5rem 0.75rem;
		font-size: 0.875rem;
		color: #374151;
		background: transparent;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		text-align: left;
		transition: background 0.15s;
	}
	:global(.user-item:hover),
	:global(.user-item[data-highlighted]) {
		background: #f3f4f6;
	}
	.user-item-icon {
		flex-shrink: 0;
		opacity: 0.6;
	}
	:global(.user-item-logout) {
		color: #b91c1c;
	}
	.admin-main {
		flex: 1;
		max-width: 1200px;
		width: 100%;
		margin: 0 auto;
		padding: 1.5rem;
		box-sizing: border-box;
	}
</style>
