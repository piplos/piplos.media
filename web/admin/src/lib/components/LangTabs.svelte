<script lang="ts">
	interface LangItem {
		code: string;
		name: string;
	}

	interface Props {
		languages: LangItem[];
		activeLang: string;
		isFilled?: (code: string) => boolean;
		ariaLabel?: string;
		onSelect: (code: string) => void;
	}

	let {
		languages,
		activeLang,
		isFilled = () => false,
		ariaLabel = 'Языки контента',
		onSelect
	}: Props = $props();
</script>

<div class="tr-tabs" role="tablist" aria-label={ariaLabel}>
	{#each languages as lang (lang.code)}
		<button
			type="button"
			role="tab"
			aria-selected={activeLang === lang.code}
			class="tr-tab"
			class:tr-tab--active={activeLang === lang.code}
			onclick={() => onSelect(lang.code)}
		>
			{lang.code.toUpperCase()}
			<span class="tr-dot" class:tr-dot--filled={isFilled(lang.code)} aria-hidden="true"></span>
		</button>
	{/each}
</div>

<style>
	.tr-tabs {
		display: flex;
		gap: 0.25rem;
		padding: 0.25rem;
		background: #f4f4f5;
		border-radius: 10px;
		width: fit-content;
	}
	.tr-tab {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.tr-tab:hover {
		color: #1a1a1a;
	}
	.tr-tab--active {
		color: #1a1a1a;
		background: #fff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
	.tr-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d4d4d8;
	}
	.tr-dot--filled {
		background: #16a34a;
	}
</style>
