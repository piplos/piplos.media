import { API_URL } from '$lib/api';
import { DEFAULT_LANG } from '$lib/i18n/routing';

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

/** Опубликованные услуги для секции на главной.
 *  lang — вернуть только этот перевод (фильтрация на сервере). */
export async function fetchServices(fetchFn: FetchFn = fetch, lang?: string): Promise<ServiceItem[]> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${API_URL}/api/v1/public/services${qs}`);
		if (!res.ok) return [];
		const data = (await res.json()) as { services: ServiceItem[] };
		return (data.services ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

/** Преобразует API-записи в формат для UI с учётом языка.
 *  Fallback: запрошенный язык → язык по умолчанию → любой доступный перевод. */
export function toServiceDisplayItems(
	services: ServiceItem[],
	lang: string
): ServiceDisplayItem[] {
	return [...services]
		.sort((a, b) => a.sort_order - b.sort_order)
		.map((item) => {
			const locale =
				item.translations[lang] ??
				item.translations[DEFAULT_LANG] ??
				Object.values(item.translations)[0];
			return {
				id: item.slug,
				title: locale?.title ?? item.slug,
				description: locale?.description ?? '',
				tags: item.tags ?? [],
				icon: item.icon
			};
		});
}
