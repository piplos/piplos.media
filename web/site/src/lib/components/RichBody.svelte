<script lang="ts">
	import { splitBodySegments } from '$lib/embeds';
	import type { PortfolioProject } from '$lib/portfolio';
	import type { ServiceItem } from '$lib/services-api';
	import EmbedBlock from './EmbedBlock.svelte';
	import SafeHtml from './SafeHtml.svelte';

	interface Props {
		html: string;
		/** Данные для embed-блоков; пустые массивы — блоки просто не выводятся. */
		projects?: PortfolioProject[];
		services?: ServiceItem[];
		class?: string;
	}
	let { html, projects = [], services = [], class: className = '' }: Props = $props();

	const segments = $derived(splitBodySegments(html));
</script>

{#each segments as segment, i (i)}
	{#if segment.type === 'html'}
		<SafeHtml html={segment.html} class={className} />
	{:else}
		<EmbedBlock params={segment.params} {projects} {services} />
	{/if}
{/each}
