<script lang="ts">
	import { page } from '$app/state';
	import { langStore } from '$lib/stores/lang.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { SITE } from '$lib/site';
	import type { Lang } from '$lib/stores/lang.svelte';

	let menuOpen = $state(false);
	let langOpen = $state(false);
	let langRootEl = $state<HTMLDivElement | null>(null);

	const langOptions: { value: Lang; label: string }[] = [
		{ value: 'en', label: 'EN' },
		{ value: 'ru', label: 'RU' }
	];

	function toggleLang(e?: Event) {
		e?.stopPropagation();
		langOpen = !langOpen;
	}

	function selectLang(next: Lang) {
		if (next !== langStore.value) langStore.set(next);
		langOpen = false;
	}

	function onLangKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			langOpen = false;
			return;
		}

		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			toggleLang();
		}
	}

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

	$effect(() => {
		document.body.classList.toggle('menu-open', menuOpen);
		document.body.style.overflow = menuOpen ? 'hidden' : '';

		return () => {
			document.body.classList.remove('menu-open');
			document.body.style.overflow = '';
		};
	});

	$effect(() => {
		if (!langOpen) return;

		function onDocumentClick(e: MouseEvent) {
			if (!langRootEl?.contains(e.target as Node)) langOpen = false;
		}

		function onDocumentKeydown(e: KeyboardEvent) {
			if (e.key === 'Escape') langOpen = false;
		}

		document.addEventListener('click', onDocumentClick);
		document.addEventListener('keydown', onDocumentKeydown);

		return () => {
			document.removeEventListener('click', onDocumentClick);
			document.removeEventListener('keydown', onDocumentKeydown);
		};
	});
</script>

<div class="header-shell">
	<header class="header">
	<div class="container">
		<Logo label="{SITE.displayName} — home" />

		<!-- Desktop nav -->
		<nav class="nav" aria-label="Primary navigation">
			{#each navLinks as link (link.href)}
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
			<div class="lang-select-root" bind:this={langRootEl}>
				<button
					type="button"
					class="lang-toggle"
					onclick={toggleLang}
					onkeydown={onLangKeydown}
					aria-haspopup="listbox"
					aria-expanded={langOpen}
					aria-controls="lang-listbox"
					aria-label={langStore.t('lang.switch')}
				>
					{langStore.value === 'en' ? 'EN' : 'RU'}
				</button>

				{#if langOpen}
					<ul
						id="lang-listbox"
						class="field-select-menu field-select-menu--inline"
						role="listbox"
						aria-label={langStore.t('lang.switch')}
					>
						{#each langOptions as option (option.value)}
							<li role="presentation">
								<button
									type="button"
									role="option"
									class="field-select-option"
									class:selected={langStore.value === option.value}
									aria-selected={langStore.value === option.value}
									onclick={() => selectLang(option.value)}
								>
									{option.label}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<ThemeToggle />

				<!-- CTA -->
				<a href="/order" class="btn-nav" aria-label="Start a project with {SITE.name}">
					{langStore.t('nav.start_project')}
				</a>

			<!-- Mobile hamburger -->
			<button
				class="hamburger"
				class:is-open={menuOpen}
				onclick={() => menuOpen = !menuOpen}
				aria-label="Toggle menu"
				aria-expanded={menuOpen}
			>
				<span class="hamburger-line"></span>
				<span class="hamburger-line"></span>
				<span class="hamburger-line"></span>
			</button>
		</div>
	</div>
	</header>

	{#if menuOpen}
		<button
			type="button"
			class="mobile-backdrop"
			aria-label="Close menu"
			onclick={() => menuOpen = false}
		></button>
		<div class="mobile-menu">
			<nav>
				{#each navLinks as link (link.href)}
					<a
						href={link.href}
						onclick={() => menuOpen = false}
						aria-current={isActive(link.href) ? 'page' : undefined}
					>
						{langStore.t(link.key)}
					</a>
				{/each}
			</nav>
		</div>
	{/if}
</div>

<style>
	.lang-select-root {
		position: relative;
		display: inline-block;
	}
	.hamburger {
		display: none;
		position: relative;
		width: 36px;
		height: 36px;
		padding: 0;
		cursor: pointer;
		background: none;
		border: none;
		margin-left: 12px;
		flex-shrink: 0;
	}
	.hamburger-line {
		position: absolute;
		left: 50%;
		top: 50%;
		width: 20px;
		height: 2px;
		margin-left: -10px;
		margin-top: -1px;
		background: var(--c-text);
		border-radius: 1px;
		transition: transform 0.25s ease, opacity 0.2s ease;
	}
	.hamburger-line:nth-child(1) { transform: translateY(-6px); }
	.hamburger-line:nth-child(2) { transform: translateY(0); }
	.hamburger-line:nth-child(3) { transform: translateY(6px); }
	.hamburger.is-open .hamburger-line:nth-child(1) { transform: translateY(0) rotate(45deg); }
	.hamburger.is-open .hamburger-line:nth-child(2) { opacity: 0; }
	.hamburger.is-open .hamburger-line:nth-child(3) { transform: translateY(0) rotate(-45deg); }
	.mobile-backdrop {
		display: none;
		position: fixed;
		inset: 0;
		top: var(--nav-h);
		border: none;
		padding: 0;
		margin: 0;
		background: rgba(8, 8, 12, 0.35);
		backdrop-filter: blur(16px) saturate(120%);
		-webkit-backdrop-filter: blur(16px) saturate(120%);
		z-index: 1;
		cursor: pointer;
	}
	.mobile-menu {
		display: none;
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		z-index: 2;
		background: var(--c-accent);
		border-bottom: none;
		box-shadow: 0 16px 40px var(--c-shadow);
	}
	.mobile-menu nav {
		display: flex;
		flex-direction: column;
		padding: 16px;
		gap: 4px;
	}
	.mobile-menu nav a {
		font-family: var(--f-brand);
		font-size: 14px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: rgba(255, 255, 255, 0.88);
		padding: 12px 16px;
		border-radius: var(--radius);
		transition: color 0.2s, background 0.2s;
	}
	.mobile-menu nav a:hover { color: #fff; background: var(--c-accent-hover); }
	.mobile-menu nav a[aria-current="page"] {
		background: var(--c-accent);
		color: #fff;
		font-weight: 700;
	}
	:global([data-theme="light"]) .mobile-backdrop {
		background: rgba(247, 248, 250, 0.45);
	}

	@media (max-width: 1024px) {
		.hamburger { display: flex; }
		.mobile-backdrop { display: block; }
		.mobile-menu { display: block; }
	}
</style>
