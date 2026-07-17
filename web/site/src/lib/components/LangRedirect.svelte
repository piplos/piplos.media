<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { localizePath, resolveInitialLang } from '$lib/i18n/routing';
	import { SITE } from '$lib/site';

	interface Props {
		/** Путь без префикса языка, например `/portfolio` или `/portfolio/foo`. */
		path?: string;
	}

	let { path = '/' }: Props = $props();

	onMount(() => {
		goto(localizePath(path, resolveInitialLang()), { replaceState: true });
	});
</script>

<svelte:head>
	<title>{SITE.displayName}</title>
</svelte:head>

<main id="main">
	<p class="redirect-msg">Redirecting…</p>
</main>

<style>
	.redirect-msg {
		padding: 48px 24px;
		text-align: center;
		color: var(--c-muted);
		font-family: var(--f-mono);
		font-size: 13px;
	}
</style>
