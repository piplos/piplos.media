<script lang="ts">
	import { langStore } from '$lib/stores/lang.svelte';

	const year = new Date().getFullYear();
</script>

<footer class="footer" role="contentinfo">
	<div class="container">
		<div class="footer-main">
			<div class="footer-brand">
				<a href="/" class="footer-logo" aria-label="piplos.dev home">
					piplos<span class="dot">.</span>dev
				</a>
				<p class="footer-tagline">
					{langStore.t('footer.tagline')}
				</p>
				<a href="https://piplos.media" class="footer-parent" rel="noopener" target="_blank">
					↗ piplos.media
				</a>
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
					<li><a href="https://piplos.media" rel="noopener" target="_blank">piplos.media →</a></li>
				</ul>
			</nav>
		</div>

		<div class="footer-bottom">
			<p class="footer-copy">© {year} {langStore.t('footer.copyright')}</p>
			<nav class="footer-legal" aria-label="Legal">
				<a href="/privacy">{langStore.t('footer.links.privacy')}</a>
				<a href="/terms">{langStore.t('footer.links.terms')}</a>
			</nav>
		</div>
	</div>
</footer>
