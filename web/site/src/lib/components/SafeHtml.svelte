<script lang="ts">
	import { browser } from '$app/environment';

	interface Props {
		html: string;
		class?: string;
	}

	let { html, class: className = '' }: Props = $props();

	const ssrHtml = $derived(html.trim());
	let clientHtml = $state<string | null>(null);

	$effect(() => {
		if (!browser) return;

		const trimmed = html.trim();
		if (!trimmed) {
			clientHtml = '';
			return;
		}

		let cancelled = false;

		import('$lib/sanitize-html').then(async ({ sanitizeCaseHtmlAsync }) => {
			const sanitized = await sanitizeCaseHtmlAsync(trimmed);
			if (!cancelled) clientHtml = sanitized;
		});

		return () => {
			cancelled = true;
		};
	});
</script>

<!-- eslint-disable-next-line svelte/no-at-html-tags — sanitized async on client; SSR uses trusted API HTML -->
<div class={className}>{@html clientHtml ?? ssrHtml}</div>
