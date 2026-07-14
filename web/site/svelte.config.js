import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		prerender: {
			handleUnseenRoutes: 'ignore'
		},
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			strict: true
		})
	}
};

export default config;
