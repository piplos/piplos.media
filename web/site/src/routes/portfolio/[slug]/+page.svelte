<script lang="ts">
	import { page } from '$app/state';
	import LangRedirect from '$lib/components/LangRedirect.svelte';
</script>

<LangRedirect path={`/portfolio/${page.params.slug}`} />
