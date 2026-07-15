<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import LegalContentEditor from '$lib/components/LegalContentEditor.svelte';
	import type { Language, LegalPage, LegalTranslations } from '$lib/types';

	interface Props {
		page: LegalPage;
		languages: Language[];
		submitLabel: string;
	}
	let { page, languages, submitLabel }: Props = $props();

	let submitting = $state(false);
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(page);
	let translations = $state<LegalTranslations>(structuredClone(initial.translations ?? {}));
</script>

<form
	method="POST"
	action="?/save"
	class="content-form"
	use:enhance={() => {
		submitting = true;
		return async ({ result, update }) => {
			submitting = false;
			if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			} else if (result.type === 'success') {
				toast.success('Документ сохранён');
			}
			await update({ reset: false });
		};
	}}
>
	<input type="hidden" name="translations" value={JSON.stringify(translations)} />

	<Card padding="sm">
		<FormField label="Путь на сайте" id="legal-path">
			<p class="legal-path">/{'{lang}'}/legal/{page.slug}</p>
		</FormField>
		<p class="legal-hint">
			SEO для legal-страниц заблокировано. На сайте для этих страниц установлен <code>noindex, nofollow</code>
			— они не индексируются поисковиками.
		</p>
	</Card>

	<Card padding="sm">
		<h2 class="section-title">Содержание документа</h2>
		<LegalContentEditor {languages} bind:translations idPrefix="legal-{page.slug}" />
	</Card>

	<div class="form-actions">
		<Button type="submit" loading={submitting}>{submitLabel}</Button>
	</div>
</form>

<style>
	.content-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.legal-path {
		margin: 0;
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5;
		color: #52525b;
		background: #f4f4f5;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.legal-hint {
		margin: 0.75rem 0 0;
		font-size: 0.8125rem;
		line-height: 1.5;
		color: #71717a;
	}
	.legal-hint code {
		font-size: 0.75rem;
		padding: 0.125rem 0.375rem;
		background: #f4f4f5;
		border-radius: 4px;
	}
	.section-title {
		margin: 0 0 0.75rem;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
	}
	.form-actions {
		display: flex;
		justify-content: flex-end;
	}
</style>
