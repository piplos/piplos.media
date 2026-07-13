<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import type { LegalPage, LegalSlug } from '$lib/legal-api';
	import { resolveLegalDocument } from '$lib/legal';

	interface Props {
		slug: LegalSlug;
		legalPages: LegalPage[];
	}

	let { slug, legalPages }: Props = $props();

	const document = $derived(resolveLegalDocument(slug, legalPages));
</script>

<svelte:head>
	<title>{document.title}</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href={l('/')}>{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t(`footer.links.${slug}`)}</span>
	</div>
</nav>

<main id="main" class="legal-page">
	<div class="container">
		<p class="legal-label">{document.label}</p>
		<h1 class="legal-title">{document.title}</h1>
		<p class="legal-updated">{document.lastUpdated}</p>

		<div class="legal-sections">
			{#each document.sections as section (section.title)}
				<article class="legal-section">
					<h2>{section.title}</h2>
					{#each section.body.split('\n\n') as paragraph, i (i)}
						<p>{paragraph}</p>
					{/each}
				</article>
			{/each}
		</div>
	</div>
</main>
