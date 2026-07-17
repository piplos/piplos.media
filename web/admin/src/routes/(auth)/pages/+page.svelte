<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import { deleteEnhance } from '$lib/delete-enhance';
	import { formatDate } from '$lib/format';
	import {
		LEGAL_SLUG_LABELS,
		PAGE_STATUS_LABELS,
		PAGE_STATUS_VARIANTS,
		pageStatus,
		type LegalPage,
		type Page
	} from '$lib/types';

	let { data } = $props();

	function pageTitle(page: Page): string {
		const langs = Object.keys(page.translations);
		const en = page.translations['en']?.title;
		const first = langs.length ? page.translations[langs[0]]?.title : '';
		return en || first || page.slug;
	}

	function legalTitle(page: LegalPage): string {
		const langs = Object.keys(page.translations);
		const en = page.translations['en']?.title;
		const first = langs.length ? page.translations[langs[0]]?.title : '';
		return en ?? first ?? LEGAL_SLUG_LABELS[page.slug] ?? page.slug;
	}
</script>

<svelte:head>
	<title>Страницы — Piplos Admin</title>
</svelte:head>

<AdminPage title="Страницы" breadcrumbs={[{ label: 'Страницы' }]}>
	{#snippet actions()}
		<Button variant="success" onclick={() => (location.href = '/pages/new')}>+ Новая страница</Button>
	{/snippet}

	{#if data.error}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">{data.error}</p>
		</div>
	{:else}
		{#if !data.pages.length}
			<div class="admin-table-wrap admin-table-wrap--empty">
				<p class="text-muted">
					Страниц пока нет. <a href="/pages/new" class="admin-text-link">Создайте первую</a> — она
					появится в разделе «Статьи» на сайте.
				</p>
			</div>
		{:else}
			<div class="admin-table-wrap">
				<table class="chart-table">
					<thead>
						<tr>
							<th>Название</th>
							<th>Путь</th>
							<th>Языки</th>
							<th>Публикация</th>
							<th>Статус</th>
							<th class="admin-table-cell-actions"></th>
						</tr>
					</thead>
					<tbody>
						{#each data.pages as page (page.id)}
							{@const status = pageStatus(page)}
							<tr>
								<td class="chart-cell-main">
									<a href="/pages/{page.id}" class="admin-text-link">{pageTitle(page)}</a>
								</td>
								<td class="chart-cell-muted">/{'{lang}'}/articles/{page.slug}</td>
								<td>
									{#each Object.keys(page.translations) as lang (lang)}
										<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
									{/each}
								</td>
								<td class="chart-cell-muted">
									{page.publish_at ? formatDate(page.publish_at) : 'Сразу'}
								</td>
								<td>
									<form
										method="POST"
										action="?/togglePublished"
										class="status-toggle-form"
										use:enhance={() => {
											return async ({ result }) => {
												if (result.type === 'success') await invalidateAll();
											};
										}}
									>
										<input type="hidden" name="id" value={page.id} />
										<button
											type="submit"
											class="status-toggle-btn"
											title={page.published ? 'Снять с публикации' : 'Опубликовать'}
											aria-label={page.published ? 'Снять с публикации' : 'Опубликовать'}
										>
											<Badge variant={PAGE_STATUS_VARIANTS[status]} pill>
												{PAGE_STATUS_LABELS[status]}
											</Badge>
										</button>
									</form>
								</td>
								<td class="admin-table-cell-actions">
									<div class="admin-actions-wrap">
										<a href="/pages/{page.id}" class="admin-action-btn" title="Редактировать" aria-label="Редактировать страницу">
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
												message: 'Удалить страницу?',
												onSuccess: async () => {
													toast.success('Страница удалена');
													await invalidateAll();
												},
												onError: () => toast.error('Не удалось удалить страницу')
											})}
										>
											<input type="hidden" name="id" value={page.id} />
											<button type="submit" class="admin-action-btn" title="Удалить" aria-label="Удалить страницу">
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

		<h2 class="pages-section-title">Правовые документы</h2>
		<p class="pages-section-hint">
			Системные страницы: их нельзя удалить или переименовать — только редактировать содержимое.
		</p>
		{#if !data.legalPages.length}
			<div class="admin-table-wrap admin-table-wrap--empty">
				<p class="text-muted">Правовые документы не найдены. Примените миграции БД.</p>
			</div>
		{:else}
			<div class="admin-table-wrap">
				<table class="chart-table">
					<thead>
						<tr>
							<th>Документ</th>
							<th>Путь</th>
							<th>Заголовок</th>
							<th>Языки</th>
							<th class="admin-table-cell-actions"></th>
						</tr>
					</thead>
					<tbody>
						{#each data.legalPages as page (page.id)}
							<tr>
								<td class="chart-cell-main">
									<a href="/pages/legal/{page.id}" class="admin-text-link">
										{LEGAL_SLUG_LABELS[page.slug] ?? page.slug}
									</a>
								</td>
								<td class="chart-cell-muted">/{'{lang}'}/legal/{page.slug}</td>
								<td class="chart-cell-muted">{legalTitle(page)}</td>
								<td>
									{#each Object.keys(page.translations) as lang (lang)}
										<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
									{/each}
								</td>
								<td class="admin-table-cell-actions">
									<a href="/pages/legal/{page.id}" class="admin-action-btn" title="Редактировать" aria-label="Редактировать документ">
										<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
											<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
											<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
										</svg>
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}
</AdminPage>

<style>
	:global(.cat-badge) {
		margin-right: 0.25rem;
	}
	.pages-section-title {
		margin: 1rem 0 0;
		font-size: 1.0625rem;
		font-weight: 600;
		color: #18181b;
	}
	.pages-section-hint {
		margin: -0.5rem 0 0;
		font-size: 0.8125rem;
		color: #71717a;
	}
	.status-toggle-form {
		display: inline;
		margin: 0;
		padding: 0;
	}
	.status-toggle-btn {
		display: inline-flex;
		margin: 0;
		padding: 0;
		border: none;
		background: none;
		cursor: pointer;
		border-radius: 9999px;
		transition: opacity 0.15s;
	}
	.status-toggle-btn:hover {
		opacity: 0.8;
	}
	.status-toggle-btn:focus-visible {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}
</style>
