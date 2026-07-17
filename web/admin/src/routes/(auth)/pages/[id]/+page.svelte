<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import PageForm from '../PageForm.svelte';

	let { data } = $props();

	const pageTitle = $derived.by(() => {
		const translations = data.page.translations ?? {};
		const langs = Object.keys(translations);
		const en = translations['en']?.title;
		const first = langs.length ? translations[langs[0]]?.title : '';
		return en || first || data.page.slug;
	});
	const breadcrumbs = $derived([
		{ label: 'Страницы', href: '/pages' },
		{ label: pageTitle }
	]);
</script>

<svelte:head>
	<title>{pageTitle} — Страницы — Piplos Admin</title>
</svelte:head>

<AdminPage title={pageTitle} {breadcrumbs}>
	{#key data.page.id}
		<PageForm page={data.page} seo={data.seo} languages={data.languages} stack={data.stack} submitLabel="Сохранить" />
	{/key}
</AdminPage>
