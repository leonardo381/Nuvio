<script>
  import { createEventDispatcher, onMount } from "svelte";
  import TinyMCE from "@/components/base/TinyMCE.svelte";
  import CommonHelper from "@/utils/CommonHelper";
  import FieldShell from "./FieldShell.svelte";

  export let field;
  export let value = "";
  export let disabled = false;
  export let error = "";
  export let path = "";

  const dispatch = createEventDispatcher();
  $: id = `schema-${String(path || field?.key || "field").replace(/[^a-zA-Z0-9_-]/g, "-")}`;

  let mounted = false;
  let mountTimer = null;
  let editorValue = normalizeValue(value);
  let lastIncomingValue = editorValue;
  let lastDispatchedValue = editorValue;

  const richTextProfileSet = new Set(["basicRichText"]);
  const richEditorModelEvents = "change input undo redo setcontent execcommand keyup blur";
  const trustedMarkupProfiles = ["trustedIconSvg", "trustedSvgIllustration", "trustedHtmlIllustration"];
  const trustedMarkupProfileSet = new Set(trustedMarkupProfiles);
  const trustedMarkupProfileLookup = new Map(trustedMarkupProfiles.map((profile) => [profile.toLowerCase(), profile]));
  const trustedSvgElements = [
    "svg", "g", "path", "rect", "circle", "ellipse", "line", "polyline", "polygon",
    "defs", "linearGradient", "radialGradient", "stop", "clipPath", "mask", "title", "desc",
  ];
  const trustedHtmlElements = ["div", "span", "ul", "ol", "li", "figure", "figcaption", "small", "strong", "em"];
  const trustedMarkupValidElements = [
    ...trustedSvgElements.map((tagName) => `${tagName}[*]`),
    ...trustedHtmlElements.map((tagName) => `${tagName}[*]`),
  ].join(",");
  const basicRichTextAllowedPasteNodes = ["P", "A", "EM", "I", "B", "STRONG", "BR", "OL", "UL", "LI"];

  $: trustedMarkupProfile = resolveTrustedMarkupProfile(field);
  $: isTrustedMarkupField = field?.trustedMarkup === true || !!trustedMarkupProfile;
  $: richTextProfile = resolveRichTextProfile(field);
  $: isRichTextField = !!richTextProfile;
  $: shouldUseRichEditor = isTrustedMarkupField || isRichTextField;
  $: plainTextRows = Math.max(3, Number(field?.options?.rows) || 5);
  $: editorHeight = field?.options?.rows ? Math.max(180, Number(field.options.rows) * 36) : 270;
  $: baseEditorConfig = {
    ...CommonHelper.defaultEditorOptions(),
    convert_urls: false,
    relative_urls: false,
    min_height: editorHeight,
    height: editorHeight,
  };
  $: editorConfig = isTrustedMarkupField
    ? createTrustedMarkupEditorConfig(baseEditorConfig)
    : isRichTextField
      ? createBasicRichTextEditorConfig(baseEditorConfig)
      : baseEditorConfig;

  $: {
    const nextIncomingValue = normalizeValue(value);
    if (nextIncomingValue !== lastIncomingValue) {
      lastIncomingValue = nextIncomingValue;
      if (nextIncomingValue !== editorValue) {
        editorValue = nextIncomingValue;
      }
      if (nextIncomingValue !== lastDispatchedValue) {
        lastDispatchedValue = nextIncomingValue;
      }
    }
  }

  $: if (mounted && editorValue !== lastDispatchedValue) {
    lastDispatchedValue = editorValue;
    dispatch("change", editorValue);
  }

  function normalizeValue(raw) {
    if (typeof raw === "string") return raw;
    if (raw === null || typeof raw === "undefined") return "";
    return String(raw);
  }

  function resolveRichTextProfile(fieldConfig) {
    if (fieldConfig?.richText !== true) return "";

    const profile = `${fieldConfig.richTextProfile || "basicRichText"}`.trim();
    return richTextProfileSet.has(profile) ? profile : "";
  }

  function resolveTrustedMarkupProfile(fieldConfig) {
    if (!fieldConfig) return "";

    const candidates = [fieldConfig.profile, fieldConfig.trustedMarkupProfile, fieldConfig.richTextProfile, fieldConfig.key, fieldConfig.name]
      .map((candidate) => `${candidate || ""}`.trim())
      .filter(Boolean);

    for (const candidate of candidates) {
      if (trustedMarkupProfileSet.has(candidate)) {
        return candidate;
      }
    }

    for (const candidate of candidates) {
      const normalizedCandidate = candidate.toLowerCase();
      for (const [normalizedProfile, profile] of trustedMarkupProfileLookup) {
        if (normalizedCandidate.includes(normalizedProfile)) {
          return profile;
        }
      }
    }

    return "";
  }

  function createBasicRichTextEditorConfig(baseConfig) {
    const { file_picker_callback, file_picker_types, paste_postprocess, ...richConfig } = baseConfig;

    return {
      ...richConfig,
      plugins: ["autoresize", "autolink", "lists", "link"],
      toolbar: "bold italic | bullist numlist | link removeformat",
      block_formats: "Paragraph=p",
      forced_root_block: "p",
      valid_elements: "p,br,strong/b,em/i,ul,ol,li,a[href|title|target|rel]",
      extended_valid_elements: "a[href|title|target|rel]",
      invalid_elements: "script,style,iframe,object,embed,img,table,thead,tbody,tfoot,tr,td,th,video,audio,source,form,input,button,textarea,select,option",
      paste_postprocess: (_editor, args) => {
        cleanupBasicRichTextNode(args.node);
      },
      content_style: `${baseConfig.content_style || ""}\na { color: #0f6f8f; }`,
    };
  }

  function createTrustedMarkupEditorConfig(baseConfig) {
    const { file_picker_callback, file_picker_types, paste_postprocess, ...trustedConfig } = baseConfig;

    return {
      ...trustedConfig,
      plugins: ["autoresize", "code"],
      toolbar: "undo redo | code",
      forced_root_block: false,
      valid_elements: trustedMarkupValidElements,
      extended_valid_elements: trustedMarkupValidElements,
      entity_encoding: "raw",
      content_style: `${baseConfig.content_style || ""}\nbody { font-family: sans-serif; } svg { display: block; max-width: 100%; height: auto; }`,
    };
  }

  function cleanupBasicRichTextNode(node) {
    if (!node) return;

    for (const child of Array.from(node.children || [])) {
      cleanupBasicRichTextNode(child);
    }

    if (!basicRichTextAllowedPasteNodes.includes(node.tagName)) {
      unwrapNode(node);
      return;
    }

    for (const attribute of Array.from(node.attributes || [])) {
      if (node.tagName === "A" && ["href", "title", "target", "rel"].includes(attribute.name)) continue;
      node.removeAttribute(attribute.name);
    }
  }

  function unwrapNode(node) {
    const parent = node.parentNode;
    if (!parent) return;

    while (node.firstChild) {
      parent.insertBefore(node.firstChild, node);
    }

    parent.removeChild(node);
  }

  onMount(() => {
    // Slight delay avoids heavy initial paint when many fields mount together.
    mountTimer = setTimeout(() => {
      mounted = true;
    }, 60);

    return () => {
      if (mountTimer) clearTimeout(mountTimer);
    };
  });
