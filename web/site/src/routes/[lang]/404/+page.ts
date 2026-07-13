import { SUPPORTED_LANGS } from '$lib/i18n/routing';
import type { EntryGenerator } from './$types';

export const entries: EntryGenerator = () =>
	SUPPORTED_LANGS.map((lang) => ({ lang }));
