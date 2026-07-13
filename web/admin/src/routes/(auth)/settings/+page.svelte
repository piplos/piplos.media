<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Drawer from '$lib/components/Drawer.svelte';
	import TabPageActions from '$lib/components/TabPageActions.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import type { Language } from '$lib/types';
	import LanguageForm from './LanguageForm.svelte';

	let { data } = $props();

	let drawerOpen = $state(false);
	let editing = $state<Language | null>(null);

	function openCreate() {
		editing = null;
		drawerOpen = true;
	}

	function openEdit(lang: Language) {
		editing = lang;
		drawerOpen = true;
	}
</script>

<svelte:head>
	<title>Общие — Настройки — Piplos Admin</title>
</svelte:head>

<TabPageActions>
	<Button variant="success" onclick={openCreate}>+ Новый язык</Button>
</TabPageActions>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else if !data.languages.length}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">Языков пока нет. Добавьте первый.</p>
	</div>
{:else}
	<div class="admin-table-wrap">
		<table class="chart-table">
			<thead>
				<tr>
					<th>Код</th>
					<th>Название</th>
					<th>Статус</th>
					<th class="admin-table-cell-actions"></th>
				</tr>
			</thead>
			<tbody>
				{#each data.languages as lang (lang.code)}
					<tr>
						<td class="chart-cell-main">
							<button type="button" class="admin-text-link row-link" onclick={() => openEdit(lang)}>
								{lang.code.toUpperCase()}
							</button>
						</td>
						<td>
							<button type="button" class="admin-text-link row-link" onclick={() => openEdit(lang)}>
								{lang.name}
							</button>
						</td>
						<td>
							{#if lang.is_default}
								<Badge variant="warning" pill>По умолчанию</Badge>
							{/if}
							{#if lang.enabled}
								<Badge variant="success" pill>Включён</Badge>
							{:else}
								<Badge variant="neutral" pill>Выключен</Badge>
							{/if}
						</td>
						<td class="admin-table-cell-actions">
							<div class="admin-actions-wrap">
								<button
									type="button"
									class="admin-action-btn"
									title="Редактировать"
									aria-label="Редактировать язык"
									onclick={() => openEdit(lang)}
								>
									<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
									</svg>
								</button>
								{#if !lang.is_default}
									<form
										method="POST"
										action="?/deleteLanguage"
										class="admin-action-form"
										use:enhance={deleteEnhance({
											message: `Удалить язык ${lang.name}?`,
											onSuccess: async () => {
												toast.success('Язык удалён');
												if (editing?.code === lang.code) drawerOpen = false;
												await invalidateAll();
											},
											onError: (message) => toast.error(message)
										})}
									>
										<input type="hidden" name="code" value={lang.code} />
										<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить язык">
											<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
												<polyline points="3 6 5 6 21 6" />
												<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
											</svg>
										</button>
									</form>
								{/if}
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}

<Drawer bind:open={drawerOpen} title={editing ? `Язык: ${editing.name}` : 'Новый язык'}>
	{#key editing?.code ?? 'new'}
		<LanguageForm language={editing ?? {}} onSaved={() => (drawerOpen = false)} />
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
</style>
