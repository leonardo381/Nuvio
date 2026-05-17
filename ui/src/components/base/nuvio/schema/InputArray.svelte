<script>
  import { createEventDispatcher } from "svelte";
  import SchemaForm from "./SchemaForm.svelte";
  import CommonHelper from "@/utils/CommonHelper";

  export let field;
  export let value = [];
  export let path = "";

  const dispatch = createEventDispatcher();

  let items = Array.isArray(value) ? [...value] : [];
  let openItemIndex = -1;
  $: supportsObjectItems = Array.isArray(field?.item?.fields) && field.item.fields.length > 0;

  function emit() {
    dispatch("change", { value: items });
  }

  function createEmptyItem() {
    if (!supportsObjectItems) {
      if (field?.item?.type === "bool") {
        return false;
      }

      return "";
    }

    const fields = field?.item?.fields ?? [];
    const obj = {};

    for (const f of fields) {
      if (f.type === "array") {
        obj[f.key] = [];
      } else if (f.type === "object") {
        obj[f.key] = {};
      } else if (f.type === "bool") {
        obj[f.key] = false;
      } else {
        obj[f.key] = "";
      }
    }

    return obj;
  }

  function addItem() {
    items = [...items, createEmptyItem()];
    openItemIndex = items.length - 1;
    emit();
  }

  function removeItem(index) {
    items = items.filter((_, i) => i !== index);
    if (openItemIndex === index) {
      openItemIndex = -1;
    } else if (openItemIndex > index) {
      openItemIndex -= 1;
    }
    emit();
  }

  function updateItem(index, nextValue) {
    items = items.map((item, i) => (i === index ? nextValue : item));
    emit();
  }

  function updatePrimitiveItem(index, rawValue) {
    let nextValue = rawValue;

    if (field?.item?.type === "bool") {
      nextValue = !!rawValue;
    } else if (field?.item?.type === "number") {
      nextValue = rawValue === "" ? "" : Number(rawValue);
    } else {
      nextValue = String(rawValue ?? "").trim();
    }

    updateItem(index, nextValue);
  }

  function toggleItem(index) {
    openItemIndex = openItemIndex === index ? -1 : index;
  }

  function normalizeText(value) {
    const raw = String(value ?? "");
    if (!raw) {
      return "";
    }

    const plain = raw.includes("<") ? CommonHelper.plainText(raw) : raw;
    return plain.replace(/\s+/g, " ").trim();
  }

  function getNamedObjectValue(item, keys = []) {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      return "";
    }

    for (const key of keys) {
      const value = normalizeText(item?.[key]);
      if (value) {
        return value;
      }
    }

    return "";
  }

  function truncateText(value, maxLength = 110) {
    const text = normalizeText(value);
    if (text.length <= maxLength) {
      return text;
    }
    return `${text.slice(0, maxLength).trim()}...`;
  }

  function getObjectItemSummary(item, index) {
    const title =
      getNamedObjectValue(item, ["title", "heading", "label", "name", "question", "text"]) ||
      `${field.itemLabel ?? "Item"} ${index + 1}`;
    const snippet = getNamedObjectValue(item, ["answer", "description", "body", "content"]);
    return {
      title: truncateText(title, 90),
      snippet: truncateText(snippet, 120),
    };
  }

  function getPrimitiveItemSummary(item, index) {
    const defaultTitle = `${field.itemLabel ?? "Item"} ${index + 1}`;

    if (field?.item?.type === "bool") {
      return {
        title: defaultTitle,
        snippet: item ? "Enabled" : "Disabled",
      };
    }

    const text = truncateText(item, 90);
    if (!text) {
      return {
        title: defaultTitle,
        snippet: "",
      };
    }

    return {
      title: text,
      snippet: "",
    };
  }

  function getItemSummary(item, index) {
    if (supportsObjectItems) {
      return getObjectItemSummary(item, index);
    }
    return getPrimitiveItemSummary(item, index);
  }

  function toSafePathSegment(value) {
    return String(value || "field").replace(/[^a-zA-Z0-9_-]/g, "-");
  }

  function getItemDomId(index) {
    return `array-item-${toSafePathSegment(path || field?.key)}-${index}`;
  }

  function getPrimitiveInputId(index) {
    return `array-${toSafePathSegment(path || field?.key)}-${index}`;
  }

  $: if (Array.isArray(value)) {
    items = [...value];
    if (openItemIndex >= items.length) {
      openItemIndex = -1;
    }
  }
