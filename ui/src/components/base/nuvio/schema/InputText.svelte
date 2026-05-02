<script>
  import { createEventDispatcher } from "svelte";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";
  export let disabled = false;
  export let error = "";
  export let path = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${String(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-")}`;

  function onInput(e) {
    value = e.currentTarget.value;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <input
    id={id}
    name={field?.key}
    type="text"
    class="form-input schema-input-text"
    {disabled}
    value={value ?? ""}
    minlength={Number.isInteger(field?.options?.min) ? field.options.min : undefined}
    maxlength={Number.isInteger(field?.options?.max) ? field.options.max : undefined}
    pattern={field?.options?.pattern || undefined}
    on:input={onInput}
  />
</FieldShell>

<style>
  .schema-input-text {
    width: 100%;
  }
</style>
