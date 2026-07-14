<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ServiceForm from '../ServiceForm.svelte';
	import { serviceEditBreadcrumbs, serviceTitle } from '../_services';

	let { data } = $props();

	const serviceLabel = $derived(data.service ? serviceTitle(data.service) : 'Редактирование услуги');
	const breadcrumbs = $derived(serviceEditBreadcrumbs(serviceLabel));
</script>

<svelte:head>
	<title>Редактирование услуги — Piplos Admin</title>
</svelte:head>

<AdminPage title={serviceLabel} breadcrumbs={breadcrumbs}>
	{#if data.service}
		{#key data.service.slug}
			<ServiceForm
				service={data.service}
				seo={data.seo}
				languages={data.languages}
				stack={data.stack}
				submitLabel="Сохранить"
			/>
		{/key}
	{/if}
</AdminPage>
