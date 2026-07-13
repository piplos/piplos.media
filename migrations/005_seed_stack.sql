-- Начальный каталог технологического стека (как на главной сайта).
-- Управление: админка → Списки → Стек.

-- +goose Up
INSERT INTO stack_items (slug, label, group_id, published, sort_order) VALUES
    ('react', 'React', 'frontend', true, 0),
    ('svelte', 'Svelte', 'frontend', true, 1),
    ('vue', 'Vue', 'frontend', true, 2),
    ('nextjs', 'Next.js', 'frontend', true, 3),
    ('typescript', 'TypeScript', 'frontend', true, 4),
    ('flutter', 'Flutter', 'mobile', true, 0),
    ('swift', 'Swift', 'mobile', true, 1),
    ('nodejs', 'Node.js', 'backend', true, 0),
    ('bun', 'Bun', 'backend', true, 1),
    ('golang', 'Go', 'backend', true, 2),
    ('python', 'Python', 'backend', true, 3),
    ('java', 'Java', 'backend', true, 4),
    ('rust', 'Rust', 'backend', true, 5),
    ('graphql', 'GraphQL', 'backend', true, 6),
    ('postgresql', 'PostgreSQL', 'data', true, 0),
    ('mysql', 'MySQL', 'data', true, 1),
    ('clickhouse', 'ClickHouse', 'data', true, 2),
    ('redis', 'Redis', 'data', true, 3),
    ('docker', 'Docker', 'devops', true, 0),
    ('kubernetes', 'Kubernetes', 'devops', true, 1),
    ('aws', 'AWS', 'devops', true, 2),
    ('terraform', 'Terraform', 'devops', true, 3),
    ('github-actions', 'GitHub Actions', 'devops', true, 4),
    ('figma', 'Figma', 'design', true, 0)
ON CONFLICT (slug) DO NOTHING;

-- +goose Down
DELETE FROM stack_items WHERE slug IN (
    'react', 'svelte', 'vue', 'nextjs', 'typescript',
    'flutter', 'swift',
    'nodejs', 'bun', 'golang', 'python', 'java', 'rust', 'graphql',
    'postgresql', 'mysql', 'clickhouse', 'redis',
    'docker', 'kubernetes', 'aws', 'terraform', 'github-actions',
    'figma'
);
