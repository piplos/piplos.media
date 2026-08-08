<script lang="ts">
	import { embedSelectionKey, splitBodySegments } from '$lib/embeds';
	import type { PortfolioProject } from '$lib/portfolio';
	import type { ServiceItem } from '$lib/services-api';
	import EmbedBlock from './EmbedBlock.svelte';
	import SafeHtml from './SafeHtml.svelte';

	interface Props {
		html: string;
		/** Выборки проектов по embedSelectionKey (дозированный API). */
		projects?: Record<string, PortfolioProject[]>;
		/** Выборки услуг по embedSelectionKey. */
		services?: Record<string, ServiceItem[]>;
		class?: string;
	}
	let { html, projects = {}, services = {}, class: className = '' }: Props = $props();

	const segments = $derived(splitBodySegments(html));
</script>

{#each segments as segment, i (i)}
	{#if segment.type === 'html'}
		<SafeHtml html={segment.html} class={className} />
	{:else}
		<EmbedBlock
			params={segment.params}
			projects={projects[embedSelectionKey(segment.params)] ?? []}
			services={services[embedSelectionKey(segment.params)] ?? []}
		/>
	{/if}
{/each}
