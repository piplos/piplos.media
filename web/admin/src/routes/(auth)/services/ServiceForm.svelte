<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import TagSelect from '$lib/components/TagSelect.svelte';
	import TranslationsEditor from '$lib/components/TranslationsEditor.svelte';
	import { DEFAULT_SERVICE_ICON, SERVICE_ICON_OPTIONS } from '$lib/service-icons';
	import type { Language, Service, StackItem, Translations } from '$lib/types';

	interface Props {
		service?: Partial<Service>;
		languages: Language[];
		stack: StackItem[];
		submitLabel: string;
	}
	let { service = {}, languages, stack, submitLabel }: Props = $props();

	let submitting = $state(false);
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(service) as Partial<Service>;
	const isEdit = Boolean(initial.id);
	let slug = $state(initial.slug ?? '');
	let icon = $state(initial.icon || DEFAULT_SERVICE_ICON);
	let tags = $state(initial.tags ?? []);
	let sortOrder = $state(String(initial.sort_order ?? 0));
	let published = $state(initial.published ?? true);
	let translations = $state<Translations>((initial.translations ?? {}) as Translations);

	const translationFields = [
		{ key: 'title', label: 'Название' },
		{ key: 'description', label: 'Краткое описание', type: 'textarea' as const, rows: 3 },
		{ key: 'body', label: 'Подробное описание страницы (Markdown)', type: 'markdown' as const }
	];

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

	<Card padding="sm">
		<div class="fields">
			<div class="grid-2">
				<FormField label="Slug" id="svc-slug">
					<Input id="svc-slug" name="slug" bind:value={slug} placeholder="web" required />
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
			<h2 class="section-title">Контент по языкам</h2>
			<p class="section-hint">
				Краткое описание — для карточек и шапки страницы. Подробное описание — Markdown с
				картинками для отдельной страницы услуги на сайте.
			</p>
			<TranslationsEditor {languages} fields={translationFields} bind:translations idPrefix="svc" />
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
