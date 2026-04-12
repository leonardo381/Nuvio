<script>
    import { createEventDispatcher } from "svelte";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import RecordFileThumb from "@/components/records/RecordFileThumb.svelte";

    const dispatch = createEventDispatcher();
    const preferredTitleKeys = ["label", "title", "name", "headline", "heading"];

    export let page = null;
    export let blocksCollection = null;

    let blocks = [];
    let isLoading = false;
    let lastLoadKey = "";

    $: pageId = page?.id || "";

    $: pageDisplayName = `${CommonHelper.displayValue(page || {}, ["title"]) || ""}`.trim() || pageId;

    $: loadKey = `${blocksCollection?.id || ""}:${pageId}`;

    $: if (!pageId) {
        blocks = [];
        lastLoadKey = "";
    } else if (blocksCollection?.id && loadKey !== lastLoadKey) {
        lastLoadKey = loadKey;
        load();
    }

    export function reload() {
        if (!pageId || !blocksCollection?.id) {
            return;
        }

        lastLoadKey = ""; // force reload on same page
        return load();
    }

    async function load() {
        if (!pageId || !blocksCollection?.id) {
            blocks = [];
            return;
        }

        isLoading = true;

        try {
            blocks = await ApiClient.collection(blocksCollection.id).getFullList({
                filter: `page="${pageId}"`,
                sort: "+created",
                requestKey: "client_page_blocks_cards_" + pageId,
            });
        } catch (err) {
            if (!err?.isAbort) {
                blocks = [];
                ApiClient.error(err);
            }
        }

        isLoading = false;
    }

    function getFriendlyComponentKey(componentKey) {
        const key = (componentKey || "").trim();
        if (!key) {
            return "Unknown component";
        }

        return key
            .replace(/[._/\\-]+/g, " ")
            .replace(/\s+/g, " ")
            .trim()
            .replace(/\b\w/g, (c) => c.toUpperCase());
    }

    function toPropsObject(rawProps) {
        if (!rawProps) {
            return {};
        }

        if (typeof rawProps === "object") {
            return rawProps;
        }

        if (typeof rawProps === "string") {
            try {
                const parsed = JSON.parse(rawProps);
                return parsed && typeof parsed === "object" ? parsed : {};
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function getDisplayName(block, i) {
        const explicitTitle = `${block?.title || ""}`.trim();
        if (explicitTitle) {
            return CommonHelper.truncate(explicitTitle, 60, true);
        }

        const props = toPropsObject(block?.props);

        for (const key of preferredTitleKeys) {
            const value = props?.[key];
            if (typeof value === "string" && value.trim()) {
                return CommonHelper.truncate(value.trim(), 60, true);
            }
        }

        if (typeof block?.component_key === "string" && block.component_key.trim()) {
            return getFriendlyComponentKey(block.component_key);
        }

        return "Block " + (i + 1);
    }

    function getCardImageFilename(block) {
        const files = CommonHelper.toArray(block?.image).filter(Boolean);
        return files[0] || "";
    }

    function formatSummaryValue(value) {
        if (typeof value === "string") {
            return CommonHelper.truncate(value.replace(/\s+/g, " ").trim(), 48, true);
        }

        if (typeof value === "number") {
            return value.toString();
        }

        if (typeof value === "boolean") {
            return value ? "Yes" : "No";
        }

        return "";
    }

    function getPropsSummary(block) {
        const props = toPropsObject(block?.props);
        const summaryParts = [];

        for (const [key, value] of Object.entries(props || {})) {
            if (summaryParts.length >= 2) {
                break;
            }

            if (value === null || typeof value === "undefined") {
                continue;
            }

            if (Array.isArray(value)) {
                if (value.length) {
                    summaryParts.push(`${key}: ${value.length} items`);
                }
                continue;
            }

            if (typeof value === "object") {
                const count = Object.keys(value).length;
                if (count) {
                    summaryParts.push(`${key}: ${count} fields`);
                }
                continue;
            }

            const formatted = formatSummaryValue(value);
            if (formatted) {
                summaryParts.push(`${key}: ${formatted}`);
            }
        }

        if (summaryParts.length) {
            return summaryParts.join(" | ");
        }

        const totalKeys = Object.keys(props || {}).length;
        if (totalKeys > 0) {
            return `Configured props: ${totalKeys}`;
        }

        return "No props summary yet.";
    }
</script>

<section class="client-page-blocks">
    <div class="section-header">
        <h5 class="m-0">Page Blocks</h5>
        {#if pageId}
            <span class="txt-sm txt-hint">
                {pageDisplayName}
                {#if blocks.length}
                    ({blocks.length})
                {/if}
            </span>
        {/if}
    </div>

    {#if !pageId}
        <div class="empty-state">Select a page to view and edit its blocks.</div>
    {:else if isLoading}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading blocks...</span>
        </div>
    {:else if !blocks.length}
        <div class="empty-state">This page has no blocks yet.</div>
    {:else}
        <div class="cards-grid">
            {#each blocks as block, i (block.id)}
                {@const cardImageFilename = getCardImageFilename(block)}
                <article class="block-card">
                    {#if cardImageFilename}
                        <div class="card-image">
                            <RecordFileThumb record={block} filename={cardImageFilename} size="xl" />
                        </div>
                    {/if}

                    <div class="card-header">
                        <h6 class="m-0">{getDisplayName(block, i)}</h6>
                        <span class="status {block.enabled ? 'status-enabled' : 'status-disabled'}">
                            {block.enabled ? "Enabled" : "Disabled"}
                        </span>
                    </div>

                    <div class="meta txt-sm txt-hint">
                        {block.component_key || "No component_key"}
                    </div>

                    <div class="summary txt-sm">{getPropsSummary(block)}</div>

                    <div class="actions">
                        <button
                            type="button"
                            class="btn btn-sm btn-secondary"
                            on:click={() => dispatch("edit", block)}
                        >
                            <i class="ri-pencil-line" />
                            <span class="txt">Edit block</span>
                        </button>
                    </div>
                </article>
            {/each}
        </div>
    {/if}
</section>

<style>
    .client-page-blocks {
        margin-top: 14px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 12px;
    }

    .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        margin-bottom: 10px;
    }

    .loading-state,
    .empty-state {
        min-height: 72px;
        border: 1px dashed var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 12px;
        color: var(--txtHintColor);
    }

    .cards-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
        gap: 10px;
    }

    .block-card {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 10px;
        background: var(--baseColor);
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .card-header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
    }

    .card-image {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        overflow: hidden;
        min-height: 88px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--baseAlt1Color);
    }

    :global(.card-image .thumb) {
        width: 100%;
        height: 120px;
        border-radius: 0;
        margin: 0;
    }

    :global(.card-image .thumb img) {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }

    .status {
        font-size: var(--xsFontSize);
        padding: 2px 7px;
        border-radius: 999px;
        border: 1px solid var(--baseAlt2Color);
        white-space: nowrap;
    }

    .status-enabled {
        color: var(--successColor);
    }

    .status-disabled {
        color: var(--txtHintColor);
    }

    .summary {
        min-height: 36px;
        color: var(--txtHintColor);
    }

    .actions {
        margin-top: auto;
        display: flex;
        justify-content: flex-end;
    }
</style>
