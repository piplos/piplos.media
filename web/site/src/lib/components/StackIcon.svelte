<script lang="ts">
	import { resolveUploadUrl } from '$lib/api';
	import { themeStore } from '$lib/stores/theme.svelte';

	type Props = {
		icon: string;
		iconAlt?: string;
	};

	let { icon, iconAlt = '' }: Props = $props();

	const src = $derived.by(() => {
		const path = themeStore.value === 'dark' && iconAlt ? iconAlt : icon;
		return path ? resolveUploadUrl(path) : '';
	});
</script>

{#if src}
	<img class="stack-logo" {src} alt="" width="24" height="24" loading="lazy" decoding="async" />
{/if}

<style>
	.stack-logo {
		display: block;
		width: 24px;
		height: 24px;
		object-fit: contain;
		flex-shrink: 0;
	}
</style>
