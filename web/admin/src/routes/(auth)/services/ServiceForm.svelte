<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import SlugInput from '$lib/components/SlugInput.svelte';
	import TagSelect from '$lib/components/TagSelect.svelte';
	import TranslationsEditor from '$lib/components/TranslationsEditor.svelte';
	import { DEFAULT_SERVICE_ICON, SERVICE_ICON_OPTIONS } from '$lib/service-icons';
	import type { Language, SEOPage, Service, StackItem, Translations } from '$lib/types';

	interface Props {
		service?: Partial<Service>;
		seo?: Partial<SEOPage> | null;
		languages: Language[];
		stack: StackItem[];
		submitLabel: string;
	}
	let { service = {}, seo = null, languages, stack, submitLabel }: Props = $props();

	let submitting = $state(false);
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(service) as Partial<Service>;
	// svelte-ignore state_referenced_locally
	const initialSeo = $state.snapshot(seo) as Partial<SEOPage> | null;
	const isEdit = Boolean(initial.id);
	let slug = $state(initial.slug ?? '');
	let icon = $state(initial.icon || DEFAULT_SERVICE_ICON);
	let tags = $state(initial.tags ?? []);
	let sortOrder = $state(String(initial.sort_order ?? 0));
	let published = $state(initial.published ?? true);
	let translations = $state<Translations>((initial.translations ?? {}) as Translations);
	let seoTranslations = $state<Translations>((initialSeo?.translations ?? {}) as Translations);

	// При очистке поля slug используем исходный slug, чтобы табы «Контент/SEO» не пропадали.
	const seoSlug = $derived(slug.trim() || initial.slug || '');
	const seoPath = $derived(isEdit && seoSlug ? `/services/${seoSlug}` : '');

	const defaultLang = $derived(languages.find((l) => l.is_default)?.code ?? languages[0]?.code ?? 'en');
	const slugSource = $derived.by(() => {
		const fromDefault = translations[defaultLang]?.title?.trim();
		if (fromDefault) return fromDefault;
		for (const lang of Object.keys(translations)) {
			const title = translations[lang]?.title?.trim();
			if (title) return title;
		}
		return '';
	});

	const translationFields = [
		{ key: 'title', label: 'Название' },
		{ key: 'description', label: 'Краткое описание', type: 'textarea' as const, rows: 3 },
		{ key: 'body', label: 'Подробное описание страницы (Markdown)', type: 'markdown' as const }
	];

	const seoFields = [
		{ key: 'title', label: 'Title' },
		{ key: 'description', label: 'Meta description', type: 'textarea' as const, rows: 3 },
		{ key: 'keywords', label: 'Keywords' },
		{ key: 'og_title', label: 'OG Title' },
		{ key: 'og_description', label: 'OG Description', type: 'textarea' as const, rows: 2 },
		{ key: 'og_image', label: 'OG Image URL', type: 'image' as const }
	];

	type ContentTab = 'content' | 'seo';
	let activeTab = $state<ContentTab>('content');

	const stackOptions = $derived(stack.map((item) => ({ value: item.label, label: item.label })));

	const iconOptions = $derived.by(() => {
		const known = SERVICE_ICON_OPTIONS.map((o) => ({ value: o.value, label: o.label }));
		if (icon && !known.some((o) => o.value === icon)) {
			return [{ value: icon, label: `${icon} (текущая)` }, ...known];
		}
		return known;
	});
</script>

<form
	id="service-save-form"
	method="POST"
	action="?/save"
	class="content-form"
	use:enhance={() => {
		submitting = true;
		return async ({ result }) => {
			submitting = false;
			if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			}
		};
	}}
