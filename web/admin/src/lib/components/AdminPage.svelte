<script lang="ts">
	import AdminBreadcrumbs from './AdminBreadcrumbs.svelte';
	import type { BreadcrumbItem } from '$lib/breadcrumbs';

	interface Props {
		title: string;
		description?: string;
		breadcrumbs?: BreadcrumbItem[];
		actions?: import('svelte').Snippet;
		children?: import('svelte').Snippet;
	}
	let { title, description = '', breadcrumbs, actions, children }: Props = $props();
</script>

<div class="admin-page">
	<div class="admin-page-head">
		<div>
			{#if breadcrumbs?.length}
				<AdminBreadcrumbs items={breadcrumbs} />
			{:else}
				<h1 class="admin-page-title">{title}</h1>
			{/if}
			{#if description}
				<p class="admin-page-lead">{description}</p>
			{/if}
		</div>
		{#if actions}
			<div class="admin-page-actions">
				{@render actions()}
			</div>
		{/if}
	</div>
	{#if children}
		{@render children()}
	{/if}
</div>

<style>
	.admin-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.admin-page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}
	.admin-page-lead {
		margin: 0.25rem 0 0;
		font-size: 1rem;
		color: #555;
	}
	.admin-page-actions {
		flex-shrink: 0;
	}
</style>
