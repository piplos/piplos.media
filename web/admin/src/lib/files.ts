/** Клиентские запросы к файловому архиву через прокси /api/files (браузер). */

export interface FolderInfo {
	name: string;
	path: string;
}

export interface FileInfo {
	name: string;
	path: string;
	url: string;
	size: number;
	mod_time: string;
}

export interface FileListing {
	path: string;
	folders: FolderInfo[];
	files: FileInfo[];
}

/** Ошибка файлового API с HTTP-статусом (404 — папки нет). */
export class FilesApiError extends Error {
	status: number;
	constructor(message: string, status: number) {
		super(message);
		this.status = status;
	}
}

async function parseOrThrow<T>(res: Response): Promise<T> {
	const data = (await res.json().catch(() => ({}))) as T & { message?: string };
	if (!res.ok) throw new FilesApiError(data.message ?? 'Ошибка запроса к файловому архиву', res.status);
	return data;
}

export async function listFiles(path: string): Promise<FileListing> {
	const res = await fetch(`/api/files?path=${encodeURIComponent(path)}`);
	return parseOrThrow<FileListing>(res);
}

async function filesAction<T>(action: string, body: unknown): Promise<T> {
	const res = await fetch(`/api/files?action=${action}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	return parseOrThrow<T>(res);
}

export function createFolder(path: string): Promise<{ path: string }> {
	return filesAction('folders', { path });
}

export function renameEntry(from: string, to: string): Promise<{ path: string; url: string }> {
	return filesAction('rename', { from, to });
}

export function moveEntries(paths: string[], dest: string): Promise<{ moved: string[] }> {
	return filesAction('move', { paths, dest });
}

export function deleteEntries(paths: string[]): Promise<{ deleted: number }> {
	return filesAction('delete', { paths });
}

export async function uploadFile(file: File, path: string): Promise<{ url: string; path: string }> {
	const fd = new FormData();
	fd.append('file', file);
	if (path) fd.append('path', path);
	fd.append('name', file.name);
	const res = await fetch('/api/upload', { method: 'POST', body: fd });
	return parseOrThrow<{ url: string; path: string }>(res);
}

const IMAGE_EXT = /\.(png|jpe?g|webp|gif|svg|avif)$/i;

export function isImageFile(name: string): boolean {
	return IMAGE_EXT.test(name);
}

export { formatBytes as formatSize } from './format';

/** Хлебные крошки для относительного пути: [{name, path}], без корня. */
export function pathCrumbs(path: string): { name: string; path: string }[] {
	if (!path) return [];
	const parts = path.split('/');
	return parts.map((name, i) => ({ name, path: parts.slice(0, i + 1).join('/') }));
}

/** Папка сущности в архиве: entityFolder('projects', 'site-dev') → 'projects/site-dev'.
 *  Slug чистится от запрещённых символов; пустой slug → корневая папка раздела. */
export function entityFolder(
	section: 'projects' | 'services' | 'pages' | 'legal',
	slug: string
): string {
	const clean = slug
		.trim()
		.replace(/[/\\:*?"<>|]/g, '-')
		.replace(/^\.+/, '')
		.slice(0, 128);
	return clean ? `${section}/${clean}` : section;
}
