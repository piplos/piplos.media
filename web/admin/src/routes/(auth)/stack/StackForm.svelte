<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import FilePickerDrawer from '$lib/components/FilePickerDrawer.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import { uploadFile } from '$lib/files';
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
	let icon = $state(initial.icon ?? '');
	let iconAlt = $state(initial.icon_alt ?? '');
	let published = $state(initial.published ?? true);
	let iconInput = $state<HTMLInputElement | null>(null);
	let iconAltInput = $state<HTMLInputElement | null>(null);
	let uploadingIcon = $state(false);
	let uploadingIconAlt = $state(false);
	let iconPickerOpen = $state(false);
	let iconAltPickerOpen = $state(false);
	let iconPickerTarget = $state<'icon' | 'icon_alt'>('icon');

	async function onIconFileChange(e: Event, target: 'icon' | 'icon_alt') {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;

		const setUploading = target === 'icon' ? (v: boolean) => (uploadingIcon = v) : (v: boolean) => (uploadingIconAlt = v);
		const setValue = target === 'icon' ? (v: string) => (icon = v) : (v: string) => (iconAlt = v);

		setUploading(true);
		try {
			const data = await uploadFile(file, 'stack');
			setValue(data.path || data.url);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Не удалось загрузить иконку');
		} finally {
			setUploading(false);
		}
	}

	function openIconPicker(target: 'icon' | 'icon_alt') {
		iconPickerTarget = target;
		if (target === 'icon') iconPickerOpen = true;
		else iconAltPickerOpen = true;
	}

	function onIconPicked(file: { path: string; url: string }) {
		if (iconPickerTarget === 'icon') icon = file.path || file.url;
		else iconAlt = file.path || file.url;
	}
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

	<FormField label="Иконка" id="stack-icon">
		<div class="icon-field">
			<div class="icon-controls">
				<Input id="stack-icon" name="icon" bind:value={icon} placeholder="/uploads/stack/postgresql.svg" />
				<div class="icon-buttons">
					<Button variant="secondary" loading={uploadingIcon} onclick={() => iconInput?.click()}>
						Загрузить
					</Button>
					<Button variant="secondary" onclick={() => openIconPicker('icon')}>Из архива</Button>
					{#if icon}
						<Button variant="ghost" onclick={() => (icon = '')}>Убрать</Button>
					{/if}
				</div>
			</div>
			{#if icon}
				<div class="icon-thumb">
					<img src={icon} alt="" />
				</div>
			{/if}
		</div>
		<input
			bind:this={iconInput}
			type="file"
			accept="image/*,.svg"
			class="sr-only"
			onchange={(e) => onIconFileChange(e, 'icon')}
		/>
	</FormField>

	<FormField label="Иконка для тёмной темы" id="stack-icon-alt" hint="Необязательно. Например, светлый логотип для Next.js.">
		<div class="icon-field">
			<div class="icon-controls">
				<Input
					id="stack-icon-alt"
					name="icon_alt"
					bind:value={iconAlt}
					placeholder="/uploads/stack/nextjs-light.svg"
				/>
				<div class="icon-buttons">
					<Button variant="secondary" loading={uploadingIconAlt} onclick={() => iconAltInput?.click()}>
						Загрузить
					</Button>
					<Button variant="secondary" onclick={() => openIconPicker('icon_alt')}>Из архива</Button>
					{#if iconAlt}
						<Button variant="ghost" onclick={() => (iconAlt = '')}>Убрать</Button>
					{/if}
				</div>
			</div>
			{#if iconAlt}
				<div class="icon-thumb icon-thumb--dark">
					<img src={iconAlt} alt="" />
				</div>
			{/if}
		</div>
		<input
			bind:this={iconAltInput}
			type="file"
			accept="image/*,.svg"
			class="sr-only"
			onchange={(e) => onIconFileChange(e, 'icon_alt')}
		/>
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

<FilePickerDrawer
	bind:open={iconPickerOpen}
	title="Иконка из архива"
	initialPath="stack"
	onselect={onIconPicked}
/>
<FilePickerDrawer
	bind:open={iconAltPickerOpen}
	title="Иконка для тёмной темы"
	initialPath="stack"
	onselect={onIconPicked}
/>

<style>
	.drawer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.form-actions {
		padding-top: 0.5rem;
	}
	.icon-field {
		display: flex;
		gap: 1rem;
		align-items: flex-start;
	}
	.icon-controls {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.icon-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.icon-thumb {
		flex-shrink: 0;
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border: 1px solid var(--c-border);
		border-radius: 8px;
		background: #fff;
	}
	.icon-thumb--dark {
		background: #1a1a1e;
	}
	.icon-thumb img {
		display: block;
		width: 32px;
		height: 32px;
		object-fit: contain;
	}
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}
</style>
