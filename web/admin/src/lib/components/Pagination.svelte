<script lang="ts">
	import { pageItems } from '$lib/pagination';

	interface Props {
		/** Текущая страница (с 1). */
		page: number;
		/** Всего страниц (всегда ≥ 1). */
		totalPages: number;
		/** Ссылка на страницу n (для n=1 обычно чистый URL без ?page). */
		makeHref: (page: number) => string;
	}

	let { page, totalPages, makeHref }: Props = $props();

	const items = $derived(pageItems(page, totalPages));
</script>

<nav class="pagination" aria-label="Постраничная навигация">
	{#if page > 1}
		<a href={makeHref(page - 1)} class="pagination-btn pagination-btn--arrow" aria-label="Предыдущая страница">←</a>
	{:else}
		<span class="pagination-btn pagination-btn--arrow pagination-btn--disabled" aria-hidden="true">←</span>
	{/if}

	{#each items as item, i (i)}
		{#if item === 'gap'}
			<span class="pagination-gap" aria-hidden="true">…</span>
		{:else if item === page}
			<span class="pagination-btn pagination-btn--active" aria-current="page">{item}</span>
		{:else}
			<a href={makeHref(item)} class="pagination-btn">{item}</a>
		{/if}
	{/each}

	{#if page < totalPages}
		<a href={makeHref(page + 1)} class="pagination-btn pagination-btn--arrow" aria-label="Следующая страница">→</a>
	{:else}
		<span class="pagination-btn pagination-btn--arrow pagination-btn--disabled" aria-hidden="true">→</span>
	{/if}
</nav>

<style>
	.pagination {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		margin-top: 0.75rem;
	}
	.pagination-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		box-sizing: border-box;
		min-width: 2rem;
		height: 2rem;
		padding: 0 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		line-height: 1;
		color: #52525b;
		background: #fff;
		border: 1px solid transparent;
		border-radius: 6px;
		text-decoration: none;
		transition: color 0.15s, background 0.15s;
	}
	a.pagination-btn:hover {
		color: #18181b;
		background: #f4f4f5;
	}
	a.pagination-btn:focus-visible {
		outline: 2px solid #7c3aed;
		outline-offset: 2px;
	}
	.pagination-btn--active {
		color: #fff;
		background: #18181b;
		font-weight: 600;
	}
	.pagination-btn--disabled {
		color: #d1d5db;
	}
	.pagination-gap {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 1.5rem;
		height: 2rem;
		color: #a1a1aa;
	}
</style>