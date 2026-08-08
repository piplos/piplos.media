import { resolveUploadUrl, preferUploadWebp } from '$lib/api';

const WEBP_RE = /\.webp(?=$|[?#])/i;
const SIZED_WEBP_RE = /-(480|960)\.webp(?=$|[?#])/i;

/** file.webp → file-480.webp (idempotent if already sized). */
export function uploadSizedWebp(url: string, width: 480 | 960): string {
	const webp = preferUploadWebp(url);
	if (!webp.includes('/uploads/') || !WEBP_RE.test(webp)) return webp;
	const master = webp.replace(SIZED_WEBP_RE, '.webp');
	return master.replace(WEBP_RE, `-${width}.webp`);
}

export interface UploadCardImage {
	/** Fallback / default src (full WebP master). */
	src: string;
	srcset: string;
	sizes: string;
}

/** Превью карточек портфолио/статей: full + 480/960 WebP. */
export function uploadCardImage(
	url: string,
	opts: { sizes?: string } = {}
): UploadCardImage | null {
	if (!url) return null;
	const resolved = resolveUploadUrl(url);
	const sizes = opts.sizes ?? '(max-width: 768px) 100vw, 640px';
	if (!resolved.includes('/uploads/')) {
		return { src: resolved, srcset: '', sizes };
	}
	const master = preferUploadWebp(resolved);
	if (!WEBP_RE.test(master)) {
		// GIF/SVG uploads — plain src, no fake multi-width srcset.
		return { src: master, srcset: '', sizes };
	}
	const w480 = uploadSizedWebp(master, 480);
	const w960 = uploadSizedWebp(master, 960);
	return {
		src: master,
		srcset: `${w480} 480w, ${w960} 960w`,
		sizes
	};
}