>
	<input type="hidden" name="id" value={initial.id ?? ''} />
	<input type="hidden" name="translations" value={JSON.stringify(translations)} />
	{#if seoPath}
		<input type="hidden" name="seo_id" value={initialSeo?.id ?? ''} />
		<input type="hidden" name="seo_path" value={seoPath} />
		<input type="hidden" name="seo_translations" value={JSON.stringify(seoTranslations)} />
	{/if}

	<Card padding="sm">
		<div class="fields">
			<div class="grid-2">
				<FormField label="Slug" id="svc-slug">
					<SlugInput id="svc-slug" name="slug" bind:value={slug} placeholder="web" required source={slugSource} />
				</FormField>
				<FormField label="Иконка (символ)" id="svc-icon">
					<div class="icon-field">
						<span class="icon-preview" aria-hidden="true">{icon}</span>
						<Select id="svc-icon" name="icon" bind:value={icon} options={iconOptions} />
					</div>
				</FormField>
			</div>
			<FormField label="Стек" id="svc-tags">
				<TagSelect
					id="svc-tags"
					name="tags"
					options={stackOptions}
					bind:values={tags}
					placeholder="Выберите технологии"
				/>
			</FormField>
			<FormField label="Порядок сортировки" id="svc-sort">
				<Input id="svc-sort" name="sort_order" type="number" bind:value={sortOrder} />
			</FormField>
			<div class="checks-row">
				<label class="check">
					<input type="checkbox" name="published" bind:checked={published} />
					Опубликована
				</label>
			</div>
		</div>
	</Card>

	<Card padding="sm">
		<div class="fields">
			{#if seoPath}
				<div class="form-tabs" aria-label="Разделы контента" role="tablist">
					<button
						type="button"
						role="tab"
						class="form-tab"
						class:form-tab--active={activeTab === 'content'}
						aria-selected={activeTab === 'content'}
						onclick={() => (activeTab = 'content')}
					>
						Контент по языкам
					</button>
					<button
						type="button"
						role="tab"
						class="form-tab"
						class:form-tab--active={activeTab === 'seo'}
						aria-selected={activeTab === 'seo'}
						onclick={() => (activeTab = 'seo')}
					>
						SEO
					</button>
				</div>
			{:else}
				<h2 class="section-title">Контент по языкам</h2>
			{/if}

			{#if !seoPath || activeTab === 'content'}
				<p class="section-hint">
					Краткое описание — для карточек и шапки страницы. Подробное описание — Markdown с
					картинками для отдельной страницы услуги на сайте.
				</p>
				<TranslationsEditor {languages} fields={translationFields} bind:translations idPrefix="svc" />
			{:else}
				<FormField label="Путь страницы" id="svc-seo-path">
					<p class="seo-path">{seoPath}</p>
				</FormField>
				<div class="seo-editor">
					<h3 class="subsection-title">Meta-теги по языкам</h3>
					<TranslationsEditor
						{languages}
						fields={seoFields}
						bind:translations={seoTranslations}
						idPrefix="svc-seo"
					/>
				</div>
			{/if}
		</div>
	</Card>
</form>

<div class="form-actions">
	{#if isEdit}
		<form
			method="POST"
			action="?/delete"
			class="delete-form"
			use:enhance={deleteEnhance({
				message: 'Удалить услугу?',
				onSuccess: async () => {},
				onError: () => toast.error('Не удалось удалить услугу')
			})}
		>
			<input type="hidden" name="id" value={initial.id} />
			<Button type="submit" variant="danger">Удалить</Button>
		</form>
	{/if}
	<Button type="submit" form="service-save-form" loading={submitting}>{submitLabel}</Button>
</div>

<style>
	.content-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.fields {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.grid-2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		align-items: start;
	}
	@media (max-width: 640px) {
		.grid-2 {
			grid-template-columns: 1fr;
		}
	}
	.icon-field {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.icon-preview {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2.25rem;
		height: 2.25rem;
		flex-shrink: 0;
		font-size: 1.125rem;
		line-height: 1;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		background: #f9fafb;
	}
	.icon-field :global(.select) {
		flex: 1;
	}
	.section-title {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
	}
	.section-hint {
		margin: 0;
		font-size: 0.8125rem;
		color: #71717a;
		line-height: 1.5;
	}
	.form-tabs {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		padding: 0.25rem;
		margin: 0;
		border-radius: 10px;
		background: #f4f4f5;
	}
	.form-tab {
		padding: 0.5rem 0.875rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.form-tab:hover {
		color: #1a1a1a;
	}
	.form-tab--active {
		color: #1a1a1a;
		background: #fff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
	.form-tab:focus,
	.form-tab:focus-visible {
		outline: none;
	}
	.subsection-title {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
		color: #18181b;
	}
	.seo-editor {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.seo-path {
		margin: 0;
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5;
		color: #52525b;
		background: #f4f4f5;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.form-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}
	.delete-form {
		margin: 0;
	}
</style>
