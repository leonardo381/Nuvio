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

  function handleChange(e) {
    objectValue = e.detail?.value ?? e.detail ?? {};
    dispatch("change", { value: objectValue });
  }

  $: if (value && typeof value === "object" && !Array.isArray(value)) {
    objectValue = { ...value };
  }
</script>

<div class="object-field">
  <div class="object-field__header">
    <div class="object-field__title">{field.label}</div>
  </div>

  <div class="object-field__body">
    <SchemaForm
      fields={field.fields ?? []}
      value={objectValue}
      showImport={false}
      path={path}
      on:change={handleChange}
    />
  </div>
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
    padding: 10px 12px;
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 62%, var(--baseColor));
  }

  .object-field__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: 1.2;
  }

  .object-field__body {
    padding: 10px;
    background: transparent;
  }

  .object-field__body :global(.pb-field:last-child) {
    margin-bottom: 0;
  }
</style>
