<script>
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import ReportsCollectionDashboard from "@/components/records/ReportsCollectionDashboard.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";

    // NUVIO CUSTOM START: Dedicated Reports section/page.
    $pageTitle = "Reports";

    let reportsCollectionDashboard;

    loadCollections();

    $: websitesCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "websites") || null;
    // NUVIO CUSTOM END: Dedicated Reports section/page.
</script>

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Reports</div>
        </nav>

        <RefreshButton on:refresh={() => reportsCollectionDashboard?.reload?.()} />
    </header>

    {#if $isCollectionsLoading && !websitesCollection}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading reports...</h1>
        </div>
    {:else if !websitesCollection}
        <div class="alert alert-danger">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>Websites collection not found. Reports module requires Websites to be available.</div>
        </div>
    {:else}
        <ReportsCollectionDashboard bind:this={reportsCollectionDashboard} {websitesCollection} />
    {/if}
</PageWrapper>
