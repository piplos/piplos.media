import { getContext, setContext } from 'svelte';
import type { Snippet } from 'svelte';

const TAB_LAYOUT_KEY = Symbol('tab-layout');

export type TabLayoutContext = {
	setActions: (actions: Snippet | null) => void;
};

export function setTabLayoutContext(ctx: TabLayoutContext) {
	setContext(TAB_LAYOUT_KEY, ctx);
}

export function useTabLayout(): TabLayoutContext {
	return getContext(TAB_LAYOUT_KEY);
}
