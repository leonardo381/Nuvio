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

<div class="space-y-3 rounded border p-4">
  <div class="font-medium">{field.label}</div>

  <SchemaForm
    fields={field.fields ?? []}
    value={objectValue}
    on:change={handleChange}
  />
</div>