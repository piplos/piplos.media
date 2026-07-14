import { API_V1 } from '$lib/api';
import { SERVICE_ICONS } from '$lib/constants/sections';
import { DEFAULT_LANG } from '$lib/i18n/routing';

export interface ServiceTranslation {
	title: string;
	description: string;
	body?: string;
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

export interface ServicePageItem {
	slug: string;
	title: string;
	description: string;
	body: string;
	tags: string[];
	icon: string;
}

type FetchFn = typeof fetch;

export function getServiceLocale(item: ServiceItem, lang: string): ServiceTranslation {
	return (
		item.translations[lang] ??
		item.translations[DEFAULT_LANG] ??
		Object.values(item.translations)[0] ?? {
			title: item.slug,
			description: '',
			body: ''
		}
	);
}

export function toServicePageItem(item: ServiceItem, lang: string): ServicePageItem {
	const locale = getServiceLocale(item, lang);
	return {
		slug: item.slug,
		title: locale.title,
		description: locale.description,
		body: locale.body ?? '',
		tags: item.tags ?? [],
		icon: item.icon || SERVICE_ICONS[item.slug] || '⬡'
	};
}

/** Опубликованные услуги для секции на главной.
 *  lang — вернуть только этот перевод (фильтрация на сервере). */
export async function fetchServices(fetchFn: FetchFn = fetch, lang?: string): Promise<ServiceItem[]> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${API_V1}/public/services${qs}`);
		if (!res.ok) return [];
		const data = (await res.json()) as { services: ServiceItem[] };
		return (data.services ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

/** Одна опубликованная услуга по slug или null. */
export async function fetchService(
	slug: string,
	fetchFn: FetchFn = fetch,
	lang?: string
): Promise<ServiceItem | null> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(
			`${API_V1}/public/services/${encodeURIComponent(slug)}${qs}`
		);
		if (!res.ok) return null;
		const data = (await res.json()) as { service: ServiceItem };
		return data.service?.published ? data.service : null;
	} catch {
		return null;
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
			const page = toServicePageItem(item, lang);
			return {
				id: page.slug,
				title: page.title,
				description: page.description,
				tags: page.tags,
				icon: page.icon
			};
		});
}

/** Список услуг для страниц сайта (только API). */
export async function loadServicePageItems(
	fetchFn: FetchFn = fetch,
	lang: string
): Promise<ServicePageItem[]> {
	const fromApi = await fetchServices(fetchFn, lang);
	return [...fromApi]
		.sort((a, b) => a.sort_order - b.sort_order)
		.map((item) => toServicePageItem(item, lang));
}

/** Одна услуга для страницы (только API). */
export async function loadServicePageItem(
	slug: string,
	fetchFn: FetchFn = fetch,
	lang: string
): Promise<ServicePageItem | null> {
	const fromApi = await fetchService(slug, fetchFn, lang);
	if (!fromApi) return null;
	return toServicePageItem(fromApi, lang);
}

/** Slug-и для prerender entries(). */
export function servicePageEntries(services: ServicePageItem[]): { slug: string }[] {
	return services.map((service) => ({ slug: service.slug }));
}
