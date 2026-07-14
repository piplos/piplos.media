import type { Project, Service } from '$lib/types';

export const ORPHAN_GROUP_ID = '__orphan__';
export const ORPHAN_URL_SLUG = 'unassigned';

export type ProjectFilter = { value: string; label: string; href: string };

export function serviceTitle(service: Service): string {
	const langs = Object.keys(service.translations);
	return service.translations['en']?.title ?? (langs.length ? service.translations[langs[0]]?.title : '') ?? service.slug;
}

export function computeProjectCounts(projects: Project[], services: Service[]): Record<string, number> {
	const serviceSlugs = new Set(services.map((s) => s.slug));
	const counts: Record<string, number> = { '': projects.length };
	for (const service of services) {
		counts[service.slug] = projects.filter((p) => p.category === service.slug).length;
	}
	counts[ORPHAN_GROUP_ID] = projects.filter((p) => !serviceSlugs.has(p.category)).length;
	return counts;
}

export function buildProjectFilters(services: Service[], counts: Record<string, number>): ProjectFilter[] {
	const sorted = [...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug));
	const filters: ProjectFilter[] = [{ value: '', label: 'Все', href: '/projects' }];
	for (const service of sorted) {
		filters.push({
			value: service.slug,
			label: serviceTitle(service),
			href: `/projects/${service.slug}`
		});
	}
	if ((counts[ORPHAN_GROUP_ID] ?? 0) > 0) {
		filters.push({
			value: ORPHAN_GROUP_ID,
			label: 'Без группы',
			href: `/projects/${ORPHAN_URL_SLUG}`
		});
	}
	return filters;
}

/** URL-сегмент услуги для проекта (включая «без группы»). */
export function serviceUrlSlug(category: string, serviceSlugs: Set<string>): string {
	if (serviceSlugs.has(category)) return category;
	return ORPHAN_URL_SLUG;
}

export function projectHref(project: Project, services: Service[]): string {
	const serviceSlugs = new Set(services.map((s) => s.slug));
	return `/projects/${serviceUrlSlug(project.category, serviceSlugs)}/${project.slug}`;
}

export function newProjectHref(category: string, services: Service[]): string {
	if (category && category !== ORPHAN_GROUP_ID) {
		return `/projects/${category}/new`;
	}
	const first = [...services].sort((a, b) => a.sort_order - b.sort_order || a.slug.localeCompare(b.slug))[0];
	return first ? `/projects/${first.slug}/new` : '/projects';
}

export type ProjectBreadcrumb = { label: string; href?: string };

export function projectsBreadcrumbs(category: string, filters: ProjectFilter[]): ProjectBreadcrumb[] {
	if (!category) return [{ label: 'Проекты' }];
	const filter = filters.find((entry) => entry.value === category);
	const label = filter?.label ?? (category === ORPHAN_GROUP_ID ? 'Без группы' : category);
	return [{ label: 'Проекты', href: '/projects' }, { label }];
}

export function projectEditBreadcrumbs(
	serviceLabel: string,
	serviceHref: string,
	currentLabel: string
): ProjectBreadcrumb[] {
	return [
		{ label: 'Проекты', href: '/projects' },
		{ label: serviceLabel, href: serviceHref },
		{ label: currentLabel }
	];
}

export function projectTitle(project: Project): string {
	const langs = Object.keys(project.translations);
	return project.translations['en']?.title ?? (langs.length ? project.translations[langs[0]]?.title : '') ?? project.slug;
}

export function categoryFromUrlSlug(slug: string, serviceSlugs: Set<string>): string | null {
	if (slug === ORPHAN_URL_SLUG) return ORPHAN_GROUP_ID;
	if (serviceSlugs.has(slug)) return slug;
	return null;
}
