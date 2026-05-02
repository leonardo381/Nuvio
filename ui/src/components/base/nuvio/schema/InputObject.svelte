<script>
  import { createEventDispatcher } from "svelte";
  import SchemaForm from "./SchemaForm.svelte";

  export let field;
  export let value = {};
  export let path = "";

  const dispatch = createEventDispatcher();

  let objectValue = value && typeof value === "object" && !Array.isArray(value)
    ? { ...value }
    : {};
  let isExpanded = true;

  $: objectFields = Array.isArray(field?.fields) ? field.fields : [];
  $: fieldCountLabel = `${objectFields.length} ${objectFields.length === 1 ? "field" : "fields"}`;
  $: objectTitle = `${field?.label || "Object"}`.trim() || "Object";

  function toSafePathSegment(rawValue) {
    return String(rawValue || field?.key || "object").replace(/[^a-zA-Z0-9_-]/g, "-");
  }

  function getBodyId() {
    return `object-field-${toSafePathSegment(path)}`;
  }

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }

  function handleChange(e) {
    objectValue = e.detail?.value ?? e.detail ?? {};
    dispatch("change", { value: objectValue });
  }

  $: if (value && typeof value === "object" && !Array.isArray(value)) {
    objectValue = { ...value };
  }
</script>

<div class="object-field">
  <button
    type="button"
    class="object-field__header"
    aria-expanded={isExpanded}
    aria-controls={getBodyId()}
    on:click={toggleExpanded}
  >
    <div class="object-field__header-main">
      <div class="object-field__title">{objectTitle}</div>
      <span class="label label-sm object-field__count">{fieldCountLabel}</span>
    </div>
    <i class="ri-arrow-down-s-line object-field__chevron" class:is-open={isExpanded} />
  </button>

  {#if isExpanded}
    <div class="object-field__body" id={getBodyId()}>
      {#if objectFields.length}
        <SchemaForm
          fields={objectFields}
          value={objectValue}
          showImport={false}
          path={path}
          on:change={handleChange}
        />
      {:else}
        <div class="object-field__empty">This group has no fields.</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .object-field {
    border: 1px solid var(--baseAlt2Color);
    border-radius: var(--baseRadius);
    background: var(--baseColor);
    overflow: hidden;
    margin: 12px 0;
  }

  .object-field__header {
    width: 100%;
    border: 0;
    margin: 0;
    color: inherit;
    cursor: pointer;
    text-align: left;
    padding: 10px 12px;
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 62%, var(--baseColor));
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .object-field__header-main {
    min-width: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .object-field__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: 1.2;
  }

  .object-field__count {
    min-height: 20px;
    border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
    color: var(--txtHintColor);
    background: var(--baseColor);
  }

  .object-field__chevron {
    font-size: 16px;
    color: var(--txtHintColor);
    transition: transform var(--baseAnimationSpeed), color var(--baseAnimationSpeed);
    flex: 0 0 auto;
  }

  .object-field__chevron.is-open {
    transform: rotate(180deg);
    color: var(--txtPrimaryColor);
  }

  .object-field__body {
    padding: 10px 10px 8px;
    background: transparent;
  }

  .object-field__empty {
    border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
    border-radius: var(--baseRadius);
    background: color-mix(in srgb, var(--baseAlt1Color) 34%, var(--baseColor));
    color: var(--txtHintColor);
    padding: 10px;
    font-size: var(--smFontSize);
    line-height: var(--smLineHeight);
  }

  .object-field__body :global(.pb-field:last-child) {
    margin-bottom: 0;
  }
</style>
