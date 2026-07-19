<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import FilePickerDrawer from '$lib/components/FilePickerDrawer.svelte';
	import Select from '$lib/components/Select.svelte';
	import { confirmAction } from '$lib/confirm.svelte';
	import { promptName } from '$lib/name-dialog.svelte';
	import {
		createFolder,
		deleteEntries,
		formatSize,
		isImageFile,
		moveEntries,
		pathCrumbs,
		renameEntry,
		uploadFile,
		type FileInfo
	} from '$lib/files';

	const PAGE_SIZE = 60;

	let { data } = $props();

	const listing = $derived(data.listing);
	const path = $derived(data.listing.path);
	const crumbs = $derived(pathCrumbs(data.listing.path));

	let selected = $state<string[]>([]);
	let search = $state('');
	let sortBy = $state('new');
	let visibleCount = $state(PAGE_SIZE);
	let uploading = $state(false);
	let busy = $state(false);
	let moveDrawerOpen = $state(false);
	let uploadInput = $state<HTMLInputElement | null>(null);

	let preview = $state<FileInfo | null>(null);
	let previewOpen = $state(false);

	let dragPaths = $state<string[]>([]);
	let dropTarget = $state<string | null>(null);
	let osDragDepth = $state(0);

	const query = $derived(search.trim().toLowerCase());
	const visibleFolders = $derived(
		listing.folders.filter((f) => !query || f.name.toLowerCase().includes(query))
	);
	const matchedFiles = $derived.by(() => {
		const files = listing.files.filter((f) => !query || f.name.toLowerCase().includes(query));
		if (sortBy === 'name') return files;
		return files.toSorted((a, b) =>
			sortBy === 'old'
				? a.mod_time.localeCompare(b.mod_time)
				: b.mod_time.localeCompare(a.mod_time)
		);
	});
	const shownFiles = $derived(matchedFiles.slice(0, visibleCount));

	const allPaths = $derived([
		...visibleFolders.map((f) => f.path),
		...matchedFiles.map((f) => f.path)
	]);
	const allSelected = $derived(allPaths.length > 0 && selected.length === allPaths.length);

	function navigate(target: string) {
		selected = [];
		search = '';
		visibleCount = PAGE_SIZE;
		void goto(target ? `/files?path=${encodeURIComponent(target)}` : '/files', {
			keepFocus: true
		});
	}

	async function refresh() {
		selected = [];
		await invalidateAll();
	}

	function toggle(entryPath: string) {
		selected = selected.includes(entryPath)
			? selected.filter((p) => p !== entryPath)
			: [...selected, entryPath];
	}

	function toggleAll() {
		selected = allSelected ? [] : [...allPaths];
	}

	function fail(e: unknown, fallback: string) {
		toast.error(e instanceof Error ? e.message : fallback);
	}

	// ---------- операции ----------

	async function uploadAll(files: File[]) {
		if (!files.length) return;
		uploading = true;
		try {
			for (const file of files) {
				await uploadFile(file, path);
			}
			toast.success(files.length === 1 ? 'Файл загружен' : `Загружено файлов: ${files.length}`);
			await refresh();
		} catch (e) {
			fail(e, 'Не удалось загрузить файл');
		} finally {
			uploading = false;
		}
	}

	function onUploadChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const files = [...(input.files ?? [])];
		input.value = '';
		void uploadAll(files);
	}

	async function onCreateFolder() {
		const name = await promptName({
			title: 'Новая папка',
			label: 'Название папки',
			placeholder: 'images'
		});
		if (!name) return;
		const next = path ? `${path}/${name}` : name;
		busy = true;
		try {
			await createFolder(next);
			navigate(next);
		} catch (e) {
			fail(e, 'Не удалось создать папку');
		} finally {
			busy = false;
		}
	}

	async function onRename(
		entryPath: string,
		currentName: string,
		kind: 'folder' | 'file' = 'folder'
	): Promise<{ path: string; url: string } | null> {
		const name = await promptName({
			title: kind === 'folder' ? 'Переименовать папку' : 'Переименовать файл',
			label: 'Новое название',
			value: currentName
		});
		if (!name || name === currentName) return null;
		const parent = entryPath.split('/').slice(0, -1).join('/');
		busy = true;
		try {
			const res = await renameEntry(entryPath, parent ? `${parent}/${name}` : name);
			await refresh();
			return res;
		} catch (e) {
			fail(e, 'Не удалось переименовать');
			return null;
		} finally {
			busy = false;
		}
	}

	async function onDelete(paths: string[]) {
		const message =
			paths.length === 1
				? `Удалить «${paths[0].split('/').pop()}»? Папки удаляются со всем содержимым.`
				: `Удалить выбранное (${paths.length})? Папки удаляются со всем содержимым.`;
		if (!(await confirmAction({ message }))) return false;
		busy = true;
		try {
			await deleteEntries(paths);
			toast.success('Удалено');
			await refresh();
			return true;
		} catch (e) {
			fail(e, 'Не удалось удалить');
			return false;
		} finally {
			busy = false;
		}
	}

	async function moveTo(paths: string[], dest: string) {
		busy = true;
		try {
			const res = await moveEntries(paths, dest);
			if (res.moved.length) toast.success(`Перемещено: ${res.moved.length}`);
			await refresh();
		} catch (e) {
			fail(e, 'Не удалось переместить');
		} finally {
			busy = false;
		}
	}

	// ---------- просмотр ----------

	function openPreview(file: FileInfo) {
		preview = file;
		previewOpen = true;
	}

	async function copyUrl(url: string) {
		try {
			await navigator.clipboard.writeText(url);
			toast.success('Ссылка скопирована');
		} catch {
			toast.error('Не удалось скопировать');
		}
	}

	async function renamePreview() {
		if (!preview) return;
		const res = await onRename(preview.path, preview.name, 'file');
		if (res) {
			preview = { ...preview, name: res.path.split('/').pop() ?? preview.name, path: res.path, url: res.url };
		}
	}

	async function deletePreview() {
		if (!preview) return;
		if (await onDelete([preview.path])) previewOpen = false;
	}

	// ---------- drag & drop ----------

	function onDragStart(e: DragEvent, entryPath: string) {
		dragPaths = selected.includes(entryPath) ? [...selected] : [entryPath];
		if (e.dataTransfer) {
			e.dataTransfer.setData('text/plain', dragPaths.join('\n'));
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function onDragEnd() {
		dragPaths = [];
		dropTarget = null;
	}

	function canDropInto(dest: string): boolean {
		if (!dragPaths.length) return false;
		return dragPaths.every(
			(p) => p !== dest && !dest.startsWith(p + '/') && p.split('/').slice(0, -1).join('/') !== dest
		);
	}

	function onFolderDragOver(e: DragEvent, dest: string) {
		if (!canDropInto(dest)) return;
		e.preventDefault();
		if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
		dropTarget = dest;
	}

	function onFolderDrop(e: DragEvent, dest: string) {
		e.preventDefault();
		dropTarget = null;
		if (!canDropInto(dest)) return;
		const paths = [...dragPaths];
		dragPaths = [];
		void moveTo(paths, dest);
	}

	function isOsFileDrag(e: DragEvent): boolean {
		return !!e.dataTransfer && [...e.dataTransfer.types].includes('Files');
	}

	function onZoneDragEnter(e: DragEvent) {
		if (!isOsFileDrag(e)) return;
		e.preventDefault();
		osDragDepth += 1;
	}

	function onZoneDragOver(e: DragEvent) {
		if (!isOsFileDrag(e)) return;
		e.preventDefault();
		if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
	}

	function onZoneDragLeave(e: DragEvent) {
		if (!isOsFileDrag(e)) return;
		osDragDepth = Math.max(0, osDragDepth - 1);
	}

	function onZoneDrop(e: DragEvent) {
		if (!isOsFileDrag(e)) return;
		e.preventDefault();
		osDragDepth = 0;
		void uploadAll([...(e.dataTransfer?.files ?? [])]);
	}
</script>

<svelte:head>
	<title>Файлы — Piplos Admin</title>
</svelte:head>

<AdminPage title="Файлы" breadcrumbs={[{ label: 'Файлы' }]}>
	{#snippet actions()}
		<div class="head-actions">
			<Button variant="secondary" disabled={busy} onclick={onCreateFolder}>Новая папка</Button>
			<Button loading={uploading} onclick={() => uploadInput?.click()}>Загрузить</Button>
		</div>
	{/snippet}

	<input
		type="file"
		accept="image/*"
		multiple
		bind:this={uploadInput}
		onchange={onUploadChange}
		hidden
	/>

	<Card padding="sm">
		<div
			class="zone"
			class:zone--os-drag={osDragDepth > 0}
			role="region"
			aria-label="Файлы текущей папки"
			ondragenter={onZoneDragEnter}
			ondragover={onZoneDragOver}
			ondragleave={onZoneDragLeave}
			ondrop={onZoneDrop}
		>
			<nav class="crumbs" aria-label="Путь">
				<button
					type="button"
					class="crumb"
					class:crumb--drop={dropTarget === ''}
					disabled={!path && !dragPaths.length}
					onclick={() => navigate('')}
					ondragover={(e) => onFolderDragOver(e, '')}
					ondragleave={() => (dropTarget = dropTarget === '' ? null : dropTarget)}
					ondrop={(e) => onFolderDrop(e, '')}
				>
					Корень
				</button>
				{#each crumbs as crumb (crumb.path)}
					<span class="crumb-sep" aria-hidden="true">/</span>
					<button
						type="button"
						class="crumb"
						class:crumb--drop={dropTarget === crumb.path}
						disabled={crumb.path === path && !dragPaths.length}
						onclick={() => navigate(crumb.path)}
						ondragover={(e) => onFolderDragOver(e, crumb.path)}
						ondragleave={() => (dropTarget = dropTarget === crumb.path ? null : dropTarget)}
						ondrop={(e) => onFolderDrop(e, crumb.path)}
					>
						{crumb.name}
					</button>
				{/each}
			</nav>

			<div class="controls" class:controls--selection={selected.length > 0}>
				<div class="controls-left">
					{#if allPaths.length}
						<label class="select-all">
							<input
								type="checkbox"
								class="admin-checkbox"
								checked={allSelected}
								indeterminate={selected.length > 0 && !allSelected}
								onchange={toggleAll}
							/>
							{#if selected.length}
								Выбрано: {selected.length}
							{:else}
								Выбрать все ({allPaths.length})
							{/if}
						</label>
					{/if}
					{#if selected.length}
						<div class="bulk-actions">
							<Button variant="secondary" disabled={busy} onclick={() => (moveDrawerOpen = true)}>
								Переместить
							</Button>
							<Button variant="danger" disabled={busy} onclick={() => onDelete(selected)}>
								Удалить
							</Button>
							<Button variant="ghost" onclick={() => (selected = [])}>Отмена</Button>
						</div>
					{/if}
				</div>
				<div class="controls-right">
					<input
						type="search"
						class="search"
						placeholder="Поиск по имени…"
						bind:value={search}
						oninput={() => (visibleCount = PAGE_SIZE)}
					/>
					<Select
						bind:value={sortBy}
						ariaLabel="Сортировка"
						class="sort-select"
						options={[
							{ value: 'new', label: 'Сначала новые' },
							{ value: 'old', label: 'Сначала старые' },
							{ value: 'name', label: 'По имени' }
						]}
					/>
				</div>
			</div>

			{#if visibleFolders.length}
				<h3 class="admin-block-label">Папки ({visibleFolders.length})</h3>
				<div class="folders" role="list" aria-label="Папки">
					{#each visibleFolders as folder (folder.path)}
						<div
							class="folder-chip"
							class:folder-chip--selected={selected.includes(folder.path)}
							class:folder-chip--drop={dropTarget === folder.path}
							role="listitem"
							draggable="true"
							ondragstart={(e) => onDragStart(e, folder.path)}
							ondragend={onDragEnd}
							ondragover={(e) => onFolderDragOver(e, folder.path)}
							ondragleave={() => (dropTarget = dropTarget === folder.path ? null : dropTarget)}
							ondrop={(e) => onFolderDrop(e, folder.path)}
						>
							<input
								type="checkbox"
								class="admin-checkbox folder-check"
								checked={selected.includes(folder.path)}
								onchange={() => toggle(folder.path)}
								aria-label="Выбрать {folder.name}"
							/>
							<button type="button" class="folder-main" onclick={() => navigate(folder.path)} title={folder.name}>
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
									<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
								</svg>
								<span class="folder-name">{folder.name}</span>
							</button>
							<button
								type="button"
								class="folder-rename"
								title="Переименовать"
								onclick={() => onRename(folder.path, folder.name)}
							>
								✎
							</button>
						</div>
					{/each}
				</div>
			{/if}

			{#if !visibleFolders.length && !shownFiles.length}
				<p class="empty">
					{query ? 'Ничего не найдено' : 'Папка пуста. Загрузите файлы или перетащите их сюда.'}
				</p>
			{:else if shownFiles.length}
				<h3 class="admin-block-label">Файлы ({matchedFiles.length})</h3>
				<div class="grid">
					{#each shownFiles as file (file.path)}
						<div
							class="item"
							class:item--selected={selected.includes(file.path)}
							class:item--dragging={dragPaths.includes(file.path)}
							draggable="true"
							ondragstart={(e) => onDragStart(e, file.path)}
							ondragend={onDragEnd}
							role="listitem"
						>
							<input
								type="checkbox"
								class="admin-checkbox item-check"
								checked={selected.includes(file.path)}
								onchange={() => toggle(file.path)}
								aria-label="Выбрать {file.name}"
							/>
							<button type="button" class="item-main" onclick={() => openPreview(file)} title="{file.name} · {formatSize(file.size)}">
								<span class="thumb">
									{#if isImageFile(file.name)}
										<img src={file.url} alt={file.name} loading="lazy" />
									{:else}
										<svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
											<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
											<polyline points="14 2 14 8 20 8" />
										</svg>
									{/if}
								</span>
								<span class="item-name">{file.name}</span>
							</button>
						</div>
					{/each}
				</div>
				{#if matchedFiles.length > visibleCount}
					<div class="more">
						<Button variant="secondary" onclick={() => (visibleCount += PAGE_SIZE)}>
							Показать ещё ({matchedFiles.length - visibleCount})
						</Button>
					</div>
				{/if}
			{/if}

		</div>
	</Card>
</AdminPage>

<Drawer bind:open={previewOpen} title={preview?.name ?? 'Файл'}>
	{#if preview}
		<div class="preview">
			{#if isImageFile(preview.name)}
				<a class="preview-img" href={preview.url} target="_blank" rel="noreferrer" title="Открыть оригинал">
					<img src={preview.url} alt={preview.name} />
				</a>
			{/if}
			<dl class="preview-meta">
				<dt>Размер</dt>
				<dd>{formatSize(preview.size)}</dd>
				<dt>Изменён</dt>
				<dd>{new Date(preview.mod_time).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' })}</dd>
				<dt>Путь</dt>
				<dd>{preview.path}</dd>
			</dl>
			<div class="preview-url">
				<input type="text" readonly value={preview.url} aria-label="Ссылка на файл" />
				<Button variant="secondary" onclick={() => preview && copyUrl(preview.url)}>Копировать</Button>
			</div>
			<div class="preview-actions">
				<Button variant="secondary" disabled={busy} onclick={renamePreview}>Переименовать</Button>
				<a class="preview-open" href={preview.url} target="_blank" rel="noreferrer">Открыть оригинал</a>
				<Button variant="danger" disabled={busy} onclick={deletePreview}>Удалить</Button>
			</div>
		</div>
	{/if}
</Drawer>

<FilePickerDrawer
	bind:open={moveDrawerOpen}
	title="Куда переместить"
	mode="folder"
	disabledPaths={selected}
	onselect={(dest) => void moveTo(selected, dest.path)}
/>

<style>
	.head-actions {
		display: flex;
		gap: 0.5rem;
	}
	.zone {
		border-radius: 10px;
		transition: box-shadow 0.15s;
	}
	.zone--os-drag {
		box-shadow: 0 0 0 2px #2563eb;
	}
	.controls {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		flex-wrap: wrap;
		min-height: 2.75rem;
		padding: 0.25rem 0.5rem;
		margin-bottom: 1rem;
		border-radius: 10px;
		transition: background 0.15s;
	}
	.controls--selection {
		background: #f4f4f5;
	}
	.controls-left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}
	.controls-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.bulk-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.select-all {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #18181b;
		cursor: pointer;
		white-space: nowrap;
	}
	.search {
		width: 13rem;
		min-height: 2.25rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		box-sizing: border-box;
	}
	.search:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	:global(.sort-select) {
		width: auto;
	}
	.crumbs {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 0.125rem;
		margin-bottom: 0.5rem;
	}
	.crumb {
		padding: 0.25rem 0.5rem;
		font-size: 0.875rem;
		color: #2563eb;
		background: transparent;
		border: 1px dashed transparent;
		border-radius: 6px;
		cursor: pointer;
	}
	.crumb:hover:not(:disabled) {
		background: #eff6ff;
	}
	.crumb:disabled {
		color: #18181b;
		font-weight: 600;
		cursor: default;
	}
	.crumb--drop {
		border-color: #2563eb;
		background: #eff6ff;
	}
	.crumb-sep {
		color: #a1a1aa;
		font-size: 0.875rem;
	}
	.folders {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1.25rem;
	}
	.folder-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		max-width: 16rem;
		padding: 0.25rem 0.375rem;
		background: #fafafa;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
	}
	.folder-chip:hover {
		border-color: #9ca3af;
	}
	.folder-chip--selected {
		border-color: #2563eb;
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
	}
	.folder-chip--drop {
		border-color: #2563eb;
		background: #eff6ff;
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.3);
	}
	.folder-check {
		margin: 0 0.125rem;
	}
	.folder-main {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		min-width: 0;
		padding: 0.125rem 0.25rem;
		font-size: 0.8125rem;
		color: #374151;
		background: transparent;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}
	.folder-main svg {
		flex-shrink: 0;
		color: #eab308;
	}
	.folder-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.folder-rename {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.375rem;
		height: 1.375rem;
		font-size: 0.75rem;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		opacity: 0;
		transition: opacity 0.15s;
	}
	.folder-chip:hover .folder-rename,
	.folder-rename:focus-visible {
		opacity: 1;
	}
	.folder-rename:hover {
		background: #f3f4f6;
		color: #18181b;
	}
	.empty {
		margin: 2.5rem 0;
		text-align: center;
		font-size: 0.875rem;
		color: #71717a;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
		gap: 0.75rem;
	}
	.item {
		position: relative;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		transition: border-color 0.15s, box-shadow 0.15s, opacity 0.15s;
	}
	.item:hover {
		border-color: #9ca3af;
	}
	.item--selected {
		border-color: #2563eb;
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
	}
	.item--dragging {
		opacity: 0.4;
	}
	.item-check {
		position: absolute;
		top: 0.5rem;
		left: 0.5rem;
		z-index: 1;
		opacity: 0;
		filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.9));
		transition: opacity 0.15s;
	}
	.item:hover .item-check,
	.item-check:checked,
	.item-check:focus-visible {
		opacity: 1;
	}
	.item-main {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		width: 100%;
		padding: 0.375rem;
		background: transparent;
		border: none;
		border-radius: 10px;
		cursor: pointer;
		box-sizing: border-box;
	}
	.thumb {
		display: flex;
		align-items: center;
		justify-content: center;
		aspect-ratio: 4 / 3;
		color: #a1a1aa;
		background: #f4f4f5;
		border-radius: 6px;
		overflow: hidden;
	}
	.thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.item-name {
		padding: 0 0.125rem 0.125rem;
		font-size: 0.75rem;
		color: #18181b;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-align: center;
	}
	.more {
		display: flex;
		justify-content: center;
		margin-top: 1rem;
	}
	.preview {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.preview-img {
		display: block;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
		background: #f4f4f5;
	}
	.preview-img img {
		display: block;
		width: 100%;
		max-height: 22rem;
		object-fit: contain;
	}
	.preview-meta {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.375rem 1rem;
		margin: 0;
		font-size: 0.8125rem;
	}
	.preview-meta dt {
		color: #71717a;
	}
	.preview-meta dd {
		margin: 0;
		color: #18181b;
		word-break: break-all;
	}
	.preview-url {
		display: flex;
		gap: 0.5rem;
	}
	.preview-url input {
		flex: 1;
		min-width: 0;
		min-height: 2.25rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		color: #374151;
		background: #fafafa;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		box-sizing: border-box;
	}
	.preview-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.preview-open {
		margin-right: auto;
		font-size: 0.875rem;
		color: #2563eb;
		text-decoration: none;
	}
	.preview-open:hover {
		text-decoration: underline;
	}
</style>
