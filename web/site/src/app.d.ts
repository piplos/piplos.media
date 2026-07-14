// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Platform {
			env?: {
				PUBLIC_API_URL?: string;
			};
		}
	}
}

export {};
