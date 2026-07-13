<script lang="ts">
  import { Button } from "bits-ui";

  type Variant = "primary" | "secondary" | "ghost" | "danger" | "success";
  interface Props {
    type?: "button" | "submit" | "reset";
    variant?: Variant;
    disabled?: boolean;
    loading?: boolean;
    fullWidth?: boolean;
    class?: string;
    form?: string;
    title?: string;
    onclick?: (e: MouseEvent) => void;
    children?: import("svelte").Snippet;
  }
  let {
    type = "button",
    variant = "primary",
    disabled = false,
    loading = false,
    fullWidth = false,
    class: className = "",
    form: formId,
    title,
    onclick,
    children,
  }: Props = $props();

  const isDisabled = $derived(disabled || loading);
</script>

<Button.Root
  {type}
  disabled={isDisabled}
  {title}
  {onclick}
  form={formId}
  class="btn btn-{variant} {fullWidth ? 'btn-full' : ''} {className}"
>
  {#if loading}
    <span class="btn-spinner" aria-hidden="true"></span>
  {:else if children}
    {@render children()}
  {/if}
</Button.Root>

<style>
  :global(.btn) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.375rem;
    min-height: 2.25rem;
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    font-weight: 400;
    border-radius: 10px;
    cursor: pointer;
    transition:
      background 0.15s,
      border-color 0.15s,
      color 0.15s;
    border: none;
    box-sizing: border-box;
  }
  :global(.btn:disabled) {
    opacity: 0.8;
    cursor: not-allowed;
  }
  :global(.btn:focus),
  :global(.btn:focus-visible) {
    outline: none;
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.2);
  }
  :global(.btn-full) {
    width: 100%;
  }
  :global(.btn-primary) {
    color: #fff;
    background: #111;
  }
  :global(.btn-primary:hover:not(:disabled)) {
    background: #333;
  }
  :global(.btn-secondary) {
    color: #374151;
    background: #fff;
    border: 1px solid #d1d5db;
  }
  :global(.btn-secondary:hover:not(:disabled)) {
    background: #f9fafb;
    border-color: #9ca3af;
  }
  :global(.btn-ghost) {
    color: #374151;
    background: transparent;
  }
  :global(.btn-ghost:hover:not(:disabled)) {
    background: #f3f4f6;
  }
  :global(.btn-danger) {
    color: #fff;
    background: #dc2626;
  }
  :global(.btn-danger:hover:not(:disabled)) {
    background: #b91c1c;
  }
  :global(.btn-success) {
    color: #fff;
    background: #16a34a;
  }
  :global(.btn-success:hover:not(:disabled)) {
    background: #15803d;
  }
  :global(.btn-success:focus),
  :global(.btn-success:focus-visible) {
    box-shadow: 0 0 0 3px rgba(22, 163, 74, 0.35);
  }
  .btn-spinner {
    display: inline-block;
    width: 1em;
    height: 1em;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: btn-spin 0.6s linear infinite;
  }
  @keyframes btn-spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
