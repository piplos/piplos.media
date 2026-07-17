<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { resolveUploadUrl } from '$lib/api';
	import { articleDate, formatArticleDate, getArticleLocale } from '$lib/articles-api';
	import { getCategoryColor, getProjectLocale } from '$lib/portfolio';
	import SafeHtml from '$lib/components/SafeHtml.svelte';
	import { SITE } from '$lib/site';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const article = $derived(data.article);
	const related = $derived(data.related);
	const loc = $derived(getArticleLocale(article, langStore.value));
	const title = $derived(loc.title || article.slug);
	const tags = $derived(article.tags ?? []);

	const seoLoc = $derived(data.seo?.[langStore.value]);
	const seoTitle = $derived(seoLoc?.title || `${title} — ${langStore.t('articles.title')} | ${SITE.displayName}`);
	const seoDescription = $derived(seoLoc?.description || loc.description || '');
	const ogTitle = $derived(seoLoc?.og_title || seoTitle);
	const ogDescription = $derived(seoLoc?.og_description || seoDescription);
	const ogImage = $derived(resolveUploadUrl(seoLoc?.og_image || article.image || ''));
	const canonicalUrl = $derived(`${SITE.url}${l(`/articles/${article.slug}`)}`);
</script>

<svelte:head>
	<title>{seoTitle}</title>
	<meta name="description" content={seoDescription} />
	<link rel="canonical" href={canonicalUrl} />
	<link rel="alternate" hreflang="en" href="{SITE.url}/en/articles/{article.slug}" />
	<link rel="alternate" hreflang="ru" href="{SITE.url}/ru/articles/{article.slug}" />
	<link rel="alternate" hreflang="x-default" href="{SITE.url}/en/articles/{article.slug}" />
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content={SITE.displayName} />
	<meta property="og:locale" content={langStore.value === 'ru' ? 'ru_RU' : 'en_US'} />
	<meta property="og:title" content={ogTitle} />
	<meta property="og:description" content={ogDescription} />
	<meta property="og:url" content={canonicalUrl} />
	<meta property="article:published_time" content={articleDate(article)} />
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
		<a href={l('/articles')}>{langStore.t('nav.articles')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{title}</span>
	</div>
</nav>

<main id="main" itemscope itemtype="https://schema.org/Article">
	<section class="article-hero" aria-labelledby="article-title">
		<div class="container">
			<a href={l('/articles')} class="article-back">
				<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
					<path d="M7 2L3 6l4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
				{langStore.t('articles.back')}
			</a>

			<time class="article-date" datetime={articleDate(article)} itemprop="datePublished">
				{formatArticleDate(articleDate(article), langStore.value)}
			</time>
			<h1 class="article-title" id="article-title" itemprop="headline">{title}</h1>
			{#if loc.description}
				<p class="article-desc" itemprop="description">{loc.description}</p>
			{/if}
		</div>
	</section>

	<section class="article-content">
		<div class="container">
			<div class="article-layout">
				<div class="article-main">
					{#if loc.body?.trim()}
						<div class="article-block" itemprop="articleBody">
							<SafeHtml html={loc.body} class="rich-text" />
						</div>
					{/if}
				</div>

				<aside class="article-sidebar">
					{#if tags.length > 0}
						<div class="article-card">
							<p class="section-label">{langStore.t('articles.stack')}</p>
							<div class="article-tags">
								{#each tags as tag (tag)}
									<span class="article-tag">{tag}</span>
								{/each}
							</div>
						</div>
					{/if}

					<a href={l('/order')} class="article-cta">
						{langStore.t('articles.start_project')}
						<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
							<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</a>
				</aside>
			</div>
		</div>
	</section>

	{#if related.length > 0}
		<section class="section" style="padding-top: 0" aria-labelledby="article-projects-heading">
			<div class="container">
				<div class="section-header">
					<div>
						<p class="section-label">{langStore.t('service_page.projects')}</p>
						<h2 class="section-title" id="article-projects-heading">{langStore.t('work.title')}</h2>
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
						{@const projectLoc = getProjectLocale(project, langStore.value)}
						<article class="work-card" role="listitem" itemscope itemtype="https://schema.org/CreativeWork">
							{#if project.image}
								<div class="pc-bg" aria-hidden="true">
									<img src={project.image} alt="" loading="lazy" />
								</div>
							{/if}
							<div class="work-type">
								<span class="work-type-dot" style="background:{getCategoryColor(project.category)}" aria-hidden="true"></span>
								{projectLoc.subtitle}
							</div>
							<h3 class="work-title" itemprop="name">
								<a href={l(`/portfolio/${project.id}`)} class="work-title-link" aria-label="View {projectLoc.title} case study">{projectLoc.title}</a>
							</h3>
							<p class="work-desc" itemprop="description">{projectLoc.description}</p>
							<a href={l(`/portfolio/${project.id}`)} class="work-link" itemprop="url" aria-label="View {projectLoc.title} case study">
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
	.article-hero {
		padding: 64px 0;
		border-bottom: 1px solid var(--c-border);
	}

	.article-back {
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

	.article-back:hover {
		color: var(--c-white);
	}

	.article-date {
		display: block;
		font-family: var(--f-mono);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-muted);
		margin-bottom: 16px;
	}

	.article-title {
		font-family: var(--f-display);
		font-size: clamp(36px, 5vw, 56px);
		font-weight: 700;
		color: var(--c-white);
		letter-spacing: -0.02em;
		line-height: 1.05;
		margin-bottom: 20px;
	}

	.article-desc {
		font-size: 18px;
		color: var(--c-muted);
		line-height: 1.7;
		max-width: 720px;
	}

	.article-content {
		padding: 80px 0 100px;
	}

	.article-layout {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: 64px;
		align-items: stretch;
	}

	.article-main {
		display: flex;
		flex-direction: column;
		min-height: 100%;
	}

	.article-block {
		flex: 1;
		max-width: 720px;
	}

	.article-block :global(img) {
		max-width: min(360px, 100%);
		margin: 2rem auto;
	}

	.article-sidebar {
		display: flex;
		flex-direction: column;
		gap: 24px;
		position: sticky;
		top: calc(var(--nav-h) + 24px);
	}

	.article-card {
		background: var(--c-surface);
		border: 1px solid var(--c-border2);
		border-radius: var(--radius);
		padding: 28px 24px;
	}

	.article-card .section-label {
		margin-bottom: 16px;
	}

	.article-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.article-tag {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-muted);
		border: 1px solid var(--c-border);
		padding: 4px 10px;
		border-radius: 100px;
	}

	.article-cta {
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

	.article-cta:hover {
		opacity: 0.88;
		transform: translateY(-1px);
	}

	:global([data-theme='light']) .article-cta {
		color: #fff;
	}

	@media (max-width: 1024px) {
		.article-layout {
			grid-template-columns: 1fr;
			gap: 48px;
		}

		.article-sidebar {
			position: static;
		}
	}

	@media (max-width: 768px) {
		.article-hero {
			padding: 48px 0;
		}

		.article-content {
			padding: 56px 0 80px;
		}
	}
</style>
