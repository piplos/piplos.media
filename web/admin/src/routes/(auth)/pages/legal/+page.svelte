<script lang="ts">
	import PagesList from '../PagesList.svelte';

	let { data } = $props();
</script>

<PagesList {data} />
