<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import LangTabs from '$lib/components/LangTabs.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import type { LegalSection, LegalTranslations } from '$lib/types';

	interface Props {
		languages: { code: string; name: string; is_default: boolean }[];
		translations: LegalTranslations;
		idPrefix?: string;
	}
	let { languages, translations = $bindable(), idPrefix = 'legal' }: Props = $props();

	const defaultLang = $derived(languages.find((l) => l.is_default)?.code ?? languages[0]?.code ?? 'en');
	let activeLang = $state('');

	$effect(() => {
		if (!activeLang && languages.length) activeLang = defaultLang;
	});

	function ensureLang(code: string) {
		if (!translations[code]) {
			translations[code] = { label: '', title: '', last_updated: '', sections: [] };
		}
		if (!translations[code].sections) {
			translations[code].sections = [];
		}
	}

	function activeLocale() {
		ensureLang(activeLang);
		return translations[activeLang];
	}

	function addSection() {
		ensureLang(activeLang);
		translations[activeLang].sections = [...translations[activeLang].sections, { title: '', body: '' }];
	}

	function removeSection(index: number) {
		ensureLang(activeLang);
		translations[activeLang].sections = translations[activeLang].sections.filter((_, i) => i !== index);
	}

	function langFilled(code: string): boolean {
		const t = translations[code];
		if (!t) return false;
		return Boolean(t.title?.trim() || t.sections?.some((s) => s.title.trim() || s.body.trim()));
	}
</script>

<div class="legal-editor">
	<LangTabs
		{languages}
		{activeLang}
		isFilled={langFilled}
		ariaLabel="Языки документа"
		onSelect={(code) => {
			activeLang = code;
			ensureLang(code);
		}}
	/>

	{#if activeLang}
		{@const locale = activeLocale()}
		<div class="lang-panel" role="tabpanel">
			<div class="grid-2">
				<FormField label="Метка (label)" id="{idPrefix}-{activeLang}-label">
					<Input id="{idPrefix}-{activeLang}-label" bind:value={locale.label} placeholder="Legal" />
				</FormField>
				<FormField label="Заголовок страницы" id="{idPrefix}-{activeLang}-title">
					<Input id="{idPrefix}-{activeLang}-title" bind:value={locale.title} required />
				</FormField>
			</div>
			<FormField label="Дата версии" id="{idPrefix}-{activeLang}-updated">
				<Input
					id="{idPrefix}-{activeLang}-updated"
					bind:value={locale.last_updated}
					placeholder="This version takes effect on ..."
				/>
			</FormField>

			<div class="sections-block">
				<div class="sections-head">
					<h3 class="sections-title">Разделы</h3>
					<Button type="button" variant="secondary" onclick={addSection}>+ Раздел</Button>
				</div>
				{#if !locale.sections.length}
					<p class="text-muted">Разделов пока нет. Добавьте первый.</p>
				{:else}
					{#each locale.sections as section, index (index)}
						<div class="section-card">
							<div class="section-card-head">
								<span class="section-num">#{index + 1}</span>
								<button
									type="button"
									class="section-remove"
									aria-label="Удалить раздел"
									onclick={() => removeSection(index)}
								>
									Удалить
								</button>
							</div>
							<FormField label="Заголовок раздела" id="{idPrefix}-{activeLang}-sec-{index}-title">
								<Input id="{idPrefix}-{activeLang}-sec-{index}-title" bind:value={section.title} />
							</FormField>
							<FormField label="Текст" id="{idPrefix}-{activeLang}-sec-{index}-body">
								<Textarea
									id="{idPrefix}-{activeLang}-sec-{index}-body"
									bind:value={section.body}
									rows={8}
								/>
							</FormField>
						</div>
					{/each}
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.legal-editor {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.lang-panel {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.grid-2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}
	@media (max-width: 640px) {
		.grid-2 {
			grid-template-columns: 1fr;
		}
	}
	.sections-block {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		padding-top: 0.5rem;
		border-top: 1px solid #f4f4f5;
	}
	.sections-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}
	.sections-title {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
		color: #18181b;
	}
	.section-card {
		padding: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		background: #fafafa;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.section-card-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.section-num {
		font-size: 0.75rem;
		font-weight: 600;
		color: #71717a;
	}
	.section-remove {
		padding: 0;
		font-size: 0.8125rem;
		color: #b91c1c;
		background: none;
		border: none;
		cursor: pointer;
	}
	.section-remove:hover {
		text-decoration: underline;
	}
</style>
