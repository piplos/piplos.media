/**
 * Утилиты для API настроек (/v1/settings). Значения — композитный JSON
 * (ключи AI, SMTP); чувствительные поля приходят маскированными ("****").
 */
import type { RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import type { Setting } from '$lib/types';

/** Маска секретных полей в ответах API. */
export const MASKED = '****';

export interface SettingsListResponse {
	settings: Setting[];
}

export function getByKey(settings: Setting[], key: string): string {
	return settings.find((s) => s.key === key)?.value ?? '';
}

export async function updateSetting(
	event: RequestEvent,
	key: string,
	value: string
): Promise<Response> {
	return fetchWithAuth(event, `/v1/settings/${encodeURIComponent(key)}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ value })
	});
}
