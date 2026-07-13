<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { Editor } from '@tiptap/core';
	import StarterKit from '@tiptap/starter-kit';
	import Image from '@tiptap/extension-image';
	import Link from '@tiptap/extension-link';
	import toast from 'svelte-french-toast';

	interface Props {
		id?: string;
		value?: string;
		/** Показать переключатель «Редактор / HTML». */
		previewable?: boolean;
	}
	let { id, value = $bindable(''), previewable = false }: Props = $props();

	let root = $state<HTMLDivElement | null>(null);
	let fileInput = $state<HTMLInputElement | null>(null);
	let editor = $state<Editor | null>(null);
	let uploading = $state(false);
	let mode = $state<'edit' | 'html'>('edit');

	function normalizeHtml(html: string): string {
		const trimmed = html.trim();
		if (!trimmed || trimmed === '<p></p>') return '';
		return trimmed;
	}

	async function uploadImage(file: File) {
		if (!browser) return;
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
			editor?.chain().focus().setImage({ src: data.url, alt: file.name }).run();
		} catch {
			toast.error('Сервис загрузки недоступен');
		} finally {
			uploading = false;
		}
	}

	function pickImage() {
		fileInput?.click();
	}

	async function onFileChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (file) await uploadImage(file);
	}

	function setLink() {
		const prev = editor?.getAttributes('link').href as string | undefined;
		const url = window.prompt('URL ссылки', prev ?? 'https://');
		if (url === null) return;
		if (url === '') {
			editor?.chain().focus().extendMarkRange('link').unsetLink().run();
			return;
		}
		editor?.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
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

	$effect(() => {
		if (!editor) return;
		const current = normalizeHtml(editor.getHTML());
		const next = normalizeHtml(value);
		if (current !== next) {
			editor.commands.setContent(next || '<p></p>', { emitUpdate: false });
		}
	});

	onMount(() => {
		if (!browser || !root) return;

		const ed = new Editor({
			element: root,
			extensions: [
				StarterKit.configure({ heading: { levels: [2, 3] } }),
				Link.configure({ openOnClick: false, autolink: true, linkOnPaste: true }),
				Image.configure({ inline: false, allowBase64: false })
			],
			content: value || '<p></p>',
			editorProps: {
				handleDrop: (_view, event) => {
					const file = imageFromDataTransfer(event.dataTransfer);
					if (!file) return false;
					event.preventDefault();
					void uploadImage(file);
					return true;
				},
				handlePaste: (_view, event) => {
					const file = imageFromDataTransfer(event.clipboardData);
					if (!file) return false;
					event.preventDefault();
					void uploadImage(file);
					return true;
				}
			},
			onUpdate: ({ editor: e }) => {
				value = normalizeHtml(e.getHTML());
			}
		});

		editor = ed;
		return () => {
			ed.destroy();
			editor = null;
		};
	});
</script>

<div class="rte" {id}>
	<div class="rte-toolbar" role="toolbar" aria-label="Форматирование">
		<button
			type="button"
			class="rte-btn"
			class:rte-btn--active={editor?.isActive('bold')}
			title="Жирный"
			disabled={!editor || mode === 'html'}
			onclick={() => editor?.chain().focus().toggleBold().run()}
		>
			B
		</button>
		<button
			type="button"
			class="rte-btn"
			class:rte-btn--active={editor?.isActive('italic')}
			title="Курсив"
			disabled={!editor || mode === 'html'}
			onclick={() => editor?.chain().focus().toggleItalic().run()}
		>
			I
		</button>
		<button
			type="button"
			class="rte-btn"
			class:rte-btn--active={editor?.isActive('bulletList')}
			title="Маркированный список"
			disabled={!editor || mode === 'html'}
			onclick={() => editor?.chain().focus().toggleBulletList().run()}
		>
			•
		</button>
		<button
			type="button"
			class="rte-btn"
			class:rte-btn--active={editor?.isActive('orderedList')}
			title="Нумерованный список"
			disabled={!editor || mode === 'html'}
			onclick={() => editor?.chain().focus().toggleOrderedList().run()}
		>
			1.
		</button>
		<button type="button" class="rte-btn" title="Ссылка" disabled={!editor || mode === 'html'} onclick={setLink}>
			🔗
		</button>
		<button
			type="button"
			class="rte-btn"
			title="Вставить изображение"
			disabled={!editor || uploading || mode === 'html'}
			onclick={pickImage}
		>
			{#if uploading}…{:else}🖼{/if}
		</button>
		{#if previewable}
			<span class="rte-toolbar-sep" aria-hidden="true"></span>
			<button
				type="button"
				class="rte-btn"
				class:rte-btn--active={mode === 'edit'}
				title="Редактор"
				disabled={!editor}
				onclick={() => (mode = 'edit')}
			>
				Редактор
			</button>
			<button
				type="button"
				class="rte-btn"
				class:rte-btn--active={mode === 'html'}
				title="HTML-код"
				onclick={() => (mode = 'html')}
			>
				HTML
			</button>
		{/if}
	</div>

	{#if previewable && mode === 'html'}
		<textarea
			class="rte-html"
			bind:value
			spellcheck="false"
			aria-label="HTML-код"
			placeholder="<p>Текст решения</p>"
		></textarea>
	{:else if browser}
		<div class="rte-surface" bind:this={root}></div>
	{:else}
		<textarea
			class="rte-html"
			bind:value
			spellcheck="false"
			aria-label="HTML-код"
			placeholder="<p>Текст</p>"
		></textarea>
	{/if}

	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/webp,image/gif"
		class="rte-file"
		onchange={onFileChange}
	/>
</div>

<style>
	.rte {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.rte-toolbar {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		padding: 0.25rem;
		background: #f4f4f5;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.rte-toolbar-sep {
		width: 1px;
		align-self: stretch;
		margin: 0.25rem 0.125rem;
		background: #d4d4d8;
	}
	.rte-btn {
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
	.rte-btn:hover:not(:disabled) {
		color: #18181b;
		background: #fafafa;
	}
	.rte-btn--active {
		color: #18181b;
		background: #e4e4e7;
	}
	.rte-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.rte-surface {
		min-height: 12rem;
		padding: 0.75rem;
		font-size: 0.875rem;
		line-height: 1.6;
		color: #18181b;
		background: #fff;
		border: 1px solid #d1d5db;
		border-radius: 8px;
	}
	.rte-surface--hidden {
		display: none;
	}
	.rte-html {
		min-height: 12rem;
		width: 100%;
		padding: 0.75rem;
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.8125rem;
		line-height: 1.55;
		color: #18181b;
		background: #fafafa;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		resize: vertical;
		box-sizing: border-box;
		tab-size: 2;
	}
	.rte-html:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 2px rgba(17, 17, 17, 0.08);
	}
	.rte-surface:focus-within {
		border-color: #111;
		box-shadow: 0 0 0 2px rgba(17, 17, 17, 0.08);
	}
	.rte-surface :global(.ProseMirror) {
		outline: none;
		min-height: 10rem;
	}
	.rte-surface :global(.ProseMirror p) {
		margin: 0 0 0.75rem;
	}
	.rte-surface :global(.ProseMirror p:last-child) {
		margin-bottom: 0;
	}
	.rte-surface :global(.ProseMirror img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 0.75rem 0;
		border-radius: 8px;
	}
	.rte-surface :global(.ProseMirror ul),
	.rte-surface :global(.ProseMirror ol) {
		margin: 0 0 0.75rem 1.25rem;
		padding: 0;
	}
	.rte-surface :global(.ProseMirror a) {
		color: #2563eb;
		text-decoration: underline;
	}
	.rte-file {
		display: none;
	}
</style>
