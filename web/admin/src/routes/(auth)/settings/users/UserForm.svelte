<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import type { AdminUser } from '$lib/types';

	interface Props {
		user?: Partial<AdminUser>;
		onSaved?: () => void;
	}
	let { user = {}, onSaved }: Props = $props();

	let submitting = $state(false);
	// Начальные значения формы фиксируются при монтировании; родитель перемонтирует форму через {#key}.
	// svelte-ignore state_referenced_locally
	const initial = $state.snapshot(user) as Partial<AdminUser>;
	const isEdit = Boolean(initial.id);
	let email = $state(initial.email ?? '');
	let password = $state('');
	let fullName = $state(initial.full_name ?? '');
	let role = $state(initial.role ?? 'manager');
	let isActive = $state(initial.is_active ?? true);
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
				toast.success(isEdit ? 'Пользователь обновлён' : 'Пользователь создан');
				await invalidateAll();
				onSaved?.();
			} else if (result.type === 'failure') {
				toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
			}
		};
	}}
>
	<input type="hidden" name="id" value={initial.id ?? ''} />
	<FormField label="Email" id="user-email">
		{#if isEdit}
			<input type="hidden" name="email" value={email} />
		{/if}
		<Input
			id="user-email"
			name={isEdit ? undefined : 'email'}
			type="email"
			bind:value={email}
			placeholder="user@piplos.media"
			required
			disabled={isEdit}
		/>
	</FormField>
	<FormField label="Имя" id="user-name">
		<Input id="user-name" name="full_name" bind:value={fullName} placeholder="Иван Иванов" />
	</FormField>
	<FormField label={isEdit ? 'Новый пароль (пусто — не менять)' : 'Пароль'} id="user-password">
		<Input
			id="user-password"
			name="password"
			type="password"
			bind:value={password}
			placeholder="минимум 8 символов"
			autocomplete="new-password"
		/>
	</FormField>
	<FormField label="Роль" id="user-role">
		<Select
			id="user-role"
			name="role"
			bind:value={role}
			options={[
				{ value: 'manager', label: 'Менеджер' },
				{ value: 'admin', label: 'Администратор' }
			]}
		/>
	</FormField>
	<label class="check">
		<input type="checkbox" name="is_active" bind:checked={isActive} />
		Активен
	</label>
	<div class="form-actions">
		<Button type="submit" loading={submitting} fullWidth>
			{isEdit ? 'Сохранить' : 'Создать пользователя'}
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
		padding-top: 0.25rem;
	}
</style>
