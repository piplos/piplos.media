<script lang="ts">
	import type { AiSettingsData, AIProviderModel, ModelSettings } from './_types';
	import { noReload } from './_ai-helpers';
	import ModelCard from './ModelCard.svelte';
	import toast from 'svelte-french-toast';

	const TOAST_MESSAGES: Record<string, string> = {
		geminiSuccess: 'Настройки Gemini сохранены',
		grokSuccess: 'Настройки Grok сохранены',
		openaiSuccess: 'Настройки OpenAI сохранены',
		openrouterSuccess: 'Настройки OpenRouter сохранены',
		testGeminiSuccess: 'Ключ Gemini действителен',
		testGrokSuccess: 'Ключ Grok действителен',
		testOpenAISuccess: 'Ключ OpenAI действителен',
		testOpenRouterSuccess: 'Ключ OpenRouter действителен'
	};

	const ERROR_KEYS = [
		'geminiError',
		'grokError',
		'openaiError',
		'openrouterError',
		'testGeminiError',
		'testGrokError',
		'testOpenAIError',
		'testOpenRouterError'
	] as const;

	function getToastFromForm(form: Record<string, unknown> | null): { message: string | null; isError: boolean } {
		if (form == null) return { message: null, isError: false };
		for (const key of Object.keys(TOAST_MESSAGES)) {
			if (form[key] === true) return { message: TOAST_MESSAGES[key], isError: false };
		}
		for (const key of ERROR_KEYS) {
			const value = form[key];
			if (typeof value === 'string') return { message: value, isError: true };
		}
		return { message: null, isError: false };
	}

	let { data, form } = $props();

	const layoutData = $derived(
		data as AiSettingsData & {
			gemini?: ModelSettings;
			grok?: ModelSettings;
			openai?: ModelSettings;
			openrouter?: ModelSettings;
			aiModels?: AIProviderModel[];
			error?: string | null;
		}
	);
	const gemini = $derived(layoutData.gemini as ModelSettings);
	const grok = $derived(layoutData.grok as ModelSettings);
	const openai = $derived(layoutData.openai as ModelSettings);
	const openrouter = $derived(layoutData.openrouter as ModelSettings);
	const pageError = $derived(layoutData.error as string | null);
	const translationProvider = $derived(layoutData.translationProvider as string);

	const allModels = $derived(layoutData.aiModels ?? []);
	const geminiModels = $derived(allModels.filter((m) => m.provider === 'gemini'));
	const grokModels = $derived(allModels.filter((m) => m.provider === 'grok'));
	const openaiModels = $derived(allModels.filter((m) => m.provider === 'openai'));
	const openrouterModels = $derived(allModels.filter((m) => m.provider === 'openrouter'));

	let geminiEnable = $state(false);
	let geminiApiKey = $state('');
	let geminiApiKeyDirty = $state(false);
	let geminiRateLimit = $state('10');
	let geminiTimeoutSeconds = $state('60');

	let grokEnable = $state(false);
	let grokApiKey = $state('');
	let grokApiKeyDirty = $state(false);
	let grokRateLimit = $state('10');
	let grokTimeoutSeconds = $state('15');

	let openaiEnable = $state(false);
	let openaiApiKey = $state('');
	let openaiApiKeyDirty = $state(false);
	let openaiRateLimit = $state('10');
	let openaiTimeoutSeconds = $state('30');

	let openrouterEnable = $state(false);
	let openrouterApiKey = $state('');
	let openrouterApiKeyDirty = $state(false);
	let openrouterRateLimit = $state('10');
	let openrouterTimeoutSeconds = $state('30');

	$effect(() => {
		geminiEnable = gemini.enable;
		geminiApiKey = gemini.apiKey;
		geminiRateLimit = String(gemini.rateLimit);
		geminiTimeoutSeconds = String(gemini.timeoutSeconds);
	});
	$effect(() => {
		grokEnable = grok.enable;
		grokApiKey = grok.apiKey;
		grokRateLimit = String(grok.rateLimit);
		grokTimeoutSeconds = String(grok.timeoutSeconds);
	});
	$effect(() => {
		openaiEnable = openai.enable;
		openaiApiKey = openai.apiKey;
		openaiRateLimit = String(openai.rateLimit);
		openaiTimeoutSeconds = String(openai.timeoutSeconds);
	});
	$effect(() => {
		openrouterEnable = openrouter.enable;
		openrouterApiKey = openrouter.apiKey;
		openrouterRateLimit = String(openrouter.rateLimit);
		openrouterTimeoutSeconds = String(openrouter.timeoutSeconds);
	});

	const geminiApiKeyPlaceholder = $derived(gemini.apiKeyMasked ? '•••••••• (задан)' : '');
	const grokApiKeyPlaceholder = $derived(grok.apiKeyMasked ? '•••••••• (задан)' : '');
	const openaiApiKeyPlaceholder = $derived(openai.apiKeyMasked ? '•••••••• (задан)' : '');
	const openrouterApiKeyPlaceholder = $derived(openrouter.apiKeyMasked ? '•••••••• (задан)' : '');

	const formToast = $derived(getToastFromForm(form as Record<string, unknown> | null));
	let lastToastKey = $state<string | null>(null);
	$effect(() => {
		const t = formToast;
		if (t.message == null) {
			lastToastKey = null;
			return;
		}
		const key = `${t.isError}:${t.message}`;
		if (key === lastToastKey) return;
		lastToastKey = key;
		if (t.isError) toast.error(t.message);
		else toast.success(t.message);
	});
