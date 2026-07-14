<script lang="ts">
	import { LEGAL_PLACEHOLDERS, legalPlaceholderToken } from '$lib/legal-placeholders';
</script>

<aside class="placeholders-hint" aria-label="Доступные подстановки в тексте">
	<p class="placeholders-title">Подстановки в тексте</p>
	<p class="placeholders-desc">
		В заголовках, дате версии и тексте разделов можно использовать переменные — на сайте они
		заменятся на актуальные значения:
	</p>
	<ul class="placeholders-list">
		{#each LEGAL_PLACEHOLDERS as item (item.key)}
			<li>
				<code>{legalPlaceholderToken(item.key)}</code>
				<span class="placeholders-label">{item.label}</span>
				<span class="placeholders-example">— {item.example}</span>
			</li>
		{/each}
	</ul>
</aside>

<style>
	.placeholders-hint {
		padding: 0.875rem 1rem;
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 10px;
	}
	.placeholders-title {
		margin: 0;
		font-size: 0.8125rem;
		font-weight: 600;
		color: #18181b;
	}
	.placeholders-desc {
		margin: 0.375rem 0 0.625rem;
		font-size: 0.8125rem;
		line-height: 1.5;
		color: #71717a;
	}
	.placeholders-list {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}
	.placeholders-list li {
		font-size: 0.8125rem;
		line-height: 1.45;
		color: #3f3f46;
	}
	.placeholders-list code {
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.75rem;
		padding: 0.125rem 0.375rem;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 4px;
		color: #18181b;
	}
	.placeholders-label {
		margin-left: 0.375rem;
	}
	.placeholders-example {
		color: #71717a;
	}
</style>
