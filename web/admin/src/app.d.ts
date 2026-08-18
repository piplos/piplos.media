// See https://svelte.dev/docs/kit/types#app.d.ts
import type { AdminUser } from './lib/types';

declare global {
	namespace App {
		interface Locals {
			accessToken: string | null;
			refreshToken: string | null;
			user: AdminUser | null;
		}

		interface Platform {
			env?: {
				ADMIN_API_URL?: string;
			};
		}
	}
}

export {};
