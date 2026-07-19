<script lang="ts">
	import { browser } from '$app/environment';
	import { tick } from 'svelte';
	import { marked } from 'marked';
	import toast from 'svelte-french-toast';
	import Button from './Button.svelte';
	import FilePickerDrawer from './FilePickerDrawer.svelte';
	import Select from './Select.svelte';
	import {
		buildEmbedToken,
		embedTokensToPreviewChips,
		EMBED_KIND_OPTIONS,
		EMBED_LAYOUT_OPTIONS,
		EMBED_SELECTION_OPTIONS,
		type EmbedKind,
		type EmbedLayout,
		type EmbedSelection
	} from '$lib/embeds';

	interface Props {
		id: string;
		value?: string;
		rows?: number;
		placeholder?: string;
		/** Папка архива для загрузок и стартовая папка пикера (например, projects/site-dev). */
		uploadPath?: string;
		/** Показывать кнопку вставки переменных ({{projects …}}/{{services …}}). */
		embeds?: boolean;
	}
	let { id, value = $bindable(''), rows = 12, placeholder, uploadPath = '', embeds = true }: Props = $props();

	const source = $derived(value ?? '');

	function setSource(next: string) {
		value = next;
	}

	let textarea = $state<HTMLTextAreaElement | null>(null);
	let pickerOpen = $state(false);
	let uploading = $state(false);
	let mode = $state<'edit' | 'preview'>('edit');
	let previewHtml = $state('');

	$effect(() => {
		if (!browser || mode !== 'preview' || !source.trim()) {
			previewHtml = '';
			return;
		}

		const html = marked.parse(source, { async: false, gfm: true }) as string;
		let cancelled = false;

		import('$lib/sanitize-html').then(async ({ sanitizeCaseHtmlAsync }) => {
			if (!cancelled) previewHtml = embedTokensToPreviewChips(await sanitizeCaseHtmlAsync(html));
		});

		return () => {
			cancelled = true;
		};
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

	// ---------- вставка ссылки ----------

	type LinkStyle = 'link' | 'btn-primary' | 'btn-secondary';

	const LINK_STYLE_OPTIONS = [
		{ value: 'link', label: 'Обычная ссылка' },
		{ value: 'btn-primary', label: 'Кнопка (акцентная)' },
		{ value: 'btn-secondary', label: 'Кнопка (контурная)' }
	];
	const LINK_TARGET_OPTIONS = [
		{ value: 'same', label: 'В текущей вкладке' },
		{ value: 'blank', label: 'В новой вкладке' }
	];

	let linkPanelOpen = $state(false);
	let linkText = $state('');
	let linkUrl = $state('');
	let linkStyle = $state<LinkStyle>('link');
	let linkTarget = $state<'same' | 'blank'>('same');
	let linkNofollow = $state(false);

	function toggleLinkPanel() {
		if (!linkPanelOpen) {
			const { start, end } = selection();
			linkText = source.slice(start, end).trim();
			embedPanelOpen = false;
		}
		linkPanelOpen = !linkPanelOpen;
	}

	function escapeAttr(value: string): string {
		return value.replaceAll('&', '&amp;').replaceAll('"', '&quot;').replaceAll('<', '&lt;');
	}

	function escapeText(value: string): string {
		return value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');
	}

	function buildLink(): string {
		const url = linkUrl.trim();
		const text = linkText.trim() || url;
		const plain = linkStyle === 'link' && linkTarget === 'same' && !linkNofollow;
		if (plain) {
			const safeUrl = /[()\s]/.test(url) ? `<${url}>` : url;
			return `[${text}](${safeUrl})`;
		}
		const attrs = [`href="${escapeAttr(url)}"`];
		const rel: string[] = [];
		if (linkTarget === 'blank') {
			attrs.push('target="_blank"');
			rel.push('noopener', 'noreferrer');
		}
		if (linkNofollow) rel.push('nofollow');
		if (rel.length) attrs.push(`rel="${rel.join(' ')}"`);
		if (linkStyle !== 'link') attrs.push(`class="${linkStyle}"`);
		return `<a ${attrs.join(' ')}>${escapeText(text)}</a>`;
	}

	function insertLink() {
		if (!linkUrl.trim()) return;
		const md = buildLink();
		const { start, end } = selection();
		const caret = start + md.length;
		void replaceRange(start, end, md, caret, caret);
		linkPanelOpen = false;
		linkText = '';
		linkUrl = '';
		linkStyle = 'link';
		linkTarget = 'same';
		linkNofollow = false;
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
			if (uploadPath) fd.append('path', uploadPath);
			fd.append('name', file.name);
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

	// ---------- переменные (embed-токены) ----------

	let embedPanelOpen = $state(false);
	let embedKind = $state<EmbedKind>('projects');
	let embedSelection = $state<EmbedSelection>('auto');
	let embedValue = $state('');
	let embedLimit = $state('3');
	let embedLayout = $state<EmbedLayout>('cards');

	const embedSelectionOptions = $derived(EMBED_SELECTION_OPTIONS[embedKind]);
	const embedNeedsValue = $derived(
		embedSelection === 'category' || embedSelection === 'tags' || embedSelection === 'slugs'
	);
	const embedValuePlaceholder = $derived(
		embedSelection === 'category'
			? 'web'
			: embedSelection === 'tags'
				? 'Go, Svelte'
				: embedKind === 'projects'
					? 'site-dev, crm-app'
					: 'web, mobile'
	);

	function onEmbedKindChange(next: string) {
		embedKind = next as EmbedKind;
		if (!EMBED_SELECTION_OPTIONS[embedKind].some((o) => o.value === embedSelection)) {
			embedSelection = 'auto';
			embedValue = '';
		}
	}

	function insertEmbed() {
		const token = buildEmbedToken({
			kind: embedKind,
			selection: embedSelection,
			value: embedValue,
			limit: Number(embedLimit),
			layout: embedLayout
		});
		const { start, end } = selection();
		const before = start > 0 && source[start - 1] !== '\n' ? '\n\n' : '';
		const md = `${before}${token}\n`;
		const caret = start + md.length;
		void replaceRange(start, end, md, caret, caret);
		embedPanelOpen = false;
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
		<button
			type="button"
			class="mde-btn"
			class:mde-btn--active={linkPanelOpen}
			title="Ссылка — текст, стиль, способ открытия"
			disabled={mode === 'preview'}
			onclick={toggleLinkPanel}
		>
			🔗
		</button>
		<button
			type="button"
			class="mde-btn"
			title="Изображение — выбрать из архива или загрузить"
			disabled={uploading || mode === 'preview'}
			onclick={() => (pickerOpen = true)}
		>
			{#if uploading}…{:else}🖼{/if}
		</button>
		{#if embeds}
			<button
				type="button"
				class="mde-btn"
				class:mde-btn--active={embedPanelOpen}
				title="Вставить блок проектов или услуг"
				disabled={mode === 'preview'}
				onclick={() => {
					linkPanelOpen = false;
					embedPanelOpen = !embedPanelOpen;
				}}
			>
				▦ Блок
			</button>
		{/if}
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

	{#if linkPanelOpen && mode === 'edit'}
		<div class="mde-embed-panel">
			<div class="mde-embed-row">
				<label class="mde-embed-label" for="{id}-link-text">Текст</label>
				<input
					id="{id}-link-text"
					type="text"
					class="mde-embed-input"
					placeholder="Текст ссылки"
					bind:value={linkText}
				/>
				<label class="mde-embed-label" for="{id}-link-url">URL</label>
				<input
					id="{id}-link-url"
					type="text"
					class="mde-embed-input"
					placeholder="https://… или /uploads/…"
					bind:value={linkUrl}
				/>
			</div>
			<div class="mde-embed-row">
				<label class="mde-embed-label" for="{id}-link-style">Стиль</label>
				<Select
					id="{id}-link-style"
					bind:value={() => linkStyle, (v) => (linkStyle = v as LinkStyle)}
					options={LINK_STYLE_OPTIONS}
				/>
				<label class="mde-embed-label" for="{id}-link-target">Открывать</label>
				<Select
					id="{id}-link-target"
					bind:value={() => linkTarget, (v) => (linkTarget = v as 'same' | 'blank')}
					options={LINK_TARGET_OPTIONS}
				/>
				<label class="mde-link-check">
					<input type="checkbox" bind:checked={linkNofollow} />
					nofollow
				</label>
				<Button variant="primary" disabled={!linkUrl.trim()} onclick={insertLink}>Вставить</Button>
			</div>
			<p class="mde-embed-hint">
				Обычная ссылка вставляется как Markdown; кнопки и открытие в новой вкладке — как HTML-тег
				<code>&lt;a&gt;</code> со стилями сайта.
			</p>
		</div>
	{/if}

	{#if embedPanelOpen && mode === 'edit'}
		<div class="mde-embed-panel">
			<div class="mde-embed-row">
				<label class="mde-embed-label" for="{id}-embed-kind">Что показать</label>
				<Select
					id="{id}-embed-kind"
					bind:value={() => embedKind, (v) => onEmbedKindChange(v)}
					options={[...EMBED_KIND_OPTIONS]}
				/>
				<label class="mde-embed-label" for="{id}-embed-selection">Выборка</label>
				<Select
					id="{id}-embed-selection"
					bind:value={() => embedSelection, (v) => (embedSelection = v as EmbedSelection)}
					options={embedSelectionOptions}
				/>
			</div>
			{#if embedNeedsValue}
				<div class="mde-embed-row">
					<label class="mde-embed-label" for="{id}-embed-value">
						{embedSelection === 'category' ? 'Slug услуги' : embedSelection === 'tags' ? 'Теги (через запятую)' : 'Slug-и (через запятую)'}
					</label>
					<input
						id="{id}-embed-value"
						type="text"
						class="mde-embed-input"
						placeholder={embedValuePlaceholder}
						bind:value={embedValue}
					/>
				</div>
			{/if}
			<div class="mde-embed-row">
				<label class="mde-embed-label" for="{id}-embed-limit">Количество</label>
				<input
					id="{id}-embed-limit"
					type="number"
					class="mde-embed-input mde-embed-input--num"
					min="1"
					max="24"
					bind:value={embedLimit}
				/>
				<label class="mde-embed-label" for="{id}-embed-layout">Дизайн</label>
				<Select
					id="{id}-embed-layout"
					bind:value={() => embedLayout, (v) => (embedLayout = v as EmbedLayout)}
					options={[...EMBED_LAYOUT_OPTIONS]}
				/>
				<Button variant="primary" onclick={insertEmbed}>Вставить</Button>
			</div>
			<p class="mde-embed-hint">
				Переменная вставится в текст как <code>{'{{projects limit=3}}'}</code> — на сайте она
				заменится блоком с выбранным дизайном.
			</p>
		</div>
	{/if}

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
</div>

<FilePickerDrawer
	bind:open={pickerOpen}
	title="Выбор изображения"
	mode="file"
	imagesOnly
	initialPath={uploadPath}
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
	.mde-embed-panel {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 0.75rem;
		background: #fafafa;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	.mde-embed-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.mde-embed-row :global(.select) {
		width: auto;
		min-width: 11rem;
	}
	.mde-embed-label {
		font-size: 0.75rem;
		font-weight: 500;
		color: #52525b;
		white-space: nowrap;
	}
	.mde-embed-input {
		flex: 1;
		min-width: 10rem;
		min-height: 2rem;
		padding: 0.25rem 0.625rem;
		font-size: 0.8125rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		box-sizing: border-box;
	}
	.mde-embed-input--num {
		flex: 0 0 5rem;
		min-width: 5rem;
	}
	.mde-embed-input:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	.mde-link-check {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.8125rem;
		color: #52525b;
		white-space: nowrap;
		cursor: pointer;
	}
	.mde-embed-hint {
		margin: 0;
		font-size: 0.75rem;
		color: #a1a1aa;
	}
	.mde-embed-hint code {
		padding: 0.0625rem 0.25rem;
		font-size: 0.6875rem;
		background: #f4f4f5;
		border-radius: 4px;
	}
	.mde-preview :global(.mde-embed-chip) {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.625rem;
		margin: 0.125rem 0;
		font-size: 0.75rem;
		font-weight: 500;
		color: #3730a3;
		background: #eef2ff;
		border: 1px dashed #a5b4fc;
		border-radius: 8px;
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
	.mde-preview :global(a.btn-primary),
	.mde-preview :global(a.btn-secondary) {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.5rem 1.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		text-decoration: none;
		border-radius: 8px;
	}
	.mde-preview :global(a.btn-primary) {
		color: #fff;
		background: #18181b;
	}
	.mde-preview :global(a.btn-secondary) {
		color: #52525b;
		border: 1px solid #d4d4d8;
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

</style>
