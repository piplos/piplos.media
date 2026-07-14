/** Ключи подстановок в legal-текстах. Должны совпадать с legalParams() на публичном сайте. */
export const LEGAL_PLACEHOLDERS = [
	{
		key: 'company',
		label: 'Название компании',
		example: 'Piplos Media'
	},
	{
		key: 'email',
		label: 'Email для связи',
		example: 'info@piplos.media'
	},
	{
		key: 'site',
		label: 'URL сайта',
		example: 'https://piplos.media'
	},
	{
		key: 'location',
		label: 'Местоположение (зависит от языка страницы)',
		example: 'Минск, Беларусь / Minsk, Belarus'
	},
	{
		key: 'phones',
		label: 'Телефоны (через запятую)',
		example: '+375 (17) 249-98-97, +375 (29) 122-61-11'
	}
] as const;

export function legalPlaceholderToken(key: string): string {
	return `{${key}}`;
}
