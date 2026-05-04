<script>
    // NUVIO CUSTOM START: Collection-backed dashboard Reviews module UI.
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addErrorToast, addInfoToast, addSuccessToast } from "@/stores/toasts";

    export let websitesCollection = null;

    const tabKeys = {
        reviews: "reviews",
        display: "display",
        sources: "sources",
    };

    const ratingFilterOptions = [
        { value: "all", label: "All ratings" },
        { value: "5", label: "5 stars" },
        { value: "4plus", label: "4+ stars" },
        { value: "3plus", label: "3+ stars" },
        { value: "2plus", label: "2+ stars" },
        { value: "1plus", label: "1+ stars" },
    ];

    const positiveResponseTemplates = [
        "Thank you so much for your kind review. We appreciate your trust and support.",
        "We are glad you had a great experience. Thank you for choosing us and sharing your feedback.",
        "Thanks for the amazing review. It means a lot to our team and helps other clients find us.",
    ];

    const negativeResponseTemplates = [
        "Thank you for the feedback. We are sorry your experience did not match expectations and we want to make it right.",
        "We appreciate your honesty. Please contact us directly so we can understand what happened and improve quickly.",
        "Sorry for the inconvenience. Your feedback is valuable and we are already working on improvements.",
    ];

    let websites = [];
    let selectedWebsiteId = "";
    let dashboard = null;
    let isLoadingWebsites = false;
    let isLoadingDashboard = false;
    let isSyncing = false;
    let syncError = "";

    let activeTab = tabKeys.reviews;
    let sourceFilter = "all";
    let ratingFilter = "all";
    let searchTerm = "";

    let lastWebsitesCollectionId = "";
    let lastDashboardWebsiteId = "";

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        dashboard = null;
        lastWebsitesCollectionId = "";
        lastDashboardWebsiteId = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: if (selectedWebsiteId && selectedWebsiteId !== lastDashboardWebsiteId) {
        lastDashboardWebsiteId = selectedWebsiteId;
        loadDashboard();
    }

    $: dashboardState = dashboard?.state || "";
    $: allReviews = Array.isArray(dashboard?.reviews) ? dashboard.reviews : [];
    $: normalizedSearchTerm = normalizeLower(searchTerm);
    $: sourceOptions = Array.from(
        new Set(allReviews.map((review) => normalizeLower(review?.source)).filter(Boolean)),
    ).sort((a, b) => a.localeCompare(b));
    $: hasGoogleSource = sourceOptions.includes("google")
        || !!`${dashboard?.googlePlaceId || ""}`.trim()
        || !!`${dashboard?.openOnGoogle || ""}`.trim();
    $: filteredReviews = allReviews.filter((review) => {
        if (sourceFilter !== "all" && normalizeLower(review?.source) !== sourceFilter) {
            return false;
        }

        if (!matchesRatingFilter(review?.rating, ratingFilter)) {
            return false;
        }

        if (!normalizedSearchTerm) {
            return true;
        }

        const haystack = [
            `${review?.authorName || ""}`,
            `${review?.text || ""}`,
            `${review?.source || ""}`,
            `${review?.relativeTime || ""}`,
            `${review?.publishedAt || ""}`,
        ].join(" ").toLowerCase();

        return haystack.includes(normalizedSearchTerm);
    });

    export function reload() {
        if (!selectedWebsiteId) {
            return;
        }

        return loadDashboard();
    }

    function normalizeLower(value) {
        return `${value || ""}`.trim().toLowerCase();
    }

    function resolveWebsitesSort(collection) {
        const preferredSortFields = ["title", "name", "slug"];
        const availableFields = new Set(
            CommonHelper.getAllCollectionIdentifiers(collection).map((field) => `${field || ""}`.trim().toLowerCase()),
        );
        const validSortFields = preferredSortFields.filter((field) => availableFields.has(field));

        if (!validSortFields.length) {
            return "+id";
        }

        return validSortFields.map((field) => `+${field}`).join(",");
    }

    async function loadWebsites() {
        if (!websitesCollection?.id) {
            websites = [];
            selectedWebsiteId = "";
            dashboard = null;
            return;
        }

        isLoadingWebsites = true;

        try {
            websites = await ApiClient.collection(websitesCollection.id).getFullList({
                sort: resolveWebsitesSort(websitesCollection),
                requestKey: "nuvio_reviews_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                dashboard = null;
                lastDashboardWebsiteId = "";
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            } else {
                lastDashboardWebsiteId = "";
                await loadDashboard();
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            dashboard = null;
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadDashboard() {
        if (!selectedWebsiteId) {
            dashboard = null;
            return;
        }

        isLoadingDashboard = true;

        try {
            dashboard = await ApiClient.send("/api/nuvio/reviews/dashboard", {
                method: "GET",
                query: {
                    websiteId: selectedWebsiteId,
                },
                requestKey: "nuvio_reviews_dashboard_" + selectedWebsiteId,
            });
        } catch (err) {
            dashboard = null;
            ApiClient.error(err);
        }

        isLoadingDashboard = false;
    }

    async function syncFromGoogle() {
        if (!selectedWebsiteId || isSyncing) {
            return;
        }

        isSyncing = true;
        syncError = "";

        try {
            dashboard = await ApiClient.send("/api/nuvio/reviews/sync", {
                method: "POST",
                body: {
                    websiteId: selectedWebsiteId,
                },
                requestKey: "nuvio_reviews_sync_" + selectedWebsiteId,
            });

            addSuccessToast("Reviews synced successfully.");
        } catch (err) {
            syncError = err?.response?.message || err?.message || "Failed to sync reviews.";
            addErrorToast(syncError);
        }

        isSyncing = false;
    }

    function handleWebsiteChange(event) {
        selectedWebsiteId = event.target.value || "";
        syncError = "";
    }

    function resolveWebsiteLabel(website) {
        return (
            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim() || website?.id || ""
        );
    }

    function resolveStateMessage(state) {
        if (state === "feature_unavailable") {
            return "Reviews are not available for this website right now.";
        }

        if (state === "disabled") {
            return "Reviews are currently turned off in Website Settings.";
        }

        if (state === "not_configured") {
            return "Connect a Google Place ID in Website Settings to sync reviews.";
        }

        if (state === "never_synced") {
            return "No reviews have been synced yet. Click Sync now to import reviews from Google.";
        }

        return "";
    }

    function resolveStateClass(state) {
        if (state === "feature_unavailable" || state === "disabled") {
            return "alert-warning";
        }

        if (state === "not_configured") {
            return "alert-danger";
        }

        if (state === "never_synced") {
            return "alert-info";
        }

        return "alert-success";
    }

    function resolveStateLabel(state) {
        if (state === "feature_unavailable") {
            return "Unavailable";
        }

        if (state === "disabled") {
            return "Disabled";
        }

        if (state === "not_configured") {
            return "Not configured";
        }

        if (state === "never_synced") {
            return "Never synced";
        }

        return "Ready";
    }

    function resolveStateLabelClass(state) {
        if (state === "feature_unavailable" || state === "disabled") {
            return "label-warning";
        }

        if (state === "not_configured") {
            return "label-danger";
        }

        if (state === "never_synced") {
            return "label-info";
        }

        return "label-success";
    }

    function formatDateTime(value) {
        const raw = `${value || ""}`.trim();
        if (!raw) {
            return "Never";
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const parsed = new Date(normalized);
        if (Number.isNaN(parsed.getTime())) {
            return raw;
        }

        return parsed.toLocaleString();
    }

    function formatRating(value) {
        const numeric = Number(value);
        if (Number.isNaN(numeric)) {
            return "-";
        }

        return numeric.toFixed(1);
    }

    function matchesRatingFilter(ratingValue, filterValue) {
        if (filterValue === "all") {
            return true;
        }

        const numeric = Number(ratingValue);
        if (Number.isNaN(numeric)) {
            return false;
        }

        if (filterValue === "5") {
            return numeric >= 5;
        }

        if (filterValue === "4plus") {
            return numeric >= 4;
        }

        if (filterValue === "3plus") {
            return numeric >= 3;
        }

        if (filterValue === "2plus") {
            return numeric >= 2;
        }

        if (filterValue === "1plus") {
            return numeric >= 1;
        }

        return true;
    }

    function buildStarSlots(value) {
        const numeric = Number(value);
        const safeRating = Number.isNaN(numeric) ? 0 : Math.max(0, Math.min(5, numeric));

        return [1, 2, 3, 4, 5].map((step) => safeRating >= step - 0.25);
    }

    function truncateText(value, max = 260) {
        return CommonHelper.truncate(`${value || ""}`.trim(), max, true);
    }

    function resolveSourceLabel(source) {
        const normalized = normalizeLower(source);
        if (!normalized) {
            return "Unknown";
        }

        if (normalized === "google") {
            return "Google";
        }

        return CommonHelper.sentenize(normalized.replace(/_/g, " "), false);
    }

    function resolveReviewTimeline(review) {
        const parts = [];

        if (`${review?.relativeTime || ""}`.trim()) {
            parts.push(`${review.relativeTime}`.trim());
        }

        if (`${review?.publishedAt || ""}`.trim()) {
            parts.push(formatDateTime(review.publishedAt));
        } else if (`${review?.syncedAt || ""}`.trim()) {
            parts.push(`Synced ${formatDateTime(review.syncedAt)}`);
        }

        return parts.join(" · ") || "No date information";
    }

    function clearReviewsFilters() {
        sourceFilter = "all";
        ratingFilter = "all";
        searchTerm = "";
    }

    function openExternal(url) {
        const normalized = `${url || ""}`.trim();
        if (!normalized) {
            return;
        }

        window.open(normalized, "_blank", "noopener,noreferrer");
    }

    async function copyText(text, label) {
        const value = `${text || ""}`.trim();
        if (!value) {
            return;
        }

        await CommonHelper.copyToClipboard(value);
        addInfoToast(`${label} copied to clipboard.`);
    }
    // NUVIO CUSTOM END: Collection-backed dashboard Reviews module UI.
</script>

<section class="reviews-dashboard-module">
    <section class="panel reviews-controls-panel">
        <div class="reviews-controls-row">
            <div class="reviews-selector-wrap">
                <label class="txt-sm txt-hint selector-label m-b-0" for="reviews-website">Website</label>
                <select
                    id="reviews-website"
                    class="input input-sm"
                    value={selectedWebsiteId}
                    disabled={isLoadingWebsites || !websites.length}
                    on:change={handleWebsiteChange}
                >
                    {#if !websites.length}
                        <option value="">No websites available</option>
                    {:else}
                        {#each websites as website (website.id)}
                            <option value={website.id}>{resolveWebsiteLabel(website)}</option>
                        {/each}
                    {/if}
                </select>
            </div>

            <div class="flex-fill" />

            <div class="reviews-controls-actions">
                <button
                    type="button"
                    class="btn btn-sm btn-outline"
                    disabled={!selectedWebsiteId || isLoadingDashboard || isLoadingWebsites}
                    on:click={() => loadDashboard()}
                >
                    <i class="ri-refresh-line" />
                    <span class="txt">Reload</span>
                </button>

                <button
                    type="button"
                    class="btn btn-sm"
                    class:btn-loading={isSyncing}
                    disabled={!selectedWebsiteId || isSyncing || isLoadingDashboard || isLoadingWebsites}
                    on:click={() => syncFromGoogle()}
                >
                    <i class="ri-loop-right-line" />
                    <span class="txt">Sync now</span>
                </button>
            </div>
        </div>
    </section>

    {#if isLoadingWebsites}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading websites...</span>
        </div>
    {:else if !websites.length}
        <div class="empty-state">No website records are available yet.</div>
    {:else if isLoadingDashboard}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading reviews...</span>
        </div>
    {:else if !dashboard}
        <div class="empty-state">Unable to load reviews for this website right now.</div>
    {:else}
        {#if dashboardState !== "ready"}
            <div class="alert {resolveStateClass(dashboardState)} m-b-sm">
                <div class="icon">
                    <i class="ri-information-line" />
                </div>
                <div>{resolveStateMessage(dashboardState)}</div>
            </div>
        {/if}

        {#if syncError}
            <div class="alert alert-danger m-b-sm">
                <div class="icon">
                    <i class="ri-error-warning-line" />
                </div>
                <div>{syncError}</div>
            </div>
        {/if}

        <div class="tabs-header compact combined left operations-tabs operations-tabs--nested">
            <button
                type="button"
                class="tab-item"
                class:active={activeTab === tabKeys.reviews}
                on:click={() => (activeTab = tabKeys.reviews)}
            >
                <i class="ri-star-line tab-icon" aria-hidden="true" />
                <span class="tab-label">Reviews</span>
            </button>
            <button
                type="button"
                class="tab-item"
                class:active={activeTab === tabKeys.display}
                on:click={() => (activeTab = tabKeys.display)}
            >
                <i class="ri-layout-grid-line tab-icon" aria-hidden="true" />
                <span class="tab-label">Display Settings</span>
            </button>
            <button
                type="button"
                class="tab-item"
                class:active={activeTab === tabKeys.sources}
                on:click={() => (activeTab = tabKeys.sources)}
            >
                <i class="ri-links-line tab-icon" aria-hidden="true" />
                <span class="tab-label">Sources</span>
            </button>
        </div>

        {#if activeTab === tabKeys.reviews}
            <section class="panel reviews-summary-panel">
                <div class="summary-badges">
                    <span class="summary-pill">
                        <i class="ri-star-smile-line" />
                        Average rating: <strong>{dashboard.summary.averageRating === null ? "-" : formatRating(dashboard.summary.averageRating)}</strong>
                    </span>
                    <span class="summary-pill">
                        <i class="ri-chat-quote-line" />
                        Total reviews: <strong>{dashboard.summary.totalReviews}</strong>
                    </span>
                    <span class="summary-pill">
                        <i class="ri-time-line" />
                        Last sync: <strong>{formatDateTime(dashboard.summary.lastSyncAt)}</strong>
                    </span>
                    <span class="summary-pill">
                        <i class="ri-google-line" />
                        Source: <strong>{hasGoogleSource ? "Google" : "Not connected"}</strong>
                    </span>
                    <span class={`label label-sm ${resolveStateLabelClass(dashboardState)}`}>{resolveStateLabel(dashboardState)}</span>
                </div>
            </section>

            <section class="panel reviews-filters-panel">
                <div class="reviews-filters-head">
                    <h5 class="m-0">Reviews list</h5>
                    <span class="txt-sm txt-hint">{filteredReviews.length} shown · {allReviews.length} total</span>
                </div>
                <div class="reviews-filter-grid">
                    <div class="control-item">
                        <label class="txt-sm txt-hint block m-b-5" for="reviews-source-filter">Source</label>
                        <select id="reviews-source-filter" class="input input-sm" bind:value={sourceFilter}>
                            <option value="all">All sources</option>
                            {#each sourceOptions as sourceOption}
                                <option value={sourceOption}>{resolveSourceLabel(sourceOption)}</option>
                            {/each}
                        </select>
                    </div>

                    <div class="control-item">
                        <label class="txt-sm txt-hint block m-b-5" for="reviews-rating-filter">Rating</label>
                        <select id="reviews-rating-filter" class="input input-sm" bind:value={ratingFilter}>
                            {#each ratingFilterOptions as option (option.value)}
                                <option value={option.value}>{option.label}</option>
                            {/each}
                        </select>
                    </div>

                    <div class="control-item reviews-search-control">
                        <label class="txt-sm txt-hint block m-b-5" for="reviews-search">Search</label>
                        <input
                            id="reviews-search"
                            type="text"
                            class="input input-sm"
                            placeholder="Search by author, review text, or source..."
                            bind:value={searchTerm}
                        />
                    </div>
                </div>
            </section>

            <section class="panel reviews-list-panel">
                {#if !allReviews.length}
                    <div class="empty-state">No synced reviews are available for this website yet.</div>
                {:else if !filteredReviews.length}
                    <div class="empty-state empty-state-stack">
                        <span>No reviews match these filters.</span>
                        <button type="button" class="btn btn-xs btn-outline" on:click={clearReviewsFilters}>
                            <span class="txt">Clear filters</span>
                        </button>
                    </div>
                {:else}
                    <div class="list list-compact">
                        <div class="list-content">
                            {#each filteredReviews as review (review.id)}
                                <article class="list-item review-row-item">
                                    <div class="content review-row-main">
                                        <div class="review-row-head">
                                            <h6 class="m-0 review-author">{review.authorName || "Google user"}</h6>
                                            <span class="label label-sm label-info">{resolveSourceLabel(review.source)}</span>
                                        </div>

                                        <div class="review-rating-row">
                                            <span class="review-stars" aria-label={`Rating ${formatRating(review.rating)} out of 5`}>
                                                {#each buildStarSlots(review.rating) as isFilled, index (`${review.id}_star_${index}`)}
                                                    <i class={isFilled ? "ri-star-fill" : "ri-star-line"} aria-hidden="true" />
                                                {/each}
                                            </span>
                                            <span class="txt-sm review-rating-value">{formatRating(review.rating)}</span>
                                        </div>

                                        <p class="txt-sm review-text">{truncateText(review.text, 260) || "No review text was provided."}</p>
                                        <div class="txt-xs txt-hint">{resolveReviewTimeline(review)}</div>
                                    </div>

                                    <div class="actions review-row-actions">
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-outline"
                                            disabled={!`${review.text || ""}`.trim()}
                                            on:click={() => copyText(review.text, "Review text")}
                                        >
                                            <i class="ri-file-copy-line" />
                                            <span class="txt">Copy text</span>
                                        </button>
                                        {#if review.reviewUrl}
                                            <button
                                                type="button"
                                                class="btn btn-xs btn-outline"
                                                on:click={() => openExternal(review.reviewUrl)}
                                            >
                                                <i class="ri-external-link-line" />
                                                <span class="txt">Open</span>
                                            </button>
                                        {/if}
                                    </div>
                                </article>
                            {/each}
                        </div>
                    </div>
                {/if}
            </section>
        {:else if activeTab === tabKeys.display}
            <section class="panel placeholder-tab-panel">
                <h5 class="m-0 m-b-xs">Display settings</h5>
                <p class="txt-sm txt-hint m-b-0">Display settings are not configured yet.</p>
                <p class="txt-sm txt-hint m-b-0">
                    Review display controls will be available once the public reviews widget is enabled.
                </p>
            </section>
        {:else}
            <section class="panel sources-tab-panel">
                <div class="sources-head">
                    <h5 class="m-0">Google source</h5>
                    <span class={`label label-sm ${resolveStateLabelClass(dashboardState)}`}>{resolveStateLabel(dashboardState)}</span>
                </div>

                <div class="sources-grid">
                    <div class="source-item">
                        <span class="source-label">Source</span>
                        <span class="source-value">{hasGoogleSource ? "Google Reviews" : "Not connected"}</span>
                    </div>
                    <div class="source-item">
                        <span class="source-label">Google Place ID</span>
                        <span class="source-value">{`${dashboard.googlePlaceId || ""}`.trim() || "Not configured"}</span>
                    </div>
                    <div class="source-item">
                        <span class="source-label">Review request link</span>
                        <span class="source-value source-value-break">{`${dashboard.reviewRequestLink || ""}`.trim() || "Not configured"}</span>
                    </div>
                    <div class="source-item">
                        <span class="source-label">Last sync</span>
                        <span class="source-value">{formatDateTime(dashboard.summary.lastSyncAt)}</span>
                    </div>
                </div>

                {#if !`${dashboard.googlePlaceId || ""}`.trim()}
                    <div class="empty-state m-t-sm">
                        Connect a Google Place ID in Website Settings to sync reviews.
                    </div>
                {/if}

                <div class="sources-actions">
                    <button
                        type="button"
                        class="btn btn-sm"
                        class:btn-loading={isSyncing}
                        disabled={!selectedWebsiteId || isSyncing || isLoadingDashboard || isLoadingWebsites}
                        on:click={() => syncFromGoogle()}
                    >
                        <i class="ri-loop-right-line" />
                        <span class="txt">Sync now</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!dashboard.openOnGoogle}
                        on:click={() => openExternal(dashboard.openOnGoogle)}
                    >
                        <i class="ri-external-link-line" />
                        <span class="txt">Open on Google</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!dashboard.reviewRequestLink}
                        on:click={() => copyText(dashboard.reviewRequestLink, "Review request link")}
                    >
                        <i class="ri-file-copy-line" />
                        <span class="txt">Copy review request link</span>
                    </button>
                </div>

                <section class="response-templates-panel m-t-sm">
                    <h6 class="m-0 m-b-xs">Response templates</h6>
                    <div class="templates-grid">
                        <article class="template-column">
                            <h6 class="m-b-xs">Positive reviews</h6>
                            {#each positiveResponseTemplates as template, index (`positive_${index}`)}
                                <div class="template-item">
                                    <p class="txt-sm m-0">{template}</p>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        on:click={() => copyText(template, "Template")}
                                    >
                                        <i class="ri-file-copy-line" />
                                        <span class="txt">Copy</span>
                                    </button>
                                </div>
                            {/each}
                        </article>
                        <article class="template-column">
                            <h6 class="m-b-xs">Needs-attention reviews</h6>
                            {#each negativeResponseTemplates as template, index (`negative_${index}`)}
                                <div class="template-item">
                                    <p class="txt-sm m-0">{template}</p>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        on:click={() => copyText(template, "Template")}
                                    >
                                        <i class="ri-file-copy-line" />
                                        <span class="txt">Copy</span>
                                    </button>
                                </div>
                            {/each}
                        </article>
                    </div>
                </section>
            </section>
        {/if}
    {/if}
</section>

<style>
    .reviews-dashboard-module {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .reviews-controls-row {
        display: flex;
        align-items: flex-end;
        gap: 10px;
        flex-wrap: wrap;
    }

    .reviews-selector-wrap {
        display: flex;
        flex-direction: column;
        gap: 5px;
        min-width: min(100%, 320px);
    }

    .reviews-controls-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .loading-state,
    .empty-state {
        border: 1px dashed var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 14px;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        color: var(--txtHintColor);
    }

    .empty-state-stack {
        flex-direction: column;
    }

    .summary-badges {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;
    }

    .reviews-filters-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
        margin-bottom: 8px;
    }

    .reviews-filter-grid {
        display: grid;
        grid-template-columns: minmax(160px, 220px) minmax(160px, 220px) minmax(260px, 1fr);
        gap: 8px;
    }

    .control-item,
    .reviews-search-control {
        min-width: 0;
    }

    .review-row-item {
        display: grid;
        grid-template-columns: minmax(0, 1fr) auto;
        align-items: flex-start;
        gap: 10px;
    }

    .review-row-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .review-row-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .review-author {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .review-rating-row {
        display: inline-flex;
        align-items: center;
        gap: 8px;
    }

    .review-stars {
        display: inline-flex;
        align-items: center;
        gap: 1px;
        color: color-mix(in srgb, var(--warningColor) 82%, var(--txtPrimaryColor));
        font-size: 14px;
    }

    .review-rating-value {
        color: var(--txtHintColor);
    }

    .review-text {
        margin: 0;
        white-space: pre-wrap;
        line-height: 1.4;
    }

    .review-row-actions {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .placeholder-tab-panel {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .sources-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
        margin-bottom: 8px;
    }

    .sources-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
        gap: 8px;
    }

    .source-item {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 9px 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .source-label {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        color: var(--txtHintColor);
    }

    .source-value {
        font-size: 13px;
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .source-value-break {
        white-space: normal;
        word-break: break-word;
    }

    .sources-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        margin-top: 10px;
    }

    .response-templates-panel {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .templates-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        gap: 10px;
    }

    .template-column {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 10px;
    }

    .template-item {
        display: flex;
        flex-direction: column;
        gap: 6px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 8px;
        margin-bottom: 8px;
        background: var(--baseColor);
    }

    .template-item:last-child {
        margin-bottom: 0;
    }

    @media (max-width: 900px) {
        .reviews-filter-grid {
            grid-template-columns: repeat(2, minmax(180px, 1fr));
        }

        .reviews-search-control {
            grid-column: 1 / -1;
        }

        .review-row-item {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 640px) {
        .reviews-filter-grid {
            grid-template-columns: 1fr;
        }

        .reviews-controls-actions,
        .sources-actions {
            width: 100%;
        }

        .reviews-controls-actions .btn,
        .sources-actions .btn {
            flex: 1 1 auto;
        }
    }
</style>
