<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import { LEGAL_SLUG_LABELS } from '$lib/types';
	import LegalForm from '../LegalForm.svelte';

	let { data } = $props();

	const docTitle = $derived(LEGAL_SLUG_LABELS[data.page.slug] ?? data.page.slug);
	const breadcrumbs = $derived([
		{ label: 'Правовое', href: '/legal' },
		{ label: docTitle }
	]);
</script>

<svelte:head>
	<title>{docTitle} — Правовое — Piplos Admin</title>
</svelte:head>

<AdminPage title={docTitle} breadcrumbs={breadcrumbs}>
	{#key data.page.id}
		<LegalForm page={data.page} languages={data.languages} submitLabel="Сохранить" />
	{/key}
</AdminPage>
