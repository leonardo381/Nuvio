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

  let localError = "";

  function validate(v) {
    if (!v || v.trim() === "") return "";
    try {
      JSON.parse(v);
      return "";
    } catch (e) {
      return "Invalid JSON";
    }
  }

  function onInput(e) {
    value = e.currentTarget.value;
    localError = validate(value);
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} error={error || localError} required={!!field?.required} hint="Must be valid JSON">
  <textarea
    id={id}
    name={field?.key}
    class="form-textarea schema-json-textarea code"
    rows={field?.options?.rows ?? 6}
    {disabled}
    on:input={onInput}
  >{value ?? ""}</textarea>
</FieldShell>

<style>
  .schema-json-textarea {
    width: 100%;
    min-height: 120px;
    line-height: 1.35;
    tab-size: 2;
  }
</style>
