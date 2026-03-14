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

<div class="space-y-4">
  <div class="font-medium">{field.label}</div>

  {#each items as item, index}
    <div class="rounded border p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="text-sm font-medium">
          Item {index + 1}
        </div>

        <button
          type="button"
          class="text-sm text-red-600"
          on:click={() => removeItem(index)}
        >
          Remove
        </button>
      </div>

      <SchemaForm
        fields={field?.item?.fields ?? []}
        value={item}
        on:change={(e) => updateItem(index, e.detail.value)}
      />
    </div>
  {/each}

  <button
    type="button"
    class="rounded border px-3 py-2 text-sm"
    on:click={addItem}
  >
    Add item
  </button>
</div>