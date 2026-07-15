<script lang="ts">
	import { deserialize, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import Badge from '$lib/components/Badge.svelte';
	import PublishedToggleBadge from '$lib/components/PublishedToggleBadge.svelte';
	import Button from '$lib/components/Button.svelte';
	import type { Project, Service } from '$lib/types';
	import {
		ORPHAN_GROUP_ID,
		buildProjectFilters,
		newProjectHref,
		projectHref,
		serviceTitle
	} from './_projects';
	import { projectsBreadcrumbs } from './_projects';

	let { data } = $props();

	// «Все» — плоский список со сквозным порядком сайта (global_sort_order).
	// Фильтр по группе — порядок внутри группы (sort_order).
	const globalMode = $derived(data.category === '');
	// В «Без группы» реордер запрещён: у группы нет соответствующей услуги.
	const canDrag = $derived(globalMode || data.category !== ORPHAN_GROUP_ID);

	// Оптимистичный порядок после drag. Привязан к ссылке на baseItems:
	// при обновлении данных (invalidateAll, смена фильтра) черновик устаревает сам.
	let reorderDraft = $state<{ base: Project[]; items: Project[] } | null>(null);
	let draggingId = $state<string | null>(null);
	let dragOverId = $state<string | null>(null);
	let reordering = $state(false);

	const filters = $derived(buildProjectFilters(data.services, data.counts));
	const breadcrumbs = $derived(projectsBreadcrumbs(data.category, filters));

	const serviceLabels = $derived(
		new Map<string, string>(
			data.services.map((service: Service): [string, string] => [service.slug, serviceTitle(service)])
		)
	);

	const baseItems = $derived(
		globalMode
			? [...data.projects].sort(
					(a: Project, b: Project) =>
						a.global_sort_order - b.global_sort_order || a.slug.localeCompare(b.slug)
				)
			: groupItems(data.projects, data.services, data.category)
	);
	const items = $derived(reorderDraft?.base === baseItems ? reorderDraft.items : baseItems);

	function title(project: Project): string {
		const langs = Object.keys(project.translations);
		return project.translations['en']?.title ?? (langs.length ? project.translations[langs[0]]?.title : '') ?? project.slug;
	}

	function groupLabel(project: Project): string | null {
		return serviceLabels.get(project.category) ?? null;
	}

	function groupItems(projects: Project[], services: Service[], category: string): Project[] {
		const serviceSlugs = new Set(services.map((s) => s.slug));
		const inGroup =
			category === ORPHAN_GROUP_ID
				? projects.filter((project: Project) => !serviceSlugs.has(project.category))
				: projects.filter((project: Project) => project.category === category);
		return inGroup.sort(
			(a: Project, b: Project) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug)
		);
	}

	function moveItem(fromId: string, toId: string): Project[] | null {
		const source = items;
		const fromIdx = source.findIndex((project) => project.id === fromId);
		const toIdx = source.findIndex((project) => project.id === toId);
		if (fromIdx < 0 || toIdx < 0 || fromIdx === toIdx) return null;
		const next = [...source];
		const [item] = next.splice(fromIdx, 1);
		next.splice(toIdx, 0, item);
		reorderDraft = { base: baseItems, items: next };
		return next;
	}

	async function persistOrder(next: Project[]) {
		if (reordering) return;

		reordering = true;
		const previous = reorderDraft;
		try {
			const fd = new FormData();
			if (globalMode) {
				fd.set('order', JSON.stringify(next.map((project) => project.id)));
			} else {
				fd.set(
					'layout',
					JSON.stringify([{ group_id: data.category, ids: next.map((project) => project.id) }])
				);
			}
			const res = await fetch(globalMode ? '?/reorderGlobal' : '?/reorder', {
				method: 'POST',
				body: fd,
				headers: {
					accept: 'application/json',
					'x-sveltekit-action': 'true'
				}
			});
			const result = deserialize(await res.text());
			if (result.type === 'success') {
				toast.success('Порядок сохранён');
				await invalidateAll();
				reorderDraft = null;
				return;
			}
			reorderDraft = previous;
			const message =
				result.type === 'failure'
					? ((result.data as { error?: string } | undefined)?.error ?? 'Не удалось сохранить порядок')
					: 'Не удалось сохранить порядок';
			toast.error(message);
		} catch {
			reorderDraft = previous;
			toast.error('Не удалось сохранить порядок');
		} finally {
			reordering = false;
		}
	}

	function onRowDragOver(e: DragEvent, itemId: string) {
		e.preventDefault();
		if (!canDrag) return;
		if (draggingId && draggingId !== itemId) dragOverId = itemId;
	}

	async function onRowDrop(e: DragEvent, itemId: string) {
		e.preventDefault();
		const fromId = e.dataTransfer?.getData('text/plain') || draggingId;
		dragOverId = null;
		draggingId = null;
		if (!canDrag || !fromId || fromId === itemId) return;
		const next = moveItem(fromId, itemId);
		if (next) await persistOrder(next);
	}
</script>

<svelte:head>
	<title>Проекты — Piplos Admin</title>
</svelte:head>

