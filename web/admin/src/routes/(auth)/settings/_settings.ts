export type SettingsBreadcrumb = { label: string; href?: string };

export function settingsBreadcrumbs(pathname: string): SettingsBreadcrumb[] {
	if (pathname === '/settings') {
		return [{ label: 'Настройки', href: '/settings' }, { label: 'Общие' }];
	}

	if (pathname === '/settings/ai' || pathname.startsWith('/settings/ai/')) {
		if (pathname === '/settings/ai/translation') {
			return [
				{ label: 'Настройки', href: '/settings' },
				{ label: 'AI-переводчик', href: '/settings/ai' },
				{ label: 'Перевод' }
			];
		}
		return [{ label: 'Настройки', href: '/settings' }, { label: 'AI-переводчик' }];
	}

	if (pathname === '/settings/smtp' || pathname.startsWith('/settings/smtp/')) {
		if (pathname === '/settings/smtp/template') {
			return [
				{ label: 'Настройки', href: '/settings' },
				{ label: 'SMTP', href: '/settings/smtp' },
				{ label: 'Шаблон письма' }
			];
		}
		return [{ label: 'Настройки', href: '/settings' }, { label: 'SMTP' }];
	}

	if (pathname === '/settings/backups' || pathname.startsWith('/settings/backups/')) {
		if (pathname === '/settings/backups/schedule') {
			return [
				{ label: 'Настройки', href: '/settings' },
				{ label: 'Бекапы', href: '/settings/backups' },
				{ label: 'Расписание' }
			];
		}
		if (pathname === '/settings/backups/s3') {
			return [
				{ label: 'Настройки', href: '/settings' },
				{ label: 'Бекапы', href: '/settings/backups' },
				{ label: 'S3' }
			];
		}
		return [
			{ label: 'Настройки', href: '/settings' },
			{ label: 'Бекапы', href: '/settings/backups' },
			{ label: 'Архивы' }
		];
	}

	if (pathname === '/settings/users' || pathname.startsWith('/settings/users/')) {
		return [{ label: 'Настройки', href: '/settings' }, { label: 'Пользователи' }];
	}

	if (pathname === '/settings/media' || pathname.startsWith('/settings/media/')) {
		if (pathname === '/settings/media/image-search') {
			return [
				{ label: 'Настройки', href: '/settings' },
				{ label: 'Медиа', href: '/settings/media' },
				{ label: 'Поиск PNG/JPG' }
			];
		}
		return [{ label: 'Настройки', href: '/settings' }, { label: 'Медиа' }];
	}

	if (pathname === '/settings/api-keys' || pathname.startsWith('/settings/api-keys/')) {
		return [{ label: 'Настройки', href: '/settings' }, { label: 'API-ключи' }];
	}

	return [{ label: 'Настройки' }];
}

