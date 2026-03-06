<script>
  import { createEventDispatcher } from "svelte";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";       // stored as string JSON
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${field?.key || "field"}`;

  let localError = "";

  function validate(v) {
    if (!v || v.trim() === "") return "";
    try { JSON.parse(v); return ""; }
    catch (e) { return "Invalid JSON"; }
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
    class="pb-textarea pb-code"
    rows={field?.options?.rows ?? 6}
    {disabled}
    on:input={onInput}
  >{value ?? ""}</textarea>
</FieldShell>
<!--
<style>
  .form-textarea { width:100%; padding:10px; border:1px solid rgba(0,0,0,.2); border-radius:8px; resize:vertical; }
  .code { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
</style>
-->
<style>
  .pb-textarea {
    width: 100%;
    padding: 10px 12px;

    font-size: 14px;
    color: #1f2937;

    background-color: #f8fafc;
    border: 1px solid #d1d5db;
    border-radius: 6px;

    outline: none;
    resize: vertical;
    min-height: 120px;

    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease,
      background-color 0.15s ease;
  }

  .pb-textarea:focus {
    background-color: #ffffff;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  .pb-textarea:disabled {
    background-color: #f1f5f9;
    color: #9ca3af;
    cursor: not-allowed;
  }

  /* monospace like PB JSON/editor fields */
  .pb-code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
    line-height: 1.35;
    tab-size: 2;
  }

  /* error state propagated by FieldShell */
  :global(.form-field.error) .pb-textarea {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15);
  }
</style>
