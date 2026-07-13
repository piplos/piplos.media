import { browser } from '$app/environment';

type Theme = 'dark' | 'light';

const STORAGE_KEY = 'piplos-theme';

function isTheme(value: unknown): value is Theme {
	return value === 'dark' || value === 'light';
}

function createThemeStore() {
	let theme = $state<Theme>('dark');

	function set(t: Theme) {
		theme = t;
		document.documentElement.setAttribute('data-theme', t);
		document
			.getElementById('theme-color-meta')
			?.setAttribute('content', t === 'light' ? '#ffffff' : '#24252a');
		localStorage.setItem(STORAGE_KEY, t);
	}

	function init() {
		if (!browser) return;
		const saved = localStorage.getItem(STORAGE_KEY);
		const preferred = window.matchMedia('(prefers-color-scheme: light)').matches
			? 'light'
			: 'dark';
		set(isTheme(saved) ? saved : preferred);
	}

	function toggle() {
		set(theme === 'dark' ? 'light' : 'dark');
	}

	return {
		get value() { return theme; },
		init,
		toggle
	};
}

export const themeStore = createThemeStore();
