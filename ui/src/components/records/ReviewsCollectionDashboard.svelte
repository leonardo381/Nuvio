<script>
    // NUVIO CUSTOM START: Collection-backed dashboard Reviews module UI.
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addErrorToast, addInfoToast, addSuccessToast } from "@/stores/toasts";

    export let websitesCollection = null;

    const tabKeys = {
        reviews: "reviews",
        request: "request",
        source: "source",
    };

    const ratingFilterOptions = [
        { value: "all", label: "All ratings" },
        { value: "5", label: "5 stars" },
        { value: "4plus", label: "4+ stars" },
        { value: "3plus", label: "3+ stars" },
    ];

    const periodFilterOptions = [
        { value: "all", label: "All time" },
        { value: "thisMonth", label: "This month" },
        { value: "thisYear", label: "This year" },
    ];

    const sortFilterOptions = [
        { value: "newest", label: "Newest" },
        { value: "ratingHigh", label: "Highest rating" },
        { value: "ratingLow", label: "Lowest rating" },
    ];

    const responseTemplates = [
        {
            key: "positive",
            title: "Positive response",
            text: "Thank you so much for your kind words. We are glad you had a great experience with our team.",
        },
        {
            key: "neutral",
            title: "Neutral response",
            text: "Thank you for sharing your feedback. We appreciate your time and will use your comments to keep improving.",
        },
        {
            key: "negative",
            title: "Negative response",
            text: "Thank you for the feedback. We are sorry your experience did not meet expectations and we would like to make it right.",
        },
    ];

    let websites = [];
    let selectedWebsiteId = "";
    let dashboard = null;
    let isLoadingWebsites = false;
    let isLoadingDashboard = false;
    let isSyncing = false;
    let syncError = "";

    let activeTab = tabKeys.reviews;
    let ratingFilter = "all";
    let periodFilter = "all";
    let sortFilter = "newest";
    let searchTerm = "";
    let selectedReviewId = "";

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
    $: dashboardSummary = dashboard?.summary || {
        averageRating: null,
        totalReviews: 0,
        lastSyncAt: "",
    };
    $: allReviews = Array.isArray(dashboard?.reviews) ? dashboard.reviews : [];
    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
    $: selectedWebsiteLabel = resolveWebsiteLabel(selectedWebsite);
    $: normalizedSearchTerm = normalizeLower(searchTerm);
    $: reviewRequestLink = normalizeString(dashboard?.reviewRequestLink);
    $: hasReviewRequestLink = !!reviewRequestLink;
    $: openOnGoogleLink = normalizeString(dashboard?.openOnGoogle);
    $: googlePlaceId = normalizeString(dashboard?.googlePlaceId);
    $: hasGoogleSource = allReviews.some((review) => normalizeLower(review?.source) === "google")
        || !!googlePlaceId
        || !!openOnGoogleLink;

    $: filteredReviews = allReviews.filter((review) => {
        if (!matchesRatingFilter(review?.rating, ratingFilter)) {
            return false;
        }

        if (!matchesPeriodFilter(review, periodFilter)) {
            return false;
        }

        if (!normalizedSearchTerm) {
            return true;
        }

        const haystack = [
            `${review?.authorName || ""}`,
            `${review?.text || ""}`,
            `${review?.source || ""}`,
        ].join(" ").toLowerCase();

        return haystack.includes(normalizedSearchTerm);
    });

    $: visibleReviews = sortReviews(filteredReviews, sortFilter);
    $: visibleReviewIds = new Set(visibleReviews.map((review) => normalizeString(review?.id)).filter(Boolean));

    $: if (activeTab === tabKeys.reviews) {
        if (!visibleReviews.length) {
            if (selectedReviewId) {
                selectedReviewId = "";
            }
        } else if (!visibleReviewIds.has(selectedReviewId)) {
            selectedReviewId = normalizeString(visibleReviews[0]?.id);
        }
    }

    $: selectedReview = allReviews.find((review) => normalizeString(review?.id) === selectedReviewId) || null;
    $: selectedReviewPhotoUrl = resolveReviewerPhotoUrl(selectedReview);
    $: selectedReviewText = normalizeString(selectedReview?.text);
    $: selectedReviewAuthorName = resolveReviewAuthorName(selectedReview);
    $: selectedReviewReviewUrl = normalizeString(selectedReview?.reviewUrl);
    $: selectedReviewSourceUrl = resolveReviewSourceUrl(selectedReview, dashboard);
    $: selectedReviewLinkForCopy = selectedReviewReviewUrl || selectedReviewSourceUrl;
    $: selectedReviewSyncedAt = selectedReview ? resolveReviewSyncedAt(selectedReview) : "";
    $: selectedReviewPublishedDisplay = selectedReview ? resolveReviewPublishedDisplay(selectedReview) : "";
    $: requestWhatsAppMessage = hasReviewRequestLink ? buildWhatsAppRequestMessage(selectedWebsiteLabel, reviewRequestLink) : "";
    $: requestEmailMessage = hasReviewRequestLink ? buildEmailRequestMessage(selectedWebsiteLabel, reviewRequestLink) : "";
    $: sourceSyncStatus = resolveSourceSyncStatus(dashboardState, dashboardSummary.lastSyncAt, isSyncing);
    $: sourceSetupGuidance = resolveSourceSetupGuidance(!!googlePlaceId, hasReviewRequestLink);

    export function reload() {
        if (!selectedWebsiteId) {
            return;
        }

        return loadDashboard();
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
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
            ApiClient.error(err);
            syncError = "Failed to sync reviews right now. Please try again.";
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

        return true;
    }

    function resolveReviewDate(review) {
        const publishedAt = normalizeString(review?.publishedAt);
        const syncedAt = normalizeString(review?.syncedAt);
        const source = publishedAt || syncedAt;
        if (!source) {
            return null;
        }

        const normalized = source.includes("T") ? source : source.replace(" ", "T");
        const parsed = new Date(normalized);
        if (Number.isNaN(parsed.getTime())) {
            return null;
        }

        return parsed;
    }

    function matchesPeriodFilter(review, filterValue) {
        if (filterValue === "all") {
            return true;
        }

        const reviewDate = resolveReviewDate(review);
        if (!reviewDate) {
            return false;
        }

        const now = new Date();

        if (filterValue === "thisMonth") {
            return reviewDate.getFullYear() === now.getFullYear() && reviewDate.getMonth() === now.getMonth();
        }

        if (filterValue === "thisYear") {
            return reviewDate.getFullYear() === now.getFullYear();
        }

        return true;
    }

    function sortReviews(reviews, filterValue) {
        const cloned = [...reviews];

        if (filterValue === "ratingHigh") {
            return cloned.sort((a, b) => resolveRatingNumber(b?.rating) - resolveRatingNumber(a?.rating));
        }

        if (filterValue === "ratingLow") {
            return cloned.sort((a, b) => resolveRatingNumber(a?.rating) - resolveRatingNumber(b?.rating));
        }

        return cloned.sort((a, b) => resolveTimestamp(b) - resolveTimestamp(a));
    }

    function resolveTimestamp(review) {
        const reviewDate = resolveReviewDate(review);
        if (!reviewDate) {
            return 0;
        }

        return reviewDate.getTime();
    }

    function resolveRatingNumber(value) {
        const numeric = Number(value);
        if (Number.isNaN(numeric)) {
            return 0;
        }

        return numeric;
    }

    function buildStarSlots(value) {
        const numeric = Number(value);
        const safeRating = Number.isNaN(numeric) ? 0 : Math.max(0, Math.min(5, numeric));

        return [1, 2, 3, 4, 5].map((step) => safeRating >= step - 0.25);
    }

    function truncateText(value, max = 240) {
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

        if (normalizeString(review?.relativeTime)) {
            parts.push(normalizeString(review.relativeTime));
        }

        if (normalizeString(review?.publishedAt)) {
            parts.push(formatDateTime(review.publishedAt));
        } else if (normalizeString(review?.syncedAt)) {
            parts.push(`Synced ${formatDateTime(review.syncedAt)}`);
        }

        return parts.join(" - ") || "No date information";
    }

    function resolveReviewPublishedDisplay(review) {
        if (!review) {
            return "";
        }

        const relativeTime = normalizeString(review?.relativeTime);
        const publishedAt = normalizeString(review?.publishedAt);

        if (relativeTime && publishedAt) {
            return `${relativeTime} - ${formatDateTime(publishedAt)}`;
        }

        if (relativeTime) {
            return relativeTime;
        }

        if (publishedAt) {
            return formatDateTime(publishedAt);
        }

        return "No published date";
    }

    function resolveReviewSyncedAt(review) {
        const syncedAt = normalizeString(review?.syncedAt);
        if (!syncedAt) {
            return "";
        }

        return formatDateTime(syncedAt);
    }

    function resolveReviewAuthorName(review) {
        return normalizeString(review?.authorName) || "Google user";
    }

    function resolveReviewerPhotoUrl(review) {
        if (!review || typeof review !== "object") {
            return "";
        }

        const reviewer = review?.reviewer && typeof review.reviewer === "object"
            ? review.reviewer
            : {};

        const candidates = [
            review?.reviewerPhotoUrl,
            review?.profilePhotoUrl,
            review?.authorPhotoUrl,
            review?.authorAvatarUrl,
            review?.avatarUrl,
            review?.photoUrl,
            reviewer?.photoUrl,
            reviewer?.profilePhotoUrl,
            reviewer?.avatarUrl,
        ];

        for (const candidate of candidates) {
            const value = normalizeString(candidate);
            if (value) {
                return value;
            }
        }

        return "";
    }

    function resolveReviewSourceUrl(review, currentDashboard) {
        const candidates = [currentDashboard?.openOnGoogle, review?.sourceUrl, review?.authorUrl];
        for (const candidate of candidates) {
            const value = normalizeString(candidate);
            if (value) {
                return value;
            }
        }
        return "";
    }

    function clearReviewsFilters() {
        ratingFilter = "all";
        periodFilter = "all";
        sortFilter = "newest";
        searchTerm = "";
    }

    function selectReview(review) {
        const reviewId = normalizeString(review?.id);
        if (!reviewId) {
            return;
        }

        selectedReviewId = reviewId;
    }

    function handleReviewCardKeyDown(event, review) {
        if (event.target !== event.currentTarget) {
            return;
        }

        if (event.key !== "Enter" && event.key !== " ") {
            return;
        }

        event.preventDefault();
        selectReview(review);
    }

    function buildWhatsAppRequestMessage(websiteLabel, reviewLink) {
        const safeBusinessName = normalizeString(websiteLabel) || "us";
        return `Hi, thank you for choosing ${safeBusinessName}. If you have a moment, could you leave us a review here? ${reviewLink}`;
    }

    function buildEmailRequestMessage(websiteLabel, reviewLink) {
        const safeBusinessName = normalizeString(websiteLabel) || "us";

        return [
            `Thank you for choosing ${safeBusinessName}.`,
            "",
            "We would really appreciate your feedback.",
            "You can leave us a review here:",
            reviewLink,
        ].join("\n");
    }

    function resolveSourceSyncStatus(state, lastSyncAt, syncing) {
        if (syncing) {
            return "Syncing now...";
        }

        if (state === "feature_unavailable") {
            return "Feature unavailable";
        }

        if (state === "disabled") {
            return "Feature disabled";
        }

        if (state === "not_configured") {
            return "Configuration required";
        }

        if (state === "never_synced") {
            return "Never synced";
        }

        if (!normalizeString(lastSyncAt)) {
            return "Sync status unavailable";
        }

        return `Last synced ${formatDateTime(lastSyncAt)}`;
    }

    function resolveSourceSetupGuidance(hasPlaceId, hasLink) {
        const guidance = [];

        if (!hasPlaceId) {
            guidance.push("Add a Google Place ID in Website Settings to sync Google reviews.");
        }

        if (!hasLink) {
            guidance.push("Add a review request link in Website Settings to make requesting reviews easier.");
        }

        if (hasPlaceId && hasLink) {
            guidance.push("Google reviews are ready to sync.");
        }

        return guidance;
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
    <section class="reviews-head operations-head panel m-b-base">
        <div class="head-main">
            <div class="summary-title-wrap">
                <h2 class="m-0">Reviews</h2>
                <p class="txt-sm txt-hint m-b-0 head-description">
                    Monitor customer reviews and reputation for this website.
                </p>
            </div>

            <div class="head-selector">
                <div class="selector-row">
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
            </div>
        </div>

        <div class="head-tools">
            <div class="reviews-header-actions">
                <button
                    type="button"
                    class="btn btn-sm"
                    class:btn-loading={isSyncing}
                    disabled={!selectedWebsiteId || isSyncing || isLoadingDashboard || isLoadingWebsites}
                    on:click={syncFromGoogle}
                >
                    <i class="ri-loop-right-line" />
                    <span class="txt">Sync now</span>
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline"
                    disabled={!reviewRequestLink}
                    on:click={() => copyText(reviewRequestLink, "Review link")}
                >
                    <i class="ri-file-copy-line" />
                    <span class="txt">Copy review link</span>
                </button>
            </div>

            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-star-smile-line" />
                    Average rating: <strong>{dashboardSummary.averageRating === null ? "-" : formatRating(dashboardSummary.averageRating)}</strong>
                </span>
                <span class="summary-pill">
                    <i class="ri-chat-quote-line" />
                    Total reviews: <strong>{dashboardSummary.totalReviews}</strong>
                </span>
                <span class="summary-pill">
                    <i class="ri-time-line" />
                    Last sync: <strong>{formatDateTime(dashboardSummary.lastSyncAt)}</strong>
                </span>
                <span class="summary-pill">
                    <i class="ri-link" />
                    Source status: <strong>{resolveStateLabel(dashboardState)}</strong>
                </span>
            </div>
        </div>
    </section>

    {#if isLoadingWebsites}
        <div class="loading-state m-b-base">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading websites...</span>
        </div>
    {:else if !websites.length}
        <div class="empty-state m-b-base">No websites are available yet.</div>
    {:else if isLoadingDashboard}
        <div class="loading-state m-b-base">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading reviews...</span>
        </div>
    {:else if !dashboard}
        <div class="empty-state m-b-base">Unable to load reviews for this website right now.</div>
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

        <div class="tabs-header compact combined left operations-tabs operations-tabs--nested m-b-sm">
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
                class:active={activeTab === tabKeys.request}
                on:click={() => (activeTab = tabKeys.request)}
            >
                <i class="ri-share-forward-line tab-icon" aria-hidden="true" />
                <span class="tab-label">Request Reviews</span>
            </button>
            <button
                type="button"
                class="tab-item"
                class:active={activeTab === tabKeys.source}
                on:click={() => (activeTab = tabKeys.source)}
            >
                <i class="ri-links-line tab-icon" aria-hidden="true" />
                <span class="tab-label">Source</span>
            </button>
        </div>

        {#if activeTab === tabKeys.reviews}
            <section class="panel reviews-workspace-panel m-b-base">
                <div class="reviews-workspace-layout">
                    <div class="reviews-left-column">
                        <div class="section-head section-head-inline m-b-sm">
                            <h4 class="m-0">Reviews</h4>
                            <span class="txt-sm txt-hint">{visibleReviews.length} shown - {allReviews.length} total</span>
                        </div>

                        <div class="reviews-filter-grid m-b-sm">
                            <div class="control-item reviews-filter-search">
                                <label class="txt-sm txt-hint block m-b-5" for="reviews-search">Search</label>
                                <input
                                    id="reviews-search"
                                    type="text"
                                    class="input input-sm"
                                    placeholder="Search by author, text, or source..."
                                    bind:value={searchTerm}
                                />
                            </div>
                            <div class="control-item">
                                <label class="txt-sm txt-hint block m-b-5" for="reviews-rating-filter">Rating</label>
                                <select id="reviews-rating-filter" class="input input-sm" bind:value={ratingFilter}>
                                    {#each ratingFilterOptions as option (option.value)}
                                        <option value={option.value}>{option.label}</option>
                                    {/each}
                                </select>
                            </div>
                            <div class="control-item">
                                <label class="txt-sm txt-hint block m-b-5" for="reviews-period-filter">Period</label>
                                <select id="reviews-period-filter" class="input input-sm" bind:value={periodFilter}>
                                    {#each periodFilterOptions as option (option.value)}
                                        <option value={option.value}>{option.label}</option>
                                    {/each}
                                </select>
                            </div>
                            <div class="control-item">
                                <label class="txt-sm txt-hint block m-b-5" for="reviews-sort-filter">Sort</label>
                                <select id="reviews-sort-filter" class="input input-sm" bind:value={sortFilter}>
                                    {#each sortFilterOptions as option (option.value)}
                                        <option value={option.value}>{option.label}</option>
                                    {/each}
                                </select>
                            </div>
                        </div>

                        <div class="reviews-filter-actions m-b-sm">
                            <button type="button" class="btn btn-sm btn-outline" on:click={clearReviewsFilters}>
                                <i class="ri-filter-off-line" />
                                <span class="txt">Reset filters</span>
                            </button>
                        </div>

                        <div class="reviews-list-wrap">
                            {#if !allReviews.length}
                                <div class="empty-state">Reviews synced from Google will appear here.</div>
                            {:else if !visibleReviews.length}
                                <div class="empty-state">
                                    No reviews match these filters.
                                </div>
                            {:else}
                                <div class="list list-compact">
                                    <div class="list-content">
                                        {#each visibleReviews as review (review.id)}
                                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                                            <article
                                                class="list-item review-list-item"
                                                class:selected={selectedReviewId === normalizeString(review.id)}
                                                role="button"
                                                tabindex="0"
                                                aria-label={`Select review from ${resolveReviewAuthorName(review)}`}
                                                on:click={() => selectReview(review)}
                                                on:keydown={(event) => handleReviewCardKeyDown(event, review)}
                                            >
                                                <div class="content review-list-main">
                                                    <div class="review-list-head">
                                                        <h6 class="m-0 review-author">{resolveReviewAuthorName(review)}</h6>
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

                                                    <p class="txt-sm review-text">{truncateText(review.text, 220) || "No review text was provided."}</p>
                                                    <div class="txt-xs txt-hint">{resolveReviewTimeline(review)}</div>
                                                    {#if resolveReviewSyncedAt(review)}
                                                        <div class="txt-xs txt-hint review-synced-meta">Synced: {resolveReviewSyncedAt(review)}</div>
                                                    {/if}
                                                </div>
                                            </article>
                                        {/each}
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>

                    <aside class="reviews-right-rail" aria-live="polite">
                        {#if selectedReview}
                            <section class="reviews-rail-block">
                                <div class="reviews-rail-head">
                                    <div class="reviews-rail-head-main">
                                        <h5 class="m-0">Review summary</h5>
                                        <p class="txt-sm txt-hint m-b-0 reviews-rail-helper">
                                            Review context and source details for follow-up.
                                        </p>
                                    </div>
                                </div>

                                <div class="reviews-detail-badges">
                                    <span class="label label-sm label-info">{resolveSourceLabel(selectedReview.source)}</span>
                                    <span class="label label-sm label-warning">Rating {formatRating(selectedReview.rating)}</span>
                                </div>

                                <div class="reviews-summary-stack">
                                    <div class="reviews-summary-item">
                                        <span class="txt-xs txt-hint">Author</span>
                                        <span class="txt-sm">{selectedReviewAuthorName}</span>
                                    </div>

                                    <div class="reviews-summary-item">
                                        <span class="txt-xs txt-hint">Source</span>
                                        <span class="txt-sm">{resolveSourceLabel(selectedReview.source)}</span>
                                    </div>

                                    <div class="reviews-summary-item">
                                        <span class="txt-xs txt-hint">Rating</span>
                                        <div class="review-rating-row">
                                            <span class="review-stars" aria-label={`Rating ${formatRating(selectedReview.rating)} out of 5`}>
                                                {#each buildStarSlots(selectedReview.rating) as isFilled, index (`${selectedReview.id}_detail_star_${index}`)}
                                                    <i class={isFilled ? "ri-star-fill" : "ri-star-line"} aria-hidden="true" />
                                                {/each}
                                            </span>
                                            <span class="txt-sm review-rating-value">{formatRating(selectedReview.rating)}</span>
                                        </div>
                                    </div>

                                    {#if selectedReviewPhotoUrl}
                                        <div class="reviews-summary-item">
                                            <span class="txt-xs txt-hint">Reviewer photo</span>
                                            <div class="reviewer-photo-wrap">
                                                <img src={selectedReviewPhotoUrl} alt={`Photo of ${selectedReviewAuthorName}`} loading="lazy" />
                                            </div>
                                        </div>
                                    {/if}

                                    <div class="reviews-summary-item">
                                        <span class="txt-xs txt-hint">Review text</span>
                                        <p class="txt-sm m-b-0 review-detail-text">{selectedReviewText || "No review text was provided."}</p>
                                    </div>

                                    <div class="reviews-summary-item">
                                        <span class="txt-xs txt-hint">Published / relative</span>
                                        <span class="txt-sm">{selectedReviewPublishedDisplay}</span>
                                    </div>

                                    {#if selectedReviewSyncedAt}
                                        <div class="reviews-summary-item">
                                            <span class="txt-xs txt-hint">Synced at</span>
                                            <span class="txt-sm">{selectedReviewSyncedAt}</span>
                                        </div>
                                    {/if}
                                </div>
                            </section>

                            <section class="reviews-rail-block">
                                <div class="reviews-rail-head">
                                    <div class="reviews-rail-head-main">
                                        <h5 class="m-0">Actions</h5>
                                        <p class="txt-sm txt-hint m-b-0 reviews-rail-helper">
                                            Copy or open this review for reputation workflows.
                                        </p>
                                    </div>
                                </div>

                                <div class="reviews-detail-actions">
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-outline"
                                        disabled={!selectedReviewText}
                                        on:click={() => copyText(selectedReviewText, "Review text")}
                                    >
                                        <i class="ri-file-copy-line" />
                                        <span class="txt">Copy review text</span>
                                    </button>

                                    {#if selectedReviewReviewUrl}
                                        <button
                                            type="button"
                                            class="btn btn-sm btn-outline"
                                            on:click={() => openExternal(selectedReviewReviewUrl)}
                                        >
                                            <i class="ri-external-link-line" />
                                            <span class="txt">Open review URL</span>
                                        </button>
                                    {/if}

                                    {#if selectedReviewSourceUrl}
                                        <button
                                            type="button"
                                            class="btn btn-sm btn-outline"
                                            on:click={() => openExternal(selectedReviewSourceUrl)}
                                        >
                                            <i class="ri-google-line" />
                                            <span class="txt">Open on Google</span>
                                        </button>
                                    {/if}

                                    {#if selectedReviewLinkForCopy}
                                        <button
                                            type="button"
                                            class="btn btn-sm btn-outline"
                                            on:click={() => copyText(selectedReviewLinkForCopy, "Review link")}
                                        >
                                            <i class="ri-link" />
                                            <span class="txt">Copy review link</span>
                                        </button>
                                    {/if}
                                </div>
                            </section>

                            <section class="reviews-rail-block">
                                <div class="reviews-rail-head">
                                    <div class="reviews-rail-head-main">
                                        <h5 class="m-0">Response templates</h5>
                                        <p class="txt-sm txt-hint m-b-0 reviews-rail-helper">
                                            Use as starting points when responding to customers.
                                        </p>
                                    </div>
                                </div>

                                <div class="response-template-columns">
                                    {#each responseTemplates as template (template.key)}
                                        <article class="template-column">
                                            <h6 class="m-0 m-b-xs">{template.title}</h6>
                                            <div class="template-item">
                                                <p class="txt-sm m-0">{template.text}</p>
                                                <button
                                                    type="button"
                                                    class="btn btn-xs btn-outline"
                                                    on:click={() => copyText(template.text, "Response template")}
                                                >
                                                    <i class="ri-file-copy-line" />
                                                    <span class="txt">Copy response</span>
                                                </button>
                                            </div>
                                        </article>
                                    {/each}
                                </div>
                            </section>
                        {:else}
                            <div class="empty-state">Select a review to view details.</div>
                        {/if}
                    </aside>
                </div>
            </section>
        {:else if activeTab === tabKeys.request}
            <section class="panel reviews-request-panel m-b-base">
                <div class="request-head m-b-sm">
                    <h4 class="m-0">Request reviews</h4>
                    <p class="txt-sm txt-hint m-b-0">
                        Share your Google review link and reuse ready-to-send messages.
                    </p>
                </div>

                <div class="request-link-row m-b-sm">
                    <div class="request-link-main">
                        <div class="txt-xs txt-hint txt-uppercase">Review link</div>
                        <div class="request-link-value">{reviewRequestLink || "Not configured"}</div>
                        {#if !hasReviewRequestLink}
                            <p class="txt-sm txt-hint m-b-0 request-link-missing">
                                Add a review request link in Website Settings to make requesting reviews easier.
                            </p>
                        {/if}
                    </div>
                    <div class="request-link-actions">
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={!reviewRequestLink}
                            on:click={() => copyText(reviewRequestLink, "Review request link")}
                        >
                            <i class="ri-file-copy-line" />
                            <span class="txt">Copy link</span>
                        </button>
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={!reviewRequestLink}
                            on:click={() => openExternal(reviewRequestLink)}
                        >
                            <i class="ri-external-link-line" />
                            <span class="txt">Open link</span>
                        </button>
                    </div>
                </div>

                <div class="request-messages-grid">
                    <article class="request-message-card">
                        <div class="txt-xs txt-hint txt-uppercase m-b-5">WhatsApp request message</div>
                        <textarea
                            class="input request-message-area"
                            rows="5"
                            readonly
                            disabled={!hasReviewRequestLink}
                            value={requestWhatsAppMessage || "Add a review request link to generate a WhatsApp request message."}
                        />
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={!hasReviewRequestLink}
                            on:click={() => copyText(requestWhatsAppMessage, "WhatsApp message")}
                        >
                            <i class="ri-file-copy-line" />
                            <span class="txt">Copy WhatsApp message</span>
                        </button>
                    </article>

                    <article class="request-message-card">
                        <div class="txt-xs txt-hint txt-uppercase m-b-5">Email request message</div>
                        <textarea
                            class="input request-message-area"
                            rows="7"
                            readonly
                            disabled={!hasReviewRequestLink}
                            value={requestEmailMessage || "Add a review request link to generate an email request message."}
                        />
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={!hasReviewRequestLink}
                            on:click={() => copyText(requestEmailMessage, "Email message")}
                        >
                            <i class="ri-file-copy-line" />
                            <span class="txt">Copy email message</span>
                        </button>
                    </article>
                </div>

                <p class="txt-xs txt-hint m-b-0 request-toolkit-note">
                    QR code support can be added later.
                </p>
            </section>
        {:else}
            <section class="panel reviews-source-panel m-b-base">
                <div class="source-overview-card m-b-sm">
                    <div class="source-head m-b-sm">
                        <div class="source-head-main">
                            <h4 class="m-0">Google source</h4>
                            <p class="txt-sm txt-hint m-b-0">
                                Review connection and sync health for this website.
                            </p>
                        </div>
                        <span class={`label label-sm ${resolveStateLabelClass(dashboardState)}`}>{resolveStateLabel(dashboardState)}</span>
                    </div>

                    <div class="summary-badges source-summary-badges m-b-sm">
                        <span class={`label label-sm ${resolveStateLabelClass(dashboardState)}`}>{resolveStateLabel(dashboardState)}</span>
                        <span class="summary-pill">
                            <i class="ri-loop-right-line" />
                            {sourceSyncStatus}
                        </span>
                        <span class="summary-pill">
                            <i class="ri-time-line" />
                            Last sync: {formatDateTime(dashboardSummary.lastSyncAt)}
                        </span>
                    </div>

                    <div class="source-status-grid">
                        <div class="source-status-item">
                            <span class="source-status-key">Source</span>
                            <span class="source-status-value">Google</span>
                        </div>
                        <div class="source-status-item">
                            <span class="source-status-key">Feature status</span>
                            <span class="source-status-value">{resolveStateLabel(dashboardState)}</span>
                        </div>
                        <div class="source-status-item">
                            <span class="source-status-key">Google Place ID</span>
                            <span class="source-status-value">{googlePlaceId ? "Configured" : "Missing"}</span>
                        </div>
                        <div class="source-status-item">
                            <span class="source-status-key">Review request link</span>
                            <span class="source-status-value">{hasReviewRequestLink ? "Configured" : "Missing"}</span>
                        </div>
                        <div class="source-status-item">
                            <span class="source-status-key">Total synced reviews</span>
                            <span class="source-status-value">{dashboardSummary.totalReviews}</span>
                        </div>
                        <div class="source-status-item">
                            <span class="source-status-key">Average rating</span>
                            <span class="source-status-value">{dashboardSummary.averageRating === null ? "Not available" : formatRating(dashboardSummary.averageRating)}</span>
                        </div>
                    </div>
                </div>

                <div class="source-actions m-b-sm">
                    <button
                        type="button"
                        class="btn btn-sm"
                        class:btn-loading={isSyncing}
                        disabled={!selectedWebsiteId || isSyncing || isLoadingDashboard || isLoadingWebsites}
                        on:click={syncFromGoogle}
                    >
                        <i class="ri-loop-right-line" />
                        <span class="txt">{isSyncing ? "Syncing..." : "Sync now"}</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!hasReviewRequestLink}
                        on:click={() => openExternal(reviewRequestLink)}
                    >
                        <i class="ri-external-link-line" />
                        <span class="txt">Open review link</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!hasReviewRequestLink}
                        on:click={() => copyText(reviewRequestLink, "Review request link")}
                    >
                        <i class="ri-file-copy-line" />
                        <span class="txt">Copy review request link</span>
                    </button>
                </div>

                <section class="source-guidance-section">
                    <div class="txt-xs txt-hint txt-uppercase m-b-5">Setup guidance</div>
                    <div class="source-guidance-list">
                        {#each sourceSetupGuidance as guidanceMessage, guidanceIndex (`source_guidance_${guidanceIndex}`)}
                            <div class="source-guidance-item">
                                <span class="label label-sm source-guidance-pill">{guidanceMessage === "Google reviews are ready to sync." ? "Ready" : "Info"}</span>
                                <span class="source-guidance-text">{guidanceMessage}</span>
                            </div>
                        {/each}
                    </div>
                </section>
            </section>
        {/if}
    {/if}
</section>

<style>
    .reviews-head.operations-head .head-description {
        max-width: 520px;
    }

    .reviews-head.operations-head .head-selector {
        width: min(100%, 560px);
    }

    .reviews-head.operations-head .summary-badges {
        justify-content: flex-end;
    }

    .reviews-dashboard-module {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .head-tools {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .reviews-header-actions {
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

    .reviews-workspace-panel {
        padding: calc(var(--baseSpacing) - 10px);
    }

    .reviews-workspace-layout {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
        gap: 12px;
        align-items: start;
    }

    .reviews-left-column {
        min-width: 0;
    }

    .reviews-right-rail {
        display: flex;
        flex-direction: column;
        gap: 10px;
        min-width: 0;
    }

    .reviews-rail-block {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .reviews-rail-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
    }

    .reviews-rail-head-main {
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .reviews-rail-helper {
        font-size: 11px;
        line-height: 1.35;
    }

    .reviews-detail-badges {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .reviews-summary-stack {
        display: flex;
        flex-direction: column;
        gap: 7px;
    }

    .reviews-summary-item {
        display: flex;
        flex-direction: column;
        gap: 3px;
        padding-top: 7px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .reviews-summary-item:first-child {
        border-top: 0;
        padding-top: 0;
    }

    .review-detail-text {
        white-space: pre-wrap;
    }

    .reviewer-photo-wrap {
        display: inline-flex;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        overflow: hidden;
        width: 56px;
        height: 56px;
    }

    .reviewer-photo-wrap img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }

    .reviews-detail-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .response-template-columns {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: 8px;
    }

    .template-column {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 8px;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .template-item {
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
        padding-top: 6px;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .template-column .template-item:first-of-type {
        border-top: 0;
        padding-top: 0;
    }

    .reviews-filter-grid {
        display: grid;
        grid-template-columns: minmax(220px, 1.6fr) repeat(3, minmax(140px, 1fr));
        gap: 8px;
    }

    .control-item,
    .reviews-filter-search {
        min-width: 0;
    }

    .reviews-filter-actions {
        display: flex;
        align-items: center;
        justify-content: flex-end;
    }

    .reviews-list-wrap {
        min-height: 120px;
    }

    .review-list-item {
        cursor: pointer;
        transition: border-color var(--baseAnimationSpeed), box-shadow var(--baseAnimationSpeed);
    }

    .review-list-item:hover,
    .review-list-item:focus-visible {
        border-color: var(--baseAlt3Color);
        box-shadow: 0 0 0 2px var(--baseAltColor);
        outline: none;
    }

    .review-list-item.selected {
        border-color: var(--baseAlt3Color);
        box-shadow: 0 0 0 2px var(--baseAltColor);
    }

    .review-list-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .review-list-head {
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

    .reviews-request-panel,
    .reviews-source-panel {
        padding: calc(var(--baseSpacing) - 10px);
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .request-link-row {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        padding: 10px;
        background: var(--baseColor);
    }

    .request-link-main {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
        flex: 1 1 300px;
    }

    .request-link-value {
        font-size: 13px;
        color: var(--txtPrimaryColor);
        overflow-wrap: anywhere;
    }

    .request-link-missing {
        line-height: 1.35;
    }

    .request-link-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .request-messages-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
        gap: 10px;
    }

    .request-message-card {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .request-message-area {
        resize: vertical;
        min-height: 92px;
        line-height: 1.4;
    }

    .request-toolkit-note {
        margin-top: 2px;
    }

    .source-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .source-head-main {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .source-overview-card {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
    }

    .source-summary-badges {
        align-items: center;
    }

    .source-status-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 8px;
    }

    .source-status-item {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 9px 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .source-status-key {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        color: var(--txtHintColor);
    }

    .source-status-value {
        font-size: 13px;
        color: var(--txtPrimaryColor);
        overflow-wrap: anywhere;
    }

    .source-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .source-guidance-section {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .source-guidance-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .source-guidance-item {
        display: flex;
        align-items: flex-start;
        gap: 8px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
        padding-top: 6px;
    }

    .source-guidance-item:first-child {
        border-top: 0;
        padding-top: 0;
    }

    .source-guidance-pill {
        flex: 0 0 auto;
    }

    .source-guidance-text {
        font-size: 13px;
        color: var(--txtPrimaryColor);
        line-height: 1.35;
    }

    @media (max-width: 1080px) {
        .reviews-workspace-layout {
            grid-template-columns: 1fr;
        }

        .reviews-right-rail {
            order: 2;
        }

        .reviews-left-column {
            order: 1;
        }
    }

    @media (max-width: 900px) {
        .reviews-filter-grid {
            grid-template-columns: repeat(2, minmax(150px, 1fr));
        }

        .reviews-filter-search {
            grid-column: 1 / -1;
        }

        .reviews-header-actions,
        .source-actions,
        .request-link-actions {
            width: 100%;
        }

        .reviews-header-actions .btn,
        .source-actions .btn,
        .request-link-actions .btn {
            flex: 1 1 auto;
        }
    }

    @media (max-width: 640px) {
        .reviews-filter-grid {
            grid-template-columns: 1fr;
        }

        .head-tools {
            flex-direction: column;
            align-items: stretch;
        }

        .reviews-head.operations-head .summary-badges {
            justify-content: flex-start;
        }

        .reviews-filter-actions {
            justify-content: stretch;
        }

        .reviews-filter-actions .btn {
            width: 100%;
        }
    }
</style>
