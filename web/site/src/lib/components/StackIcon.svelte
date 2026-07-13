<script lang="ts">
	import { themeStore } from '$lib/stores/theme.svelte';
	import type { StackItemId } from '$lib/constants/stack';

	type Props = {
		id: StackItemId;
	};

	let { id }: Props = $props();

	const themeAware = new Set<StackItemId>(['nextjs']);

	let src = $derived.by(() => {
		if (themeAware.has(id) && themeStore.value === 'dark') {
			return `/stack/${id}-light.svg`;
		}
		return `/stack/${id}.svg`;
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
