<script lang="ts">
	import { untrack } from 'svelte';
	import { Dialog } from 'bits-ui';
	import { fade, fly } from 'svelte/transition';
	import toast from 'svelte-french-toast';
	import Button from './Button.svelte';
	import {
		listFiles,
		isImageFile,
		pathCrumbs,
		formatSize,
		type FileInfo,
		type FileListing
	} from '$lib/files';

	const PAGE_SIZE = 60;

	interface Props {
		open?: boolean;
		title?: string;
		/** file — выбор файла; folder — выбор папки назначения. */
		mode?: 'file' | 'folder';
		/** Показывать только изображения (для mode="file"). */
		imagesOnly?: boolean;
		/** Пути, которые нельзя выбрать/открыть (например, перемещаемые папки). */
		disabledPaths?: string[];
		onselect: (value: { path: string; url: string }) => void;
	}
	let {
		open = $bindable(false),
		title = 'Файловый архив',
		mode = 'file',
		imagesOnly = true,
		disabledPaths = [],
		onselect
	}: Props = $props();

	let path = $state('');
	let listing = $state<FileListing | null>(null);
	let loading = $state(false);
	let search = $state('');
	let visibleCount = $state(PAGE_SIZE);

	async function load(target: string) {
		loading = true;
		try {
			listing = await listFiles(target);
			path = target;
			search = '';
			visibleCount = PAGE_SIZE;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Не удалось загрузить список файлов');
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (open) void load(untrack(() => path));
	});

	const crumbs = $derived(pathCrumbs(path));
	const blocked = $derived(new Set(disabledPaths));
	const query = $derived(search.trim().toLowerCase());

	const visibleFolders = $derived(
		(listing?.folders ?? []).filter((f) => !query || f.name.toLowerCase().includes(query))
	);
	const matchedFiles = $derived(
		mode === 'folder'
			? []
			: (listing?.files ?? [])
					.filter((f) => !imagesOnly || isImageFile(f.name))
					.filter((f) => !query || f.name.toLowerCase().includes(query))
					.toSorted((a, b) => b.mod_time.localeCompare(a.mod_time))
	);
	const shownFiles = $derived(matchedFiles.slice(0, visibleCount));

	function pickFile(file: FileInfo) {
		onselect({ path: file.path, url: file.url });
		open = false;
	}

	function pickFolder() {
		onselect({ path, url: '' });
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay forceMount>
			{#snippet child({ props: overlayProps, open: isOpen })}
				{#if isOpen}
					<div
						{...overlayProps}
						class="fp-overlay"
						in:fade={{ duration: 200 }}
						out:fade={{ duration: 150 }}
					></div>
				{/if}
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content forceMount>
			{#snippet child({ props: contentProps, open: isOpen })}
				{#if isOpen}
					<div
						{...contentProps}
						class="fp-panel"
						in:fly={{ duration: 250, x: 560, opacity: 1 }}
						out:fly={{ duration: 200, x: 560, opacity: 1 }}
					>
						<header class="fp-head">
							<Dialog.Title class="fp-title">{title}</Dialog.Title>
							<Dialog.Close class="fp-close" aria-label="Закрыть">
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
									<line x1="18" y1="6" x2="6" y2="18" />
									<line x1="6" y1="6" x2="18" y2="18" />
								</svg>
							</Dialog.Close>
						</header>

						<div class="fp-toolbar">
							<nav class="fp-crumbs" aria-label="Путь">
								<button type="button" class="fp-crumb" disabled={!path} onclick={() => load('')}>
									Корень
								</button>
								{#each crumbs as crumb (crumb.path)}
									<span class="fp-crumb-sep" aria-hidden="true">/</span>
									<button
										type="button"
										class="fp-crumb"
										disabled={crumb.path === path}
										onclick={() => load(crumb.path)}
									>
										{crumb.name}
									</button>
								{/each}
							</nav>
							{#if mode === 'file'}
								<input
									type="search"
									class="fp-search"
									placeholder="Поиск по имени…"
									bind:value={search}
								/>
							{/if}
						</div>

						<div class="fp-body">
							{#if loading}
								<p class="fp-empty">Загрузка…</p>
							{:else if visibleFolders.length === 0 && shownFiles.length === 0}
								<p class="fp-empty">{query ? 'Ничего не найдено' : 'Папка пуста'}</p>
							{:else}
								{#if visibleFolders.length}
									<div class="fp-folders">
										{#each visibleFolders as folder (folder.path)}
											<button
												type="button"
												class="fp-folder"
												disabled={blocked.has(folder.path)}
												onclick={() => load(folder.path)}
												title={folder.name}
											>
												<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
													<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
												</svg>
												<span class="fp-folder-name">{folder.name}</span>
											</button>
										{/each}
									</div>
								{/if}
								{#if shownFiles.length}
									<div class="fp-grid">
										{#each shownFiles as file (file.path)}
											<button
												type="button"
												class="fp-item"
												onclick={() => pickFile(file)}
												title="{file.name} · {formatSize(file.size)}"
											>
												<span class="fp-thumb">
													{#if isImageFile(file.name)}
														<img src={file.url} alt={file.name} loading="lazy" />
													{:else}
														<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
															<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
															<polyline points="14 2 14 8 20 8" />
														</svg>
													{/if}
												</span>
												<span class="fp-name">{file.name}</span>
											</button>
										{/each}
									</div>
									{#if matchedFiles.length > visibleCount}
										<div class="fp-more">
											<Button variant="secondary" onclick={() => (visibleCount += PAGE_SIZE)}>
												Показать ещё ({matchedFiles.length - visibleCount})
											</Button>
										</div>
									{/if}
								{/if}
							{/if}
						</div>

						{#if mode === 'folder'}
							<footer class="fp-footer">
								<Button variant="primary" onclick={pickFolder}>Переместить сюда</Button>
							</footer>
						{/if}
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.fp-overlay {
		position: fixed;
		inset: 0;
		z-index: 60;
		background: rgba(24, 24, 27, 0.35);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
	}
	.fp-panel {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		z-index: 70;
		width: min(40rem, 100vw);
		display: flex;
		flex-direction: column;
		background: #fff;
		box-shadow: -12px 0 40px rgba(0, 0, 0, 0.15);
		outline: none;
	}
	.fp-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid #e5e7eb;
		flex-shrink: 0;
	}
	:global(.fp-title) {
		margin: 0;
		font-size: 1.0625rem;
		font-weight: 600;
		color: #18181b;
	}
	:global(.fp-close) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		padding: 0;
		color: #71717a;
		background: transparent;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	:global(.fp-close:hover) {
		background: #f4f4f5;
		color: #18181b;
	}
	.fp-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		flex-wrap: wrap;
		padding: 0.625rem 1.25rem;
		border-bottom: 1px solid #f1f1f2;
		flex-shrink: 0;
	}
	.fp-crumbs {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 0.125rem;
		min-width: 0;
	}
	.fp-crumb {
		padding: 0.125rem 0.375rem;
		font-size: 0.8125rem;
		color: #2563eb;
		background: transparent;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}
	.fp-crumb:hover:not(:disabled) {
		background: #eff6ff;
	}
	.fp-crumb:disabled {
		color: #18181b;
		font-weight: 500;
		cursor: default;
	}
	.fp-crumb-sep {
		color: #a1a1aa;
		font-size: 0.8125rem;
	}
	.fp-search {
		width: 12rem;
		min-height: 2rem;
		padding: 0.25rem 0.625rem;
		font-size: 0.8125rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		box-sizing: border-box;
	}
	.fp-search:focus {
		outline: none;
		border-color: #111;
		box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
	}
	.fp-body {
		flex: 1;
		overflow-y: auto;
		padding: 1rem 1.25rem;
	}
	.fp-empty {
		margin: 2.5rem 0;
		text-align: center;
		font-size: 0.875rem;
		color: #71717a;
	}
	.fp-folders {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}
	.fp-folder {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		max-width: 100%;
		padding: 0.375rem 0.625rem;
		font-size: 0.8125rem;
		color: #374151;
		background: #fafafa;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		cursor: pointer;
		transition: border-color 0.15s, background 0.15s;
	}
	.fp-folder svg {
		flex-shrink: 0;
		color: #eab308;
	}
	.fp-folder:hover:not(:disabled) {
		background: #f4f4f5;
		border-color: #9ca3af;
	}
	.fp-folder:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	.fp-folder-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.fp-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr));
		gap: 0.5rem;
	}
	.fp-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 0.375rem;
		background: transparent;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		cursor: pointer;
		transition: border-color 0.15s, background 0.15s;
	}
	.fp-item:hover {
		background: #f9fafb;
		border-color: #9ca3af;
	}
	.fp-thumb {
		display: flex;
		align-items: center;
		justify-content: center;
		aspect-ratio: 4 / 3;
		color: #a1a1aa;
		background: #f4f4f5;
		border-radius: 6px;
		overflow: hidden;
	}
	.fp-thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.fp-name {
		font-size: 0.6875rem;
		color: #374151;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-align: center;
	}
	.fp-more {
		display: flex;
		justify-content: center;
		margin-top: 1rem;
	}
	.fp-footer {
		display: flex;
		justify-content: flex-end;
		padding: 0.75rem 1.25rem;
		border-top: 1px solid #e5e7eb;
		flex-shrink: 0;
	}
</style>
