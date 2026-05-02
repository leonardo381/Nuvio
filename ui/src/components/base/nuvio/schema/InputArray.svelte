<script>
  import { createEventDispatcher } from "svelte";
  import SchemaForm from "./SchemaForm.svelte";

  export let field;
  export let value = [];
  export let path = "";

  const dispatch = createEventDispatcher();

  let items = Array.isArray(value) ? [...value] : [];
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
    emit();
  }

  function removeItem(index) {
    items = items.filter((_, i) => i !== index);
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

  $: if (Array.isArray(value)) {
    items = [...value];
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
      class="btn btn-sm btn-outline array-field__add-button"
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
        <div class="array-item">
          <div class="array-item__header">
            <div class="array-item__title-wrap">
              <div class="array-item__index">{index + 1}</div>
              <div class="array-item__title">
                {field.itemLabel ?? "Item"} {index + 1}
              </div>
            </div>

            <button
              type="button"
              class="btn btn-sm btn-outline btn-danger array-item__remove-button"
              on:click={() => removeItem(index)}
            >
              Remove
            </button>
          </div>

          <div class="array-item__body">
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
                for={"array-" + String(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-") + "-" + index}
              >
                {field?.item?.label || "Value"}
              </label>
              <input
                id={"array-" + String(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-") + "-" + index}
                class="form-input array-item__primitive-input"
                type={field?.item?.type === "number" ? "number" : "text"}
                value={item ?? ""}
                on:input={(e) => updatePrimitiveItem(index, e.currentTarget.value)}
              />
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .array-field {
    border: 1px solid var(--baseAlt2Color);
    border-radius: var(--baseRadius);
    background: var(--baseColor);
    overflow: hidden;
    margin: 12px 0;
  }

  .array-field__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 12px;
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 62%, var(--baseColor));
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
    border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
    color: var(--txtHintColor);
    background: var(--baseColor);
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
    margin: 10px 12px 12px;
    padding: 12px;
    border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
    border-radius: var(--baseRadius);
    background: color-mix(in srgb, var(--baseAlt1Color) 38%, var(--baseColor));
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
    padding: 10px 12px 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .array-item {
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
    border-radius: var(--baseRadius);
    overflow: hidden;
    background: var(--baseColor);
  }

  .array-item__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 8px 10px;
    border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 84%, transparent);
    background: color-mix(in srgb, var(--baseAlt1Color) 46%, var(--baseColor));
  }

  .array-item__title-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .array-item__index {
    width: 22px;
    height: 22px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
    background: var(--baseColor);
    color: var(--txtHintColor);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 700;
    flex-shrink: 0;
  }

  .array-item__title {
    font-size: var(--smFontSize);
    font-weight: 600;
    color: var(--txtPrimaryColor);
    line-height: 1.2;
  }

  .array-item__remove-button {
    white-space: nowrap;
  }

  .array-item__body {
    padding: 10px;
    background: transparent;
  }

  .array-item__body :global(.pb-field:last-child) {
    margin-bottom: 0;
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

    .array-field__add-button,
    .array-item__remove-button {
      width: 100%;
      justify-content: center;
    }
  }
</style>
