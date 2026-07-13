import { fail, isRedirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, updateSetting, type SettingsListResponse } from '$lib/settings.server';
import type { Setting } from '$lib/types';
import { loadAIModels } from './_ai-settings.server';
import type { AIProviderModel, ModelSettings } from './_types';

type ActionEvent = Parameters<Actions['updateGemini']>[0];

const MASKED = '****';

function parseModelFromComposite(settings: Setting[], compositeKey: string): ModelSettings {
	const raw = getByKey(settings, compositeKey);
	if (!raw) {
		return { enable: false, apiKey: '', apiKeyMasked: false, rateLimit: 0, timeoutSeconds: 0 };
	}
	try {
		const obj = JSON.parse(raw) as Record<string, unknown>;
		const apiKeyVal = (obj.apiKey as string) ?? '';
		return {
			enable: obj.enable === true,
			apiKey: apiKeyVal === MASKED ? '' : apiKeyVal,
			apiKeyMasked: apiKeyVal === MASKED,
			rateLimit: Number(obj.rateLimit) || 0,
			timeoutSeconds: Number(obj.timeoutSeconds) || 0
		};
	} catch {
		return { enable: false, apiKey: '', apiKeyMasked: false, rateLimit: 0, timeoutSeconds: 0 };
	}
}

export interface AiModelsPageData {
	gemini: ModelSettings;
	grok: ModelSettings;
	openai: ModelSettings;
	openrouter: ModelSettings;
	aiModels: AIProviderModel[];
	error: string | null;
}

