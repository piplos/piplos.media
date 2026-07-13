<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE, getCompanyYears } from '$lib/site';
	import { SERVICE_ICONS } from '$lib/constants/sections';
	import { toServiceDisplayItems, type ServiceDisplayItem } from '$lib/services-api';
	import StackIcon from '$lib/components/StackIcon.svelte';
	import GridPlaceholder from '$lib/components/GridPlaceholder.svelte';
	import { getCategoryColor, getProjectLocale } from '$lib/portfolio';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const featuredProjects = $derived(data.projects.filter((p) => p.featured).slice(0, 3));
	const stackItems = $derived(data.stackItems);

	type ProcessStep = { title: string; description: string };

	let tickerTrack = $derived.by(() => {
		const items = langStore.get<string[]>('services.ticker') ?? [];
		return [...items, ...items];
	});

	let services = $derived.by(() => {
		const items: (Omit<ServiceDisplayItem, 'icon'> & { icon?: string })[] =
			data.services.length > 0
				? toServiceDisplayItems(data.services, langStore.value)
				: (langStore.get<Omit<ServiceDisplayItem, 'icon'>[]>('services.items') ?? []);
		return items.map((item, i) => ({
			...item,
			num: String(i + 1).padStart(2, '0'),
			icon: item.icon || SERVICE_ICONS[item.id] || '⬡'
		}));
	});

	const SERVICES_COLUMNS = 2;
	const servicesPlaceholderCount = $derived(
		(SERVICES_COLUMNS - (services.length % SERVICES_COLUMNS)) % SERVICES_COLUMNS
	);

	let process = $derived(
		(langStore.get<ProcessStep[]>('process.steps') ?? []).map((step, i) => ({
			num: String(i + 1).padStart(2, '0'),
			...step
		}))
	);

	const companyYears = getCompanyYears();
</script>

<svelte:head>
	<title>{SITE.name} — Web, Mobile & Backend</title>
	<meta name="description" content="{SITE.name} builds web applications, mobile apps, backend systems and DevOps infrastructure for startups and enterprises." />
	<meta property="og:title" content="{SITE.name} — Web, Mobile & Backend" />
	<meta property="og:description" content="Engineering-first team building web apps, mobile apps and backend systems." />
	<meta property="og:url" content={SITE.url} />
	<meta property="og:type" content="website" />
	<link rel="canonical" href="{SITE.url}{l('/')}" />
</svelte:head>

