import type { Lang } from '$lib/stores/lang.svelte';

export type ProjectLocale = {
	title: string;
	subtitle: string;
	description: string;
	challenge: string;
	solution: string;
	result: string;
	stack_detail: string;
};

export type PortfolioProject = {
	id: string;
	category: string;
	categories: string[];
	tags: string[];
	year: number;
	featured: boolean;
	image: string;
	en: ProjectLocale;
	ru: ProjectLocale;
};

const CATEGORY_COLORS: Record<string, string> = {
	saas: 'var(--c-accent3)',
	web: 'var(--c-accent3)',
	mobile: 'var(--c-accent2)',
	fintech: 'var(--c-accent)',
	ecommerce: 'var(--c-accent2)',
	devops: 'var(--c-muted)'
};

export function getProjectLocale(project: PortfolioProject, lang: Lang): ProjectLocale {
	return project[lang];
}

export function getCategoryColor(category: string): string {
	return CATEGORY_COLORS[category] ?? 'var(--c-accent)';
}

export function getProjectStackItems(project: PortfolioProject, lang: Lang): string[] {
	if (project.tags?.length) return project.tags;
	return getProjectLocale(project, lang)
		.stack_detail.split(',')
		.map((item) => item.trim())
		.filter(Boolean);
}
