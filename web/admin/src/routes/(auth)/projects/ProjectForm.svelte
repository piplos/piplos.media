<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import TagSelect from '$lib/components/TagSelect.svelte';
	import TranslationsEditor from '$lib/components/TranslationsEditor.svelte';
	import type { Language, Project, SEOPage, Service, StackItem, Translations } from '$lib/types';

	interface Props {
		project?: Partial<Project>;
		seo?: Partial<SEOPage> | null;
		languages: Language[];
		services: Service[];
		stack: StackItem[];
		submitLabel: string;
	}
	let { project = {}, seo = null, languages = [], services = [], stack = [], submitLabel }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; страница перемонтирует форму через {#key}.
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(project) as Partial<Project>;
	// svelte-ignore state_referenced_locally
	const initialSeo = $state.snapshot(seo) as Partial<SEOPage> | null;
	const isEdit = Boolean(initial.id);
	let slug = $state(initial.slug ?? '');
	// svelte-ignore state_referenced_locally
	const defaultCategory =
		[...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug))[0]?.slug ?? '';
	// svelte-ignore state_referenced_locally
	let category = $state(
		initial.category && services.some((s) => s.slug === initial.category) ? initial.category : defaultCategory
	);
	let tags = $state(initial.tags ?? []);
	let year = $state(String(initial.year ?? new Date().getFullYear()));
	let featured = $state(initial.featured ?? false);
	let published = $state(initial.published ?? true);
	let translations = $state<Translations>((initial.translations ?? {}) as Translations);
	let seoTranslations = $state<Translations>((initialSeo?.translations ?? {}) as Translations);

	const seoPath = $derived(isEdit && slug ? `/portfolio/${slug}` : '');

	const stackOptions = $derived.by(() => {
		const fromStack = stack.map((item) => ({ value: item.label, label: item.label }));
		const known = new Set(fromStack.map((option) => option.value));
		const extras = tags
			.filter((tag) => !known.has(tag))
			.map((tag) => ({ value: tag, label: tag }));
		return [...extras, ...fromStack];
	});

	function serviceTitle(service: Service): string {
		const langs = Object.keys(service.translations);
		return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
	}

	const serviceOptions = $derived(
		[...services]
			.sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug))
			.map((service) => ({ value: service.slug, label: `${serviceTitle(service)} (${service.slug})` }))
	);

	const translationFields = [
		{ key: 'title', label: 'Название' },
		{ key: 'subtitle', label: 'Подзаголовок' },
		{ key: 'description', label: 'Описание', type: 'textarea' as const },
		{ key: 'challenge', label: 'Задача (challenge)', type: 'textarea' as const },
		{ key: 'solution', label: 'Решение (solution)', type: 'richtext' as const, preview: true },
		{ key: 'result', label: 'Результат (result)', type: 'textarea' as const }
	];

	const seoFields = [
		{ key: 'title', label: 'Title' },
		{ key: 'description', label: 'Meta description', type: 'textarea' as const, rows: 3 },
		{ key: 'keywords', label: 'Keywords' },
		{ key: 'og_title', label: 'OG Title' },
		{ key: 'og_description', label: 'OG Description', type: 'textarea' as const, rows: 2 },
		{ key: 'og_image', label: 'OG Image URL' }
	];

	type ContentTab = 'content' | 'seo';
	let activeTab = $state<ContentTab>('content');
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
			}
			await update({ reset: false });
		};
	}}
>
	<input type="hidden" name="translations" value={JSON.stringify(translations)} />
	{#if seoPath}
		<input type="hidden" name="seo_id" value={initialSeo?.id ?? ''} />
		<input type="hidden" name="seo_path" value={seoPath} />
		<input type="hidden" name="seo_translations" value={JSON.stringify(seoTranslations)} />
	{/if}

	<Card padding="sm">
		<div class="fields">
			<div class="grid-2">
				<FormField label="Slug" id="project-slug">
					<Input id="project-slug" name="slug" bind:value={slug} placeholder="analytics-dashboard" required />
				</FormField>
				<FormField label="Год" id="project-year">
					<Input id="project-year" name="year" type="number" bind:value={year} />
				</FormField>
			</div>
			<FormField label="Группа (услуга)" id="project-category">
				<Select
					id="project-category"
					name="category"
					options={serviceOptions}
					bind:value={category}
					disabled={!serviceOptions.length}
				/>
			</FormField>
			<FormField label="Стек" id="project-stack">
				<TagSelect
					id="project-stack"
					name="tags"
					options={stackOptions}
					bind:values={tags}
					placeholder={stackOptions.length ? 'Выберите технологии' : 'Список стека пуст'}
				/>
			</FormField>
			<div class="checks-row">
				<label class="check">
					<input type="checkbox" name="featured" bind:checked={featured} />
					Избранный (featured)
				</label>
				<label class="check">
					<input type="checkbox" name="published" bind:checked={published} />
					Опубликован
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
				<TranslationsEditor {languages} fields={translationFields} bind:translations idPrefix="project" />
			{:else}
				<FormField label="Путь страницы" id="seo-path" hint="Соответствует URL кейса на сайте">
					<p class="seo-path">{seoPath}</p>
				</FormField>
				<div class="seo-editor">
					<h3 class="subsection-title">Meta-теги по языкам</h3>
					<TranslationsEditor
						{languages}
						fields={seoFields}
						bind:translations={seoTranslations}
						idPrefix="project-seo"
					/>
				</div>
			{/if}
		</div>
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
	.fields {
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
	.checks-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1.5rem;
	}
	.check {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}
	.section-title {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
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
		justify-content: flex-end;
	}
</style>