export const load: PageServerLoad = async (event): Promise<AiModelsPageData> => {
	const emptyModel: ModelSettings = {
		enable: false,
		apiKey: '',
		apiKeyMasked: false,
		rateLimit: 0,
		timeoutSeconds: 0
	};
	const defaultModels = {
		gemini: { ...emptyModel },
		grok: { ...emptyModel },
		openai: { ...emptyModel },
		openrouter: { ...emptyModel }
	};
	try {
		const [res, aiModels] = await Promise.all([
			fetchWithAuth(event, '/api/v1/settings'),
			loadAIModels(event)
		]);
		if (!res.ok) {
			return {
				...defaultModels,
				aiModels,
				error: apiLoadErrorMessage(res, 'Не удалось загрузить настройки')
			};
		}
		const data = (await res.json()) as SettingsListResponse;
		const list = data.settings ?? [];
		return {
			gemini: parseModelFromComposite(list, 'GEMINI'),
			grok: parseModelFromComposite(list, 'GROK'),
			openai: parseModelFromComposite(list, 'OPENAI'),
			openrouter: parseModelFromComposite(list, 'OPENROUTER'),
			aiModels,
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { ...defaultModels, aiModels: [], error: 'Ошибка запроса к API' };
	}
};

interface ParsedModelForm {
	enable: boolean;
	apiKey: string;
	rateLimit: number;
	timeoutSeconds: number;
	apiKeyDirty: boolean;
}

function parseModelForm(fd: FormData): ParsedModelForm {
	const enable = fd.get('enable') === 'on' || fd.get('enable') === 'true';
	const apiKey = (fd.get('apiKey') as string)?.trim() ?? '';
	const rateLimit = Math.max(0, parseInt((fd.get('rateLimit') as string) || '0', 10));
	const timeoutSeconds = Math.max(0, parseInt((fd.get('timeoutSeconds') as string) || '0', 10));
	const apiKeyDirty = (fd.get('apiKeyDirty') as string) === 'true';
	return { enable, apiKey, rateLimit, timeoutSeconds, apiKeyDirty };
}

async function applyModelUpdate(event: ActionEvent, compositeKey: string, parsed: ParsedModelForm) {
	const payload: Record<string, unknown> = {
		enable: parsed.enable,
		rateLimit: parsed.rateLimit,
		timeoutSeconds: parsed.timeoutSeconds
	};
	if (parsed.apiKeyDirty) {
		payload.apiKey = parsed.apiKey;
	} else {
		const res = await fetchWithAuth(event, `/api/v1/settings/${encodeURIComponent(compositeKey)}`);
		if (res.ok) {
			const data = (await res.json()) as { value?: string };
			try {
				const current = JSON.parse(data.value ?? '{}') as Record<string, unknown>;
				payload.apiKey = current.apiKey ?? '';
			} catch {
				payload.apiKey = '';
			}
		}
	}
	await updateSetting(event, compositeKey, JSON.stringify(payload));
}

async function testApiKey(
	event: ActionEvent,
	compositeKey: string,
	apiKeyValue: string,
	successKey: string,
	errorKey: string,
	errorLabel: string
) {
	try {
		const payload: Record<string, unknown> = { key: compositeKey, value: '' };
		if (apiKeyValue) {
			payload.value = JSON.stringify({ enable: true, apiKey: apiKeyValue });
		}
		const res = await fetchWithAuth(event, '/api/v1/settings/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		const body = (await res.json().catch(() => ({}))) as { message?: string };
		if (!res.ok) {
			return fail(res.status === 400 ? 400 : 500, {
				[errorKey]: body.message ?? 'Ключ недействителен'
			});
		}
		return { [successKey]: true };
	} catch (e) {
		if (isRedirect(e)) throw e;
		return fail(500, { [errorKey]: errorLabel });
	}
}

export const actions: Actions = {
	updateGemini: async (event) => {
		try {
			await applyModelUpdate(event, 'GEMINI', parseModelForm(await event.request.formData()));
			return { geminiSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { geminiError: 'Не удалось сохранить настройки Gemini' });
		}
	},
	updateGrok: async (event) => {
		try {
			await applyModelUpdate(event, 'GROK', parseModelForm(await event.request.formData()));
			return { grokSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { grokError: 'Не удалось сохранить настройки Grok' });
		}
	},
	updateOpenAI: async (event) => {
		try {
			await applyModelUpdate(event, 'OPENAI', parseModelForm(await event.request.formData()));
			return { openaiSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { openaiError: 'Не удалось сохранить настройки OpenAI' });
		}
	},
	updateOpenRouter: async (event) => {
		try {
			await applyModelUpdate(event, 'OPENROUTER', parseModelForm(await event.request.formData()));
			return { openrouterSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { openrouterError: 'Не удалось сохранить настройки OpenRouter' });
		}
	},
	testGeminiKey: async (event) => {
		const fd = await event.request.formData();
		const apiKey = (fd.get('apiKey') as string)?.trim() ?? '';
		return testApiKey(event, 'GEMINI', apiKey, 'testGeminiSuccess', 'testGeminiError', 'Ошибка запроса к API');
	},
	testGrokKey: async (event) => {
		const fd = await event.request.formData();
		const apiKey = (fd.get('apiKey') as string)?.trim() ?? '';
		return testApiKey(event, 'GROK', apiKey, 'testGrokSuccess', 'testGrokError', 'Ошибка запроса к API');
	},
	testOpenAIKey: async (event) => {
		const fd = await event.request.formData();
		const apiKey = (fd.get('apiKey') as string)?.trim() ?? '';
		return testApiKey(event, 'OPENAI', apiKey, 'testOpenAISuccess', 'testOpenAIError', 'Ошибка запроса к API');
	},
	testOpenRouterKey: async (event) => {
		const fd = await event.request.formData();
		const apiKey = (fd.get('apiKey') as string)?.trim() ?? '';
		return testApiKey(
			event,
			'OPENROUTER',
			apiKey,
			'testOpenRouterSuccess',
			'testOpenRouterError',
			'Ошибка запроса к API'
		);
	},
	createModel: async (event) => {
		const fd = await event.request.formData();
		const provider = ((fd.get('provider') as string) ?? '').trim().toLowerCase();
		const modelId = ((fd.get('model_id') as string) ?? '').trim();
		const displayName = ((fd.get('display_name') as string) ?? '').trim();
		if (!provider || !modelId || !displayName) {
			return fail(400, { createError: 'Все поля обязательны' });
		}
		try {
			const res = await fetchWithAuth(event, '/api/v1/ai-models', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ provider, model_id: modelId, display_name: displayName })
			});
			if (!res.ok) {
				const body = (await res.json().catch(() => ({}))) as { message?: string };
				const msg = body.message ?? (res.status === 409 ? 'Модель уже существует' : 'Ошибка создания');
				return fail(res.status, { createError: msg });
			}
			return { createSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { createError: 'Ошибка запроса к API' });
		}
	},
	updateModel: async (event) => {
		const fd = await event.request.formData();
		const id = (fd.get('id') as string) ?? '';
		const displayName = ((fd.get('display_name') as string) ?? '').trim();
		const enabled = fd.get('enabled') === 'true';
		if (!id) return fail(400, { updateError: 'ID обязателен' });
		try {
			const res = await fetchWithAuth(event, `/api/v1/ai-models/${id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ display_name: displayName || undefined, enabled })
			});
			if (!res.ok) {
				const body = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, { updateError: body.message ?? 'Ошибка обновления' });
			}
			return { updateSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { updateError: 'Ошибка запроса к API' });
		}
	},
	deleteModel: async (event) => {
		const fd = await event.request.formData();
		const id = (fd.get('id') as string) ?? '';
		if (!id) return fail(400, { deleteError: 'ID обязателен' });
		try {
			const res = await fetchWithAuth(event, `/api/v1/ai-models/${id}`, { method: 'DELETE' });
			if (!res.ok) {
				const body = (await res.json().catch(() => ({}))) as { message?: string };
				return fail(res.status, { deleteError: body.message ?? 'Ошибка удаления' });
			}
			return { deleteSuccess: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { deleteError: 'Ошибка запроса к API' });
		}
	}
};
