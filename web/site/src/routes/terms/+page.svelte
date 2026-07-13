<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';

	const sections = [
		{
			title: '1. Acceptance of Terms',
			body: `By accessing and using the ${SITE.name} website, you accept and agree to be bound by these Terms of Service. If you do not agree to these terms, please do not use our website or services.`
		},
		{
			title: '2. Services',
			body: `${SITE.name} provides software development services including web application development, mobile application development, backend and API development, DevOps infrastructure, and technology consulting. The specific terms of any engagement are governed by a separate project agreement or statement of work.`
		},
		{
			title: '3. Intellectual Property',
			body: `Upon full payment, clients receive full ownership of all custom code, designs, and deliverables created specifically for their project. ${SITE.name} retains the right to use general methodologies, frameworks, and non-client-specific code. We may reference completed projects in our portfolio unless otherwise agreed in writing.`
		},
		{
			title: '4. Confidentiality',
			body: 'We treat all client information as confidential. We are happy to sign a Non-Disclosure Agreement (NDA) before any project discussion. Our team members are bound by confidentiality obligations.'
		},
		{
			title: '5. Limitation of Liability',
			body: `${SITE.name} shall not be liable for any indirect, incidental, special, or consequential damages arising from the use of our services. Our total liability for any claim shall not exceed the total fees paid by the client for the specific project giving rise to the claim.`
		},
		{
			title: '6. Payment Terms',
			body: 'Payment terms are specified in individual project agreements. Typically, projects require a deposit before work begins, with remaining payments tied to project milestones. Late payments may result in work suspension.'
		},
		{
			title: '7. Governing Law',
			body: 'These Terms of Service are governed by the laws of the Republic of Belarus. Any disputes shall be resolved through good-faith negotiation. If negotiation fails, disputes shall be submitted to the competent courts of Minsk, Belarus.'
		},
		{
			title: '8. Changes to Terms',
			body: 'We reserve the right to modify these Terms of Service at any time. Changes will be posted on this page with an updated date. Continued use of our website after changes constitutes acceptance of the new terms.'
		},
		{
			title: '9. Contact',
			body: `For questions about these Terms of Service, contact us at ${SITE.email}.`
		}
	];
</script>

<svelte:head>
	<title>Terms of Service — {SITE.name}</title>
	<meta name="description" content="Terms of Service for {SITE.name} — the terms governing use of our website and services." />
	<link rel="canonical" href="{SITE.url}/terms" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t('footer.links.terms')}</span>
	</div>
</nav>

<main id="main" class="legal-page">
	<div class="container">
		<p class="legal-label">Legal</p>
		<h1 class="legal-title">{langStore.t('terms.title')}</h1>
		<p class="legal-updated">{langStore.t('terms.last_updated')}</p>

		<div class="legal-sections">
			{#each sections as section (section.title)}
				<article class="legal-section">
					<h2>{section.title}</h2>
					<p>{section.body}</p>
				</article>
			{/each}
		</div>
	</div>
</main>
