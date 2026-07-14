<script lang="ts">
	import { browser } from '$app/environment';
	import { tick } from 'svelte';
	import { marked } from 'marked';
	import toast from 'svelte-french-toast';
	import FilePickerDrawer from './FilePickerDrawer.svelte';
	import { sanitizeCaseHtml } from '$lib/sanitize-html';

	interface Props {
		id: string;
		value?: string;
		rows?: number;
		placeholder?: string;
	}
	let { id, value = $bindable(''), rows = 12, placeholder }: Props = $props();

	const source = $derived(value ?? '');

	function setSource(next: string) {
		value = next;
	}

	let textarea = $state<HTMLTextAreaElement | null>(null);
	let fileInput = $state<HTMLInputElement | null>(null);
	let pickerOpen = $state(false);
	let uploading = $state(false);
	let mode = $state<'edit' | 'preview'>('edit');

	const previewHtml = $derived.by(() => {
		if (!browser || mode !== 'preview' || !source.trim()) return '';
		const html = marked.parse(source, { async: false, gfm: true }) as string;
		return sanitizeCaseHtml(html);
	});

	async function replaceRange(
		start: number,
		end: number,
		text: string,
		selectFrom: number,
		selectTo: number
	) {
		const next = source.slice(0, start) + text + source.slice(end);
		setSource(next);
		await tick();
		textarea?.focus();
		textarea?.setSelectionRange(selectFrom, selectTo);
	}

	function selection(): { start: number; end: number } {
		return textarea
			? { start: textarea.selectionStart, end: textarea.selectionEnd }
			: { start: source.length, end: source.length };
	}

	function wrapSelection(prefix: string, suffix: string, fallback: string) {
		const { start, end } = selection();
		const selected = source.slice(start, end) || fallback;
		void replaceRange(
			start,
			end,
			prefix + selected + suffix,
			start + prefix.length,
			start + prefix.length + selected.length
		);
	}

	/** Добавляет префикс каждой выделенной строке (заголовки, списки, цитаты). */
	function prefixLines(prefix: string | ((index: number) => string)) {
		const { start, end } = selection();
		const lineStart = source.lastIndexOf('\n', start - 1) + 1;
		const nextBreak = source.indexOf('\n', end);
		const lineEnd = nextBreak < 0 ? source.length : nextBreak;
		const block = source.slice(lineStart, lineEnd) || (typeof prefix === 'string' ? 'Текст' : 'Пункт');
		const prefixed = block
			.split('\n')
			.map((line, i) => (typeof prefix === 'function' ? prefix(i) : prefix) + line)
			.join('\n');
		void replaceRange(lineStart, lineEnd, prefixed, lineStart, lineStart + prefixed.length);
	}

	function insertLink() {
		const url = window.prompt('URL ссылки', 'https://');
		if (!url) return;
		const { start, end } = selection();
		const text = source.slice(start, end) || 'ссылка';
		const md = `[${text}](${url})`;
		void replaceRange(start, end, md, start + 1, start + 1 + text.length);
	}

	function insertImage(url: string, alt: string) {
		const { start, end } = selection();
		const before = start > 0 && source[start - 1] !== '\n' ? '\n\n' : '';
		const md = `${before}![${alt}](${url})\n`;
		const caret = start + md.length;
		void replaceRange(start, end, md, caret, caret);
	}

	async function uploadImage(file: File) {
		uploading = true;
		try {
			const fd = new FormData();
			fd.append('file', file);
			const res = await fetch('/api/upload', { method: 'POST', body: fd });
			const data = (await res.json().catch(() => ({}))) as { url?: string; message?: string };
			if (!res.ok || !data.url) {
				toast.error(data.message ?? 'Не удалось загрузить изображение');
				return;
			}
			insertImage(data.url, file.name.replace(/\.[^.]+$/, ''));
		} catch {
			toast.error('Сервис загрузки недоступен');
		} finally {
			uploading = false;
		}
	}

	async function onFileChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (file) await uploadImage(file);
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

	function onPaste(e: ClipboardEvent) {
		const file = imageFromDataTransfer(e.clipboardData);
		if (!file) return;
		e.preventDefault();
		void uploadImage(file);
	}

	function onDrop(e: DragEvent) {
		const file = imageFromDataTransfer(e.dataTransfer);
		if (!file) return;
		e.preventDefault();
		void uploadImage(file);
	}
</script>

