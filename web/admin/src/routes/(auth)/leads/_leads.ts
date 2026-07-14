import type { Lead, LeadStatus } from '$lib/types';

export const LEAD_STATUSES = ['new', 'in_progress', 'done', 'spam'] as const satisfies readonly LeadStatus[];

/** API status → сегмент URL. */
export const STATUS_TO_SLUG: Record<LeadStatus, string> = {
	new: 'new',
	in_progress: 'in-progress',
	done: 'done',
	spam: 'spam'
};

/** Сегмент URL → API status. */
export const SLUG_TO_STATUS: Record<string, LeadStatus> = {
	new: 'new',
	'in-progress': 'in_progress',
	done: 'done',
	spam: 'spam'
};

export function isLeadStatus(value: string): value is LeadStatus {
	return (LEAD_STATUSES as readonly string[]).includes(value);
}

export const LEAD_FILTERS: { value: '' | LeadStatus; label: string; href: string }[] = [
	{ value: '', label: 'Все', href: '/leads' },
	{ value: 'new', label: 'Новые', href: '/leads/new' },
	{ value: 'in_progress', label: 'В работе', href: '/leads/in-progress' },
	{ value: 'done', label: 'Завершённые', href: '/leads/done' },
	{ value: 'spam', label: 'Спам', href: '/leads/spam' }
];

export type LeadsPageData = {
	leads: Lead[];
	total: number;
	counts: Record<'' | LeadStatus, number>;
	status: '' | LeadStatus;
	page: number;
	error: string | null;
};

export type LeadBreadcrumb = { label: string; href?: string };

export function leadsBreadcrumbs(status: '' | LeadStatus): LeadBreadcrumb[] {
	if (!status) return [{ label: 'Заявки' }];
	const filter = LEAD_FILTERS.find((entry) => entry.value === status);
	return [{ label: 'Заявки', href: '/leads' }, { label: filter?.label ?? status }];
}
