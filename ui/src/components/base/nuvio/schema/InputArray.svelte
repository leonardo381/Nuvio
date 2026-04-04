<script>
  import { createEventDispatcher } from "svelte";
  import SchemaForm from "./SchemaForm.svelte";

  export let field;
  export let value = [];

  const dispatch = createEventDispatcher();

  let items = Array.isArray(value) ? [...value] : [];

  function emit() {
    dispatch("change", { value: items });
  }

    function createEmptyItem() {
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

  $: if (Array.isArray(value)) {
    items = [...value];
  }
</script>

<div class="array-field">
  <div class="array-field__header">
    <div class="array-field__header-left">
      <div class="array-field__title">{field.label}</div>
      <div class="array-field__count">
        {items.length} {items.length === 1 ? "item" : "items"}
      </div>
    </div>

    <button
      type="button"
      class="array-field__add-button"
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
              class="array-item__remove-button"
              on:click={() => removeItem(index)}
            >
              Remove
            </button>
          </div>

          <div class="array-item__body">
            <SchemaForm
              fields={field?.item?.fields ?? []}
              value={item}
              showImport={false}
              on:change={(e) => updateItem(index, e.detail.value)}
            />
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .array-field {
  border: 1px solid #d9e0e7;
  border-radius: 10px;
  background: #ffffff;
  overflow: hidden;
  margin: 0;
  margin-top: 16px;
  margin-bottom: 16px;
}

.array-field__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid #e5ebf2;
  background: #fbfcfd;
}

.array-field__header-left {
  min-width: 0;
}

.array-field__title {
  font-size: 14px;
  font-weight: 600;
  color: #475467;
  line-height: 1.2;
}

.array-field__count {
  margin-top: 6px;
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  background: #f2f4f7;
  color: #667085;
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
}

.array-field__add-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid #d0d5dd;
  background: #ffffff;
  color: #475467;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  white-space: nowrap;
}

.array-field__add-button:hover {
  background: #f9fafb;
  border-color: #c7ced8;
  color: #344054;
}

.array-field__add-icon {
  font-size: 14px;
  line-height: 1;
  font-weight: 700;
}

.array-field__empty {
  margin: 14px 16px 16px;
  padding: 16px;
  border: 1px dashed #d6dde6;
  border-radius: 10px;
  background: #fbfcfd;
  text-align: center;
}

.array-field__empty-title {
  font-size: 13px;
  font-weight: 600;
  color: #475467;
}

.array-field__empty-text {
  margin-top: 4px;
  font-size: 12px;
  color: #667085;
}

.array-field__items {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.array-item {
  border: 1px solid #e1e7ee;
  border-radius: 10px;
  overflow: hidden;
  background: #ffffff;
}

.array-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid #e8edf3;
  background: #fafbfc;
}

.array-item__title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.array-item__index {
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: #f2f4f7;
  color: #475467;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.array-item__title {
  font-size: 13px;
  font-weight: 600;
  color: #475467;
  line-height: 1.2;
}

.array-item__remove-button {
  border: 1px solid #f2d4d7;
  background: #fffafa;
  color: #c2414c;
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  white-space: nowrap;
}

.array-item__remove-button:hover {
  background: #fff3f4;
  border-color: #e8bcc2;
  color: #b42318;
}

.array-item__body {
  padding: 12px;
  background: #ffffff;
}

</style>