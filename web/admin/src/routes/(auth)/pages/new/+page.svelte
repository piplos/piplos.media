<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import PageForm from '../PageForm.svelte';

	let { data } = $props();

	const breadcrumbs = [
		{ label: 'Страницы', href: '/pages' },
		{ label: 'Новая страница' }
	];
</script>

<svelte:head>
	<title>Новая страница — Piplos Admin</title>
</svelte:head>

<AdminPage title="Новая страница" {breadcrumbs}>
	<PageForm languages={data.languages} stack={data.stack} submitLabel="Создать" />
</AdminPage>
