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
	<img class="stack-logo" {src} alt="" width="32" height="32" loading="lazy" decoding="async" />
{/if}

<style>
	.stack-logo {
		display: block;
		width: 32px;
		height: 32px;
		object-fit: contain;
		flex-shrink: 0;
	}
</style>
