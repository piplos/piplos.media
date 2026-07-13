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

	const ORPHAN_GROUP_ID = '__orphan__';

	type ProjectGroup = { id: string; label: string; items: Project[] };

	let { data } = $props();

	let groups = $state<ProjectGroup[]>([]);
	let draggingId = $state<string | null>(null);
	let dragOver = $state<{ groupId: string; itemId: string | null } | null>(null);
	let reordering = $state(false);

	function title(project: Project): string {
		const langs = Object.keys(project.translations);
		return project.translations['en']?.title ?? (langs.length ? project.translations[langs[0]]?.title : '') ?? project.slug;
	}

	function serviceTitle(service: Service): string {
		const langs = Object.keys(service.translations);
		return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
	}

	function buildGroups(services: Service[], projects: Project[]): ProjectGroup[] {
		const sortedServices = [...services].sort(
			(a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug)
		);
		const serviceSlugs = new Set(sortedServices.map((s) => s.slug));

		const result: ProjectGroup[] = sortedServices.map((service) => ({
			id: service.slug,
			label: serviceTitle(service),
			items: projects
				.filter((project) => project.category === service.slug)
				.sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug))
		}));

		const orphans = projects
			.filter((project) => !serviceSlugs.has(project.category))
			.sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug));
		if (orphans.length) {
			result.push({ id: ORPHAN_GROUP_ID, label: 'Без группы', items: orphans });
		}
		return result;
	}

	$effect(() => {
		groups = buildGroups(data.services, data.projects);
	});

	function findItem(itemId: string): { groupIdx: number; itemIdx: number } | null {
		for (let groupIdx = 0; groupIdx < groups.length; groupIdx++) {
			const itemIdx = groups[groupIdx].items.findIndex((item) => item.id === itemId);
			if (itemIdx >= 0) return { groupIdx, itemIdx };
		}
		return null;
	}

	function moveItem(fromId: string, toGroupId: string, beforeItemId: string | null) {
		if (toGroupId === ORPHAN_GROUP_ID) return;
		const from = findItem(fromId);
		if (!from) return;
		const toGroupIdx = groups.findIndex((group) => group.id === toGroupId);
		if (toGroupIdx < 0) return;

		const next = groups.map((group) => ({ ...group, items: [...group.items] }));
		const [item] = next[from.groupIdx].items.splice(from.itemIdx, 1);

		let insertIdx = beforeItemId
			? next[toGroupIdx].items.findIndex((entry) => entry.id === beforeItemId)
			: next[toGroupIdx].items.length;
		if (insertIdx < 0) insertIdx = next[toGroupIdx].items.length;
		if (from.groupIdx === toGroupIdx && from.itemIdx < insertIdx) insertIdx--;

		next[toGroupIdx].items.splice(insertIdx, 0, item);
		groups = next;
	}

	function layoutPayload(source: ProjectGroup[]) {
		return source
			.filter((group) => group.id !== ORPHAN_GROUP_ID)
			.map((group) => ({ group_id: group.id, ids: group.items.map((item) => item.id) }));
	}

	async function persistLayout() {
		if (reordering) return;
		const orphanGroup = groups.find((group) => group.id === ORPHAN_GROUP_ID);
		if (orphanGroup?.items.length) {
			toast.error('Переместите проекты без группы в одну из услуг');
			return;
		}

		reordering = true;
		const previous = groups.map((group) => ({ ...group, items: [...group.items] }));
		try {
			const fd = new FormData();
			fd.set('layout', JSON.stringify(layoutPayload(groups)));
			const res = await fetch('?/reorder', {
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
				return;
			}
			groups = previous;
			const message =
				result.type === 'failure'
					? ((result.data as { error?: string } | undefined)?.error ?? 'Не удалось сохранить порядок')
					: 'Не удалось сохранить порядок';
			toast.error(message);
		} catch {
			groups = previous;
			toast.error('Не удалось сохранить порядок');
		} finally {
			reordering = false;
		}
	}

	function onRowDragOver(e: DragEvent, groupId: string, itemId: string) {
		e.preventDefault();
		if (groupId === ORPHAN_GROUP_ID) return;
		if (draggingId && draggingId !== itemId) dragOver = { groupId, itemId };
	}

	function onGroupDragOver(e: DragEvent, groupId: string) {
		e.preventDefault();
		if (groupId === ORPHAN_GROUP_ID) return;
		if (draggingId) dragOver = { groupId, itemId: null };
	}

	async function onRowDrop(e: DragEvent, groupId: string, itemId: string) {
		e.preventDefault();
		const fromId = e.dataTransfer?.getData('text/plain') || draggingId;
		if (!fromId || fromId === itemId || groupId === ORPHAN_GROUP_ID) return;
		moveItem(fromId, groupId, itemId);
		dragOver = null;
		draggingId = null;
		await persistLayout();
	}

	async function onGroupDrop(e: DragEvent, groupId: string) {
		e.preventDefault();
		const fromId = e.dataTransfer?.getData('text/plain') || draggingId;
		if (!fromId || groupId === ORPHAN_GROUP_ID) return;
		moveItem(fromId, groupId, null);
		dragOver = null;
		draggingId = null;
		await persistLayout();
	}

	const hasItems = $derived(groups.some((group) => group.items.length > 0));
</script>

<svelte:head>
	<title>Проекты — Piplos Admin</title>
</svelte:head>

<AdminPage title="Проекты">
	{#snippet actions()}
		<Button variant="success" onclick={() => (location.href = '/projects/new')} disabled={!data.services.length}>
			+ Новый проект
		</Button>
	{/snippet}

	{#if data.error}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">{data.error}</p>
		</div>
	{:else if !data.services.length}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">Сначала создайте услугу в разделе «Списки → Услуги».</p>
		</div>
	{:else if !hasItems}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">Проектов пока нет. Создайте первый.</p>
		</div>
	{:else}
		<div class="admin-table-wrap" class:admin-table-wrap--busy={reordering}>
			<table class="chart-table">
				<thead>
					<tr>
						<th class="admin-table-cell-drag" aria-label="Порядок"></th>
						<th>Название</th>
						<th>Slug</th>
						<th>Год</th>
						<th>Языки</th>
						<th>Статус</th>
						<th class="admin-table-cell-actions"></th>
					</tr>
				</thead>
				<tbody ondragover={(e) => e.preventDefault()}>
					{#each groups as group (group.id)}
						{#if group.items.length > 0 || group.id !== ORPHAN_GROUP_ID}
							<tr
								class="project-group-header"
								class:project-group-header--warning={group.id === ORPHAN_GROUP_ID}
								class:project-group-header--over={dragOver?.groupId === group.id && dragOver.itemId === null}
								ondragover={(e) => onGroupDragOver(e, group.id)}
								ondragleave={() => {
									if (dragOver?.groupId === group.id && dragOver.itemId === null) dragOver = null;
								}}
								ondrop={(e) => void onGroupDrop(e, group.id)}
							>
								<td colspan="7">{group.label}</td>
							</tr>
						{/if}
						{#each group.items as project (project.id)}
							<tr
								class:admin-table-row--dragging={draggingId === project.id}
								class:admin-table-row--over={dragOver?.groupId === group.id && dragOver.itemId === project.id}
								ondragover={(e) => onRowDragOver(e, group.id, project.id)}
								ondragleave={() => {
									if (dragOver?.groupId === group.id && dragOver.itemId === project.id) dragOver = null;
								}}
								ondrop={(e) => void onRowDrop(e, group.id, project.id)}
							>
								<td class="admin-table-cell-drag">
									<button
										type="button"
										class="admin-drag-handle"
										draggable="true"
										title="Перетащите для сортировки и группировки"
										aria-label="Перетащите для изменения порядка или группы"
										disabled={reordering}
										ondragstart={(e) => {
											draggingId = project.id;
											e.dataTransfer?.setData('text/plain', project.id);
											if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
										}}
										ondragend={() => {
											draggingId = null;
											dragOver = null;
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
									<a href="/projects/{project.slug}" class="admin-text-link">{title(project)}</a>
									{#if project.featured}
										<Badge variant="warning" title="Избранный">★</Badge>
									{/if}
								</td>
								<td class="chart-cell-muted">{project.slug}</td>
								<td class="chart-cell-muted">{project.year || '—'}</td>
								<td>
									{#each Object.keys(project.translations) as lang (lang)}
										<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
									{/each}
								</td>
								<td>
									<PublishedToggleBadge id={project.slug} published={project.published} />
								</td>
								<td class="admin-table-cell-actions">
									<div class="admin-actions-wrap">
										<a href="/projects/{project.slug}" class="admin-action-btn" title="Редактировать" aria-label="Редактировать проект">
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
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</AdminPage>

<style>
	:global(.cat-badge) {
		margin-right: 0.25rem;
	}
	:global(.project-group-header td) {
		padding: 0.75rem 1rem 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #6b7280;
		background: #f9fafb;
		border-top: 1px solid #e5e7eb;
	}
	:global(.project-group-header--warning td) {
		color: #b45309;
		background: #fffbeb;
	}
	:global(.project-group-header--over td) {
		background: #eff6ff;
	}
</style>
