<script>
  import { createEventDispatcher } from "svelte";
  import ApiClient from "@/utils/ApiClient";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = null; // { collection, recordId, filename } OR string OR null
  export let disabled = false;
  export let error = "";
  export let path = "";

  const dispatch = createEventDispatcher();

  const ASSET_COLLECTION = "Assets";

  $: id = `schema-${(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-")}`;

  let isUploading = false;
  let localError = "";

  let showPicker = false;
  let isLoadingAssets = false;
  let assetSearch = "";
  let assets = [];
  let pickerError = "";

  async function sha256(file) {
    const buffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest("SHA-256", buffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
  }

  async function findExistingAssetByChecksum(checksum) {
    try {
      return await ApiClient
        .collection(ASSET_COLLECTION)
        .getFirstListItem(`checksum="${checksum}"`);
    } catch {
      return null;
    }
  }

  async function handleFileChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    localError = "";

    try {
      const checksum = await sha256(file);

      const existing = await findExistingAssetByChecksum(checksum);

      if (existing) {
        const reusedVal = {
          collection: ASSET_COLLECTION,
          recordId: existing.id,
          filename: existing.file,
        };

        value = reusedVal;
        dispatch("change", reusedVal);
        return;
      }

      const formData = new FormData();
      formData.append("file", file);
      formData.append("checksum", checksum);
      formData.append("originalName", file.name);
      formData.append("mimeType", file.type || "");
      formData.append("size", String(file.size));

      const rec = await ApiClient.collection(ASSET_COLLECTION).create(formData);

      const newVal = {
        collection: ASSET_COLLECTION,
        recordId: rec.id,
        filename: rec.file,
      };

      value = newVal;
      dispatch("change", newVal);
    } catch (err) {
      console.error("Upload failed", err);
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

  async function openPicker() {
    showPicker = true;
    await loadAssets();
  }

  async function loadAssets() {
    isLoadingAssets = true;
    pickerError = "";

    try {
      const filter = assetSearch?.trim()
        ? `originalName ~ "${assetSearch.replace(/"/g, '\\"')}" || file ~ "${assetSearch.replace(/"/g, '\\"')}"`
        : "";

      assets = await ApiClient.collection(ASSET_COLLECTION).getFullList({
        sort: "-created",
        filter
      });
    } catch (err) {
      console.error("Failed to load assets", err);
      pickerError = "Failed to load existing assets.";
      assets = [];
    } finally {
      isLoadingAssets = false;
    }
  }

  function chooseExisting(asset) {
    const selectedVal = {
      collection: ASSET_COLLECTION,
      recordId: asset.id,
      filename: asset.file,
    };

    value = selectedVal;
    dispatch("change", selectedVal);
    showPicker = false;
  }

  function closePicker() {
    showPicker = false;
    pickerError = "";
    assetSearch = "";
  }

  async function handleSearchInput() {
    await loadAssets();
  }

  function assetUrl(asset) {
    return ApiClient.files.getURL?.(asset, asset.file) ?? "";
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

  <div class="file-actions">
    <input
      id={id}
      name={path || field?.key}
      class="form-input"
      type="file"
      disabled={disabled || isUploading}
      on:change={handleFileChange}
    />

    <button
      type="button"
      class="pb-btn pb-btn-light"
      on:click={openPicker}
      disabled={disabled || isUploading}
    >
      Choose existing
    </button>
  </div>

  {#if showPicker}
    <div class="picker-backdrop" on:click={closePicker}></div>

    <div class="picker-modal">
      <div class="picker-header">
        <strong>Choose existing asset</strong>
        <button type="button" class="pb-btn pb-btn-light" on:click={closePicker}>
          Close
        </button>
      </div>

      <div class="picker-toolbar">
        <input
          class="form-input"
          type="text"
          bind:value={assetSearch}
          placeholder="Search by filename..."
          on:input={handleSearchInput}
        />
      </div>

      {#if isLoadingAssets}
        <div class="picker-empty">Loading assets...</div>
      {:else if pickerError}
        <div class="picker-empty">{pickerError}</div>
      {:else if assets.length === 0}
        <div class="picker-empty">No assets found.</div>
      {:else}
        <div class="asset-grid">
          {#each assets as asset}
            <button
              type="button"
              class="asset-card"
              on:click={() => chooseExisting(asset)}
            >
              <div class="asset-thumb">
                {#if asset.file}
                  <img src={assetUrl(asset)} alt={asset.originalName || asset.file} />
                {/if}
              </div>

              <div class="asset-meta">
                <div class="asset-name">{asset.originalName || asset.file}</div>
                <div class="asset-file">{asset.file}</div>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
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

  .file-actions {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .file-actions .form-input {
    flex: 1;
  }

  .pb-btn{
    border: 1px solid rgba(15,23,42,0.12);
    border-radius: 10px;
    padding: 8px 10px;
    font-size: 13px;
    background: #f8fafc;
    cursor: pointer;
    white-space: nowrap;
  }

  .pb-btn:disabled{
    opacity:.65;
    cursor:not-allowed;
  }

  .pb-btn-light{
    background:#f8fafc;
  }

  .form-input[type="file"] {
    height: 46px;
    padding: 8px 10px;
    display: flex;
    align-items: center;
    box-sizing: border-box;
    line-height: 1;
  }

  .form-input[type="file"]::file-selector-button {
    height: 28px;
    padding: 0 12px;
    border: 1px solid rgba(15,23,42,0.12);
    border-radius: 8px;
    background: #f8fafc;
    font-size: 13px;
    cursor: pointer;
  }

  .picker-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.35);
    z-index: 999;
  }

  .picker-modal {
    position: fixed;
    inset: 50% auto auto 50%;
    transform: translate(-50%, -50%);
    width: min(900px, 92vw);
    max-height: 80vh;
    overflow: auto;
    background: white;
    border: 1px solid rgba(15,23,42,0.12);
    border-radius: 16px;
    padding: 16px;
    z-index: 1000;
    box-shadow: 0 20px 60px rgba(15, 23, 42, 0.18);
  }

  .picker-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .picker-toolbar {
    margin-bottom: 12px;
  }

  .picker-empty {
    padding: 24px 8px;
    text-align: center;
    color: #64748b;
  }

  .asset-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 12px;
  }

  .asset-card {
    border: 1px solid rgba(15,23,42,0.12);
    background: #fff;
    border-radius: 12px;
    padding: 10px;
    text-align: left;
    cursor: pointer;
  }

  .asset-card:hover {
    background: #f8fafc;
  }

  .asset-thumb {
    width: 100%;
    aspect-ratio: 1 / 1;
    border-radius: 10px;
    overflow: hidden;
    background: #f1f5f9;
    margin-bottom: 8px;
  }

  .asset-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .asset-meta {
    min-width: 0;
  }

  .asset-name,
  .asset-file {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .asset-name {
    font-size: 13px;
    color: #0f172a;
  }

  .asset-file {
    font-size: 12px;
    color: #64748b;
  }
</style>