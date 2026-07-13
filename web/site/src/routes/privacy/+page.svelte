<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import { SITE } from '$lib/site';

	const sections = [
		{
			title: '1. Information We Collect',
			body: 'We collect information you provide directly to us when you fill out the project request form on our website. This includes your name, email address, company name, phone number, and project details. We also collect standard web analytics data such as pages visited, time on site, and referral sources through anonymised analytics tools.'
		},
		{
			title: '2. How We Use Your Information',
			body: 'We use the information we collect to respond to your project enquiries, prepare project estimates and proposals, communicate with you about our services, and improve our website and services. We do not sell, rent, or share your personal information with third parties for their marketing purposes.'
		},
		{
			title: '3. Data Storage and Security',
			body: 'Your information is stored on secure servers. We implement appropriate technical and organisational measures to protect your personal data against unauthorised access, alteration, disclosure, or destruction. Project briefs and contact information are retained for up to 3 years unless you request deletion.'
		},
		{
			title: '4. Cookies',
			body: 'Our website uses essential cookies required for basic functionality (theme preference, language selection). We do not use advertising or tracking cookies. You can disable cookies in your browser settings, though this may affect certain features of the website.'
		},
		{
			title: '5. Third-Party Services',
			body: 'We may use third-party services such as analytics providers and hosting platforms. These services have their own privacy policies and we encourage you to review them. We ensure all third-party processors comply with applicable data protection regulations.'
		},
		{
			title: '6. Your Rights',
			body: `You have the right to access, correct, or delete your personal data. You may also object to processing or request data portability. To exercise these rights, contact us at ${SITE.email}. We will respond to your request within 30 days.`
		},
		{
			title: '7. Contact',
			body: `If you have questions about this Privacy Policy or our data practices, please contact us at ${SITE.email} or by post at: ${SITE.name}, ${SITE.location}.`
		}
	];
</script>

<svelte:head>
	<title>Privacy Policy — {SITE.name}</title>
	<meta name="description" content="Privacy Policy for {SITE.name} — how we collect, use and protect your data." />
	<link rel="canonical" href="{SITE.url}/privacy" />
</svelte:head>

<nav class="breadcrumb-bar" aria-label="Breadcrumb">
	<div class="container">
		<a href="/">{langStore.t('nav.home')}</a>
		<span class="sep" aria-hidden="true">/</span>
		<span class="current" aria-current="page">{langStore.t('footer.links.privacy')}</span>
	</div>
</nav>

<main id="main" class="legal-page">
	<div class="container">
		<p class="legal-label">Legal</p>
		<h1 class="legal-title">{langStore.t('privacy.title')}</h1>
		<p class="legal-updated">{langStore.t('privacy.last_updated')}</p>

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
