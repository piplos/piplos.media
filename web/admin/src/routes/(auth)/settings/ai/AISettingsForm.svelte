<script lang="ts">
  import type { Snippet } from "svelte";
  import { enhance } from "$app/forms";
  import type { AiSettingsData } from "./_types";
  import { groupModelsByProvider, filterEnabledProviders, noReload } from "./_ai-helpers";
  import Card from "$lib/components/Card.svelte";
  import FormField from "$lib/components/FormField.svelte";
  import Button from "$lib/components/Button.svelte";
  import Select from "$lib/components/Select.svelte";
  import Textarea from "$lib/components/Textarea.svelte";
  import { createFormToastHandler } from "$lib/utils/form-toasts";

  interface Props {
    aiData: AiSettingsData;
    form: Record<string, unknown> | null;
    /** camelCase-префикс контракта страницы: имена полей формы, actions и ключей form-result. */
    prefix: string;
    title: string;
    /** Текущие значения из load (перезаписывают локальный state при каждой загрузке). */
    provider: string;
    model: string;
    prompt: string;
    saveSuccessMessage: string;
    /** Рендерить кнопку «Тестировать» и блок результата (action `?/test{Prefix}`). */
    withTest?: boolean;
    testSuccessMessage?: string;
    /** Формат ответа модели: JSON с отступами или сырой текст. */
    testResultFormat?: "json" | "text";
    promptLabel?: string;
    promptPlaceholder?: string;
    promptRows?: number;
    promptHint?: Snippet;
    /** Дополнительные поля внутри формы (например, промпты text-translation). */
    extraFields?: Snippet;
  }

  let {
    aiData,
    form,
    prefix,
    title,
    provider,
    model,
    prompt,
    saveSuccessMessage,
    withTest = true,
    testSuccessMessage,
    testResultFormat = "json",
    promptLabel = "Системный промпт",
    promptPlaceholder = "Пусто — используется встроенный промпт",
    promptRows = 10,
    promptHint,
    extraFields,
  }: Props = $props();

  const cap = $derived(prefix.charAt(0).toUpperCase() + prefix.slice(1));
  const kebab = $derived(prefix.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`));
  const formId = $derived(`${kebab}-form`);

  const MODELS_BY_PROVIDER = $derived(groupModelsByProvider(aiData.aiModels ?? []));
  const enabledProviders = $derived(filterEnabledProviders(aiData));

  // Writable deriveds: редактируются через bind, пересинхронизируются при смене
  // данных load; невалидные значения нормализуются к первому доступному варианту.
  let providerValue = $derived.by(() => {
    if (
      enabledProviders.length > 0 &&
      !enabledProviders.some((p) => p.value === provider)
    ) {
      return enabledProviders[0].value;
    }
    return provider;
  });
  let modelValue = $derived.by(() => {
    const list = MODELS_BY_PROVIDER[providerValue] ?? [];
    if (list.length > 0 && !list.some((m) => m.value === model)) {
      return list[0].value;
    }
    return model;
  });
  let promptValue = $derived(prompt);
  let showUserPrompt = $state(false);
  let testInProgress = $state(false);

  const models = $derived(MODELS_BY_PROVIDER[providerValue] ?? []);

  function testEnhance() {
    testInProgress = true;
    return async ({
      update,
    }: {
      update: (opts?: { invalidateAll?: boolean; reset?: boolean }) => Promise<void>;
    }) => {
      try {
        await update?.({ invalidateAll: false, reset: false });
      } finally {
        testInProgress = false;
      }
    };
  }

  const testError = $derived(form?.[`test${cap}Error`] as string | undefined);
  const testResponse = $derived(form?.[`test${cap}Response`]);
  const testUserPrompt = $derived(form?.[`test${cap}UserPrompt`] as string | undefined);

  // Пересоздаётся только при смене конфигурации (в реальности props статичны).
  const handleFormToasts = $derived(
    createFormToastHandler({
      [prefix]: { success: saveSuccessMessage },
      ...(withTest && testSuccessMessage !== undefined
        ? { [`test${cap}`]: { success: testSuccessMessage } }
        : {}),
    }),
  );
  $effect(() => {
    handleFormToasts(form);
  });
</script>

{#if aiData.error}
  <div class="admin-table-wrap admin-table-wrap--empty">
    <p class="text-muted">Данные недоступны</p>
  </div>
{:else}
<div class="ai-form">
  <Card class="ai-settings-card">
    <div class="ai-card-content">
      <form id={formId} method="POST" action="?/update{cap}" use:enhance={noReload}>
        <h2 class="ai-section-title">{title}</h2>
        <div class="ai-fields-row">
          <FormField label="Провайдер" id="{kebab}-provider">
            <Select
              id="{kebab}-provider"
              name="{prefix}Provider"
              bind:value={providerValue}
              options={enabledProviders}
              class="ai-select"
            />
          </FormField>
          {#if models.length > 1}
            <FormField label="Модель" id="{kebab}-model">
              <Select
                id="{kebab}-model"
                name="{prefix}Model"
                bind:value={modelValue}
                options={models}
                class="ai-select"
              />
            </FormField>
          {:else}
            <input type="hidden" name="{prefix}Model" value={modelValue} />
          {/if}
        </div>
        <div class="ai-prompt-field">
          <FormField label={promptLabel} id="{kebab}-prompt">
            <Textarea
              id="{kebab}-prompt"
              name="{prefix}Prompt"
              bind:value={promptValue}
              rows={promptRows}
              placeholder={promptPlaceholder}
            />
            {#if promptHint}
              {@render promptHint()}
            {/if}
          </FormField>
        </div>
        {#if extraFields}
          {@render extraFields()}
        {/if}
      </form>
      <div class="ai-actions">
        <Button
          type="button"
          onclick={() =>
            (document.getElementById(formId) as HTMLFormElement | null)?.requestSubmit()}
        >
          Сохранить
        </Button>
        {#if withTest}
          <form method="POST" action="?/test{cap}" use:enhance={testEnhance} class="ai-test-form">
            <Button type="submit" variant="secondary" loading={testInProgress}>
              Тестировать
            </Button>
          </form>
        {/if}
      </div>
      {#if withTest}
        {#if testError}
          <p class="ai-test-error">{testError}</p>
        {:else if testResponse != null}
          <div class="today-test-result">
            {#if testUserPrompt != null && testUserPrompt !== ""}
              <button
                type="button"
                class="ai-test-prompt-toggle"
                onclick={() => (showUserPrompt = !showUserPrompt)}
                aria-expanded={showUserPrompt}
              >
                <span class="ai-test-prompt-label">User prompt</span>
                <svg
                  class="ai-test-prompt-chevron"
                  class:ai-test-prompt-chevron--open={showUserPrompt}
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M6 9l6 6 6-6" />
                </svg>
              </button>
              {#if showUserPrompt}
                <pre class="ai-test-pre ai-test-prompt">{testUserPrompt}</pre>
              {/if}
            {/if}
            <p class="ai-test-prompt-label">Ответ модели</p>
            <pre class="ai-test-pre">{testResultFormat === "json"
              ? JSON.stringify(testResponse, null, 2)
              : String(testResponse)}</pre>
          </div>
        {/if}
      {/if}
    </div>
  </Card>
</div>
{/if}

<style>
  .text-muted {
    color: #71717a;
  }
  .ai-card-content {
    margin: -2rem -2rem -2.25rem -2rem;
    padding: 2rem 2rem 2.25rem 2rem;
    min-height: 200px;
    box-sizing: border-box;
  }
</style>
