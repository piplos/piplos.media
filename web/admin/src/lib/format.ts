/** Форматирует размер в байтах в компактный вид («1.5 МБ», «12 ГБ»). */
export function formatBytes(bytes: number): string {
	const units = ['Б', 'КБ', 'МБ', 'ГБ', 'ТБ'];
	let v = bytes;
	let i = 0;
	while (v >= 1024 && i < units.length - 1) {
		v /= 1024;
		i++;
	}
	return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}

/** Форматирует ISO-дату в компактный вид для таблиц. */
export function formatDate(iso: string): string {
	if (!iso) return '—';
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return '—';
	return d.toLocaleDateString('ru-RU', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	});
}
