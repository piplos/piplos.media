<script lang="ts">
	import { enhance } from '$app/forms';
	import type { SubmitFunction } from '@sveltejs/kit';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import { confirmAction } from '$lib/confirm.svelte';
	import type { ImageFileRef, ImageRefUsage } from './+page.server';

	let { data } = $props();

	const files = $derived(data.files as ImageFileRef[]);
	const replaceable = $derived(files.filter((f) => f.webp_exists));
	const missingWebP = $derived(files.length - replaceable.length);

	// Кнопка замены конкретного файла, отправляющая форму его action'ом.
	// replacing блокирует обе формы на время любой замены: параллельные
	// запросы перезаписывали бы строки снимками разного возраста.
	let replacing = $state(false);
	let replacingPath = $state<string | null>(null);

	const ENTITY_LABELS: Record<string, string> = {
		project: 'Проект',
		service: 'Услуга',
		stack: 'Стек',
		seo: 'SEO',
		legal: 'Документ',
		page: 'Страница'
	};

	function entityLabel(entity: string) {
		return ENTITY_LABELS[entity] ?? entity;
	}

	function fileTitle(refPath: string) {
		const name = refPath.split('/').pop() ?? refPath;
		return name.replace(/\.(png|jpe?g)$/i, '');
	}

	function usageTitle(usage: ImageRefUsage) {
		const base = `${entityLabel(usage.entity)} · ${usage.label || usage.id}`;
		return usage.href ? `${base} — открыть в новой вкладке` : base;
	}

	// Одна сущность = один чип: поля использования собираются в подсказку.
	function groupedUsages(file: ImageFileRef) {
		const groups = new Map<string, { usage: ImageRefUsage; fields: string[] }>();
		for (const usage of file.usages) {
			const key = `${usage.entity} ${usage.id}`;
			const group = groups.get(key);
			if (group) group.fields.push(usage.field);
			else groups.set(key, { usage, fields: [usage.field] });
		}
		return [...groups.values()];
	}

	// Как deleteEnhance, но с кнопкой подтверждения «Заменить» и общим
	// для обеих форм пост-обработкой (тост + перезагрузка данных).
	function replaceEnhance(message: string, onStart?: () => void): SubmitFunction {
		return async ({ cancel }) => {
			if (
				!(await confirmAction({
					message,
					confirmLabel: 'Заменить',
					cancelLabel: 'Отмена'
				}))
			) {
				cancel();
				return;
			}
			onStart?.();
			replacing = true;
			return async ({ result }) => {
				if (result.type === 'success') {
					const updated = Number((result.data as { updated?: number } | undefined)?.updated ?? 0);
					toast.success(
						updated === 1 ? 'Готово: обновлена 1 запись' : `Готово: обновлено записей ${updated}`
					);
					replacingPath = null;
					await invalidateAll();
				} else if (result.type === 'failure') {
					replacingPath = null;
					toast.error((result.data?.error as string) ?? 'Не удалось заменить ссылки');
				} else {
					replacingPath = null;
				}
				replacing = false;
			};
		};
	}
</script>

<svelte:head>
	<title>Поиск PNG/JPG — Настройки — Piplos Admin</title>