<AdminPage title="Проекты" breadcrumbs={breadcrumbs}>
	{#snippet actions()}
		<Button
			variant="success"
			onclick={() => (location.href = newProjectHref(data.category, data.services))}
			disabled={!data.services.length}
		>
			+ Новый проект
		</Button>
	{/snippet}

	<div class="admin-sidebar-row">
		<nav class="admin-sidebar-nav" aria-label="Фильтр по группе">
			{#each filters as f (f.value)}
				<a href={f.href} class="sidebar-link" class:active={data.category === f.value}>
					<span class="sidebar-label">{f.label}</span>
					<span class="sidebar-count">{data.counts[f.value] ?? 0}</span>
				</a>
			{/each}
		</nav>

		<div class="admin-sidebar-content admin-sidebar-content--no-box">
			{#if data.error}
				<div class="admin-table-wrap admin-table-wrap--empty">
					<p class="text-muted">{data.error}</p>
				</div>
			{:else if !data.services.length}
				<div class="admin-table-wrap admin-table-wrap--empty">
					<p class="text-muted">Сначала создайте услугу в разделе «Услуги».</p>
				</div>
			{:else if !items.length}
				<div class="admin-table-wrap admin-table-wrap--empty">
					<p class="text-muted">
						{data.category === '' ? 'Проектов пока нет. Создайте первый.' : 'В этой группе проектов нет.'}
					</p>
				</div>
			{:else}
				<div class="admin-table-wrap" class:admin-table-wrap--busy={reordering}>
					<table class="chart-table">
						<thead>
							<tr>
								<th class="admin-table-cell-drag" aria-label="Порядок"></th>
								<th>Название</th>
								<th>Языки</th>
								<th>Статус</th>
								<th class="admin-table-cell-actions"></th>
							</tr>
						</thead>
						<tbody ondragover={(e) => e.preventDefault()}>
							{#each items as project (project.id)}
								<tr
									class:admin-table-row--dragging={draggingId === project.id}
									class:admin-table-row--over={dragOverId === project.id}
									ondragover={(e) => onRowDragOver(e, project.id)}
									ondragleave={() => {
										if (dragOverId === project.id) dragOverId = null;
									}}
									ondrop={(e) => void onRowDrop(e, project.id)}
								>
									<td class="admin-table-cell-drag">
										<button
											type="button"
											class="admin-drag-handle"
											draggable="true"
											title={globalMode
												? 'Перетащите — так проекты идут на сайте в разделе «все проекты»'
												: 'Перетащите для сортировки внутри группы'}
											aria-label="Перетащите для изменения порядка"
											disabled={reordering || !canDrag}
											ondragstart={(e) => {
												draggingId = project.id;
												e.dataTransfer?.setData('text/plain', project.id);
												if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
											}}
											ondragend={() => {
												draggingId = null;
												dragOverId = null;
											}}
										>
											<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
												<circle cx="9" cy="6" r="1.5" />
												<circle cx="15" cy="6" r="1.5" />
												<circle cx="9" cy="12" r="1.5" />
												<circle cx="15" cy="12" r="1.5" />
												<circle cx="9" cy="18" r="1.5" />
												<circle cx="15" cy="18" r="1.5" />
											</svg>
										</button>
									</td>
									<td class="chart-cell-main">
										<div class="cell-title">
											<a href={projectHref(project, data.services)} class="admin-text-link">{title(project)}</a>
											{#if project.featured}
												<Badge variant="warning" title="Избранный">★</Badge>
											{/if}
										</div>
										<div class="cell-sub">
											{#if globalMode}
												{#if groupLabel(project)}
													<span>{groupLabel(project)}</span>
												{:else}
													<Badge variant="warning">Без группы</Badge>
												{/if}
												<span class="cell-sub-sep" aria-hidden="true">·</span>
											{/if}
											<span>{project.slug}</span>
											{#if project.year}
												<span class="cell-sub-sep" aria-hidden="true">·</span>
												<span>{project.year}</span>
											{/if}
										</div>
									</td>
									<td class="cell-langs">
										{#each Object.keys(project.translations) as lang (lang)}
											<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
										{/each}
									</td>
									<td>
										<PublishedToggleBadge id={project.slug} published={project.published} />
									</td>
									<td class="admin-table-cell-actions">
										<div class="admin-actions-wrap">
											<a href={projectHref(project, data.services)} class="admin-action-btn" title="Редактировать" aria-label="Редактировать проект">
												<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
													<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
													<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
												</svg>
											</a>
											<form
												method="POST"
												action="?/delete"
												class="admin-action-form"
												use:enhance={deleteEnhance({
													message: 'Удалить проект?',
													onSuccess: async () => {
														toast.success('Проект удалён');
														await invalidateAll();
													},
													onError: () => toast.error('Не удалось удалить проект')
												})}
											>
												<input type="hidden" name="id" value={project.slug} />
												<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить проект">
													<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
														<polyline points="3 6 5 6 21 6" />
														<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
													</svg>
												</button>
											</form>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</AdminPage>

<style>
	:global(.admin-sidebar-nav) .sidebar-link {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.sidebar-label {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.sidebar-count {
		flex-shrink: 0;
		min-width: 1.375rem;
		height: 1.375rem;
		padding: 0 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		line-height: 1.375rem;
		text-align: center;
		color: #71717a;
		background: #e5e7eb;
		border-radius: 6px;
		box-sizing: border-box;
	}
	:global(.admin-sidebar-nav) a.active .sidebar-count {
		color: #374151;
		background: #f4f4f5;
	}
	:global(.cat-badge) {
		margin-right: 0.25rem;
	}
	.cell-title {
		display: flex;
		align-items: center;
		gap: 0.375rem;
	}
	.cell-sub {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 0.375rem;
		margin-top: 0.2rem;
		font-size: 0.75rem;
		font-weight: 400;
		color: #a1a1aa;
	}
	.cell-sub-sep {
		color: #d4d4d8;
	}
	.cell-langs {
		white-space: nowrap;
	}
</style>
