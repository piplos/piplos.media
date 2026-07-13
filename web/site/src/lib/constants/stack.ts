export const STACK_GROUPS = [
	{
		id: 'frontend',
		items: [
			{ id: 'react', label: 'React' },
			{ id: 'svelte', label: 'Svelte' },
			{ id: 'vue', label: 'Vue' },
			{ id: 'nextjs', label: 'Next.js' },
			{ id: 'typescript', label: 'TypeScript' }
		]
	},
	{
		id: 'mobile',
		items: [{ id: 'flutter', label: 'Flutter' }]
	},
	{
		id: 'backend',
		items: [
			{ id: 'nodejs', label: 'Node.js' },
			{ id: 'bun', label: 'Bun' },
			{ id: 'golang', label: 'Go' },
			{ id: 'python', label: 'Python' },
			{ id: 'rust', label: 'Rust' },
			{ id: 'graphql', label: 'GraphQL' }
		]
	},
	{
		id: 'data',
		items: [
			{ id: 'postgresql', label: 'PostgreSQL' },
			{ id: 'mysql', label: 'MySQL' },
			{ id: 'redis', label: 'Redis' }
		]
	},
	{
		id: 'devops',
		items: [
			{ id: 'docker', label: 'Docker' },
			{ id: 'kubernetes', label: 'Kubernetes' },
			{ id: 'aws', label: 'AWS' },
			{ id: 'terraform', label: 'Terraform' },
			{ id: 'github-actions', label: 'GitHub Actions' }
		]
	},
	{
		id: 'design',
		items: [{ id: 'figma', label: 'Figma' }]
	}
] as const;

export type StackItemId = (typeof STACK_GROUPS)[number]['items'][number]['id'];

export type StackItem = { readonly id: StackItemId; readonly label: string };

export const STACK_ITEMS: readonly StackItem[] = STACK_GROUPS.flatMap(
	(group): readonly StackItem[] => group.items
);
