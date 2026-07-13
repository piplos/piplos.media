import { browser } from '$app/environment';

type Theme = 'dark' | 'light';

function createThemeStore() {
	let theme = $state<Theme>('dark');

	function init() {
		if (!browser) return;
		const saved = localStorage.getItem('piplos-theme') as Theme | null;
		const preferred = saved ?? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
		theme = preferred;
		apply(preferred);
	}

	function apply(t: Theme) {
		document.documentElement.setAttribute('data-theme', t);
		const meta = document.getElementById('theme-color-meta');
		if (meta) meta.setAttribute('content', t === 'light' ? '#ffffff' : '#24252a');
		if (browser) localStorage.setItem('piplos-theme', t);
	}

	function toggle() {
		theme = theme === 'dark' ? 'light' : 'dark';
		apply(theme);
	}

	return {
		get value() { return theme; },
		init,
		toggle
	};
}

export const themeStore = createThemeStore();
