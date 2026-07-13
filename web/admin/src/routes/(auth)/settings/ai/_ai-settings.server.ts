import { fail, isRedirect } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiLoadErrorMessage, fetchWithAuth } from '$lib/api.server';
import { getByKey, updateSetting, type SettingsListResponse } from '$lib/settings.server';
import type { AiSettingsData, AIProviderModel } from './_types';

export const COMPOSITE_KEYS = {
	translation: 'AI_TRANSLATION'
} as const;

interface AITaskJSON {
	provider?: string;
	model?: string;
	prompt?: string;
}

function parseAITask(raw: string): { provider: string; model: string; prompt: string } {
	if (!raw) return { provider: '', model: '', prompt: '' };
	try {
		const obj = JSON.parse(raw) as AITaskJSON;
		return {
			provider: obj.provider ?? '',
			model: obj.model ?? '',
			prompt: obj.prompt ?? ''
		};
	} catch {
		return { provider: '', model: '', prompt: '' };
	}
}

function parseProviderEnable(compositeJson: string): boolean {
	if (!compositeJson) return false;
	try {
		const obj = JSON.parse(compositeJson) as { enable?: boolean };
		return obj.enable === true;
	} catch {
		return false;
	}
}

const defaults: AiSettingsData = {
	translationProvider: '',
	translationModel: '',
	translationPrompt: '',
	geminiEnable: false,
	grokEnable: false,
	openaiEnable: false,
	openrouterEnable: false,
	aiModels: [],
	error: null
};

export async function loadAIModels(event: RequestEvent): Promise<AIProviderModel[]> {
	try {
		const res = await fetchWithAuth(event, '/api/v1/ai-models');
		if (!res.ok) return [];
		const data = (await res.json()) as { models: AIProviderModel[] };
		return data.models ?? [];
	} catch {
		return [];
	}
}

export async function loadAiSettings(event: RequestEvent): Promise<AiSettingsData> {
	try {
		const [settingsRes, aiModels] = await Promise.all([
			fetchWithAuth(event, '/api/v1/settings'),
			loadAIModels(event)
		]);
		if (!settingsRes.ok) {
			return {
				...defaults,
				aiModels,
				error: apiLoadErrorMessage(settingsRes, 'Не удалось загрузить настройки')
			};
		}
		const data = (await settingsRes.json()) as SettingsListResponse;
		const list = data.settings ?? [];
		const trans = parseAITask(getByKey(list, COMPOSITE_KEYS.translation));

		return {
			translationProvider: trans.provider,
			translationModel: trans.model,
			translationPrompt: trans.prompt,
			geminiEnable: parseProviderEnable(getByKey(list, 'GEMINI')),
			grokEnable: parseProviderEnable(getByKey(list, 'GROK')),
			openaiEnable: parseProviderEnable(getByKey(list, 'OPENAI')),
			openrouterEnable: parseProviderEnable(getByKey(list, 'OPENROUTER')),
			aiModels,
			error: null
		};
	} catch (e) {
		if (isRedirect(e)) throw e;
		return { ...defaults, error: 'Ошибка запроса к API' };
	}
}

async function saveCompositeKey(
	event: RequestEvent,
	key: string,
	value: Record<string, unknown>
): Promise<{ message?: string } | null> {
	const res = await updateSetting(event, key, JSON.stringify(value));
	if (!res.ok) {
		const err = (await res.json().catch(() => ({}))) as { message?: string };
		return { message: err?.message ?? 'Не удалось сохранить настройки' };
	}
	return null;
}

function capitalize(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

export function makeUpdateAction(prefix: string, compositeKey: string) {
	return async (event: RequestEvent) => {
		const fd = await event.request.formData();
		const value: Record<string, unknown> = {
			provider: ((fd.get(`${prefix}Provider`) as string) ?? '').trim(),
			model: ((fd.get(`${prefix}Model`) as string) ?? '').trim(),
			prompt: (fd.get(`${prefix}Prompt`) as string) ?? ''
		};
		try {
			const err = await saveCompositeKey(event, compositeKey, value);
			if (err !== null) return fail(500, { [`${prefix}Error`]: err.message });
			return { [`${prefix}Success`]: true };
		} catch (e) {
			if (isRedirect(e)) throw e;
			return fail(500, { [`${prefix}Error`]: 'Ошибка запроса к API' });
		}
	};
}

export function makeTestAction(prefix: string, endpoint: string, timeoutMs = 100_000) {
	const cap = capitalize(prefix);
	const errorKey = `test${cap}Error`;

	return async (event: RequestEvent) => {
		const controller = new AbortController();
		const timeoutId = setTimeout(() => controller.abort(), timeoutMs);
		try {
			const res = await fetchWithAuth(event, endpoint, {
				method: 'POST',
				signal: controller.signal
			});
			const body = (await res.json().catch(() => ({}))) as {
				response?: { content?: unknown; user_prompt?: string };
				message?: string;
			};
			if (!res.ok) {
				const msg =
					res.status === 504
						? 'Превышено время ожидания ответа модели. Попробуйте позже.'
						: (body.message ?? 'Ошибка запроса');
				return fail(res.status, { [errorKey]: msg });
			}
			const response = body.response;
			if (response?.content == null) {
				return fail(500, { [errorKey]: 'Некорректный ответ API' });
			}
			return {
				[`test${cap}Success`]: true,
				[`test${cap}Response`]: response.content,
				[`test${cap}UserPrompt`]: response.user_prompt ?? ''
			};
		} catch (e: unknown) {
			if (isRedirect(e)) throw e;
			const isAbort =
				e !== null && typeof e === 'object' && 'name' in e && (e as { name: string }).name === 'AbortError';
			return fail(500, {
				[errorKey]: isAbort
					? 'Превышено время ожидания ответа модели. Попробуйте позже.'
					: 'Ошибка запроса к API'
			});
		} finally {
			clearTimeout(timeoutId);
		}
	};
}
