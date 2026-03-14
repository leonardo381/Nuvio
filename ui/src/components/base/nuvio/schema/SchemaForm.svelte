<script>
  import { onMount, createEventDispatcher } from "svelte";

  import InputText from "./InputText.svelte";
  import InputTextArea from "./InputTextArea.svelte";
  import InputBool from "./InputBool.svelte";
  import InputDate from "./InputDate.svelte";
  import InputJson from "./InputJson.svelte";
  import InputSelect from "./InputSelect.svelte";
  import InputFile from "./InputFile.svelte";
  import InputRelation from "./InputRelation.svelte";
  import InputArray from "./InputArray.svelte";
  import InputObject from "./InputObject.svelte";

  export let block = null;
  export let fields = null;
  export let value = null;

  const dispatch = createEventDispatcher();

  let schema = { fields: [] };
  let values = {};
  let lastKey = null;

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

{#each schema.fields as field (field.key)}
  {#if field.type === "text"}
    <InputText field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "textarea"}
    <InputTextArea field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "bool"}
    <InputBool field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "date"}
    <InputDate field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "json"}
    <InputJson field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "select"}
    <InputSelect field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "file"}
    <InputFile field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "relation"}
    <InputRelation field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "array"}
    <InputArray field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else if field.type === "object"}
    <InputObject field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />

  {:else}
    <InputText field={field} value={values[field.key]} on:change={(e) => update(field.key, e.detail?.value ?? e.detail)} />
  {/if}
{/each}