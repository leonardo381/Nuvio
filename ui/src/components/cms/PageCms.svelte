<script>
    import { querystring } from "svelte-spa-router";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import SchemaForm from "@/components/base/nuvio/schema/SchemaForm.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addSuccessToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";

    $pageTitle = "Conteúdo do Website";
    loadCollections();

    const initialQueryParams = new URLSearchParams($querystring);

    let websites = [];
    let pages = [];
    let blocks = [];
    let components = [];

    let selectedWebsiteId = initialQueryParams.get("cmsWebsite") || "";
    let selectedPageId = initialQueryParams.get("cmsPage") || "";
    let openSectionId = "";

    let isLoadingWebsites = false;
    let isLoadingPages = false;
    let isLoadingBlocks = false;
    let isLoadingComponents = false;
    let isSavingPage = false;

    let pageError = "";
    let pageEditForm = {
        title: "",
        enabled: true,
    };

    let sectionPropsDraftById = {};
    let sectionErrorById = {};
    let isSavingSectionById = {};

    let lastCollectionsKey = "";
    let lastPersistedContextKey = "";
    let lastPageSeedId = "";
    let lastSectionsSeedKey = "";

    $: websitesCollection = findCollection("websites");
    $: pagesCollection = findCollection("pages");
    $: blocksCollection = findCollection("blocks");
    $: componentsCollection = findCollection("components");

    $: hasCmsCollections = !!websitesCollection?.id && !!pagesCollection?.id && !!blocksCollection?.id;

    $: missingCollections = [];
    $: if (!websitesCollection?.id) {
        missingCollections.push("websites");
    }
    $: if (!pagesCollection?.id) {
        missingCollections.push("pages");
    }
    $: if (!blocksCollection?.id) {
        missingCollections.push("blocks");
    }

    $: websiteNameField = resolveFieldName(websitesCollection, ["name", "title", "label"]);
    $: websiteSlugField = resolveFieldName(websitesCollection, ["slug"]);
    $: websiteDomainField = resolveFieldName(websitesCollection, ["domain", "url", "host"]);

    $: pageTitleField = resolveFieldName(pagesCollection, ["title", "name", "label"]);
    $: pageSlugField = resolveFieldName(pagesCollection, ["slug"]);
    $: pageEnabledField = resolveFieldName(pagesCollection, ["enabled", "published", "active"]);

    $: blockTitleField = resolveFieldName(blocksCollection, ["title", "name", "label"]);
    $: blockPropsField = resolveFieldName(blocksCollection, ["props"]);
    $: blockComponentKeyField = resolveFieldName(blocksCollection, ["component_key", "componentKey"]);
    $: blockComponentRelationField = resolveRelationFieldName(blocksCollection, componentsCollection?.id, ["component"]);
    $: blockPageRelationField = resolveRelationFieldName(blocksCollection, pagesCollection?.id, ["page"]);

    $: componentKeyField = resolveFieldName(componentsCollection, ["key", "component_key"]);
    $: componentNameField = resolveFieldName(componentsCollection, ["name", "title", "label"]);
    $: componentSchemaField = resolveFieldName(componentsCollection, ["schema"]);

    $: pageWebsiteRelationField = resolveRelationFieldName(pagesCollection, websitesCollection?.id, ["website"]);

    $: selectedWebsite = websites.find((record) => record.id === selectedWebsiteId) || null;
    $: selectedPage = pages.find((record) => record.id === selectedPageId) || null;

    $: websitePublicUrl = getWebsitePublicUrl(selectedWebsite);
    $: selectedPagePath = getPagePath(selectedPage);

    $: componentsById = new Map(components.map((record) => [record.id, record]));
    $: componentsByKey = new Map(
        components
            .map((record) => {
                const key = normalizeString(componentKeyField ? record?.[componentKeyField] : "");
                return [key, record];
            })
            .filter(([key]) => !!key),
    );

    $: if (selectedPage?.id && selectedPage.id !== lastPageSeedId) {
        lastPageSeedId = selectedPage.id;
        pageEditForm = {
            title: pageTitleField ? `${selectedPage?.[pageTitleField] || ""}` : "",
            enabled: pageEnabledField ? !!selectedPage?.[pageEnabledField] : true,
        };
        pageError = "";
    }

    $: if (!selectedPage?.id) {
        lastPageSeedId = "";
    }

    $: if (!hasCmsCollections) {
        websites = [];
        pages = [];
        blocks = [];
        components = [];
        selectedWebsiteId = "";
        selectedPageId = "";
        openSectionId = "";
        lastCollectionsKey = "";
    } else {
        const nextKey = `${websitesCollection?.id || ""}|${pagesCollection?.id || ""}|${blocksCollection?.id || ""}|${componentsCollection?.id || ""}`;
        if (nextKey !== lastCollectionsKey) {
            lastCollectionsKey = nextKey;
            reload();
        }
    }

    $: if (hasCmsCollections) {
        const nextContextKey = `${selectedWebsiteId || ""}|${selectedPageId || ""}`;
        if (nextContextKey !== lastPersistedContextKey) {
            lastPersistedContextKey = nextContextKey;
            CommonHelper.replaceHashQueryParams({
                cmsWebsite: selectedWebsiteId || null,
                cmsPage: selectedPageId || null,
            });
        }
    }

    $: {
        const nextSeedKey = `${blockPropsField || ""}|${blocks.map((block) => `${block.id}:${block?.updated || ""}`).join("|")}`;
        if (nextSeedKey !== lastSectionsSeedKey) {
            lastSectionsSeedKey = nextSeedKey;

            const nextDraft = {};
            const nextErrors = {};
            const nextSaving = {};

            for (const block of blocks) {
                nextDraft[block.id] = toPropsObject(blockPropsField ? block?.[blockPropsField] : {});
                nextErrors[block.id] = "";
                nextSaving[block.id] = false;
            }

            sectionPropsDraftById = nextDraft;
            sectionErrorById = nextErrors;
            isSavingSectionById = nextSaving;

            if (!openSectionId || !blocks.some((block) => block.id === openSectionId)) {
                openSectionId = blocks[0]?.id || "";
            }
        }
    }

    function findCollection(name) {
        return $collections.find((item) => `${item?.name || ""}`.toLowerCase() === `${name || ""}`.toLowerCase()) || null;
    }

    function getCollectionFieldNames(collection) {
        return new Set((collection?.fields || []).map((field) => `${field?.name || ""}`.toLowerCase()));
    }

    function resolveFieldName(collection, candidates = []) {
        const fieldNames = getCollectionFieldNames(collection);
        for (const candidate of candidates) {
            if (fieldNames.has(`${candidate || ""}`.toLowerCase())) {
                return candidate;
            }
        }
        return "";
    }

    function resolveRelationFieldName(collection, targetCollectionId = "", fallbackCandidates = []) {
        const fields = collection?.fields || [];
        const relationByTarget = fields.find(
            (field) => field?.type === "relation" && `${field.collectionId || ""}` === `${targetCollectionId || ""}`,
        );

        if (relationByTarget?.name) {
            return relationByTarget.name;
        }

        const normalizedFallback = fallbackCandidates.map((item) => `${item || ""}`.toLowerCase());
        const relationByName = fields.find(
            (field) => field?.type === "relation" && normalizedFallback.includes(`${field?.name || ""}`.toLowerCase()),
        );

        return relationByName?.name || "";
    }

    function resolveSort(collection, candidates = []) {
        const fieldNames = getCollectionFieldNames(collection);
        const valid = candidates.filter((candidate) => fieldNames.has(`${candidate || ""}`.toLowerCase()));
        if (!valid.length) {
            return "+id";
        }
        return valid.map((field) => `+${field}`).join(",");
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function toPropsObject(value) {
        if (value && typeof value === "object" && !Array.isArray(value)) {
            return structuredClone(value);
        }

        if (typeof value === "string" && value.trim()) {
            try {
                const parsed = JSON.parse(value);
                if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
                    return parsed;
                }
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function parseSchemaFields(rawSchema) {
        let schema = rawSchema;
        if (typeof schema === "string") {
            try {
                schema = JSON.parse(schema);
            } catch (_) {
                return [];
            }
        }

        if (!schema || typeof schema !== "object") {
            return [];
        }

        return Array.isArray(schema.fields) ? schema.fields : [];
    }

    function setPayloadField(payload, fieldName, value) {
        if (!fieldName) {
            return;
        }
        payload[fieldName] = value;
    }

    function getWebsiteLabel(record) {
        return `${CommonHelper.displayValue(record || {}, [websiteNameField, websiteSlugField, websiteDomainField], "") || ""}`.trim() || record?.id || "-";
    }

    function getPageLabel(record) {
        return `${CommonHelper.displayValue(record || {}, [pageTitleField, pageSlugField], "") || ""}`.trim() || record?.id || "-";
    }

    function getPagePath(record) {
        const slug = normalizeString(pageSlugField ? record?.[pageSlugField] : "");
        if (!slug) {
            return "/";
        }
        return slug.startsWith("/") ? slug : `/${slug}`;
    }

    function getWebsitePublicUrl(record) {
        const domain = normalizeString(websiteDomainField ? record?.[websiteDomainField] : "");
        if (!domain) {
            return "";
        }
        if (/^https?:\/\//i.test(domain)) {
            return domain;
        }
        return `https://${domain}`;
    }

    function getExpandedComponent(block) {
        if (!blockComponentRelationField) {
            return null;
        }

        const expanded = block?.expand?.[blockComponentRelationField];
        if (Array.isArray(expanded)) {
            return expanded[0] || null;
        }
        return expanded || null;
    }

    function getBlockComponentId(block) {
        if (!blockComponentRelationField) {
            return "";
        }

        const relationValue = block?.[blockComponentRelationField];
        if (Array.isArray(relationValue)) {
            return normalizeString(relationValue[0]);
        }
        if (typeof relationValue === "string") {
            return normalizeString(relationValue);
        }

        const expanded = getExpandedComponent(block);
        return normalizeString(expanded?.id);
    }

    function getBlockComponentKey(block) {
        if (blockComponentKeyField) {
            const rawKey = normalizeString(block?.[blockComponentKeyField]);
            if (rawKey) {
                return rawKey;
            }
        }

        const expanded = getExpandedComponent(block);
        if (expanded && componentKeyField) {
            const expandedKey = normalizeString(expanded?.[componentKeyField]);
            if (expandedKey) {
                return expandedKey;
            }
        }

        const componentId = getBlockComponentId(block);
        if (componentId && componentsById.has(componentId) && componentKeyField) {
            return normalizeString(componentsById.get(componentId)?.[componentKeyField]);
        }

        return "";
    }

    function getComponentForBlock(block) {
        const key = getBlockComponentKey(block);
        if (key && componentsByKey.has(key)) {
            return componentsByKey.get(key);
        }

        const componentId = getBlockComponentId(block);
        if (componentId && componentsById.has(componentId)) {
            return componentsById.get(componentId);
        }

        return null;
    }

    function getSectionSchemaFields(block) {
        const component = getComponentForBlock(block);
        if (!component || !componentSchemaField) {
            return [];
        }
        return parseSchemaFields(component?.[componentSchemaField]);
    }

    function getSectionTitle(block, index) {
        const title = normalizeString(blockTitleField ? block?.[blockTitleField] : "");
        if (title) {
            return title;
        }

        const component = getComponentForBlock(block);
        const componentName = normalizeString(componentNameField ? component?.[componentNameField] : "");
        if (componentName) {
            return componentName;
        }

        return `Section ${index + 1}`;
    }

    function getSectionDescription(block) {
        const component = getComponentForBlock(block);
        const componentName = normalizeString(componentNameField ? component?.[componentNameField] : "");
        if (componentName) {
            return componentName;
        }
        return "Edit section content.";
    }

    function toggleSection(blockId) {
        if (openSectionId === blockId) {
            openSectionId = "";
            return;
        }
        openSectionId = blockId;
    }

    function updateSectionDraft(blockId, nextValue) {
        sectionPropsDraftById = {
            ...sectionPropsDraftById,
            [blockId]: toPropsObject(nextValue),
        };
    }

    async function reload() {
        if (!hasCmsCollections) {
            return;
        }

        await loadWebsites();
        await loadComponents();
        await loadPages();
        await loadBlocks();
    }

    async function loadWebsites() {
        if (!websitesCollection?.id) {
            websites = [];
            selectedWebsiteId = "";
            return;
        }

        isLoadingWebsites = true;

        try {
            websites = await ApiClient.collection(websitesCollection.id).getFullList({
                sort: resolveSort(websitesCollection, ["name", "title", "slug", "domain"]),
                requestKey: "nuvio_cms_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                selectedPageId = "";
            } else if (!websites.some((record) => record.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
                selectedPageId = "";
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            selectedPageId = "";
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadComponents() {
        if (!componentsCollection?.id) {
            components = [];
            return;
        }

        isLoadingComponents = true;

        try {
            components = await ApiClient.collection(componentsCollection.id).getFullList({
                sort: resolveSort(componentsCollection, ["name", "title", "key"]),
                requestKey: "nuvio_cms_components",
            });
        } catch (err) {
            components = [];
            ApiClient.error(err);
        }

        isLoadingComponents = false;
    }

    async function loadPages() {
        if (!pagesCollection?.id || !selectedWebsiteId || !pageWebsiteRelationField) {
            pages = [];
            selectedPageId = "";
            return;
        }

        isLoadingPages = true;

        try {
            pages = await ApiClient.collection(pagesCollection.id).getFullList({
                filter: `${pageWebsiteRelationField}="${selectedWebsiteId}"`,
                sort: resolveSort(pagesCollection, ["title", "name", "slug"]),
                requestKey: "nuvio_cms_pages_" + selectedWebsiteId,
            });

            if (!pages.length) {
                selectedPageId = "";
            } else if (!pages.some((record) => record.id === selectedPageId)) {
                selectedPageId = pages[0].id;
            }
        } catch (err) {
            pages = [];
            selectedPageId = "";
            ApiClient.error(err);
        }

        isLoadingPages = false;
    }

    async function loadBlocks() {
        if (!blocksCollection?.id || !selectedPageId || !blockPageRelationField) {
            blocks = [];
            openSectionId = "";
            return;
        }

        isLoadingBlocks = true;

        try {
            const query = {
                filter: `${blockPageRelationField}="${selectedPageId}"`,
                sort: resolveSort(blocksCollection, ["title", "slot", "variant"]),
                requestKey: "nuvio_cms_blocks_" + selectedPageId,
            };

            if (blockComponentRelationField) {
                query.expand = blockComponentRelationField;
            }

            blocks = await ApiClient.collection(blocksCollection.id).getFullList(query);
            if (!openSectionId || !blocks.some((block) => block.id === openSectionId)) {
                openSectionId = blocks[0]?.id || "";
            }
        } catch (err) {
            blocks = [];
            openSectionId = "";
            ApiClient.error(err);
        }

        isLoadingBlocks = false;
    }

    async function selectWebsite(websiteId) {
        if (`${websiteId || ""}` === `${selectedWebsiteId || ""}`) {
            return;
        }

        selectedWebsiteId = `${websiteId || ""}`;
        selectedPageId = "";
        openSectionId = "";

        await loadPages();
        await loadBlocks();
    }

    async function selectPage(pageId) {
        if (`${pageId || ""}` === `${selectedPageId || ""}`) {
            return;
        }

        selectedPageId = `${pageId || ""}`;
        openSectionId = "";

        await loadBlocks();
    }

    async function savePage() {
        pageError = "";

        if (!selectedPage?.id || !pagesCollection?.id || !selectedWebsiteId || !pageWebsiteRelationField) {
            pageError = "Select a page first.";
            return;
        }

        const payload = {};
        setPayloadField(payload, pageWebsiteRelationField, selectedWebsiteId);

        if (pageTitleField) {
            setPayloadField(payload, pageTitleField, normalizeString(pageEditForm.title));
        }
        if (pageEnabledField) {
            setPayloadField(payload, pageEnabledField, !!pageEditForm.enabled);
        }

        isSavingPage = true;

        try {
            await ApiClient.collection(pagesCollection.id).update(selectedPage.id, payload);
            addSuccessToast("Page updated.");
            await loadPages();
        } catch (err) {
            ApiClient.error(err);
            pageError = err?.response?.message || err?.message || "Failed to save page.";
        }

        isSavingPage = false;
    }

    async function saveSection(block) {
        const blockId = `${block?.id || ""}`;
        if (!blockId) {
            return;
        }

        sectionErrorById = { ...sectionErrorById, [blockId]: "" };

        if (!blocksCollection?.id || !blockPropsField) {
            sectionErrorById = {
                ...sectionErrorById,
                [blockId]: "Section cannot be saved because props field is missing.",
            };
            return;
        }

        const payload = {};
        setPayloadField(payload, blockPropsField, toPropsObject(sectionPropsDraftById?.[blockId]));

        isSavingSectionById = {
            ...isSavingSectionById,
            [blockId]: true,
        };

        try {
            await ApiClient.collection(blocksCollection.id).update(blockId, payload);
            addSuccessToast("Section updated.");
            await loadBlocks();
        } catch (err) {
            ApiClient.error(err);
            sectionErrorById = {
                ...sectionErrorById,
                [blockId]: err?.response?.message || err?.message || "Failed to save section.",
            };
        }

        isSavingSectionById = {
            ...isSavingSectionById,
            [blockId]: false,
        };
    }

    async function handleWebsiteChange(event) {
        await selectWebsite(event.currentTarget.value);
    }
</script>

<PageWrapper>
    {#if $isCollectionsLoading && !hasCmsCollections}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading website content...</h1>
        </div>
    {:else if !hasCmsCollections}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>
                Missing required collections: {missingCollections.join(", ")}.<br />
                Keep using Collections until those collections are available.
            </div>
        </div>
    {:else}
        <section class="cms-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <div class="title-row">
                        <h2 class="m-0">Conteúdo do Website</h2>
                        <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={reload} />
                    </div>
                    <p class="txt-sm txt-hint m-b-0 head-description">Edite as páginas e secções do seu site.</p>
                </div>

                <div class="head-selector">
                    <div class="selector-label-row">
                        <label class="txt-sm txt-hint selector-label m-b-0" for="cms-website">Website</label>
                    </div>
                    <div class="selector-controls">
                        <select
                            id="cms-website"
                            class="input input-sm"
                            value={selectedWebsiteId}
                            disabled={isLoadingWebsites || !websites.length}
                            on:change={handleWebsiteChange}
                        >
                            <option value="">Select website</option>
                            {#each websites as website}
                                <option value={website.id}>{getWebsiteLabel(website)}</option>
                            {/each}
                        </select>

                        {#if websitePublicUrl}
                            <a href={websitePublicUrl} target="_blank" rel="noreferrer noopener" class="btn btn-sm btn-outline">
                                View website
                            </a>
                        {/if}
                    </div>
                </div>
            </div>

            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-file-list-3-line" />
                    {pages.length} pages
                </span>
                <span class="summary-pill">
                    <i class="ri-layout-grid-line" />
                    {blocks.length} sections
                </span>
                {#if !componentsCollection?.id}
                    <span class="summary-pill warning">
                        <i class="ri-alert-line" />
                        components missing
                    </span>
                {/if}
            </div>
        </section>

        <section class="panel cms-section-panel m-b-base">
            <div class="content-workspace-grid">
                <aside class="pages-list-panel">
                    <div class="pages-list-head">
                        <h4 class="m-0">Pages</h4>
                        <span class="txt-sm txt-hint">{pages.length} total</span>
                    </div>

                    {#if !selectedWebsiteId}
                        <div class="txt-hint m-t-sm">Select a website first.</div>
                    {:else if !pageWebsiteRelationField}
                        <div class="txt-hint txt-danger m-t-sm">Pages relation to websites is missing.</div>
                    {:else if isLoadingPages}
                        <div class="txt-hint m-t-sm">Loading pages...</div>
                    {:else if !pages.length}
                        <div class="txt-hint m-t-sm">No pages for this website.</div>
                    {:else}
                        <div class="pages-list-body m-t-sm">
                            {#each pages as page}
                                <button
                                    type="button"
                                    class="page-row"
                                    class:active={page.id === selectedPageId}
                                    on:click={() => selectPage(page.id)}
                                >
                                    <span class="page-row-title">{getPageLabel(page)}</span>
                                    <span class="page-row-path">{getPagePath(page)}</span>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </aside>

                <div class="page-editor-panel">
                    {#if selectedPage}
                        <div class="page-editor-head">
                            <div>
                                <h4 class="m-0">Page details</h4>
                                <p class="txt-sm txt-hint m-b-0 m-t-6">Path: {selectedPagePath}</p>
                            </div>
                        </div>

                        <div class="form-grid two-col m-t-sm">
                            {#if pageTitleField}
                                <div class="form-field">
                                    <label class="txt-sm txt-hint block m-b-5" for="cms-page-title-content">Title</label>
                                    <input id="cms-page-title-content" class="input input-sm" bind:value={pageEditForm.title} />
                                </div>
                            {/if}

                            <div class="form-field read-only-field">
                                <label class="txt-sm txt-hint block m-b-5">URL</label>
                                <div class="read-only-value">{selectedPagePath}</div>
                            </div>
                        </div>

                        {#if pageEnabledField}
                            <label class="checkbox-field m-t-8">
                                <input type="checkbox" bind:checked={pageEditForm.enabled} />
                                <span>Enabled</span>
                            </label>
                        {/if}

                        <div class="form-actions m-t-sm">
                            <button type="button" class="btn btn-sm btn-strong" disabled={isSavingPage} on:click={savePage}>
                                {isSavingPage ? "Saving..." : "Save page"}
                            </button>
                        </div>

                        {#if pageError}
                            <p class="txt-danger m-t-8 m-b-0">{pageError}</p>
                        {/if}

                        <div class="sections-wrap m-t-base">
                            <div class="sections-head">
                                <h5 class="m-0">Sections on this page</h5>
                                <span class="txt-sm txt-hint">{blocks.length} total</span>
                            </div>

                            {#if !blockPageRelationField}
                                <p class="txt-sm txt-danger m-t-8 m-b-0">Blocks relation to pages is missing.</p>
                            {:else if !blockPropsField}
                                <p class="txt-sm txt-danger m-t-8 m-b-0">Blocks props field is missing.</p>
                            {:else if isLoadingBlocks || isLoadingComponents}
                                <p class="txt-sm txt-hint m-t-8 m-b-0">Loading sections...</p>
                            {:else if !blocks.length}
                                <p class="txt-sm txt-hint m-t-8 m-b-0">No sections linked to this page.</p>
                            {:else}
                                <div class="sections-list m-t-sm">
                                    {#each blocks as block, index}
                                        <article class="section-card" class:open={openSectionId === block.id}>
                                            <button type="button" class="section-toggle" on:click={() => toggleSection(block.id)}>
                                                <span class="section-toggle-content">
                                                    <strong>{getSectionTitle(block, index)}</strong>
                                                    <small>{getSectionDescription(block)}</small>
                                                </span>
                                                <i class={openSectionId === block.id ? "ri-arrow-up-s-line" : "ri-arrow-down-s-line"} />
                                            </button>

                                            {#if openSectionId === block.id}
                                                <div class="section-body">
                                                    {#if getSectionSchemaFields(block).length}
                                                        <SchemaForm
                                                            fields={getSectionSchemaFields(block)}
                                                            value={sectionPropsDraftById[block.id] || {}}
                                                            showImport={false}
                                                            path={`sections.${block.id}`}
                                                            on:propsChange={(event) => updateSectionDraft(block.id, event.detail)}
                                                        />
                                                    {:else}
                                                        <p class="txt-sm txt-hint m-b-0">
                                                            This section has no editable schema fields yet.
                                                        </p>
                                                    {/if}

                                                    <div class="form-actions m-t-sm">
                                                        <button
                                                            type="button"
                                                            class="btn btn-sm btn-strong"
                                                            disabled={!!isSavingSectionById[block.id] || !blockPropsField}
                                                            on:click={() => saveSection(block)}
                                                        >
                                                            {isSavingSectionById[block.id] ? "Saving..." : "Save section"}
                                                        </button>
                                                    </div>

                                                    {#if sectionErrorById[block.id]}
                                                        <p class="txt-danger m-t-8 m-b-0">{sectionErrorById[block.id]}</p>
                                                    {/if}
                                                </div>
                                            {/if}
                                        </article>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    {:else}
                        <div class="txt-hint">Select a page to edit.</div>
                    {/if}
                </div>
            </div>
        </section>
    {/if}
</PageWrapper>

<style>
    .cms-head {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .head-main {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 12px;
        flex-wrap: wrap;
    }

    .summary-title-wrap {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 260px;
    }

    .title-row {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 0;
    }

    .head-description {
        max-width: 520px;
    }

    .head-selector {
        width: min(100%, 620px);
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .selector-controls {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .selector-controls .input {
        flex: 1 1 auto;
        min-width: 260px;
    }

    .summary-badges {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        justify-content: flex-end;
        margin-left: auto;
    }

    .summary-pill {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseAlt1Color);
        color: var(--txtHintColor);
        font-size: 12px;
        padding: 5px 9px;
        white-space: nowrap;
    }

    .summary-pill i {
        color: var(--txtPrimaryColor);
        opacity: 0.85;
        font-size: 13px;
    }

    .summary-pill.warning {
        color: color-mix(in srgb, var(--dangerColor) 65%, var(--txtHintColor));
    }

    .cms-section-panel {
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .content-workspace-grid {
        display: grid;
        grid-template-columns: minmax(260px, 0.85fr) minmax(520px, 1.8fr);
        gap: 12px;
        align-items: stretch;
    }

    .pages-list-panel,
    .page-editor-panel {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        min-height: 460px;
        overflow: hidden;
        padding: 12px;
    }

    .pages-list-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .pages-list-body {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 620px;
        overflow: auto;
    }

    .page-row {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 10px;
        text-align: left;
        display: flex;
        flex-direction: column;
        gap: 4px;
        cursor: pointer;
    }

    .page-row:hover {
        background: color-mix(in srgb, var(--baseAlt1Color) 72%, var(--baseColor));
    }

    .page-row.active {
        border-color: color-mix(in srgb, var(--primaryColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 8%, var(--baseColor));
        box-shadow: inset 2px 0 0 color-mix(in srgb, var(--primaryColor) 60%, transparent);
    }

    .page-row-title {
        color: var(--txtPrimaryColor);
        font-weight: 600;
    }

    .page-row-path {
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
    }

    .page-editor-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .form-grid {
        display: grid;
        gap: 10px;
    }

    .form-grid.two-col {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .form-field {
        min-width: 0;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 8px 10px 10px;
    }

    .form-field .input {
        width: 100%;
        background: var(--baseColor);
    }

    .read-only-field .read-only-value {
        min-height: 32px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        display: flex;
        align-items: center;
        padding: 0 10px;
        color: var(--txtHintColor);
    }

    .checkbox-field {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
    }

    .form-actions {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
    }

    .sections-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .sections-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .sections-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .section-card {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        overflow: hidden;
    }

    .section-toggle {
        width: 100%;
        border: 0;
        background: transparent;
        text-align: left;
        padding: 10px 12px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        cursor: pointer;
    }

    .section-toggle:hover {
        background: color-mix(in srgb, var(--baseAlt1Color) 72%, var(--baseColor));
    }

    .section-toggle-content {
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
    }

    .section-toggle-content strong {
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .section-toggle-content small {
        color: var(--txtHintColor);
        font-size: 12px;
    }

    .section-body {
        border-top: 1px solid var(--baseAlt2Color);
        background: var(--baseColor);
        padding: 12px;
    }

    @media (max-width: 1200px) {
        .content-workspace-grid {
            grid-template-columns: 1fr;
        }

        .pages-list-panel,
        .page-editor-panel {
            min-height: auto;
        }
    }

    @media (max-width: 840px) {
        .cms-head,
        .cms-section-panel {
            padding: calc(var(--baseSpacing) - 12px) calc(var(--baseSpacing) - 10px);
        }

        .selector-controls {
            flex-direction: column;
            align-items: stretch;
        }

        .selector-controls .input {
            min-width: 0;
        }

        .summary-badges {
            justify-content: flex-start;
            margin-left: 0;
        }

        .form-grid.two-col {
            grid-template-columns: 1fr;
        }
    }
</style>
