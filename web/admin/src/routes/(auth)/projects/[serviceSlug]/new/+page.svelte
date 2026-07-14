<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ProjectForm from '../../ProjectForm.svelte';
	import { ORPHAN_URL_SLUG, projectEditBreadcrumbs, serviceTitle } from '../../_projects';

	let { data } = $props();

	const service = $derived(data.services.find((entry) => entry.slug === data.serviceSlug));
	const serviceLabel = $derived(
		data.serviceSlug === ORPHAN_URL_SLUG
			? 'Без группы'
			: service
				? serviceTitle(service)
				: data.serviceSlug
	);
	const breadcrumbs = $derived(
		projectEditBreadcrumbs(serviceLabel, `/projects/${data.serviceSlug}`, 'Новый проект')
	);
</script>

<svelte:head>
	<title>Новый проект — Piplos Admin</title>
</svelte:head>

<AdminPage title="Новый проект" breadcrumbs={breadcrumbs}>
	<ProjectForm
		project={{ category: data.serviceSlug }}
		languages={data.languages}
		services={data.services}
		stack={data.stack}
		submitLabel="Создать проект"
	/>
</AdminPage>
