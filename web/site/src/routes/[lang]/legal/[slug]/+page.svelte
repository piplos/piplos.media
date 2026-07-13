<script lang="ts">
	import LegalPage from '$lib/components/LegalPage.svelte';

	let { data } = $props();
</script>

<LegalPage slug={data.slug} legalPages={data.legalPages} />
