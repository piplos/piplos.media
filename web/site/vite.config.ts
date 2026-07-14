import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { svelteStyleFallback } from '../vite/svelte-style-fallback';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		svelteStyleFallback()
	],
	server: {
		port: 5173,
		strictPort: true,
		allowedHosts: true
	}
});
