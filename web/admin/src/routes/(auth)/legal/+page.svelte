<script lang="ts">
	import AdminPage from '$lib/components/AdminPage.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import { LEGAL_SLUG_LABELS, type LegalPage } from '$lib/types';

	let { data } = $props();

	function title(page: LegalPage): string {
		const langs = Object.keys(page.translations);
		const en = page.translations['en']?.title;
		const first = langs.length ? page.translations[langs[0]]?.title : '';
		return en ?? first ?? LEGAL_SLUG_LABELS[page.slug] ?? page.slug;
	}
</script>

<svelte:head>
	<title>Правовое — Piplos Admin</title>
</svelte:head>

<AdminPage title="Правовое" breadcrumbs={[{ label: 'Правовое' }]}>
	{#if data.error}
		<div class="admin-table-wrap admin-table-wrap--empty">
			<p class="text-muted">{data.error}</p>
		</div>
	{:else if !data.pages.length}
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
					{#each data.pages as page (page.id)}
						<tr>
							<td class="chart-cell-main">
								<a href="/legal/{page.id}" class="admin-text-link">
									{LEGAL_SLUG_LABELS[page.slug] ?? page.slug}
								</a>
							</td>
							<td class="chart-cell-muted">{page.path}</td>
							<td class="chart-cell-muted">{title(page)}</td>
							<td>
								{#each Object.keys(page.translations) as lang (lang)}
									<Badge variant={lang} class="cat-badge">{lang.toUpperCase()}</Badge>
								{/each}
							</td>
							<td class="admin-table-cell-actions">
								<a href="/legal/{page.id}" class="admin-action-btn" title="Редактировать" aria-label="Редактировать документ">
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
</AdminPage>

<style>
	:global(.cat-badge) {
		margin-right: 0.25rem;
	}
</style>
