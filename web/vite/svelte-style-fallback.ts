import { readFileSync } from 'node:fs';
import type { Plugin } from 'vite';

const VIRTUAL_STYLE_RE = /[?&]svelte&type=style&lang\.css/;
const STYLE_BLOCK_RE = /<style[^>]*>([\s\S]*?)<\/style>/;

/**
 * When vite-plugin-svelte cannot serve a cached virtual CSS module (cold SSR/dev start),
 * Vite falls back to the raw .svelte file and @tailwindcss/vite crashes parsing JS as CSS.
 */
export function svelteStyleFallback(): Plugin {
	return {
		name: 'svelte-style-fallback',
		enforce: 'post',
		load(id) {
			if (!VIRTUAL_STYLE_RE.test(id)) return;

			const filename = id.split('?')[0];
			try {
				const source = readFileSync(filename, 'utf-8');
				const match = STYLE_BLOCK_RE.exec(source);
				if (!match) return { code: '', moduleType: 'css' as const };

				return {
					code: match[1],
					moduleType: 'css' as const,
					meta: { vite: { cssScopeTo: [filename, 'default'] } }
				};
			} catch {
				return;
			}
		}
	};
}
