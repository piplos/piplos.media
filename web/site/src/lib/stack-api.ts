import { getApiV1, type ApiRequestContext } from '$lib/api';

export interface StackItem {
	id: string;
	slug: string;
	label: string;
	group_id: string;
	published: boolean;
	sort_order: number;
}

type FetchFn = typeof fetch;

/** Опубликованные технологии для секции «Стек» на сайте.
 *  Сортировка повторяет API: group_id, sort_order, label. */
export async function fetchStackItems(
	fetchFn: FetchFn = fetch,
	ctx?: ApiRequestContext
): Promise<StackItem[]> {
	try {
		const res = await fetchFn(`${getApiV1(ctx)}/public/stack`);
		if (!res.ok) return [];
		const data = (await res.json()) as { stack: StackItem[] };
		return (data.stack ?? [])
			.filter((item) => item.published)
			.sort(
				(a, b) =>
					a.group_id.localeCompare(b.group_id) ||
					a.sort_order - b.sort_order ||
					a.label.localeCompare(b.label)
			);
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
