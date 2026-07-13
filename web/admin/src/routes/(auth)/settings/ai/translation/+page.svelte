<script lang="ts">
	import type { AiSettingsData } from '../_types';
	import AISettingsForm from '../AISettingsForm.svelte';

	let { data, form } = $props();
	const aiData = $derived(data as AiSettingsData);
</script>

<svelte:head>
	<title>Перевод контента — AI — Настройки — Piplos Admin</title>
</svelte:head>

<AISettingsForm
	{aiData}
	form={form as Record<string, unknown> | null}
	prefix="translation"
	title="Перевод контента"
	provider={aiData.translationProvider}
	model={aiData.translationModel}
	prompt={aiData.translationPrompt}
	saveSuccessMessage="Настройки перевода сохранены"
	testSuccessMessage="Тест перевода выполнен"
>
	{#snippet promptHint()}
		<p class="ai-placeholders-hint">
			<code>{'{target_language}'}</code> — название целевого языка перевода.
		</p>
	{/snippet}
</AISettingsForm>
