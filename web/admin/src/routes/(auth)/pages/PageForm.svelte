<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import ImageField from '$lib/components/ImageField.svelte';
	import Input from '$lib/components/Input.svelte';
	import SlugInput from '$lib/components/SlugInput.svelte';
	import TagSelect from '$lib/components/TagSelect.svelte';
	import TranslationsEditor from '$lib/components/TranslationsEditor.svelte';
	import { entityFolder } from '$lib/files';
	import type { Language, Page, SEOPage, StackItem, Translations } from '$lib/types';

	interface Props {
		page?: Partial<Page>;
		seo?: Partial<SEOPage> | null;
		languages: Language[];
		stack?: StackItem[];
		submitLabel: string;
	}
	let { page = {}, seo = null, languages = [], stack = [], submitLabel }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; страница перемонтирует форму через {#key}.
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(page) as Partial<Page>;
	// svelte-ignore state_referenced_locally
	const initialSeo = $state.snapshot(seo) as Partial<SEOPage> | null;

	let slug = $state(initial.slug ?? '');
	// Новые страницы публикуются сразу (как проекты/услуги); черновик — снять галочку.
	let published = $state(initial.published ?? true);
	let image = $state(initial.image ?? '');
	let tags = $state(initial.tags ?? []);
	let translations = $state<Translations>((initial.translations ?? {}) as Translations);
	let seoTranslations = $state<Translations>((initialSeo?.translations ?? {}) as Translations);

	/** ISO → значение для input type="datetime-local" в локальном времени. */
	function toLocalInput(iso: string | null | undefined): string {
		if (!iso) return '';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return '';
		const pad = (n: number) => String(n).padStart(2, '0');
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}

	let publishAtLocal = $state(toLocalInput(initial.publish_at));
	// В скрытое поле уходит ISO (UTC): datetime-local интерпретируется в браузере пользователя.
	const publishAtISO = $derived.by(() => {
		if (!publishAtLocal) return '';
		const d = new Date(publishAtLocal);
		return Number.isNaN(d.getTime()) ? '' : d.toISOString();
	});

	// При очистке поля slug используем исходный slug, чтобы табы «Контент/SEO» не пропадали.
	const seoSlug = $derived(slug.trim() || initial.slug || '');
	// SEO-ключ в БД без языка (как у проектов/услуг); в UI показываем публичный URL с /{lang}.
	const seoPath = $derived(seoSlug ? `/articles/${seoSlug}` : '');
	const seoPathDisplay = $derived(seoPath ? `/{lang}${seoPath}` : '');
	// Папка в файловом архиве для картинок этой страницы (создаётся при первой загрузке).
	const uploadPath = $derived(entityFolder('pages', seoSlug));

	const defaultLang = $derived(languages.find((l) => l.is_default)?.code ?? languages[0]?.code ?? 'en');
	const stackOptions = $derived(stack.map((item) => ({ value: item.label, label: item.label })));
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
		{ key: 'title', label: 'Заголовок' },
		{ key: 'description', label: 'Краткое описание (анонс)', type: 'textarea' as const, rows: 3 },
		{ key: 'body', label: 'Текст (Markdown)', type: 'markdown' as const }
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
				toast.success('Страница сохранена');
			}
			await update({ reset: false });
		};
	}}
>
	<input type="hidden" name="translations" value={JSON.stringify(translations)} />
	<input type="hidden" name="publish_at" value={publishAtISO} />
	<input type="hidden" name="seo_id" value={initialSeo?.id ?? ''} />
	<input type="hidden" name="seo_path" value={seoPath} />
	<input type="hidden" name="seo_translations" value={JSON.stringify(seoTranslations)} />

	<Card padding="sm">
		<div class="fields">
			<div class="grid-2">
				<FormField label="Slug" id="page-slug">
					<SlugInput
						id="page-slug"
						name="slug"
						bind:value={slug}
						placeholder="company-news"
						required
						source={slugSource}
					/>
				</FormField>
				<FormField label="Отложенная публикация" id="page-publish-at">
					<Input
						id="page-publish-at"
						type="datetime-local"
						bind:value={publishAtLocal}
					/>
				</FormField>
			</div>
			<FormField label="Превью (картинка)" id="page-image">
				<ImageField id="page-image" name="image" bind:value={image} {uploadPath} alt="Превью статьи" />
			</FormField>
			<FormField label="Стек" id="page-tags">
				<TagSelect
					id="page-tags"
					name="tags"
					options={stackOptions}
					bind:values={tags}
					placeholder={stackOptions.length ? 'Выберите технологии' : 'Список стека пуст'}
				/>
			</FormField>
			<p class="page-hint">
				Страница появится на сайте по адресу <code>/{'{lang}'}/articles/{seoSlug || '…'}</code>
				сразу после сохранения (если включено «Опубликована»).
				Превью показывается фоном в карточке списка статей (как в портфолио).
				Дата отложенной публикации необязательна: заполните её, чтобы отложить показ на сайте.
			</p>
			<label class="check">
				<input type="checkbox" name="published" bind:checked={published} />
				Опубликована
			</label>
		</div>
	</Card>

	<Card padding="sm">
		<div class="fields">
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

			{#if activeTab === 'content'}
				<TranslationsEditor {languages} fields={translationFields} bind:translations idPrefix="page" {uploadPath} />
			{:else if seoPath}
				<FormField label="Путь страницы" id="page-seo-path">
					<p class="seo-path">{seoPathDisplay}</p>
				</FormField>
				<div class="seo-editor">
					<h3 class="subsection-title">Meta-теги по языкам</h3>
					<TranslationsEditor
						{languages}
						fields={seoFields}
						bind:translations={seoTranslations}
						idPrefix="page-seo"
						{uploadPath}
					/>
				</div>
			{:else}
				<p class="page-hint">Укажите slug выше — тогда можно заполнить SEO (путь <code>/{'{lang}'}/articles/…</code>).</p>
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
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		gap: 1rem;
		align-items: start;
	}
	@media (max-width: 640px) {
		.grid-2 {
			grid-template-columns: 1fr;
		}
	}
	.page-hint {
		margin: 0;
		font-size: 0.8125rem;
		line-height: 1.5;
		color: #71717a;
	}
	.page-hint code {
		font-size: 0.75rem;
		padding: 0.125rem 0.375rem;
		background: #f4f4f5;
		border-radius: 4px;
	}
	.check {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #374151;
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
