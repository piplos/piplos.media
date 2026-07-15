<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import toast from 'svelte-french-toast';
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import MarkdownEditor from '$lib/components/MarkdownEditor.svelte';
	import type { LeadEmailTemplate } from './+page.server';

	let { data } = $props();

	const template = $derived(data.template as LeadEmailTemplate);

	let saving = $state(false);
	// Writable derived: значения из load, редактируемые локально; сбрасываются при invalidateAll.
	let subject = $derived(template.subject);
	let body = $derived(template.body);

	const variables = [
		{ name: 'display_name', hint: 'Название проекта или имя клиента' },
		{ name: 'project_name', hint: 'Название проекта' },
		{ name: 'types', hint: 'Тип проекта' },
		{ name: 'description', hint: 'Описание' },
		{ name: 'stack', hint: 'Стек' },
		{ name: 'references', hint: 'Референсы' },
		{ name: 'budget', hint: 'Бюджет с валютой' },
		{ name: 'timeline', hint: 'Сроки' },
		{ name: 'stage', hint: 'Стадия проекта' },
		{ name: 'first_name', hint: 'Имя' },
		{ name: 'last_name', hint: 'Фамилия' },
		{ name: 'name', hint: 'Имя и фамилия' },
		{ name: 'email', hint: 'Email клиента' },
		{ name: 'company', hint: 'Компания' },
		{ name: 'phone', hint: 'Телефон' },
		{ name: 'how_found', hint: 'Как нашли' },
		{ name: 'notes', hint: 'Примечания' },
		{ name: 'lang', hint: 'Язык заявки' },
		{ name: 'created_at', hint: 'Дата заявки' },
		{ name: 'lead_url', hint: 'Ссылка на заявку в админке' },
		{ name: 'id', hint: 'ID заявки' }
	];

	async function copyVariable(name: string) {
		const placeholder = `{{${name}}}`;
		try {
			await navigator.clipboard.writeText(placeholder);
			toast.success(`${placeholder} — скопировано`);
		} catch {
			toast.error('Не удалось скопировать');
		}
	}
</script>

<svelte:head>
	<title>Шаблон письма — SMTP — Настройки — Piplos Admin</title>
</svelte:head>

{#if data.error}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">{data.error}</p>
	</div>
{:else}
	<Card padding="sm">
		<h2 class="section-title">Письмо о новой заявке</h2>

		<form
			method="POST"
			action="?/updateTemplate"
			class="template-form"
			use:enhance={() => {
				saving = true;
				return async ({ result }) => {
					saving = false;
					if (result.type === 'success') {
						toast.success('Шаблон письма сохранён');
						await invalidateAll();
					} else if (result.type === 'failure') {
						toast.error((result.data?.error as string) ?? 'Не удалось сохранить');
					}
				};
			}}
		>
			<FormField label="Тема письма" id="tpl-subject" hint="Пусто — название проекта или имя клиента.">
				<Input
					id="tpl-subject"
					name="subject"
					bind:value={subject}
					placeholder={'{{name}} — новая заявка ({{project_name}})'}
				/>
			</FormField>
			<FormField label="Текст письма" id="tpl-body">
				<input type="hidden" name="body" value={body} />
				<MarkdownEditor
					id="tpl-body"
					bind:value={body}
					rows={14}
					placeholder={'## {{name}}\n{{email}} · {{phone}}\n...\n\n[Открыть в админке →]({{lead_url}})'}
				/>
			</FormField>
			<div class="form-actions">
				<Button type="submit" loading={saving}>Сохранить</Button>
			</div>
		</form>
	</Card>

	<Card padding="sm">
		<h2 class="section-title">Доступные переменные</h2>
		<p class="section-hint">Нажмите на переменную, чтобы скопировать её в буфер обмена.</p>
		<ul class="vars-list">
			{#each variables as v (v.name)}
				<li class="vars-item">
					<button
						type="button"
						class="var-chip"
						title="Скопировать"
						onclick={() => copyVariable(v.name)}
					>
						{'{{' + v.name + '}}'}
					</button>
					<span class="var-hint">{v.hint}</span>
				</li>
			{/each}
		</ul>
	</Card>
{/if}

<style>
	.text-muted {
		color: #71717a;
		font-size: 0.875rem;
	}
	.section-title {
		margin: 0 0 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		color: #18181b;
	}
	.section-hint {
		margin: 0 0 1rem;
		font-size: 0.8125rem;
		color: #71717a;
		line-height: 1.5;
	}
	.template-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.form-actions {
		display: flex;
		justify-content: flex-end;
	}
	.vars-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
		gap: 0.375rem 1rem;
	}
	.vars-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}
	.var-chip {
		flex-shrink: 0;
		padding: 0.125rem 0.5rem;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.75rem;
		color: #18181b;
		background: #f4f4f5;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s;
	}
	.var-chip:hover {
		background: #e5e7eb;
		border-color: #d4d4d8;
	}
	.var-hint {
		font-size: 0.75rem;
		color: #71717a;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
