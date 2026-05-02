<script>
  import { onMount, createEventDispatcher } from "svelte";

  import InputText from "./InputText.svelte";
  import InputTextArea from "./InputTextArea.svelte";
  import InputBool from "./InputBool.svelte";
  //import InputDate from "./InputDate.svelte";
  import InputJson from "./InputJson.svelte";
  import InputSelect from "./InputSelect.svelte";
  import InputFile from "./InputFile.svelte";
  import InputRelation from "./InputRelation.svelte";
  import InputArray from "./InputArray.svelte";
  import InputObject from "./InputObject.svelte";

  export let block = null;
  export let fields = null;
  export let value = null;
  export let showImport = true;
  export let path = "schema";

  const dispatch = createEventDispatcher();

  let schema = { fields: [] };
  let values = {};
  let lastKey = null;
  let importInput;

  async function loadSchema() {
    if (fields) {
      schema = { fields };
      values = value || {};
      return;
    }

    if (!block?.component_key) return;

    const filter = encodeURIComponent(`key="${block.component_key}"`);
    const res = await fetch(`/api/collections/components/records?filter=${filter}&perPage=1`);
    const data = await res.json();
    const cmp = data?.items?.[0];

    let s = cmp?.schema || {};
    if (typeof s === "string") {
      try { s = JSON.parse(s); } catch { s = {}; }
    }
    if (!Array.isArray(s.fields)) s.fields = [];

    schema = s;

    try {
      values = typeof block.props === "string"
        ? JSON.parse(block.props || "{}")
        : (block.props || {});
    } catch {
      values = {};
    }
  }

  function update(key, val) {
    values = { ...values, [key]: val };
    dispatch("change", { value: values });
    dispatch("propsChange", values);
  }

  function openImportPicker() {
    importInput?.click();
  }

  async function handleImportFile(event) {
    const file = event.currentTarget?.files?.[0];
    if (!file) return;

    try {
      const text = await file.text();
      const parsed = JSON.parse(text);

      values = parsed || {};
      dispatch("change", { value: values });
      dispatch("propsChange", values);
    } catch (err) {
      console.error("Failed to import JSON:", err);
      alert("Invalid JSON file.");
    } finally {
      event.currentTarget.value = "";
    }
  }

  onMount(() => {
    lastKey = block?.component_key;
    loadSchema();
  });

  $: if (fields) {
    schema = { fields };
    values = value || {};
  }

  $: if (block?.component_key && block.component_key !== lastKey) {
    lastKey = block.component_key;
    loadSchema();
  }

  $: if (fields && value) {
    values = value;
  }
</script>

{#if showImport}
  <div style="margin-bottom: 12px; display: flex; justify-content: flex-end;">
    <button
      type="button"
      on:click={openImportPicker}
      style="
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding: 8px 14px;
        border: 1px solid #d9e0e7;
        border-radius: 10px;
        background: #ffffff;
        color: #344054;
        font-size: 13px;
        font-weight: 600;
        cursor: pointer;
        box-shadow: 0 1px 2px rgba(16, 24, 40, 0.05);
      "
    >
      Import JSON
    </button>

    <input
      bind:this={importInput}
      type="file"
      accept=".json,application/json"
      style="display: none;"
      on:change={handleImportFile}
    />
  </div>
{/if}

{#each schema.fields as field (field.key)}
  {#if field.type === "text"}
    <InputText path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "textarea"}
    <InputTextArea path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "bool"}
    <InputBool path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "json"}
    <InputJson path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "select"}
    <InputSelect path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "file"}
    <InputFile path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "relation"}
    <InputRelation path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "array"}
    <InputArray path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "object"}
    <InputObject path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else}
    <InputText path={`${path}.${field.key}`} field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />
  {/if}
{/each}

