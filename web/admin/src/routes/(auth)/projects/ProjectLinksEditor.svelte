<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import {
		emptyProjectLink,
		isProjectLinkKind,
		PROJECT_LINK_KIND_OPTIONS,
		withUpdatedUrl
	} from '$lib/project-links';
	import type { ProjectLink } from '$lib/types';

	interface Props {
		links?: ProjectLink[];
		idPrefix?: string;
	}

	let { links = $bindable([]), idPrefix = 'project-link' }: Props = $props();

	function patch(index: number, next: ProjectLink) {
		const copy = [...links];
		if (!copy[index]) return;
		copy[index] = next;
		links = copy;
	}

	function addLink() {
		links = [...links, emptyProjectLink()];
	}

	function removeLink(index: number) {
		links = links.filter((_, i) => i !== index);
	}

	function setUrl(index: number, url: string) {
		const current = links[index];
		if (!current) return;
		patch(index, withUpdatedUrl(current, url));
	}

	function setLabel(index: number, label: string) {
		const current = links[index];
		if (!current) return;
		patch(index, { ...current, label });
	}

	function setKind(index: number, kind: string) {
		const current = links[index];
		if (!current || !isProjectLinkKind(kind)) return;
		patch(index, { ...current, kind });
	}
</script>

<div class="links-editor">
	<div class="links-head">
		<h2 class="section-title">Ссылки проекта</h2>
		<Button type="button" variant="secondary" onclick={addLink}>Добавить ссылку</Button>
	</div>

	{#if !links.length}
		<p class="links-empty">Нет ссылок. Добавьте сайт проекта, App Store или Google Play.</p>
	{:else}
		<ul class="links-list">
			{#each links as link, index (index)}
				<li class="link-row">
					<FormField label="URL" id="{idPrefix}-{index}-url">
						<Input
							id="{idPrefix}-{index}-url"
							type="url"
							placeholder="https://example.com"
							bind:value={() => link.url, (v) => setUrl(index, v)}
						/>
					</FormField>
					<FormField label="Подпись" id="{idPrefix}-{index}-label">
						<Input
							id="{idPrefix}-{index}-label"
							placeholder="Сайт / Google Play"
							bind:value={() => link.label, (v) => setLabel(index, v)}
						/>
					</FormField>
					<FormField label="Тип" id="{idPrefix}-{index}-kind">
						<Select
							id="{idPrefix}-{index}-kind"
							options={PROJECT_LINK_KIND_OPTIONS}
							bind:value={() => link.kind, (v) => setKind(index, v)}
						/>
					</FormField>
					<div class="link-actions">
						<Button type="button" variant="ghost" onclick={() => removeLink(index)}>Удалить</Button>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.links-editor {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.links-head {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}
	.section-title {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
		color: #18181b;
	}
	.links-empty {
		margin: 0;
		font-size: 0.875rem;
		color: #71717a;
	}
	.links-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.link-row {
		display: grid;
		grid-template-columns: minmax(0, 2fr) minmax(0, 1.25fr) minmax(7.5rem, 0.85fr) auto;
		gap: 0.75rem;
		align-items: end;
		padding: 0.75rem;
		background: #fafafa;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
	}
	.link-actions {
		display: flex;
		align-items: center;
		padding-bottom: 0.125rem;
	}
	@media (max-width: 720px) {
		.link-row {
			grid-template-columns: 1fr;
			align-items: stretch;
		}
		.link-actions {
			justify-content: flex-end;
			padding-bottom: 0;
		}
	}
</style>
