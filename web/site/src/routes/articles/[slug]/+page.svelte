<script lang="ts">
	import { page } from '$app/state';
	import LangRedirect from '$lib/components/LangRedirect.svelte';
</script>

<LangRedirect path={`/articles/${page.params.slug}`} />
