<script lang="ts">
	import { themeStore } from '$lib/stores/theme.svelte';

	type Props = {
		slug: string;
	};

	let { slug }: Props = $props();

	const themeAware = new Set(['nextjs']);

	let src = $derived.by(() => {
		if (themeAware.has(slug) && themeStore.value === 'dark') {
			return `/stack/${slug}-light.svg`;
		}
		return `/stack/${slug}.svg`;
	});
</script>

<img class="stack-logo" {src} alt="" width="32" height="32" loading="lazy" decoding="async" />

<style>
	.stack-logo {
		display: block;
		width: 32px;
		height: 32px;
		object-fit: contain;
		flex-shrink: 0;
	}
</style>
