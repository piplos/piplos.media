import { langStore } from '$lib/stores/lang.svelte';
import { SITE } from '$lib/site';
import type { LegalPage, LegalSlug } from '$lib/legal-api';

export type LegalSection = {
	title: string;
	body: string;
};

export type LegalDocument = {
	label: string;
	title: string;
	lastUpdated: string;
	sections: LegalSection[];
};

type LegalSectionKey = 'privacy.sections' | 'terms.sections' | 'cookies.sections';

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

function interpolateSections(sections: LegalSection[]): LegalSection[] {
	return sections.map((section) => ({
		title: interpolate(section.title),
		body: interpolate(section.body)
	}));
}

const SLUG_META: Record<
	LegalSlug,
	{ sectionKey: LegalSectionKey; labelKey: string; titleKey: string; lastUpdatedKey: string }
> = {
	privacy: {
		sectionKey: 'privacy.sections',
		labelKey: 'privacy.label',
		titleKey: 'privacy.title',
		lastUpdatedKey: 'privacy.last_updated'
	},
	terms: {
		sectionKey: 'terms.sections',
		labelKey: 'terms.label',
		titleKey: 'terms.title',
		lastUpdatedKey: 'terms.last_updated'
	},
	cookies: {
		sectionKey: 'cookies.sections',
		labelKey: 'cookies.label',
		titleKey: 'cookies.title',
		lastUpdatedKey: 'cookies.last_updated'
	}
};

function fromStatic(slug: LegalSlug): LegalDocument {
	const keys = SLUG_META[slug];
	return {
		label: langStore.t(keys.labelKey),
		title: langStore.t(keys.titleKey),
		lastUpdated: langStore.t(keys.lastUpdatedKey),
		sections: interpolateSections(langStore.get<LegalSection[]>(keys.sectionKey) ?? [])
	};
}

function fromApi(slug: LegalSlug, pages: LegalPage[]): LegalDocument | null {
	const page = pages.find((p) => p.slug === slug);
	const locale = page?.translations[langStore.value];
	if (!locale) return null;
	const hasContent =
		Boolean(locale.title?.trim()) ||
		locale.sections?.some((s) => s.title?.trim() || s.body?.trim());
	if (!hasContent) return null;
	return {
		label: interpolate(locale.label ?? ''),
		title: interpolate(locale.title ?? ''),
		lastUpdated: interpolate(locale.last_updated ?? ''),
		sections: interpolateSections(locale.sections ?? [])
	};
}

export function resolveLegalDocument(slug: LegalSlug, pages: LegalPage[]): LegalDocument {
	return fromApi(slug, pages) ?? fromStatic(slug);
}
