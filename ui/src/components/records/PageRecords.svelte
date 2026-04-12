<script>
    import { tick } from "svelte";
    import { querystring } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import tooltip from "@/actions/tooltip";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import Searchbar from "@/components/base/Searchbar.svelte";
    import CollectionDocsPanel from "@/components/collections/CollectionDocsPanel.svelte";
    import CollectionUpsertPanel from "@/components/collections/CollectionUpsertPanel.svelte";
    import CollectionsSidebar from "@/components/collections/CollectionsSidebar.svelte";
    import RecordPreviewPanel from "@/components/records/RecordPreviewPanel.svelte";
    import RecordUpsertPanel from "@/components/records/RecordUpsertPanel.svelte";
    import RecordsCount from "@/components/records/RecordsCount.svelte";
    import RecordsList from "@/components/records/RecordsList.svelte";
    import { hideControls, pageTitle } from "@/stores/app";
    import {
        activeCollection,
        changeActiveCollectionByIdOrName,
        collections,
        isCollectionsLoading,
        loadCollections,
    } from "@/stores/collections";
    import ClientRoleUi from "@/utils/ClientRoleUi";

    const initialQueryParams = new URLSearchParams($querystring);
    const clientVisibleCollectionNames = ClientRoleUi.clientEditableCollectionNames;

    let collectionUpsertPanel;
    let collectionDocsPanel;
    let recordUpsertPanel;
    let recordPreviewPanel;
    let clientBlocksUpsertPanel;
    let recordsList;
    let recordsCount;
    let filter = initialQueryParams.get("filter") || "";
    let sort = initialQueryParams.get("sort") || "-@rowid";
    let selectedCollectionIdOrName = initialQueryParams.get("collection") || $activeCollection?.id;
    let totalCount = 0; // used to manully change the count without the need of reloading the recordsCount component
    let canManageCollectionActions = false;
    let canManageRecordActions = false;
    let canCreateRecords = false;
    let isClientCollectionMode = false;

    loadCollections(selectedCollectionIdOrName);

    $: canManageCollectionActions = ApiClient.isAdminSuperuser();

    $: isClientCollectionMode = ApiClient.isClientSuperuser();

    $: canManageRecordActions = ApiClient.isAdminSuperuser()
        || (isClientCollectionMode
            && clientVisibleCollectionNames.has(($activeCollection?.name || "").toLowerCase()));

    $: canCreateRecords = ApiClient.isAdminSuperuser();

    $: reactiveParams = new URLSearchParams($querystring);

    $: collectionQueryParam = reactiveParams.get("collection");

    $: if (
        !$isCollectionsLoading &&
        collectionQueryParam &&
        collectionQueryParam != selectedCollectionIdOrName &&
        collectionQueryParam != $activeCollection?.id &&
        collectionQueryParam != $activeCollection?.name
    ) {
        changeActiveCollectionByIdOrName(collectionQueryParam);
    }

    $: allowedClientCollections = $collections.filter((c) =>
        clientVisibleCollectionNames.has((c?.name || "").toLowerCase()),
    );

    $: blocksCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "blocks") || null;

    $: roleVisibleCollectionsCount = isClientCollectionMode
        ? allowedClientCollections.length
        : $collections.length;

    $: isClientActiveCollectionAllowed = !isClientCollectionMode
        || clientVisibleCollectionNames.has(($activeCollection?.name || "").toLowerCase());

    $: if (
        isClientCollectionMode &&
        !$isCollectionsLoading &&
        $activeCollection?.id &&
        !clientVisibleCollectionNames.has(($activeCollection?.name || "").toLowerCase())
    ) {
        const fallback = allowedClientCollections[0];
        if (fallback) {
            changeActiveCollectionByIdOrName(fallback.id);
            selectedCollectionIdOrName = fallback.id;
            updateQueryParams({
                collection: fallback.id,
                recordId: null,
            });
        }
    }

    // reset filter and sort on collection change
    $: if (
        $activeCollection?.id &&
        selectedCollectionIdOrName != $activeCollection.id &&
        selectedCollectionIdOrName != $activeCollection.name
    ) {
        reset();
    }

    $: if ($activeCollection?.id) {
        normalizeSort();
    }

    $: if (!$isCollectionsLoading && initialQueryParams.get("recordId")) {
        showRecordById(initialQueryParams.get("recordId"));
    }

    // keep the url params in sync
    $: if (!$isCollectionsLoading && (sort || filter || $activeCollection?.id)) {
        updateQueryParams();
    }

    $: $pageTitle = $activeCollection?.name || "Collections";

    async function showRecordById(recordId) {
        await tick(); // ensure that the reactive component params are resolved

        $activeCollection?.type === "view" || !canManageRecordActions
            ? recordPreviewPanel.show(recordId)
            : recordUpsertPanel?.show(recordId);
    }

    function handleRecordSelect(record) {
        updateQueryParams({
            recordId: record.id,
        });

        const showModel = record._partial ? record.id : record;

        $activeCollection.type === "view" || !canManageRecordActions
            ? recordPreviewPanel?.show(showModel)
            : recordUpsertPanel?.show(showModel);
    }

    function reset() {
        selectedCollectionIdOrName = $activeCollection?.id;
        filter = "";
        sort = "-@rowid";

        normalizeSort();

        updateQueryParams({ recordId: null });

        // close any open collection panels
        collectionUpsertPanel?.forceHide();
        collectionDocsPanel?.hide();
    }

    // ensures that the sort fields exist in the collection
    async function normalizeSort() {
        if (!sort) {
            return; // nothing to normalize
        }

        const collectionFields = CommonHelper.getAllCollectionIdentifiers($activeCollection);

        const sortFields = sort.split(",").map((f) => {
            if (f.startsWith("+") || f.startsWith("-")) {
                return f.substring(1);
            }
            return f;
        });

        // invalid sort expression or missing sort field
        if (sortFields.filter((f) => collectionFields.includes(f)).length != sortFields.length) {
            if ($activeCollection?.type != "view") {
                sort = "-@rowid"; // all collections with exception to the view has this field
            } else if (collectionFields.includes("created")) {
                // common autodate field
                sort = "-created";
            } else {
                sort = "";
            }
        }
    }

    function updateQueryParams(extra = {}) {
        const queryParams = Object.assign(
            {
                collection: $activeCollection?.id || "",
                filter: filter,
                sort: sort,
            },
            extra,
        );

        CommonHelper.replaceHashQueryParams(queryParams);
    }
