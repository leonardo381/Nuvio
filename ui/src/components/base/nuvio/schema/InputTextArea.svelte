<script>
  import { createEventDispatcher, onMount } from "svelte";
  import TinyMCE from "@/components/base/TinyMCE.svelte";
  import CommonHelper from "@/utils/CommonHelper";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";
  export let disabled = false;
  export let error = "";
  export let path = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${String(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-")}`;

  let mounted = false;
  let mountTimer = null;
  let editorValue = normalizeValue(value);
  let lastDispatchedValue = editorValue;

  $: editorConfig = {
    ...CommonHelper.defaultEditorOptions(),
    convert_urls: false,
    relative_urls: false,
    min_height: field?.options?.rows ? Math.max(180, Number(field.options.rows) * 36) : 270,
    height: field?.options?.rows ? Math.max(180, Number(field.options.rows) * 36) : 270,
  };

  $: {
    const next = normalizeValue(value);
    if (next !== editorValue) {
      editorValue = next;
      lastDispatchedValue = next;
    }
  }

  $: if (mounted && editorValue !== lastDispatchedValue) {
    lastDispatchedValue = editorValue;
    dispatch("change", editorValue);
  }

  function normalizeValue(raw) {
    if (typeof raw === "string") return raw;
    if (raw === null || typeof raw === "undefined") return "";
    return String(raw);
  }

  onMount(() => {
    // Slight delay avoids heavy initial paint when many fields mount together.
    mountTimer = setTimeout(() => {
      mounted = true;
    }, 60);

    return () => {
      if (mountTimer) clearTimeout(mountTimer);
    };
  });
</script>

<FieldShell {field} {id} {error} required={!!field?.required}>
  {#if mounted}
    <TinyMCE id={id} conf={editorConfig} bind:value={editorValue} {disabled} />
  {:else}
    <div class="textarea-editor-skeleton" aria-hidden="true"></div>
  {/if}
</FieldShell>
<style>
  .textarea-editor-skeleton {
    width: 100%;
    min-height: 220px;
    border: 1px solid rgba(15, 23, 42, 0.12);
    border-radius: 10px;
    background: color-mix(in srgb, var(--baseAlt1Color) 70%, #fff);
  }
</style>
