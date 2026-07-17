<script lang="ts">
	import { page } from '$app/state';
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE, getCompanyYears } from '$lib/site';
	import { PORTFOLIO_FILTER_KEYS } from '$lib/constants/sections';
	import { getCategoryColor, getProjectLocale } from '$lib/portfolio';
	import { sortProjectsByGroupOrder } from '$lib/portfolio-api';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const projects = $derived(data.projects);

	// ?filter=web — входной фильтр (редиректы старых разделов /portfolio/{type}).
	const requestedFilter = page.url.searchParams.get('filter');
	let activeFilter = $state(
		requestedFilter && (PORTFOLIO_FILTER_KEYS as readonly string[]).includes(requestedFilter)
			? requestedFilter
			: 'all'
	);

	let categories = $derived(
		PORTFOLIO_FILTER_KEYS.map((key) => ({
			key,
			label: langStore.t(`portfolio.filters.${key}`)
		}))
	);

	// «Все» — сквозной порядок из админки (порядок массива);
	// фильтр по группе — порядок внутри группы (sort_order).
	let filtered = $derived(
		activeFilter === 'all'
			? projects
			: sortProjectsByGroupOrder(projects.filter((p) => p.categories?.includes(activeFilter)))
	);

	function getCount(key: string) {
		if (key === 'all') return projects.length;
		return projects.filter((p) => p.categories?.includes(key)).length;
	}

	const companyYears = getCompanyYears();
</script>

<svelte:head>
	<title>Portfolio — {SITE.displayName}</title>
	<meta name="description" content="Browse {SITE.displayName} portfolio of 240+ projects: web apps, mobile apps, SaaS platforms, fintech and enterprise systems." />
	<link rel="canonical" href="{SITE.url}{l('/portfolio')}" />
</svelte:head>

<!-- Breadcrumb -->
<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href={l('/')}>{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t('nav.portfolio')}</span>
	</div>
</nav>

