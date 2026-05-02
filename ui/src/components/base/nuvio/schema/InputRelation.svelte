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

<FieldShell
  {field}
  {id}
  {error}
  required={!!field?.required}
  hint="Enter record ID(s), separated by commas"
>
  <input
    id={id}
    name={field?.key}
    class="form-input schema-relation-input"
    type="text"
    {disabled}
    value={value ?? ""}
    placeholder="e.g. rec123, rec456"
    on:input={onInput}
  />
</FieldShell>

<style>
  .schema-relation-input {
    width: 100%;
  }
</style>
