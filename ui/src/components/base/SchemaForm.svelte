<!--
<script>
  import { onMount } from "svelte";

  // The block record we’re editing (passed as {record} from RecordUpsertPanel)
  export let block;

  let loading = true;
  let error = "";
  let schema = { fields: [] };
  let formValues = {};
  

  async function loadSchema() {
    loading = true;
    error = "";
    schema = { fields: [] };

    if (!block || !block.component_key) {
      error = "This block has no component_key.";
      loading = false;
      return;
    }

    try {
      // Build filter: key="hero_simple"
      const filter = encodeURIComponent(`key="${block.component_key}"`);

      // Call PocketBase REST API for the components collection
      const res = await fetch(
        `/api/collections/components/records?filter=${filter}&perPage=1`
      );

      if (!res.ok) {
        error = `Failed to load schema (status ${res.status})`;
        console.error("Schema load error:", await res.text());
        loading = false;
        return;
      }

      const data = await res.json();
      const cmp = data?.items?.[0];

      if (!cmp || !cmp.schema || !cmp.schema.fields || !cmp.schema.fields.length) {
        error = "No schema defined for this component.";
        loading = false;
        return;
      }

      schema = cmp.schema;
      // Pre-fill with existing props if any
      formValues = { ...(block.props || {}) };
    } catch (e) {
      console.error("Schema load exception:", e);
      error = "Error loading schema.";
    } finally {
      loading = false;
    }
  }

  onMount(loadSchema);

  function updateField(key, value) {
    formValues = { ...formValues, [key]: value };
    // This is what PocketBase will save when you click "Save changes"
    block.props = formValues;
  }
</script>
-->
<script>
  import { onMount } from "svelte";

  // The block record we’re editing (passed as {record} from RecordUpsertPanel)
  export let block;

  let loading = false;
  let error = "";
  let schema = { fields: [] };
  let formValues = {};
  let lastKey = null; // remember last loaded component_key

  async function loadSchema() {
    const key = block?.component_key?.trim();

    // no key → clear schema + props and show message
    if (!key) {
      schema = { fields: [] };
      formValues = {};
      block.props = formValues; // normalise to {} instead of null
      error = "This block has no component_key.";
      loading = false;
      return;
    }

    loading = true;
    error = "";
    schema = { fields: [] };

    try {
      // Build filter: key="hero_simple"
      const filter = encodeURIComponent(`key="${key}"`);

      // Call PocketBase REST API for the components collection
      const res = await fetch(
        `/api/collections/components/records?filter=${filter}&perPage=1`
      );

      if (!res.ok) {
        error = `Failed to load schema (status ${res.status})`;
        console.error("Schema load error:", await res.text());
        loading = false;
        return;
      }

      const data = await res.json();
      const cmp = data?.items?.[0];

      if (!cmp || !cmp.schema || !cmp.schema.fields || !cmp.schema.fields.length) {
        error = "No schema defined for this component.";
        loading = false;
        return;
      }

      schema = cmp.schema;

      // Pre-fill with existing props if any, otherwise {}
      formValues = { ...(block.props || {}) };
      block.props = formValues;   // ensure props is never null once schema is loaded
    } catch (e) {
      console.error("Schema load exception:", e);
      error = "Error loading schema.";
    } finally {
      loading = false;
    }
  }

  // initial load on mount (for existing blocks that already have component_key)
  onMount(() => {
    if (block?.component_key) {
      lastKey = block.component_key;
      loadSchema();
    }
  });

  // 🔁 re-run whenever component_key changes (new block or edited key)
  $: if (block?.component_key && block.component_key !== lastKey) {
    lastKey = block.component_key;
    loadSchema();
  }

  function updateField(key, value) {
    formValues = { ...formValues, [key]: value };
    // This is what PocketBase will save when you click "Save changes"
    block.props = formValues;
    // optional debug:
    // console.log("Updated props:", block.props);
  }
</script>

{#if loading}
  <p>Loading schema…</p>

{:else if error}
  <p>{error}</p>

{:else}
  {#each schema.fields as f (f.key)}
    <div class="field">
      <label>{f.label}</label>

      {#if f.type === "text"}
        <input
          type="text"
          value={formValues[f.key] ?? ""}
          on:input={(e) => updateField(f.key, e.currentTarget.value)}
        />

      {:else if f.type === "textarea"}
        <textarea
          value={formValues[f.key] ?? ""}
          on:input={(e) => updateField(f.key, e.currentTarget.value)}
        ></textarea>

      {:else if f.type === "url"}
        <input
          type="url"
          value={formValues[f.key] ?? ""}
          on:input={(e) => updateField(f.key, e.currentTarget.value)}
        />

      {:else if f.type === "file"}
        <!-- simplest version: just store file id or path as text for now -->
        <input
          type="text"
          placeholder="file id or filename"
          value={formValues[f.key] ?? ""}
          on:input={(e) => updateField(f.key, e.currentTarget.value)}
        />

      {:else}
        <!-- fallback to text -->
        <input
          type="text"
          value={formValues[f.key] ?? ""}
          on:input={(e) => updateField(f.key, e.currentTarget.value)}
        />
      {/if}
    </div>
  {/each}
{/if}
