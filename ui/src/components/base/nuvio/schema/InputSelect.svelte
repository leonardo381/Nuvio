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

  $: options = Array.isArray(field?.options?.values) ? field.options.values : [];

  function onChange(e) {
    value = e.currentTarget.value;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <select
    id={id}
    name={field?.key}
    class="form-select"
    {disabled}
    bind:value
    on:change={onChange}
  >
    <option value="">Select…</option>
    {#each options as opt}
      <option value={opt}>{opt}</option>
    {/each}
  </select>
</FieldShell>

<style>
  .form-select { width:100%; padding:10px; border:1px solid rgba(0,0,0,.2); border-radius:8px; background:white; }
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

  $: options = Array.isArray(field?.options?.values)
    ? field.options.values
    : [];

  function onChange(e) {
    value = e.currentTarget.value;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <div class="pb-select-wrapper">
    <select
      id={id}
      name={field?.key}
      class="form-select"
      {disabled}
      bind:value
      on:change={onChange}
    >
      <option value="">Select…</option>
      {#each options as opt}
        <option value={opt}>{opt}</option>
      {/each}
    </select>

    <!-- arrow -->
    <span class="pb-select-arrow"></span>
  </div>
</FieldShell>

<!--
<style>
  .pb-select-wrapper {
    position: relative;
    width: 100%;
  }

  .pb-select {
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;

    width: 100%;
    padding: 10px 36px 10px 12px;

    font-size: 14px;
    color: #1f2937;

    background-color: #f8fafc;
    border: 1px solid #d1d5db;
    border-radius: 6px;

    outline: none;
    cursor: pointer;

    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease,
      background-color 0.15s ease;
  }

  .pb-select:focus {
    background-color: #ffffff;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  .pb-select:disabled {
    background-color: #f1f5f9;
    color: #9ca3af;
    cursor: not-allowed;
  }

  /* dropdown arrow */
  .pb-select-arrow {
    position: absolute;
    top: 50%;
    right: 12px;
    width: 0;
    height: 0;
    pointer-events: none;

    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 6px solid #6b7280;

    transform: translateY(-50%);
  }

  /* error state (if FieldShell adds .error) */
  :global(.form-field.error) .pb-select {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15);
  }
</style>
-->