<div class="mde" {id}>
	<div class="mde-toolbar" role="toolbar" aria-label="Форматирование Markdown">
		<button
			type="button"
			class="mde-btn"
			title="Жирный"
			disabled={mode === 'preview'}
			onclick={() => wrapSelection('**', '**', 'жирный')}
		>
			B
		</button>
		<button
			type="button"
			class="mde-btn mde-btn--italic"
			title="Курсив"
			disabled={mode === 'preview'}
			onclick={() => wrapSelection('*', '*', 'курсив')}
		>
			I
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Заголовок H2"
			disabled={mode === 'preview'}
			onclick={() => prefixLines('## ')}
		>
			H2
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Заголовок H3"
			disabled={mode === 'preview'}
			onclick={() => prefixLines('### ')}
		>
			H3
		</button>
		<span class="mde-toolbar-sep" aria-hidden="true"></span>
		<button
			type="button"
			class="mde-btn"
			title="Маркированный список"
			disabled={mode === 'preview'}
			onclick={() => prefixLines('- ')}
		>
			•
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Нумерованный список"
			disabled={mode === 'preview'}
			onclick={() => prefixLines((i) => `${i + 1}. `)}
		>
			1.
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Цитата"
			disabled={mode === 'preview'}
			onclick={() => prefixLines('> ')}
		>
			❝
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Код"
			disabled={mode === 'preview'}
			onclick={() => wrapSelection('`', '`', 'код')}
		>
			&lt;/&gt;
		</button>
		<span class="mde-toolbar-sep" aria-hidden="true"></span>
		<button type="button" class="mde-btn" title="Ссылка" disabled={mode === 'preview'} onclick={insertLink}>
			🔗
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Загрузить изображение"
			disabled={uploading || mode === 'preview'}
			onclick={() => fileInput?.click()}
		>
			{#if uploading}…{:else}🖼{/if}
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Изображение из архива"
			disabled={mode === 'preview'}
			onclick={() => (pickerOpen = true)}
		>
			📁
		</button>
		<span class="mde-toolbar-sep mde-toolbar-sep--push" aria-hidden="true"></span>
		<button
			type="button"
			class="mde-btn"
			class:mde-btn--active={mode === 'edit'}
			title="Редактор"
			onclick={() => (mode = 'edit')}
		>
			Редактор
		</button>
		<button
			type="button"
			class="mde-btn"
			class:mde-btn--active={mode === 'preview'}
			title="Просмотр"
			onclick={() => (mode = 'preview')}
		>
			Просмотр
		</button>
	</div>

	{#if mode === 'preview'}
		<div class="mde-preview">
			{#if previewHtml}
				<!-- eslint-disable-next-line svelte/no-at-html-tags — HTML прошёл через DOMPurify -->
				{@html previewHtml}
			{:else}
				<p class="mde-preview-empty">Нет содержимого</p>
			{/if}
		</div>
	{:else}
		<textarea
			bind:this={textarea}
			value={source}
			oninput={(e) => setSource(e.currentTarget.value)}
			{rows}
			class="mde-source"
			spellcheck="false"
			placeholder={placeholder ?? 'Markdown: **жирный**, ## заголовок, - список, ![alt](url картинки)'}
			aria-label="Markdown"
			onpaste={onPaste}
			ondrop={onDrop}
		></textarea>
	{/if}

	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/webp,image/gif"
		class="mde-file"
		onchange={onFileChange}
	/>
</div>

<FilePickerDrawer
	bind:open={pickerOpen}
	title="Выбор изображения"
	mode="file"
	imagesOnly
	onselect={(file) => {
		pickerOpen = false;
		insertImage(file.url, file.path.split('/').pop()?.replace(/\.[^.]+$/, '') ?? '');
	}}
/>

<style>
	.mde {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.mde-toolbar {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		padding: 0.25rem;
		background: #f4f4f5;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.mde-toolbar-sep {
		width: 1px;
		align-self: stretch;
		margin: 0.25rem 0.125rem;
		background: #d4d4d8;
	}
	.mde-toolbar-sep--push {
		margin-left: auto;
		background: transparent;
	}
	.mde-btn {
		min-width: 2rem;
		height: 2rem;
		padding: 0 0.5rem;
		font-size: 0.8125rem;
		font-weight: 600;
		color: #52525b;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	.mde-btn--italic {
		font-style: italic;
	}
	.mde-btn:hover:not(:disabled) {
		color: #18181b;
		background: #fafafa;
	}
	.mde-btn--active {
		color: #18181b;
		background: #e4e4e7;
	}
	.mde-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.mde-source {
		width: 100%;
		min-height: 12rem;
		padding: 0.75rem;
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.8125rem;
		line-height: 1.6;
		color: #18181b;
		background: #fff;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		resize: vertical;
		box-sizing: border-box;
		tab-size: 2;
	}
	.mde-source::placeholder {
		color: #9ca3af;
	}
	.mde-source:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 2px rgba(17, 17, 17, 0.08);
	}
	.mde-preview {
		min-height: 12rem;
		padding: 0.75rem;
		font-size: 0.875rem;
		line-height: 1.6;
		color: #18181b;
		background: #fafafa;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		box-sizing: border-box;
		overflow-wrap: break-word;
	}
	.mde-preview-empty {
		margin: 0;
		color: #a1a1aa;
	}
	.mde-preview :global(p) {
		margin: 0 0 0.75rem;
	}
	.mde-preview :global(p:last-child) {
		margin-bottom: 0;
	}
	.mde-preview :global(h2),
	.mde-preview :global(h3),
	.mde-preview :global(h4) {
		margin: 1rem 0 0.5rem;
		color: #18181b;
	}
	.mde-preview :global(ul),
	.mde-preview :global(ol) {
		margin: 0 0 0.75rem 1.25rem;
		padding: 0;
	}
	.mde-preview :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 0.75rem 0;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.mde-preview :global(a) {
		color: #2563eb;
		text-decoration: underline;
	}
	.mde-preview :global(blockquote) {
		margin: 0 0 0.75rem;
		padding: 0.25rem 0.75rem;
		color: #52525b;
		border-left: 3px solid #d4d4d8;
	}
	.mde-preview :global(code) {
		padding: 0.125rem 0.25rem;
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.8125em;
		background: #f4f4f5;
		border-radius: 4px;
	}
	.mde-preview :global(pre) {
		padding: 0.75rem;
		background: #f4f4f5;
		border-radius: 8px;
		overflow-x: auto;
	}
	.mde-preview :global(pre code) {
		padding: 0;
		background: none;
	}
	.mde-file {
		display: none;
	}
</style>
