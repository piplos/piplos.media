<script lang="ts">
	import LeadsList from '../LeadsList.svelte';

	let { data } = $props();
</script>

<LeadsList {data} />
