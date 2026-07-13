<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import { getProjectLocale, type PortfolioProject } from '$lib/portfolio';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let project = $derived(data.project as PortfolioProject);
	let loc = $derived(getProjectLocale(project, langStore.value));
</script>

<svelte:head>
	<title>{loc.title} — Case Study | piplos.dev</title>
	<meta name="description" content={loc.description} />
	<link rel="canonical" href="https://piplos.dev/portfolio/{project.id}" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<a href="/portfolio">{langStore.t('nav.portfolio')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{loc.title}</span>
	</div>
</nav>

<main id="main">

	<section class="cs-hero" aria-labelledby="cs-title">
		<div class="container">
			<a href="/portfolio" class="cs-back">
				<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
					<path d="M7 2L3 6l4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
				{langStore.t('case_study.back')}
			</a>

			<div class="cs-meta">
				<span class="cs-subtitle">{loc.subtitle}</span>
				<span class="cs-year">{project.year}</span>
			</div>

			<h1 class="cs-title" id="cs-title">{loc.title}</h1>
			<p class="cs-desc">{loc.description}</p>
		</div>
	</section>

	<section class="cs-content">
		<div class="container">
			<div class="cs-layout">

				<div class="cs-main">

					<article class="cs-block">
						<p class="cs-label">{langStore.t('case_study.challenge')}</p>
						<h2 class="cs-heading">{loc.challenge_title ?? langStore.t('case_study.challenge')}</h2>
						<p class="cs-text">{loc.challenge ?? loc.description}</p>
					</article>

					<article class="cs-block">
						<p class="cs-label">{langStore.t('case_study.solution')}</p>
						<h2 class="cs-heading">{loc.solution_title ?? langStore.t('case_study.solution')}</h2>
						<p class="cs-text">{loc.solution ?? 'We designed and built a scalable, maintainable solution using modern technologies and best engineering practices.'}</p>
					</article>

					{#if project.metrics}
						<article class="cs-block">
							<p class="cs-label">{langStore.t('case_study.result')}</p>
							<div class="cs-metrics">
								{#each project.metrics as metric}
									<div class="cs-metric">
										<div class="cs-metric-value">{metric.value}</div>
										<div class="cs-metric-label">{metric.label}</div>
									</div>
								{/each}
							</div>
						</article>
					{:else if loc.result}
						<article class="cs-block">
							<p class="cs-label">{langStore.t('case_study.result')}</p>
							<h2 class="cs-heading">{langStore.t('case_study.result')}</h2>
							<p class="cs-text">{loc.result}</p>
						</article>
					{/if}
				</div>

				<aside class="cs-sidebar">

					<div class="cs-card">
						<p class="cs-label">{langStore.t('case_study.stack')}</p>
						<div class="cs-stack-tags">
							{#each project.tags as tag}
								<span class="cs-stack-tag">{tag}</span>
							{/each}
						</div>
						{#if loc.stack_detail}
							<p class="cs-stack-detail">{loc.stack_detail}</p>
						{/if}
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

					<a href="/order" class="cs-cta">
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

	.cs-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		margin-bottom: 24px;
	}

	.cs-subtitle {
		font-family: var(--f-mono);
		font-size: 11px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--c-muted);
		background: var(--c-surface2);
		padding: 6px 12px;
		border-radius: var(--radius);
	}

	.cs-year {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-dim);
		letter-spacing: 0.1em;
	}

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

	.cs-metrics {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1px;
		background: var(--c-border);
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
		overflow: hidden;
	}

	.cs-metric {
		background: var(--c-surface);
		padding: 28px 24px;
	}

	.cs-metric-value {
		font-family: var(--f-display);
		font-size: 32px;
		font-weight: 700;
		color: var(--c-accent);
		line-height: 1;
		margin-bottom: 8px;
	}

	.cs-metric-label {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-muted);
		letter-spacing: 0.12em;
		text-transform: uppercase;
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

	.cs-stack-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 16px;
	}

	.cs-stack-tag {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-muted);
		border: 1px solid var(--c-border);
		padding: 4px 10px;
		border-radius: 100px;
	}

	.cs-stack-detail {
		font-size: 13px;
		color: var(--c-dim);
		line-height: 1.6;
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
		color: #000;
		background: var(--c-accent);
		border-radius: var(--radius);
		transition: opacity 0.2s, transform 0.2s;
	}

	.cs-cta:hover { opacity: 0.88; transform: translateY(-1px); }

	:global([data-theme="light"]) .cs-cta { color: #fff; }

	@media (max-width: 1024px) {
		.cs-layout { grid-template-columns: 1fr; gap: 48px; }
		.cs-sidebar { position: static; }
		.cs-metrics { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 768px) {
		.cs-hero { padding: 48px 0; }
		.cs-content { padding: 56px 0 80px; }
		.cs-metrics { grid-template-columns: 1fr; }
	}
</style>
