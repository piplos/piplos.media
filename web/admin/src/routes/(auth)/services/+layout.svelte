<script lang="ts">
	import { page } from '$app/stores';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Button from '$lib/components/Button.svelte';

	let { data, children } = $props();

	const isOverview = $derived($page.url.pathname === '/services');
</script>

{#if isOverview}
	<AdminPage title="Услуги" breadcrumbs={[{ label: 'Услуги' }]}>
		{#snippet actions()}
			<Button variant="success" onclick={() => (location.href = '/services/new')}>+ Новая услуга</Button>
		{/snippet}
		{@render children()}
	</AdminPage>
{:else}
	{@render children()}
{/if}