</script>

{#if $isCollectionsLoading && !roleVisibleCollectionsCount}
    <PageWrapper center>
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading collections...</h1>
        </div>
    </PageWrapper>
{:else if !roleVisibleCollectionsCount}
    <PageWrapper center>
        <div class="placeholder-section m-b-base">
            <div class="icon">
                <i class="ri-database-2-line" />
            </div>
            {#if isClientCollectionMode}
                <h1 class="m-b-10">No client collections available yet.</h1>
            {:else if $hideControls}
                <h1 class="m-b-10">You don't have any collections yet.</h1>
            {:else if canManageCollectionActions}
                <h1 class="m-b-10">Create your first collection to add records!</h1>
                <button
                    type="button"
                    class="btn btn-expanded-lg btn-lg"
                    on:click={() => collectionUpsertPanel?.show()}
                >
                    <i class="ri-add-line" />
                    <span class="txt">Create new collection</span>
                </button>
            {/if}
        </div>
    </PageWrapper>
{:else if isClientCollectionMode && !isClientActiveCollectionAllowed}
    <PageWrapper center>
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading collections...</h1>
        </div>
    </PageWrapper>
{:else}
    <CollectionsSidebar />

    <PageWrapper class="flex-content">
        <header class="page-header">
            <nav class="breadcrumbs">
                <div class="breadcrumb-item">Collections</div>
                <div class="breadcrumb-item">{$activeCollection.name}</div>
            </nav>

            <div class="inline-flex gap-5">
                {#if !$hideControls && canManageCollectionActions}
                    <button
                        type="button"
                        aria-label="Edit collection"
                        class="btn btn-transparent btn-circle"
                        use:tooltip={{ text: "Edit collection", position: "right" }}
                        on:click={() => collectionUpsertPanel?.show($activeCollection)}
                    >
                        <i class="ri-settings-4-line" />
                    </button>
                {/if}

                <RefreshButton
                    on:refresh={() => {
                        recordsList?.load();
                        recordsCount?.reload();
                    }}
                />
            </div>

            <div class="btns-group">
                {#if canManageCollectionActions}
                    <button
                        type="button"
                        class="btn btn-outline"
                        on:click={() => collectionDocsPanel?.show($activeCollection)}
                    >
                        <i class="ri-code-s-slash-line" />
                        <span class="txt">API Preview</span>
                    </button>
                {/if}

                {#if $activeCollection.type !== "view" && canCreateRecords}
                    <button type="button" class="btn btn-expanded" on:click={() => recordUpsertPanel?.show()}>
                        <i class="ri-add-line" />
                        <span class="txt">New record</span>
                    </button>
                {/if}
            </div>
        </header>

        <Searchbar
            value={filter}
            autocompleteCollection={$activeCollection}
            on:submit={(e) => (filter = e.detail)}
        />

        <div class="clearfix m-b-sm" />

        <RecordsList
            bind:this={recordsList}
            collection={$activeCollection}
            bind:filter
            bind:sort
            on:select={(e) => handleRecordSelect(e.detail)}
            on:delete={() => {
                recordsCount?.reload();
            }}
            on:new={() => {
                if (!canCreateRecords) {
                    return;
                }
                recordUpsertPanel?.show();
            }}
        />

        <svelte:fragment slot="footer">
            <RecordsCount
                bind:this={recordsCount}
                class="m-r-auto txt-sm txt-hint"
                collection={$activeCollection}
                {filter}
                bind:totalCount
            />
        </svelte:fragment>
    </PageWrapper>
{/if}

<CollectionUpsertPanel
    bind:this={collectionUpsertPanel}
    on:truncate={() => {
        recordsList?.load();
        recordsCount?.reload();
    }}
/>

<CollectionDocsPanel bind:this={collectionDocsPanel} />

<RecordUpsertPanel
    bind:this={recordUpsertPanel}
    collection={$activeCollection}
    on:clientblockedit={(e) => clientBlocksUpsertPanel?.show(e.detail)}
    on:hide={() => {
        updateQueryParams({ recordId: null });
    }}
    on:save={(e) => {
        if (filter) {
            // if there is applied filter, reload the count since we
            // don't know after the save whether the record satisfies it
            recordsCount?.reload();
        } else if (e.detail.isNew) {
            totalCount++;
        }

        recordsList?.reloadLoadedPages();
    }}
    on:delete={(e) => {
        if (!filter || recordsList?.hasRecord(e.detail.id)) {
            totalCount--;
        }

        recordsList?.reloadLoadedPages();
    }}
/>

<RecordUpsertPanel
    bind:this={clientBlocksUpsertPanel}
    collection={blocksCollection}
    on:save={() => recordUpsertPanel?.reloadClientPageBlocks?.()}
    on:delete={() => recordUpsertPanel?.reloadClientPageBlocks?.()}
/>

<RecordPreviewPanel
    bind:this={recordPreviewPanel}
    collection={$activeCollection}
    on:hide={() => {
        updateQueryParams({ recordId: null });
    }}
/>
