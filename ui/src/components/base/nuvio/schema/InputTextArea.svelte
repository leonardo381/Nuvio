<!--
<script>
  import { createEventDispatcher } from "svelte";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${field?.key || "field"}`;

  function onInput(e) {
    value = e.currentTarget.value;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <textarea
    id={id}
    name={field?.key}
    class="form-textarea"
    rows={field?.options?.rows ?? 4}
    {disabled}
    on:input={onInput}
  >{value ?? ""}</textarea>
</FieldShell>
-->
<script>
  import { createEventDispatcher } from "svelte";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${field?.key || "field"}`;

  function onInput(e) {
    value = e.currentTarget.value;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <textarea
    id={id}
    name={field?.key}
    class="form-textarea"
    rows={field?.options?.rows ?? 4}
    {disabled}
    on:input={onInput}
  >{value ?? ""}</textarea>
</FieldShell>

<!--
<style>
  .pb-textarea {
    width: 100%;
    min-height: 96px;
    padding: 10px 12px;

    font-size: 14px;
    color: #1f2937;
    line-height: 1.4;

    background-color: #f8fafc;
    border: 1px solid #d1d5db;
    border-radius: 6px;

    resize: vertical;
    outline: none;

    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease,
      background-color 0.15s ease;
  }

  .pb-textarea::placeholder {
    color: #9ca3af;
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

  /* error state (optional, if FieldShell sets error class) */
  :global(.form-field.error) .pb-textarea {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15);
  }
</style>
-->