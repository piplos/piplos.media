<script lang="ts">
	import type { BreadcrumbItem } from '$lib/breadcrumbs';

	let { items }: { items: BreadcrumbItem[] } = $props();
</script>

<nav class="admin-breadcrumbs" aria-label="Breadcrumb">
	<ol class="admin-breadcrumbs-list">
		{#each items as crumb, index (crumb.label + (crumb.href ?? ''))}
			<li class="admin-breadcrumb-item">
				{#if crumb.href}
					<a href={crumb.href}>{crumb.label}</a>
				{:else}
					<h1 class="admin-breadcrumb-current">{crumb.label}</h1>
				{/if}
				{#if index < items.length - 1}
					<span class="admin-breadcrumb-sep" aria-hidden="true">/</span>
				{/if}
			</li>
		{/each}
	</ol>
</nav>
