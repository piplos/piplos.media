<script lang="ts">
	import { SITE } from '$lib/site';

	type Props = {
		href?: string;
		variant?: 'header' | 'footer';
		label?: string;
	};

	let {
		href = '/',
		variant = 'header',
		label = `${SITE.displayName} — home`
	}: Props = $props();

	const [brandLine1, brandLine2] = SITE.displayName.split(' ');
</script>

<a {href} class="logo logo--{variant}" aria-label={label}>
	<svg
		class="logo-mark"
		viewBox="0 0 75 75"
		xmlns="http://www.w3.org/2000/svg"
		aria-hidden="true"
	>
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
		transition: opacity 0.2s;
	}

	.logo:hover {
		opacity: 0.9;
	}

	.logo-mark {
		flex-shrink: 0;
		display: block;
		fill: var(--c-logo);
	}

	.logo-text {
		display: inline-flex;
		flex-direction: row;
		align-items: center;
		gap: 0.32em;
		font-family: var(--f-brand);
		font-weight: 400;
		line-height: 1;
		letter-spacing: 0.02em;
		text-transform: uppercase;
	}

	.logo-line {
		display: block;
	}

	.logo--header {
		--logo-mark-size: 42px;
		--logo-text-size: 1.75rem;
	}

	.logo--header .logo-mark {
		width: var(--logo-mark-size);
		height: var(--logo-mark-size);
	}

	.logo--header .logo-text {
		flex-direction: row;
		align-items: center;
		height: auto;
		font-size: var(--logo-text-size);
		white-space: nowrap;
	}

	.logo--footer {
		--logo-mark-size: 44px;
		--logo-text-size: 1.75rem;
	}

	.logo--footer .logo-mark {
		width: var(--logo-mark-size);
		height: var(--logo-mark-size);
	}

	.logo--footer .logo-text {
		flex-direction: column;
		align-items: flex-start;
		justify-content: space-between;
		height: var(--logo-mark-size);
		gap: 0;
		font-size: calc(var(--logo-mark-size) * 0.46);
		line-height: 1;
	}
</style>
