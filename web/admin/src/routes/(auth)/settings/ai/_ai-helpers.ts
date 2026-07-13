import type { AiSettingsData, AIProviderModel } from './_types';

const PROVIDERS = [
	{ value: 'gemini', label: 'Gemini' },
	{ value: 'grok', label: 'Grok' },
	{ value: 'openai', label: 'OpenAI' },
	{ value: 'openrouter', label: 'OpenRouter' }
] as const;

export function groupModelsByProvider(
	models: AIProviderModel[]
): Record<string, { value: string; label: string }[]> {
	const groups: Record<string, { value: string; label: string }[]> = {};
	for (const m of models) {
		if (!m.enabled) continue;
		if (!groups[m.provider]) groups[m.provider] = [];
		groups[m.provider].push({ value: m.model_id, label: m.display_name });
	}
	return groups;
}

export function filterEnabledProviders(aiData: AiSettingsData) {
	return PROVIDERS.filter((p) => {
		if (p.value === 'gemini') return aiData.geminiEnable;
		if (p.value === 'grok') return aiData.grokEnable;
		if (p.value === 'openai') return aiData.openaiEnable;
		return aiData.openrouterEnable;
	});
}

export function noReload() {
	return async ({
		update
	}: {
		update: (opts?: { invalidateAll?: boolean; reset?: boolean }) => Promise<void>;
	}) => {
		await update?.({ invalidateAll: false, reset: false });
	};
}
