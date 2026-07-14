import type { PageServerLoad } from './$types';
import { createProjectsLoad, projectsActions } from './_projects.server';

export const load: PageServerLoad = createProjectsLoad('');
export const actions = projectsActions;
