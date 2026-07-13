// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Locals {
			accessToken: string | null;
			refreshToken: string | null;
			user: { id: string; email: string; full_name: string; role: string } | null;
		}
	}
}

export {};
