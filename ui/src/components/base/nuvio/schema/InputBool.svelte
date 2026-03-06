<script>
  import { createEventDispatcher } from "svelte";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = false;
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${field?.key || "field"}`;

  function onToggle(e) {
    value = !!e.currentTarget.checked;
    dispatch("change", value);
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
<!--
  <label class="switch">
    <input id={id} name={field?.key} type="checkbox" checked={!!value} {disabled} on:change={onToggle} />
    <span class="slider"></span>
  </label>
-->
<label class="pb-switch">
  <input
    id={id}
    name={field?.key}
    type="checkbox"
    checked={!!value}
    {disabled}
    on:change={onToggle}
  />
  <span class="pb-slider"></span>
</label>

</FieldShell>
<!--
<style>
  .switch { position:relative; display:inline-block; width:44px; height:24px; }
  .switch input { display:none; }
  .slider { position:absolute; cursor:pointer; inset:0; background:#bbb; border-radius:999px; transition:.15s; }
  .slider:before { content:""; position:absolute; height:18px; width:18px; left:3px; top:3px; background:white; border-radius:999px; transition:.15s; }
  input:checked + .slider { background:#444; }
  input:checked + .slider:before { transform: translateX(20px); }
</style>
-->
<style>
  /* wrapper */
  .pb-switch {
    position: relative;
    display: inline-flex;
    align-items: center;
    width: 36px;
    height: 20px;
  }

  /* hide native checkbox */
  .pb-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  /* track */
  .pb-slider {
    position: absolute;
    inset: 0;
    background-color: #d1d5db;
    border-radius: 999px;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  /* knob */
  .pb-slider::before {
    content: "";
    position: absolute;
    height: 14px;
    width: 14px;
    left: 3px;
    top: 3px;
    background-color: #ffffff;
    border-radius: 999px;
    transition: transform 0.15s ease;
    box-shadow: 0 1px 2px rgba(0,0,0,0.2);
  }

  /* checked */
  .pb-switch input:checked + .pb-slider {
    background-color: #3b82f6;
  }

  .pb-switch input:checked + .pb-slider::before {
    transform: translateX(16px);
  }

  /* disabled */
  .pb-switch input:disabled + .pb-slider {
    background-color: #e5e7eb;
    cursor: not-allowed;
  }

  .pb-switch input:disabled + .pb-slider::before {
    background-color: #f3f4f6;
  }

  /* error state from FieldShell */
  :global(.form-field.error) .pb-slider {
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.25);
  }
</style>