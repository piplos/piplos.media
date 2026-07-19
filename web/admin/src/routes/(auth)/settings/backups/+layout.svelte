<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const pathname = $derived(page.url.pathname);

	const subTabs = [
		{ href: '/settings/backups', label: 'Архивы' },
		{ href: '/settings/backups/schedule', label: 'Расписание' },
		{ href: '/settings/backups/s3', label: 'S3' }
	] as const;

	function isActive(href: (typeof subTabs)[number]['href']) {
		if (href === '/settings/backups') return pathname === '/settings/backups';
		return pathname === href || pathname.startsWith(href + '/');
	}
</script>

<div class="admin-sidebar-row">
	<nav class="admin-sidebar-nav" aria-label="Подразделы бекапов">
		{#each subTabs as tab (tab.href)}
			<a href={resolve(tab.href)} class:active={isActive(tab.href)}>{tab.label}</a>
		{/each}
	</nav>
	<div class="admin-sidebar-content admin-sidebar-content--no-box">
		{#key pathname}
			{@render children()}
		{/key}
	</div>
</div>
