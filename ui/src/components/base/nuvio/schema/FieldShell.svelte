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
  .pb-field {
    min-width: 0;
    margin: 0 0 var(--baseSpacing);
  }

  .pb-label {
    display: inline-flex;
    align-items: center;
    flex-wrap: wrap;
    margin: 0 0 6px;
    color: var(--txtHintColor);
    font-size: var(--smFontSize);
    font-weight: 600;
    line-height: 1.25;
    letter-spacing: 0.1px;
  }

  .pb-req {
    margin-left: 4px;
    color: var(--dangerColor);
    font-size: 0.95em;
    line-height: 1;
  }

  .pb-control {
    min-width: 0;
  }

  .pb-hint,
  .pb-error {
    margin-top: 6px;
    font-size: var(--smFontSize);
    line-height: var(--smLineHeight);
    word-break: break-word;
  }

  .pb-hint {
    color: var(--txtHintColor);
  }

  .pb-error {
    color: var(--dangerColor);
  }

  .pb-field.has-error .pb-label {
    color: var(--dangerColor);
  }

  .pb-field.has-error :global(input:not([type="checkbox"]):not([type="radio"])),
  .pb-field.has-error :global(select),
  .pb-field.has-error :global(textarea) {
    border-color: color-mix(in srgb, var(--dangerColor) 55%, var(--baseAlt2Color));
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--dangerColor) 22%, transparent);
  }
</style>
