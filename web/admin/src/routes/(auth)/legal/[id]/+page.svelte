<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import { LEGAL_SLUG_LABELS } from '$lib/types';
	import LegalForm from '../LegalForm.svelte';

	let { data } = $props();

	const docTitle = $derived(LEGAL_SLUG_LABELS[data.page.slug] ?? data.page.slug);
</script>

<svelte:head>
	<title>{docTitle} — Права — Piplos Admin</title>
</svelte:head>

<AdminPage title={docTitle}>
	{#snippet actions()}
		<a href="/legal" class="back-link">← К списку</a>
	{/snippet}
	{#key data.page.id}
		<LegalForm page={data.page} languages={data.languages} submitLabel="Сохранить" />
	{/key}
</AdminPage>

<style>
	.back-link {
		font-size: 0.875rem;
		color: #71717a;
		text-decoration: none;
	}
	.back-link:hover {
		color: #2563eb;
	}
</style>
