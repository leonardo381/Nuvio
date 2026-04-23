<script>
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import ReviewsCollectionDashboard from "@/components/records/ReviewsCollectionDashboard.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";

    // NUVIO CUSTOM START: Dedicated Reviews section/page backed by the reviews collection storage.
    $pageTitle = "Reviews";

    let reviewsCollectionDashboard;

    loadCollections();

    $: websitesCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "websites") || null;
    // NUVIO CUSTOM END: Dedicated Reviews section/page backed by the reviews collection storage.
</script>

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Reviews</div>
        </nav>

        <RefreshButton on:refresh={() => reviewsCollectionDashboard?.reload?.()} />
    </header>

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
            <div>Websites collection not found. Reviews module requires Websites to be available.</div>
        </div>
    {:else}
        <ReviewsCollectionDashboard bind:this={reviewsCollectionDashboard} {websitesCollection} />
    {/if}
</PageWrapper>
