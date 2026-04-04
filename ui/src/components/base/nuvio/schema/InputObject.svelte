<script>
  import { createEventDispatcher } from "svelte";
  import SchemaForm from "./SchemaForm.svelte";

  export let field;
  export let value = {};

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
      on:change={handleChange}
    />
  </div>
</div>

<style>
  .object-field {
    border: 1px solid #d9e0e7;
    border-radius: 10px;
    background: #ffffff;
    overflow: hidden;
    margin: 0;
    margin-top: 16px;
    margin-bottom: 16px;
  }

  .object-field__header {
    padding: 12px 16px 10px;
    border-bottom: 1px solid #e5ebf2;
    background: #fbfcfd;
  }

  .object-field__title {
    font-size: 14px;
    font-weight: 600;
    color: #475467;
    line-height: 1.2;
  }

  .object-field__body {
    padding: 14px;
    background: #ffffff;
  }

</style>