</script>

<svelte:head>
	<title>Провайдеры — AI — Настройки — Piplos Admin</title>
</svelte:head>

{#if pageError}
	<div class="admin-table-wrap admin-table-wrap--empty">
		<p class="text-muted">Данные недоступны</p>
	</div>
{:else}
	<div class="models-grid">
		<ModelCard
			title="Gemini"
			bind:enable={geminiEnable}
			bind:apiKey={geminiApiKey}
			bind:apiKeyDirty={geminiApiKeyDirty}
			bind:rateLimit={geminiRateLimit}
			bind:timeoutSeconds={geminiTimeoutSeconds}
			apiKeyPlaceholder={geminiApiKeyPlaceholder}
			updateAction="?/updateGemini"
			testAction="?/testGeminiKey"
			{noReload}
			idPrefix="gemini"
			usedForTranslation={translationProvider === 'gemini'}
			models={geminiModels}
			providerKey="gemini"
		/>
		<ModelCard
			title="Grok"
			bind:enable={grokEnable}
			bind:apiKey={grokApiKey}
			bind:apiKeyDirty={grokApiKeyDirty}
			bind:rateLimit={grokRateLimit}
			bind:timeoutSeconds={grokTimeoutSeconds}
			apiKeyPlaceholder={grokApiKeyPlaceholder}
			updateAction="?/updateGrok"
			testAction="?/testGrokKey"
			{noReload}
			idPrefix="grok"
			usedForTranslation={translationProvider === 'grok'}
			models={grokModels}
			providerKey="grok"
		/>
		<ModelCard
			title="OpenAI"
			bind:enable={openaiEnable}
			bind:apiKey={openaiApiKey}
			bind:apiKeyDirty={openaiApiKeyDirty}
			bind:rateLimit={openaiRateLimit}
			bind:timeoutSeconds={openaiTimeoutSeconds}
			apiKeyPlaceholder={openaiApiKeyPlaceholder}
			updateAction="?/updateOpenAI"
			testAction="?/testOpenAIKey"
			{noReload}
			idPrefix="openai"
			usedForTranslation={translationProvider === 'openai'}
			models={openaiModels}
			providerKey="openai"
		/>
		<ModelCard
			title="OpenRouter"
			bind:enable={openrouterEnable}
			bind:apiKey={openrouterApiKey}
			bind:apiKeyDirty={openrouterApiKeyDirty}
			bind:rateLimit={openrouterRateLimit}
			bind:timeoutSeconds={openrouterTimeoutSeconds}
			apiKeyPlaceholder={openrouterApiKeyPlaceholder}
			updateAction="?/updateOpenRouter"
			testAction="?/testOpenRouterKey"
			{noReload}
			idPrefix="openrouter"
			usedForTranslation={translationProvider === 'openrouter'}
			models={openrouterModels}
			providerKey="openrouter"
		/>
	</div>
{/if}

<style>
	.text-muted {
		color: #71717a;
	}
	.models-grid {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}
</style>
