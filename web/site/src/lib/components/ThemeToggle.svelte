<script lang="ts">
	import { themeStore } from '$lib/stores/theme.svelte';
	import { langStore } from '$lib/stores/lang.svelte';

	let isLight = $derived(themeStore.value === 'light');
	let label = $derived(
		isLight ? langStore.t('theme.toggle_dark') : langStore.t('theme.toggle_light')
	);
</script>

<button
	type="button"
	class="theme-switch"
	class:is-light={isLight}
	onclick={() => themeStore.toggle()}
	aria-label={label}
	title={label}
	aria-pressed={isLight}
>
	<span class="theme-switch-track">
		<span class="theme-switch-thumb" aria-hidden="true"></span>

		<span class="theme-switch-icon theme-switch-icon--sun" class:is-active={isLight} aria-hidden="true">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<circle cx="12" cy="12" r="5"/>
				<line x1="12" y1="1" x2="12" y2="3"/>
				<line x1="12" y1="21" x2="12" y2="23"/>
				<line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
				<line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
				<line x1="1" y1="12" x2="3" y2="12"/>
				<line x1="21" y1="12" x2="23" y2="12"/>
				<line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
				<line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
			</svg>
		</span>

		<span class="theme-switch-icon theme-switch-icon--moon" class:is-active={!isLight} aria-hidden="true">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
			</svg>
		</span>
	</span>
</button>
