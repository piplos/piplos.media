<script lang="ts">
	interface Props {
		variant?: string;
		pill?: boolean;
		title?: string;
		class?: string;
		children?: import('svelte').Snippet;
	}
	let { variant = 'neutral', pill = false, title, class: className = '', children }: Props = $props();
</script>

<span class="badge badge--{variant} {pill ? 'badge--pill' : ''} {className}" {title}>
	{#if children}
		{@render children()}
	{/if}
</span>

<style>
	.badge {
		display: inline-block;
		box-sizing: border-box;
		font-size: 0.75rem;
		font-weight: 400;
		line-height: 1.25;
		letter-spacing: 0.02em;
		vertical-align: middle;
		padding: 0.25rem 0.5rem;
		border-radius: 6px;
	}
	.badge--pill {
		padding: 0.25rem 0.625rem;
		border-radius: 9999px;
	}
	.badge--success {
		background: rgba(22, 163, 74, 0.15);
		color: #15803d;
	}
	.badge--neutral {
		background: #f4f4f5;
		color: #71717a;
	}
	.badge--warning {
		background: rgba(234, 179, 8, 0.2);
		color: #a16207;
	}
	.badge--danger {
		background: rgba(239, 68, 68, 0.15);
		color: #b91c1c;
	}
	.badge--info {
		background: rgba(59, 130, 246, 0.15);
		color: #1d4ed8;
	}
	.badge--ru {
		background: rgba(239, 68, 68, 0.15);
		color: #b91c1c;
	}
	.badge--en {
		background: rgba(59, 130, 246, 0.15);
		color: #1d4ed8;
	}
</style>
