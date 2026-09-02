import { describe, expect, it } from 'vitest';
import { pageItems } from './pagination';

describe('pageItems', () => {
	it('shows every page when there are seven or fewer', () => {
		expect(pageItems(1, 5)).toEqual([1, 2, 3, 4, 5]);
		expect(pageItems(7, 7)).toEqual([1, 2, 3, 4, 5, 6, 7]);
	});

	it('collapses distant pages into gaps', () => {
		expect(pageItems(6, 12)).toEqual([1, 'gap', 5, 6, 7, 'gap', 12]);
	});

	it('keeps neighbours of the first page', () => {
		expect(pageItems(1, 12)).toEqual([1, 2, 'gap', 12]);
		expect(pageItems(2, 12)).toEqual([1, 2, 3, 'gap', 12]);
	});

	it('keeps neighbours of the last page', () => {
		expect(pageItems(11, 12)).toEqual([1, 'gap', 10, 11, 12]);
		expect(pageItems(12, 12)).toEqual([1, 'gap', 11, 12]);
	});
});
