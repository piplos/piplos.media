<script lang="ts">
	import { browser } from '$app/environment';
	import toast from 'svelte-french-toast';
	import type { Language, Translations } from '$lib/types';
	import Button from '$lib/components/Button.svelte';
	import ImageField from '$lib/components/ImageField.svelte';
	import LangTabs from '$lib/components/LangTabs.svelte';
	import Input from '$lib/components/Input.svelte';
	import MarkdownEditor from '$lib/components/MarkdownEditor.svelte';
	import Textarea from '$lib/components/Textarea.svelte';

	interface FieldDef {
		key: string;
		label: string;
		/** markdown — моноширинный textarea (HTML генерируется на стороне API);
		 *  image — URL картинки с выбором из файлового архива. */
		type?: 'input' | 'textarea' | 'markdown' | 'image';
		rows?: number;
	}
	interface Props {
		languages: Language[];
		fields: FieldDef[];
		translations: Translations;
		idPrefix?: string;
		/** Папка архива для загрузок в markdown/image-полях (например, projects/site-dev). */
		uploadPath?: string;
	}
	let { languages, fields, translations = $bindable(), idPrefix = 'tr', uploadPath = '' }: Props = $props();

	const defaultLang = $derived(languages.find((l) => l.is_default)?.code ?? languages[0]?.code ?? 'en');
	let activeLang = $state('');
	let sourceLang = $state('');
	let translating = $state(false);
	let translatePanelOpen = $state(false);

	$effect(() => {
		if (!activeLang && languages.length) activeLang = defaultLang;
		if (!sourceLang && languages.length) sourceLang = defaultLang;
		if (activeLang && activeLang === sourceLang) {
			const alt = languages.find((l) => l.code !== activeLang);
			if (alt) sourceLang = alt.code;
		}
	});

	function ensureLang(code: string) {
		if (!translations[code]) translations[code] = {};
	}

	function fieldText(code: string, key: string): string {
		return (translations[code]?.[key] ?? '').trim();
	}

	function langFilled(code: string): boolean {
		const t = translations[code];
		if (!t) return false;
		return fields.some((f) => fieldText(code, f.key) !== '');
	}

	async function translateActive() {
		if (!browser) return;
		const source = translations[sourceLang] ?? {};
		const payload: Record<string, string> = {};
		for (const f of fields) {
			const v = (source[f.key] ?? '').trim();
			if (v) payload[f.key] = v;
		}
		if (Object.keys(payload).length === 0) {
			toast.error(`Нет текста на языке «${sourceLang}» для перевода`);
			return;
		}
		translating = true;
		try {
			const res = await fetch('/api/translate', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ fields: payload, target_lang: activeLang })
			});
			const data = (await res.json().catch(() => ({}))) as {
				fields?: Record<string, string>;
				message?: string;
			};
			if (!res.ok || !data.fields) {
				toast.error(data.message ?? 'Не удалось выполнить перевод');
				return;
			}
			ensureLang(activeLang);
			for (const [key, value] of Object.entries(data.fields)) {
				translations[activeLang][key] = value;
			}
			toast.success(`Переведено на «${activeLang}»`);
		} catch {
			toast.error('Сервис перевода недоступен');
		} finally {
			translating = false;
		}
	}

	function canTranslateActive(): boolean {
		if (!activeLang || activeLang === sourceLang) return false;
		const source = translations[sourceLang] ?? {};
		return fields.some((f) => (source[f.key] ?? '').trim() !== '');
	}

	function setField(key: string, value: string) {
		ensureLang(activeLang);
		translations[activeLang][key] = value;
	}
</script>

<div class="tr-editor">
	<div class="practice-edit-badges-row">
		<div class="practice-edit-lang-badges">
			<LangTabs
				{languages}
				{activeLang}
				isFilled={langFilled}
				onSelect={(code) => {
					ensureLang(code);
					activeLang = code;
				}}
			/>
		</div>
		{#if languages.length > 1}
			<button
				type="button"
				class="admin-action-btn practice-edit-ai-trigger"
				class:practice-edit-ai-trigger--active={translatePanelOpen}
				onclick={() => (translatePanelOpen = !translatePanelOpen)}
				aria-expanded={translatePanelOpen}
				aria-label="Автоматический перевод"
				title="Автоматический перевод"
			>
				<svg
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path
						d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09L9 18.75l-.813-2.846z"
					/>
					<path
						d="M18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456L18 9.75z"
					/>
				</svg>
			</button>
		{/if}
	</div>

	{#if translatePanelOpen && languages.length > 1}
		<div class="practice-edit-translate-panel">
			<div class="practice-edit-translate-row">
				<label class="practice-edit-label" for="{idPrefix}-source-lang">Источник</label>
				<select
					id="{idPrefix}-source-lang"
					class="practice-edit-source-select"
					bind:value={sourceLang}
					aria-label="Исходный язык"
				>
					{#each languages.filter((l) => l.code !== activeLang) as lang (lang.code)}
						<option value={lang.code}>{lang.name} ({lang.code})</option>
					{/each}
				</select>
				<Button
					type="button"
					variant="secondary"
					disabled={translating || !canTranslateActive()}
					loading={translating}
					onclick={translateActive}
				>
					Текущий язык
				</Button>
			</div>
		</div>
	{/if}

	{#if activeLang}
		{#each fields as field (field.key)}
			{@const fieldId = `${idPrefix}-${activeLang}-${field.key}`}
			<div class="tr-field">
				<label class="tr-label" for={fieldId}>{field.label}</label>
				{#if field.type === 'markdown'}
					<MarkdownEditor
						id={fieldId}
						rows={field.rows ?? 12}
						{uploadPath}
						bind:value={
							() => translations[activeLang]?.[field.key] ?? '',
							(v) => {
								ensureLang(activeLang);
								translations[activeLang][field.key] = v;
							}
						}
					/>
					<p class="tr-hint">
						Поддерживается Markdown. HTML для сайта генерируется автоматически при выводе через API.
						Кнопка «▦ Блок» вставляет переменную с проектами или услугами.
					</p>
				{:else if field.type === 'textarea'}
					<Textarea
						id={fieldId}
						rows={field.rows ?? 4}
						bind:value={
							() => translations[activeLang]?.[field.key] ?? '',
							(v) => {
								ensureLang(activeLang);
								translations[activeLang][field.key] = v;
							}
						}
					/>
				{:else if field.type === 'image'}
					<ImageField
						id={fieldId}
						{uploadPath}
						alt={field.label}
						bind:value={
							() => translations[activeLang]?.[field.key] ?? '',
							(v) => setField(field.key, v)
						}
					/>
				{:else}
					<Input
						id={fieldId}
						bind:value={
							() => translations[activeLang]?.[field.key] ?? '',
							(v) => {
								ensureLang(activeLang);
								translations[activeLang][field.key] = v;
							}
						}
					/>
				{/if}
			</div>
		{/each}
	{/if}
</div>

<style>
	.tr-editor {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.tr-field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.tr-label {
		font-size: 0.8125rem;
		font-weight: 500;
		color: #52525b;
	}
	.tr-hint {
		margin: 0;
		font-size: 0.75rem;
		color: #a1a1aa;
	}
</style>
