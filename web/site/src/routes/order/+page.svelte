<script lang="ts">
	import { page } from '$app/state';
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';
	import FieldSelect from '$lib/components/FieldSelect.svelte';
	import { getOrderPrefillFromProject, type OrderPrefill } from '$lib/order-prefill';
	import type { PortfolioProject } from '$lib/portfolio';
	import portfolioData from '$lib/data/portfolio.json';

	type Currency = 'USD' | 'BYN' | 'EUR';
	type Option = { value: string; label: string };
	type OptionCard = { id: string; label: string; hint: string };
	type PromiseItem = { title: string; text: string };

	const BUDGET_MIN = 3000;
	const BUDGET_MAX = 50000;
	const BUDGET_STEP = 1000;
	const TOTAL_STEPS = 4;

	const CURRENCY_SYMBOLS: Record<Currency, string> = { USD: '$', EUR: '€', BYN: '' };

	const currencyOptions: Option[] = ['USD', 'EUR', 'BYN'].map((c) => ({ value: c, label: c }));

	let step = $state(1);
	let submitted = $state(false);
	let appliedPrefill: OrderPrefill | null = null;

	// Currency follows the language until the user picks one manually
	let manualCurrency = $state<Currency | null>(null);
	let currency = $derived(manualCurrency ?? (langStore.value === 'ru' ? 'BYN' : 'USD'));

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

	let stepLabels = $derived(langStore.get<string[]>('order.steps') ?? []);
	let projectTypes = $derived(langStore.get<OptionCard[]>('order.types') ?? []);
	let stageOptions = $derived(langStore.get<OptionCard[]>('order.stages') ?? []);
	let timelineOptions = $derived(langStore.get<Option[]>('order.timelines') ?? []);
	let howFoundOptions = $derived(langStore.get<Option[]>('order.how_found_options') ?? []);
	let promises = $derived(langStore.get<PromiseItem[]>('order.promises') ?? []);

	function formatBudgetAmount(v: number) {
		const locale = langStore.value === 'ru' ? 'ru-BY' : 'en-US';
		const amount = v.toLocaleString(locale) + (v >= BUDGET_MAX ? '+' : '');
		return CURRENCY_SYMBOLS[currency] + amount;
	}

	function formatRangeLabel(v: number) {
		const suffix = v >= BUDGET_MAX ? '+' : '';
		const short = v >= 1000 ? `${v / 1000}k` : `${v}`;
		return `${CURRENCY_SYMBOLS[currency]}${short}${suffix}`;
	}

	function toggleType(id: string) {
		if (form.types.includes(id)) form.types = form.types.filter((t) => t !== id);
		else form.types = [...form.types, id];
	}

	const stepValidators = [
		() => form.types.length > 0,
		() => !!form.projectName.trim() && !!form.description.trim(),
		() => !!form.timeline && !!form.stage,
		() => !!form.firstName.trim() && !!form.email.trim()
	];

	function canGoTo(n: number) {
		return stepValidators.slice(0, n - 1).every((isValid) => isValid());
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

	let canNext = $derived(stepValidators[step - 1]());

	// Prefill from a portfolio project (/order?from=<id>).
	// Re-applies on language change (e.g. lang init after first run) until the user edits the fields.
	$effect(() => {
		const fromId = page.url.searchParams.get('from');
		if (!fromId) return;

		const project = (portfolioData as PortfolioProject[]).find((item) => item.id === fromId);
		if (!project) return;

		const untouched =
			!appliedPrefill ||
			(form.projectName === appliedPrefill.projectName &&
				form.description === appliedPrefill.description &&
				form.stack === appliedPrefill.stack &&
				form.references === appliedPrefill.references);
		if (!untouched) return;

		const prefill = getOrderPrefillFromProject(project, langStore.value);
		form.types = prefill.types;
		form.projectName = prefill.projectName;
		form.description = prefill.description;
		form.stack = prefill.stack;
		form.references = prefill.references;
		appliedPrefill = prefill;
	});
</script>

{#snippet arrow(size: number)}
	<svg width={size} height={size} viewBox="0 0 12 12" fill="none" aria-hidden="true"><path d="M1 6h10M7 2l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
{/snippet}

{#snippet formNav(stepNum: number)}
	<div class="form-nav">
		{#if stepNum > 1}
			<button type="button" class="btn-prev" onclick={back}>← {langStore.t('order.back')}</button>
		{/if}
		<span class="step-prog" aria-live="polite">{langStore.t('order.step_of', { current: stepNum, total: TOTAL_STEPS })}</span>
		{#if stepNum < TOTAL_STEPS}
			<button type="button" class="btn-next" onclick={next} disabled={!canNext}>
				{langStore.t('order.next')}
				{@render arrow(12)}
			</button>
		{:else}
			<button type="submit" class="btn-submit" disabled={!canNext}>
				{langStore.t('order.submit')}
				{@render arrow(14)}
			</button>
		{/if}
	</div>
{/snippet}

{#snippet optionCards(name: string, options: OptionCard[], isSelected: (id: string) => boolean, onSelect: (id: string) => void, multiple: boolean)}
	{#each options as option (option.id)}
		<label class="opt-card" class:sel={isSelected(option.id)}>
			<input
				type={multiple ? 'checkbox' : 'radio'}
				{name}
				value={option.id}
				checked={isSelected(option.id)}
				onchange={() => onSelect(option.id)}
				class="sr-only"
			/>
			<span class="opt-check" aria-hidden="true"></span>
			<div>
				<div class="opt-label">{option.label}</div>
				<div class="opt-hint">{option.hint}</div>
			</div>
		</label>
	{/each}
{/snippet}

<svelte:head>
	<title>{langStore.t('order.title_1')} {langStore.t('order.title_2')} — {SITE.name}</title>
	<meta name="description" content="Tell us about your project — web app, mobile app, backend or DevOps. Get a free estimate within 24 hours." />
	<link rel="canonical" href="{SITE.url}/order" />
</svelte:head>

<!-- Breadcrumb -->
<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t('nav.start_project')}</span>
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
					<h2 class="success-h2">{langStore.t('order.success_title')}</h2>
					<p class="success-text">{langStore.t('order.success_text')}</p>
					<a href="/portfolio" class="btn-submit" style="display:inline-flex;margin-top:0;">
						{langStore.t('hero.cta_primary')}
						{@render arrow(14)}
					</a>
				</div>
			{:else}
				<p class="form-eyebrow">{langStore.t('order.eyebrow')}</p>
				<h1 class="form-h1">{langStore.t('order.title_1')} <span class="a">{langStore.t('order.title_2')}</span></h1>

				<!-- STEP TABS -->
				<div class="steps" role="tablist" aria-label="Form steps">
					{#each stepLabels as label, i (i)}
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
							<h2 class="step-title">{langStore.t('order.step1_title')}</h2>
							<div class="field">
								<span class="field-label" id="type-lbl">{langStore.t('order.type_label')} <span class="field-required">*</span></span>
								<div class="options-grid" role="group" aria-labelledby="type-lbl">
									{@render optionCards('project_type', projectTypes, (id) => form.types.includes(id), toggleType, true)}
								</div>
							</div>
							{@render formNav(1)}
						</div>
					{/if}

					<!-- STEP 2 — Details -->
					{#if step === 2}
						<div role="tabpanel" aria-labelledby="tab-2">
							<h2 class="step-title">{langStore.t('order.step2_title')}</h2>
							<div class="field">
								<label class="field-label" for="project_name">{langStore.t('order.fields.project_name.label')} <span class="field-required">*</span></label>
								<input type="text" id="project_name" name="project_name" class="field-input" placeholder={langStore.t('order.fields.project_name.placeholder')} required autocomplete="off" bind:value={form.projectName} />
							</div>
							<div class="field">
								<label class="field-label" for="project_desc">{langStore.t('order.fields.description.label')} <span class="field-required">*</span></label>
								<p class="field-hint">{langStore.t('order.fields.description.hint')}</p>
								<textarea id="project_desc" name="project_desc" class="field-textarea" placeholder={langStore.t('order.fields.description.placeholder')} required rows="5" bind:value={form.description}></textarea>
							</div>
							<div class="field">
								<label class="field-label" for="tech_stack">{langStore.t('order.fields.stack.label')}</label>
								<p class="field-hint">{langStore.t('order.fields.stack.hint')}</p>
								<input type="text" id="tech_stack" name="tech_stack" class="field-input" placeholder={langStore.t('order.fields.stack.placeholder')} bind:value={form.stack} />
							</div>
							<div class="field">
								<label class="field-label" for="references">{langStore.t('order.fields.references.label')}</label>
								<input type="text" id="references" name="references" class="field-input" placeholder={langStore.t('order.fields.references.placeholder')} bind:value={form.references} />
							</div>
							{@render formNav(2)}
						</div>
					{/if}

					<!-- STEP 3 — Budget -->
					{#if step === 3}
						<div role="tabpanel" aria-labelledby="tab-3">
							<h2 class="step-title">{langStore.t('order.step3_title')}</h2>
							<div class="field">
								<span class="field-label" id="budget-lbl">{langStore.t('order.budget_label')}</span>
								<div class="budget-val" aria-live="polite">
									<span class="budget-amount">{formatBudgetAmount(form.budget)}</span>
									<FieldSelect
										id="budget_currency"
										name="budget_currency"
										variant="inline"
										placeholder=""
										ariaLabel={langStore.t('order.currency_label')}
										options={currencyOptions}
										bind:value={() => currency, (v) => manualCurrency = v as Currency}
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
								<label class="field-label" for="timeline">{langStore.t('order.timeline_label')}</label>
								<FieldSelect
									id="timeline"
									name="timeline"
									placeholder={langStore.t('order.timeline_placeholder')}
									options={timelineOptions}
									bind:value={form.timeline}
								/>
							</div>
							<div class="field">
								<span class="field-label" id="stage-lbl">{langStore.t('order.stage_label')}</span>
								<div class="options-grid" role="group" aria-labelledby="stage-lbl">
									{@render optionCards('stage', stageOptions, (id) => form.stage === id, (id) => form.stage = id, false)}
								</div>
							</div>
							{@render formNav(3)}
						</div>
					{/if}

					<!-- STEP 4 — Contact -->
					{#if step === 4}
						<div role="tabpanel" aria-labelledby="tab-4">
							<h2 class="step-title">{langStore.t('order.step4_title')}</h2>
							<div class="field-row">
								<div class="field">
									<label class="field-label" for="first_name">{langStore.t('order.fields.first_name.label')} <span class="field-required">*</span></label>
									<input type="text" id="first_name" name="first_name" class="field-input" placeholder={langStore.t('order.fields.first_name.placeholder')} required autocomplete="given-name" bind:value={form.firstName} />
								</div>
								<div class="field">
									<label class="field-label" for="last_name">{langStore.t('order.fields.last_name.label')}</label>
									<input type="text" id="last_name" name="last_name" class="field-input" placeholder={langStore.t('order.fields.last_name.placeholder')} autocomplete="family-name" bind:value={form.lastName} />
								</div>
							</div>
							<div class="field">
								<label class="field-label" for="email">{langStore.t('order.fields.email.label')} <span class="field-required">*</span></label>
								<input type="email" id="email" name="email" class="field-input" placeholder={langStore.t('order.fields.email.placeholder')} required autocomplete="email" bind:value={form.email} />
							</div>
							<div class="field-row">
								<div class="field">
									<label class="field-label" for="company">{langStore.t('order.fields.company.label')}</label>
									<input type="text" id="company" name="company" class="field-input" placeholder={langStore.t('order.fields.company.placeholder')} autocomplete="organization" bind:value={form.company} />
								</div>
								<div class="field">
									<label class="field-label" for="phone">{langStore.t('order.fields.phone.label')}</label>
									<input type="tel" id="phone" name="phone" class="field-input" placeholder={langStore.t('order.fields.phone.placeholder')} autocomplete="tel" bind:value={form.phone} />
								</div>
							</div>
							<div class="field">
								<label class="field-label" for="how_found">{langStore.t('order.fields.how_found.label')}</label>
								<FieldSelect
									id="how_found"
									name="how_found"
									placeholder={langStore.t('order.fields.how_found.placeholder')}
									options={howFoundOptions}
									bind:value={form.howFound}
								/>
							</div>
							<div class="field">
								<label class="field-label" for="extra_notes">{langStore.t('order.fields.notes.label')}</label>
								<textarea id="extra_notes" name="extra_notes" class="field-textarea" placeholder={langStore.t('order.fields.notes.placeholder')} rows="3" bind:value={form.notes}></textarea>
							</div>
							{@render formNav(4)}
						</div>
					{/if}

				</form>
			{/if}
		</div>

		<!-- ─── INFO SIDE ─────────────────────────────── -->
		<aside class="info-side" aria-label="Why work with {SITE.name}">

			<div class="info-block">
				<p class="info-label">{langStore.t('order.promise_label')}</p>
				<h2 class="info-title">{langStore.t('order.why_title')} {SITE.name}</h2>
				<ul class="promise-list" role="list">
					{#each promises as p (p.title)}
						<li class="promise-item">
							<span class="promise-check" aria-hidden="true">✓</span>
							<p class="promise-text"><strong>{p.title}</strong> {p.text}</p>
						</li>
					{/each}
				</ul>
			</div>

			<div class="info-block">
				<p class="info-label">{langStore.t('order.testimonial_label')}</p>
				<blockquote class="testimonial">
					<p class="testimonial-text">{langStore.t('order.testimonial_text')}</p>
					<cite class="testimonial-author">{langStore.t('order.testimonial_author')}</cite>
				</blockquote>
			</div>

			<div class="info-block">
				<p class="info-label">{langStore.t('order.contact_label')}</p>
				<div class="contact-list">
					<div class="contact-item">
						<div class="contact-item-lbl">{langStore.t('contact.email')}</div>
						<a href="mailto:{SITE.email}" class="contact-item-val">{SITE.email}</a>
					</div>
					<div class="contact-item">
						<div class="contact-item-lbl">{langStore.t('contact.phone')}</div>
						<div class="contact-phones">
							{#each SITE.phones as phone (phone.tel)}
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
	.field-required { color: var(--c-accent2); margin-left: 4px; }
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
	input[type="range"] { appearance: none; -webkit-appearance: none; width: 100%; height: 3px; background: var(--c-border2); outline: none; cursor: pointer; border-radius: 2px; }
	input[type="range"]::-webkit-slider-thumb { appearance: none; -webkit-appearance: none; width: 20px; height: 20px; background: var(--c-accent); border: 2px solid var(--c-bg); border-radius: 50%; cursor: pointer; transition: transform 0.2s; }
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
