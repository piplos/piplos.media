<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import { resolveUploadUrl } from '$lib/api';
	import { getProjectLocale, getProjectStackItems, type PortfolioProject } from '$lib/portfolio';
	import { extractAndStripProjectLinks } from '$lib/project-links';
	import ProjectLinkIcon from '$lib/components/ProjectLinkIcon.svelte';
	import SafeHtml from '$lib/components/SafeHtml.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let project = $derived(data.project as PortfolioProject);
	let loc = $derived(getProjectLocale(project, langStore.value));
	let stackItems = $derived(getProjectStackItems(project, langStore.value));
	let solutionParts = $derived(extractAndStripProjectLinks(loc.solution));
	let projectLinks = $derived(solutionParts.links);

	let seoLoc = $derived(data.seo?.[langStore.value]);
	let seoTitle = $derived(seoLoc?.title || `${loc.title} — Case Study | ${SITE.name}`);
	let seoDescription = $derived(seoLoc?.description || loc.description);
	let ogTitle = $derived(seoLoc?.og_title || seoTitle);
	let ogDescription = $derived(seoLoc?.og_description || seoDescription);
	let ogImage = $derived(resolveUploadUrl(seoLoc?.og_image || project.image || ''));
	let canonicalUrl = $derived(`${SITE.url}${l(`/portfolio/${project.id}`)}`);
</script>

