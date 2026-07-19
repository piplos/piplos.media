<script lang="ts">
	import { beforeNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { useTabLayout } from '$lib/tab-layout.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	const layout = useTabLayout();

	// Пока сниппет страницы отрендерен в родительском layout, SvelteKit не может
	// размонтировать страницу — клиентская навигация обновляет URL, но не контент.
	// Сбрасываем actions до ухода на другой pathname; навигации в пределах
	// текущей страницы (тот же URL, query) не размонтируют её — actions остаются.
	let mounted = $state(true);

	beforeNavigate(({ to }) => {
		if (to?.url.pathname === page.url.pathname) return;
		mounted = false;
		layout.setActions(null);
	});

	$effect(() => {
		if (!mounted) return;
		layout.setActions(children);
		return () => layout.setActions(null);
	});
</script>
