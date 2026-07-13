<script lang="ts">
	import { deserialize, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import PublishedToggleBadge from '$lib/components/PublishedToggleBadge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import TabPageActions from '$lib/components/TabPageActions.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import type { Service } from '$lib/types';
	import ServiceForm from './ServiceForm.svelte';

	let { data } = $props();

	let drawerOpen = $state(false);
	let editing = $state<Service | null>(null);
	let services = $state<Service[]>([]);
	let draggingId = $state<string | null>(null);
	let dragOverId = $state<string | null>(null);
	let reordering = $state(false);

	$effect(() => {
		services = [...data.services].sort(
			(a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug)
		);
	});

	function openCreate() {
		editing = null;
		drawerOpen = true;
	}

	function openEdit(service: Service) {
		editing = service;
		drawerOpen = true;
	}

	function title(service: Service): string {
		const langs = Object.keys(service.translations);
		return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
	}

	function moveItem(fromId: string, toId: string) {
		const fromIdx = services.findIndex((s) => s.id === fromId);
		const toIdx = services.findIndex((s) => s.id === toId);
		if (fromIdx < 0 || toIdx < 0 || fromIdx === toIdx) return;
		const next = [...services];
		const [item] = next.splice(fromIdx, 1);
		next.splice(toIdx, 0, item);
		services = next;
	}

	async function persistOrder() {
		if (reordering) return;
		reordering = true;
		const previous = [...services];
		try {
			const fd = new FormData();
			fd.set('order', JSON.stringify(services.map((s) => s.id)));
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
			services = previous;
			const message =
				result.type === 'failure'
					? ((result.data as { error?: string } | undefined)?.error ?? 'Не удалось сохранить порядок')
					: 'Не удалось сохранить порядок';
			toast.error(message);
		} catch {
			services = previous;
			toast.error('Не удалось сохранить порядок');
		} finally {
			reordering = false;
		}
	}

	function onRowDragOver(e: DragEvent, serviceId: string) {
		e.preventDefault();
		if (draggingId && draggingId !== serviceId) dragOverId = serviceId;
	}

	async function onRowDrop(e: DragEvent, serviceId: string) {
		e.preventDefault();
		const fromId = e.dataTransfer?.getData('text/plain') || draggingId;
		if (!fromId || fromId === serviceId) return;
		moveItem(fromId, serviceId);
		dragOverId = null;
		draggingId = null;
		await persistOrder();
	}
</script>

<svelte:head>
	<title>Списки: Услуги — Piplos Admin</title>
</svelte:head>

<TabPageActions>
	<Button variant="success" onclick={openCreate}>+ Новая услуга</Button>
</TabPageActions>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else if !services.length}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">Услуг пока нет. Создайте первую.</p>
	</div>
{:else}
	<div class="admin-table-wrap" class:admin-table-wrap--busy={reordering}>
		<table class="chart-table">
			<thead>
				<tr>
					<th class="admin-table-cell-drag" aria-label="Порядок"></th>
					<th>Название</th>
					<th>Slug</th>
					<th>Стек</th>
					<th>Языки</th>
					<th>Статус</th>
					<th class="admin-table-cell-actions"></th>
				</tr>
			</thead>
			<tbody
				ondragover={(e) => e.preventDefault()}
			>
				{#each services as service (service.id)}
					<tr
						class:admin-table-row--dragging={draggingId === service.id}
						class:admin-table-row--over={dragOverId === service.id}
						ondragover={(e) => onRowDragOver(e, service.id)}
						ondragleave={() => {
							if (dragOverId === service.id) dragOverId = null;
						}}
						ondrop={(e) => void onRowDrop(e, service.id)}
					>
						<td class="admin-table-cell-drag">
							<button
								type="button"
								class="admin-drag-handle"
								draggable="true"
								title="Перетащите для сортировки"
								aria-label="Перетащите для изменения порядка"
								disabled={reordering}
								ondragstart={(e) => {
									draggingId = service.id;
									e.dataTransfer?.setData('text/plain', service.id);
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
							{#if service.icon}<span class="svc-icon" aria-hidden="true">{service.icon}</span>{/if}
							<button type="button" class="admin-text-link row-link" onclick={() => openEdit(service)}>
								{title(service)}
							</button>
						</td>
						<td class="chart-cell-muted">{service.slug}</td>
						<td>
							{#each service.tags as t (t)}
								<Badge variant="neutral" class="cat-badge">{t}</Badge>
							{/each}
						</td>
						<td>
							{#each Object.keys(service.translations) as lang (lang)}
								<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
							{/each}
						</td>
						<td>
							<PublishedToggleBadge
								id={service.id}
								published={service.published}
								publishedLabel="Опубликована"
							/>
						</td>
						<td class="admin-table-cell-actions">
							<div class="admin-actions-wrap">
								<button
									type="button"
									class="admin-action-btn"
									title="Редактировать"
									aria-label="Редактировать услугу"
									onclick={() => openEdit(service)}
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
										message: 'Удалить услугу?',
										onSuccess: async () => {
											toast.success('Услуга удалена');
											await invalidateAll();
										},
										onError: () => toast.error('Не удалось удалить услугу')
									})}
								>
									<input type="hidden" name="id" value={service.id} />
									<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить услугу">
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

<Drawer bind:open={drawerOpen} title={editing ? `Услуга: ${title(editing)}` : 'Новая услуга'}>
	{#key editing?.id ?? 'new'}
		<ServiceForm
			service={editing ?? {}}
			languages={data.languages}
			stack={data.stack}
			onSaved={() => (drawerOpen = false)}
		/>
	{/key}
</Drawer>

<style>
	.svc-icon {
		margin-right: 0.375rem;
	}
	.row-link {
		padding: 0;
		font-size: inherit;
		background: none;
		border: none;
		cursor: pointer;
	}
	:global(.cat-badge) {
		margin-right: 0.25rem;
	}
</style>
