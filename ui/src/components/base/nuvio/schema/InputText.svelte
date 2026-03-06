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
  <input
    id={id}
    name={field?.key}
    class="form-input"
    type="text"
    {disabled}
    value={value ?? ""}
    minlength={Number.isInteger(field?.options?.min) ? field.options.min : null}
    maxlength={Number.isInteger(field?.options?.max) ? field.options.max : null}
    pattern={field?.options?.pattern ? field.options.pattern : null}
    on:input={onInput}
  />
</FieldShell>

<style>
  .form-input { width:100%; padding:10px; border:1px solid rgba(0,0,0,.2); border-radius:8px; }
</style>
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
  <input
    id={id}
    name={field?.key}
    type="text"
    class="form-input"
    {disabled}
    value={value ?? ""}
    minlength={Number.isInteger(field?.options?.min) ? field.options.min : undefined}
    maxlength={Number.isInteger(field?.options?.max) ? field.options.max : undefined}
    pattern={field?.options?.pattern || undefined}
    on:input={onInput}
  />
</FieldShell>


<!--
<style>
  .pb-input {
    width: 100%;
    height: 38px;
    padding: 0 12px;

    font-size: 14px;
    color: #1f2937;

    background-color: #f8fafc;
    border: 1px solid #d1d5db;
    border-radius: 6px;

    outline: none;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease,
      background-color 0.15s ease;
  }

  .pb-input::placeholder {
    color: #9ca3af;
  }

  .pb-input:focus {
    background-color: #ffffff;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  .pb-input:disabled {
    background-color: #f1f5f9;
    color: #9ca3af;
    cursor: not-allowed;
  }

  /* error state (optional, if FieldShell sets error class) */
  :global(.form-field.error) .pb-input {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15);
  }
</style>
-->