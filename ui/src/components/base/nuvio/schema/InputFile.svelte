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
  const DEFERRED_HINT = "File uploads are managed by an administrator for now.";
  const SCOPED_WEBSITE_REQUIRED_HINT = "Select a website before managing files.";
  const SCOPED_ALLOWED_MIME_TYPES = new Set(["image/jpeg", "image/png", "image/webp", "image/gif"]);
  const SCOPED_MAX_FILE_SIZE_BYTES = 8 * 1024 * 1024;
  const SCOPED_MAX_FILE_SIZE_MB = 8;

  $: id = `schema-${(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-")}`;
  $: scopedAssetMode = toBooleanFlag(field?.nuvioUseScopedAssets);
  $: scopedAssetWebsiteId = scopedAssetMode ? `${field?.nuvioAssetWebsiteId || ""}`.trim() : "";
  $: assetActionsDeferred = !!field?.nuvioDisableAssetActions || (!scopedAssetMode && ApiClient.isClientSuperuser());
  $: fieldDisabled = disabled || !!field?.disabled || !!field?.readonly;
  $: assetActionsDisabled = fieldDisabled || assetActionsDeferred;
  $: resolvedHint = assetActionsDeferred
    ? (field?.hint || DEFERRED_HINT)
    : (field?.hint || "");

  let isUploading = false;
  let localError = "";

  let showPicker = false;
  let isLoadingAssets = false;
  let assetSearch = "";
  let assets = [];
  let pickerError = "";

  $: if (assetActionsDeferred && showPicker) {
    closePicker();
  }

  $: if (scopedAssetMode && !scopedAssetWebsiteId && showPicker) {
    closePicker();
  }

  $: if (assetActionsDeferred) {
    localError = "";
    pickerError = "";
  }

  function toBooleanFlag(value) {
    if (typeof value === "string") {
      const normalized = value.trim().toLowerCase();
      return normalized === "true" || normalized === "1" || normalized === "yes";
    }

    return !!value;
  }

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

  function getScopedAssetFileRef(asset) {
    const recordId = `${asset?.id || asset?.file?.recordId || ""}`.trim();
    const filename = `${asset?.filename || asset?.file?.filename || ""}`.trim();
    const collection = `${asset?.collection || asset?.file?.collection || ASSET_COLLECTION}`.trim() || ASSET_COLLECTION;

    if (!recordId || !filename) {
      return null;
    }

    return {
      recordId,
      filename,
      collection,
    };
  }

  function getScopedAssetDisplayName(asset) {
    const originalName = `${asset?.originalName || ""}`.trim();
    if (originalName) {
      return originalName;
    }

    return getScopedAssetFileRef(asset)?.filename || "";
  }

  function getScopedAssetSearchValue(asset) {
    const displayName = getScopedAssetDisplayName(asset).toLowerCase();
    const filename = (getScopedAssetFileRef(asset)?.filename || "").toLowerCase();
    return `${displayName} ${filename}`.trim();
  }

  function filterScopedAssetsBySearch(items, searchTerm) {
    const normalizedSearch = `${searchTerm || ""}`.trim().toLowerCase();
    if (!normalizedSearch) {
      return Array.isArray(items) ? items : [];
    }

    return (Array.isArray(items) ? items : []).filter((asset) => {
      const searchValue = getScopedAssetSearchValue(asset);
      return searchValue.includes(normalizedSearch);
    });
  }

  function mapScopedAssetDtoToFieldValue(asset) {
    const fileRef = getScopedAssetFileRef(asset);
    if (!fileRef) {
      return null;
    }

    return {
      collection: fileRef.collection,
      recordId: fileRef.recordId,
      filename: fileRef.filename,
    };
  }

  function isScopedAssetFileAllowed(file) {
    const mimeType = `${file?.type || ""}`.trim().toLowerCase();
    return SCOPED_ALLOWED_MIME_TYPES.has(mimeType);
  }

  function isScopedAssetFileSizeAllowed(file) {
    const size = Number(file?.size || 0);
    return size > 0 && size <= SCOPED_MAX_FILE_SIZE_BYTES;
  }

  async function handleFileChange(e) {
    if (assetActionsDisabled) {
      if (e?.currentTarget) {
        e.currentTarget.value = "";
      }
      return;
    }

    if (scopedAssetMode && !scopedAssetWebsiteId) {
      localError = SCOPED_WEBSITE_REQUIRED_HINT;
      if (e?.currentTarget) {
        e.currentTarget.value = "";
      }
      return;
    }

    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    localError = "";

    try {
      if (scopedAssetMode) {
        if (!isScopedAssetFileAllowed(file)) {
          localError = "Unsupported file type. Allowed types: JPG, PNG, WebP, GIF.";
          return;
        }
        if (!isScopedAssetFileSizeAllowed(file)) {
          localError = `File exceeds the maximum allowed size of ${SCOPED_MAX_FILE_SIZE_MB}MB.`;
          return;
        }

        const asset = await ApiClient.uploadCMSAsset(
          {
            websiteId: scopedAssetWebsiteId,
            file,
          },
          {
            requestKey: `nuvio_cms_asset_upload_${scopedAssetWebsiteId || "unknown"}`,
          },
        );

        const scopedValue = mapScopedAssetDtoToFieldValue(asset);
        if (!scopedValue) {
          throw new Error("Invalid asset response.");
        }

        value = scopedValue;
        dispatch("change", scopedValue);
      } else {
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
      }
    } catch (err) {
      console.error("Upload failed", err);
      localError = err?.response?.message || err?.data?.message || err?.message || "Upload failed.";
    } finally {
      isUploading = false;
      e.currentTarget.value = "";
    }
  }

  function clearFile() {
    if (assetActionsDisabled) {
      return;
    }

    value = null;
    dispatch("change", null);
  }

  async function openPicker() {
    if (assetActionsDisabled) {
      return;
    }

    if (scopedAssetMode && !scopedAssetWebsiteId) {
      localError = SCOPED_WEBSITE_REQUIRED_HINT;
      return;
    }

    showPicker = true;
    localError = "";
    await loadAssets();
  }

  async function loadAssets() {
    if (assetActionsDisabled) {
      closePicker();
      return;
    }

    if (scopedAssetMode && !scopedAssetWebsiteId) {
      closePicker();
      localError = SCOPED_WEBSITE_REQUIRED_HINT;
      return;
    }

    isLoadingAssets = true;
    pickerError = "";

    try {
      if (scopedAssetMode) {
        const scopedAssets = await ApiClient.getCMSAssets({
          websiteId: scopedAssetWebsiteId,
          requestKey: `nuvio_cms_assets_picker_${scopedAssetWebsiteId || "unknown"}`,
        });
        assets = filterScopedAssetsBySearch(scopedAssets, assetSearch);
      } else {
        const filter = assetSearch?.trim()
          ? `originalName ~ "${assetSearch.replace(/"/g, '\\"')}" || file ~ "${assetSearch.replace(/"/g, '\\"')}"`
          : "";

        assets = await ApiClient.collection(ASSET_COLLECTION).getFullList({
          sort: "-created",
          filter
        });
      }
    } catch (err) {
      console.error("Failed to load assets", err);
      const statusCode = Number(err?.status || err?.response?.status || 0);
      if (scopedAssetMode && statusCode === 403) {
        pickerError = "You do not have access to assets for this website.";
      } else {
        pickerError = err?.response?.message || err?.data?.message || err?.message || "Failed to load existing assets.";
      }
      assets = [];
    } finally {
      isLoadingAssets = false;
    }
  }

  function chooseExisting(asset) {
    if (assetActionsDisabled) {
      return;
    }

    const selectedVal = scopedAssetMode
      ? mapScopedAssetDtoToFieldValue(asset)
      : {
          collection: ASSET_COLLECTION,
          recordId: asset.id,
          filename: asset.file,
        };

    if (!selectedVal) {
      pickerError = "Selected asset is missing a valid file reference.";
      return;
    }

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
    if (assetActionsDisabled) {
      return;
    }

    await loadAssets();
  }

  function assetUrl(asset) {
    const scopedFileRef = getScopedAssetFileRef(asset);
    const filename = scopedFileRef?.filename || `${asset?.file || ""}`.trim();
    const recordId = scopedFileRef?.recordId || `${asset?.id || ""}`.trim();
    const collectionName = scopedFileRef?.collection || `${asset?.collection || ASSET_COLLECTION}`.trim() || ASSET_COLLECTION;

    if (!filename || !recordId || !collectionName) {
      return "";
    }

    const pbFileURL = ApiClient.files.getURL?.(
      {
        id: recordId,
        collectionName,
      },
      filename,
    );
    if (pbFileURL) {
      return pbFileURL;
    }

    const backendURL = `${import.meta.env.VITE_PB_BACKEND_URL || ""}`.trim().replace(/\/+$/, "");
    if (!backendURL) {
      return "";
    }

    return `${backendURL}/api/files/${encodeURIComponent(collectionName)}/${encodeURIComponent(recordId)}/${encodeURIComponent(filename)}`;
  }
