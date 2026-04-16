<script>
    import PageSidebar from "@/components/base/PageSidebar.svelte";
    import CollectionSidebarItem from "@/components/collections/CollectionSidebarItem.svelte";
    import CollectionUpsertPanel from "@/components/collections/CollectionUpsertPanel.svelte";
    import { hideControls } from "@/stores/app";
    import { activeCollection, collections, isCollectionsLoading } from "@/stores/collections";
    import ApiClient from "@/utils/ApiClient";

    const pinnedStorageKey = "@pinnedCollections";
    const clientVisibleCollectionNames = new Set(["assets", "blocks", "pages", "websites"]);
    // NUVIO CUSTOM START: Admin collection sidebar grouped aggregators.
    const customCollectionsGroups = [
        {
            key: "site",
            label: "Site",
            collectionNames: new Set(["websites", "website", "pages", "page", "blocks", "components", "assets"]),
        },
        {
            key: "leads",
            label: "Leads",
            collectionNames: new Set(["leads", "lead", "contacts", "contact", "forms", "submissions"]),
        },
        {
            key: "markting",
            label: "Markting",
            collectionNames: new Set(["markting", "marketing", "campaigns", "newsletter", "newsletters"]),
        },
        {
            key: "reviews",
            label: "Reviews",
            collectionNames: new Set(["reviews", "review", "testimonials", "ratings"]),
        },
    ];
    // NUVIO CUSTOM END: Admin collection sidebar grouped aggregators.

    let collectionPanel;
    let searchTerm = "";
    let pinnedIds = [];
    let showSystemSection = false;
    // NUVIO CUSTOM START: Expand/collapse state per custom sidebar group.
    let showCustomGroupSections = {
        site: true,
        leads: true,
        markting: true,
        reviews: true,
    };
    // NUVIO CUSTOM END: Expand/collapse state per custom sidebar group.
    let oldCollectionId;
    let canManageCollections = false;
    let isClientCollectionMode = false;

    loadPinned();

    $: if ($collections) {
        syncPinned();
        scrollIntoView();
    }

    $: normalizedSearch = searchTerm.replace(/\s+/g, "").toLowerCase();

    $: hasSearch = searchTerm !== "";

    $: if (pinnedIds) {
        localStorage.setItem(pinnedStorageKey, JSON.stringify(pinnedIds));
    }

    $: canManageCollections = ApiClient.isAdminSuperuser();

    $: isClientCollectionMode = ApiClient.isClientSuperuser();

    $: visibleCollections = $collections.filter((c) => {
        if (!isClientCollectionMode) {
            return true;
        }

        return clientVisibleCollectionNames.has((c?.name || "").toLowerCase());
    });

    $: filtered = visibleCollections.filter((c) => {
        return c.id == searchTerm || c.name?.replace(/\s+/g, "")?.toLowerCase()?.includes(normalizedSearch);
    });

    $: pinnedCollections = filtered.filter((c) => pinnedIds.includes(c.id));

    $: unpinnedRegularCollections = filtered.filter((c) => !c.system && !pinnedIds.includes(c.id));

    $: unpinnedSystemCollections = filtered.filter((c) => c.system && !pinnedIds.includes(c.id));

    // NUVIO CUSTOM START: Partition regular collections into custom groups + fallback Others.
    $: groupedRegularCollections = customCollectionsGroups.map((group) => {
        return {
            ...group,
            collections: unpinnedRegularCollections.filter((collection) => {
                return group.collectionNames.has((collection?.name || "").toLowerCase());
            }),
        };
    });

    $: groupedCollectionNames = new Set(
        groupedRegularCollections.flatMap((group) => group.collections.map((collection) => collection.id)),
    );

    $: ungroupedRegularCollections = unpinnedRegularCollections.filter((collection) => {
        return !groupedCollectionNames.has(collection.id);
    });
    // NUVIO CUSTOM END: Partition regular collections into custom groups + fallback Others.

    $: if ($activeCollection?.id && oldCollectionId != $activeCollection.id) {
        oldCollectionId = $activeCollection.id;
        if ($activeCollection.system && !pinnedCollections.find((c) => c.id == $activeCollection.id)) {
            showSystemSection = true;
        } else {
            showSystemSection = false;
        }

        // NUVIO CUSTOM START: Auto-expand the active collection custom group.
        if (!$activeCollection.system && !pinnedCollections.find((c) => c.id == $activeCollection.id)) {
            for (const group of customCollectionsGroups) {
                if (group.collectionNames.has(($activeCollection?.name || "").toLowerCase())) {
                    if (!showCustomGroupSections[group.key]) {
                        showCustomGroupSections[group.key] = true;
                        showCustomGroupSections = { ...showCustomGroupSections };
                    }
                    break;
                }
            }
        }
        // NUVIO CUSTOM END: Auto-expand the active collection custom group.
    }

    function scrollIntoView() {
        setTimeout(() => {
            const activeItem = document.querySelector(".collection-sidebar .sidebar-list-item.active");
            if (activeItem) {
                activeItem?.scrollIntoView({ block: "nearest" });
            }
        }, 0);
    }

    function loadPinned() {
        pinnedIds = [];

        try {
            const encoded = localStorage.getItem(pinnedStorageKey);
            if (encoded) {
                pinnedIds = JSON.parse(encoded) || [];
            }
        } catch (_) {}
    }

    function syncPinned() {
        pinnedIds = pinnedIds.filter((id) => !!$collections.find((c) => c.id == id));
    }
