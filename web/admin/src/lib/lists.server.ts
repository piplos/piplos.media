/** Загрузка справочников «Списки» (услуги, стек) для форм. */
import { isRedirect, type RequestEvent } from '@sveltejs/kit';
import { fetchWithAuth } from '$lib/api.server';
import type { Service, StackItem } from '$lib/types';

export async function loadServices(event: RequestEvent): Promise<Service[]> {
	try {
		const res = await fetchWithAuth(event, '/api/v1/services');
		if (!res.ok) return [];
		const data = (await res.json()) as { services: Service[] };
		return data.services ?? [];
	} catch (e) {
		if (isRedirect(e)) throw e;
		return [];
	}
}

export async function loadStack(event: RequestEvent): Promise<StackItem[]> {
	try {
		const res = await fetchWithAuth(event, '/api/v1/stack');
		if (!res.ok) return [];
		const data = (await res.json()) as { stack: StackItem[] };
		return data.stack ?? [];
	} catch (e) {
		if (isRedirect(e)) throw e;
		return [];
	}
}
