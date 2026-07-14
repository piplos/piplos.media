<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ServiceForm from '../ServiceForm.svelte';

	let { data } = $props();

	const breadcrumbs = [{ label: 'Услуги', href: '/services' }, { label: 'Новая услуга' }];
</script>

<svelte:head>
	<title>Новая услуга — Piplos Admin</title>
</svelte:head>

<AdminPage title="Новая услуга" breadcrumbs={breadcrumbs}>
	<ServiceForm languages={data.languages} stack={data.stack} submitLabel="Создать услугу" />
</AdminPage>
