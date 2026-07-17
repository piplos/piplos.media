<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import { articleDate, formatArticleDate, getArticleLocale } from '$lib/articles-api';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const articles = $derived(data.articles);

	const pageTitle = $derived(`${langStore.t('articles.title')} — ${SITE.displayName}`);
	const pageDescription = $derived(langStore.t('articles.description'));
	const canonicalUrl = $derived(`${SITE.url}${l('/articles')}`);
</script>

<svelte:head>
	<title>{pageTitle}</title>
	<meta name="description" content={pageDescription} />
	<link rel="canonical" href={canonicalUrl} />
	<link rel="alternate" hreflang="en" href="{SITE.url}/en/articles" />
	<link rel="alternate" hreflang="ru" href="{SITE.url}/ru/articles" />
	<link rel="alternate" hreflang="x-default" href="{SITE.url}/en/articles" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content={SITE.displayName} />
	<meta property="og:locale" content={langStore.value === 'ru' ? 'ru_RU' : 'en_US'} />
	<meta property="og:title" content={pageTitle} />
	<meta property="og:description" content={pageDescription} />
	<meta property="og:url" content={canonicalUrl} />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href={l('/')}>{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t('nav.articles')}</span>
	</div>
</nav>

<main id="main">
	<section class="page-hero" aria-labelledby="articles-h1">
		<div class="container">
			<p class="page-eyebrow">{langStore.t('articles.eyebrow')}</p>
			<h1 class="page-h1" id="articles-h1">{langStore.t('articles.title')}</h1>
			<p class="page-desc">{langStore.t('articles.description')}</p>
		</div>
	</section>

	<section class="articles-section" aria-labelledby="articles-list-heading">
		<div class="container">
			<h2 id="articles-list-heading" class="sr-only">{langStore.t('articles.title')}</h2>
			{#if !articles.length}
				<p class="articles-empty">{langStore.t('articles.empty')}</p>
			{:else}
				<div class="articles-grid" role="list">
					{#each articles as article (article.id)}
						{@const loc = getArticleLocale(article, langStore.value)}
						<article class="article-card" role="listitem" itemscope itemtype="https://schema.org/Article">
							{#if article.image}
								<div class="article-bg" aria-hidden="true">
									<img src={article.image} alt="" loading="lazy" />
								</div>
							{/if}
							<time class="article-date" datetime={articleDate(article)} itemprop="datePublished">
								{formatArticleDate(articleDate(article), langStore.value)}
							</time>
							<h3 class="article-title" itemprop="headline">
								<a href={l(`/articles/${article.slug}`)} class="article-title-link" itemprop="url">
									{loc.title || article.slug}
								</a>
							</h3>
							{#if loc.description}
								<p class="article-desc" itemprop="description">{loc.description}</p>
							{/if}
							<a href={l(`/articles/${article.slug}`)} class="article-link" aria-label="{langStore.t('articles.read')}: {loc.title || article.slug}">
								{langStore.t('articles.read')}
								<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
							</a>
						</article>
					{/each}
				</div>
			{/if}
		</div>
	</section>
</main>

<style>
	.articles-section {
		padding: 64px 0 100px;
	}

	.articles-empty {
		padding: 48px 0;
		color: var(--c-muted);
		font-family: var(--f-mono);
		font-size: 13px;
	}

	.articles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 24px;
	}

	.article-card {
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 28px 24px;
		background: var(--c-surface);
		border: 1px solid var(--c-border2);
		border-radius: var(--radius);
		transition: border-color 0.2s, transform 0.2s, background 0.2s;
	}

	.article-card:hover {
		border-color: var(--c-accent);
		transform: translateY(-2px);
	}

	.article-card > :not(.article-bg) {
		position: relative;
		z-index: 1;
	}

	.article-bg {
		position: absolute;
		inset: 0;
		overflow: hidden;
		pointer-events: none;
	}

	.article-bg img {
		position: absolute;
		right: -4%;
		bottom: -8%;
		height: 82%;
		width: auto;
		opacity: 0.09;
		filter: grayscale(60%);
		transition: opacity 0.35s ease, transform 0.35s ease, filter 0.35s ease;
	}

	.article-card:hover .article-bg img {
		opacity: 0.28;
		filter: grayscale(0);
		transform: scale(1.05);
	}

	.article-date {
		font-family: var(--f-mono);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-muted);
	}

	.article-title {
		font-family: var(--f-display);
		font-size: 22px;
		font-weight: 700;
		line-height: 1.25;
		letter-spacing: -0.01em;
	}

	.article-title-link {
		color: var(--c-white);
		transition: color 0.2s;
	}

	.article-title-link:hover {
		color: var(--c-accent);
	}

	.article-desc {
		flex: 1;
		font-size: 15px;
		color: var(--c-muted);
		line-height: 1.65;
	}

	.article-link {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-family: var(--f-mono);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-accent);
		transition: gap 0.2s;
	}

	.article-link:hover {
		gap: 12px;
	}

	@media (max-width: 768px) {
		.articles-section {
			padding: 48px 0 80px;
		}
	}
</style>
