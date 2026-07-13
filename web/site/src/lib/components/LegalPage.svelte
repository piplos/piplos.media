<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import { getLegalSections, type LegalSectionKey } from '$lib/legal';

	interface Props {
		sectionKey: LegalSectionKey;
		metaTitleKey: string;
		metaDescriptionKey: string;
		labelKey: string;
		titleKey: string;
		lastUpdatedKey: string;
		breadcrumbKey: string;
		canonicalPath: string;
	}

	let {
		sectionKey,
		metaTitleKey,
		metaDescriptionKey,
		labelKey,
		titleKey,
		lastUpdatedKey,
		breadcrumbKey,
		canonicalPath
	}: Props = $props();

	const metaParams = { company: SITE.displayName };

	let sections = $derived.by(() => {
		langStore.value;
		return getLegalSections(sectionKey);
	});
</script>

<svelte:head>
	<title>{langStore.t(metaTitleKey, metaParams)}</title>
	<meta name="description" content={langStore.t(metaDescriptionKey, metaParams)} />
	<link rel="canonical" href="{SITE.url}{canonicalPath}" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t(breadcrumbKey)}</span>
	</div>
</nav>

<main id="main" class="legal-page">
	<div class="container">
		<p class="legal-label">{langStore.t(labelKey)}</p>
		<h1 class="legal-title">{langStore.t(titleKey)}</h1>
		<p class="legal-updated">{langStore.t(lastUpdatedKey)}</p>

		<div class="legal-sections">
			{#each sections as section (section.title)}
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
