<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ProjectForm from '../ProjectForm.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Проект: {data.project.slug} — Piplos Admin</title>
</svelte:head>

<AdminPage title="Редактирование проекта">
	{#snippet actions()}
		<a href="/projects" class="back-link">← К списку</a>
	{/snippet}
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