</script>

<FieldShell {field} {id} {error} required={!!field?.required} hint={field?.hint || ""}>
  {#if shouldUseRichEditor}
    {#if mounted}
      <TinyMCE id={id} conf={editorConfig} modelEvents={richEditorModelEvents} bind:value={editorValue} {disabled} />
    {:else}
      <div class="textarea-editor-skeleton" aria-hidden="true"></div>
    {/if}
  {:else}
    <textarea
      {id}
      class="schema-plain-textarea"
      rows={plainTextRows}
      bind:value={editorValue}
      {disabled}
      aria-invalid={error ? "true" : undefined}
    ></textarea>
  {/if}
</FieldShell>
<style>
  .textarea-editor-skeleton,
  .schema-plain-textarea {
    width: 100%;
    border: 1px solid rgba(15, 23, 42, 0.12);
    border-radius: 10px;
    background: color-mix(in srgb, var(--baseAlt1Color) 70%, #fff);
  }

  .textarea-editor-skeleton {
    min-height: 220px;
  }

  .schema-plain-textarea {
    min-height: 132px;
    padding: 12px 14px;
    color: var(--txtPrimaryColor);
    font: inherit;
    line-height: 1.5;
    resize: vertical;
    transition: border-color 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
  }

  .schema-plain-textarea:focus {
    border-color: color-mix(in srgb, var(--primaryColor) 60%, #0ea5e9);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--primaryColor) 16%, transparent);
    outline: none;
  }

  .schema-plain-textarea:disabled {
    cursor: not-allowed;
    opacity: 0.72;
  }
</style>
