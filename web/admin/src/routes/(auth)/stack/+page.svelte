<script lang="ts">
	import { deserialize, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import PublishedToggleBadge from '$lib/components/PublishedToggleBadge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import type { Service, StackItem } from '$lib/types';
	import StackForm from './StackForm.svelte';

	const ORPHAN_GROUP_ID = '__orphan__';

	type StackGroup = { id: string; label: string; items: StackItem[] };

	let { data } = $props();

	let drawerOpen = $state(false);
	let editing = $state<StackItem | null>(null);
	let groupsOverride = $state<StackGroup[] | null>(null);
	let draggingId = $state<string | null>(null);
	let dragOver = $state<{ groupId: string; itemId: string | null } | null>(null);
	let reordering = $state(false);

	const baseGroups = $derived(buildGroups(data.services, data.stack));
	const groups = $derived(groupsOverride ?? baseGroups);

	$effect(() => {
		data.stack;
		data.services;
		groupsOverride = null;
	});

	function serviceTitle(service: Service): string {
		const langs = Object.keys(service.translations);
		return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
	}

	function buildGroups(services: Service[], stack: StackItem[]): StackGroup[] {
		const sortedServices = [...services].sort(
			(a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug)
		);
		const serviceSlugs = new Set(sortedServices.map((s) => s.slug));

		const result: StackGroup[] = sortedServices.map((service) => ({
			id: service.slug,
			label: serviceTitle(service),
			items: stack
				.filter((item) => item.group_id === service.slug)
				.sort((a, b) => a.sort_order - b.sort_order || a.label.localeCompare(b.label))
		}));

		const orphans = stack
			.filter((item) => !serviceSlugs.has(item.group_id))
			.sort((a, b) => a.sort_order - b.sort_order || a.label.localeCompare(b.label));
		if (orphans.length) {
			result.push({ id: ORPHAN_GROUP_ID, label: 'Без группы', items: orphans });
		}
		return result;
	}

	function openCreate() {
		editing = null;
		drawerOpen = true;
	}

	function openEdit(item: StackItem) {
		editing = item;
		drawerOpen = true;
	}

	function findItemIn(source: StackGroup[], itemId: string): { groupIdx: number; itemIdx: number } | null {
		for (let groupIdx = 0; groupIdx < source.length; groupIdx++) {
			const itemIdx = source[groupIdx].items.findIndex((item) => item.id === itemId);
			if (itemIdx >= 0) return { groupIdx, itemIdx };
		}
		return null;
	}

	function moveItem(fromId: string, toGroupId: string, beforeItemId: string | null): StackGroup[] | null {
		if (toGroupId === ORPHAN_GROUP_ID) return null;
		const source = groupsOverride ?? baseGroups;
		const from = findItemIn(source, fromId);
		if (!from) return null;
		const toGroupIdx = source.findIndex((group) => group.id === toGroupId);
		if (toGroupIdx < 0) return null;

		const next = source.map((group) => ({ ...group, items: [...group.items] }));
		const [item] = next[from.groupIdx].items.splice(from.itemIdx, 1);

		let insertIdx = beforeItemId
			? next[toGroupIdx].items.findIndex((entry) => entry.id === beforeItemId)
			: next[toGroupIdx].items.length;
		if (insertIdx < 0) insertIdx = next[toGroupIdx].items.length;
		if (from.groupIdx === toGroupIdx && from.itemIdx < insertIdx) insertIdx--;

		next[toGroupIdx].items.splice(insertIdx, 0, item);
		groupsOverride = next;
		return next;
	}

	function layoutPayload(source: StackGroup[]) {
		return source
			.filter((group) => group.id !== ORPHAN_GROUP_ID)
			.map((group) => ({ group_id: group.id, ids: group.items.map((item) => item.id) }));
	}

	async function persistLayout(nextGroups: StackGroup[]) {
		if (reordering) return;

		reordering = true;
		const previous = groupsOverride;
		try {
			const fd = new FormData();
			fd.set('layout', JSON.stringify(layoutPayload(nextGroups)));
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
				groupsOverride = null;
				await invalidateAll();
				return;
			}
			groupsOverride = previous;
			const message =
				result.type === 'failure'
					? ((result.data as { error?: string } | undefined)?.error ?? 'Не удалось сохранить порядок')
					: 'Не удалось сохранить порядок';
			toast.error(message);
		} catch {
			groupsOverride = previous;
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
		const next = moveItem(fromId, groupId, itemId);
		dragOver = null;
		draggingId = null;
		if (next) await persistLayout(next);
	}

	async function onGroupDrop(e: DragEvent, groupId: string) {
		e.preventDefault();
		const fromId = e.dataTransfer?.getData('text/plain') || draggingId;
		if (!fromId || groupId === ORPHAN_GROUP_ID) return;
		const next = moveItem(fromId, groupId, null);
		dragOver = null;
		draggingId = null;
		if (next) await persistLayout(next);
	}

	const hasItems = $derived(groups.some((group) => group.items.length > 0));
</script>

<svelte:head>
	<title>Стек — Piplos Admin</title>
</svelte:head>

<AdminPage title="Стек" breadcrumbs={[{ label: 'Стек' }]}>
	{#snippet actions()}
		<Button variant="success" onclick={openCreate} disabled={!data.services.length}>+ Новая технология</Button>
	{/snippet}

	{#if data.error}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">{data.error}</p>
		</div>
	{:else if !data.services.length}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">Сначала создайте услугу в разделе «Услуги».</p>
		</div>
	{:else if !hasItems}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">Стек пуст. Добавьте первую технологию.</p>
		</div>
	{:else}
		<div class="admin-table-wrap" class:admin-table-wrap--busy={reordering}>
			<table class="chart-table">
				<thead>
					<tr>
						<th class="admin-table-cell-drag" aria-label="Порядок"></th>
						<th>Технология</th>
						<th>Slug</th>
						<th>Статус</th>
						<th class="admin-table-cell-actions"></th>
					</tr>
				</thead>
				<tbody ondragover={(e) => e.preventDefault()}>
					{#each groups as group (group.id)}
						{#if group.items.length > 0 || group.id !== ORPHAN_GROUP_ID}
							<tr
								class="stack-group-header"
								class:stack-group-header--warning={group.id === ORPHAN_GROUP_ID}
								class:stack-group-header--over={dragOver?.groupId === group.id && dragOver.itemId === null}
								ondragover={(e) => onGroupDragOver(e, group.id)}
								ondragleave={() => {
									if (dragOver?.groupId === group.id && dragOver.itemId === null) dragOver = null;
								}}
								ondrop={(e) => void onGroupDrop(e, group.id)}
							>
								<td colspan="5">{group.label}</td>
							</tr>
						{/if}
						{#each group.items as item (item.id)}
							<tr
								class:admin-table-row--dragging={draggingId === item.id}
								class:admin-table-row--over={dragOver?.groupId === group.id && dragOver.itemId === item.id}
								ondragover={(e) => onRowDragOver(e, group.id, item.id)}
								ondragleave={() => {
									if (dragOver?.groupId === group.id && dragOver.itemId === item.id) dragOver = null;
								}}
								ondrop={(e) => void onRowDrop(e, group.id, item.id)}
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
											draggingId = item.id;
											e.dataTransfer?.setData('text/plain', item.id);
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
									<button type="button" class="admin-text-link row-link" onclick={() => openEdit(item)}>
										{item.label}
									</button>
								</td>
								<td class="chart-cell-muted">{item.slug}</td>
								<td>
									<PublishedToggleBadge
										id={item.id}
										published={item.published}
										publishedLabel="Виден"
										draftLabel="Скрыт"
									/>
								</td>
								<td class="admin-table-cell-actions">
									<div class="admin-actions-wrap">
										<button
											type="button"
											class="admin-action-btn"
											title="Редактировать"
											aria-label="Редактировать технологию"
											onclick={() => openEdit(item)}
										>
											<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
												<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
												<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
											</svg>
										</button>
										<form
											method="POST"
											action="?/delete"
											class="admin-action-form"
											use:enhance={deleteEnhance({
												message: `Удалить «${item.label}»?`,
												onSuccess: async () => {
													toast.success('Технология удалена');
													if (editing?.id === item.id) drawerOpen = false;
													await invalidateAll();
												},
												onError: () => toast.error('Не удалось удалить')
											})}
										>
											<input type="hidden" name="id" value={item.id} />
											<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить технологию">
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

<Drawer bind:open={drawerOpen} title={editing ? `Технология: ${editing.label}` : 'Новая технология'}>
	{#key editing?.id ?? 'new'}
		<StackForm item={editing ?? {}} onSaved={() => (drawerOpen = false)} />
	{/key}
</Drawer>

<style>
	.row-link {
		padding: 0;
		font-size: inherit;
		background: none;
		border: none;
		cursor: pointer;
	}
	:global(.stack-group-header td) {
		padding: 0.75rem 1rem 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #6b7280;
		background: #f9fafb;
		border-top: 1px solid #e5e7eb;
	}
	:global(.stack-group-header--warning td) {
		color: #b45309;
		background: #fffbeb;
	}
	:global(.stack-group-header--over td) {
		background: #eff6ff;
	}
</style>
