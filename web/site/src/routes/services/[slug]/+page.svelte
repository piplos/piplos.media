<script lang="ts">
	import { page } from '$app/state';
	import LangRedirect from '$lib/components/LangRedirect.svelte';
</script>

<LangRedirect path={`/services/${page.params.slug}`} />
