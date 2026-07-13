<script lang="ts">
	import { onMount } from 'svelte';
	import { isLang, localizePath } from '$lib/i18n/routing';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';

	interface Props {
		status?: number;
	}

	let { status = 404 }: Props = $props();

	const kind = $derived(status === 404 ? '404' : '500');

	// Статичный 404.html отдаётся для любого неизвестного URL (включая /ru/*),
	// поэтому язык определяем из первого сегмента реального пути после гидратации.
	onMount(() => {
		const seg = location.pathname.split('/').filter(Boolean)[0];
		if (isLang(seg) && seg !== langStore.value) langStore.set(seg);
	});

	function href(path: string): string {
		return localizePath(path, langStore.value);
	}
</script>

<svelte:head>
	<title>{status} — {langStore.t(`error.${kind}.title`)} | {SITE.name}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<main id="main" class="error-page">
	<div class="container">
		<div class="error-inner">
			<div class="error-content">
				<p class="error-label">{langStore.t(`error.${kind}.label`)}</p>
				<h1 class="error-title">{langStore.t(`error.${kind}.title`)}</h1>
				<p class="error-desc">{langStore.t(`error.${kind}.description`)}</p>
				<div class="error-actions">
					<a href={href('/')} class="btn-primary">
						{langStore.t('error.home')}
						<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
					</a>
					<a href={href('/portfolio')} class="btn-secondary">
						{langStore.t('error.portfolio')}
					</a>
				</div>
			</div>
			<div class="error-illustration">
				<img
					src="/illustrations/cat-{kind}.png"
					alt={langStore.t(`error.${kind}.alt`)}
					width="560"
					height="560"
					loading="eager"
				/>
			</div>
		</div>
	</div>
</main>

<style>
	.error-page {
		position: relative;
		overflow: hidden;
		padding: 80px 0 100px;
	}
	.error-page::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(var(--c-border) 1px, transparent 1px),
			linear-gradient(90deg, var(--c-border) 1px, transparent 1px);
		background-size: 64px 64px;
		opacity: 0.35;
		pointer-events: none;
	}
	.error-page .container {
		position: relative;
		z-index: 1;
	}
	.error-inner {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 64px;
		align-items: center;
		min-height: calc(100vh - var(--nav-h) - 180px);
	}
	.error-label {
		font-family: var(--f-mono);
		font-size: 11px;
		color: var(--c-accent);
		letter-spacing: 0.2em;
		text-transform: uppercase;
		margin-bottom: 16px;
	}
	.error-title {
		font-family: var(--f-display);
		font-size: clamp(40px, 5.5vw, 80px);
		font-weight: 700;
		line-height: 1.05;
		letter-spacing: -0.02em;
		color: var(--c-white);
		margin-bottom: 24px;
	}
	.error-desc {
		font-size: 18px;
		color: var(--c-muted);
		line-height: 1.7;
		max-width: 480px;
		margin-bottom: 40px;
	}
	.error-actions {
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
	}
	.error-illustration {
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.error-illustration img {
		width: 100%;
		max-width: 560px;
	}

	@media (max-width: 1024px) {
		.error-inner {
			grid-template-columns: 1fr;
			gap: 32px;
			min-height: 0;
		}
		.error-illustration {
			order: -1;
		}
		.error-illustration img {
			max-width: 360px;
		}
	}

	@media (max-width: 768px) {
		.error-page {
			padding: 48px 0 64px;
		}
	}
</style>
