import { getApiV1, resolveUploadUrl, type ApiRequestContext } from '$lib/api';
import { DEFAULT_LANG } from '$lib/i18n/routing';

export interface ArticleTranslation {
	title?: string;
	description?: string;
	/** HTML: Markdown рендерится на стороне API. */
	body?: string;
}

/** Пользовательская страница из админки (раздел «Статьи»). */
export interface ArticleItem {
	id: string;
	slug: string;
	published: boolean;
	publish_at: string | null;
	/** Превью: фон карточки в списке (как у проектов). */
	image: string;
	/** Технологический стек; пустой — не показывать в сайдбаре. */
	tags: string[];
	translations: Record<string, ArticleTranslation>;
	created_at: string;
}

type FetchFn = typeof fetch;

export function getArticleLocale(item: ArticleItem, lang: string): ArticleTranslation {
	return (
		item.translations[lang] ??
		item.translations[DEFAULT_LANG] ??
		Object.values(item.translations)[0] ?? { title: item.slug }
	);
}

/** Дата публикации статьи для отображения. */
export function articleDate(item: ArticleItem): string {
	return item.publish_at ?? item.created_at;
}

export function formatArticleDate(iso: string, lang: string): string {
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return '';
	return d.toLocaleDateString(lang === 'ru' ? 'ru-RU' : 'en-US', {
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	});
}

function normalizeArticle(raw: ArticleItem, ctx?: ApiRequestContext): ArticleItem {
	return {
		...raw,
		image: resolveUploadUrl(raw.image ?? '', ctx),
		tags: raw.tags ?? []
	};
}

/** Опубликованные статьи (сервер уже скрывает черновики и отложенные).
 *  lang — вернуть только этот перевод. */
export async function fetchArticles(
	fetchFn: FetchFn = fetch,
	lang?: string,
	ctx?: ApiRequestContext
): Promise<ArticleItem[]> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(`${getApiV1(ctx)}/public/pages${qs}`);
		if (!res.ok) return [];
		const data = (await res.json()) as { pages: ArticleItem[] };
		return (data.pages ?? []).map((p) => normalizeArticle(p, ctx));
	} catch {
		return [];
	}
}

/** Одна опубликованная статья по slug или null. */
export async function fetchArticle(
	slug: string,
	fetchFn: FetchFn = fetch,
	lang?: string,
	ctx?: ApiRequestContext
): Promise<ArticleItem | null> {
	try {
		const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
		const res = await fetchFn(
			`${getApiV1(ctx)}/public/pages/${encodeURIComponent(slug)}${qs}`
		);
		if (!res.ok) return null;
		const data = (await res.json()) as { page: ArticleItem };
		return data.page ? normalizeArticle(data.page, ctx) : null;
	} catch {
		return null;
	}
}
