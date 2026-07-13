<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import ProjectForm from '../ProjectForm.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Новый проект — Piplos Admin</title>
</svelte:head>

<AdminPage title="Новый проект">
	{#snippet actions()}
		<a href="/projects" class="back-link">← К списку</a>
	{/snippet}
	<ProjectForm
		languages={data.languages}
		services={data.services}
		stack={data.stack}
		submitLabel="Создать проект"
	/>
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
