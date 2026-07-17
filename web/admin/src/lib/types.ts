export interface AdminUser {
	id: string;
	email: string;
	full_name: string;
	role: 'admin' | 'manager';
	is_active: boolean;
	/** Получать письма о новых заявках (меняет только админ). */
	notify_leads: boolean;
	created_at: string;
	updated_at: string;
}

export interface Language {
	code: string;
	name: string;
	is_default: boolean;
	enabled: boolean;
	sort_order: number;
}

export type Translations = Record<string, Record<string, string>>;

export interface Project {
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
	/** Сквозной порядок в разделе «все проекты» на сайте. */
	global_sort_order: number;
	image: string;
	translations: Translations;
	created_at: string;
	updated_at: string;
}

export interface Service {
	id: string;
	slug: string;
	icon: string;
	tags: string[];
	published: boolean;
	sort_order: number;
	translations: Translations;
	created_at: string;
	updated_at: string;
}

export interface StackItem {
	id: string;
	slug: string;
	label: string;
	icon: string;
	icon_alt: string;
	group_id: string;
	published: boolean;
	sort_order: number;
	created_at: string;
	updated_at: string;
}

export interface SEOPage {
	id: string;
	path: string;
	translations: Translations;
	created_at: string;
	updated_at: string;
}

export interface LegalSection {
	title: string;
	body: string;
}

export interface LegalLocale {
	label: string;
	title: string;
	last_updated: string;
	sections: LegalSection[];
}

export type LegalTranslations = Record<string, LegalLocale>;

export interface LegalPage {
	id: string;
	slug: string;
	path: string;
	sort_order: number;
	translations: LegalTranslations;
	created_at: string;
	updated_at: string;
}

export const LEGAL_SLUG_LABELS: Record<string, string> = {
	privacy: 'Персональные данные',
	terms: 'Пользовательское соглашение',
	cookies: 'Cookie'
};

/** Пользовательская страница (раздел «Статьи» на сайте). */
export interface Page {
	id: string;
	slug: string;
	published: boolean;
	/** ISO-дата отложенной публикации или null (сразу). */
	publish_at: string | null;
	/** Превью: фон карточки в списке статей. */
	image: string;
	/** Технологический стек (метки из каталога стека). */
	tags: string[];
	translations: Translations;
	created_at: string;
	updated_at: string;
}

export type PageStatus = 'draft' | 'scheduled' | 'published';

/** Статус страницы: черновик / запланирована / опубликована. */
export function pageStatus(page: Pick<Page, 'published' | 'publish_at'>, now = new Date()): PageStatus {
	if (!page.published) return 'draft';
	if (page.publish_at && new Date(page.publish_at) > now) return 'scheduled';
	return 'published';
}

export const PAGE_STATUS_LABELS: Record<PageStatus, string> = {
	draft: 'Черновик',
	scheduled: 'Запланирована',
	published: 'Опубликована'
};

export const PAGE_STATUS_VARIANTS: Record<PageStatus, 'neutral' | 'warning' | 'success'> = {
	draft: 'neutral',
	scheduled: 'warning',
	published: 'success'
};

export type LeadStatus = 'new' | 'in_progress' | 'done' | 'spam';

export interface Lead {
	id: string;
	types: string[];
	project_name: string;
	description: string;
	stack: string;
	references: string;
	budget: number;
	currency: string;
	timeline: string;
	stage: string;
	first_name: string;
	last_name: string;
	email: string;
	company: string;
	phone: string;
	how_found: string;
	notes: string;
	lang: string;
	status: LeadStatus;
	created_at: string;
	updated_at: string;
}

export interface Setting {
	key: string;
	value: string;
	updated_at: string;
}

export const LEAD_STATUS_LABELS: Record<LeadStatus, string> = {
	new: 'Новая',
	in_progress: 'В работе',
	done: 'Завершена',
	spam: 'Спам'
};

export const LEAD_STATUS_ORDER: LeadStatus[] = ['new', 'in_progress', 'done', 'spam'];

export const LEAD_STATUS_VARIANTS: Record<LeadStatus, 'info' | 'warning' | 'success' | 'danger'> = {
	new: 'info',
	in_progress: 'warning',
	done: 'success',
	spam: 'danger'
};

export function nextLeadStatus(current: LeadStatus): LeadStatus {
	const index = LEAD_STATUS_ORDER.indexOf(current);
	if (index < 0) return LEAD_STATUS_ORDER[0];
	return LEAD_STATUS_ORDER[(index + 1) % LEAD_STATUS_ORDER.length];
}