</script>

<PageSidebar class="collection-sidebar">
    <header class="sidebar-header">
        <div class="form-field search" class:active={hasSearch}>
            <div class="form-field-addon">
                <button
                    type="button"
                    class="btn btn-xs btn-transparent btn-circle btn-clear"
                    class:hidden={!hasSearch}
                    on:click={() => (searchTerm = "")}
                >
                    <i class="ri-close-line" />
                </button>
            </div>
            <input
                type="text"
                placeholder="Search collections..."
                name="collections-search"
                bind:value={searchTerm}
            />
        </div>
    </header>

    <hr class="m-t-5 m-b-xs" />

    <div
        class="sidebar-content"
        class:fade={$isCollectionsLoading}
        class:sidebar-content-compact={filtered.length > 20}
    >
        {#if isClientCollectionMode}
            {#each filtered as collection (collection.id)}
                <CollectionSidebarItem {collection} bind:pinnedIds />
            {/each}
        {:else}
            {#if pinnedCollections.length}
                <div class="sidebar-title">Pinned</div>
                {#each pinnedCollections as collection (collection.id)}
                    <CollectionSidebarItem {collection} bind:pinnedIds />
                {/each}
            {/if}

            {#if unpinnedRegularCollections.length}
                <!-- NUVIO CUSTOM START: Render custom grouped sidebar sections (Site/Leads/Markting/Reviews). -->
                {#each groupedRegularCollections as group (group.key)}
                    <button
                        type="button"
                        class="sidebar-title m-b-xs"
                        class:link-hint={!normalizedSearch.length}
                        aria-label={showCustomGroupSections[group.key]
                            ? "Collapse grouped collections"
                            : "Expand grouped collections"}
                        aria-expanded={showCustomGroupSections[group.key] || normalizedSearch.length}
                        disabled={normalizedSearch.length}
                        on:click={() => {
                            if (!normalizedSearch.length) {
                                showCustomGroupSections[group.key] = !showCustomGroupSections[group.key];
                                showCustomGroupSections = { ...showCustomGroupSections };
                            }
                        }}
                    >
                        <span class="txt">{group.label}</span>
                        {#if !normalizedSearch.length}
                            <i
                                class="ri-arrow-{showCustomGroupSections[group.key] ? 'up' : 'down'}-s-line"
                                aria-hidden="true"
                            />
                        {/if}
                    </button>
                    {#if showCustomGroupSections[group.key] || normalizedSearch.length}
                        {#each group.collections as collection (collection.id)}
                            <CollectionSidebarItem {collection} bind:pinnedIds />
                        {/each}
                    {/if}
                {/each}

                {#if ungroupedRegularCollections.length}
                    <div class="sidebar-title">Others</div>
                    {#each ungroupedRegularCollections as collection (collection.id)}
                        <CollectionSidebarItem {collection} bind:pinnedIds />
                    {/each}
                {/if}
                <!-- NUVIO CUSTOM END: Render custom grouped sidebar sections (Site/Leads/Markting/Reviews). -->
            {/if}

            {#if unpinnedSystemCollections.length}
                <button
                    type="button"
                    class="sidebar-title m-b-xs"
                    class:link-hint={!normalizedSearch.length}
                    aria-label={showSystemSection ? "Expand system collections" : "Collapse system collections"}
                    aria-expanded={showSystemSection || normalizedSearch.length}
                    disabled={normalizedSearch.length}
                    on:click={() => {
                        if (!normalizedSearch.length) {
                            showSystemSection = !showSystemSection;
                        }
                    }}
                >
                    <span class="txt">System</span>
                    {#if !normalizedSearch.length}
                        <i class="ri-arrow-{showSystemSection ? 'up' : 'down'}-s-line" aria-hidden="true" />
                    {/if}
                </button>
                {#if showSystemSection || normalizedSearch.length}
                    {#each unpinnedSystemCollections as collection (collection.id)}
                        <CollectionSidebarItem {collection} bind:pinnedIds />
                    {/each}
                {/if}
            {/if}
        {/if}

        {#if normalizedSearch.length && !filtered.length}
            <p class="txt-hint m-t-10 m-b-10 txt-center">No collections found.</p>
        {/if}
    </div>

    {#if !$hideControls && canManageCollections}
        <footer class="sidebar-footer">
            <button type="button" class="btn btn-block btn-outline" on:click={() => collectionPanel?.show()}>
                <i class="ri-add-line" />
                <span class="txt">New collection</span>
            </button>
        </footer>
    {/if}
</PageSidebar>

<CollectionUpsertPanel bind:this={collectionPanel} />
