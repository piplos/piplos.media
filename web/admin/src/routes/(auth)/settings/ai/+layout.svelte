<script lang="ts">
  let { children } = $props();
</script>

<div class="ai-settings-main">
  {@render children()}
</div>

<style>
  .ai-settings-main {
    width: 100%;
    min-width: 0;
  }

  /* Shared form styles for text-generation settings pages */
  :global(.ai-settings-card) {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-width: 0;
  }
  :global(.ai-section-title) {
    margin: 0 0 1rem;
    font-size: 1.25rem;
    font-weight: 600;
    color: #18181b;
  }
  :global(.ai-fields-row) {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: flex-end;
  }
  :global(.ai-fields-row .field) {
    flex: 1;
    min-width: 10rem;
  }
  :global(.ai-prompt-field) {
    margin-top: 1.25rem;
  }
  :global(.ai-select) {
    width: 100%;
    min-height: 2.75rem;
    padding: 0.5rem 0.75rem;
    font-size: 1rem;
    line-height: 1.5;
    border: 1px solid #d1d5db;
    border-radius: 10px;
    background: #fff;
    color: #111;
    box-sizing: border-box;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.75rem center;
    padding-right: 2.5rem;
    cursor: pointer;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  :global(.ai-select:hover) {
    border-color: #9ca3af;
  }
  :global(.ai-select:focus) {
    outline: none;
    border-color: #111;
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.08);
  }
  :global(.ai-placeholders-hint) {
    margin: 0.5rem 0 0;
    font-size: 0.875rem;
    color: #71717a;
    line-height: 1.4;
  }
  :global(.ai-placeholders-hint code) {
    font-family: ui-monospace, monospace;
    background: #f4f4f5;
    padding: 0.125rem 0.375rem;
    border-radius: 4px;
    font-size: 0.8125rem;
  }
  :global(.ai-actions) {
    margin-top: 1.25rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
  }
  :global(.ai-test-form) {
    margin: 0;
  }
  :global(.ai-test-error) {
    margin: 1.5rem 0 0;
    color: #b91c1c;
    font-size: 0.875rem;
  }
  :global(.today-test-result) {
    margin-top: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  :global(.ai-test-prompt-toggle) {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    margin: 0;
    padding: 0;
    border: none;
    background: none;
    cursor: pointer;
    font: inherit;
    color: inherit;
    text-align: left;
  }
  :global(.ai-test-prompt-chevron) {
    flex-shrink: 0;
    transition: transform 0.2s;
    color: #71717a;
  }
  :global(.ai-test-prompt-chevron--open) {
    transform: rotate(180deg);
  }
  :global(.ai-test-prompt-label) {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
  }
  :global(.ai-test-pre) {
    padding: 0.75rem;
    background: #1e293b;
    color: #e2e8f0;
    border-radius: 6px;
    font-size: 0.8125rem;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-word;
  }
  :global(.ai-test-pre.ai-test-prompt) {
    margin-top: 0.25rem;
  }
</style>
