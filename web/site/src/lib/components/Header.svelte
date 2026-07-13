<script lang="ts">
	import { page } from '$app/state';
	import { langStore } from '$lib/stores/lang.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let menuOpen = $state(false);

	const navLinks = [
		{ key: 'nav.services', href: '/#services' },
		{ key: 'nav.portfolio', href: '/portfolio' },
		{ key: 'nav.stack', href: '/#stack' },
		{ key: 'nav.about', href: '/#about' }
	];

	function isActive(href: string) {
		const { pathname, hash } = page.url;
		const [path, fragment] = href.split('#');

		if (!path || path === '/') {
			if (!fragment) return pathname === '/';
			return pathname === '/' && hash === `#${fragment}`;
		}

		return pathname.startsWith(path);
	}
</script>

<header class="header" role="banner">
	<div class="container">
		<a href="/" class="logo" aria-label="piplos.dev — home">
			piplos<span class="logo-dot">.</span>dev
		</a>

		<!-- Desktop nav -->
		<nav class="nav" aria-label="Primary navigation">
			{#each navLinks as link}
				<a
					href={link.href}
					aria-current={isActive(link.href) ? 'page' : undefined}
				>
					{langStore.t(link.key)}
				</a>
			{/each}
		</nav>

		<!-- Right controls -->
		<div class="header-right">
			<!-- Language toggle -->
			<button
				class="lang-toggle"
				onclick={() => langStore.toggle()}
				aria-label={langStore.t('lang.switch')}
			>
				{langStore.t('lang.switch')}
			</button>

			<ThemeToggle />

				<!-- CTA -->
				<a href="/order" class="btn-nav" aria-label="Start a project with piplos.dev">
					{langStore.t('nav.start_project')}
				</a>

			<!-- Mobile hamburger -->
			<button
				class="hamburger"
				onclick={() => menuOpen = !menuOpen}
				aria-label="Toggle menu"
				aria-expanded={menuOpen}
			>
				<span style="transform: {menuOpen ? 'rotate(45deg) translate(4px, 4px)' : 'none'}"></span>
				<span style="opacity: {menuOpen ? 0 : 1}"></span>
				<span style="transform: {menuOpen ? 'rotate(-45deg) translate(4px, -4px)' : 'none'}"></span>
			</button>
		</div>
	</div>

	<!-- Mobile menu -->
	{#if menuOpen}
		<div class="mobile-menu">
			<nav>
				{#each navLinks as link}
					<a
						href={link.href}
						onclick={() => menuOpen = false}
						aria-current={isActive(link.href) ? 'page' : undefined}
					>
						{langStore.t(link.key)}
					</a>
				{/each}
				<a href="/order" onclick={() => menuOpen = false} class="mobile-cta">
					{langStore.t('nav.start_project')}
				</a>
			</nav>
		</div>
	{/if}
</header>

<style>
	.hamburger {
		display: none;
		flex-direction: column;
		gap: 5px;
		padding: 8px;
		cursor: pointer;
		background: none;
		border: none;
		margin-left: 12px;
	}
	.hamburger span {
		display: block;
		width: 20px;
		height: 2px;
		background: var(--c-text);
		transition: all 0.2s;
	}
	.mobile-menu {
		display: none;
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		background: var(--c-surface);
		border-bottom: 1px solid var(--c-border);
	}
	.mobile-menu nav {
		display: flex;
		flex-direction: column;
		padding: 16px;
		gap: 4px;
	}
	.mobile-menu nav a {
		font-family: var(--f-mono);
		font-size: 12px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-muted);
		padding: 12px 16px;
		border-radius: var(--radius);
		transition: color 0.2s, background 0.2s;
	}
	.mobile-menu nav a:hover { color: var(--c-white); background: var(--c-surface2); }
	.mobile-menu nav a[aria-current="page"] {
		background: var(--c-accent);
		color: #fff;
		font-weight: 700;
	}
	.mobile-cta {
		margin-top: 8px;
		background: var(--c-accent) !important;
		color: #000 !important;
		text-align: center;
		font-weight: 700;
	}

	@media (max-width: 768px) {
		.hamburger { display: flex; }
		.mobile-menu { display: block; }
	}
</style>
