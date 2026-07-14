import type { Actions } from './$types';
import { COMPOSITE_KEYS, makeTestAction, makeUpdateAction } from '../_ai-settings.server';

export const actions: Actions = {
	updateTranslation: makeUpdateAction('translation', COMPOSITE_KEYS.translation),
	testTranslation: makeTestAction('translation', '/v1/settings/test/translation', 100_000)
};
