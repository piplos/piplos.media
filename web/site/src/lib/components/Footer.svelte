<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { SITE } from '$lib/site';

	const year = new Date().getFullYear();
</script>

<footer class="footer">
	<div class="container">
		<div class="footer-main">
			<div class="footer-brand">
				<Logo variant="footer" label="{SITE.displayName} home" />
				<p class="footer-tagline">
					© {year} {SITE.displayName} <br>
					{langStore.t('footer.tagline')}
				</p>
			</div>

			<nav class="footer-nav" aria-label="Services">
				<h3>{langStore.t('footer.services')}</h3>
				<ul>
					<li><a href="/#services">{langStore.t('footer.links.web')}</a></li>
					<li><a href="/#services">{langStore.t('footer.links.mobile')}</a></li>
					<li><a href="/#services">{langStore.t('footer.links.backend')}</a></li>
					<li><a href="/#services">{langStore.t('footer.links.devops')}</a></li>
					<li><a href="/order">{langStore.t('footer.links.consulting')}</a></li>
				</ul>
			</nav>

			<nav class="footer-nav" aria-label="Company">
				<h3>{langStore.t('footer.company')}</h3>
				<ul>
					<li><a href="/portfolio">{langStore.t('footer.links.portfolio')}</a></li>
					<li><a href="/#stack">{langStore.t('footer.links.stack')}</a></li>
					<li><a href="/#about">{langStore.t('nav.about')}</a></li>
					<li><a href="/order">{langStore.t('footer.links.start')}</a></li>
				</ul>
			</nav>

			<nav class="footer-nav" aria-label="Legal">
				<h3>{langStore.t('footer.legal')}</h3>
				<ul>
					<li><a href="/privacy">{langStore.t('footer.links.privacy')}</a></li>
					<li><a href="/terms">{langStore.t('footer.links.terms')}</a></li>
				</ul>
			</nav>
		</div>
	</div>
</footer>
