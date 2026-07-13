<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
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
		onSaved?: () => void;
	}
	let { service = {}, languages, stack, onSaved }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; родитель перемонтирует форму через {#key}.
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
		{ key: 'description', label: 'Описание', type: 'textarea' as const }
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
	method="POST"
	action="?/save"
	class="drawer-form"
	use:enhance={() => {
		submitting = true;
		return async ({ result }) => {
			submitting = false;
			if (result.type === 'success') {
				toast.success(isEdit ? 'Услуга обновлена' : 'Услуга создана');
				await invalidateAll();
				onSaved?.();
			} else if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			}
		};
	}}
>
	<input type="hidden" name="id" value={initial.id ?? ''} />
	<input type="hidden" name="translations" value={JSON.stringify(translations)} />

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
	</div>
	<label class="check">
		<input type="checkbox" name="published" bind:checked={published} />
		Опубликована
	</label>

	<div class="form-section">
		<h3 class="section-title">Контент по языкам</h3>
		<TranslationsEditor {languages} fields={translationFields} bind:translations idPrefix="svc" />
	</div>

	<div class="form-actions">
		<Button type="submit" loading={submitting} fullWidth>
			{isEdit ? 'Сохранить' : 'Создать услугу'}
		</Button>
	</div>
</form>

<style>
	.drawer-form {
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
	.check {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}
	.form-section {
		padding-top: 0.5rem;
		border-top: 1px solid #f4f4f5;
	}
	.section-title {
		margin: 0 0 0.75rem;
		font-size: 0.9375rem;
		font-weight: 600;
		color: #18181b;
	}
	.form-actions {
		padding-top: 0.5rem;
	}
</style>
