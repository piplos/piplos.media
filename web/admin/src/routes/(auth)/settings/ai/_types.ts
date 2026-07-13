export interface AIProviderModel {
	id: string;
	provider: string;
	model_id: string;
	display_name: string;
	enabled: boolean;
}

export interface AiSettingsData {
	translationProvider: string;
	translationModel: string;
	translationPrompt: string;
	geminiEnable: boolean;
	grokEnable: boolean;
	openaiEnable: boolean;
	openrouterEnable: boolean;
	aiModels: AIProviderModel[];
	error: string | null;
}

export interface ModelSettings {
	enable: boolean;
	apiKey: string;
	apiKeyMasked: boolean;
	rateLimit: number;
	timeoutSeconds: number;
}
