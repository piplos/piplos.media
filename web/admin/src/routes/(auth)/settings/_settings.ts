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
		return [{ label: 'Настройки', href: '/settings' }, { label: 'SMTP' }];
	}

	if (pathname === '/settings/users' || pathname.startsWith('/settings/users/')) {
		return [{ label: 'Настройки', href: '/settings' }, { label: 'Пользователи' }];
	}

	return [{ label: 'Настройки' }];
}
