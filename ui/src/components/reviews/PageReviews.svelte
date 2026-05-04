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
    <section class="reviews-page-head operations-head panel m-b-base">
        <div class="head-main">
            <div class="summary-title-wrap">
                <div class="title-row">
                    <h2 class="m-0">Reviews</h2>
                    <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={() => reviewsCollectionDashboard?.reload?.()} />
                </div>
                <p class="txt-sm txt-hint m-b-0 head-description">
                    Monitor customer reviews and social proof for this website.
                </p>
            </div>
        </div>
    </section>

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

<style>
    .reviews-page-head.operations-head .head-description {
        max-width: 520px;
    }
</style>
