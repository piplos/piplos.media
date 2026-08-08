import {
	getApiV1,
	resolveUploadUrl,
	resolveUploadUrlsInHtml,
	type ApiRequestContext
} from '$lib/api';
import {
	collectEmbedItems,
	embedParamsToProjectsQuery,
	selectProjects,
	type EmbedProjectsQuery
} from '$lib/embeds';
import { DEFAULT_LANG } from '$lib/i18n/routing';
import type { PortfolioProject, ProjectLocale, ProjectLink } from '$lib/portfolio';

export interface ApiProject {
	id: string;
	slug: string;
	category: string;
	categories: string[];
	tags: string[];
	year: number;
	featured: boolean;
	published: boolean;
	/** Порядок внутри группы (услуги). */
	sort_order: number;
	/** Сквозной порядок раздела «все проекты» (задаётся в админке). */
	global_sort_order: number;
	image: string;
	links?: ProjectLink[];
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

/** `full` — страница кейса; `summary` — списки/эмбеды без тяжёлого HTML (solution/result). */
export type ProjectPayloadMode = 'full' | 'summary';

function toProjectLocale(
	data: Record<string, string> | undefined,
	ctx?: ApiRequestContext,
	mode: ProjectPayloadMode = 'full'
): ProjectLocale {
	const src = data ?? {};
	const locale = Object.fromEntries(
		LOCALE_FIELDS.map((key) => [key, src[key] ?? ''])
	) as ProjectLocale;
	if (mode === 'summary') {
		// Списки/эмбеды: title/subtitle/description/challenge достаточно; solution тянет
		// сотни KiB HTML с <img> в data-payload и раздувает документ до 1+ MiB.
		locale.solution = '';
		locale.result = '';
		return locale;
	}
	// solution приходит из API готовым HTML с относительными /uploads/ ссылками
	locale.solution = resolveUploadUrlsInHtml(locale.solution, ctx);
	return locale;
}

/** Преобразует запись API в формат сайта (id = slug для URL /portfolio/{slug}).
 *  Отсутствующий перевод заменяется языком по умолчанию, чтобы не отдавать пустые страницы. */
export function toPortfolioProject(
	project: ApiProject,
	ctx?: ApiRequestContext,
	mode: ProjectPayloadMode = 'full'
): PortfolioProject {
	const fallback = project.translations[DEFAULT_LANG];
	return {
		id: project.slug,
		category: project.category,
		categories: project.categories ?? [],
		tags: project.tags ?? [],
		year: project.year,
		featured: project.featured,
		sort_order: project.sort_order,
		image: resolveUploadUrl(project.image ?? '', ctx),
		links: Array.isArray(project.links) ? project.links : [],
		en: toProjectLocale(project.translations.en ?? fallback, ctx, mode),
		ru: toProjectLocale(project.translations.ru ?? fallback, ctx, mode)
	};
}

type FetchFn = typeof fetch;

export interface ProjectsQuery extends Omit<EmbedProjectsQuery, 'limit' | 'mode'> {
	/** Максимум записей (сервер режет ответ). */
	limit?: number;
	/** Без solution/result HTML — для списков, эмбедов и префилла заказа. */
	mode?: ProjectPayloadMode;
}

function projectsQueryString(query: ProjectsQuery = {}): string {
	const params = new URLSearchParams();
	if (query.lang) params.set('lang', query.lang);
	if (query.featured) params.set('featured', 'true');
	if (query.category) params.set('category', query.category);
	if (query.tags?.length) params.set('tags', query.tags.join(','));
	if (query.slugs?.length) params.set('slugs', query.slugs.join(','));
	if (query.limit && query.limit > 0) params.set('limit', String(query.limit));
	if (query.mode === 'summary') params.set('mode', 'summary');
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

/** Загружает портфолио из API. Порядок задаёт бэкенд:
 *  category → group sort_order; slugs → порядок запроса;
 *  иначе ListProjects (global_sort_order, year DESC). Для UI-фильтра
 *  внутри группы используйте sortProjectsByGroupOrder. */
export async function loadPortfolioProjects(
	fetchFn: FetchFn = fetch,
	query: ProjectsQuery = {},
	ctx?: ApiRequestContext
): Promise<PortfolioProject[]> {
	const mode = query.mode ?? 'full';
	const fromApi = await fetchPortfolioProjects(fetchFn, query, ctx);
	return fromApi.map((project) => toPortfolioProject(project, ctx, mode));
}

/** Порядок внутри группы (sort_order из админки), затем год по убыванию. */
export function sortProjectsByGroupOrder(projects: PortfolioProject[]): PortfolioProject[] {
	return [...projects].sort((a, b) => a.sort_order - b.sort_order || b.year - a.year);
}

/** Связанные проекты услуги: дозированный запрос category+limit на API. */
export async function loadRelatedProjectsForCategory(
	category: string,
	fetchFn: FetchFn = fetch,
	opts: { lang?: string; limit?: number } = {},
	ctx?: ApiRequestContext
): Promise<PortfolioProject[]> {
	return loadPortfolioProjects(
		fetchFn,
		{
			lang: opts.lang,
			category,
			limit: opts.limit ?? 3,
			mode: 'summary'
		},
		ctx
	);
}

/** Связанные проекты статьи: tags+limit, иначе первые N сквозного порядка. */
export async function loadRelatedProjectsForTags(
	tags: string[],
	fetchFn: FetchFn = fetch,
	opts: { lang?: string; limit?: number } = {},
	ctx?: ApiRequestContext
): Promise<PortfolioProject[]> {
	const limit = opts.limit ?? 3;
	const clean = tags.filter(Boolean);
	if (clean.length > 0) {
		const matched = await loadPortfolioProjects(
			fetchFn,
			{ lang: opts.lang, tags: clean, limit, mode: 'summary' },
			ctx
		);
		if (matched.length > 0) return matched;
	}
	return loadPortfolioProjects(fetchFn, { lang: opts.lang, limit, mode: 'summary' }, ctx);
}

/** Проекты только под `{{projects …}}` в HTML — по одному дозированному запросу на уникальный токен. */
export async function loadProjectsForEmbedHtml(
	html: string,
	fetchFn: FetchFn = fetch,
	opts: { lang?: string } = {},
	ctx?: ApiRequestContext
): Promise<Record<string, PortfolioProject[]>> {
	return collectEmbedItems(
		html,
		'projects',
		(params) => loadPortfolioProjects(fetchFn, embedParamsToProjectsQuery(params, opts), ctx),
		selectProjects
	);
}

/** Slug-и для prerender entries(). */
export function portfolioProjectEntries(projects: PortfolioProject[]): { slug: string }[] {
	return projects.map((p) => ({ slug: p.id }));
}
