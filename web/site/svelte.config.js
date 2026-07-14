import adapter from '@sveltejs/adapter-cloudflare';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		prerender: {
			handleUnseenRoutes: 'ignore',
			handleHttpError: ({ path }) => {
				// Uploads и stack-иконки отдаёт Go API, при build-time crawl их может не быть.
				if (path.startsWith('/uploads/') || path.startsWith('/stack/')) return;
			}
		},
		adapter: adapter()
	}
};

export default config;
