<!--
<script>
  export let field;
  export let id;
  export let required = false;
  export let error = "";
  export let hint = "";
</script>

<div class="form-field {required ? 'required' : ''} {error ? 'has-error' : ''}">
  {#if field?.label}
    <label for={id} class="form-label">
      {field.label}
      {#if required}<span class="req">*</span>{/if}
    </label>
  {/if}

  <div class="form-control">
    <slot />
  </div>

  {#if hint && !error}
    <div class="form-hint">{hint}</div>
  {/if}

  {#if error}
    <div class="form-error">{error}</div>
  {/if}
</div>

<style>
  .form-field { margin: 12px 0; }
  .form-label { display:block; font-weight:600; margin-bottom:6px; }
  .req { margin-left:6px; opacity:0.7; }
  .form-hint { margin-top:6px; font-size:12px; opacity:0.75; }
  .form-error { margin-top:6px; font-size:12px; color: #d33; }
</style>
-->
<script>
  export let field;
  export let id;
  export let required = false;
  export let error = "";
  export let hint = "";
</script>

<div class="pb-field {required ? 'required' : ''} {error ? 'has-error' : ''}">
  {#if field?.label}
    <label for={id} class="pb-label">
      {field.label}
      {#if required}<span class="pb-req">*</span>{/if}
    </label>
  {/if}

  <div class="pb-control">
    <slot />
  </div>

  {#if hint && !error}
    <div class="pb-hint">{hint}</div>
  {/if}

  {#if error}
    <div class="pb-error">{error}</div>
  {/if}
</div>

<style>
  /* Field wrapper (PocketBase-like block) */
  .pb-field{
    margin: 12px 0;
    padding: 10px 12px 12px;
    border-radius: 10px;
    background: #eef2f7;
    border: 1px solid rgba(15, 23, 42, 0.06);
  }

  .pb-label{
    display: block;
    font-size: 12px;
    font-weight: 600;
    color: rgba(15, 23, 42, 0.65);
    margin-bottom: 8px;
  }

  .pb-req{
    margin-left: 6px;
    opacity: 0.7;
  }

  .pb-control{ }

  .pb-hint{
    margin-top: 8px;
    font-size: 12px;
    color: rgba(15, 23, 42, 0.55);
  }

  .pb-error{
    margin-top: 8px;
    font-size: 12px;
    color: #d33;
  }

  /* ✅ IMPORTANT: Make ALL your inputs match PB (shared classes) */
  :global(.form-input),
  :global(.form-textarea),
  :global(.form-select){
    width: 100%;
    box-sizing: border-box;

    background: #f8fafc;
    border: 1px solid rgba(15, 23, 42, 0.12);
    border-radius: 10px;

    padding: 10px 12px;
    font-size: 14px;
    color: rgba(15, 23, 42, 0.9);

    outline: none;
  }

  :global(.form-textarea){
    resize: vertical;
    min-height: 96px;
  }

  :global(.form-input:focus),
  :global(.form-textarea:focus),
  :global(.form-select:focus){
    border-color: rgba(15, 23, 42, 0.28);
    box-shadow: 0 0 0 3px rgba(15, 23, 42, 0.08);
  }

  :global(.form-input:disabled),
  :global(.form-textarea:disabled),
  :global(.form-select:disabled){
    opacity: 0.65;
    cursor: not-allowed;
  }

  /* Optional: PB-like “invalid” */
  .has-error{
    border-color: rgba(220, 38, 38, 0.35);
  }
</style>