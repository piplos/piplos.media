<script lang="ts">
	import { page } from '$app/state';
	import Logo from '$lib/components/Logo.svelte';
	import ErrorPage from '$lib/components/ErrorPage.svelte';
	import { SITE } from '$lib/site';
</script>

<!-- Вне [lang]-контекста полноценные Header/Footer нелокализуемы, поэтому только логотип. -->
<div class="header-shell">
	<header class="header">
		<div class="container">
			<Logo href="/" label="{SITE.displayName} — home" />
		</div>
	</header>
</div>
<div id="page-wrap">
	<ErrorPage status={page.status} />
</div>
