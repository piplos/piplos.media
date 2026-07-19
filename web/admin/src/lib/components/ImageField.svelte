<script lang="ts">
	import toast from 'svelte-french-toast';
	import Button from './Button.svelte';
	import FilePickerDrawer from './FilePickerDrawer.svelte';
	import Input from './Input.svelte';
	import { uploadFile } from '$lib/files';

	interface Props {
		id: string;
		name?: string;
		value?: string;
		placeholder?: string;
		/** Папка архива для загрузки и стартовая папка пикера (например, projects/site-dev). */
		uploadPath?: string;
		pickerTitle?: string;
		alt?: string;
	}
	let {
		id,
		name,
		value = $bindable(''),
		placeholder = '/uploads/… или https://…',
		uploadPath = '',
		pickerTitle = 'Выбор картинки из архива',
		alt = 'Превью'
	}: Props = $props();

	let fileInput = $state<HTMLInputElement | null>(null);
	let uploading = $state(false);
	let pickerOpen = $state(false);
	let dragDepth = $state(0);

	async function upload(file: File) {
		uploading = true;
		try {
			const data = await uploadFile(file, uploadPath);
			value = data.url;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Не удалось загрузить изображение');
		} finally {
			uploading = false;
		}
	}

	function onFileChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (file) void upload(file);
	}

	function imageFromDataTransfer(dt: DataTransfer | null): File | null {
		if (!dt) return null;
		for (const item of dt.items) {
			if (item.kind === 'file' && item.type.startsWith('image/')) {
				return item.getAsFile();
			}
		}
		return null;
	}

	function isFileDrag(e: DragEvent): boolean {
		return !!e.dataTransfer && [...e.dataTransfer.types].includes('Files');
	}

	function onDragEnter(e: DragEvent) {
		if (!isFileDrag(e)) return;
		e.preventDefault();
		dragDepth += 1;
	}

	function onDragOver(e: DragEvent) {
		if (!isFileDrag(e)) return;
		e.preventDefault();
		if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
	}

	function onDragLeave(e: DragEvent) {
		if (!isFileDrag(e)) return;
		dragDepth = Math.max(0, dragDepth - 1);
	}

	function onDrop(e: DragEvent) {
		if (!isFileDrag(e)) return;
		e.preventDefault();
		dragDepth = 0;
		const file = imageFromDataTransfer(e.dataTransfer);
		if (file) void upload(file);
		else toast.error('Перетащите файл изображения');
	}
</script>

<div
	class="image-field"
	class:image-field--drag={dragDepth > 0}
	role="group"
	aria-label="Картинка: поле и загрузка"
	ondragenter={onDragEnter}
	ondragover={onDragOver}
	ondragleave={onDragLeave}
	ondrop={onDrop}
>
	<div class="image-controls">
		<Input {id} {name} bind:value {placeholder} />
		<div class="image-buttons">
			<Button variant="secondary" loading={uploading} onclick={() => fileInput?.click()}>
				Загрузить
			</Button>
			<Button variant="secondary" onclick={() => (pickerOpen = true)}>Из архива</Button>
			{#if value}
				<Button variant="ghost" onclick={() => (value = '')}>Убрать</Button>
			{/if}
		</div>
		<p class="image-hint">
			Можно перетащить картинку прямо на этот блок{uploadPath ? ` — она попадёт в папку «${uploadPath}»` : ''}.
		</p>
	</div>
	{#if value}
		<a class="image-thumb" href={value} target="_blank" rel="noreferrer" title="Открыть в новой вкладке">
			<img src={value} {alt} />
		</a>
	{/if}
</div>
<input type="file" accept="image/*" bind:this={fileInput} onchange={onFileChange} hidden />

<FilePickerDrawer
	bind:open={pickerOpen}
	title={pickerTitle}
	initialPath={uploadPath}
	onselect={(file) => (value = file.url)}
/>

<style>
	.image-field {
		display: flex;
		gap: 1rem;
		align-items: flex-start;
		padding: 0.375rem;
		margin: -0.375rem;
		border-radius: 10px;
		transition: box-shadow 0.15s;
	}
	.image-field--drag {
		box-shadow: inset 0 0 0 2px #2563eb;
	}
	.image-controls {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.image-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.image-hint {
		margin: 0;
		font-size: 0.75rem;
		color: #a1a1aa;
	}
	.image-thumb {
		flex-shrink: 0;
		display: block;
		width: 10rem;
		height: 6.25rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		overflow: hidden;
		background: #f4f4f5;
	}
	.image-thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	@media (max-width: 640px) {
		.image-field {
			flex-direction: column;
		}
	}
</style>
