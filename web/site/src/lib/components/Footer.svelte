<script lang="ts">
	import { l } from '$lib/i18n/link';
	import { langStore } from '$lib/stores/lang.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { SITE } from '$lib/site';

	type FooterService = { slug: string; title: string };

	let { services = [] }: { services?: FooterService[] } = $props();

	const year = new Date().getFullYear();
</script>

<footer class="footer">
	<div class="container">
		<div class="footer-main">
			<div class="footer-brand">
				<Logo variant="footer" href={l('/')} label="{SITE.displayName} home" />
				<p class="footer-tagline">
					© {year} {SITE.displayName} <br>
					{langStore.t('footer.tagline')}
				</p>
			</div>

			{#if services.length > 0}
				<nav class="footer-nav" aria-label="Services">
					<h3>{langStore.t('footer.services')}</h3>
					<ul>
						{#each services as service (service.slug)}
							<li><a href={l(`/services/${service.slug}`)}>{service.title}</a></li>
						{/each}
					</ul>
				</nav>
			{/if}

			<nav class="footer-nav" aria-label="Company">
				<h3>{langStore.t('footer.company')}</h3>
				<ul>
					<li><a href={l('/portfolio')}>{langStore.t('footer.links.portfolio')}</a></li>
					<li><a href={l('/articles')}>{langStore.t('footer.links.articles')}</a></li>
					<li><a href={l('/#stack')}>{langStore.t('footer.links.stack')}</a></li>
					<li><a href={l('/#about')}>{langStore.t('nav.about')}</a></li>
					<li><a href={l('/order')}>{langStore.t('footer.links.start')}</a></li>
				</ul>
			</nav>

			<nav class="footer-nav" aria-label="Legal">
				<h3>{langStore.t('footer.legal')}</h3>
				<ul>
					<li><a href={l('/legal/privacy')}>{langStore.t('footer.links.privacy')}</a></li>
					<li><a href={l('/legal/cookies')}>{langStore.t('footer.links.cookies')}</a></li>
					<li><a href={l('/legal/terms')}>{langStore.t('footer.links.terms')}</a></li>
				</ul>
			</nav>
		</div>
	</div>
</footer>
