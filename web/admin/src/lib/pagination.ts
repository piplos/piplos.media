/**
 * Номера страниц с разрывами для постраничной навигации: 1 … 4 5 6 … 12.
 * При ≤ 7 страницах показывает все номера подряд.
 */
export function pageItems(current: number, last: number): (number | 'gap')[] {
	if (last <= 7) return Array.from({ length: last }, (_, i) => i + 1);
	const items: (number | 'gap')[] = [1];
	const start = Math.max(2, current - 1);
	const end = Math.min(last - 1, current + 1);
	if (start > 2) items.push('gap');
	for (let n = start; n <= end; n++) items.push(n);
	if (end < last - 1) items.push('gap');
	items.push(last);
	return items;
}
