import { DEFAULT_LANG } from '$lib/i18n/routing';
import { servicePageEntries, loadServicePageItems } from '$lib/services-api';
import type { EntryGenerator } from './$types';

export const entries: EntryGenerator = async () =>
	servicePageEntries(await loadServicePageItems(fetch, DEFAULT_LANG));
