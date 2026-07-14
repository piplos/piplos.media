<script lang="ts">
	import ProjectsList from '../ProjectsList.svelte';

	let { data } = $props();
</script>

<ProjectsList {data} />