<main id="main">

	<!-- Page Hero -->
	<section class="page-hero" aria-labelledby="portfolio-h1">
		<div class="container">
			<p class="page-eyebrow">{langStore.t('portfolio.eyebrow')}</p>
			<h1 class="page-h1" id="portfolio-h1">{langStore.t('portfolio.title')}</h1>
			<p class="page-desc">{langStore.t('portfolio.description')}</p>
		</div>
	</section>

	<!-- Filter Bar -->
	<div class="filter-bar" role="navigation" aria-label="Filter by category">
		<div class="container">
			<div class="filter-inner">
				<span class="filter-label" aria-hidden="true">{langStore.t('portfolio.filter_label')}</span>
				{#each categories as cat (cat.key)}
					<button
						class="filter-btn"
						class:active={activeFilter === cat.key}
						onclick={() => activeFilter = cat.key}
						aria-pressed={activeFilter === cat.key}
					>
						{cat.label} <span class="filter-count">{getCount(cat.key)}</span>
					</button>
				{/each}
			</div>
		</div>
	</div>

	<!-- Portfolio Grid -->
	<section class="portfolio-section" aria-labelledby="grid-heading">
		<div class="container">
			<h2 id="grid-heading" class="sr-only">Project case studies</h2>
			<div class="portfolio-grid" role="list">
				{#each filtered as project, i (project.id)}
					{@const loc = getProjectLocale(project, langStore.value)}
					<article
						class="portfolio-card"
						class:featured={i === 0 && activeFilter === 'all'}
						role="listitem"
						itemscope
						itemtype="https://schema.org/CreativeWork"
					>
						{#if project.image}
							<div class="pc-bg" aria-hidden="true">
								<img src={project.image} alt="" loading="lazy" />
							</div>
						{/if}
						<div class="pc-top">
							<span class="pc-type">
								<span class="pc-dot" style="background:{getCategoryColor(project.category)}" aria-hidden="true"></span>
								{loc.subtitle}
							</span>
							<span class="pc-year">{project.year}</span>
						</div>
						<h3 class="pc-title" itemprop="name">
							<a href={l(`/portfolio/${project.id}`)} class="pc-title-link" aria-label="View {loc.title} case study">{loc.title}</a>
						</h3>
						<p class="pc-desc" itemprop="description">{loc.description}</p>
						<div class="pc-tags">
							{#each project.tags as tag (tag)}
								<span class="pc-tag">{tag}</span>
							{/each}
						</div>
						<a href={l(`/portfolio/${project.id}`)} class="pc-link" itemprop="url" aria-label="View {loc.title} case study">
							{langStore.t('portfolio.case_study')}
							<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
						</a>
					</article>
				{/each}
			</div>
		</div>
	</section>

	<!-- Stats Bar -->
	<div class="stats-bar" role="region" aria-label="Company statistics">
		<div class="container">
			<div class="stats-inner">
				<div class="stat-item">
					<div class="stat-num">240<span class="a">+</span></div>
					<div class="stat-label">{langStore.t('portfolio.stats.projects')}</div>
				</div>
				<div class="stat-item">
					<div class="stat-num">{companyYears}<span class="a">yr</span></div>
					<div class="stat-label">{langStore.t('portfolio.stats.years')}</div>
				</div>
				<div class="stat-item">
					<div class="stat-num">40<span class="a">+</span></div>
					<div class="stat-label">{langStore.t('portfolio.stats.technologies')}</div>
				</div>
				<div class="stat-item">
					<div class="stat-num">98<span class="a">%</span></div>
					<div class="stat-label">{langStore.t('portfolio.stats.retention')}</div>
				</div>
			</div>
		</div>
	</div>

	<!-- CTA -->
	<section class="cta-section" aria-labelledby="cta-h">
		<div class="container">
			<div class="cta-inner">
				<div>
					<h2 class="cta-title" id="cta-h">{langStore.t('portfolio.cta_title')}</h2>
					<p class="cta-sub">{langStore.t('portfolio.cta_sub')}</p>
				</div>
				<a href={l('/order')} class="btn-primary" aria-label="Start a project with {SITE.name}">
					{langStore.t('nav.start_project')}
					<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</a>
			</div>
		</div>
	</section>

</main>

<style>
	/* Filter Inner */
	.filter-inner { display: flex; align-items: center; overflow-x: auto; scrollbar-width: none; }
	.filter-inner::-webkit-scrollbar { display: none; }
	.filter-label { font-family: var(--f-mono); font-size: 11px; color: var(--c-dim); letter-spacing: 0.15em; text-transform: uppercase; padding: 0 24px 0 0; white-space: nowrap; flex-shrink: 0; }
	.filter-count { font-size: 10px; opacity: 0.5; margin-left: 5px; }

	/* Portfolio section */
	.portfolio-section { padding: 64px 0 100px; }

	/* Stats Bar */
	.stats-bar { border-top: 1px solid var(--c-border); border-bottom: 1px solid var(--c-border); background: var(--c-surface); }
	.stats-inner { display: grid; grid-template-columns: repeat(4, 1fr); }
	.stat-item { padding: 36px 0; text-align: center; border-right: 1px solid var(--c-border); }
	.stat-item:last-child { border-right: none; }
	.stat-num { font-family: var(--f-display); font-size: 40px; font-weight: 700; color: var(--c-white); line-height: 1; margin-bottom: 6px; }
	.stat-num .a { color: var(--c-accent); }
	.stat-label { font-family: var(--f-mono); font-size: 11px; color: var(--c-muted); letter-spacing: 0.15em; text-transform: uppercase; }

	/* CTA */
	.cta-section { padding: 100px 0; }
	.cta-inner { background: var(--c-surface); border: 1px solid var(--c-border2); border-radius: 8px; padding: 72px 64px; display: flex; align-items: center; justify-content: space-between; gap: 48px; }
	.cta-title { font-family: var(--f-display); font-size: clamp(28px, 3.5vw, 48px); font-weight: 700; color: var(--c-white); letter-spacing: -0.02em; line-height: 1.1; margin-bottom: 12px; }
	.cta-sub { font-size: 15px; color: var(--c-muted); line-height: 1.7; }

	@media (max-width: 1024px) {
		.stats-inner { grid-template-columns: repeat(2, 1fr); }
		.stat-item:nth-child(2) { border-right: none; }
		.stat-item:nth-child(3) { border-right: 1px solid var(--c-border); border-top: 1px solid var(--c-border); }
		.stat-item:nth-child(4) { border-top: 1px solid var(--c-border); }
		.cta-inner { flex-direction: column; align-items: flex-start; }
	}
	@media (max-width: 768px) {
		.cta-inner { padding: 40px 28px; }
	}
</style>
