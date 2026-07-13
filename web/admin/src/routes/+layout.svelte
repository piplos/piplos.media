<script lang="ts">
	import '../app.css';
	import { Toaster } from 'svelte-french-toast';

	let { children }: { children: import('svelte').Snippet } = $props();
</script>

<Toaster position="bottom-right" />
{@render children()}
