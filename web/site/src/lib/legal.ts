import { langStore } from '$lib/stores/lang.svelte';
import { SITE } from '$lib/site';

export type LegalSection = {
	title: string;
	body: string;
};

export type LegalSectionKey = 'privacy.sections' | 'terms.sections' | 'cookies.sections';

const legalParams = () => ({
	company: SITE.displayName,
	email: SITE.email,
	site: SITE.url,
	location: langStore.t('site.location'),
	phones: SITE.phones.map((phone) => phone.display).join(', ')
});

function interpolate(template: string): string {
	const params = legalParams();
	return template.replace(/\{(\w+)\}/g, (match, name) =>
		name in params ? String(params[name as keyof typeof params]) : match
	);
}

export function getLegalSections(key: LegalSectionKey): LegalSection[] {
	const raw = langStore.get<LegalSection[]>(key) ?? [];
	return raw.map((section) => ({
		title: section.title,
		body: interpolate(section.body)
	}));
}
