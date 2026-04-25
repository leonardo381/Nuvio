<script>
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { collections } from "@/stores/collections";

    export let websitesCollection = null;

    let websites = [];
    let selectedWebsiteId = "";
    let contacts = [];
    let subscribers = [];
    let campaigns = [];
    let whatsappInteractions = [];
    let reportSnapshots = [];
    let reviews = [];

    let isLoadingWebsites = false;
    let isLoadingDashboard = false;
    let selectedSnapshotIndex = 0;

    let lastWebsitesCollectionId = "";
    let lastDashboardDataKey = "";

    const now = new Date();
    const currentMonthStartTimestamp = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0).getTime();

    $: contactsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "contacts") || null;
    $: subscribersCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "subscribers") || null;
    $: campaignsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "campaigns") || null;
    $: whatsappInteractionsCollection = $collections.find((c) => {
        const normalized = (c?.name || "").toLowerCase();
        return normalized === "whatsapp_interactions" || normalized === "whatsapp_clicks";
    }) || null;
    $: reportsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "reports") || null;
    $: reviewsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "reviews") || null;

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        contacts = [];
        subscribers = [];
        campaigns = [];
        whatsappInteractions = [];
        reportSnapshots = [];
        reviews = [];
        lastWebsitesCollectionId = "";
        lastDashboardDataKey = "";
        selectedSnapshotIndex = 0;
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: dashboardDataKey = [
        selectedWebsiteId,
        contactsCollection?.id || "",
        subscribersCollection?.id || "",
        campaignsCollection?.id || "",
        whatsappInteractionsCollection?.id || "",
        reportsCollection?.id || "",
        reviewsCollection?.id || "",
    ].join(":");

    $: if (selectedWebsiteId && dashboardDataKey !== lastDashboardDataKey) {
        lastDashboardDataKey = dashboardDataKey;
        loadDashboardData();
    }

    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
    $: reportsFeatureAvailable = resolveReportsFeatureAvailable(selectedWebsite);
    $: selectedSnapshot = reportSnapshots[selectedSnapshotIndex] || null;
    $: selectedSnapshotData = parseReportData(selectedSnapshot?.data);

    $: contactsByChannel = {
        form: contacts.filter((record) => normalizeValue(record?.channel) === "form").length,
        booking: contacts.filter((record) => normalizeValue(record?.channel) === "booking").length,
        whatsapp: whatsappInteractions.length,
    };
    $: totalContacts = contacts.length;
    $: totalLeads = totalContacts + contactsByChannel.whatsapp;
    $: contactsThisMonth = contacts.filter((record) => isCurrentMonth(record?.created)).length;
    $: bookingsThisMonth = contacts.filter((record) => normalizeValue(record?.channel) === "booking" && isCurrentMonth(record?.created)).length;
    $: whatsappThisMonth = whatsappInteractions.filter((record) => isCurrentMonth(record?.created)).length;
    $: totalLeadsThisMonth = contactsThisMonth + whatsappThisMonth;

    $: activeSubscribers = subscribers.filter((record) => normalizeValue(record?.status) === "active").length;
    $: newSubscribersThisMonth = subscribers.filter((record) => isCurrentMonth(record?.created)).length;
    $: draftCampaigns = campaigns.filter((record) => normalizeValue(record?.status) === "draft").length;
    $: sentCampaignsThisMonth = campaigns.filter((record) => {
        if (normalizeValue(record?.status) !== "sent") {
            return false;
        }
        return isCurrentMonth(record?.sentAt || record?.updated || record?.created);
    }).length;

    $: trafficVisitors = resolveMetric(selectedSnapshotData, "traffic.visitors", "-");
    $: trafficPageviews = resolveMetric(selectedSnapshotData, "traffic.pageviews", "-");
    $: trafficBounceRate = resolveMetric(selectedSnapshotData, "traffic.bounceRate", "-");
    $: trafficAvgDuration = resolveMetric(selectedSnapshotData, "traffic.avgDuration", "-");
    $: trafficTopPages = Array.isArray(selectedSnapshotData?.traffic?.topPages) ? selectedSnapshotData.traffic.topPages : [];
    $: trafficSources = Array.isArray(selectedSnapshotData?.traffic?.sources) ? selectedSnapshotData.traffic.sources : [];
    $: trafficDevices = selectedSnapshotData?.traffic?.devices && typeof selectedSnapshotData.traffic.devices === "object"
        ? selectedSnapshotData.traffic.devices
        : null;

    $: reviewStats = resolveReviewStats({
        snapshotData: selectedSnapshotData,
        reviews,
    });

    $: reportRecommendations = buildReportRecommendations({
        totalLeadsThisMonth,
        bookingsThisMonth,
        whatsappThisMonth,
        activeSubscribers,
        newSubscribersThisMonth,
        sentCampaignsThisMonth,
        snapshotData: selectedSnapshotData,
    });

    $: reportHistoryRows = reportSnapshots.map((snapshot) => {
        const snapshotData = parseReportData(snapshot?.data);
        return {
            id: snapshot?.id || "",
            month: snapshot?.month || resolveMetric(snapshotData, "period", ""),
            visitors: resolveMetric(snapshotData, "traffic.visitors", "-"),
            contacts: resolveMetric(snapshotData, "contacts.total", "-"),
            bookings: resolveMetric(snapshotData, "bookings.total", "-"),
            subscribers: resolveMetric(snapshotData, "newsletter.activeSubscribers", "-"),
            sentAt: snapshot?.sent_at || snapshot?.sentAt || snapshot?.created || "",
        };
    });

    $: leadChannelRows = [
        { key: "form", label: "Form", count: contactsByChannel.form },
        { key: "whatsapp", label: "WhatsApp", count: contactsByChannel.whatsapp },
        { key: "booking", label: "Booking", count: contactsByChannel.booking },
    ];

    $: sourceStatusRows = [
        {
            label: "Leads (contacts + WhatsApp)",
            ok: !!contactsCollection?.id || !!whatsappInteractionsCollection?.id,
            message: contactsCollection?.id || whatsappInteractionsCollection?.id
                ? "Live collection data available."
                : "Collections missing.",
        },
        {
            label: "Newsletter (subscribers + campaigns)",
            ok: !!subscribersCollection?.id || !!campaignsCollection?.id,
            message: subscribersCollection?.id || campaignsCollection?.id
                ? "Live collection data available."
                : "Collections missing.",
        },
        {
            label: "Monthly snapshots",
            ok: !!reportsCollection?.id,
            message: reportsCollection?.id
                ? `${reportSnapshots.length} snapshot(s) loaded.`
                : "Reports collection not found.",
        },
        {
            label: "Reviews",
            ok: !!reviewsCollection?.id,
            message: reviewsCollection?.id
                ? `${reviews.length} review record(s) loaded.`
                : "Reviews collection optional and currently unavailable.",
        },
    ];

    export function reload() {
        if (!selectedWebsiteId) {
            return;
        }

        return loadDashboardData();
    }

    function normalizeValue(value) {
        return `${value || ""}`.trim().toLowerCase();
    }

    function toTimestamp(value) {
        const raw = `${value || ""}`.trim();
        if (!raw) {
            return 0;
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const parsed = new Date(normalized).getTime();
        return Number.isNaN(parsed) ? 0 : parsed;
    }

    function isCurrentMonth(value) {
        const timestamp = toTimestamp(value);
        return timestamp >= currentMonthStartTimestamp;
    }

    function formatDateTime(value) {
        const raw = `${value || ""}`.trim();
        if (!raw) {
            return "-";
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const parsed = new Date(normalized);
        if (Number.isNaN(parsed.getTime())) {
            return raw;
        }

        return parsed.toLocaleString();
    }

    function formatMonth(value) {
        const raw = `${value || ""}`.trim();
        if (!raw) {
            return "-";
        }

        const matched = raw.match(/^(\d{4})-(\d{2})$/);
        if (!matched) {
            return raw;
        }

        const parsed = new Date(Number(matched[1]), Number(matched[2]) - 1, 1);
        if (Number.isNaN(parsed.getTime())) {
            return raw;
        }

        return parsed.toLocaleDateString(undefined, { month: "long", year: "numeric" });
    }

    function parseSettings(rawSettings) {
        if (rawSettings && typeof rawSettings === "object") {
            return rawSettings;
        }

        if (typeof rawSettings === "string") {
            try {
                const parsed = JSON.parse(rawSettings);
                return parsed && typeof parsed === "object" ? parsed : {};
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function resolveReportsFeatureAvailable(website) {
        const settings = parseSettings(website?.settings);
        const featureFlags = settings?.featureFlags && typeof settings.featureFlags === "object"
            ? settings.featureFlags
            : {};
        const reportsSettings = settings?.reports && typeof settings.reports === "object"
            ? settings.reports
            : {};

        if (featureFlags.reports === false) {
            return false;
        }

        if (reportsSettings.enabled === false) {
            return false;
        }

        return true;
    }

    function parseReportData(value) {
        if (!value) {
            return {};
        }

        if (typeof value === "object") {
            return value;
        }

        if (typeof value === "string") {
            try {
                const parsed = JSON.parse(value);
                return parsed && typeof parsed === "object" ? parsed : {};
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function resolveMetric(source, path, fallback = "-") {
        const keys = Array.isArray(path) ? path : `${path || ""}`.split(".").filter(Boolean);
        let current = source;

        for (const key of keys) {
            if (!current || typeof current !== "object" || !(key in current)) {
                return fallback;
            }
            current = current[key];
        }

        if (typeof current === "number") {
            return current;
        }

        if (typeof current === "string" && current.trim()) {
            return current.trim();
        }

        return fallback;
    }

    function resolveReviewStats(payload = {}) {
        const snapshotReviews = payload.snapshotData?.reviews && typeof payload.snapshotData.reviews === "object"
            ? payload.snapshotData.reviews
            : {};
        const reviewRecords = Array.isArray(payload.reviews) ? payload.reviews : [];

        const ratings = reviewRecords
            .map((record) => Number(record?.rating ?? record?.stars ?? record?.score))
            .filter((value) => Number.isFinite(value) && value > 0);
        const avgFromRecords = ratings.length
            ? Math.round((ratings.reduce((sum, value) => sum + value, 0) / ratings.length) * 10) / 10
            : 0;

        return {
            rating: snapshotReviews.rating || avgFromRecords || "-",
            total: snapshotReviews.total || reviewRecords.length || "-",
            newThisMonth: snapshotReviews.newThisMonth || reviewRecords.filter((record) => isCurrentMonth(record?.created)).length || 0,
        };
    }

    function buildReportRecommendations(payload = {}) {
        const recommendations = [];
        const liveContacts = Number(payload.totalLeadsThisMonth || 0);
        const liveBookings = Number(payload.bookingsThisMonth || 0);
        const liveWhatsApp = Number(payload.whatsappThisMonth || 0);
        const liveActiveSubscribers = Number(payload.activeSubscribers || 0);
        const liveNewSubscribers = Number(payload.newSubscribersThisMonth || 0);
        const liveSentCampaigns = Number(payload.sentCampaignsThisMonth || 0);
        const previousContacts = Number(resolveMetric(payload.snapshotData, "contacts.total", 0) || 0);

        if (previousContacts && liveContacts < previousContacts) {
            recommendations.push("Contacts are below the latest snapshot. Review CTA copy on homepage and contact sections.");
        }

        if (liveWhatsApp < 1 && liveContacts < 5) {
            recommendations.push("WhatsApp channel is underused. Add or highlight WhatsApp CTA on high-intent pages.");
        }

        if (liveBookings < 1 && liveContacts > 0) {
            recommendations.push("You have inbound leads but no bookings this month. Simplify booking form and follow-up speed.");
        }

        if (liveActiveSubscribers > 0 && liveSentCampaigns < 1) {
            recommendations.push("No newsletter was sent this month. Schedule at least one campaign to keep engagement active.");
        }

        if (liveActiveSubscribers > 0 && liveNewSubscribers < 1) {
            recommendations.push("Subscriber growth is flat. Add stronger newsletter entry points on service/contact pages.");
        }

        if (!recommendations.length) {
            recommendations.push("Metrics are stable this month. Keep cadence and test one conversion improvement for next month.");
        }

        return recommendations.slice(0, 4);
    }

    function resolveChannelPercent(count) {
        if (!totalLeads || totalLeads < 1) {
            return 0;
        }
        const value = Math.round((Number(count || 0) / totalLeads) * 100);
        return Number.isNaN(value) ? 0 : Math.max(0, Math.min(100, value));
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

    function resolveWebsiteLabel(website) {
        return `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim() || website?.id || "";
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
                sort: resolveWebsitesSort(websitesCollection),
                requestKey: "nuvio_reports_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                contacts = [];
                subscribers = [];
                campaigns = [];
                whatsappInteractions = [];
                reportSnapshots = [];
                reviews = [];
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            } else {
                lastDashboardDataKey = "";
                await loadDashboardData();
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            contacts = [];
            subscribers = [];
            campaigns = [];
            whatsappInteractions = [];
            reportSnapshots = [];
            reviews = [];
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadRecordsByWebsite(collection, requestKeyPrefix, sort = "-created") {
        if (!collection?.id || !selectedWebsiteId) {
            return [];
        }

        try {
            return await ApiClient.collection(collection.id).getFullList({
                filter: `website="${selectedWebsiteId}"`,
                sort,
                requestKey: `${requestKeyPrefix}_${selectedWebsiteId}`,
            });
        } catch (err) {
            ApiClient.error(err);
            return [];
        }
    }

    async function loadDashboardData() {
        if (!selectedWebsiteId) {
            contacts = [];
            subscribers = [];
            campaigns = [];
            whatsappInteractions = [];
            reportSnapshots = [];
            reviews = [];
            return;
        }

        isLoadingDashboard = true;

        const [
            nextContacts,
            nextSubscribers,
            nextCampaigns,
            nextWhatsAppInteractions,
            nextReportSnapshots,
            nextReviews,
        ] = await Promise.all([
            loadRecordsByWebsite(contactsCollection, "nuvio_reports_contacts"),
            loadRecordsByWebsite(subscribersCollection, "nuvio_reports_subscribers"),
            loadRecordsByWebsite(campaignsCollection, "nuvio_reports_campaigns"),
            loadRecordsByWebsite(whatsappInteractionsCollection, "nuvio_reports_whatsapp"),
            loadRecordsByWebsite(reportsCollection, "nuvio_reports_snapshots", "-month,-created"),
            loadRecordsByWebsite(reviewsCollection, "nuvio_reports_reviews"),
        ]);

        contacts = nextContacts;
        subscribers = nextSubscribers;
        campaigns = nextCampaigns;
        whatsappInteractions = nextWhatsAppInteractions;
        reportSnapshots = nextReportSnapshots;
        reviews = nextReviews;
        selectedSnapshotIndex = 0;

        isLoadingDashboard = false;
    }

    function handleWebsiteChange(event) {
        selectedWebsiteId = event?.target?.value || "";
        selectedSnapshotIndex = 0;
    }

    function selectOlderSnapshot() {
        if (selectedSnapshotIndex >= reportSnapshots.length - 1) {
            return;
        }

        selectedSnapshotIndex += 1;
    }

    function selectNewerSnapshot() {
        if (selectedSnapshotIndex <= 0) {
            return;
        }

        selectedSnapshotIndex -= 1;
    }
</script>

<section class="reports-dashboard-module">
    <section class="panel reports-controls">
        <div class="reports-controls-main">
            <h4 class="m-0">Monthly Reports</h4>
            <p class="txt-sm txt-hint m-b-0">
                Monitor live performance and monthly snapshots from one place.
            </p>
        </div>
        <div class="reports-controls-side">
            <label class="txt-sm txt-hint" for="reports-website">Website</label>
            <select
                id="reports-website"
                class="input input-sm"
                bind:value={selectedWebsiteId}
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
    </section>

    {#if !selectedWebsiteId}
        <div class="empty-state">
            Select a website to view reports.
        </div>
    {:else if !reportsFeatureAvailable}
        <div class="alert alert-warning">
            <div class="icon">
                <i class="ri-information-line" />
            </div>
            <div>Reports are disabled for this website in settings.</div>
        </div>
    {:else if isLoadingDashboard}
        <div class="loading-state">
            <span class="loader loader-sm" />
            <span class="txt-hint">Loading report metrics...</span>
        </div>
    {:else}
        <section class="panel reports-period-panel m-b-base">
            <div>
                <h5 class="m-0">Reporting period</h5>
                <p class="txt-sm txt-hint m-b-0">
                    {#if selectedSnapshot}
                        Snapshot selected: <strong>{formatMonth(selectedSnapshot?.month || resolveMetric(selectedSnapshotData, "period", ""))}</strong>
                    {:else}
                        Live view for <strong>{formatMonth(`${now.getFullYear()}-${`${now.getMonth() + 1}`.padStart(2, "0")}`)}</strong> (no snapshot selected)
                    {/if}
                </p>
            </div>
            <div class="reports-period-actions">
                <button
                    type="button"
                    class="btn btn-sm btn-outline"
                    disabled={!reportSnapshots.length || selectedSnapshotIndex >= reportSnapshots.length - 1}
                    on:click={selectOlderSnapshot}
                >
                    <span class="txt">Older</span>
                </button>
                <button
                    type="button"
                    class="btn btn-sm btn-outline"
                    disabled={!reportSnapshots.length || selectedSnapshotIndex <= 0}
                    on:click={selectNewerSnapshot}
                >
                    <span class="txt">Newer</span>
                </button>
            </div>
        </section>

        <section class="reports-kpi-grid">
            <article class="panel reports-kpi-card">
                <h6 class="m-0 txt-hint">Leads this month</h6>
                <div class="reports-kpi-value">{totalLeadsThisMonth}</div>
                <div class="txt-sm txt-hint">Total leads: {totalLeads}</div>
            </article>
            <article class="panel reports-kpi-card">
                <h6 class="m-0 txt-hint">Bookings this month</h6>
                <div class="reports-kpi-value">{bookingsThisMonth}</div>
                <div class="txt-sm txt-hint">From contacts channel: booking</div>
            </article>
            <article class="panel reports-kpi-card">
                <h6 class="m-0 txt-hint">Active subscribers</h6>
                <div class="reports-kpi-value">{activeSubscribers}</div>
                <div class="txt-sm txt-hint">{newSubscribersThisMonth} new this month</div>
            </article>
            <article class="panel reports-kpi-card">
                <h6 class="m-0 txt-hint">Campaign delivery</h6>
                <div class="reports-kpi-value">{sentCampaignsThisMonth}</div>
                <div class="txt-sm txt-hint">{draftCampaigns} drafts pending</div>
            </article>
        </section>

        <section class="reports-grid-two m-t-base">
            <article class="panel">
                <div class="section-head m-b-sm">
                    <h5 class="m-0">Lead channels</h5>
                    <p class="txt-sm txt-hint m-b-0">Real-time breakdown from collections.</p>
                </div>
                <div class="reports-channel-list">
                    {#each leadChannelRows as row (row.key)}
                        <div class="reports-channel-item">
                            <div class="reports-channel-head">
                                <span>{row.label}</span>
                                <span class="txt-sm txt-hint">{row.count} ({resolveChannelPercent(row.count)}%)</span>
                            </div>
                            <div class="reports-channel-track">
                                <span class="reports-channel-fill" style={`width: ${resolveChannelPercent(row.count)}%;`} />
                            </div>
                        </div>
                    {/each}
                </div>
            </article>

            <article class="panel">
                <div class="section-head m-b-sm">
                    <h5 class="m-0">Traffic snapshot</h5>
                    <p class="txt-sm txt-hint m-b-0">From monthly snapshot data.</p>
                </div>

                {#if selectedSnapshot}
                    <div class="reports-traffic-grid">
                        <div class="reports-stat-box">
                            <span class="txt-xs txt-hint">Visitors</span>
                            <strong>{trafficVisitors}</strong>
                        </div>
                        <div class="reports-stat-box">
                            <span class="txt-xs txt-hint">Pageviews</span>
                            <strong>{trafficPageviews}</strong>
                        </div>
                        <div class="reports-stat-box">
                            <span class="txt-xs txt-hint">Bounce rate</span>
                            <strong>{trafficBounceRate}</strong>
                        </div>
                        <div class="reports-stat-box">
                            <span class="txt-xs txt-hint">Avg duration</span>
                            <strong>{trafficAvgDuration}</strong>
                        </div>
                    </div>

                    <div class="reports-mini-columns m-t-sm">
                        <div>
                            <h6 class="m-b-xs">Top pages</h6>
                            {#if trafficTopPages.length}
                                <ul class="reports-mini-list">
                                    {#each trafficTopPages.slice(0, 4) as page}
                                        <li>
                                            <span>{page?.page || "-"}</span>
                                            <strong>{page?.visitors || 0}</strong>
                                        </li>
                                    {/each}
                                </ul>
                            {:else}
                                <div class="txt-sm txt-hint">No top pages in this snapshot.</div>
                            {/if}
                        </div>
                        <div>
                            <h6 class="m-b-xs">Sources</h6>
                            {#if trafficSources.length}
                                <ul class="reports-mini-list">
                                    {#each trafficSources.slice(0, 4) as source}
                                        <li>
                                            <span>{source?.source || "-"}</span>
                                            <strong>{source?.visitors || 0}</strong>
                                        </li>
                                    {/each}
                                </ul>
                            {:else}
                                <div class="txt-sm txt-hint">No traffic sources in this snapshot.</div>
                            {/if}
                        </div>
                    </div>

                    {#if trafficDevices}
                        <div class="txt-sm txt-hint m-t-sm">
                            Devices: Mobile {trafficDevices.mobile ?? "-"}% · Desktop {trafficDevices.desktop ?? "-"}%
                        </div>
                    {/if}
                {:else}
                    <div class="empty-state">
                        No report snapshot yet. Live metrics are available, and traffic data appears after first snapshot.
                    </div>
                {/if}
            </article>
        </section>

        <section class="reports-grid-two m-t-base">
            <article class="panel">
                <div class="section-head m-b-sm">
                    <h5 class="m-0">Recommendations</h5>
                    <p class="txt-sm txt-hint m-b-0">Actionable suggestions based on current data.</p>
                </div>
                <ul class="reports-recommendations">
                    {#each reportRecommendations as recommendation}
                        <li>{recommendation}</li>
                    {/each}
                </ul>
            </article>

            <article class="panel">
                <div class="section-head m-b-sm">
                    <h5 class="m-0">Reviews</h5>
                    <p class="txt-sm txt-hint m-b-0">Snapshot + live collection overview.</p>
                </div>
                <div class="reports-traffic-grid">
                    <div class="reports-stat-box">
                        <span class="txt-xs txt-hint">Average rating</span>
                        <strong>{reviewStats.rating}</strong>
                    </div>
                    <div class="reports-stat-box">
                        <span class="txt-xs txt-hint">Total reviews</span>
                        <strong>{reviewStats.total}</strong>
                    </div>
                    <div class="reports-stat-box">
                        <span class="txt-xs txt-hint">New this month</span>
                        <strong>{reviewStats.newThisMonth}</strong>
                    </div>
                </div>

                <div class="reports-source-status m-t-sm">
                    {#each sourceStatusRows as row}
                        <div class="reports-source-row">
                            <span>{row.label}</span>
                            <span class:txt-success={row.ok} class:txt-hint={!row.ok}>{row.message}</span>
                        </div>
                    {/each}
                </div>
            </article>
        </section>

        <section class="panel m-t-base">
            <div class="section-head m-b-sm">
                <h5 class="m-0">Report history</h5>
                <p class="txt-sm txt-hint m-b-0">Immutable monthly snapshots generated by automation.</p>
            </div>

            {#if !reportHistoryRows.length}
                <div class="empty-state">
                    No snapshot history available yet for this website.
                </div>
            {:else}
                <div class="reports-history-list">
                    {#each reportHistoryRows as row (row.id)}
                        <div class="reports-history-item">
                            <div class="reports-history-month">{formatMonth(row.month)}</div>
                            <div class="reports-history-metrics">
                                <span>{row.visitors} visitors</span>
                                <span>{row.contacts} contacts</span>
                                <span>{row.bookings} bookings</span>
                                <span>{row.subscribers} active subscribers</span>
                            </div>
                            <div class="txt-sm txt-hint">Sent: {formatDateTime(row.sentAt)}</div>
                        </div>
                    {/each}
                </div>
            {/if}
        </section>
    {/if}
</section>

<style>
    .reports-dashboard-module {
        display: grid;
        gap: var(--baseSpacing);
    }

    .reports-controls {
        display: flex;
        align-items: flex-end;
        justify-content: space-between;
        gap: 16px;
        flex-wrap: wrap;
    }

    .reports-controls-main {
        display: grid;
        gap: 6px;
    }

    .reports-controls-side {
        min-width: min(100%, 380px);
        display: grid;
        gap: 6px;
    }

    .reports-period-panel {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        flex-wrap: wrap;
    }

    .reports-period-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
    }

    .reports-kpi-grid {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(4, minmax(150px, 1fr));
    }

    .reports-kpi-card {
        display: grid;
        gap: 6px;
    }

    .reports-kpi-value {
        font-size: 26px;
        line-height: 1.1;
        font-weight: 700;
        color: var(--txtPrimaryColor);
    }

    .reports-grid-two {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-channel-list {
        display: grid;
        gap: 10px;
    }

    .reports-channel-item {
        display: grid;
        gap: 5px;
    }

    .reports-channel-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .reports-channel-track {
        height: 9px;
        border-radius: 999px;
        background: var(--baseAlt1Color);
        border: 1px solid var(--baseAlt2Color);
        overflow: hidden;
    }

    .reports-channel-fill {
        display: block;
        height: 100%;
        border-radius: inherit;
        background: color-mix(in srgb, var(--primaryColor) 70%, var(--successColor));
    }

    .reports-traffic-grid {
        display: grid;
        gap: 8px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-stat-box {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 8px 10px;
        display: grid;
        gap: 5px;
    }

    .reports-stat-box strong {
        font-size: 17px;
        line-height: 1.2;
    }

    .reports-mini-columns {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-mini-list {
        margin: 0;
        padding: 0;
        list-style: none;
        display: grid;
        gap: 6px;
    }

    .reports-mini-list li {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 6px 8px;
    }

    .reports-recommendations {
        margin: 0;
        padding-left: 18px;
        display: grid;
        gap: 8px;
    }

    .reports-source-status {
        display: grid;
        gap: 6px;
    }

    .reports-source-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .reports-history-list {
        display: grid;
        gap: 8px;
    }

    .reports-history-item {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 8px 10px;
        display: grid;
        gap: 4px;
    }

    .reports-history-month {
        font-weight: 600;
    }

    .reports-history-metrics {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
    }

    @media (max-width: 1180px) {
        .reports-kpi-grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 980px) {
        .reports-grid-two {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 640px) {
        .reports-kpi-grid {
            grid-template-columns: 1fr;
        }

        .reports-traffic-grid,
        .reports-mini-columns {
            grid-template-columns: 1fr;
        }

        .reports-period-panel {
            align-items: flex-start;
        }
    }
</style>
