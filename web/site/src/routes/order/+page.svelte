<script lang="ts">
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import FieldSelect from '$lib/components/FieldSelect.svelte';
	import { getOrderPrefillFromProject } from '$lib/order-prefill';
	import type { PortfolioProject } from '$lib/portfolio';
	import portfolioData from '$lib/data/portfolio.json';

	type Currency = 'USD' | 'BYN' | 'EUR';

	const BUDGET_MIN = 3000;
	const BUDGET_MAX = 50000;
	const BUDGET_STEP = 1000;

	let step = $state(1);
	let submitted = $state(false);
	let prefillApplied = $state(false);
	let currencyManual = $state(false);
	let currency = $state<Currency>('USD');

	let form = $state({
		types: [] as string[],
		projectName: '',
		description: '',
		stack: '',
		references: '',
		budget: 15000,
		timeline: '',
		stage: '',
		firstName: '',
		lastName: '',
		email: '',
		company: '',
		phone: '',
		howFound: '',
		notes: ''
	});

	const projectTypes = [
		{ id: 'web',        label: 'Web App',        desc: 'SPA, dashboard, portal, CRM' },
		{ id: 'mobile',     label: 'Mobile App',      desc: 'iOS, Android, cross-platform' },
		{ id: 'backend',    label: 'Backend / API',   desc: 'Microservices, REST, GraphQL' },
		{ id: 'devops',     label: 'DevOps',          desc: 'CI/CD, cloud, containers' },
		{ id: 'saas',       label: 'SaaS Product',    desc: 'Multi-tenant, subscriptions' },
		{ id: 'consulting', label: 'Consulting',      desc: 'Audit, architecture, review' }
	];

	const currencyOptions = [
		{ value: 'USD', label: 'USD' },
		{ value: 'EUR', label: 'EUR' },
		{ value: 'BYN', label: 'BYN' }
	];

	const stageOptions = [
		{ id: 'idea',   label: 'Just an Idea',     desc: 'Starting from scratch' },
		{ id: 'design', label: 'Have Design',       desc: 'Wireframes or mockups ready' },
		{ id: 'mvp',    label: 'Existing MVP',      desc: 'Rebuild or scale' },
		{ id: 'legacy', label: 'Legacy Migration',  desc: 'Modernize old system' }
	];

	const timelineOptions = [
		{ value: '', label: '— Select timeline —' },
		{ value: '1m', label: 'Less than 1 month' },
		{ value: '1-3m', label: '1–3 months' },
		{ value: '3-6m', label: '3–6 months' },
		{ value: '6-12m', label: '6–12 months' },
		{ value: '12m+', label: '12+ months / Ongoing' },
		{ value: 'flexible', label: 'Flexible' }
	];

	const howFoundOptions = [
		{ value: '', label: '— Select —' },
		{ value: 'google', label: 'Google Search' },
		{ value: 'referral', label: 'Referral' },
		{ value: 'social', label: 'Social Media' },
		{ value: 'github', label: 'GitHub' },
		{ value: 'other', label: 'Other' }
	];

	const promises = [
		{ bold: '24h response', text: '— we review every brief personally and reply within one business day.' },
		{ bold: 'Free estimate', text: '— detailed scope, timeline and cost breakdown at no charge.' },
		{ bold: 'No lock-in', text: '— you own all code, repos and infrastructure from day one.' },
		{ bold: 'Fixed-price or T&M', text: '— we adapt to your preferred engagement model.' },
		{ bold: 'NDA on request', text: '— confidentiality agreement available before any discussion.' }
	];

	function formatBudgetAmount(v: number) {
		const locale = langStore.value === 'ru' ? 'ru-BY' : 'en-US';
		const amount =
			v >= BUDGET_MAX
				? `${BUDGET_MAX.toLocaleString(locale)}+`
				: v.toLocaleString(locale);

		if (currency === 'USD') return `$${amount}`;
		if (currency === 'EUR') return `€${amount}`;
		return amount;
	}

	function formatRangeLabel(v: number) {
		const suffix = v >= BUDGET_MAX ? '+' : '';
		const short = v >= 1000 ? `${v / 1000}k` : `${v}`;

		if (currency === 'USD') return `$${short}${suffix}`;
		if (currency === 'EUR') return `€${short}${suffix}`;
		return `${short}${suffix}`;
	}

	function onCurrencyChange() {
		currencyManual = true;
	}

	function toggleType(id: string) {
		const idx = form.types.indexOf(id);
		if (idx >= 0) form.types = form.types.filter(t => t !== id);
		else form.types = [...form.types, id];
	}

	function isStep1Valid() {
		return form.types.length > 0;
	}

	function isStep2Valid() {
		return !!form.projectName.trim() && !!form.description.trim();
	}

	function isStep3Valid() {
		return !!form.timeline && !!form.stage;
	}

	function isStep4Valid() {
		return !!form.firstName.trim() && !!form.email.trim();
	}

	function canGoTo(n: number) {
		if (n === 1) return true;
		if (n === 2) return isStep1Valid();
		if (n === 3) return isStep1Valid() && isStep2Valid();
		if (n === 4) return isStep1Valid() && isStep2Valid() && isStep3Valid();
		return false;
	}

	function goTo(n: number) {
		if (canGoTo(n)) step = n;
	}

	function next() {
		if (canNext) step++;
	}

	function back() {
		if (step > 1) step--;
	}

	function submit(e: Event) {
		e.preventDefault();
		submitted = true;
	}

	let canNext = $derived(
		step === 1 ? isStep1Valid() :
		step === 2 ? isStep2Valid() :
		step === 3 ? isStep3Valid() :
		isStep4Valid()
	);

	$effect(() => {
		if (!browser || prefillApplied) return;

		const fromId = page.url.searchParams.get('from');
		if (!fromId) return;

		const project = (portfolioData as PortfolioProject[]).find((item) => item.id === fromId);
		if (!project) return;

		const prefill = getOrderPrefillFromProject(project, langStore.value);
		form.types = prefill.types;
		form.projectName = prefill.projectName;
		form.description = prefill.description;
		form.stack = prefill.stack;
		form.references = prefill.references;
		prefillApplied = true;
	});

	$effect(() => {
		langStore.value;

		if (!currencyManual) {
			currency = langStore.value === 'ru' ? 'BYN' : 'USD';
		}
	});

	$effect(() => {
		if (form.budget > BUDGET_MAX) form.budget = BUDGET_MAX;
		if (form.budget < BUDGET_MIN) form.budget = BUDGET_MIN;
	});

	$effect(() => {
		if (canGoTo(step)) return;

		for (let n = step - 1; n >= 1; n--) {
			if (canGoTo(n)) {
				step = n;
				return;
			}
		}

		step = 1;
	});

	const stepLabels = ['Type', 'Details', 'Budget', 'Contact'];