</script>

<FieldShell {field} {id} error={error || localError} required={!!field?.required} hint={resolvedHint}>
  {#if value}
    <div class="file-current">
      <div class="file-current-main">
        <span class="label label-sm file-current-label">Current file</span>
        <span class="file-current-name" title={value?.filename ?? value}>{value?.filename ?? value}</span>
      </div>

      {#if !assetActionsDeferred}
        <button
          type="button"
          class="btn btn-sm btn-outline file-remove-btn"
          on:click={clearFile}
          disabled={assetActionsDisabled || isUploading}
        >
          Remove
        </button>
      {/if}
    </div>
  {/if}

  {#if !assetActionsDeferred}
    <div class="file-actions">
      <div class="file-input-wrap">
        <input
          id={id}
          name={path || field?.key}
          class="form-input file-native-input"
          type="file"
          disabled={fieldDisabled || isUploading}
          on:change={handleFileChange}
        />
      </div>

      <button
        type="button"
        class="btn btn-sm btn-outline"
        on:click={openPicker}
        disabled={fieldDisabled || isUploading}
      >
        Choose existing
      </button>
    </div>
  {/if}

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
        <div class="schema-file-picker-header">
          <div class="schema-file-picker-head">
            <h5 class="m-0">Choose existing asset</h5>
            <span class="txt-sm txt-hint">Select a file from Assets.</span>
          </div>
          <button
            type="button"
            class="btn btn-sm btn-outline schema-file-picker-close-btn"
            on:click={closePicker}
          >
            Close
          </button>
        </div>
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
            {@const scopedFileRef = getScopedAssetFileRef(asset)}
            {@const displayFilename = scopedFileRef?.filename || `${asset?.file || ""}`.trim()}
            {@const displayName = getScopedAssetDisplayName(asset) || displayFilename}
            <button
              type="button"
              class="schema-file-asset-card"
              on:click={() => chooseExisting(asset)}
            >
              <div class="schema-file-asset-thumb">
                {#if displayFilename}
                  <img src={assetUrl(asset)} alt={displayName || displayFilename} />
                {/if}
              </div>

              <div class="schema-file-asset-meta">
                <div class="schema-file-asset-name">{displayName}</div>
                <div class="schema-file-asset-file">{displayFilename}</div>
              </div>
            </button>
          {/each}
        </div>
      {/if}

      <svelte:fragment slot="footer">
        <div class="schema-file-picker-footer">
          <button
            type="button"
            class="btn btn-sm btn-outline"
            on:click={closePicker}
          >
            Close
          </button>
        </div>
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

  .schema-file-picker-header {
    width: 100%;
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
  }

  .schema-file-picker-close-btn {
    margin-left: auto;
    flex: 0 0 auto;
  }

  .schema-file-picker-footer {
    width: 100%;
    display: flex;
    justify-content: flex-end;
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
