<script lang="ts">
	import { enhance } from '$app/forms';
	import { applyAction } from '$app/forms';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Button from '$lib/components/Button.svelte';
	import toast from 'svelte-french-toast';
	import type { AIProviderModel } from './_types';

	interface Props {
		title: string;
		enable: boolean;
		apiKey: string;
		apiKeyDirty: boolean;
		rateLimit: string;
		timeoutSeconds: string;
		apiKeyPlaceholder: string;
		updateAction: string;
		testAction: string;
		noReload: () => (opts: { update: (o?: { invalidateAll?: boolean; reset?: boolean }) => Promise<void> }) => Promise<void>;
		idPrefix: string;
		usedForTranslation?: boolean;
		models?: AIProviderModel[];
		providerKey?: string;
	}

	let {
		title,
		enable = $bindable(),
		apiKey = $bindable(),
		apiKeyDirty = $bindable(),
		rateLimit = $bindable(),
		timeoutSeconds = $bindable(),
		apiKeyPlaceholder,
		updateAction,
		testAction,
		noReload,
		idPrefix,
		usedForTranslation = false,
		models = undefined,
		providerKey = ''
	}: Props = $props();

	const cannotDisable = $derived(usedForTranslation);
	const usageMessage = $derived(usedForTranslation ? 'Используется для перевода' : null);

	let saveSubmitting = $state(false);
	let testSubmitting = $state(false);

	function enhanceSave() {
		saveSubmitting = true;
		const run = noReload();
		return (opts: { update: (o?: { invalidateAll?: boolean; reset?: boolean }) => Promise<void> }) => {
			return run(opts).finally(() => {
				saveSubmitting = false;
			});
		};
	}
	function enhanceTest() {
		testSubmitting = true;
		return async ({ result }: { result: Parameters<typeof applyAction>[0] }) => {
			try {
				await applyAction(result);
			} finally {
				testSubmitting = false;
			}
		};
	}

	// Models section state
	let newModelId = $state('');
	let newDisplayName = $state('');
	let addingModel = $state(false);
	let editingModelId = $state<string | null>(null);
	let editDisplayName = $state('');
	let deletingModelId = $state<string | null>(null);

	type ModelActionResult = {
		createSuccess?: boolean;
		createError?: string;
		updateSuccess?: boolean;
		updateError?: string;
		deleteSuccess?: boolean;
		deleteError?: string;
	};

	function enhanceModel() {
		return async ({
			result,
			update
		}: {
			result: { type: string; data?: unknown };
			update: (opts?: { invalidateAll?: boolean; reset?: boolean }) => Promise<void>;
		}) => {
			if (result.type === 'success') {
				const data = result.data as ModelActionResult;
				if (data?.createSuccess) {
					toast.success('Модель добавлена');
					newModelId = '';
					newDisplayName = '';
					addingModel = false;
				} else if (data?.updateSuccess) {
					toast.success('Модель обновлена');
					editingModelId = null;
				} else if (data?.deleteSuccess) {
					toast.success('Модель удалена');
					deletingModelId = null;
				}
			} else if (result.type === 'failure') {
				const data = result.data as ModelActionResult;
				if (data?.createError) toast.error(data.createError);
				else if (data?.updateError) toast.error(data.updateError);
				else if (data?.deleteError) toast.error(data.deleteError);
			}
			await update({ invalidateAll: true, reset: false });
		};
	}

	function startEditModel(m: AIProviderModel) {
		editingModelId = m.id;
		editDisplayName = m.display_name;
	}
	function cancelEditModel() {
		editingModelId = null;
	}
</script>

