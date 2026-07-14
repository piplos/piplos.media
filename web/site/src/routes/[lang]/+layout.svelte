<script lang="ts">
	import { onMount } from 'svelte';
	import Header from '$lib/components/Header.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { themeStore } from '$lib/stores/theme.svelte';
	import { langStore } from '$lib/stores/lang.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

	function syncLang() {
		if (data.lang !== langStore.value) {
			langStore.set(data.lang);
		}
	}

	// Вызов в теле скрипта нужен для SSR/prerender: $effect на сервере не выполняется,
	// без него /ru-страницы пререндерятся с английским контентом.
	syncLang();

	$effect(syncLang);

	onMount(() => {
		themeStore.init();
	});
</script>

<Header />
<div id="page-wrap">
	{@render children()}
</div>
<Footer services={data.footerServices} />