<main id="main">

	<!-- ─── HERO ─────────────────────────────────────── -->
	<section class="hero" aria-labelledby="hero-heading">
		<div class="container">
			<div class="hero-inner">
				<div class="hero-content">
					<h1 class="hero-title" id="hero-heading">
						{langStore.t('hero.headline_1')}<br>
							<span class="line-accent">{langStore.t('hero.headline_2')}</span>
					</h1>
					<p class="hero-sub">
						{langStore.t('hero.description')}
					</p>
					<div class="hero-actions">
						<a href={l('/portfolio')} class="btn-primary" aria-label="View our portfolio of software projects">
							{langStore.t('hero.cta_primary')}
							<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
						</a>
						<a href={l('/order')} class="btn-secondary" aria-label="Start a project with {SITE.name}">
							{langStore.t('hero.cta_secondary')}
						</a>
					</div>
				</div>
				<div class="hero-illustration">
					<img
						src="/hero-cat-isometric.png"
						alt="Piplos Media — cat mascot on a stack of software interfaces"
						width="640"
						height="660"
						loading="eager"
					/>
				</div>
			</div>
		</div>
	</section>

	<!-- ─── STATS ─────────────────────────────────────── -->
	<div class="stats" role="region" aria-label="Company statistics">
		<div class="container">
			<div class="stat">
				<div class="stat-num">240<span class="accent">+</span></div>
				<div class="stat-label">{langStore.t('stats.projects')}</div>
			</div>
			<div class="stat">
				<div class="stat-num">{companyYears}<span class="accent">yr</span></div>
				<div class="stat-label">{langStore.t('stats.experience')}</div>
			</div>
			<div class="stat">
				<div class="stat-num">40<span class="accent">+</span></div>
				<div class="stat-label">{langStore.t('stats.technologies')}</div>
			</div>
			<div class="stat">
				<div class="stat-num">98<span class="accent">%</span></div>
				<div class="stat-label">{langStore.t('stats.retention')}</div>
			</div>
		</div>
	</div>

	<!-- ─── TICKER ────────────────────────────────────── -->
	<div class="ticker" aria-hidden="true">
		<div class="ticker-track">
			{#each tickerTrack as item, i (i)}
				<span class="ticker-item">{item}</span>
			{/each}
		</div>
	</div>

	<!-- ─── SERVICES ──────────────────────────────────── -->
	<section class="section" id="services" aria-labelledby="services-heading">
		<div class="container">
			<div class="section-header">
				<div>
					<p class="section-label">{langStore.t('services.section_label')}</p>
					<h2 class="section-title" id="services-heading">{langStore.t('services.title')}</h2>
				</div>
				<a href={l('/order')} class="section-link" aria-label="Start a project">
					{langStore.t('services.cta')}
					<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</a>
			</div>
			<div class="services-grid" role="list">
				{#each services as svc (svc.id)}
					<article class="service-card" role="listitem" itemscope itemtype="https://schema.org/Service">
						<div class="service-num">{svc.num}</div>
						<div class="service-icon" aria-hidden="true">{svc.icon}</div>
						<h3 class="service-title" itemprop="name">{svc.title}</h3>
						<p class="service-desc" itemprop="description">{svc.description}</p>
						<div class="service-tags" aria-label="Technologies">
							{#each svc.tags as tag (tag)}
								<span class="tag">{tag}</span>
							{/each}
						</div>
					</article>
				{/each}
				{#each Array(servicesPlaceholderCount) as _, i (`placeholder-${i}`)}
					<GridPlaceholder label={langStore.t('services.coming_soon')} variant="service" />
				{/each}
			</div>
		</div>
	</section>

	<!-- ─── WORK ──────────────────────────────────────── -->
	<section class="section" id="work" aria-labelledby="work-heading" style="padding-top:0">
		<div class="container">
			<div class="section-header">
				<div>
					<p class="section-label">{langStore.t('work.section_label')}</p>
					<h2 class="section-title" id="work-heading">{langStore.t('work.title')}</h2>
				</div>
				<a href={l('/portfolio')} class="section-link" aria-label="View full portfolio">
						{langStore.t('work.cta')}
					<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</a>
			</div>
			<div class="work-grid" role="list">
				{#each featuredProjects as project (project.id)}
					{@const loc = getProjectLocale(project, langStore.value)}
					<article class="work-card" role="listitem" itemscope itemtype="https://schema.org/CreativeWork">
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
							<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
						</a>
					</article>
				{/each}
			</div>
		</div>
	</section>

	<!-- ─── STACK ─────────────────────────────────────── -->
	<section class="section stack-section" id="stack" aria-labelledby="stack-heading">
		<div class="container">
			<div class="section-header">
				<div>
					<p class="section-label">{langStore.t('stack.section_label')}</p>
					<h2 class="section-title" id="stack-heading">{langStore.t('stack.title')}</h2>
				</div>
			</div>
			<div class="stack-grid" role="list" aria-label="Technologies we work with">
				{#each stackItems as item (item.slug)}
					<div class="stack-item" role="listitem">
						<StackIcon slug={item.slug} />
						<span class="stack-name">{item.label}</span>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- ─── PROCESS ───────────────────────────────────── -->
	<section class="section" id="about" aria-labelledby="process-heading">
		<div class="container">
			<div class="section-header">
				<div>
					<p class="section-label">{langStore.t('process.section_label')}</p>
					<h2 class="section-title" id="process-heading">{langStore.t('process.title')}</h2>
				</div>
			</div>
			<div class="process-grid" role="list">
				{#each process as step (step.num)}
					<div class="process-step" role="listitem">
						<div class="process-num">{step.num}</div>
						<h3 class="process-title">{step.title}</h3>
						<p class="process-desc">{step.description}</p>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- ─── CTA ───────────────────────────────────────── -->
	<section class="cta-section" aria-labelledby="cta-heading">
		<div class="container">
			<div class="cta-inner">
				<div>
					<h2 class="cta-title" id="cta-heading">
						{langStore.t('cta.title')}
					</h2>
					<p class="cta-sub">{langStore.t('cta.description')}</p>
					<a href={l('/order')} class="btn-primary" style="margin-top:32px;display:inline-flex" aria-label="Start a project with {SITE.name}">
						{langStore.t('cta.button')}
						<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
					</a>
				</div>
				<div class="cta-contacts">
					<div class="cta-contact-item">
						<div class="cta-contact-label">Email</div>
						<a href="mailto:{SITE.email}" class="cta-contact-val" aria-label="Email {SITE.name}">{SITE.email}</a>
					</div>
					<div class="cta-contact-item">
						<div class="cta-contact-label">Phone</div>
						<div class="cta-contact-phones">
							{#each SITE.phones as phone (phone.tel)}
								<a href="tel:{phone.tel}" class="cta-contact-val" aria-label="Call {SITE.name}">{phone.display}</a>
							{/each}
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

</main>
