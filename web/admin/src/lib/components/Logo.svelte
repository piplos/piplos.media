<script lang="ts">
	import { SITE } from '$lib/site';

	type Props = {
		href?: string;
		label?: string;
	};

	let { href = '/', label = `${SITE.displayName} — home` }: Props = $props();

	const [brandLine1, brandLine2] = SITE.displayName.split(' ');
</script>

<a {href} class="logo" aria-label={label}>
	<svg class="logo-mark" viewBox="0 0 75 75" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
		<path
			d="M2.5.5h70c1.1045695 0 2 .8954305 2 2v70c0 1.1045695-.8954305 2-2 2h-70c-1.1045695 0-2-.8954305-2-2v-70c0-1.1045695.8954305-2 2-2zM17 41.36225l9.0088326 7.2212457v7.5805026h3.3894618V51.379923L37.5936562 58 58 41.3604447l-2.1692555-2.4371704h-2.5866946V20H41.7073924v13.5091453h-8.4147848V20H21.8754732v18.7445484h-2.8239569L17 41.36225z"
		/>
	</svg>
	<span class="logo-text" aria-hidden="true">
		<span class="logo-line">{brandLine1}</span>
		<span class="logo-line">{brandLine2}</span>
	</span>
</a>

<style>
	.logo {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		color: var(--c-logo-text);
		text-decoration: none;
		transition: opacity 0.2s;
		flex-shrink: 0;
	}

	.logo:hover {
		opacity: 0.9;
	}

	.logo-mark {
		flex-shrink: 0;
		display: block;
		width: 34px;
		height: 34px;
		fill: var(--c-logo);
	}

	.logo-text {
		display: inline-flex;
		flex-direction: row;
		align-items: center;
		gap: 0.32em;
		font-family: var(--f-brand);
		font-size: 1.25rem;
		font-weight: 400;
		line-height: 1;
		letter-spacing: 0.02em;
		text-transform: uppercase;
		white-space: nowrap;
	}

	.logo-line {
		display: block;
	}
</style>
