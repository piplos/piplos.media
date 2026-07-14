<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { getCategoryColor, getProjectLocale } from '$lib/portfolio';
	import { sanitizeCaseHtml } from '$lib/sanitize-html';
	import { SITE } from '$lib/site';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const service = $derived(data.service);
	const related = $derived(data.related);
	const bodyHtml = $derived(sanitizeCaseHtml(service.body));
</script>

<svelte:head>
	<title>{service.title} — {langStore.t('services.title')} | {SITE.name}</title>
	<meta name="description" content={service.description} />
	<link rel="canonical" href="{SITE.url}{l(`/services/${service.slug}`)}" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href={l('/')}>{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<a href={l('/#services')}>{langStore.t('services.title')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{service.title}</span>
	</div>
</nav>

<main id="main">
	<section class="svc-hero" aria-labelledby="svc-title">
		<div class="container">
			<a href={l('/#services')} class="svc-back">
				<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
					<path d="M7 2L3 6l4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
				{langStore.t('service_page.back')}
			</a>

			<div class="svc-hero-row">
				<div class="svc-icon" aria-hidden="true">{service.icon}</div>
				<div>
					<h1 class="svc-title" id="svc-title">{service.title}</h1>
					<p class="svc-desc">{service.description}</p>
				</div>
			</div>
		</div>
	</section>

	<section class="svc-content">
		<div class="container">
			<div class="svc-layout">
				<div class="svc-main">
					{#if bodyHtml}
						<div class="svc-block">
							<div class="svc-rich">{@html bodyHtml}</div>
						</div>
					{/if}
				</div>

				<aside class="svc-sidebar">
					{#if service.tags.length > 0}
						<div class="svc-card">
							<p class="section-label">{langStore.t('service_page.stack')}</p>
							<div class="svc-tags">
								{#each service.tags as tag (tag)}
									<span class="svc-tag">{tag}</span>
								{/each}
							</div>
						</div>
					{/if}

					<a href="{l('/order')}?type={service.slug}" class="svc-cta">
						{langStore.t('service_page.start_project')}
						<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
							<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</a>
				</aside>
			</div>
		</div>
	</section>

	{#if related.length > 0}
		<section class="section" style="padding-top: 0" aria-labelledby="svc-projects-heading">
			<div class="container">
				<div class="section-header">
					<div>
						<p class="section-label">{langStore.t('service_page.projects')}</p>
						<h2 class="section-title" id="svc-projects-heading">{langStore.t('work.title')}</h2>
					</div>
					<a href={l('/portfolio')} class="section-link" aria-label="View full portfolio">
						{langStore.t('work.cta')}
						<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
							<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</a>
				</div>
				<div class="work-grid" role="list">
					{#each related as project (project.id)}
						{@const loc = getProjectLocale(project, langStore.value)}
						<article class="work-card" role="listitem" itemscope itemtype="https://schema.org/CreativeWork">
							{#if project.image}
								<div class="pc-bg" aria-hidden="true">
									<img src={project.image} alt="" loading="lazy" />
								</div>
							{/if}
							<div class="work-type">
								<span class="work-type-dot" style="background:{getCategoryColor(project.category)}" aria-hidden="true"></span>
								{loc.subtitle}
							</div>
							<h3 class="work-title" itemprop="name">
								<a href={l(`/portfolio/${project.id}`)} class="work-title-link" aria-label="View {loc.title} case study">{loc.title}</a>
							</h3>
							<p class="work-desc" itemprop="description">{loc.description}</p>
							<a href={l(`/portfolio/${project.id}`)} class="work-link" itemprop="url" aria-label="View {loc.title} case study">
								{langStore.t('work.case_study')}
								<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
									<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
								</svg>
							</a>
						</article>
					{/each}
				</div>
			</div>
		</section>
	{/if}
</main>

<style>
	.svc-hero {
		padding: 64px 0;
		border-bottom: 1px solid var(--c-border);
	}

	.svc-back {
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

	.svc-back:hover { color: var(--c-white); }

	.svc-hero-row {
		display: flex;
		align-items: flex-start;
		gap: 24px;
	}

	.svc-icon {
		flex-shrink: 0;
		width: 72px;
		height: 72px;
		display: grid;
		place-items: center;
		font-size: 32px;
		color: var(--c-accent);
		background: var(--c-surface);
		border: 1px solid var(--c-border2);
		border-radius: var(--radius);
	}

	.svc-title {
		font-family: var(--f-display);
		font-size: clamp(36px, 5vw, 56px);
		font-weight: 700;
		color: var(--c-white);
		letter-spacing: -0.02em;
		line-height: 1.05;
		margin-bottom: 20px;
	}

	.svc-desc {
		font-size: 18px;
		color: var(--c-muted);
		line-height: 1.7;
		max-width: 720px;
	}

	.svc-content { padding: 80px 0 100px; }

	.svc-layout {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: 64px;
		align-items: stretch;
	}

	.svc-main {
		display: flex;
		flex-direction: column;
		min-height: 100%;
	}

	.svc-block {
		flex: 1;
	}

	.svc-rich {
		font-size: 16px;
		color: var(--c-muted);
		line-height: 1.75;
		max-width: 720px;
	}

	.svc-rich :global(p) {
		margin: 0 0 1rem;
	}

	.svc-rich :global(p:last-child) {
		margin-bottom: 0;
	}

	.svc-rich :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 1.25rem 0;
		border-radius: var(--radius);
	}

	.svc-rich :global(ul),
	.svc-rich :global(ol) {
		margin: 0 0 1rem 1.25rem;
		padding: 0;
	}

	.svc-rich :global(a) {
		color: var(--c-accent);
		text-decoration: underline;
	}

	.svc-rich :global(h2),
	.svc-rich :global(h3) {
		font-family: var(--f-display);
		color: var(--c-white);
		margin: 1.5rem 0 0.75rem;
	}

	.svc-sidebar {
		display: flex;
		flex-direction: column;
		gap: 24px;
		position: sticky;
		top: calc(var(--nav-h) + 24px);
	}

	.svc-card {
		background: var(--c-surface);
		border: 1px solid var(--c-border2);
		border-radius: var(--radius);
		padding: 28px 24px;
	}

	.svc-card .section-label {
		margin-bottom: 16px;
	}

	.svc-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.svc-tag {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-muted);
		border: 1px solid var(--c-border);
		padding: 4px 10px;
		border-radius: 100px;
	}

	.svc-cta {
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

	.svc-cta:hover { opacity: 0.88; transform: translateY(-1px); }

	:global([data-theme="light"]) .svc-cta { color: #fff; }

	@media (max-width: 1024px) {
		.svc-layout { grid-template-columns: 1fr; gap: 48px; }
		.svc-sidebar { position: static; }
	}

	@media (max-width: 768px) {
		.svc-hero { padding: 48px 0; }
		.svc-content { padding: 56px 0 80px; }
		.svc-hero-row { flex-direction: column; }
	}
</style>
