<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { selectProjects, selectServices, type EmbedParams } from '$lib/embeds';
	import { getCategoryColor, getProjectLocale, type PortfolioProject } from '$lib/portfolio';
	import { getServiceLocale, type ServiceItem } from '$lib/services-api';
	import { SERVICE_ICONS } from '$lib/constants/sections';

	interface Props {
		params: EmbedParams;
		projects: PortfolioProject[];
		services: ServiceItem[];
	}
	let { params, projects, services }: Props = $props();

	const projectItems = $derived(
		params.kind === 'projects' ? selectProjects(projects, params) : []
	);
	const serviceItems = $derived(
		params.kind === 'services'
			? selectServices(services, params).map((item) => {
					const loc = getServiceLocale(item, langStore.value);
					return {
						slug: item.slug,
						title: loc.title,
						description: loc.description,
						tags: item.tags ?? [],
						icon: item.icon || SERVICE_ICONS[item.slug] || '⬡'
					};
				})
			: []
	);
	const empty = $derived(
		params.kind === 'projects' ? projectItems.length === 0 : serviceItems.length === 0
	);
</script>

{#if !empty}
	<div class="embed embed--{params.layout}" data-embed={params.kind}>
		{#if params.kind === 'projects'}
			{#if params.layout === 'cards'}
				<div class="embed-grid">
					{#each projectItems as project (project.id)}
						{@const loc = getProjectLocale(project, langStore.value)}
						<a href={l(`/portfolio/${project.id}`)} class="embed-card">
							{#if project.image}
								<span class="embed-card-bg" aria-hidden="true">
									<img src={project.image} alt="" loading="lazy" />
								</span>
							{/if}
							<span class="embed-type">
								<span
									class="embed-type-dot"
									style="background:{getCategoryColor(project.category)}"
									aria-hidden="true"
								></span>
								{loc.subtitle}
							</span>
							<span class="embed-title">{loc.title}</span>
							<span class="embed-desc">{loc.description}</span>
							<span class="embed-link">
								{langStore.t('work.case_study')}
								<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
									<path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
								</svg>
							</span>
						</a>
					{/each}
				</div>
			{:else if params.layout === 'list'}
				<ul class="embed-list">
					{#each projectItems as project (project.id)}
						{@const loc = getProjectLocale(project, langStore.value)}
						<li class="embed-row">
							<a href={l(`/portfolio/${project.id}`)} class="embed-row-title">
								<span
									class="embed-type-dot"
									style="background:{getCategoryColor(project.category)}"
									aria-hidden="true"
								></span>
								{loc.title}
							</a>
							<span class="embed-row-desc">{loc.description}</span>
						</li>
					{/each}
				</ul>
			{:else}
				<div class="embed-chips">
					{#each projectItems as project (project.id)}
						{@const loc = getProjectLocale(project, langStore.value)}
						<a href={l(`/portfolio/${project.id}`)} class="embed-chip">
							<span
								class="embed-type-dot"
								style="background:{getCategoryColor(project.category)}"
								aria-hidden="true"
							></span>
							{loc.title}
						</a>
					{/each}
				</div>
			{/if}
		{:else if params.layout === 'cards'}
			<div class="embed-grid">
				{#each serviceItems as svc (svc.slug)}
					<a href={l(`/services/${svc.slug}`)} class="embed-card">
						<span class="embed-svc-icon" aria-hidden="true">{svc.icon}</span>
						<span class="embed-title">{svc.title}</span>
						<span class="embed-desc">{svc.description}</span>
						{#if svc.tags.length}
							<span class="embed-tags">
								{#each svc.tags.slice(0, 4) as tag (tag)}
									<span class="embed-tag">{tag}</span>
								{/each}
							</span>
						{/if}
					</a>
				{/each}
			</div>
		{:else if params.layout === 'list'}
			<ul class="embed-list">
				{#each serviceItems as svc (svc.slug)}
					<li class="embed-row">
						<a href={l(`/services/${svc.slug}`)} class="embed-row-title">
							<span class="embed-svc-icon embed-svc-icon--sm" aria-hidden="true">{svc.icon}</span>
							{svc.title}
						</a>
						<span class="embed-row-desc">{svc.description}</span>
					</li>
				{/each}
			</ul>
		{:else}
			<div class="embed-chips">
				{#each serviceItems as svc (svc.slug)}
					<a href={l(`/services/${svc.slug}`)} class="embed-chip">
						<span aria-hidden="true">{svc.icon}</span>
						{svc.title}
					</a>
				{/each}
			</div>
		{/if}
	</div>
{/if}

<style>
	.embed {
		margin: 32px 0;
	}

	/* ── cards ── */
	.embed-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1px;
		background: var(--c-border);
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
		overflow: hidden;
	}
	.embed-card {
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 24px 22px;
		background: var(--c-surface);
		text-decoration: none;
		transition: background 0.2s;
	}
	.embed-card:hover {
		background: var(--c-surface2);
	}
	.embed-card > :global(:not(.embed-card-bg)) {
		position: relative;
		z-index: 1;
	}
	.embed-card-bg {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}
	.embed-card-bg img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0.12;
		filter: grayscale(1);
		transition: opacity 0.2s, filter 0.2s, transform 0.2s;
	}
	.embed-card:hover .embed-card-bg img {
		opacity: 0.24;
		filter: grayscale(0);
		transform: scale(1.04);
	}
	.embed-type {
		display: flex;
		align-items: center;
		gap: 8px;
		font-family: var(--f-mono);
		font-size: 11px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--c-muted);
	}
	.embed-type-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.embed-title {
		font-family: var(--f-display);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-white);
		letter-spacing: -0.01em;
		line-height: 1.25;
	}
	.embed-card:hover .embed-title {
		color: var(--c-accent);
	}
	.embed-desc {
		font-size: 13.5px;
		color: var(--c-muted);
		line-height: 1.6;
		flex: 1;
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.embed-link {
		display: flex;
		align-items: center;
		gap: 6px;
		font-family: var(--f-mono);
		font-size: 11px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--c-muted);
		transition: color 0.2s;
	}
	.embed-card:hover .embed-link {
		color: var(--c-accent);
	}
	.embed-svc-icon {
		width: 44px;
		height: 44px;
		display: grid;
		place-items: center;
		font-size: 22px;
		color: var(--c-accent);
		background: var(--c-surface2);
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
	}
	.embed-svc-icon--sm {
		width: 28px;
		height: 28px;
		font-size: 15px;
		flex-shrink: 0;
	}
	.embed-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.embed-tag {
		font-family: var(--f-mono);
		font-size: 10.5px;
		color: var(--c-muted);
		border: 1px solid var(--c-border);
		padding: 3px 9px;
		border-radius: 100px;
	}

	/* ── list ── */
	.embed-list {
		margin: 0;
		padding: 0;
		list-style: none;
		border: 1px solid var(--c-border);
		border-radius: var(--radius);
		overflow: hidden;
	}
	.embed-row {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 16px 20px;
		background: var(--c-surface);
	}
	.embed-row + .embed-row {
		border-top: 1px solid var(--c-border);
	}
	.embed-row-title {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		font-family: var(--f-display);
		font-size: 16px;
		font-weight: 600;
		color: var(--c-white);
		text-decoration: none;
		transition: color 0.2s;
	}
	.embed-row-title:hover {
		color: var(--c-accent);
	}
	.embed-row-desc {
		font-size: 13.5px;
		color: var(--c-muted);
		line-height: 1.6;
	}

	/* ── compact ── */
	.embed-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.embed-chip {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		font-family: var(--f-mono);
		font-size: 12px;
		color: var(--c-white);
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 100px;
		text-decoration: none;
		transition: border-color 0.2s, color 0.2s;
	}
	.embed-chip:hover {
		border-color: var(--c-accent);
		color: var(--c-accent);
	}
</style>
