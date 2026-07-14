<script lang="ts">
	interface Props {
		label: string;
		variant?: 'service' | 'work' | 'stack';
	}
	let { label, variant = 'service' }: Props = $props();
</script>

<article class="grid-placeholder grid-placeholder--{variant}" role="listitem" aria-label={label}>
	<div class="grid-placeholder-icon" aria-hidden="true">
		<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
			<rect x="3" y="3" width="18" height="18" rx="4" stroke-dasharray="3 3" />
			<path d="M12 8v8M8 12h8" stroke-linecap="round" />
		</svg>
	</div>
	<p class="grid-placeholder-text">{label}</p>
</article>
