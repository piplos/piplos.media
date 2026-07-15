<script lang="ts">
	import { page } from '$app/state';
	import LangRedirect from '$lib/components/LangRedirect.svelte';
</script>

<LangRedirect path={`/legal/${page.params.slug}`} />