<Card class="model-card">
	<div class="model-card-header">
		<h2 class="model-card-title">{title}</h2>
		<div class="model-switch-block">
			<label class="model-switch-label">
				<input
					type="checkbox"
					name="enable"
					class="model-switch"
					bind:checked={enable}
					disabled={cannotDisable}
				/>
				<span>Включить модель</span>
			</label>
			{#if usageMessage}
				<p class="model-usage-msg" role="status">{usageMessage}</p>
			{/if}
		</div>
	</div>
	<div class="model-form" class:model-form--disabled={!enable}>
		<div class="model-api-key-wrap" oninput={() => (apiKeyDirty = true)} role="presentation">
			<FormField label="API ключ" id="{idPrefix}-apiKey">
				<Input
					id="{idPrefix}-apiKey"
					name="apiKey"
					type="password"
					bind:value={apiKey}
					placeholder={apiKeyPlaceholder}
					autocomplete="off"
				/>
			</FormField>
		</div>
		<div class="model-fields-row">
			<FormField label="Лимит запросов (rate limit)" id="{idPrefix}-rateLimit">
				<Input id="{idPrefix}-rateLimit" name="rateLimit" type="number" bind:value={rateLimit} />
			</FormField>
			<FormField label="Таймаут (сек)" id="{idPrefix}-timeoutSeconds">
				<Input
					id="{idPrefix}-timeoutSeconds"
					name="timeoutSeconds"
					type="number"
					bind:value={timeoutSeconds}
				/>
			</FormField>
		</div>
	</div>
	<div class="model-actions-row" class:model-actions-row--disabled={!enable}>
		<form method="POST" action={updateAction} class="model-actions-form" use:enhance={enhanceSave}>
			<input type="hidden" name="apiKeyDirty" value={apiKeyDirty ? 'true' : 'false'} />
			<input type="hidden" name="enable" value={enable ? 'on' : 'off'} />
			<input type="hidden" name="apiKey" value={apiKey} />
			<input type="hidden" name="rateLimit" value={rateLimit} />
			<input type="hidden" name="timeoutSeconds" value={timeoutSeconds} />
			<Button type="submit" loading={saveSubmitting}>Сохранить</Button>
		</form>
		<form method="POST" action={testAction} class="model-test-form" use:enhance={enhanceTest}>
			<input type="hidden" name="apiKey" value={apiKey} />
			<Button type="submit" variant="secondary" loading={testSubmitting}>Тестировать</Button>
		</form>
	</div>

	{#if models !== undefined}
		<div class="models-section">
			<div class="models-section-header">
				<h3 class="models-section-title">Модели</h3>
				<button
					type="button"
					class="models-add-trigger"
					class:models-add-trigger--active={addingModel}
					onclick={() => (addingModel = !addingModel)}
					aria-expanded={addingModel}
					aria-label="Добавить модель"
				>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="models-add-trigger-icon" class:models-add-trigger-icon--open={addingModel}>
						<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
					</svg>
					Добавить модель
				</button>
			</div>

			{#if addingModel}
				<div class="models-add-panel">
					<form method="POST" action="?/createModel" use:enhance={enhanceModel} class="models-add-form">
						<input type="hidden" name="provider" value={providerKey} />
						<div class="models-add-row">
							<FormField label="Model ID" id="{idPrefix}-new-model-id">
								<Input
									id="{idPrefix}-new-model-id"
									name="model_id"
									bind:value={newModelId}
									placeholder="например: deepseek/deepseek-v3"
								/>
							</FormField>
							<FormField label="Отображаемое имя" id="{idPrefix}-new-display-name">
								<Input
									id="{idPrefix}-new-display-name"
									name="display_name"
									bind:value={newDisplayName}
									placeholder="например: DeepSeek V3"
								/>
							</FormField>
							<div class="models-add-btn-wrap">
								<Button type="submit">Добавить</Button>
								<button type="button" class="models-cancel-btn" onclick={() => (addingModel = false)}>Отмена</button>
							</div>
						</div>
					</form>
				</div>
			{/if}

			{#if models.length === 0}
				<p class="models-empty">Нет моделей</p>
			{:else}
				<div class="models-table-wrapper">
					<table class="models-table">
						<thead>
							<tr>
								<th>Model ID</th>
								<th>Имя</th>
								<th>Статус</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each models as m (m.id)}
								<tr class:models-row-disabled={!m.enabled}>
									<td class="models-cell-id"><code>{m.model_id}</code></td>
									<td>
										{#if editingModelId === m.id}
											<form method="POST" action="?/updateModel" use:enhance={enhanceModel} class="models-edit-form">
												<input type="hidden" name="id" value={m.id} />
												<input type="hidden" name="enabled" value={String(m.enabled)} />
												<Input id="edit-{idPrefix}-{m.id}" name="display_name" bind:value={editDisplayName} />
												<div class="models-edit-actions">
													<Button type="submit" variant="secondary">OK</Button>
													<button type="button" class="models-cancel-btn" onclick={cancelEditModel}>Отмена</button>
												</div>
											</form>
										{:else}
											<span class="models-display-name">{m.display_name}</span>
										{/if}
									</td>
									<td>
										<form method="POST" action="?/updateModel" use:enhance={enhanceModel} class="models-toggle-form">
											<input type="hidden" name="id" value={m.id} />
											<input type="hidden" name="display_name" value={m.display_name} />
											<input type="hidden" name="enabled" value={String(!m.enabled)} />
											<button type="submit" class="models-status-btn" class:models-status-enabled={m.enabled} class:models-status-disabled={!m.enabled}>
												{m.enabled ? 'Вкл' : 'Выкл'}
											</button>
										</form>
									</td>
									<td class="models-cell-actions">
										{#if editingModelId !== m.id}
											<button type="button" class="models-action-btn" title="Редактировать" onclick={() => startEditModel(m)}>
												<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
											</button>
										{/if}
										{#if deletingModelId === m.id}
											<form method="POST" action="?/deleteModel" use:enhance={enhanceModel} class="models-delete-confirm">
												<input type="hidden" name="id" value={m.id} />
												<span class="models-delete-text">Удалить?</span>
												<Button type="submit" variant="secondary">Да</Button>
												<button type="button" class="models-cancel-btn" onclick={() => (deletingModelId = null)}>Нет</button>
											</form>
										{:else}
											<button type="button" class="models-action-btn models-action-delete" title="Удалить" onclick={() => (deletingModelId = m.id)}>
												<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
											</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</Card>

<style>
	:global(.model-card) {
		display: flex;
		flex-direction: column;
	}
	.model-card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		flex-wrap: wrap;
		margin-bottom: 1.25rem;
	}
	.model-card-title {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: #18181b;
	}
	.model-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		transition: opacity 0.2s ease;
	}
	.model-form--disabled {
		opacity: 0.55;
	}
	.model-fields-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: flex-end;
	}
	:global(.model-fields-row .field) {
		flex: 1;
		min-width: 8rem;
	}
	.model-switch-block {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.25rem;
	}
	.model-switch-label {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		cursor: pointer;
	}
	.model-switch:disabled + span {
		opacity: 0.8;
		cursor: default;
	}
	.model-switch:disabled {
		cursor: not-allowed;
	}
	.model-usage-msg {
		margin: 0;
		font-size: 0.8125rem;
		color: #71717a;
	}
	.model-actions-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
		margin-top: 1.25rem;
		transition: opacity 0.2s ease;
	}
	.model-actions-row--disabled {
		opacity: 0.55;
	}
	.model-actions-form,
	.model-test-form {
		display: inline-block;
	}
	.model-test-form {
		margin-left: auto;
	}

	/* Models section */
	.models-section {
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid #e5e7eb;
	}
	.models-section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}
	.models-section-title {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
		color: #374151;
	}
	.models-add-trigger {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.3125rem 0.625rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #374151;
		background: transparent;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s, border-color 0.15s;
		white-space: nowrap;
	}
	.models-add-trigger:hover {
		color: #111;
		background: #f4f4f5;
		border-color: #9ca3af;
	}
	.models-add-trigger--active {
		color: #111;
		background: #e4e4e7;
		border-color: #9ca3af;
	}
	.models-add-trigger-icon {
		flex-shrink: 0;
		transition: transform 0.2s;
	}
	.models-add-trigger-icon--open {
		transform: rotate(45deg);
	}
	.models-add-panel {
		margin-bottom: 0.875rem;
		padding: 1rem;
		background: #f4f4f5;
		border-radius: 8px;
		border: 1px solid #e4e4e7;
	}
	.models-add-form {
		margin: 0;
	}
	.models-add-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		align-items: flex-end;
	}
	.models-add-row :global(.field) {
		flex: 1;
		min-width: 9rem;
	}
	.models-add-btn-wrap {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
		padding-bottom: 0.125rem;
	}
	.models-empty {
		color: #71717a;
		font-size: 0.875rem;
		margin: 0 0 1rem;
	}
	.models-table-wrapper {
		overflow-x: auto;
		margin-bottom: 1rem;
	}
	.models-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}
	.models-table th {
		text-align: left;
		font-weight: 600;
		color: #374151;
		padding: 0.4rem 0.625rem;
		border-bottom: 1px solid #e5e7eb;
		white-space: nowrap;
	}
	.models-table td {
		padding: 0.4rem 0.625rem;
		border-bottom: 1px solid #f3f4f6;
		vertical-align: middle;
	}
	.models-table tr:last-child td {
		border-bottom: none;
	}
	.models-row-disabled {
		opacity: 0.5;
	}
	.models-cell-id code {
		font-family: ui-monospace, monospace;
		font-size: 0.8125rem;
		background: #f4f4f5;
		padding: 0.125rem 0.375rem;
		border-radius: 4px;
	}
	.models-display-name {
		color: #18181b;
	}
	.models-toggle-form {
		margin: 0;
		display: inline;
	}
	.models-status-btn {
		display: inline-block;
		padding: 0.2rem 0.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		border: 1px solid transparent;
		border-radius: 9999px;
		cursor: pointer;
		transition: all 0.15s;
	}
	.models-status-enabled {
		background: #dcfce7;
		color: #166534;
		border-color: #bbf7d0;
	}
	.models-status-enabled:hover {
		background: #bbf7d0;
	}
	.models-status-disabled {
		background: #fef2f2;
		color: #991b1b;
		border-color: #fecaca;
	}
	.models-status-disabled:hover {
		background: #fecaca;
	}
	.models-cell-actions {
		white-space: nowrap;
		text-align: right;
	}
	.models-action-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.75rem;
		height: 1.75rem;
		padding: 0;
		border: none;
		background: none;
		color: #9ca3af;
		border-radius: 5px;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.models-action-btn:hover {
		color: #111;
		background: #f4f4f5;
	}
	.models-action-delete:hover {
		color: #b91c1c;
		background: #fef2f2;
	}
	.models-edit-form {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0;
	}
	.models-edit-actions {
		display: flex;
		gap: 0.25rem;
		align-items: center;
	}
	.models-cancel-btn {
		padding: 0.25rem 0.5rem;
		font-size: 0.8125rem;
		color: #6b7280;
		background: none;
		border: none;
		cursor: pointer;
		text-decoration: underline;
	}
	.models-cancel-btn:hover {
		color: #111;
	}
	.models-delete-confirm {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		margin: 0;
	}
	.models-delete-text {
		font-size: 0.8125rem;
		color: #b91c1c;
		font-weight: 500;
	}
</style>
