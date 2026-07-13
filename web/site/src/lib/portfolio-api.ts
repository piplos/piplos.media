import { API_URL } from '$lib/api';
import staticPortfolio from '$lib/data/portfolio.json';
import type { PortfolioProject, ProjectLocale } from '$lib/portfolio';

export interface ApiProject {
	id: string;
	slug: string;
	category: string;
	categories: string[];
	tags: string[];
	year: number;
	featured: boolean;
	published: boolean;
	sort_order: number;
	translations: Record<string, Record<string, string>>;
}

const LOCALE_FIELDS = [
	'title',
	'subtitle',
	'description',
	'challenge',
	'solution',
	'result',
	'stack_detail'
] as const;

function toProjectLocale(data: Record<string, string> | undefined): ProjectLocale {
	const src = data ?? {};
	return Object.fromEntries(LOCALE_FIELDS.map((key) => [key, src[key] ?? ''])) as ProjectLocale;
}

/** Преобразует запись API в формат сайта (id = slug для URL /portfolio/{slug}). */
export function toPortfolioProject(project: ApiProject): PortfolioProject {
	return {
		id: project.slug,
		category: project.category,
		categories: project.categories ?? [],
		tags: project.tags ?? [],
		year: project.year,
		featured: project.featured,
		en: toProjectLocale(project.translations.en),
		ru: toProjectLocale(project.translations.ru)
	};
}

type FetchFn = typeof fetch;

/** Опубликованные проекты портфолио. */
export async function fetchPortfolioProjects(fetchFn: FetchFn = fetch): Promise<ApiProject[]> {
	try {
		const res = await fetchFn(`${API_URL}/api/v1/public/projects`);
		if (!res.ok) return [];
		const data = (await res.json()) as { projects: ApiProject[] };
		return (data.projects ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

/** Загружает портфолио из API или статический fallback. */
export async function loadPortfolioProjects(fetchFn: FetchFn = fetch): Promise<PortfolioProject[]> {
	const fromApi = await fetchPortfolioProjects(fetchFn);
	if (fromApi.length > 0) {
		return [...fromApi]
			.sort((a, b) => a.sort_order - b.sort_order)
			.map(toPortfolioProject);
	}
	return staticPortfolio as PortfolioProject[];
}

/** Slug-и для prerender entries(). */
export function portfolioProjectEntries(projects: PortfolioProject[]): { slug: string }[] {
	return projects.map((p) => ({ slug: p.id }));
}
