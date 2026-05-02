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

  $: rawOptions = Array.isArray(field?.options)
    ? field.options
    : Array.isArray(field?.options?.values)
      ? field.options.values
      : [];

  $: options = rawOptions.map((opt) =>
    typeof opt === "string"
      ? { label: opt, value: opt }
      : {
          label: opt?.label ?? opt?.value ?? "",
          value: opt?.value ?? ""
        }
  );

  function onChange(e) {
    value = e.currentTarget.value;
    dispatch("change", { value });
  }
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  <div class="schema-select-wrap">
    <select
      id={id}
      name={field?.key}
      class="form-select schema-select-input"
      {disabled}
      bind:value
      on:change={onChange}
    >
      <option value="">Select...</option>

      {#each options as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>

    <span class="schema-select-arrow" aria-hidden="true"></span>
  </div>
</FieldShell>

<style>
  .schema-select-wrap {
    position: relative;
    width: 100%;
  }

  .schema-select-input {
    width: 100%;
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
    padding-right: 34px;
  }

  .schema-select-arrow {
    position: absolute;
    right: 12px;
    top: 50%;
    width: 0;
    height: 0;
    transform: translateY(-50%);
    border-left: 4px solid transparent;
    border-right: 4px solid transparent;
    border-top: 5px solid color-mix(in srgb, var(--txtHintColor) 90%, var(--txtPrimaryColor));
    pointer-events: none;
  }

  .schema-select-input:disabled + .schema-select-arrow {
    opacity: 0.55;
  }
</style>
