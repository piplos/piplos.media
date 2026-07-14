<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ProjectForm from '../../ProjectForm.svelte';
	import {
		ORPHAN_URL_SLUG,
		projectEditBreadcrumbs,
		projectTitle,
		serviceTitle
	} from '../../_projects';

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
		projectEditBreadcrumbs(
			serviceLabel,
			`/projects/${data.serviceSlug}`,
			projectTitle(data.project)
		)
	);
</script>

<svelte:head>
	<title>Проект: {data.project.slug} — Piplos Admin</title>
</svelte:head>

<AdminPage title={projectTitle(data.project)} breadcrumbs={breadcrumbs}>
	{#key data.project.slug}
		<ProjectForm
			project={data.project}
			seo={data.seo}
			languages={data.languages}
			services={data.services}
			stack={data.stack}
			submitLabel="Сохранить"
		/>
	{/key}
</AdminPage>
