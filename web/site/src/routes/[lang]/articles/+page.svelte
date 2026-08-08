<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import { articleDate, formatArticleDate, getArticleLocale } from '$lib/articles-api';
	import { uploadCardImage } from '$lib/upload-image';
	import GridPlaceholder from '$lib/components/GridPlaceholder.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const articles = $derived(data.articles);
	/** Индекс первой карточки с обложкой — кандидат в LCP на mobile. */
	const lcpImageIndex = $derived(articles.findIndex((article) => Boolean(article.image)));
	const lcpImage = $derived(
		lcpImageIndex >= 0 ? uploadCardImage(articles[lcpImageIndex]?.image ?? '') : null
	);

	/** Совпадает с фактическим числом колонок CSS-сетки. */
	let articlesGrid = $state<HTMLDivElement | null>(null);
	let articlesColumns = $state(3);

	$effect(() => {
		const el = articlesGrid;
		if (!el) return;

		const update = () => {
			const cols = getComputedStyle(el)
				.gridTemplateColumns.split(' ')
				.filter(Boolean).length;
			articlesColumns = Math.max(1, cols);
		};

		update();
		const observer = new ResizeObserver(update);
		observer.observe(el);
		return () => observer.disconnect();
	});

	const articlesPlaceholderCount = $derived(
		articles.length === 0
			? 0
			: (articlesColumns - (articles.length % articlesColumns)) % articlesColumns
	);

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
	{#if lcpImage}
		<link
			rel="preload"
			as="image"
			type="image/webp"
			href={lcpImage.src}
			imagesrcset={lcpImage.srcset || undefined}
			imagesizes={lcpImage.sizes}
			fetchpriority="high"
		/>
	{/if}
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
				<div class="articles-grid" role="list" bind:this={articlesGrid}>
					{#each articles as article, i (article.id)}
						{@const loc = getArticleLocale(article, langStore.value)}
						{@const isLcpImage = i === lcpImageIndex}
						{@const cardImg = uploadCardImage(article.image)}
						<div role="listitem">
							<a
								href={l(`/articles/${article.slug}`)}
								class="article-card"
								itemscope
								itemtype="https://schema.org/Article"
								itemprop="url"
							>
								{#if cardImg}
									<div class="article-bg" aria-hidden="true">
										<img
											src={cardImg.src}
											srcset={cardImg.srcset || undefined}
											sizes={cardImg.sizes}
											alt=""
											width="640"
											height="400"
											decoding="async"
											loading={isLcpImage ? 'eager' : 'lazy'}
											fetchpriority={isLcpImage ? 'high' : 'auto'}
										/>
									</div>
								{/if}
								<time class="article-date" datetime={articleDate(article)} itemprop="datePublished">
									{formatArticleDate(articleDate(article), langStore.value)}
								</time>
								<h3 class="article-title" itemprop="headline">{loc.title || article.slug}</h3>
								{#if loc.description}
									<p class="article-desc" itemprop="description">{loc.description}</p>
								{/if}
								<span class="article-link">
									{langStore.t('articles.read')}
									<span class="sr-only">: {loc.title || article.slug}</span>
									<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
								</span>
							</a>
						</div>
					{/each}
					{#if articlesPlaceholderCount > 0}
						<div role="listitem" style:grid-column={`span ${articlesPlaceholderCount}`}>
							<GridPlaceholder
								label={langStore.t('services.coming_soon')}
								variant="work"
							/>
						</div>
					{/if}
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
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1px;
		background: var(--c-border);
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
		overflow: hidden;
	}

	.articles-grid > [role='listitem'] {
		display: flex;
		min-width: 0;
		min-height: 0;
	}

	.articles-grid > [role='listitem'] > :global(.grid-placeholder) {
		flex: 1;
		width: 100%;
		box-sizing: border-box;
	}

	.article-card {
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 36px 32px;
		background: var(--c-surface);
		text-decoration: none;
		color: inherit;
		flex: 1;
		width: 100%;
		min-width: 0;
		transition: background 0.2s;
		content-visibility: auto;
		contain-intrinsic-size: auto 320px;
	}

	.article-card:hover {
		background: var(--c-surface2);
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
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center;
		opacity: 0.12;
		filter: grayscale(60%);
		transition: opacity 0.35s ease, transform 0.35s ease, filter 0.35s ease;
	}

	.article-card:hover .article-bg img {
		opacity: 0.32;
		filter: grayscale(0);
		transform: scale(1.04);
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
		font-size: 24px;
		font-weight: 600;
		line-height: 1.2;
		letter-spacing: -0.01em;
		color: var(--c-white);
		overflow-wrap: anywhere;
		transition: color 0.2s;
	}

	.article-card:hover .article-title {
		color: var(--c-accent);
	}

	.article-desc {
		flex: 1;
		font-size: 14px;
		color: var(--c-muted);
		line-height: 1.65;
	}

	.article-link {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-top: 8px;
		font-family: var(--f-mono);
		font-size: 12px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--c-muted);
		transition: color 0.2s;
	}

	.article-card:hover .article-link {
		color: var(--c-accent);
	}

	@media (max-width: 1024px) {
		.articles-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.articles-section {
			padding: 48px 0 80px;
		}

		.articles-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
