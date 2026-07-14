<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import type { LegalPage, LegalSlug } from '$lib/legal-api';
	import { resolveLegalDocument } from '$lib/legal';
	import { sanitizeCaseHtml } from '$lib/sanitize-html';

	interface Props {
		slug: LegalSlug;
		legalPages: LegalPage[];
	}

	let { slug, legalPages }: Props = $props();

	const document = $derived(resolveLegalDocument(slug, legalPages));

	function sectionBodyHtml(body: string): string {
		return sanitizeCaseHtml(body);
	}
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
					<div class="legal-body">{@html sectionBodyHtml(section.body)}</div>
				</article>
			{/each}
		</div>
	</div>
</main>

<style>
	.legal-body {
		font-size: 16px;
		color: var(--c-muted);
		line-height: 1.75;
	}

	.legal-body :global(p) {
		margin: 0 0 1rem;
	}

	.legal-body :global(p:last-child) {
		margin-bottom: 0;
	}

	.legal-body :global(ul),
	.legal-body :global(ol) {
		margin: 0 0 1rem 1.25rem;
		padding: 0;
	}

	.legal-body :global(a) {
		color: var(--c-accent);
		text-decoration: underline;
	}

	.legal-body :global(h2),
	.legal-body :global(h3),
	.legal-body :global(h4) {
		font-family: var(--f-display);
		color: var(--c-white);
		margin: 1.25rem 0 0.75rem;
	}

	.legal-body :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 1rem 0;
		border-radius: var(--radius);
	}
</style>
