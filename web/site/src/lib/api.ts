import { env } from '$env/dynamic/public';
import { dev } from '$app/environment';

// Static build: значение PUBLIC_API_URL фиксируется во время сборки (Cloudflare Pages).
export const API_URL = env.PUBLIC_API_URL ?? (dev ? 'http://localhost:3001' : 'https://api.piplos.media');
