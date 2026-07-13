import { error } from '@sveltejs/kit';
import portfolioData from '$lib/data/portfolio.json';
import type { PageLoad } from './$types';

export const prerender = true;

export function entries() {
	return portfolioData.map((p) => ({ id: p.id }));
}

export const load: PageLoad = ({ params }) => {
	const project = portfolioData.find((p) => p.id === params.id);
	if (!project) throw error(404, 'Project not found');
	return { project };
};
