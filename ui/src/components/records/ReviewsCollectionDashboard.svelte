<script>
    // NUVIO CUSTOM START: Collection-backed dashboard Reviews module UI.
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addErrorToast, addInfoToast, addSuccessToast } from "@/stores/toasts";

    export let websitesCollection = null;

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

    export function reload() {
        if (!selectedWebsiteId) {
            return;
        }

        return loadDashboard();
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
            return "Reviews is unavailable for this website based on feature flags.";
        }

        if (state === "disabled") {
            return "Reviews is available but disabled in website settings.";
        }

        if (state === "not_configured") {
            return "Reviews is enabled but missing Google Place ID configuration.";
        }

        if (state === "never_synced") {
            return "This website has no synced reviews yet. Click Refresh reviews to import from Google.";
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

    function buildRatingStars(value) {
        const numeric = Number(value);
        if (Number.isNaN(numeric)) {
            return "";
        }

        const fullStars = Math.max(0, Math.min(5, Math.round(numeric)));
        return "*".repeat(fullStars) + "-".repeat(5 - fullStars);
    }

    function truncateText(value, max = 260) {
        return CommonHelper.truncate(`${value || ""}`.trim(), max, true);
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
    <div class="section-head">
        <h4 class="m-0">Reviews</h4>
        <p class="txt-sm txt-hint m-0">Collection-backed Google reputation module.</p>
    </div>

    <div class="toolbar-row">
        <label class="txt-sm txt-hint" for="reviews-website">Website</label>
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

        <div class="flex-fill" />

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
            class="btn btn-sm btn-secondary"
            class:btn-loading={isSyncing}
            disabled={!selectedWebsiteId || isSyncing || isLoadingDashboard || isLoadingWebsites}
            on:click={() => syncFromGoogle()}
        >
            <i class="ri-loop-right-line" />
            <span class="txt">Refresh reviews</span>
        </button>
    </div>

    {#if isLoadingWebsites}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading websites...</span>
        </div>
    {:else if !websites.length}
        <div class="empty-state">No website records found.</div>
    {:else if isLoadingDashboard}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading Reviews dashboard...</span>
        </div>
    {:else if dashboard}
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

        <div class="summary-grid">
            <article class="summary-card">
                <h6 class="m-0 txt-hint">Average rating</h6>
                <div class="summary-value">{dashboard.summary.averageRating === null ? "-" : formatRating(dashboard.summary.averageRating)}</div>
                {#if dashboard.summary.averageRating !== null}
                    <div class="txt-sm txt-hint">{buildRatingStars(dashboard.summary.averageRating)}</div>
                {/if}
            </article>

            <article class="summary-card">
                <h6 class="m-0 txt-hint">Total synced reviews</h6>
                <div class="summary-value">{dashboard.summary.totalReviews}</div>
                <div class="txt-sm txt-hint">Stored in `Reviews` collection</div>
            </article>

            <article class="summary-card">
                <h6 class="m-0 txt-hint">Last sync</h6>
                <div class="summary-value summary-small">{formatDateTime(dashboard.summary.lastSyncAt)}</div>
                {#if dashboard.syncedCount}
                    <div class="txt-sm txt-hint">Updated {dashboard.syncedCount} record(s) in last refresh</div>
                {/if}
            </article>
        </div>

        <div class="helper-actions">
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

        <p class="txt-sm txt-hint m-t-xs m-b-sm">
            Ask for reviews right after positive interactions and share the copied review request link directly in WhatsApp or email.
        </p>

        <section class="templates-section">
            <h5 class="m-b-xs">Response templates</h5>

            <div class="templates-grid">
                <article class="template-column">
                    <h6 class="m-b-xs">Positive reviews</h6>
                    {#each positiveResponseTemplates as template, index (`positive_${index}`)}
                        <div class="template-item">
                            <p class="txt-sm m-0">{template}</p>
                            <button
                                type="button"
                                class="btn btn-xs btn-transparent"
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
                                class="btn btn-xs btn-transparent"
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

        <section class="reviews-list-section">
            <h5 class="m-b-xs">Recent reviews</h5>

            {#if !dashboard.reviews?.length}
                <div class="empty-state">No synced reviews yet for this website.</div>
            {:else}
                <div class="reviews-grid">
                    {#each dashboard.reviews as review (review.id)}
                        <article class="review-card">
                            <div class="review-head">
                                <h6 class="m-0">{review.authorName || "Google user"}</h6>
                                <div class="rating-chip" title="Rating">
                                    <span>{formatRating(review.rating)}</span>
                                    <span class="stars">{buildRatingStars(review.rating)}</span>
                                </div>
                            </div>

                            <div class="txt-sm txt-hint m-b-xs">
                                {review.relativeTime || "No relative time"}
                                {#if review.publishedAt}
                                    | {formatDateTime(review.publishedAt)}
                                {/if}
                            </div>

                            <p class="txt-sm review-text">{truncateText(review.text, 260) || "No review text."}</p>

                            <div class="review-actions">
                                <button
                                    type="button"
                                    class="btn btn-xs btn-transparent"
                                    on:click={() => copyText(review.text, "Review text")}
                                >
                                    <i class="ri-file-copy-line" />
                                    <span class="txt">Copy text</span>
                                </button>

                                {#if review.reviewUrl}
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-transparent"
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
            {/if}
        </section>
    {/if}
</section>

<style>
    .reviews-dashboard-module {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .section-head {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
        gap: 8px;
    }

    .toolbar-row {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .toolbar-row select {
        min-width: 240px;
    }

    .loading-state,
    .empty-state {
        border: 1px dashed var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        color: var(--txtHintColor);
    }

    .summary-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: 10px;
    }

    .summary-card {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 12px;
        display: flex;
        flex-direction: column;
        gap: 6px;
        min-height: 92px;
    }

    .summary-value {
        font-size: 1.3rem;
        font-weight: 700;
        line-height: 1.1;
    }

    .summary-small {
        font-size: 1rem;
    }

    .helper-actions {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;
    }

    .templates-section,
    .reviews-list-section {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 12px;
        background: var(--baseColor);
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

    .reviews-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        gap: 10px;
    }

    .review-card {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .review-head {
        display: flex;
        justify-content: space-between;
        gap: 8px;
    }

    .rating-chip {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: var(--smFontSize);
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        padding: 2px 8px;
        white-space: nowrap;
        background: var(--baseColor);
    }

    .rating-chip .stars {
        color: var(--warningColor);
        letter-spacing: 0.5px;
    }

    .review-text {
        white-space: pre-wrap;
        line-height: 1.35;
        margin: 0;
    }

    .review-actions {
        display: flex;
        justify-content: flex-end;
        gap: 4px;
        margin-top: auto;
    }
</style>
