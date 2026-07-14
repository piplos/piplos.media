<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import type { StackItem } from '$lib/types';

	interface Props {
		item?: Partial<StackItem>;
		onSaved?: () => void;
	}
	let { item = {}, onSaved }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; родитель перемонтирует форму через {#key}.
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(item) as Partial<StackItem>;
	const isEdit = Boolean(initial.id);
	let slug = $state(initial.slug ?? '');
	let label = $state(initial.label ?? '');
	let published = $state(initial.published ?? true);
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
				toast.success(isEdit ? 'Технология обновлена' : 'Технология добавлена');
				await invalidateAll();
				onSaved?.();
			} else if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			}
		};
	}}
>
	<input type="hidden" name="id" value={initial.id ?? ''} />
	<FormField label="Название" id="stack-label">
		<Input id="stack-label" name="label" bind:value={label} placeholder="PostgreSQL" required />
	</FormField>
	<FormField label="Slug" id="stack-slug">
		<Input id="stack-slug" name="slug" bind:value={slug} placeholder="postgresql" required />
	</FormField>
	<label class="check">
		<input type="checkbox" name="published" bind:checked={published} />
		Виден на сайте
	</label>
	<div class="form-actions">
		<Button type="submit" loading={submitting} fullWidth>
			{isEdit ? 'Сохранить' : 'Добавить'}
		</Button>
	</div>
</form>

<style>
	.drawer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.form-actions {
		padding-top: 0.5rem;
	}
</style>