</script>

<svelte:head>
	<title>Start a Project — {SITE.name}</title>
	<meta name="description" content="Tell us about your project — web app, mobile app, backend or DevOps. Get a free estimate within 24 hours." />
	<link rel="canonical" href="{SITE.url}/order" />
</svelte:head>

<!-- Breadcrumb -->
<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">Start a Project</span>
	</div>
</nav>

<main id="main">
<section class="order-section">
	<div class="container order-layout">

		<!-- ─── FORM SIDE ─────────────────────────────── -->
		<div class="form-side">

			{#if submitted}
				<!-- SUCCESS STATE -->
				<div class="success-state" aria-live="polite" role="status">
					<div class="success-icon" aria-hidden="true">✓</div>
					<h2 class="success-h2">Request Sent!</h2>
					<p class="success-text">We've received your project brief. Our team will review it and get back to you within <strong>24 hours</strong> with a plan and initial estimate.</p>
					<a href="/portfolio" class="btn-submit" style="display:inline-flex;margin-top:0;">
						{langStore.t('hero.cta_primary')}
						<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
					</a>
				</div>
			{:else}
				<p class="form-eyebrow">Response within 24 hours</p>
				<h1 class="form-h1">Start a<br><span class="a">Project</span></h1>

				<!-- STEP TABS -->
				<div class="steps" role="tablist" aria-label="Form steps">
					{#each stepLabels as label, i}
						{@const stepNum = i + 1}
						<button
							type="button"
							class="step-tab"
							class:active={step === stepNum}
							class:done={step > stepNum}
							class:locked={!canGoTo(stepNum)}
							role="tab"
							aria-selected={step === stepNum}
							aria-disabled={!canGoTo(stepNum)}
							id="tab-{stepNum}"
							disabled={!canGoTo(stepNum)}
							onclick={() => goTo(stepNum)}
						>
							<span class="step-n">0{stepNum}</span>
							<span class="step-lbl">{label}</span>
						</button>
					{/each}
				</div>

				<form onsubmit={submit} novalidate aria-label="Project order form">

					<!-- STEP 1 — Project Type -->
					{#if step === 1}
						<div role="tabpanel" aria-labelledby="tab-1">
							<h2 class="step-title">What are you building?</h2>
							<div class="field">
								<span class="field-label" id="type-lbl">Project Type <span style="color:var(--c-accent2);margin-left:4px;">*</span></span>
								<div class="options-grid" role="group" aria-labelledby="type-lbl">
									{#each projectTypes as t}
										<label class="opt-card" class:sel={form.types.includes(t.id)}>
											<input type="checkbox" name="project_type" value={t.id} checked={form.types.includes(t.id)} onchange={() => toggleType(t.id)} class="sr-only" />
											<span class="opt-check" aria-hidden="true"></span>
											<div>
												<div class="opt-label">{t.label}</div>
												<div class="opt-hint">{t.desc}</div>
											</div>
										</label>
									{/each}
								</div>
							</div>
							<div class="form-nav">
								<span class="step-prog" aria-live="polite">Step 1 of 4</span>
								<button type="button" class="btn-next" onclick={next} disabled={!canNext} aria-label="Go to step 2">
									Next
									<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
								</button>
							</div>
						</div>
					{/if}

					<!-- STEP 2 — Details -->
					{#if step === 2}
						<div role="tabpanel" aria-labelledby="tab-2">
							<h2 class="step-title">Tell us more</h2>
							<div class="field">
								<label class="field-label" for="project_name">Project Name <span style="color:var(--c-accent2);">*</span></label>
								<input type="text" id="project_name" name="project_name" class="field-input" placeholder="e.g. Analytics Platform, Delivery App..." required autocomplete="off" bind:value={form.projectName} />
							</div>
							<div class="field">
								<label class="field-label" for="project_desc">Description <span style="color:var(--c-accent2);">*</span></label>
								<p class="field-hint">What problem does it solve? Who are the users? What's the core functionality?</p>
								<textarea id="project_desc" name="project_desc" class="field-textarea" placeholder="Describe your project..." required rows="5" bind:value={form.description}></textarea>
							</div>
							<div class="field">
								<label class="field-label" for="tech_stack">Preferred Tech Stack</label>
								<p class="field-hint">Leave blank — we'll recommend the best stack.</p>
								<input type="text" id="tech_stack" name="tech_stack" class="field-input" placeholder="e.g. React, Node.js, PostgreSQL..." bind:value={form.stack} />
							</div>
							<div class="field">
								<label class="field-label" for="references">References / Inspiration</label>
								<input type="text" id="references" name="references" class="field-input" placeholder="URLs to similar products..." bind:value={form.references} />
							</div>
							<div class="form-nav">
								<button type="button" class="btn-prev" onclick={back}>← Back</button>
								<span class="step-prog" aria-live="polite">Step 2 of 4</span>
								<button type="button" class="btn-next" onclick={next} disabled={!canNext}>
									Next
									<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
								</button>
							</div>
						</div>
					{/if}

					<!-- STEP 3 — Budget -->
					{#if step === 3}
						<div role="tabpanel" aria-labelledby="tab-3">
							<h2 class="step-title">Budget & Timeline</h2>
							<div class="field">
								<span class="field-label" id="budget-lbl">Estimated Budget</span>
								<div class="budget-val" aria-live="polite">
									<span class="budget-amount">{formatBudgetAmount(form.budget)}</span>
									<FieldSelect
										id="budget_currency"
										name="budget_currency"
										variant="inline"
										placeholder=""
										ariaLabel="Currency"
										options={currencyOptions}
										bind:value={currency}
										onchange={onCurrencyChange}
									/>
								</div>
								<input
									type="range"
									id="budget_range"
									name="budget_range"
									min={BUDGET_MIN}
									max={BUDGET_MAX}
									step={BUDGET_STEP}
									bind:value={form.budget}
									aria-labelledby="budget-lbl"
								/>
								<div class="range-labels">
									<span>{formatRangeLabel(BUDGET_MIN)}</span>
									<span>{formatRangeLabel(20000)}</span>
									<span>{formatRangeLabel(35000)}</span>
									<span>{formatRangeLabel(BUDGET_MAX)}</span>
								</div>
							</div>
							<div class="field">
								<label class="field-label" for="timeline">Desired Timeline</label>
								<FieldSelect
									id="timeline"
									name="timeline"
									placeholder="— Select timeline —"
									options={timelineOptions}
									bind:value={form.timeline}
								/>
							</div>
							<div class="field">
								<span class="field-label" id="stage-lbl">Current Stage</span>
								<div class="options-grid" role="group" aria-labelledby="stage-lbl">
									{#each stageOptions as s}
										<label class="opt-card" class:sel={form.stage === s.id}>
											<input type="radio" name="stage" value={s.id} checked={form.stage === s.id} onchange={() => form.stage = s.id} class="sr-only" />
											<span class="opt-check" aria-hidden="true"></span>
											<div>
												<div class="opt-label">{s.label}</div>
												<div class="opt-hint">{s.desc}</div>
											</div>
										</label>
									{/each}
								</div>
							</div>
							<div class="form-nav">
								<button type="button" class="btn-prev" onclick={back}>← Back</button>
								<span class="step-prog" aria-live="polite">Step 3 of 4</span>
								<button type="button" class="btn-next" onclick={next} disabled={!canNext}>
									Next
									<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
								</button>
							</div>
						</div>
					{/if}

					<!-- STEP 4 — Contact -->
					{#if step === 4}
						<div role="tabpanel" aria-labelledby="tab-4">
							<h2 class="step-title">Your Contact Info</h2>
							<div class="field-row">
								<div class="field">
									<label class="field-label" for="first_name">First Name <span style="color:var(--c-accent2);">*</span></label>
									<input type="text" id="first_name" name="first_name" class="field-input" placeholder="Alex" required autocomplete="given-name" bind:value={form.firstName} />
								</div>
								<div class="field">
									<label class="field-label" for="last_name">Last Name</label>
									<input type="text" id="last_name" name="last_name" class="field-input" placeholder="Johnson" autocomplete="family-name" bind:value={form.lastName} />
								</div>
							</div>
							<div class="field">
								<label class="field-label" for="email">Email <span style="color:var(--c-accent2);">*</span></label>
								<input type="email" id="email" name="email" class="field-input" placeholder="alex@company.com" required autocomplete="email" bind:value={form.email} />
							</div>
							<div class="field-row">
								<div class="field">
									<label class="field-label" for="company">Company</label>
									<input type="text" id="company" name="company" class="field-input" placeholder="Acme Corp" autocomplete="organization" bind:value={form.company} />
								</div>
								<div class="field">
									<label class="field-label" for="phone">Phone</label>
									<input type="tel" id="phone" name="phone" class="field-input" placeholder="+1 555 000 0000" autocomplete="tel" bind:value={form.phone} />
								</div>
							</div>
							<div class="field">
								<label class="field-label" for="how_found">How did you find us?</label>
								<FieldSelect
									id="how_found"
									name="how_found"
									placeholder="— Select —"
									options={howFoundOptions}
									bind:value={form.howFound}
								/>
							</div>
							<div class="field">
								<label class="field-label" for="extra_notes">Anything else?</label>
								<textarea id="extra_notes" name="extra_notes" class="field-textarea" placeholder="NDA required, specific constraints, preferred communication channel..." rows="3" bind:value={form.notes}></textarea>
							</div>
							<div class="form-nav">
								<button type="button" class="btn-prev" onclick={back}>← Back</button>
								<span class="step-prog" aria-live="polite">Step 4 of 4</span>
								<button type="submit" class="btn-submit" disabled={!canNext}>
									Send Request
									<svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true"><path d="M1 7h12M8 3l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
								</button>
							</div>
						</div>
					{/if}

				</form>
			{/if}
		</div>

		<!-- ─── INFO SIDE ─────────────────────────────── -->
		<aside class="info-side" aria-label="Why work with {SITE.name}">

			<div class="info-block">
				<p class="info-label">Our Promise</p>
				<h2 class="info-title">Why {SITE.name}</h2>
				<ul class="promise-list" role="list">
					{#each promises as p}
						<li class="promise-item">
							<span class="promise-check" aria-hidden="true">✓</span>
							<p class="promise-text"><strong>{p.bold}</strong> {p.text}</p>
						</li>
					{/each}
				</ul>
			</div>

			<div class="info-block">
				<p class="info-label">Testimonial</p>
				<blockquote class="testimonial">
					<p class="testimonial-text">"{SITE.name} delivered our analytics platform on time and under budget. Their code quality and communication are exceptional — we've been working together for 3 years."</p>
					<cite class="testimonial-author">— CTO, B2B SaaS Startup</cite>
				</blockquote>
			</div>

			<div class="info-block">
				<p class="info-label">Direct Contact</p>
				<div class="contact-list">
					<div class="contact-item">
						<div class="contact-item-lbl">Email</div>
						<a href="mailto:{SITE.email}" class="contact-item-val">{SITE.email}</a>
					</div>
					<div class="contact-item">
						<div class="contact-item-lbl">Phone</div>
						<div class="contact-phones">
							{#each SITE.phones as phone}
								<a href="tel:{phone.tel}" class="contact-item-val">{phone.display}</a>
							{/each}
						</div>
					</div>
				</div>
			</div>

		</aside>
	</div>
</section>
</main>

<style>
	.field { margin-bottom: 24px; }
	.field-hint { font-size: 13px; color: var(--c-dim); margin-bottom: 8px; line-height: 1.5; }
	.options-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
	.form-eyebrow { font-family: var(--f-mono); font-size: 11px; color: var(--c-accent); letter-spacing: 0.2em; text-transform: uppercase; margin-bottom: 16px; }
	.form-h1 { font-family: var(--f-display); font-size: clamp(36px, 5vw, 64px); font-weight: 700; color: var(--c-white); letter-spacing: -0.02em; line-height: 1.05; margin-bottom: 48px; }
	.form-h1 .a { color: var(--c-accent); }
	.steps { display: flex; border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; margin-bottom: 48px; }
	.step-tab { flex: 1; padding: 14px 16px; border: none; border-right: 1px solid var(--c-border); display: flex; align-items: center; gap: 10px; background: var(--c-surface); transition: background 0.2s; cursor: pointer; font: inherit; color: inherit; text-align: left; }
	.step-tab:last-child { border-right: none; }
	.step-tab.active { background: var(--c-surface2); }
	.step-tab.done { background: var(--c-accent-soft); }
	.step-tab:disabled,
	.step-tab.locked {
		opacity: 0.4;
		cursor: not-allowed;
	}
	.step-n { font-family: var(--f-mono); font-size: 18px; font-weight: 700; color: var(--c-dim); line-height: 1; }
	.step-tab.active .step-n { color: var(--c-accent); }
	.step-tab.done .step-n { color: rgba(253, 83, 63, 0.45); }
	.step-lbl { font-family: var(--f-mono); font-size: 11px; color: var(--c-dim); letter-spacing: 0.1em; text-transform: uppercase; }
	.step-tab.active .step-lbl { color: var(--c-muted); }
	.step-title { font-family: var(--f-display); font-size: 24px; font-weight: 600; color: var(--c-white); margin-bottom: 32px; letter-spacing: -0.01em; }
	.form-nav { display: flex; align-items: center; justify-content: space-between; margin-top: 40px; gap: 16px; }
	.step-prog { font-family: var(--f-mono); font-size: 12px; color: var(--c-dim); }
	.btn-prev { font-family: var(--f-mono); font-size: 12px; font-weight: 500; letter-spacing: 0.1em; text-transform: uppercase; color: var(--c-muted); border: 1px solid var(--c-border2); padding: 12px 24px; border-radius: var(--radius); transition: color 0.2s, border-color 0.2s; background: transparent; cursor: pointer; }
	.btn-prev:hover { color: var(--c-white); border-color: var(--c-muted); }
	.btn-next { font-family: var(--f-brand); font-size: 12px; font-weight: 400; letter-spacing: 0.08em; text-transform: uppercase; color: #fff; background: var(--c-accent); padding: 12px 28px; border-radius: var(--radius); transition: opacity 0.2s; display: flex; align-items: center; gap: 8px; cursor: pointer; border: none; }
	.btn-next:hover { opacity: 0.88; }
	.btn-next:disabled { opacity: 0.35; pointer-events: none; }
	.btn-submit { font-family: var(--f-brand); font-size: 13px; font-weight: 400; letter-spacing: 0.08em; text-transform: uppercase; color: #fff; background: var(--c-accent); padding: 14px 36px; border-radius: var(--radius); transition: opacity 0.2s; display: flex; align-items: center; gap: 8px; cursor: pointer; border: none; }
	.btn-submit:hover { opacity: 0.88; }
	.btn-submit:disabled { opacity: 0.35; pointer-events: none; }
	.budget-val {
		display: flex;
		align-items: baseline;
		gap: 10px;
		flex-wrap: wrap;
		font-family: var(--f-display);
		font-size: 48px;
		font-weight: 700;
		color: var(--c-white);
		line-height: 1;
		margin-bottom: 16px;
	}
	.budget-amount { letter-spacing: -0.02em; }
	input[type="range"] { -webkit-appearance: none; width: 100%; height: 3px; background: var(--c-border2); outline: none; cursor: pointer; border-radius: 2px; }
	input[type="range"]::-webkit-slider-thumb { -webkit-appearance: none; width: 20px; height: 20px; background: var(--c-accent); border: 2px solid var(--c-bg); border-radius: 50%; cursor: pointer; transition: transform 0.2s; }
	input[type="range"]::-webkit-slider-thumb:hover { transform: scale(1.2); }
	.range-labels { display: flex; justify-content: space-between; margin-top: 8px; }
	.range-labels span { font-family: var(--f-mono); font-size: 11px; color: var(--c-dim); }
	.success-state { padding: 80px 0; text-align: center; }
	.success-icon { width: 80px; height: 80px; background: var(--c-accent-soft); border: 1px solid var(--c-accent-ring); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 32px; font-size: 32px; }
	.success-h2 { font-family: var(--f-display); font-size: 40px; font-weight: 700; color: var(--c-white); letter-spacing: -0.02em; margin-bottom: 16px; }
	.success-text { font-size: 16px; color: var(--c-muted); line-height: 1.7; max-width: 440px; margin: 0 auto 36px; }
	.info-block { padding-bottom: 36px; margin-bottom: 36px; border-bottom: 1px solid var(--c-border); }
	.info-block:last-child { border-bottom: none; margin-bottom: 0; padding-bottom: 0; }
	.info-label { font-family: var(--f-mono); font-size: 11px; color: var(--c-accent); letter-spacing: 0.2em; text-transform: uppercase; margin-bottom: 16px; }
	.info-title { font-family: var(--f-display); font-size: 20px; font-weight: 600; color: var(--c-white); margin-bottom: 16px; letter-spacing: -0.01em; }
	@media (max-width: 768px) {
		.options-grid { grid-template-columns: 1fr; }
		.step-lbl { display: none; }
	}
</style>