<svelte:head>
	<title>{seoTitle}</title>
	<meta name="description" content={seoDescription} />
	<link rel="canonical" href={canonicalUrl} />
	<link rel="alternate" hreflang="en" href="{SITE.url}/en/portfolio/{project.id}" />
	<link rel="alternate" hreflang="ru" href="{SITE.url}/ru/portfolio/{project.id}" />
	<link rel="alternate" hreflang="x-default" href="{SITE.url}/en/portfolio/{project.id}" />
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content={SITE.displayName} />
	<meta property="og:locale" content={langStore.value === 'ru' ? 'ru_RU' : 'en_US'} />
	<meta property="og:title" content={ogTitle} />
	<meta property="og:description" content={ogDescription} />
	<meta property="og:url" content={canonicalUrl} />
	{#if ogImage}
		<meta property="og:image" content={ogImage} />
	{/if}
	<meta name="twitter:card" content={ogImage ? 'summary_large_image' : 'summary'} />
	<meta name="twitter:title" content={ogTitle} />
	<meta name="twitter:description" content={ogDescription} />
	{#if ogImage}
		<meta name="twitter:image" content={ogImage} />
	{/if}
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href={l('/')}>{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<a href={l('/portfolio')}>{langStore.t('nav.portfolio')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{loc.title}</span>
	</div>
</nav>

<main id="main">

	<section class="cs-hero" aria-labelledby="cs-title">
		<div class="container">
			<a href={l('/portfolio')} class="cs-back">
				<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
					<path d="M7 2L3 6l4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
				{langStore.t('case_study.back')}
			</a>

			<h1 class="cs-title" id="cs-title">{loc.title}</h1>
			<p class="cs-desc">{loc.description}</p>
		</div>
	</section>

	<section class="cs-content">
		<div class="container">
			<div class="cs-layout">

				<div class="cs-main">

					<article class="cs-block">
						<h2 class="cs-heading">{langStore.t('case_study.challenge')}</h2>
						<p class="cs-text">{loc.challenge}</p>
					</article>

					<article class="cs-block">
						<h2 class="cs-heading">{langStore.t('case_study.solution')}</h2>
						{#if solutionParts.html.trim()}
							<SafeHtml html={solutionParts.html} class="cs-text cs-rich" />
						{:else}
							<p class="cs-text"></p>
						{/if}
					</article>

					<article class="cs-block">
						<h2 class="cs-heading">{langStore.t('case_study.result')}</h2>
						<p class="cs-text">{loc.result}</p>
					</article>
				</div>

				<aside class="cs-sidebar">

					{#if projectLinks.length}
						<div class="cs-card">
							<p class="cs-label">{langStore.t('case_study.links')}</p>
							<div class="cs-links">
								{#each projectLinks as link (link.url)}
									<a
										href={link.url}
										class="cs-link"
										target="_blank"
										rel="noopener noreferrer"
									>
										<span class="cs-link-main">
											<ProjectLinkIcon kind={link.kind} />
											<span class="cs-link-label">{link.label}</span>
										</span>
										<svg class="cs-link-arrow" width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
											<path d="M3.5 8.5L8.5 3.5M8.5 3.5H4.5M8.5 3.5V7.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
										</svg>
									</a>
								{/each}
							</div>
						</div>
					{/if}

					<div class="cs-card">
						<p class="cs-label">{langStore.t('case_study.stack')}</p>
						<div class="cs-stack-tags">
							{#each stackItems as item (item)}
								<span class="cs-stack-tag">{item}</span>
							{/each}
						</div>
					</div>

					<div class="cs-card">
						<p class="cs-label">Project Info</p>
						<dl class="cs-info">
							<div>
								<dt>Year</dt>
								<dd>{project.year}</dd>
							</div>
							<div>
								<dt>Category</dt>
								<dd class="capitalize">{project.category}</dd>
							</div>
						</dl>
					</div>

					<a href="{l('/order')}?from={project.id}" class="cs-cta">
						{langStore.t('case_study.start_project')}
						<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
							<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</a>
				</aside>
			</div>
		</div>
	</section>

</main>

<style>
	.cs-hero {
		padding: 64px 0;
		border-bottom: 1px solid var(--c-border);
	}

	.cs-back {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-family: var(--f-mono);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-muted);
		margin-bottom: 32px;
		transition: color 0.2s;
	}

	.cs-back:hover { color: var(--c-white); }

	.cs-title {
		font-family: var(--f-display);
		font-size: clamp(40px, 6vw, 72px);
		font-weight: 700;
		color: var(--c-white);
		letter-spacing: -0.02em;
		line-height: 1.05;
		margin-bottom: 24px;
	}

	.cs-desc {
		font-size: 18px;
		color: var(--c-muted);
		line-height: 1.7;
		max-width: 720px;
	}

	.cs-content { padding: 80px 0 100px; }

	.cs-layout {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: 64px;
		align-items: start;
	}

	.cs-main { display: flex; flex-direction: column; gap: 56px; }

	.cs-label {
		font-family: var(--f-mono);
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--c-accent);
		margin-bottom: 16px;
	}

	.cs-heading {
		font-family: var(--f-display);
		font-size: 28px;
		font-weight: 600;
		color: var(--c-white);
		letter-spacing: -0.01em;
		margin-bottom: 16px;
	}

	.cs-text {
		font-size: 16px;
		color: var(--c-muted);
		line-height: 1.75;
	}

	.cs-rich :global(p) {
		margin: 0 0 1rem;
	}

	.cs-rich :global(p:last-child) {
		margin-bottom: 0;
	}

	.cs-rich :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 1.25rem 0;
		border-radius: var(--radius);
	}

	.cs-rich :global(ul),
	.cs-rich :global(ol) {
		margin: 0 0 1rem 1.25rem;
		padding: 0;
	}

	.cs-rich :global(ul) { list-style: disc; }
	.cs-rich :global(ol) { list-style: decimal; }

	.cs-rich :global(li::marker) {
		color: var(--c-accent);
	}

	.cs-rich :global(a) {
		color: var(--c-accent);
		text-decoration: underline;
	}

	.cs-rich :global(h2),
	.cs-rich :global(h3) {
		font-family: var(--f-display);
		color: var(--c-white);
		margin: 1.5rem 0 0.75rem;
	}

	.cs-sidebar {
		display: flex;
		flex-direction: column;
		gap: 24px;
		position: sticky;
		top: calc(var(--nav-h) + 24px);
	}

	.cs-card {
		background: var(--c-surface);
		border: 1px solid var(--c-border2);
		border-radius: var(--radius);
		padding: 28px 24px;
	}

	.cs-links {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.cs-link {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 10px 12px;
		font-size: 13px;
		font-weight: 500;
		color: var(--c-text);
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
		transition: color 0.2s, border-color 0.2s, background 0.2s;
	}

	.cs-link:hover {
		color: var(--c-white);
		border-color: var(--c-accent);
		background: color-mix(in srgb, var(--c-accent) 12%, transparent);
	}

	.cs-link-main {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.cs-link-arrow {
		flex-shrink: 0;
		color: var(--c-dim);
	}

	.cs-link-label {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.cs-stack-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.cs-stack-tag {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-muted);
		border: 1px solid var(--c-border);
		padding: 4px 10px;
		border-radius: 100px;
	}

	.cs-info {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.cs-info dt {
		font-family: var(--f-mono);
		font-size: 10px;
		color: var(--c-dim);
		letter-spacing: 0.2em;
		text-transform: uppercase;
		margin-bottom: 4px;
	}

	.cs-info dd {
		font-size: 14px;
		font-weight: 500;
		color: var(--c-text);
	}

	.cs-cta {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
		padding: 16px 24px;
		font-family: var(--f-mono);
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #fff;
		background: var(--c-accent);
		border-radius: var(--radius);
		transition: opacity 0.2s, transform 0.2s;
	}

	.cs-cta:hover { opacity: 0.88; transform: translateY(-1px); }

	:global([data-theme="light"]) .cs-cta { color: #fff; }

	@media (max-width: 1024px) {
		.cs-layout { grid-template-columns: 1fr; gap: 48px; }
		.cs-sidebar { position: static; }
	}

	@media (max-width: 768px) {
		.cs-hero { padding: 48px 0; }
		.cs-content { padding: 56px 0 80px; }
	}
</style>
