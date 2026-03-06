<!--
<script>
  import { createEventDispatcher } from "svelte";
  import ApiClient from "@/utils/ApiClient";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = null;     // { collection, recordId, filename } OR string OR null
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();

  $: id = `schema-${field?.key || "field"}`;

  let isUploading = false;
  let localError = "";

  async function handleFileChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    localError = "";

    try {
      const formData = new FormData();
      formData.append("file", file);

      // ⚠️ IMPORTANT:
      // Use your real collection name (most PB setups are lowercase "assets")
      const rec = await ApiClient.collection("Assets").create(formData);

      // ⚠️ IMPORTANT:
      // Use your real file field name from the assets collection (usually "file")
      const filename = rec.file;

      const newVal = {
        collection: "Assets",
        recordId: rec.id,
        filename,
      };

      // ✅ new pattern: emit change with the new value
      value = newVal;
      dispatch("change", newVal);
    } catch (err) {
      console.error("Upload failed", err);
      localError = "Upload failed.";
    } finally {
      isUploading = false;
      // reset input so selecting same file again triggers change
      e.currentTarget.value = "";
    }
  }

  function clearFile() {
    value = null;
    dispatch("change", null);
  }
</script>

<FieldShell {field} {id} error={error || localError} required={!!field?.required}>
  {#if value}
    <div class="file-preview">
      <div class="file-name">
        <span class="tag">Current</span>
        <code>{value?.filename ?? value}</code>
      </div>

      <button
        type="button"
        class="pb-btn pb-btn-light"
        on:click={clearFile}
        disabled={disabled || isUploading}
      >
        Remove
      </button>
    </div>
  {/if}

  <input
    id={id}
    name={field?.key}
    class="form-input"
    type="file"
    disabled={disabled || isUploading}
    on:change={handleFileChange}
  />
</FieldShell>

<style>
  .file-preview{
    display:flex;
    align-items:center;
    justify-content:space-between;
    gap:12px;
    margin-bottom:10px;
  }
  .file-name{
    display:flex;
    align-items:center;
    gap:8px;
    min-width:0;
  }
  .tag{
    font-size:12px;
    padding:2px 8px;
    border-radius:999px;
    background: rgba(15,23,42,0.08);
  }
  code{
    white-space:nowrap;
    overflow:hidden;
    text-overflow:ellipsis;
    max-width: 360px;
    display:inline-block;
    vertical-align:bottom;
  }

  /* small PB-ish button */
  .pb-btn{
    border: 1px solid rgba(15,23,42,0.12);
    border-radius: 10px;
    padding: 8px 10px;
    font-size: 13px;
    background: #f8fafc;
    cursor: pointer;
  }
  .pb-btn:disabled{ opacity:.65; cursor:not-allowed; }
  .pb-btn-light{ background:#f8fafc; }
</style>
-->
<script>
  import { createEventDispatcher } from "svelte";
  import ApiClient from "@/utils/ApiClient";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = null;
  export let disabled = false;
  export let error = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${field?.key || "field"}`;

  let isUploading = false;
  let localError = "";

  async function handleFileChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    localError = "";

    try {
      const formData = new FormData();
      formData.append("file", file);

      const rec = await ApiClient.collection("assets").create(formData);

      const newVal = {
        collection: "assets",
        recordId: rec.id,
        filename: rec.file,
      };

      value = newVal;
      dispatch("change", newVal);
    } catch (err) {
      console.error(err);
      localError = "Upload failed.";
    } finally {
      isUploading = false;
      e.currentTarget.value = "";
    }
  }

  function clearFile() {
    value = null;
    dispatch("change", null);
  }
</script>

<FieldShell {field} {id} error={error || localError} required={!!field?.required}>
  {#if value}
    <div class="file-preview">
      <code>{value.filename}</code>

      <button
        type="button"
        class="pb-btn pb-btn-light"
        on:click={clearFile}
        disabled={disabled}
      >
        Remove
      </button>
    </div>
  {/if}

  <!-- hidden real input -->
  <input
    id={id}
    type="file"
    class="file-input-hidden"
    disabled={disabled || isUploading}
    on:change={handleFileChange}
  />

  <!-- fake PocketBase button -->
  <label for={id} class="pb-btn pb-btn-primary">
    {isUploading ? "Uploading…" : "Choose file"}
  </label>
</FieldShell>

<style>
  .file-preview{
    display:flex;
    align-items:center;
    justify-content:space-between;
    gap:12px;
    margin-bottom:10px;
  }
  .file-name{
    display:flex;
    align-items:center;
    gap:8px;
    min-width:0;
  }
  .tag{
    font-size:12px;
    padding:2px 8px;
    border-radius:999px;
    background: rgba(15,23,42,0.08);
  }
  code{
    white-space:nowrap;
    overflow:hidden;
    text-overflow:ellipsis;
    max-width: 360px;
    display:inline-block;
    vertical-align:bottom;
  }

  /* small PB-ish button */
  .pb-btn{
    border: 1px solid rgba(15,23,42,0.12);
    border-radius: 10px;
    padding: 8px 10px;
    font-size: 13px;
    background: #f8fafc;
    cursor: pointer;
  }
    
  .pb-btn:disabled{ opacity:.65; cursor:not-allowed; }
  .pb-btn-light{ background:#f8fafc; }

  .file-input-hidden {
  display: none;
}

.file-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.pb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  padding: 0 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid rgba(15, 23, 42, 0.15);
  cursor: pointer;
  user-select: none;
}

.pb-btn-primary {
  background: #0f172a;
  color: white;
}

.pb-btn-primary:hover {
  background: #020617;
}

.pb-btn-light {
  background: #f8fafc;
}

.pb-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

