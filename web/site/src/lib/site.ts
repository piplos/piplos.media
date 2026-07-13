export const SITE = {
	name: 'piplos.media',
	displayName: 'Piplos Media',
	url: 'https://piplos.media',
	email: 'info@piplos.media',
	phones: [
		{ display: '+375 (17) 249-98-97', tel: '+375172499897' },
		{ display: '+375 (29) 122-61-11', tel: '+375291226111' }
	],
	location: 'Minsk, Belarus',
	foundedAt: '2012-11-06'
} as const;

export function getCompanyYears(now = new Date()): number {
	const founded = new Date(SITE.foundedAt);
	let years = now.getFullYear() - founded.getFullYear();

	if (
		now.getMonth() < founded.getMonth() ||
		(now.getMonth() === founded.getMonth() && now.getDate() < founded.getDate())
	) {
		years--;
	}

	return years;
}
