<script>
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import ReviewsCollectionDashboard from "@/components/records/ReviewsCollectionDashboard.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";

    // NUVIO CUSTOM START: Dedicated Reviews section/page backed by the reviews collection storage.
    $pageTitle = "Reviews";

    loadCollections();

    $: websitesCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "websites") || null;
    // NUVIO CUSTOM END: Dedicated Reviews section/page backed by the reviews collection storage.
</script>

<PageWrapper>
    {#if $isCollectionsLoading && !websitesCollection}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading reviews...</h1>
        </div>
    {:else if !websitesCollection}
        <div class="alert alert-danger">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>Website data is unavailable. Reviews requires website setup.</div>
        </div>
    {:else}
        <ReviewsCollectionDashboard {websitesCollection} />
    {/if}
</PageWrapper>
