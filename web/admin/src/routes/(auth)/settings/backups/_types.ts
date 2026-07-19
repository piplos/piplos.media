/** Типы раздела «Бекапы» (ответы Go API /v1/backups*). */

export type BackupType = 'full' | 'db' | 'files';
export type BackupStorage = 'local' | 's3';

export interface BackupArchive {
	name: string;
	type: BackupType;
	storage: BackupStorage;
	size: number;
	mod_time: string;
}

export interface BackupOpResult {
	op: 'backup' | 'restore';
	type?: BackupType;
	archive?: string;
	storage?: BackupStorage;
	ok: boolean;
	error?: string;
	started_at: string;
	finished_at: string;
	size_bytes?: number;
}

export interface BackupStatus {
	running: boolean;
	op?: 'backup' | 'restore';
	started_at?: string;
	last?: BackupOpResult;
}

export interface BackupSettingsForm {
	enabled: boolean;
	type: BackupType;
	intervalHours: string;
	keep: string;
	storage: BackupStorage;
}

export interface S3SettingsForm {
	endpoint: string;
	region: string;
	bucket: string;
	accessKeyId: string;
	accessKeyIdMasked: boolean;
	secretAccessKey: string;
	secretAccessKeyMasked: boolean;
	usePathStyle: boolean;
}