</script>

<div class="array-field">
  <div class="array-field__header">
    <div class="array-field__header-left">
      <div class="array-field__title">{field.label}</div>
      <div class="label label-sm array-field__count">
        {items.length} {items.length === 1 ? "item" : "items"}
      </div>
    </div>

    <button
      type="button"
      class="btn btn-sm array-field__add-button"
      on:click={addItem}
    >
      <span class="array-field__add-icon">+</span>
      <span>Add item</span>
    </button>
  </div>

  {#if items.length === 0}
    <div class="array-field__empty">
      <div class="array-field__empty-title">No items added yet</div>
      <div class="array-field__empty-text">
        Use the button above to create the first item.
      </div>
    </div>
  {:else}
    <div class="array-field__items">
      {#each items as item, index}
        {@const summary = getItemSummary(item, index)}
        <div class="array-item" class:is-open={openItemIndex === index}>
          <div class="array-item__header">
            <button
              type="button"
              class="array-item__toggle"
              aria-expanded={openItemIndex === index}
              aria-controls={getItemDomId(index)}
              on:click={() => toggleItem(index)}
            >
              <div class="array-item__title-wrap">
                <div class="array-item__index">{index + 1}</div>
                <div class="array-item__title-content">
                  <div class="array-item__title">{summary.title}</div>
                  {#if summary.snippet}
                    <div class="array-item__snippet">{summary.snippet}</div>
                  {/if}
                </div>
              </div>
              <i class="ri-arrow-down-s-line array-item__chevron" class:is-open={openItemIndex === index} />
            </button>
          </div>

          {#if openItemIndex === index}
            <div class="array-item__body" id={getItemDomId(index)}>
              {#if supportsObjectItems}
                <SchemaForm
                  fields={field?.item?.fields ?? []}
                  value={item}
                  showImport={false}
                  path={`${path}[${index}]`}
                  on:change={(e) => updateItem(index, e.detail.value)}
                />
              {:else if field?.item?.type === "bool"}
                <label class="array-item__primitive-checkbox">
                  <input
                    type="checkbox"
                    checked={!!item}
                    on:change={(e) => updatePrimitiveItem(index, e.currentTarget.checked)}
                  />
                  <span>{field?.item?.label || "Enabled"}</span>
                </label>
              {:else}
                <label
                  class="array-item__primitive-label"
                  for={getPrimitiveInputId(index)}
                >
                  {field?.item?.label || "Value"}
                </label>
                <input
                  id={getPrimitiveInputId(index)}
                  class="form-input array-item__primitive-input"
                  type={field?.item?.type === "number" ? "number" : "text"}
                  value={item ?? ""}
                  on:input={(e) => updatePrimitiveItem(index, e.currentTarget.value)}
                />
              {/if}

              <div class="array-item__body-actions">
                <button
                  type="button"
                  class="btn btn-sm btn-outline array-item__remove-button"
                  on:click={() => removeItem(index)}
                >
                  <i class="ri-delete-bin-line" aria-hidden="true" />
                  <span>Remove item</span>
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .array-field {
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 66%, transparent);
    border-radius: var(--baseRadius);
    background: var(--baseColor);
    overflow: hidden;
    margin: 8px 0 12px;
  }

  .array-field__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 12px;
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 70%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 14%, var(--baseColor));
  }

  .array-field__header-left {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .array-field__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: 1.2;
  }

  .array-field__count {
    min-height: 20px;
    border-color: color-mix(in srgb, var(--baseAlt2Color) 78%, transparent);
    color: color-mix(in srgb, var(--txtHintColor) 96%, var(--txtPrimaryColor));
    background: color-mix(in srgb, var(--baseAlt1Color) 20%, var(--baseColor));
  }

  .array-field__add-button {
    gap: 5px;
    white-space: nowrap;
  }

  .array-field__add-icon {
    font-size: 13px;
    line-height: 1;
    font-weight: 700;
  }

  .array-field__empty {
    margin: 9px 10px 10px;
    padding: 10px 11px;
    border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 70%, transparent);
    border-radius: var(--baseRadius);
    background: color-mix(in srgb, var(--baseAlt1Color) 12%, var(--baseColor));
    text-align: left;
  }

  .array-field__empty-title {
    font-size: var(--smFontSize);
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: var(--smLineHeight);
  }

  .array-field__empty-text {
    margin-top: 2px;
    font-size: var(--smFontSize);
    line-height: var(--smLineHeight);
    color: var(--txtHintColor);
  }

  .array-field__items {
    padding: 9px 10px 10px;
    display: flex;
    flex-direction: column;
    gap: 7px;
  }

  .array-item {
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 62%, transparent);
    border-radius: calc(var(--baseRadius) - 1px);
    overflow: hidden;
    background: color-mix(in srgb, var(--baseAlt1Color) 8%, var(--baseColor));
  }

  .array-item.is-open {
    border-color: color-mix(in srgb, var(--primaryColor) 28%, var(--baseAlt2Color));
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--primaryColor) 12%, transparent);
  }

  .array-item__header {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    gap: 10px;
    padding: 8px 10px 8px 9px;
    background: color-mix(in srgb, var(--baseAlt1Color) 16%, var(--baseColor));
  }

  .array-item.is-open .array-item__header {
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 62%, transparent);
  }

  .array-item__toggle {
    width: 100%;
    min-width: 0;
    border: 0;
    background: transparent;
    padding: 0;
    margin: 0;
    color: inherit;
    text-align: left;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .array-item__title-wrap {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    min-width: 0;
    flex: 1 1 auto;
  }

  .array-item__index {
    width: 20px;
    height: 20px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 76%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 18%, var(--baseColor));
    color: var(--txtHintColor);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 700;
    flex-shrink: 0;
  }

  .array-item__title-content {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .array-item__title {
    font-size: var(--smFontSize);
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: 1.25;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .array-item__snippet {
    font-size: 11px;
    line-height: 1.3;
    color: color-mix(in srgb, var(--txtHintColor) 97%, var(--txtPrimaryColor));
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .array-item__chevron {
    font-size: 16px;
    color: var(--txtHintColor);
    transition: transform var(--baseAnimationSpeed), color var(--baseAnimationSpeed);
    flex: 0 0 auto;
  }

  .array-item__chevron.is-open {
    transform: rotate(180deg);
    color: var(--txtPrimaryColor);
  }

  .array-item__remove-button {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    min-height: 28px;
    color: color-mix(in srgb, var(--dangerColor) 56%, var(--txtHintColor));
    border-color: color-mix(in srgb, var(--dangerColor) 18%, var(--baseAlt2Color));
    background: color-mix(in srgb, var(--baseAlt1Color) 6%, var(--baseColor));
  }

  .array-item__remove-button:hover,
  .array-item__remove-button:focus-visible {
    color: color-mix(in srgb, var(--dangerColor) 72%, var(--txtHintColor));
    border-color: color-mix(in srgb, var(--dangerColor) 28%, var(--baseAlt2Color));
    background: color-mix(in srgb, var(--dangerColor) 6%, var(--baseColor));
  }

  .array-item__body {
    padding: 10px 11px 11px;
    background: color-mix(in srgb, var(--baseColor) 94%, var(--baseAlt1Color));
  }

  .array-item__body :global(.pb-field) {
    margin-bottom: 11px;
  }

  .array-item__body :global(.pb-field:last-of-type) {
    margin-bottom: 0;
  }

  .array-item__body :global(.pb-label) {
    margin-bottom: 5px;
  }

  .array-item__body :global(.tinymce-wrapper) {
    margin-top: 2px;
  }

  .array-item__body :global(.tox.tox-tinymce) {
    border-color: color-mix(in srgb, var(--baseAlt2Color) 68%, transparent);
    border-radius: calc(var(--baseRadius) - 1px);
  }

  .array-item__body-actions {
    margin-top: 10px;
    padding-top: 9px;
    border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 62%, transparent);
    display: flex;
    justify-content: flex-end;
  }

  .array-item__primitive-label {
    display: inline-block;
    margin-bottom: 6px;
    font-size: var(--smFontSize);
    font-weight: 600;
    color: var(--txtHintColor);
  }

  .array-item__primitive-input {
    width: 100%;
  }

  .array-item__primitive-checkbox {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: var(--smFontSize);
    line-height: var(--smLineHeight);
    color: var(--txtPrimaryColor);
  }

  @media (max-width: 720px) {
    .array-field__header,
    .array-item__header {
      flex-wrap: wrap;
    }

    .array-item__toggle {
      width: 100%;
    }

    .array-field__add-button {
      width: 100%;
      justify-content: center;
    }

    .array-item__body-actions {
      justify-content: stretch;
    }

    .array-item__remove-button {
      width: 100%;
      justify-content: center;
    }
  }
</style>
