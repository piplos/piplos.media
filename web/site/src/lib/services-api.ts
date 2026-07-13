import { API_URL } from '$lib/api';

export interface ServiceTranslation {
	title: string;
	description: string;
}

export interface ServiceItem {
	id: string;
	slug: string;
	icon: string;
	tags: string[];
	published: boolean;
	sort_order: number;
	translations: Record<string, ServiceTranslation>;
}

export interface ServiceDisplayItem {
	id: string;
	title: string;
	description: string;
	tags: string[];
	icon: string;
}

type FetchFn = typeof fetch;

/** Опубликованные услуги для секции на главной. */
export async function fetchServices(fetchFn: FetchFn = fetch): Promise<ServiceItem[]> {
	try {
		const res = await fetchFn(`${API_URL}/api/v1/public/services`);
		if (!res.ok) return [];
		const data = (await res.json()) as { services: ServiceItem[] };
		return (data.services ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

/** Преобразует API-записи в формат для UI с учётом языка. */
export function toServiceDisplayItems(
	services: ServiceItem[],
	lang: string
): ServiceDisplayItem[] {
	return [...services]
		.sort((a, b) => a.sort_order - b.sort_order)
		.map((item) => {
			const locale = item.translations[lang] ?? item.translations.ru ?? item.translations.en;
			return {
				id: item.slug,
				title: locale?.title ?? item.slug,
				description: locale?.description ?? '',
				tags: item.tags ?? [],
				icon: item.icon
			};
		});
}
