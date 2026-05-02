<script>
  import { createEventDispatcher } from "svelte";
  import ApiClient from "@/utils/ApiClient";
  import OverlayPanel from "@/components/base/OverlayPanel.svelte";
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
    <div class="file-current">
      <div class="file-current-main">
        <span class="label label-sm file-current-label">Current file</span>
        <span class="file-current-name" title={value?.filename ?? value}>{value?.filename ?? value}</span>
      </div>

      <button
        type="button"
        class="btn btn-sm btn-outline file-remove-btn"
        on:click={clearFile}
        disabled={disabled || isUploading}
      >
        Remove
      </button>
    </div>
  {/if}

  <div class="file-actions">
    <div class="file-input-wrap">
      <input
        id={id}
        name={path || field?.key}
        class="form-input file-native-input"
        type="file"
        disabled={disabled || isUploading}
        on:change={handleFileChange}
      />
    </div>

    <button
      type="button"
      class="btn btn-sm btn-outline"
      on:click={openPicker}
      disabled={disabled || isUploading}
    >
      Choose existing
    </button>
  </div>

  {#if showPicker}
    <OverlayPanel
      popup
      class="overlay-panel-xl schema-file-picker"
      active={true}
      overlayClose={true}
      escClose={false}
      btnClose={false}
      on:hide={closePicker}
    >
      <svelte:fragment slot="header">
        <div class="schema-file-picker-head">
          <h5 class="m-0">Choose existing asset</h5>
          <span class="txt-sm txt-hint">Select a file from Assets.</span>
        </div>
        <button type="button" class="btn btn-sm btn-outline" on:click={closePicker}>
          Close
        </button>
      </svelte:fragment>

      <div class="schema-file-picker-toolbar">
        <input
          class="form-input"
          type="text"
          bind:value={assetSearch}
          placeholder="Search by filename..."
          on:input={handleSearchInput}
        />
      </div>

      {#if isLoadingAssets}
        <div class="schema-file-picker-empty">Loading assets...</div>
      {:else if pickerError}
        <div class="schema-file-picker-empty">{pickerError}</div>
      {:else if assets.length === 0}
        <div class="schema-file-picker-empty">No assets found.</div>
      {:else}
        <div class="schema-file-asset-grid">
          {#each assets as asset}
            <button
              type="button"
              class="schema-file-asset-card"
              on:click={() => chooseExisting(asset)}
            >
              <div class="schema-file-asset-thumb">
                {#if asset.file}
                  <img src={assetUrl(asset)} alt={asset.originalName || asset.file} />
                {/if}
              </div>

              <div class="schema-file-asset-meta">
                <div class="schema-file-asset-name">{asset.originalName || asset.file}</div>
                <div class="schema-file-asset-file">{asset.file}</div>
              </div>
            </button>
          {/each}
        </div>
      {/if}

      <svelte:fragment slot="footer">
        <button type="button" class="btn btn-sm btn-outline" on:click={closePicker}>
          Close
        </button>
      </svelte:fragment>
    </OverlayPanel>
  {/if}
</FieldShell>

<style>
  .file-current {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 74%, transparent);
    border-radius: var(--baseRadius);
    background: color-mix(in srgb, var(--baseAlt1Color) 12%, var(--baseColor));
    padding: 8px 10px;
    margin-bottom: 9px;
  }

  .file-current-main {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .file-current-label {
    min-height: 18px;
    border-color: color-mix(in srgb, var(--baseAlt2Color) 76%, transparent);
    color: color-mix(in srgb, var(--txtHintColor) 96%, var(--txtPrimaryColor));
    background: color-mix(in srgb, var(--baseAlt1Color) 20%, var(--baseColor));
  }

  .file-current-name {
    min-width: 0;
    display: inline-block;
    max-width: min(100%, 420px);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--txtPrimaryColor);
    font-size: var(--smFontSize);
  }

  .file-actions {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 72%, transparent);
    border-radius: var(--baseRadius);
    background: var(--baseColor);
    padding: 9px 10px;
  }

  .file-input-wrap {
    flex: 1 1 260px;
    min-width: 0;
  }

  .file-remove-btn {
    color: color-mix(in srgb, var(--dangerColor) 62%, var(--txtHintColor));
    border-color: color-mix(in srgb, var(--dangerColor) 20%, var(--baseAlt2Color));
    background: color-mix(in srgb, var(--baseAlt1Color) 8%, var(--baseColor));
  }

  .file-remove-btn:hover,
  .file-remove-btn:focus-visible {
    color: color-mix(in srgb, var(--dangerColor) 78%, var(--txtHintColor));
    border-color: color-mix(in srgb, var(--dangerColor) 34%, var(--baseAlt2Color));
    background: color-mix(in srgb, var(--dangerColor) 6%, var(--baseColor));
  }

  .file-native-input {
    width: 100%;
    height: 44px;
    padding: 8px 10px;
    display: flex;
    align-items: center;
    box-sizing: border-box;
    line-height: 1;
    background: var(--baseColor);
    border-color: color-mix(in srgb, var(--baseAlt2Color) 78%, transparent);
  }

  .file-native-input::file-selector-button {
    height: 28px;
    padding: 0 12px;
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 76%, transparent);
    border-radius: 8px;
    background: color-mix(in srgb, var(--baseAlt1Color) 16%, var(--baseColor));
    color: var(--txtPrimaryColor);
    font-size: var(--smFontSize);
    cursor: pointer;
  }

  :global(.overlay-panel.schema-file-picker) {
    width: min(92vw, 980px);
    max-height: 88vh;
  }

  .schema-file-picker-head {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .schema-file-picker-toolbar {
    margin-bottom: 12px;
  }

  .schema-file-picker-empty {
    padding: 22px 8px;
    text-align: center;
    color: var(--txtHintColor);
    font-size: var(--smFontSize);
  }

  .schema-file-asset-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 10px;
  }

  .schema-file-asset-card {
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 74%, transparent);
    background: var(--baseColor);
    border-radius: 10px;
    padding: 9px;
    text-align: left;
    cursor: pointer;
    transition: background-color var(--baseAnimationSpeed), border-color var(--baseAnimationSpeed);
  }

  .schema-file-asset-card:hover,
  .schema-file-asset-card:focus-visible {
    background: color-mix(in srgb, var(--baseAlt1Color) 22%, var(--baseColor));
    border-color: color-mix(in srgb, var(--primaryColor) 28%, var(--baseAlt2Color));
  }

  .schema-file-asset-thumb {
    width: 100%;
    aspect-ratio: 1 / 1;
    border-radius: 10px;
    overflow: hidden;
    background: color-mix(in srgb, var(--baseAlt1Color) 34%, var(--baseColor));
    margin-bottom: 8px;
  }

  .schema-file-asset-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .schema-file-asset-meta {
    min-width: 0;
  }

  .schema-file-asset-name,
  .schema-file-asset-file {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .schema-file-asset-name {
    font-size: 13px;
    color: var(--txtPrimaryColor);
  }

  .schema-file-asset-file {
    font-size: 12px;
    color: var(--txtHintColor);
  }

  @media (max-width: 720px) {
    .file-current {
      flex-wrap: wrap;
    }

    .file-remove-btn,
    .file-actions > .btn {
      width: 100%;
      justify-content: center;
    }
  }
</style>
