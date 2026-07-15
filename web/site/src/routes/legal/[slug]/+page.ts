export const prerender = true;

import { LEGAL_SLUGS } from '$lib/legal-api';
import type { EntryGenerator } from './$types';

export const entries: EntryGenerator = () =>
	LEGAL_SLUGS.map((slug) => ({ slug }));
