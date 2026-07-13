<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import { SERVICE_ICONS } from '$lib/constants/sections';
	import { getCategoryColor, getProjectLocale, type PortfolioProject } from '$lib/portfolio';
	import portfolioData from '$lib/data/portfolio.json';

	const projects = portfolioData as PortfolioProject[];
	const featuredProjects = projects.filter((p) => p.featured).slice(0, 3);

	type ServiceItem = { id: string; title: string; description: string; tags: string[] };
	type ProcessStep = { title: string; description: string };

	const stack = [
		{ icon: '⚛', label: 'React' },
		{ icon: '▲', label: 'Next.js' },
		{ icon: 'TS', label: 'TypeScript' },
		{ icon: '⬡', label: 'Node.js' },
		{ icon: '🐘', label: 'PostgreSQL' },
		{ icon: '⚡', label: 'Redis' },
		{ icon: '🐳', label: 'Docker' },
		{ icon: '☸', label: 'Kubernetes' },
		{ icon: '☁', label: 'AWS' },
		{ icon: '📱', label: 'React Native' },
		{ icon: '◈', label: 'GraphQL' },
		{ icon: '🦀', label: 'Rust' },
		{ icon: '🐍', label: 'Python' },
		{ icon: '⚙', label: 'Terraform' },
		{ icon: '🔄', label: 'GitHub Actions' },
		{ icon: '🎨', label: 'Figma' }
	];

	const tickerItems = ['Web Applications', 'Mobile Apps', 'Backend Systems', 'DevOps & Cloud', 'API Development', 'Tech Consulting', 'UI/UX Engineering', 'SaaS Platforms'];

	let services = $derived(
		(langStore.get<ServiceItem[]>('services.items') ?? []).map((item, i) => ({
			num: String(i + 1).padStart(2, '0'),
			icon: SERVICE_ICONS[item.id] ?? '⬡',
			...item
		}))
	);

	let process = $derived(
		(langStore.get<ProcessStep[]>('process.steps') ?? []).map((step, i) => ({
			num: String(i + 1).padStart(2, '0'),
			...step
		}))
	);
</script>

<svelte:head>
	<title>piplos.dev — Software Development Studio | Web, Mobile & Backend</title>
	<meta name="description" content="piplos.dev is an engineering-first software development studio. We build web applications, mobile apps, backend systems and DevOps infrastructure for startups and enterprises." />
	<meta property="og:title" content="piplos.dev — Software Development Studio" />
	<meta property="og:description" content="Engineering-first studio building web apps, mobile apps and backend systems." />
	<meta property="og:url" content="https://piplos.dev" />
	<meta property="og:type" content="website" />
	<link rel="canonical" href="https://piplos.dev" />
</svelte:head>

<main id="main">

	<!-- ─── HERO ─────────────────────────────────────── -->
	<section class="hero" aria-labelledby="hero-heading">
		<div class="container">
			<div class="hero-inner">
				<div class="hero-content">
					<div class="hero-eyebrow" aria-hidden="true">
						<span class="hero-eyebrow-dot"></span>
						<span>{langStore.t('hero.eyebrow')}</span>
					</div>
					<h1 class="hero-title" id="hero-heading">
						{langStore.t('hero.headline_1')}<br>
							<span class="line-accent">{langStore.t('hero.headline_2')}</span>
					</h1>
					<p class="hero-sub">
						{langStore.t('hero.description')}
					</p>
					<div class="hero-actions">
						<a href="/portfolio" class="btn-primary" aria-label="View our portfolio of software projects">
							{langStore.t('hero.cta_primary')}
							<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
						</a>
						<a href="/order" class="btn-secondary" aria-label="Start a project with piplos.dev">
							{langStore.t('hero.cta_secondary')}
						</a>
					</div>
				</div>
				<div class="hero-illustration">
					<img
						src="/hero-illustration.png"
						alt="Software development illustration"
						width="520"
						height="420"
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
				<div class="stat-num">120<span class="accent">+</span></div>
				<div class="stat-label">{langStore.t('stats.projects')}</div>
			</div>
			<div class="stat">
				<div class="stat-num">8<span class="accent">yr</span></div>
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
			{#each [...tickerItems, ...tickerItems] as item}
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
				<a href="/order" class="section-link" aria-label="Start a project">
					{langStore.t('services.cta')}
					<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</a>
			</div>
			<div class="services-grid" role="list">
				{#each services as svc}
					<article class="service-card" role="listitem" itemscope itemtype="https://schema.org/Service">
						<div class="service-num">{svc.num}</div>
						<div class="service-icon" aria-hidden="true">{svc.icon}</div>
						<h3 class="service-title" itemprop="name">{svc.title}</h3>
						<p class="service-desc" itemprop="description">{svc.description}</p>
						<div class="service-tags" aria-label="Technologies">
							{#each svc.tags as tag}
								<span class="tag">{tag}</span>
							{/each}
						</div>
					</article>
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
				<a href="/portfolio" class="section-link" aria-label="View full portfolio">
						{langStore.t('work.cta')}
					<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</a>
			</div>
			<div class="work-grid" role="list">
				{#each featuredProjects as project}
					{@const loc = getProjectLocale(project, langStore.value)}
					<article class="work-card" role="listitem" itemscope itemtype="https://schema.org/CreativeWork">
						<div class="work-type">
							<span class="work-type-dot" style="background:{getCategoryColor(project.category)}" aria-hidden="true"></span>
							{loc.subtitle}
						</div>
						<h3 class="work-title" itemprop="name">{loc.title}</h3>
						<p class="work-desc" itemprop="description">{loc.description}</p>
						<a href="/portfolio/{project.id}" class="work-link" itemprop="url" aria-label="View {loc.title} case study">
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
				{#each stack as item}
					<div class="stack-item" role="listitem">
						<span class="stack-icon" aria-hidden="true">{item.icon}</span>
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
				{#each process as step}
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
					<a href="/order" class="btn-primary" style="margin-top:32px;display:inline-flex" aria-label="Start a project with piplos.dev">
						{langStore.t('cta.button')}
						<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
					</a>
				</div>
				<div class="cta-contacts">
					<div class="cta-contact-item">
						<div class="cta-contact-label">Email</div>
						<a href="mailto:dev@piplos.media" class="cta-contact-val" aria-label="Email piplos.dev">dev@piplos.media</a>
					</div>
					<div class="cta-contact-item">
						<div class="cta-contact-label">Phone</div>
						<a href="tel:+375172499897" class="cta-contact-val" aria-label="Call piplos.dev">+375 17 249-98-97</a>
					</div>
					<div class="cta-contact-item">
						<div class="cta-contact-label">Telegram</div>
						<a href="https://t.me/piplosdev" class="cta-contact-val" rel="noopener" target="_blank" aria-label="Message piplos.dev on Telegram">@piplosdev</a>
					</div>
				</div>
			</div>
		</div>
	</section>

</main>
