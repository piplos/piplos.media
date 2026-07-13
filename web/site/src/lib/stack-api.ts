import { API_URL } from '$lib/api';

export interface StackItem {
	id: string;
	slug: string;
	label: string;
	group_id: string;
	published: boolean;
	sort_order: number;
}

type FetchFn = typeof fetch;

/** Опубликованные технологии для секции «Стек» на сайте. */
export async function fetchStackItems(fetchFn: FetchFn = fetch): Promise<StackItem[]> {
	try {
		const res = await fetchFn(`${API_URL}/api/v1/public/stack`);
		if (!res.ok) return [];
		const data = (await res.json()) as { stack: StackItem[] };
		return (data.stack ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

export interface StackDisplayItem {
	slug: string;
	label: string;
}

/** Преобразует API-записи в формат для UI (slug + label). */
export function toStackDisplayItems(items: StackItem[]): StackDisplayItem[] {
	return items.map((item) => ({ slug: item.slug, label: item.label }));
}
