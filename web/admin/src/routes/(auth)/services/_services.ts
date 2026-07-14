import type { Service } from '$lib/types';

export function serviceTitle(service: Service): string {
	const langs = Object.keys(service.translations);
	return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
}

export function sortServices(services: Service[]): Service[] {
	return [...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug));
}

export function findServiceBySlug(services: Service[], slug: string): Service | undefined {
	return services.find((service) => service.slug === slug);
}

export type ServiceBreadcrumb = { label: string; href?: string };

export function serviceEditBreadcrumbs(serviceLabel: string): ServiceBreadcrumb[] {
	return [{ label: 'Услуги', href: '/services' }, { label: serviceLabel }];
}
