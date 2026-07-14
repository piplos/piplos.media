<script lang="ts">
	import type { ProjectLinkKind } from '$lib/project-links';

	type Props = {
		kind: ProjectLinkKind;
	};

	let { kind }: Props = $props();
</script>

{#if kind === 'google_play'}
	<img class="link-icon" src="/stores/google-play.svg" alt="" width="20" height="20" loading="lazy" decoding="async" />
{:else if kind === 'app_store'}
	<svg class="link-icon link-icon--svg" width="20" height="20" viewBox="0 0 24 24" aria-hidden="true">
		<path
			fill="currentColor"
			d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"
		/>
	</svg>
{:else}
	<svg class="link-icon link-icon--svg" width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
		<circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.5"/>
		<path d="M3 12h18M12 3c-2.5 2.8-3.8 5.8-3.8 9s1.3 6.2 3.8 9M12 3c2.5 2.8 3.8 5.8 3.8 9s-1.3 6.2-3.8 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
	</svg>
{/if}

<style>
	.link-icon {
		display: block;
		width: 20px;
		height: 20px;
		flex-shrink: 0;
		object-fit: contain;
	}

	.link-icon--svg {
		color: var(--c-muted);
	}
</style>
