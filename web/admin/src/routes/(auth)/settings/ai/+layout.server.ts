import type { LayoutServerLoad } from './$types';
import { loadAiSettings } from './_ai-settings.server';

export const load: LayoutServerLoad = loadAiSettings;