</svelte:head>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	<Card padding="sm">
		<div class="search-head">
			<div>
				<h2 class="section-title">Поиск PNG/JPG</h2>
				<p class="section-desc">
					Ищет вставленные в тексты и формы PNG- и JPG-файлы из медиатеки. Если для файла уже
					есть WebP-версия, ссылку можно заменить одним нажатием. Для файлов без WebP сначала
					запустите пересборку на вкладке «WebP-варианты».
				</p>
			</div>
			{#if replaceable.length > 0}
				<form
					method="POST"
					action="?/replaceAll"
					use:enhance={replaceEnhance(
						`Заменить все найденные ссылки (${replaceable.length} файлов) на WebP?`
					)}
				>
					<Button type="submit" variant="success" disabled={replacing}>
						Заменить все ({replaceable.length})
					</Button>
				</form>
			{/if}
		</div>

		{#if !files.length}
			<div class="empty-state">
				<p class="text-muted">PNG/JPG-ссылки в контенте не найдены — всё уже в WebP.</p>
			</div>
		{:else}
			<dl class="search-stats">
				<div>
					<dt>Найдено файлов</dt>
					<dd>{files.length}</dd>
				</div>
				<div>
					<dt>Есть WebP</dt>
					<dd>{replaceable.length}</dd>
				</div>
				<div>
					<dt>Без WebP</dt>
					<dd>{missingWebP}</dd>
				</div>
			</dl>

			<div class="table-wrap">
				<table class="image-table">
					<thead>
						<tr>
							<th>Файл</th>
							<th>Используется в</th>
							<th>WebP</th>
							<th class="cell-actions"></th>
						</tr>
					</thead>
					<tbody>
						{#each files as file (file.path)}
							<tr>
								<td class="cell-file">
									<span class="file-name">{fileTitle(file.path)}</span>
									<span class="file-path" title={file.path}>{file.path}</span>
								</td>
								<td>
									<div class="usage-list">
										{#each groupedUsages(file) as group (group.usage.entity + group.usage.id)}
											{#if group.usage.href}
												<a
													class="usage-chip"
													href={group.usage.href}
													target="_blank"
													rel="noopener noreferrer"
													title={`${usageTitle(group.usage)} · поля: ${group.fields.join(', ')}`}
												>
													{entityLabel(group.usage.entity)}: {group.usage.label || group.usage.id}
													<svg
														width="10"
														height="10"
														viewBox="0 0 24 24"
														fill="none"
														stroke="currentColor"
														stroke-width="2.5"
														stroke-linecap="round"
														stroke-linejoin="round"
														aria-hidden="true"
													>
														<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
														<polyline points="15 3 21 3 21 9" />
														<line x1="10" y1="14" x2="21" y2="3" />
													</svg>
												</a>
											{:else}
												<span
													class="usage-chip"
													title={`${usageTitle(group.usage)} · поля: ${group.fields.join(', ')}`}
												>
													{entityLabel(group.usage.entity)}: {group.usage.label || group.usage.id}
												</span>
											{/if}
										{/each}
									</div>
								</td>
								<td class="cell-webp">
									{#if file.webp_exists}
										<span class="webp-icon webp-icon--ok" title="WebP-версия на диске">
											<svg
												width="16"
												height="16"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
												aria-hidden="true"
											>
												<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
												<polyline points="22 4 12 14.01 9 11.01" />
											</svg>
										</span>
									{:else}
										<span
											class="webp-icon webp-icon--missing"
											title="WebP-версия не найдена — запустите пересборку на вкладке «WebP-варианты»"
										>
											<svg
												width="16"
												height="16"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
												aria-hidden="true"
											>
												<path
													d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
												/>
												<line x1="12" y1="9" x2="12" y2="13" />
												<line x1="12" y1="17" x2="12.01" y2="17" />
											</svg>
										</span>
									{/if}
								</td>
								<td class="cell-actions">
									{#if file.webp_exists}
										<form
											method="POST"
											action="?/replace"
											use:enhance={replaceEnhance(
												`Заменить все ссылки на ${file.path} на WebP?`,
												() => (replacingPath = file.path)
											)}
										>
											<input type="hidden" name="path" value={file.path} />
											<button
												type="submit"
												class="icon-btn image-replace-btn"
												title="Заменить ссылки на WebP"
												aria-label={`Заменить ${file.path} на WebP`}
												disabled={replacing}
											>
												{#if replacingPath === file.path}
													<span class="image-replace-spinner" aria-hidden="true"></span>
												{:else}
													<svg
														width="15"
														height="15"
														viewBox="0 0 24 24"
														fill="none"
														stroke="currentColor"
														stroke-width="2"
														stroke-linecap="round"
														stroke-linejoin="round"
														aria-hidden="true"
													>
														<polyline points="17 1 21 5 17 9" />
														<path d="M3 11V9a4 4 0 0 1 4-4h14" />
														<polyline points="7 23 3 19 7 15" />
														<path d="M21 13v2a4 4 0 0 1-4 4H3" />
													</svg>
												{/if}
											</button>
										</form>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</Card>
{/if}

<style>
	.search-head {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}
	.section-title {
		margin: 0 0 0.35rem;
		font-size: 1rem;
		font-weight: 600;
	}
	.section-desc {
		margin: 0;
		max-width: 40rem;
		font-size: 0.875rem;
		line-height: 1.45;
		color: #71717a;
	}
	.empty-state {
		padding: 0.5rem 0 0.25rem;
	}
	.search-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
		gap: 0.75rem 1.25rem;
		margin: 0 0 1.25rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}
	.search-stats dt {
		margin: 0 0 0.2rem;
		font-size: 0.75rem;
		color: #a1a1aa;
		text-transform: uppercase;
		letter-spacing: 0.02em;
	}
	.search-stats dd {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
	}

	/* Таблица — в стиле таблицы моделей на /settings/ai. */
	.table-wrap {
		overflow-x: auto;
	}
	.image-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}
	.image-table th {
		text-align: left;
		font-weight: 600;
		color: #374151;
		padding: 0.4rem 0.625rem;
		border-bottom: 1px solid #e5e7eb;
		white-space: nowrap;
	}
	.image-table td {
		padding: 0.4rem 0.625rem;
		border-bottom: 1px solid #f3f4f6;
		vertical-align: middle;
	}
	.image-table tr:last-child td {
		border-bottom: none;
	}
	.cell-file {
		max-width: 18rem;
	}
	.file-name {
		display: block;
		font-weight: 500;
		color: #18181b;
	}
	/* Путь в одну строку с многоточием — не раздвигает соседние колонки. */
	.file-path {
		display: block;
		font-size: 0.75rem;
		color: #a1a1aa;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.usage-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}
	/* Чип в палитре Badge--neutral; ссылка открывается в новой вкладке. */
	.usage-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.75rem;
		line-height: 1.25;
		letter-spacing: 0.02em;
		padding: 0.25rem 0.5rem;
		border-radius: 6px;
		background: #f4f4f5;
		color: #71717a;
		text-decoration: none;
	}
	a.usage-chip {
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	a.usage-chip:hover {
		background: #e4e4e7;
		color: #18181b;
	}
	.cell-webp {
		white-space: nowrap;
	}
	.webp-icon {
		display: inline-flex;
		vertical-align: middle;
	}
	.webp-icon--ok {
		color: #16a34a;
	}
	.webp-icon--missing {
		color: #ca8a04;
		cursor: help;
	}
	/* Иконная кнопка — геометрия models-action-btn со страницы AI. */
	.icon-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.75rem;
		height: 1.75rem;
		padding: 0;
		border: none;
		background: none;
		border-radius: 5px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.image-replace-btn {
		color: #16a34a;
	}
	.image-replace-btn:hover:not(:disabled) {
		color: #15803d;
		background: #dcfce7;
	}
	.image-replace-btn:disabled {
		opacity: 0.6;
		cursor: wait;
	}
	.image-replace-spinner {
		display: inline-block;
		width: 15px;
		height: 15px;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		animation: image-replace-spin 0.7s linear infinite;
	}
	@keyframes image-replace-spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
