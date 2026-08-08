<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Badge from '$lib/components/Badge.svelte';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import { formatDate } from '$lib/format';
	import type { RebuildVariantsStatus } from './+page.server';

	let { data } = $props();

	let starting = $state(false);

	const status = $derived(data.status as RebuildVariantsStatus);
	const running = $derived(status.running);

	// Поллинг, пока идёт фоновая пересборка (как у бекапов).
	$effect(() => {
		if (!running) return;
		const timer = setInterval(async () => {
			try {
				const res = await fetch('/api/uploads/rebuild-variants');
				if (!res.ok) return;
				const payload = (await res.json()) as { status?: RebuildVariantsStatus };
				const next = payload.status;
				if (!next || next.running) return;
				clearInterval(timer);
				if (next.error) {
					toast.error(next.error);
				} else {
					toast.success(`Готово: обработано ${next.ok}, ошибок ${next.failed}`);
				}
				await invalidateAll();
			} catch {
				// Сетевые сбои поллинга игнорируем.
			}
		}, 2500);
		return () => clearInterval(timer);
	});
</script>

<svelte:head>
	<title>Медиа — Настройки — Piplos Admin</title>
</svelte:head>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	{#if running}
		<div class="status-banner">
			<span class="status-spinner" aria-hidden="true"></span>
			Идёт пересборка WebP-вариантов…
			{#if status.started_at}
				<span class="status-meta">старт {formatDate(status.started_at)}</span>
			{/if}
		</div>
	{:else if status.error}
		<div class="status-banner status-banner--error">
			Последняя пересборка завершилась с ошибкой: {status.error}
		</div>
	{:else if status.finished_at}
		<div class="status-banner status-banner--ok">
			Последняя пересборка: обработано {status.ok}, ошибок {status.failed}
			<span class="status-meta">{formatDate(status.finished_at)}</span>
		</div>
	{/if}

	<Card padding="sm">
		<div class="media-head">
			<div>
				<h2 class="section-title">WebP-варианты</h2>
				<p class="section-desc">
					Пересобирает full / 480 / 960 WebP для загруженных PNG и JPEG в архиве.
					Оригиналы после успешной конвертации удаляются. Новые загрузки обрабатываются
					автоматически — эта кнопка нужна для старых файлов после деплоя.
				</p>
			</div>
			<form
				method="POST"
				action="?/rebuild"
				use:enhance={() => {
					starting = true;
					return async ({ result }) => {
						starting = false;
						if (result.type === 'success') {
							toast.success('Пересборка запущена');
							await invalidateAll();
						} else if (result.type === 'failure') {
							toast.error((result.data?.error as string) ?? 'Не удалось запустить');
						}
					};
				}}
			>
				<Button type="submit" variant="success" loading={starting} disabled={running || starting}>
					Пересобрать WebP
				</Button>
			</form>
		</div>

		<dl class="media-stats">
			<div>
				<dt>Статус</dt>
				<dd>
					{#if running}
						<Badge variant="warning" pill>В процессе</Badge>
					{:else}
						<Badge variant="neutral" pill>Ожидание</Badge>
					{/if}
				</dd>
			</div>
			<div>
				<dt>Обработано</dt>
				<dd>{status.ok}</dd>
			</div>
			<div>
				<dt>Ошибок</dt>
				<dd>{status.failed}</dd>
			</div>
		</dl>
	</Card>
{/if}

<style>
	.status-banner {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem 0.75rem;
		padding: 0.75rem 1rem;
		margin-bottom: 1rem;
		border-radius: 10px;
		background: #eff6ff;
		color: #1e3a8a;
		font-size: 0.875rem;
	}
	.status-banner--error {
		background: #fef2f2;
		color: #991b1b;
	}
	.status-banner--ok {
		background: #f0fdf4;
		color: #166534;
	}
	.status-meta {
		opacity: 0.75;
		font-size: 0.8125rem;
	}
	.status-spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	.media-head {
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
	.media-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
		gap: 0.75rem 1.25rem;
		margin: 0;
		padding-top: 1rem;
		border-top: 1px solid #e4e4e7;
	}
	.media-stats dt {
		margin: 0 0 0.2rem;
		font-size: 0.75rem;
		color: #a1a1aa;
		text-transform: uppercase;
		letter-spacing: 0.02em;
	}
	.media-stats dd {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
	}
</style>
