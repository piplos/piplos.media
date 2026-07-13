<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props {
    id: string;
    name?: string;
    type?: string;
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean;
    autocomplete?: HTMLInputAttributes["autocomplete"];
    required?: boolean;
    class?: string;
    step?: string;
    min?: string;
    max?: string;
  }
  let {
    id,
    name,
    type = "text",
    value = $bindable(""),
    placeholder,
    disabled = false,
    error = false,
    autocomplete,
    required = false,
    class: className = "",
    step,
    min,
    max,
  }: Props = $props();
</script>

<input
  {id}
  {name}
  {type}
  bind:value
  {placeholder}
  {disabled}
  {autocomplete}
  {required}
  {step}
  {min}
  {max}
  class="input {error ? 'input-error' : ''} {className}"
  aria-invalid={error}
/>

<style>
  .input {
    width: 100%;
    min-height: 2.25rem;
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.5;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    background: #fff;
    color: #111;
    box-sizing: border-box;
    transition:
      border-color 0.15s,
      box-shadow 0.15s;
  }
  .input::placeholder {
    color: #9ca3af;
  }
  .input:hover:not(:disabled) {
    border-color: #9ca3af;
  }
  .input:focus {
    outline: none;
    border-color: #111;
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
  }
  .input:disabled {
    background: #f4f4f5;
    cursor: not-allowed;
    opacity: 1;
  }
  .input-error {
    border-color: #dc2626;
  }
  .input-error:focus {
    border-color: #dc2626;
    box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.2);
  }
</style>
