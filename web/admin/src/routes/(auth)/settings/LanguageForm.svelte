<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import type { Language } from '$lib/types';

	interface Props {
		language?: Partial<Language>;
		onSaved?: () => void;
	}
	let { language = {}, onSaved }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; родитель перемонтирует форму через {#key}.
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(language) as Partial<Language>;
	const isEdit = Boolean(initial.code);
	let code = $state(initial.code ?? '');
	let name = $state(initial.name ?? '');
	let isDefault = $state(initial.is_default ?? false);
	let enabled = $state(initial.enabled ?? true);
	let sortOrder = $state(String(initial.sort_order ?? 0));
</script>

<form
	method="POST"
	action="?/saveLanguage"
	class="drawer-form"
	use:enhance={() => {
		submitting = true;
		return async ({ result }) => {
			submitting = false;
			if (result.type === 'success') {
				toast.success(isEdit ? 'Язык обновлён' : 'Язык добавлен');
				await invalidateAll();
				onSaved?.();
			} else if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			}
		};
	}}
>
	<FormField label="Код" id="lang-code" hint="ISO: en, ru, de...">
		{#if isEdit}
			<input type="hidden" name="code" value={code} />
		{/if}
		<Input
			id="lang-code"
			name={isEdit ? undefined : 'code'}
			bind:value={code}
			placeholder="de"
			required
			disabled={isEdit}
		/>
	</FormField>
	<FormField label="Название" id="lang-name">
		<Input id="lang-name" name="name" bind:value={name} placeholder="Deutsch" required />
	</FormField>
	<FormField label="Порядок сортировки" id="lang-sort">
		<Input id="lang-sort" name="sort_order" type="number" bind:value={sortOrder} />
	</FormField>
	<div class="checks-row">
		<label class="check">
			<input type="checkbox" name="enabled" bind:checked={enabled} />
			Включён
		</label>
		<label class="check">
			<input type="checkbox" name="is_default" bind:checked={isDefault} />
			Язык по умолчанию
		</label>
	</div>
	<div class="form-actions">
		<Button type="submit" loading={submitting} fullWidth>
			{isEdit ? 'Сохранить' : 'Добавить язык'}
		</Button>
	</div>
</form>

<style>
	.drawer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.checks-row {
		display: flex;
		gap: 1.25rem;
	}
	.check {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}
	.form-actions {
		padding-top: 0.25rem;
	}
</style>
