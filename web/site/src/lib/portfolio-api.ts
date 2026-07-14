import { API_URL } from '$lib/api';
import { DEFAULT_LANG } from '$lib/i18n/routing';
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

/** Преобразует запись API в формат сайта (id = slug для URL /portfolio/{slug}).
 *  Отсутствующий перевод заменяется языком по умолчанию, чтобы не отдавать пустые страницы. */
export function toPortfolioProject(project: ApiProject): PortfolioProject {
	const fallback = project.translations[DEFAULT_LANG];
	return {
		id: project.slug,
		category: project.category,
		categories: project.categories ?? [],
		tags: project.tags ?? [],
		year: project.year,
		featured: project.featured,
		en: toProjectLocale(project.translations.en ?? fallback),
		ru: toProjectLocale(project.translations.ru ?? fallback)
	};
}

type FetchFn = typeof fetch;

export interface ProjectsQuery {
	/** Вернуть только этот перевод (payload меньше «пачки» всех языков). */
	lang?: string;
	/** Только featured-проекты (для главной). */
	featured?: boolean;
}

function projectsQueryString(query: ProjectsQuery = {}): string {
	const params = new URLSearchParams();
	if (query.lang) params.set('lang', query.lang);
	if (query.featured) params.set('featured', 'true');
	const qs = params.toString();
	return qs ? `?${qs}` : '';
}

/** Опубликованные проекты портфолио (фильтрация по языку/featured — на сервере). */
export async function fetchPortfolioProjects(
	fetchFn: FetchFn = fetch,
	query: ProjectsQuery = {}
): Promise<ApiProject[]> {
	try {
		const res = await fetchFn(`${API_URL}/api/v1/public/projects${projectsQueryString(query)}`);
		if (!res.ok) return [];
		const data = (await res.json()) as { projects: ApiProject[] };
		return (data.projects ?? []).filter((item) => item.published);
	} catch {
		return [];
	}
}

/** Один опубликованный проект по slug или null. */
export async function fetchPortfolioProject(
	slug: string,
	fetchFn: FetchFn = fetch,
	lang?: string
): Promise<PortfolioProject | null> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${API_URL}/api/v1/public/projects/${encodeURIComponent(slug)}${qs}`);
		if (!res.ok) return null;
		const data = (await res.json()) as { project: ApiProject };
		return data.project ? toPortfolioProject(data.project) : null;
	} catch {
		return null;
	}
}

/** Загружает портфолио из API или статический fallback.
 *  Сортировка повторяет API: sort_order, затем год по убыванию. */
export async function loadPortfolioProjects(
	fetchFn: FetchFn = fetch,
	query: ProjectsQuery = {}
): Promise<PortfolioProject[]> {
	const fromApi = await fetchPortfolioProjects(fetchFn, query);
	if (fromApi.length > 0) {
		return [...fromApi]
			.sort((a, b) => a.sort_order - b.sort_order || b.year - a.year)
			.map(toPortfolioProject);
	}
	const fallback = staticPortfolio as PortfolioProject[];
	return query.featured ? fallback.filter((p) => p.featured) : fallback;
}

/** Slug-и для prerender entries(). */
export function portfolioProjectEntries(projects: PortfolioProject[]): { slug: string }[] {
	return projects.map((p) => ({ slug: p.id }));
}
