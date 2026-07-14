import {
	getApiV1,
	resolveUploadUrl,
	resolveUploadUrlsInHtml,
	type ApiRequestContext
} from '$lib/api';
import { DEFAULT_LANG } from '$lib/i18n/routing';
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
	image: string;
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

function toProjectLocale(
	data: Record<string, string> | undefined,
	ctx?: ApiRequestContext
): ProjectLocale {
	const src = data ?? {};
	const locale = Object.fromEntries(
		LOCALE_FIELDS.map((key) => [key, src[key] ?? ''])
	) as ProjectLocale;
	// solution приходит из API готовым HTML с относительными /uploads/ ссылками
	locale.solution = resolveUploadUrlsInHtml(locale.solution, ctx);
	return locale;
}

/** Преобразует запись API в формат сайта (id = slug для URL /portfolio/{slug}).
 *  Отсутствующий перевод заменяется языком по умолчанию, чтобы не отдавать пустые страницы. */
export function toPortfolioProject(project: ApiProject, ctx?: ApiRequestContext): PortfolioProject {
	const fallback = project.translations[DEFAULT_LANG];
	return {
		id: project.slug,
		category: project.category,
		categories: project.categories ?? [],
		tags: project.tags ?? [],
		year: project.year,
		featured: project.featured,
		image: resolveUploadUrl(project.image ?? '', ctx),
		en: toProjectLocale(project.translations.en ?? fallback, ctx),
		ru: toProjectLocale(project.translations.ru ?? fallback, ctx)
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
	query: ProjectsQuery = {},
	ctx?: ApiRequestContext
): Promise<ApiProject[]> {
	try {
		const res = await fetchFn(`${getApiV1(ctx)}/public/projects${projectsQueryString(query)}`);
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
	lang?: string,
	ctx?: ApiRequestContext
): Promise<PortfolioProject | null> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${getApiV1(ctx)}/public/projects/${encodeURIComponent(slug)}${qs}`);
		if (!res.ok) return null;
		const data = (await res.json()) as { project: ApiProject };
		return data.project ? toPortfolioProject(data.project, ctx) : null;
	} catch {
		return null;
	}
}

/** Загружает портфолио из API. Сортировка: sort_order, затем год по убыванию. */
export async function loadPortfolioProjects(
	fetchFn: FetchFn = fetch,
	query: ProjectsQuery = {},
	ctx?: ApiRequestContext
): Promise<PortfolioProject[]> {
	const fromApi = await fetchPortfolioProjects(fetchFn, query, ctx);
	return [...fromApi]
		.sort((a, b) => a.sort_order - b.sort_order || b.year - a.year)
		.map((project) => toPortfolioProject(project, ctx));
}

/** Slug-и для prerender entries(). */
export function portfolioProjectEntries(projects: PortfolioProject[]): { slug: string }[] {
	return projects.map((p) => ({ slug: p.id }));
}
