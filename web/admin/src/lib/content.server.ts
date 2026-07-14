/** Общие парсеры форм контента (сервер). */
import type { LegalTranslations, Project, Service, Translations } from '$lib/types';

export function parseTranslations(fd: FormData, field = 'translations'): Translations {
	try {
		const raw = fd.get(field)?.toString() ?? '{}';
		const parsed = JSON.parse(raw) as Translations;
		return typeof parsed === 'object' && parsed !== null ? parsed : {};
	} catch {
		return {};
	}
}

export function parseList(fd: FormData, name: string): string[] {
	return (fd.get(name)?.toString() ?? '')
		.split(',')
		.map((s) => s.trim())
		.filter(Boolean);
}

export function parseIntField(fd: FormData, name: string, fallback = 0): number {
	const n = Number(fd.get(name)?.toString() ?? '');
	return Number.isFinite(n) ? Math.trunc(n) : fallback;
}

export function projectPayload(fd: FormData) {
	return {
		slug: fd.get('slug')?.toString().trim() ?? '',
		category: fd.get('category')?.toString().trim() ?? '',
		tags: parseList(fd, 'tags'),
		year: parseIntField(fd, 'year', new Date().getFullYear()),
		featured: fd.get('featured') === 'on',
		published: fd.get('published') === 'on',
		image: fd.get('image')?.toString().trim() ?? '',
		translations: parseTranslations(fd)
	};
}

export function firstServiceSlug(services: Service[]): string | null {
	const sorted = [...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug));
	return sorted[0]?.slug ?? null;
}

export function nextProjectSortOrder(projects: Project[], category: string): number {
	const inGroup = projects.filter((project) => project.category === category);
	if (!inGroup.length) return 0;
	return Math.min(...inGroup.map((project) => project.sort_order)) - 1;
}

export function projectSaveBody(
	payload: ReturnType<typeof projectPayload>,
	options: { existing?: Project; services: Service[]; projects?: Project[] }
) {
	const category =
		payload.category && options.services.some((service) => service.slug === payload.category)
			? payload.category
			: options.existing?.category ?? firstServiceSlug(options.services);

	if (!category) return null;

	if (options.existing) {
		return {
			...payload,
			category,
			categories: [category],
			sort_order: options.existing.sort_order
		};
	}

	return {
		...payload,
		category,
		categories: [category],
		sort_order: nextProjectSortOrder(options.projects ?? [], category)
	};
}

export function servicePayload(fd: FormData) {
	return {
		slug: fd.get('slug')?.toString().trim() ?? '',
		icon: fd.get('icon')?.toString().trim() ?? '',
		tags: parseList(fd, 'tags'),
		published: fd.get('published') === 'on',
		sort_order: parseIntField(fd, 'sort_order'),
		translations: parseTranslations(fd)
	};
}

export function seoPayload(fd: FormData) {
	const path = fd.get('seo_path')?.toString().trim() ?? '';
	if (!path) return null;
	return {
		id: fd.get('seo_id')?.toString() ?? '',
		path,
		translations: parseTranslations(fd, 'seo_translations')
	};
}

function seoTranslationsFilled(translations: Translations): boolean {
	return Object.values(translations).some((fields) =>
		Object.values(fields).some((value) => value.trim() !== '')
	);
}

export function shouldSaveSeo(seo: { id: string; translations: Translations }): boolean {
	return Boolean(seo.id) || seoTranslationsFilled(seo.translations);
}

export function legalPayload(fd: FormData): { translations: LegalTranslations } {
	try {
		const raw = fd.get('translations')?.toString() ?? '{}';
		const parsed = JSON.parse(raw) as LegalTranslations;
		return { translations: typeof parsed === 'object' && parsed !== null ? parsed : {} };
	} catch {
		return { translations: {} };
	}
}

export function stackPayload(fd: FormData) {
	return {
		slug: fd.get('slug')?.toString().trim() ?? '',
		label: fd.get('label')?.toString().trim() ?? '',
		icon: fd.get('icon')?.toString().trim() ?? '',
		icon_alt: fd.get('icon_alt')?.toString().trim() ?? '',
		published: fd.get('published') === 'on'
	};
}
