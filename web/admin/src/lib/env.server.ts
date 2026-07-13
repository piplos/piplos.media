/** Базовый URL API backend'а (Go, Fiber). */
export function getApiBaseUrl(): string {
	return process.env.ADMIN_API_URL ?? 'http://localhost:3001';
}
