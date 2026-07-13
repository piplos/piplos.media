<script lang="ts">
	interface Props {
		class?: string;
		padding?: 'none' | 'sm' | 'md' | 'lg';
		children?: import('svelte').Snippet;
	}
	let { class: className = '', padding = 'md', children }: Props = $props();
</script>

<div class="card padding-{padding} {className}">
	{#if children}
		{@render children()}
	{/if}
</div>

<style>
	.card {
		background: #fff;
		border-radius: 16px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
		overflow: visible;
	}
	.padding-none {
		padding: 0;
	}
	.padding-sm {
		padding: 1rem 1.25rem;
	}
	.padding-md {
		padding: 2rem 2rem 2.25rem;
	}
	.padding-lg {
		padding: 2.5rem;
	}
</style>
