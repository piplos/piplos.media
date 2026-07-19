/** Общие опции селектов и подписи раздела «Бекапы». */
import type { BackupStorage, BackupType } from './_types';

export const typeOptions = [
	{ value: 'full', label: 'Полный (база + файлы)' },
	{ value: 'db', label: 'Только база (SQL)' },
	{ value: 'files', label: 'Только файлы' }
];

export const storageOptions = [
	{ value: 'local', label: 'Локально (на сервере)' },
	{ value: 's3', label: 'S3 (Cloudflare R2)' }
];

export function typeLabel(type: BackupType): string {
	if (type === 'db') return 'База';
	if (type === 'files') return 'Файлы';
	return 'Полный';
}

export function typeBadgeVariant(type: BackupType): string {
	if (type === 'db') return 'info';
	if (type === 'files') return 'warning';
	return 'success';
}

export function storageLabel(storage: BackupStorage): string {
	return storage === 's3' ? 'S3' : 'Локально';
}
