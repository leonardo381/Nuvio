<!--
<script>
  import ApiClient from "@/utils/ApiClient";

  export let field;        // { key, label, ... }
  export let value = null; // current props value (if any)
  export let onChange;     // callback (newValue) or you can use dispatch

  let isUploading = false;
  let error = "";

  async function handleFileChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    error = "";

    try {
      const formData = new FormData();
      formData.append("file", file);

      // upload to the assets collection
      const rec = await ApiClient.collection("Assets").create(formData);

      const newVal = {
        collection: "Assets",
        recordId: rec.id,
        filename: rec.file, // field name in assets
      };

      onChange && onChange(newVal);
    } catch (err) {
      console.error("Upload failed", err);
      error = "Upload failed.";
    } finally {
      isUploading = false;
    }
  }
</script>

<div class="form-field">
  <label>{field.label}</label>

  {#if value}
    <div class="m-b-xs">
      <small>Current file:</small><br />
      <code>{value.filename}</code>
    </div>
  {/if}

  <input type="file" on:change={handleFileChange} disabled={isUploading} />

  {#if error}
    <div class="txt-danger">{error}</div>
  {/if}
</div>
-->
<script>
  import ApiClient from "@/utils/ApiClient";
  import Field from "@/components/base/Field.svelte";
  import FieldLabel from "@/components/records/fields/FieldLabel.svelte";

  export let field;        // { name / key, label, ... }
  export let value = null; // current props value
  export let onChange;     // callback(newValue)

  let isUploading = false;
  let error = "";

  async function handleFileChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;

    isUploading = true;
    error = "";

    try {
      const formData = new FormData();
      formData.append("file", file);

      const rec = await ApiClient.collection("Assets").create(formData);

      const newVal = {
        collection: "Assets",
        recordId: rec.id,
        filename: rec.file, // change if your field name is different
      };

      onChange && onChange(newVal);
    } catch (err) {
      console.error("Upload failed", err);
      error = "Upload failed.";
    } finally {
      isUploading = false;
    }
  }
</script>