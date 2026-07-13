<script lang="ts">
	import { useTabLayout } from '$lib/tab-layout.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	const layout = useTabLayout();

	$effect(() => {
		layout.setActions(children);
		return () => layout.setActions(null);
	});
</script>
