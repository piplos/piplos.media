import type { Lang } from '$lib/stores/lang.svelte';
import { getProjectLocale, getProjectStackItems, type PortfolioProject } from '$lib/portfolio';
import { SITE } from '$lib/site';

const CATEGORY_TO_TYPES: Record<string, string[]> = {
	web: ['web'],
	saas: ['saas', 'web'],
	mobile: ['mobile'],
	fintech: ['mobile', 'backend'],
	ecommerce: ['web', 'backend'],
	devops: ['devops']
};

export type OrderPrefill = {
	types: string[];
	projectName: string;
	description: string;
	stack: string;
	references: string;
};

export function getOrderPrefillFromProject(project: PortfolioProject, lang: Lang): OrderPrefill {
	const loc = getProjectLocale(project, lang);
	const types = new Set<string>();

	for (const category of project.categories ?? [project.category]) {
		for (const type of CATEGORY_TO_TYPES[category] ?? []) {
			types.add(type);
		}
	}

	if (types.size === 0) types.add('web');

	const projectName =
		lang === 'ru'
			? `Проект по аналогии с «${loc.title}»`
			: `Project similar to "${loc.title}"`;

	const description = [loc.description, loc.challenge].filter(Boolean).join('\n\n');

	return {
		types: [...types],
		projectName,
		description,
		stack: getProjectStackItems(project, lang).join(', '),
		references: `${SITE.url}/portfolio/${project.id}`
	};
}
