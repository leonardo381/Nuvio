<script>
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { normalizeWebsiteSettingsValue } from "@/utils/WebsiteSettingsSchema";

    $pageTitle = "Reports";

    const reportsTabs = [
        { key: "overview", label: "Overview", icon: "ri-dashboard-line" },
        { key: "traffic", label: "Traffic", icon: "ri-line-chart-line" },
        { key: "leads", label: "Leads", icon: "ri-mail-line" },
        { key: "booking", label: "Booking", icon: "ri-calendar-check-line" },
        { key: "newsletter", label: "Newsletter", icon: "ri-megaphone-line" },
        { key: "seo", label: "SEO", icon: "ri-earth-line" },
        { key: "history", label: "History", icon: "ri-history-line" },
    ];

    const periodOptions = [
        { key: "thisMonth", label: "This month" },
        { key: "lastMonth", label: "Last month" },
        { key: "last30Days", label: "Last 30 days" },
        { key: "allTime", label: "All time" },
    ];

    const archivedStatusAliases = ["archived", "archive"];
    const clientTrafficUnavailableMessage = "Traffic analytics are not configured for this website yet.";

    let activeTab = "overview";
    let selectedPeriod = "thisMonth";

    let websites = [];
    let selectedWebsiteId = "";

    let contactsRecords = [];
    let whatsappRecords = [];
    let appointmentsRecords = [];
    let bookingServicesRecords = [];
    let subscribersRecords = [];
    let campaignsRecords = [];
    let pagesRecords = [];
    let dashboardResponse = null;

    let isLoadingWebsites = false;
    let isLoadingData = false;
    let isLoadingTraffic = false;
    let reportsLoadError = "";

    let lastWebsitesCollectionId = "";
    let lastDataKey = "";
    let lastTrafficKey = "";
    let dataLoadToken = 0;
    let trafficLoadToken = 0;
    let lastSelectedWebsiteHydrationKey = "";
    let lastTrafficAnalyticsSetupSeed = "";
    let isSavingTrafficAnalyticsSetup = false;
    let trafficAnalyticsSetupError = "";
    let trafficAnalyticsSetupSuccess = "";

    let trafficResponse = null;
    let trafficAnalyticsSetupDraft = {
        enabled: false,
        siteId: "",
        scriptEnabled: false,
        scriptUrl: "",
        scrollDepth: false,
    };

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: websiteSettingsField = resolveWebsiteSettingsField(websitesCollection);

    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
    $: selectedWebsiteHydrationKey = `${websitesCollection?.id || ""}:${selectedWebsiteId || ""}`;
    $: if (selectedWebsiteHydrationKey !== lastSelectedWebsiteHydrationKey) {
        lastSelectedWebsiteHydrationKey = selectedWebsiteHydrationKey;
        hydrateSelectedWebsiteRecord();
    }
    $: selectedWebsiteSettings = normalizeWebsiteSettingsValue(selectedWebsite?.[websiteSettingsField]);
    $: selectedWebsiteFeatureFlags = selectedWebsiteSettings?.featureFlags && typeof selectedWebsiteSettings.featureFlags === "object"
        ? selectedWebsiteSettings.featureFlags
        : {};
    $: selectedWebsiteReportsSettings = selectedWebsiteSettings?.reports && typeof selectedWebsiteSettings.reports === "object"
        ? selectedWebsiteSettings.reports
        : {};
    $: selectedWebsiteReportsAnalytics = selectedWebsiteReportsSettings?.analytics && typeof selectedWebsiteReportsSettings.analytics === "object"
        ? selectedWebsiteReportsSettings.analytics
        : {};
    $: selectedWebsiteReportsEvents = selectedWebsiteReportsAnalytics?.events && typeof selectedWebsiteReportsAnalytics.events === "object"
        ? selectedWebsiteReportsAnalytics.events
        : {};
    $: reportsFeatureAvailable = resolveReportsFeatureAvailable(selectedWebsite);
    $: canConfigureTrafficAnalytics = ApiClient.isAdminSuperuser();
    $: trafficAnalyticsSetupMissingReasons = resolveTrafficAnalyticsSetupMissingReasons(selectedWebsiteReportsAnalytics);
    $: showTrafficAnalyticsSetup = canConfigureTrafficAnalytics && trafficAnalyticsSetupMissingReasons.length > 0;
    $: trafficAnalyticsSetupSeed = [
        normalizeString(selectedWebsite?.id),
        normalizeString(selectedWebsite?.updated),
        selectedWebsiteReportsAnalytics?.enabled ? "1" : "0",
        normalizeLower(selectedWebsiteReportsAnalytics?.provider),
        normalizeString(selectedWebsiteReportsAnalytics?.siteId),
        selectedWebsiteReportsAnalytics?.scriptEnabled ? "1" : "0",
        normalizeString(selectedWebsiteReportsAnalytics?.scriptUrl),
        selectedWebsiteReportsEvents?.scrollDepth ? "1" : "0",
    ].join("|");
    $: if (trafficAnalyticsSetupSeed !== lastTrafficAnalyticsSetupSeed) {
        lastTrafficAnalyticsSetupSeed = trafficAnalyticsSetupSeed;
        trafficAnalyticsSetupDraft = buildTrafficAnalyticsSetupDraft(selectedWebsiteReportsAnalytics);
        trafficAnalyticsSetupError = "";
        trafficAnalyticsSetupSuccess = "";
    }

    $: selectedPeriodLabel = periodOptions.find((option) => option.key === selectedPeriod)?.label || "This month";
    $: selectedTrafficPeriod = mapPeriodToTrafficPeriod(selectedPeriod);
    $: trafficState = normalizeLower(trafficResponse?.state) || "";
    $: trafficPeriod = trafficResponse?.period || null;
    $: trafficSummary = trafficState === "ok" && trafficResponse?.summary && typeof trafficResponse.summary === "object"
        ? trafficResponse.summary
        : null;
    $: trafficTopPages = trafficState === "ok" && Array.isArray(trafficResponse?.topPages)
        ? trafficResponse.topPages
        : [];
    $: trafficSources = trafficState === "ok" && Array.isArray(trafficResponse?.sources)
        ? trafficResponse.sources
        : [];
    $: trafficDevices = trafficState === "ok" && Array.isArray(trafficResponse?.devices)
        ? trafficResponse.devices
        : [];
    $: trafficEntryPages = trafficState === "ok"
        ? normalizeTrafficEntryRows(trafficResponse?.entryPages)
        : [];
    $: trafficExitPages = trafficState === "ok"
        ? normalizeTrafficEntryRows(trafficResponse?.exitPages)
        : [];
    $: trafficCountries = trafficState === "ok"
        ? normalizeTrafficDimensionRows(trafficResponse?.countries, ["country", "name", "label", "dimension"], "Unknown")
        : [];
    $: trafficRegions = trafficState === "ok"
        ? normalizeTrafficDimensionRows(trafficResponse?.regions, ["region", "name", "label", "dimension"], "Unknown")
        : [];
    $: trafficCities = trafficState === "ok"
        ? normalizeTrafficDimensionRows(trafficResponse?.cities, ["city", "name", "label", "dimension"], "Unknown")
        : [];
    $: trafficBrowsers = trafficState === "ok"
        ? normalizeTrafficDimensionRows(trafficResponse?.browsers, ["browser", "name", "label", "dimension"], "Unknown")
        : [];
    $: trafficOperatingSystems = trafficState === "ok"
        ? normalizeTrafficDimensionRows(trafficResponse?.operatingSystems, ["operatingSystem", "os", "name", "label", "dimension"], "Unknown")
        : [];
    $: trafficInsights = normalizeTrafficInsights(trafficResponse?.insights);
    $: trafficConversions = trafficState === "ok" && trafficResponse?.conversions && typeof trafficResponse.conversions === "object"
        ? trafficResponse.conversions
        : null;
    $: trafficEngagement = trafficState === "ok" && trafficResponse?.engagement && typeof trafficResponse.engagement === "object"
        ? trafficResponse.engagement
        : null;
    $: trafficConversionsState = normalizeLower(trafficConversions?.state) || "unavailable";
    $: trafficConversionsMessage = normalizeString(trafficConversions?.message)
        || (trafficConversionsState === "ok"
            ? "Conversion event metrics are available."
            : "Conversion event metrics are unavailable.");
    $: trafficConversionTotals = trafficConversions?.totals && typeof trafficConversions.totals === "object"
        ? trafficConversions.totals
        : { allEvents: 0, uniqueEventTypes: 0 };
    $: trafficConversionByType = trafficState === "ok"
        ? normalizeTrafficCountRows(trafficConversions?.byType, ["event", "name", "label"], "Unknown action", ["count", "value", "visitors"])
        : [];
    $: trafficConversionByPage = trafficState === "ok"
        ? normalizeTrafficCountRows(trafficConversions?.byPage, ["pageSlug", "page", "path", "label"], "Unknown page", ["count", "value", "visitors"])
        : [];
    $: trafficConversionBySourceBlock = trafficState === "ok"
        ? normalizeTrafficCountRows(trafficConversions?.bySourceBlock, ["sourceBlock", "source", "label"], "Unknown source", ["count", "value", "visitors"])
        : [];
    $: trafficConversionByCtaType = trafficState === "ok"
        ? normalizeTrafficCountRows(trafficConversions?.byCtaType, ["ctaType", "type", "label"], "Unknown CTA", ["count", "value", "visitors"])
        : [];
    $: trafficScrollDepth = trafficEngagement?.scrollDepth && typeof trafficEngagement.scrollDepth === "object"
        ? trafficEngagement.scrollDepth
        : null;
    $: trafficScrollDepthState = normalizeLower(trafficScrollDepth?.state) || "unavailable";
    $: trafficScrollDepthMessage = normalizeString(trafficScrollDepth?.message)
        || (trafficScrollDepthState === "ok"
            ? "Scroll depth events are available."
            : "Scroll depth metrics are unavailable.");
    $: trafficScrollDepthThresholds = trafficState === "ok"
        ? normalizeTrafficDepthRows(trafficScrollDepth?.thresholds)
        : [];
    $: trafficScrollDepthByPage = trafficState === "ok"
        ? normalizeTrafficCountRows(trafficScrollDepth?.byPage, ["pageSlug", "page", "path", "label"], "Unknown page", ["count", "value", "visitors"])
        : [];
    $: trafficDisplayState = trafficState || "analytics_not_configured";
    $: topTrafficPage = trafficTopPages[0] || null;
    $: topTrafficSource = trafficSources[0] || null;
    $: topTrafficDevice = trafficDevices[0] || null;
    $: trafficTopPageRows = buildTrafficPageBreakdownRows(trafficTopPages);
    $: trafficEntryPageRows = buildTrafficPageBreakdownRows(trafficEntryPages);
    $: trafficExitPageRows = buildTrafficPageBreakdownRows(trafficExitPages);
    $: trafficSourceRows = buildTrafficSimpleBreakdownRows(trafficSources, ["source", "label"], "Direct", ["visitors", "count", "value"]);
    $: trafficDeviceRows = buildTrafficSimpleBreakdownRows(trafficDevices, ["device", "label"], "Unknown", ["visitors", "count", "value"]);
    $: trafficCountryRows = buildTrafficSimpleBreakdownRows(trafficCountries, ["label", "country"], "Unknown", ["visitors", "count", "value"]);
    $: trafficRegionRows = buildTrafficSimpleBreakdownRows(trafficRegions, ["label", "region"], "Unknown", ["visitors", "count", "value"]);
    $: trafficCityRows = buildTrafficSimpleBreakdownRows(trafficCities, ["label", "city"], "Unknown", ["visitors", "count", "value"]);
    $: trafficBrowserRows = buildTrafficSimpleBreakdownRows(trafficBrowsers, ["label", "browser"], "Unknown", ["visitors", "count", "value"]);
    $: trafficOperatingSystemRows = buildTrafficSimpleBreakdownRows(trafficOperatingSystems, ["label", "operatingSystem", "os"], "Unknown", ["visitors", "count", "value"]);
    $: trafficConversionByTypeRows = trafficConversionByType
        .map((row) => ({
            label: resolveConversionEventLabel(row.label),
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficConversionByPageRows = trafficConversionByPage
        .map((row) => ({
            label: row.label || "Unknown page",
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficConversionBySourceRows = trafficConversionBySourceBlock
        .map((row) => ({
            label: row.label || "Unknown source",
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficConversionByCtaRows = trafficConversionByCtaType
        .map((row) => ({
            label: row.label || "Unknown CTA",
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficScrollDepthThresholdRows = trafficScrollDepthThresholds
        .map((row) => ({
            label: `${row.depth}% reached`,
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficScrollDepthByPageRows = trafficScrollDepthByPage
        .map((row) => ({
            label: row.label || "Unknown page",
            count: Number(row.count) || 0,
            meta: "",
        }))
        .slice(0, 8);
    $: trafficOverviewStatus = resolveTrafficOverviewStatus({
        trafficState,
        trafficInsights,
        trafficMessage: trafficResponse?.message,
    });
    $: trafficOverviewStatusLabel = resolveMetricStateLabel(trafficOverviewStatus, trafficState === "ok" ? "Ready" : "Check");
    $: trafficOverviewMessage = resolveTrafficOverviewMessage({
        trafficState,
        trafficStatus: trafficOverviewStatus,
        trafficMessage: trafficResponse?.message,
    });
    $: trafficConversionTotalsAllEvents = Number(trafficConversionTotals?.allEvents) || 0;
    $: trafficConversionTotalsUniqueEventTypes = Number(trafficConversionTotals?.uniqueEventTypes) || 0;
    $: trafficConversionIsOk = trafficConversionsState === "ok";
    $: trafficConversionIsPartial = trafficConversionsState === "partial";
    $: trafficConversionHasKnownTotals = trafficConversionTotalsAllEvents > 0 || trafficConversionTotalsUniqueEventTypes > 0;
    $: trafficConversionShowMeasuredSummary = trafficConversionIsOk;
    $: trafficConversionShowKnownSummary = trafficConversionIsPartial && trafficConversionHasKnownTotals;
    $: trafficConversionHasDirectionalData = hasPositiveTrafficRows(trafficConversionByTypeRows)
        || hasPositiveTrafficRows(trafficConversionByPageRows)
        || hasPositiveTrafficRows(trafficConversionBySourceRows)
        || hasPositiveTrafficRows(trafficConversionByCtaRows);
    $: trafficConversionShouldRenderBreakdowns = trafficConversionIsOk
        ? trafficConversionTotalsAllEvents > 0
        : trafficConversionIsPartial && trafficConversionHasDirectionalData;
    $: trafficConversionUnavailableMessage = resolveConversionUnavailableMessage(trafficConversionsState, trafficConversionsMessage);
    $: trafficConversionByTypeRenderableRows = trafficConversionIsPartial
        ? trafficConversionByTypeRows.filter((row) => Number(row?.count) > 0)
        : trafficConversionByTypeRows;
    $: trafficConversionByPageRenderableRows = trafficConversionIsPartial
        ? trafficConversionByPageRows.filter((row) => Number(row?.count) > 0)
        : trafficConversionByPageRows;
    $: trafficConversionBySourceRenderableRows = trafficConversionIsPartial
        ? trafficConversionBySourceRows.filter((row) => Number(row?.count) > 0)
        : trafficConversionBySourceRows;
    $: trafficConversionByCtaRenderableRows = trafficConversionIsPartial
        ? trafficConversionByCtaRows.filter((row) => Number(row?.count) > 0)
        : trafficConversionByCtaRows;
    $: trafficScrollDepthIsOk = trafficScrollDepthState === "ok";
    $: trafficScrollDepthIsPartial = trafficScrollDepthState === "partial";
    $: trafficScrollDepthHasDirectionalData = hasPositiveTrafficRows(trafficScrollDepthThresholdRows)
        || hasPositiveTrafficRows(trafficScrollDepthByPageRows);
    $: trafficScrollDepthShouldRenderBreakdowns = trafficScrollDepthIsOk
        ? trafficScrollDepthHasDirectionalData
        : trafficScrollDepthIsPartial && trafficScrollDepthHasDirectionalData;
    $: trafficScrollDepthUnavailableMessage = resolveScrollDepthUnavailableMessage(trafficScrollDepthState, trafficScrollDepthMessage);
    $: trafficScrollDepthThresholdRenderableRows = trafficScrollDepthIsPartial
        ? trafficScrollDepthThresholdRows.filter((row) => Number(row?.count) > 0)
        : trafficScrollDepthThresholdRows;
    $: trafficScrollDepthByPageRenderableRows = trafficScrollDepthIsPartial
        ? trafficScrollDepthByPageRows.filter((row) => Number(row?.count) > 0)
        : trafficScrollDepthByPageRows;
    $: trafficVisitorsCount = Number(trafficSummary?.visitors) || 0;
    $: trafficVisitsCount = Number(trafficSummary?.visits) || 0;
    $: trafficPageviewsCount = Number(trafficSummary?.pageviews) || 0;
    $: trafficHasNativeBreakdowns = hasPositiveTrafficRows(trafficTopPageRows)
        || hasPositiveTrafficRows(trafficEntryPageRows)
        || hasPositiveTrafficRows(trafficExitPageRows)
        || hasPositiveTrafficRows(trafficSourceRows)
        || hasPositiveTrafficRows(trafficCountryRows)
        || hasPositiveTrafficRows(trafficRegionRows)
        || hasPositiveTrafficRows(trafficCityRows)
        || hasPositiveTrafficRows(trafficDeviceRows)
        || hasPositiveTrafficRows(trafficBrowserRows)
        || hasPositiveTrafficRows(trafficOperatingSystemRows);
    $: trafficHasConversionSignals = hasPositiveTrafficRows(trafficConversionByTypeRenderableRows)
        || hasPositiveTrafficRows(trafficConversionByPageRenderableRows)
        || hasPositiveTrafficRows(trafficConversionBySourceRenderableRows)
        || hasPositiveTrafficRows(trafficConversionByCtaRenderableRows);
    $: trafficHasScrollSignals = hasPositiveTrafficRows(trafficScrollDepthThresholdRenderableRows)
        || hasPositiveTrafficRows(trafficScrollDepthByPageRenderableRows);
    $: trafficHasAnyMeasuredActivity = trafficVisitorsCount > 0
        || trafficVisitsCount > 0
        || trafficPageviewsCount > 0
        || trafficHasNativeBreakdowns
        || trafficHasConversionSignals
        || trafficHasScrollSignals;
    $: trafficNoDataYet = trafficState === "ok" && !trafficHasAnyMeasuredActivity;
    $: trafficNoDataMessage = canConfigureTrafficAnalytics
        ? "No traffic data yet for this period. Publish and visit the public website to start collecting analytics."
        : "No traffic data yet for this period. Traffic metrics will appear after the public website receives visits.";
    $: trafficHeroCards = [
        {
            key: "visitors",
            label: "Visitors",
            value: formatMetricNumber(trafficVisitorsCount),
            hint: trafficPeriod?.label || selectedPeriodLabel,
            meta: `${formatMetricNumber(trafficPageviewsCount)} pageviews tracked`,
            icon: "ri-user-search-line",
            badgeLabel: trafficOverviewStatusLabel,
            badgeClass: resolveMetricStatePillClass(trafficOverviewStatus),
        },
        {
            key: "visits",
            label: "Visits",
            value: formatMetricNumber(trafficVisitsCount),
            hint: "Total sessions in this period",
            meta: trafficVisitorsCount > 0
                ? formatShareOfTotal(trafficVisitsCount, trafficVisitorsCount, "sessions per visitor (directional)")
                : "",
            icon: "ri-footprint-line",
            badgeLabel: trafficVisitsCount > 0 ? "Active" : "No visits yet",
            badgeClass: trafficVisitsCount > 0 ? "label-success" : "",
        },
        {
            key: "pageviews",
            label: "Pageviews",
            value: formatMetricNumber(trafficPageviewsCount),
            hint: "Pages viewed across your website",
            meta: trafficVisitsCount > 0
                ? formatShareOfTotal(trafficPageviewsCount, trafficVisitsCount, "views per visit (directional)")
                : "",
            icon: "ri-file-chart-line",
            badgeLabel: trafficPageviewsCount > 0 ? "Reading activity" : "No pageviews yet",
            badgeClass: trafficPageviewsCount > 0 ? "label-success" : "",
        },
        {
            key: "bounceRate",
            label: "Bounce / interaction rate",
            value: formatBounceRate(trafficSummary?.bounceRate),
            hint: "Single-page visits compared with interactions",
            meta: trafficOverviewMessage || "",
            icon: "ri-bar-chart-grouped-line",
            badgeLabel: trafficOverviewStatusLabel,
            badgeClass: resolveMetricStatePillClass(trafficOverviewStatus),
        },
        {
            key: "visitDuration",
            label: "Average visit duration",
            value: formatDurationSeconds(trafficSummary?.visitDurationSeconds),
            hint: "Time spent per visit on average",
            meta: trafficNoDataYet ? "Data appears after public visits are tracked." : "",
            icon: "ri-timer-line",
            badgeLabel: trafficNoDataYet ? "Awaiting traffic" : "Measured",
            badgeClass: trafficNoDataYet ? "label-warning" : "label-success",
        },
    ];

    $: dataKey = [
        selectedWebsiteId,
    ].join(":");
    $: trafficKey = [
        selectedWebsiteId,
        selectedTrafficPeriod,
    ].join(":");

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        clearDataRecords();
        clearTrafficRecords();
        reportsLoadError = "";
        lastWebsitesCollectionId = "";
        lastDataKey = "";
        lastTrafficKey = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: if (selectedWebsiteId && dataKey !== lastDataKey) {
        lastDataKey = dataKey;
        loadDashboardData();
    }

    $: if (!selectedWebsiteId || !reportsFeatureAvailable) {
        clearTrafficRecords();
        lastTrafficKey = "";
    } else if (trafficKey !== lastTrafficKey) {
        lastTrafficKey = trafficKey;
        loadTrafficData();
    }

    $: websiteLabelById = new Map(
        websites.map((website) => [normalizeString(website?.id), resolveWebsiteLabel(website)]),
    );

    $: normalizedContacts = normalizeContactLeads(contactsRecords, websiteLabelById);
    $: normalizedWhatsApp = normalizeWhatsAppLeads(whatsappRecords, websiteLabelById);
    $: normalizedLeadRecords = [...normalizedContacts, ...normalizedWhatsApp];

    $: periodLeadRecords = normalizedLeadRecords.filter((lead) => isTimestampInPeriod(lead.createdTs, selectedPeriod));
    $: leadsByStatusCounts = {
        new: periodLeadRecords.filter((lead) => lead.statusKey === "new").length,
        read: periodLeadRecords.filter((lead) => lead.statusKey === "read").length,
        archived: periodLeadRecords.filter((lead) => lead.statusKey === "archived").length,
    };
    $: leadsByTypeCounts = {
        contact: periodLeadRecords.filter((lead) => lead.sourceKey === "contact").length,
        whatsapp: periodLeadRecords.filter((lead) => lead.sourceKey === "whatsapp").length,
        booking: periodLeadRecords.filter((lead) => lead.sourceKey === "booking").length,
    };

    $: leadsSummary = {
        total: periodLeadRecords.length,
        newCount: leadsByStatusCounts.new,
        readCount: leadsByStatusCounts.read,
        archivedCount: leadsByStatusCounts.archived,
        contactCount: leadsByTypeCounts.contact,
        whatsappCount: leadsByTypeCounts.whatsapp,
        bookingCount: leadsByTypeCounts.booking,
    };

    $: sortedRecentLeads = [...periodLeadRecords]
        .sort((a, b) => (b.createdTs || 0) - (a.createdTs || 0))
        .slice(0, 8);

    $: serviceNameById = new Map(
        (bookingServicesRecords || []).map((service) => [normalizeString(service?.id), normalizeString(service?.name) || "Untitled service"]),
    );

    $: normalizedAppointments = normalizeAppointments(appointmentsRecords, serviceNameById, websiteLabelById);
    $: periodAppointments = normalizedAppointments.filter((appointment) => isTimestampInPeriod(appointment.createdTs, selectedPeriod));

    $: bookingSummary = {
        total: periodAppointments.length,
        pendingCount: periodAppointments.filter((appointment) => appointment.statusKey === "pending").length,
        confirmedCount: periodAppointments.filter((appointment) => appointment.statusKey === "confirmed").length,
        cancelledCount: periodAppointments.filter((appointment) => appointment.statusKey === "cancelled").length,
        upcomingCount: normalizedAppointments.filter((appointment) => {
            if (appointment.statusKey === "cancelled") {
                return false;
            }
            return (appointment.scheduledTs || 0) >= Date.now();
        }).length,
    };

    $: topBookingServices = buildTopServices(periodAppointments);
    $: upcomingAppointments = [...normalizedAppointments]
        .filter((appointment) => {
            if (appointment.statusKey === "cancelled") {
                return false;
            }
            return (appointment.scheduledTs || 0) >= Date.now();
        })
        .sort((a, b) => (a.scheduledTs || 0) - (b.scheduledTs || 0))
        .slice(0, 8);
    $: leadSourceBreakdownRows = [
        {
            label: "Contact form",
            count: leadsSummary.contactCount,
            meta: formatShareOfTotal(leadsSummary.contactCount, leadsSummary.total, "leads"),
        },
        {
            label: "WhatsApp",
            count: leadsSummary.whatsappCount,
            meta: formatShareOfTotal(leadsSummary.whatsappCount, leadsSummary.total, "leads"),
        },
        {
            label: "Booking",
            count: leadsSummary.bookingCount,
            meta: formatShareOfTotal(leadsSummary.bookingCount, leadsSummary.total, "leads"),
        },
    ];
    $: leadsHasRecords = leadsSummary.total > 0;
    $: leadSourcesHasData = hasPositiveTrafficRows(leadSourceBreakdownRows);
    $: bookingStatusBreakdownRows = [
        {
            label: "Pending",
            count: bookingSummary.pendingCount,
            meta: formatShareOfTotal(bookingSummary.pendingCount, bookingSummary.total, "requests"),
        },
        {
            label: "Confirmed",
            count: bookingSummary.confirmedCount,
            meta: formatShareOfTotal(bookingSummary.confirmedCount, bookingSummary.total, "requests"),
        },
        {
            label: "Cancelled",
            count: bookingSummary.cancelledCount,
            meta: formatShareOfTotal(bookingSummary.cancelledCount, bookingSummary.total, "requests"),
        },
    ];
    $: bookingServiceBreakdownRows = topBookingServices
        .map((serviceRow) => ({
            label: serviceRow?.label || "Untitled service",
            count: Number(serviceRow?.count) || 0,
            meta: formatShareOfTotal(serviceRow?.count, bookingSummary.total, "requests"),
        }))
        .slice(0, 8);
    $: bookingHasRequests = bookingSummary.total > 0;
    $: bookingStatusHasData = hasPositiveTrafficRows(bookingStatusBreakdownRows);
    $: bookingServicesHasData = hasPositiveTrafficRows(bookingServiceBreakdownRows);
    $: leadsActionNeededCount = leadsSummary.newCount;
    $: leadsHeroCards = [
        {
            key: "totalLeads",
            label: "Total leads",
            value: formatMetricNumber(leadsSummary.total),
            hint: `Tracked in ${selectedPeriodLabel.toLowerCase()}`,
            meta: `${formatMetricNumber(leadsSummary.contactCount)} contact | ${formatMetricNumber(leadsSummary.whatsappCount)} WhatsApp | ${formatMetricNumber(leadsSummary.bookingCount)} booking`,
            icon: "ri-mail-line",
            badgeLabel: leadsHasRecords ? "Active period" : "No activity",
            badgeClass: leadsHasRecords ? "label-success" : "",
        },
        {
            key: "newLeads",
            label: "New leads",
            value: formatMetricNumber(leadsSummary.newCount),
            hint: "Awaiting first follow-up",
            meta: formatShareOfTotal(leadsSummary.newCount, leadsSummary.total, "leads"),
            icon: "ri-notification-3-line",
            badgeLabel: leadsSummary.newCount > 0 ? "Action needed" : "On track",
            badgeClass: leadsSummary.newCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "readLeads",
            label: "Read leads",
            value: formatMetricNumber(leadsSummary.readCount),
            hint: "Reviewed in backoffice",
            meta: formatShareOfTotal(leadsSummary.readCount, leadsSummary.total, "leads"),
            icon: "ri-mail-open-line",
            badgeLabel: leadsSummary.readCount > 0 ? "In progress" : "None yet",
            badgeClass: leadsSummary.readCount > 0 ? "label-success" : "",
        },
        {
            key: "archivedLeads",
            label: "Archived leads",
            value: formatMetricNumber(leadsSummary.archivedCount),
            hint: "Closed or archived records",
            meta: formatShareOfTotal(leadsSummary.archivedCount, leadsSummary.total, "leads"),
            icon: "ri-archive-stack-line",
            badgeLabel: leadsSummary.archivedCount > 0 ? "Archived" : "None",
            badgeClass: leadsSummary.archivedCount > 0 ? "" : "",
        },
        {
            key: "followUpNeeded",
            label: "Follow-up needed",
            value: formatMetricNumber(leadsActionNeededCount),
            hint: leadsActionNeededCount > 0
                ? "Prioritize new incoming leads"
                : "No urgent follow-up right now",
            meta: `${formatMetricNumber(leadsSummary.total)} total leads in period`,
            icon: "ri-todo-line",
            badgeLabel: leadsActionNeededCount > 0 ? "Needs attention" : "Healthy",
            badgeClass: leadsActionNeededCount > 0 ? "label-warning" : "label-success",
        },
    ];
    $: bookingHeroCards = [
        {
            key: "totalRequests",
            label: "Total requests",
            value: formatMetricNumber(bookingSummary.total),
            hint: `Tracked in ${selectedPeriodLabel.toLowerCase()}`,
            meta: `${formatMetricNumber(bookingSummary.pendingCount)} pending | ${formatMetricNumber(bookingSummary.confirmedCount)} confirmed`,
            icon: "ri-calendar-check-line",
            badgeLabel: bookingHasRequests ? "Active period" : "No activity",
            badgeClass: bookingHasRequests ? "label-success" : "",
        },
        {
            key: "pendingBookings",
            label: "Pending bookings",
            value: formatMetricNumber(bookingSummary.pendingCount),
            hint: "Waiting for confirmation",
            meta: formatShareOfTotal(bookingSummary.pendingCount, bookingSummary.total, "requests"),
            icon: "ri-time-line",
            badgeLabel: bookingSummary.pendingCount > 0 ? "Action needed" : "On track",
            badgeClass: bookingSummary.pendingCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "confirmedBookings",
            label: "Confirmed bookings",
            value: formatMetricNumber(bookingSummary.confirmedCount),
            hint: "Confirmed appointments",
            meta: formatShareOfTotal(bookingSummary.confirmedCount, bookingSummary.total, "requests"),
            icon: "ri-checkbox-circle-line",
            badgeLabel: bookingSummary.confirmedCount > 0 ? "Scheduled" : "None yet",
            badgeClass: bookingSummary.confirmedCount > 0 ? "label-success" : "",
        },
        {
            key: "cancelledBookings",
            label: "Cancelled bookings",
            value: formatMetricNumber(bookingSummary.cancelledCount),
            hint: "Cancelled or declined requests",
            meta: formatShareOfTotal(bookingSummary.cancelledCount, bookingSummary.total, "requests"),
            icon: "ri-close-circle-line",
            badgeLabel: bookingSummary.cancelledCount > 0 ? "Monitor" : "Stable",
            badgeClass: bookingSummary.cancelledCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "upcomingAppointments",
            label: "Upcoming appointments",
            value: formatMetricNumber(bookingSummary.upcomingCount),
            hint: "Future confirmed or pending slots",
            meta: `${formatMetricNumber(upcomingAppointments.length)} listed in operational view`,
            icon: "ri-calendar-event-line",
            badgeLabel: bookingSummary.upcomingCount > 0 ? "Upcoming" : "No upcoming",
            badgeClass: bookingSummary.upcomingCount > 0 ? "label-success" : "",
        },
    ];

    $: normalizedSubscribers = normalizeSubscribers(subscribersRecords);
    $: normalizedCampaigns = normalizeCampaigns(campaignsRecords);

    $: newsletterSummary = {
        activeSubscribers: normalizedSubscribers.filter((subscriber) => subscriber.statusKey === "active").length,
        newSubscribersPeriod: normalizedSubscribers.filter((subscriber) => isTimestampInPeriod(subscriber.createdTs, selectedPeriod)).length,
        sentCampaignsPeriod: normalizedCampaigns.filter((campaign) => {
            if (campaign.statusKey !== "sent") {
                return false;
            }
            return isTimestampInPeriod(campaign.sentTs || campaign.updatedTs || campaign.createdTs, selectedPeriod);
        }).length,
        recipientsReachedPeriod: normalizedCampaigns
            .filter((campaign) => {
                if (campaign.statusKey !== "sent") {
                    return false;
                }
                return isTimestampInPeriod(campaign.sentTs || campaign.updatedTs || campaign.createdTs, selectedPeriod);
            })
            .reduce((sum, campaign) => sum + Number(campaign.recipientsCount || 0), 0),
        draftCampaigns: normalizedCampaigns.filter((campaign) => campaign.statusKey === "draft").length,
    };
    $: pendingSubscribersCount = normalizedSubscribers.filter((subscriber) => subscriber.statusKey === "pending").length;
    $: unsubscribedSubscribersCount = normalizedSubscribers.filter((subscriber) => subscriber.statusKey === "unsubscribed").length;
    $: failedCampaignSubmissionsPeriod = normalizedCampaigns.filter((campaign) => {
        const status = normalizeLower(campaign?.statusKey);
        if (!["failed", "error", "rejected"].includes(status)) {
            return false;
        }
        return isTimestampInPeriod(campaign.sentTs || campaign.updatedTs || campaign.createdTs, selectedPeriod);
    }).length;
    $: newsletterHeroCards = [
        {
            key: "activeSubscribers",
            label: "Active subscribers",
            value: formatMetricNumber(newsletterSummary.activeSubscribers),
            hint: "Current active audience",
            meta: `${formatMetricNumber(newsletterSummary.newSubscribersPeriod)} new in ${selectedPeriodLabel.toLowerCase()}`,
            icon: "ri-user-follow-line",
            badgeLabel: newsletterSummary.activeSubscribers > 0 ? "Audience live" : "No audience yet",
            badgeClass: newsletterSummary.activeSubscribers > 0 ? "label-success" : "",
        },
        {
            key: "newSubscribers",
            label: "New subscribers",
            value: formatMetricNumber(newsletterSummary.newSubscribersPeriod),
            hint: `Added during ${selectedPeriodLabel.toLowerCase()}`,
            meta: formatShareOfTotal(newsletterSummary.newSubscribersPeriod, normalizedSubscribers.length, "subscribers"),
            icon: "ri-user-add-line",
            badgeLabel: newsletterSummary.newSubscribersPeriod > 0 ? "Growing" : "Steady",
            badgeClass: newsletterSummary.newSubscribersPeriod > 0 ? "label-success" : "",
        },
        {
            key: "pendingConfirmations",
            label: "Pending confirmations",
            value: formatMetricNumber(pendingSubscribersCount),
            hint: "Waiting to confirm subscription",
            meta: formatShareOfTotal(pendingSubscribersCount, normalizedSubscribers.length, "subscribers"),
            icon: "ri-time-line",
            badgeLabel: pendingSubscribersCount > 0 ? "Needs attention" : "On track",
            badgeClass: pendingSubscribersCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "unsubscribed",
            label: "Unsubscribed subscribers",
            value: formatMetricNumber(unsubscribedSubscribersCount),
            hint: "Audience churn to monitor",
            meta: formatShareOfTotal(unsubscribedSubscribersCount, normalizedSubscribers.length, "subscribers"),
            icon: "ri-user-unfollow-line",
            badgeLabel: unsubscribedSubscribersCount > 0 ? "Review" : "Stable",
            badgeClass: unsubscribedSubscribersCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "campaignsSubmitted",
            label: "Campaigns submitted",
            value: formatMetricNumber(newsletterSummary.sentCampaignsPeriod),
            hint: `${formatMetricNumber(newsletterSummary.recipientsReachedPeriod)} recipients submitted`,
            meta: `${formatMetricNumber(newsletterSummary.draftCampaigns)} drafts | ${formatMetricNumber(failedCampaignSubmissionsPeriod)} failed`,
            icon: "ri-send-plane-2-line",
            badgeLabel: newsletterSummary.sentCampaignsPeriod > 0 ? "Active sending" : "No campaigns sent",
            badgeClass: newsletterSummary.sentCampaignsPeriod > 0 ? "label-success" : "label-warning",
        },
    ];
    $: newsletterSubscriberStatusRows = [
        {
            label: "Active subscribers",
            count: newsletterSummary.activeSubscribers,
            meta: formatShareOfTotal(newsletterSummary.activeSubscribers, normalizedSubscribers.length, "subscribers"),
        },
        {
            label: "Pending confirmations",
            count: pendingSubscribersCount,
            meta: formatShareOfTotal(pendingSubscribersCount, normalizedSubscribers.length, "subscribers"),
        },
        {
            label: "Unsubscribed",
            count: unsubscribedSubscribersCount,
            meta: formatShareOfTotal(unsubscribedSubscribersCount, normalizedSubscribers.length, "subscribers"),
        },
    ];
    $: newsletterCampaignStatusRows = [
        {
            label: "Campaigns submitted",
            count: newsletterSummary.sentCampaignsPeriod,
            meta: formatShareOfTotal(newsletterSummary.sentCampaignsPeriod, newsletterSummary.sentCampaignsPeriod + newsletterSummary.draftCampaigns, "tracked campaigns"),
        },
        {
            label: "Draft campaigns",
            count: newsletterSummary.draftCampaigns,
            meta: formatShareOfTotal(newsletterSummary.draftCampaigns, newsletterSummary.sentCampaignsPeriod + newsletterSummary.draftCampaigns, "tracked campaigns"),
        },
        {
            label: "Failed submissions",
            count: failedCampaignSubmissionsPeriod,
            meta: "",
        },
    ];
    $: newsletterHasSubscribers = normalizedSubscribers.length > 0;
    $: newsletterHasCampaignOutput = newsletterSummary.sentCampaignsPeriod > 0 || newsletterSummary.draftCampaigns > 0 || failedCampaignSubmissionsPeriod > 0;
    $: newsletterHasAnyCampaigns = normalizedCampaigns.length > 0;
    $: newsletterSubscriberStatusHasData = hasPositiveTrafficRows(newsletterSubscriberStatusRows);
    $: newsletterCampaignStatusHasData = hasPositiveTrafficRows(newsletterCampaignStatusRows);
    $: maxNewsletterSubscriberStatus = resolveTrafficMaxCount(newsletterSubscriberStatusRows);
    $: maxNewsletterCampaignStatus = resolveTrafficMaxCount(newsletterCampaignStatusRows);
    $: newsletterOperationalInsights = [
        ...(pendingSubscribersCount > 0
            ? [{
                id: "pending-confirmations",
                severity: "warning",
                title: `${formatMetricNumber(pendingSubscribersCount)} pending confirmation${pendingSubscribersCount === 1 ? "" : "s"}`,
                detail: "Review pending subscribers and prompt confirmations to improve active audience growth.",
            }]
            : []),
        ...(!newsletterHasCampaignOutput && newsletterHasSubscribers
            ? [{
                id: "no-campaign-output",
                severity: "warning",
                title: "No campaign output in this period",
                detail: "Consider sending at least one campaign to engage your active subscribers.",
            }]
            : []),
        ...(newsletterHasSubscribers && newsletterSummary.sentCampaignsPeriod > 0
            ? [{
                id: "campaigns-active",
                severity: "success",
                title: "Campaign activity detected",
                detail: `${formatMetricNumber(newsletterSummary.sentCampaignsPeriod)} campaign${newsletterSummary.sentCampaignsPeriod === 1 ? "" : "s"} submitted in this period.`,
            }]
            : []),
        ...(!newsletterHasSubscribers
            ? [{
                id: "no-subscribers",
                severity: "neutral",
                title: "No subscribers recorded",
                detail: "Start capturing newsletter signups to build your subscriber base.",
            }]
            : []),
    ].slice(0, 4);
    $: recentNewsletterCampaigns = [...normalizedCampaigns]
        .filter((campaign) => isTimestampInPeriod(campaign.sentTs || campaign.updatedTs || campaign.createdTs, selectedPeriod))
        .sort((a, b) => (b.sentTs || b.updatedTs || b.createdTs || 0) - (a.sentTs || a.updatedTs || a.createdTs || 0))
        .slice(0, 8);

    $: normalizedPages = normalizePages(pagesRecords, websiteLabelById);
    $: selectedWebsiteSeo = normalizeWebsiteSeo(dashboardResponse?.websiteSeoDefaults || selectedWebsite);
    $: seoSummary = buildSeoSummary(normalizedPages, selectedWebsiteSeo);
    $: seoHeroCards = [
        {
            key: "totalPages",
            label: "Total pages",
            value: formatMetricNumber(seoSummary.totalPages),
            hint: "Pages included in this SEO audit",
            meta: `${formatMetricNumber(seoSummary.good)} page${seoSummary.good === 1 ? "" : "s"} in good shape`,
            icon: "ri-file-list-3-line",
            badgeLabel: seoSummary.totalPages > 0 ? "Audited" : "No pages",
            badgeClass: seoSummary.totalPages > 0 ? "label-success" : "",
        },
        {
            key: "goodPages",
            label: "Pages in good shape",
            value: formatMetricNumber(seoSummary.good),
            hint: "Title, description and sharing image covered",
            meta: formatShareOfTotal(seoSummary.good, seoSummary.totalPages, "pages"),
            icon: "ri-checkbox-circle-line",
            badgeLabel: seoSummary.good > 0 ? "Healthy" : "Needs work",
            badgeClass: seoSummary.good > 0 ? "label-success" : "label-warning",
        },
        {
            key: "missingBasics",
            label: "Pages missing basics",
            value: formatMetricNumber(seoSummary.missingBasics),
            hint: "Missing title or description",
            meta: `${formatMetricNumber(seoSummary.missingTitle)} title | ${formatMetricNumber(seoSummary.missingDescription)} description`,
            icon: "ri-alert-line",
            badgeLabel: seoSummary.missingBasics > 0 ? "Priority fixes" : "Covered",
            badgeClass: seoSummary.missingBasics > 0 ? "label-warning" : "label-success",
        },
        {
            key: "noindexPages",
            label: "Noindex pages",
            value: formatMetricNumber(seoSummary.noindexPages),
            hint: "Pages excluded from search index",
            meta: formatShareOfTotal(seoSummary.noindexPages, seoSummary.totalPages, "pages"),
            icon: "ri-eye-off-line",
            badgeLabel: seoSummary.noindexPages > 0 ? "Review" : "None",
            badgeClass: seoSummary.noindexPages > 0 ? "label-warning" : "label-success",
        },
        {
            key: "needsAttention",
            label: "Pages needing attention",
            value: formatMetricNumber(seoSummary.needsAttention + seoSummary.missingBasics),
            hint: "Pages with missing SEO items or review flags",
            meta: `${formatMetricNumber(seoSummary.missingSocialImage)} missing sharing image`,
            icon: "ri-search-eye-line",
            badgeLabel: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "Action needed" : "Healthy",
            badgeClass: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "label-warning" : "label-success",
        },
    ];
    $: seoIssueBreakdownRows = [
        {
            label: "Missing SEO title",
            count: seoSummary.missingTitle,
            meta: formatShareOfTotal(seoSummary.missingTitle, seoSummary.totalPages, "pages"),
        },
        {
            label: "Missing SEO description",
            count: seoSummary.missingDescription,
            meta: formatShareOfTotal(seoSummary.missingDescription, seoSummary.totalPages, "pages"),
        },
        {
            label: "Missing sharing image",
            count: seoSummary.missingSocialImage,
            meta: formatShareOfTotal(seoSummary.missingSocialImage, seoSummary.totalPages, "pages"),
        },
        {
            label: "Noindex enabled",
            count: seoSummary.noindexPages,
            meta: formatShareOfTotal(seoSummary.noindexPages, seoSummary.totalPages, "pages"),
        },
    ];
    $: seoIssueAuditRows = seoIssueBreakdownRows.map((row) => ({
        ...row,
        description: resolveSeoIssueDescription(row.label),
        severityLabel: resolveSeoIssueSeverityLabel(row.label),
        severityClass: resolveSeoIssueSeverityPillClass(row.label),
    }));
    $: seoIssueBreakdownHasData = hasPositiveTrafficRows(seoIssueBreakdownRows);
    $: maxSeoIssueCount = resolveTrafficMaxCount(seoIssueBreakdownRows);
    $: seoAttentionInsights = [
        ...(seoSummary.missingBasics > 0
            ? [{
                id: "seo-basics",
                severity: "warning",
                title: `${formatMetricNumber(seoSummary.missingBasics)} page${seoSummary.missingBasics === 1 ? "" : "s"} missing SEO basics`,
                detail: "Prioritize pages missing title or description to improve visibility clarity.",
            }]
            : []),
        ...(seoSummary.needsAttention > 0
            ? [{
                id: "seo-attention",
                severity: "warning",
                title: `${formatMetricNumber(seoSummary.needsAttention)} page${seoSummary.needsAttention === 1 ? "" : "s"} need SEO review`,
                detail: "Review sharing image coverage and noindex configuration where applicable.",
            }]
            : []),
        ...(seoSummary.noindexPages > 0
            ? [{
                id: "seo-noindex",
                severity: "neutral",
                title: `${formatMetricNumber(seoSummary.noindexPages)} page${seoSummary.noindexPages === 1 ? "" : "s"} marked noindex`,
                detail: "Confirm noindex is intentional for pages that should stay out of search results.",
            }]
            : []),
        ...(seoSummary.totalPages > 0 && (seoSummary.needsAttention + seoSummary.missingBasics) < 1
            ? [{
                id: "seo-healthy",
                severity: "success",
                title: "No SEO issues detected in this report",
                detail: "Core SEO visibility checks look healthy for the selected website and period.",
            }]
            : []),
        ...(seoSummary.totalPages < 1
            ? [{
                id: "seo-no-pages",
                severity: "neutral",
                title: "No pages available for SEO analysis",
                detail: "Publish pages to start receiving SEO health insights here.",
            }]
            : []),
    ].slice(0, 4);
    $: prioritizedSeoRows = [...seoSummary.pageRows].sort((a, b) => {
        const aPriority = a?.healthKey === "missing-basics" ? 0 : 1;
        const bPriority = b?.healthKey === "missing-basics" ? 0 : 1;
        if (aPriority !== bPriority) {
            return aPriority - bPriority;
        }
        return normalizeString(a?.title).localeCompare(normalizeString(b?.title));
    });

    $: isDashboardDataReady = !!selectedWebsiteId && !isLoadingData && !reportsLoadError;
    $: sourceReadinessRows = [
        {
            label: "Lead data",
            ok: isDashboardDataReady,
            message: isDashboardDataReady
                ? `${formatMetricNumber(normalizedLeadRecords.length)} interaction(s) loaded.`
                : (reportsLoadError ? "Lead data source unavailable." : "Loading lead data."),
        },
        {
            label: "Booking data",
            ok: isDashboardDataReady,
            message: isDashboardDataReady
                ? `${formatMetricNumber(normalizedAppointments.length)} appointment(s) loaded.`
                : (reportsLoadError ? "Booking appointments data source unavailable." : "Loading booking data."),
        },
        {
            label: "Newsletter data",
            ok: isDashboardDataReady,
            message: isDashboardDataReady
                ? `${formatMetricNumber(normalizedSubscribers.length)} subscriber(s) and ${formatMetricNumber(normalizedCampaigns.length)} campaign(s) loaded.`
                : (reportsLoadError ? "Newsletter data sources unavailable." : "Loading newsletter data."),
        },
        {
            label: "SEO pages",
            ok: isDashboardDataReady,
            message: isDashboardDataReady
                ? `${normalizedPages.length} page(s) loaded.`
                : (reportsLoadError ? "Pages data source unavailable." : "Loading SEO pages."),
        },
    ];

    $: reportWarnings = buildReportWarnings({
        leadsSummary,
        bookingSummary,
        newsletterSummary,
        seoSummary,
        trafficState,
        isLoadingTraffic,
    });

    $: reportSuggestions = buildReportSuggestions({
        leadsSummary,
        bookingSummary,
        newsletterSummary,
        seoSummary,
        trafficState,
        isLoadingTraffic,
    });

    $: overviewMetricCards = [
        {
            key: "leads",
            label: "Leads this period",
            value: formatMetricNumber(leadsSummary.total),
            hint: `${formatMetricNumber(leadsSummary.newCount)} new leads in this period`,
            meta: `${formatMetricNumber(leadsSummary.contactCount)} contact | ${formatMetricNumber(leadsSummary.whatsappCount)} WhatsApp | ${formatMetricNumber(leadsSummary.bookingCount)} booking`,
            icon: "ri-mail-line",
            badgeLabel: leadsSummary.newCount > 0 ? `${formatMetricNumber(leadsSummary.newCount)} new` : "Stable",
            badgeClass: leadsSummary.newCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "bookings",
            label: "Booking requests",
            value: formatMetricNumber(bookingSummary.total),
            hint: `${formatMetricNumber(bookingSummary.pendingCount)} pending requests`,
            meta: `${formatMetricNumber(bookingSummary.confirmedCount)} confirmed | ${formatMetricNumber(bookingSummary.cancelledCount)} cancelled`,
            icon: "ri-calendar-check-line",
            badgeLabel: bookingSummary.pendingCount > 0 ? "Needs follow-up" : "On track",
            badgeClass: bookingSummary.pendingCount > 0 ? "label-warning" : "label-success",
        },
        {
            key: "subscribers",
            label: "Active subscribers",
            value: formatMetricNumber(newsletterSummary.activeSubscribers),
            hint: `${formatMetricNumber(newsletterSummary.newSubscribersPeriod)} new in ${selectedPeriodLabel.toLowerCase()}`,
            meta: `${formatMetricNumber(pendingSubscribersCount)} pending confirmations`,
            icon: "ri-user-follow-line",
            badgeLabel: newsletterSummary.newSubscribersPeriod > 0 ? "Growing" : "Steady",
            badgeClass: newsletterSummary.newSubscribersPeriod > 0 ? "label-success" : "",
        },
        {
            key: "campaigns",
            label: "Campaigns sent",
            value: formatMetricNumber(newsletterSummary.sentCampaignsPeriod),
            hint: `${formatMetricNumber(newsletterSummary.recipientsReachedPeriod)} recipients submitted`,
            meta: `${formatMetricNumber(newsletterSummary.draftCampaigns)} drafts`,
            icon: "ri-send-plane-2-line",
            badgeLabel: newsletterSummary.sentCampaignsPeriod > 0 ? "Active" : "No campaigns",
            badgeClass: newsletterSummary.sentCampaignsPeriod > 0 ? "label-success" : "label-warning",
        },
        {
            key: "seoAttention",
            label: "SEO pages needing attention",
            value: formatMetricNumber(seoSummary.needsAttention + seoSummary.missingBasics),
            hint: `${formatMetricNumber(seoSummary.missingTitle)} missing title | ${formatMetricNumber(seoSummary.missingDescription)} missing description`,
            meta: `${formatMetricNumber(seoSummary.missingSocialImage)} missing sharing image`,
            icon: "ri-search-eye-line",
            badgeLabel: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "Needs fixes" : "Healthy",
            badgeClass: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "label-warning" : "label-success",
        },
        ...(trafficState === "ok" && trafficSummary
            ? [{
                key: "trafficVisitors",
                label: "Visitors",
                value: formatMetricNumber(trafficSummary.visitors),
                hint: `${formatMetricNumber(trafficSummary.pageviews)} pageviews`,
                icon: "ri-line-chart-line",
                badgeLabel: trafficOverviewStatusLabel,
                badgeClass: resolveMetricStatePillClass(trafficOverviewStatus),
                meta: `Bounce rate ${formatBounceRate(trafficSummary.bounceRate)}`,
            }]
            : []),
    ];
    $: overviewLeadBookingRows = [
        {
            label: "Contact form leads",
            count: leadsSummary.contactCount,
            meta: formatShareOfTotal(leadsSummary.contactCount, leadsSummary.total, "leads"),
        },
        {
            label: "WhatsApp leads",
            count: leadsSummary.whatsappCount,
            meta: formatShareOfTotal(leadsSummary.whatsappCount, leadsSummary.total, "leads"),
        },
        {
            label: "Booking leads",
            count: leadsSummary.bookingCount,
            meta: formatShareOfTotal(leadsSummary.bookingCount, leadsSummary.total, "leads"),
        },
        {
            label: "Pending booking requests",
            count: bookingSummary.pendingCount,
            meta: formatShareOfTotal(bookingSummary.pendingCount, bookingSummary.total, "requests"),
        },
        {
            label: "Confirmed bookings",
            count: bookingSummary.confirmedCount,
            meta: formatShareOfTotal(bookingSummary.confirmedCount, bookingSummary.total, "requests"),
        },
    ];
    $: overviewNewsletterSeoRows = [
        {
            label: "Active subscribers",
            count: newsletterSummary.activeSubscribers,
            meta: "",
        },
        {
            label: "New subscribers",
            count: newsletterSummary.newSubscribersPeriod,
            meta: selectedPeriodLabel,
        },
        {
            label: "Campaigns submitted",
            count: newsletterSummary.sentCampaignsPeriod,
            meta: `${formatMetricNumber(newsletterSummary.recipientsReachedPeriod)} recipients submitted`,
        },
        {
            label: "Pages needing SEO attention",
            count: seoSummary.needsAttention,
            meta: formatShareOfTotal(seoSummary.needsAttention, seoSummary.totalPages, "pages"),
        },
        {
            label: "Pages missing SEO basics",
            count: seoSummary.missingBasics,
            meta: formatShareOfTotal(seoSummary.missingBasics, seoSummary.totalPages, "pages"),
        },
    ];
    $: overviewTrafficHighlightRows = trafficState === "ok"
        ? [
            {
                label: "Top page",
                value: topTrafficPage
                    ? `${topTrafficPage.page || "/"} | ${formatMetricNumber(topTrafficPage.visitors)} visitors`
                    : "No page data available.",
            },
            {
                label: "Top source",
                value: topTrafficSource
                    ? `${topTrafficSource.source || "Direct"} | ${formatMetricNumber(topTrafficSource.visitors)} visitors`
                    : "No source data available.",
            },
            {
                label: "Top device",
                value: topTrafficDevice
                    ? `${topTrafficDevice.device || "Unknown"} | ${formatMetricNumber(topTrafficDevice.visitors)} visitors`
                    : "No device data available.",
            },
        ]
        : [];
    $: maxLeadBookingRow = resolveTrafficMaxCount(overviewLeadBookingRows);
    $: maxNewsletterSeoRow = resolveTrafficMaxCount(overviewNewsletterSeoRows);
    $: overviewAttentionItems = buildOverviewAttentionItems({
        reportWarnings,
        sourceReadinessRows,
        trafficState,
        isLoadingTraffic,
        trafficInsights,
    });
    $: overviewNextActions = buildOverviewNextActions({
        reportSuggestions,
        bookingSummary,
        leadsSummary,
        seoSummary,
        trafficState,
        isLoadingTraffic,
        newsletterSummary,
        trafficInsights,
    });
    $: overviewHasBackendInsights = Array.isArray(trafficInsights) && trafficInsights.length > 0;
    $: reportHealthState = resolveReportHealthState(
        overviewHasBackendInsights
            ? overviewAttentionItems.length
            : reportWarnings.length,
    );
    $: overviewDataConfidenceRows = buildOverviewDataConfidenceRows({
        sourceReadinessRows,
        trafficDisplayState,
        isLoadingTraffic,
        trafficPeriod,
        selectedPeriodLabel,
        trafficResponse,
        trafficInsights,
        canConfigureTrafficAnalytics,
    });

    $: historyPlaceholderMessage = "Report archive entries will appear here once scheduled snapshots are enabled.";
    $: historyReadySourcesCount = overviewDataConfidenceRows.filter((sourceRow) => normalizeLower(sourceRow?.status) === "ready").length;
    $: historyTotalSourcesCount = overviewDataConfidenceRows.length;
    $: historyArchiveStatusLabel = historyTotalSourcesCount > 0 ? "Awaiting snapshots" : "Checking sources";
    $: historyArchiveStatusClass = historyTotalSourcesCount > 0 ? "label-warning" : "";
    $: latestLeadActivityTs = resolveMaxTimestamp(periodLeadRecords.map((lead) => lead.createdTs));
    $: latestBookingActivityTs = resolveMaxTimestamp(
        periodAppointments.map((appointment) => Math.max(Number(appointment?.createdTs) || 0, Number(appointment?.scheduledTs) || 0)),
    );
    $: latestNewsletterActivityTs = resolveMaxTimestamp([
        ...normalizedSubscribers.map((subscriber) => subscriber.createdTs),
        ...normalizedCampaigns.map((campaign) => Math.max(
            Number(campaign?.sentTs) || 0,
            Number(campaign?.updatedTs) || 0,
            Number(campaign?.createdTs) || 0,
        )),
    ]);
    $: latestSeoActivityTs = resolveMaxTimestamp(
        (pagesRecords || []).map((page) => Math.max(
            toTimestamp(page?.updated),
            toTimestamp(page?.created),
        )),
    );
    $: historyActivityRows = [
        ...(latestLeadActivityTs > 0
            ? [{
                id: "leads",
                title: "Latest lead activity",
                detail: `${formatMetricNumber(leadsSummary.total)} leads tracked in ${selectedPeriodLabel.toLowerCase()}.`,
                timestampLabel: formatDateTime(latestLeadActivityTs),
                statusLabel: leadsSummary.newCount > 0 ? "Follow-up needed" : "Stable",
                statusClass: leadsSummary.newCount > 0 ? "label-warning" : "label-success",
            }]
            : []),
        ...(latestBookingActivityTs > 0
            ? [{
                id: "booking",
                title: "Latest booking activity",
                detail: `${formatMetricNumber(bookingSummary.total)} booking requests tracked in this period.`,
                timestampLabel: formatDateTime(latestBookingActivityTs),
                statusLabel: bookingSummary.pendingCount > 0 ? "Pending requests" : "On track",
                statusClass: bookingSummary.pendingCount > 0 ? "label-warning" : "label-success",
            }]
            : []),
        ...(latestNewsletterActivityTs > 0
            ? [{
                id: "newsletter",
                title: "Latest newsletter activity",
                detail: `${formatMetricNumber(newsletterSummary.sentCampaignsPeriod)} campaigns submitted in this period.`,
                timestampLabel: formatDateTime(latestNewsletterActivityTs),
                statusLabel: newsletterSummary.sentCampaignsPeriod > 0 ? "Campaign activity" : "No campaigns yet",
                statusClass: newsletterSummary.sentCampaignsPeriod > 0 ? "label-success" : "",
            }]
            : []),
        ...(latestSeoActivityTs > 0
            ? [{
                id: "seo",
                title: "Latest SEO audit signal",
                detail: `${formatMetricNumber(seoSummary.needsAttention + seoSummary.missingBasics)} pages currently need review.`,
                timestampLabel: formatDateTime(latestSeoActivityTs),
                statusLabel: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "Needs review" : "Healthy",
                statusClass: seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "label-warning" : "label-success",
            }]
            : []),
    ];
    $: historyHasActivityRows = historyActivityRows.length > 0;

    export async function reload() {
        if (!websitesCollection?.id) {
            return;
        }

        if (!selectedWebsiteId) {
            await loadWebsites();
            return;
        }

        await Promise.all([
            loadDashboardData(),
            loadTrafficData(),
        ]);
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
    }

    function mapPeriodToTrafficPeriod(periodKey) {
        if (periodKey === "lastMonth") {
            return "lastMonth";
        }
        if (periodKey === "last30Days") {
            return "last30Days";
        }
        if (periodKey === "allTime") {
            return "allTime";
        }
        return "thisMonth";
    }

    function resolveTrafficStateMessage(stateKey, options = {}) {
        const normalizedState = normalizeLower(stateKey);
        const isAdminViewer = !!options?.isAdminViewer;

        if (normalizedState === "feature_unavailable") {
            return "Traffic analytics are not available for this website.";
        }

        if (!isAdminViewer) {
            return clientTrafficUnavailableMessage;
        }

        if (normalizedState === "analytics_disabled") {
            return "Traffic analytics are disabled.";
        }

        if (normalizedState === "analytics_not_configured") {
            return "Analytics are not configured yet.";
        }

        if (normalizedState === "provider_unconfigured") {
            return "Analytics provider is not configured on the server yet.";
        }

        if (normalizedState === "provider_auth_missing") {
            return "Analytics provider authentication is not configured on the server yet.";
        }

        if (normalizedState === "provider_auth_error") {
            return "Unable to authenticate with the analytics provider.";
        }

        if (normalizedState === "provider_not_found") {
            return "Analytics website was not found for this configuration.";
        }

        if (normalizedState === "provider_not_implemented" || normalizedState === "provider_unsupported") {
            return "Analytics provider is not connected yet.";
        }

        if (normalizedState === "provider_error") {
            return "Unable to load traffic analytics right now.";
        }

        return "Unable to load traffic analytics right now.";
    }

    function resolveMetricStatePillClass(stateKey) {
        const normalized = normalizeLower(stateKey);
        if (normalized === "ok" || normalized === "ready") {
            return "label-success";
        }
        if (normalized === "partial" || normalized === "loading" || normalized === "check") {
            return "label-warning";
        }
        return "";
    }

    function resolveMetricStateLabel(stateKey, fallback = "Check") {
        const normalized = normalizeLower(stateKey);
        if (normalized === "ok" || normalized === "ready") {
            return "Ready";
        }
        if (normalized === "partial") {
            return "Partial";
        }
        if (normalized === "loading") {
            return "Checking";
        }
        if (normalized === "disabled") {
            return "Disabled";
        }
        if (normalized === "unconfigured") {
            return "Unconfigured";
        }
        if (normalized === "unavailable") {
            return "Unavailable";
        }
        return fallback;
    }

    function hasPositiveTrafficRows(rows = []) {
        if (!Array.isArray(rows)) {
            return false;
        }
        return rows.some((row) => Number(row?.count) > 0);
    }

    function resolveTrafficOverviewStatus(payload = {}) {
        const normalizedState = normalizeLower(payload?.trafficState);
        if (normalizedState === "partial") {
            return "partial";
        }
        if (normalizedState !== "ok") {
            return "check";
        }

        const insights = Array.isArray(payload?.trafficInsights) ? payload.trafficInsights : [];
        const hasPartialInsight = insights.some((insight) => normalizeLower(insight?.id) === "traffic-partial");
        if (hasPartialInsight) {
            return "partial";
        }

        // Legacy fallback: older responses may only signal partial coverage in message text.
        const message = normalizeLower(payload?.trafficMessage);
        if (message.includes("partial")) {
            return "partial";
        }

        return "ready";
    }

    function resolveTrafficOverviewMessage(payload = {}) {
        const state = normalizeLower(payload?.trafficState);
        const status = normalizeLower(payload?.trafficStatus);
        const message = normalizeString(payload?.trafficMessage);

        if (state !== "ok") {
            return "";
        }

        if (status === "partial") {
            return message || "Traffic analytics connected, but some breakdowns are unavailable.";
        }

        return message || "";
    }

    function resolveConversionUnavailableMessage(stateKey, fallbackMessage = "") {
        const normalized = normalizeLower(stateKey);
        if (normalized === "partial") {
            return "Some conversion event breakdowns are unavailable.";
        }
        if (normalized === "disabled") {
            return "Conversion event tracking is disabled.";
        }
        if (normalized === "unconfigured") {
            return "Conversion event tracking is not configured.";
        }
        if (normalized === "ok") {
            return "";
        }
        return normalizeString(fallbackMessage) || "Conversion event data is unavailable.";
    }

    function resolveScrollDepthUnavailableMessage(stateKey, fallbackMessage = "") {
        const normalized = normalizeLower(stateKey);
        if (normalized === "disabled") {
            return "Scroll depth tracking is disabled.";
        }
        if (normalized === "partial") {
            return "Some scroll depth metrics are unavailable.";
        }
        if (normalized === "ok") {
            return "";
        }
        return normalizeString(fallbackMessage) || "Scroll depth data is unavailable.";
    }

    function formatMetricNumber(value) {
        const numeric = Number(value);
        if (!Number.isFinite(numeric)) {
            return "—";
        }
        return Math.round(numeric).toLocaleString();
    }

    function formatShareOfTotal(value, total, noun = "items") {
        const numericValue = Number(value);
        const numericTotal = Number(total);
        if (!Number.isFinite(numericValue) || !Number.isFinite(numericTotal) || numericTotal <= 0) {
            return "";
        }
        const percentage = Math.round((numericValue / numericTotal) * 100);
        return `${percentage}% of ${noun}`;
    }

    function formatBounceRate(value) {
        const numeric = Number(value);
        if (!Number.isFinite(numeric)) {
            return "—";
        }

        const percent = numeric <= 1 ? numeric * 100 : numeric;
        return `${percent.toFixed(1)}%`;
    }

    function formatDurationSeconds(value) {
        const numeric = Number(value);
        if (!Number.isFinite(numeric)) {
            return "—";
        }

        const total = Math.max(0, Math.round(numeric));
        const hours = Math.floor(total / 3600);
        const minutes = Math.floor((total % 3600) / 60);
        const seconds = total % 60;

        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        }

        if (minutes > 0) {
            return `${minutes}m ${seconds}s`;
        }

        return `${seconds}s`;
    }

    function toObject(value) {
        return value && typeof value === "object" && !Array.isArray(value)
            ? value
            : {};
    }

    function resolveFirstStringValue(source, keys = [], fallback = "") {
        const row = toObject(source);

        for (const key of keys) {
            const value = row[key];
            if (typeof value === "string") {
                const normalized = value.trim();
                if (normalized) {
                    return normalized;
                }
            }
        }

        return fallback;
    }

    function resolveFirstNumberValue(source, keys = [], fallback = 0) {
        const row = toObject(source);

        for (const key of keys) {
            const numeric = Number(row[key]);
            if (Number.isFinite(numeric)) {
                return Math.max(0, Math.round(numeric));
            }
        }

        return fallback;
    }

    function normalizeTrafficEntryRows(rawRows) {
        if (!Array.isArray(rawRows)) {
            return [];
        }

        return rawRows.map((row) => ({
            page: resolveFirstStringValue(row, ["page", "path", "name", "label", "dimension"], "/"),
            visitors: resolveFirstNumberValue(row, ["visitors", "count", "value"], 0),
            pageviews: resolveFirstNumberValue(row, ["pageviews", "views"], 0),
        }));
    }

    function normalizeTrafficDimensionRows(rawRows, keyCandidates = [], fallbackLabel = "Unknown") {
        if (!Array.isArray(rawRows)) {
            return [];
        }

        return rawRows.map((row) => ({
            label: resolveFirstStringValue(row, keyCandidates, fallbackLabel),
            visitors: resolveFirstNumberValue(row, ["visitors", "count", "value"], 0),
        }));
    }

    function normalizeTrafficCountRows(rawRows, labelKeys = [], fallbackLabel = "Unknown", countKeys = ["count", "value", "visitors"]) {
        if (!Array.isArray(rawRows)) {
            return [];
        }

        return rawRows
            .map((row) => ({
                label: resolveFirstStringValue(row, labelKeys, fallbackLabel),
                count: resolveFirstNumberValue(row, countKeys, 0),
            }))
            .filter((row) => row.label || row.count > 0);
    }

    function normalizeTrafficDepthRows(rawRows) {
        if (!Array.isArray(rawRows)) {
            return [];
        }

        return rawRows
            .map((row) => ({
                depth: resolveFirstNumberValue(row, ["depth", "threshold", "value"], 0),
                count: resolveFirstNumberValue(row, ["count", "value", "visitors"], 0),
            }))
            .filter((row) => row.depth > 0)
            .sort((a, b) => a.depth - b.depth);
    }

    function buildTrafficPageBreakdownRows(rows = []) {
        if (!Array.isArray(rows)) {
            return [];
        }

        return rows.slice(0, 8).map((row) => {
            const page = resolveFirstStringValue(row, ["page", "path", "pageSlug", "label"], "/");
            const visitors = resolveFirstNumberValue(row, ["visitors", "count", "value"], 0);
            const pageviews = resolveFirstNumberValue(row, ["pageviews", "views"], 0);
            const meta = pageviews > 0
                ? `${formatMetricNumber(pageviews)} pageviews`
                : "";
            return {
                label: page || "/",
                count: visitors,
                meta,
            };
        });
    }

    function buildTrafficSimpleBreakdownRows(rows = [], labelKeys = ["label"], fallbackLabel = "Unknown", countKeys = ["count"]) {
        if (!Array.isArray(rows)) {
            return [];
        }

        return rows.slice(0, 8).map((row) => ({
            label: resolveFirstStringValue(row, labelKeys, fallbackLabel),
            count: resolveFirstNumberValue(row, countKeys, 0),
            meta: "",
        }));
    }

    function resolveTrafficMaxCount(rows = []) {
        let max = 0;
        for (const row of rows || []) {
            const numeric = Number(row?.count);
            if (Number.isFinite(numeric) && numeric > max) {
                max = numeric;
            }
        }
        return max;
    }

    function resolveTrafficBarWidth(value, maxValue) {
        const numeric = Number(value);
        const max = Number(maxValue);
        if (!Number.isFinite(numeric) || numeric <= 0 || !Number.isFinite(max) || max <= 0) {
            return 0;
        }
        const ratio = (numeric / max) * 100;
        return Math.max(4, Math.min(100, Math.round(ratio)));
    }

    function resolveConversionEventLabel(eventName) {
        const normalized = normalizeLower(eventName);
        if (normalized === "contact_form_submitted") {
            return "Contact form submitted";
        }
        if (normalized === "whatsapp_click") {
            return "WhatsApp click";
        }
        if (normalized === "booking_submitted") {
            return "Booking submitted";
        }
        if (normalized === "newsletter_signup") {
            return "Newsletter signup";
        }
        if (normalized === "phone_click") {
            return "Phone click";
        }
        if (normalized === "email_click") {
            return "Email click";
        }
        if (normalized === "directions_click") {
            return "Directions click";
        }
        return normalizeString(eventName) || "Unknown action";
    }

    function stripHtml(value) {
        return `${value || ""}`
            .replace(/<[^>]*>/g, " ")
            .replace(/\s+/g, " ")
            .trim();
    }

    function truncate(value, max = 120) {
        const normalized = normalizeString(value);
        if (!normalized) {
            return "";
        }
        if (normalized.length <= max) {
            return normalized;
        }
        return `${normalized.slice(0, Math.max(0, max - 1)).trimEnd()}...`;
    }

    function resolveReportsFeatureAvailable(website) {
        const settings = normalizeWebsiteSettingsValue(
            website?.[websiteSettingsField] ?? website?.settings,
        );
        const featureFlags = settings?.featureFlags && typeof settings.featureFlags === "object"
            ? settings.featureFlags
            : {};

        if (featureFlags.reports === false) {
            return false;
        }

        return true;
    }

    function resolveWebsiteSettingsField(collection) {
        const allIdentifiers = CommonHelper.getAllCollectionIdentifiers(collection)
            .map((field) => normalizeLower(field));

        if (allIdentifiers.includes("settings")) {
            return "settings";
        }

        return "settings";
    }

    function buildTrafficAnalyticsSetupDraft(analyticsSettings) {
        const settings = analyticsSettings && typeof analyticsSettings === "object"
            ? analyticsSettings
            : {};
        const events = settings?.events && typeof settings.events === "object"
            ? settings.events
            : {};

        return {
            enabled: !!settings.enabled,
            siteId: normalizeString(settings.siteId),
            scriptEnabled: !!settings.scriptEnabled,
            scriptUrl: normalizeString(settings.scriptUrl),
            scrollDepth: !!events.scrollDepth,
        };
    }

    function resolveTrafficAnalyticsSetupMissingReasons(analyticsSettings) {
        const settings = analyticsSettings && typeof analyticsSettings === "object"
            ? analyticsSettings
            : {};
        const missing = [];

        if (normalizeLower(settings.provider) !== "umami") {
            missing.push("Analytics provider must be set to Umami.");
        }
        if (!settings.enabled) {
            missing.push("Enable analytics tracking.");
        }
        if (!normalizeString(settings.siteId)) {
            missing.push("Add the Umami site ID.");
        }

        return missing;
    }

    function resolveCollectionByAliases(aliases = []) {
        const normalizedAliases = aliases.map((alias) => normalizeLower(alias)).filter(Boolean);
        return $collections.find((collection) => normalizedAliases.includes(normalizeLower(collection?.name))) || null;
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
        return CommonHelper.websiteDisplayLabel(website, { missingValue: "" });
    }

    function clearDataRecords() {
        contactsRecords = [];
        whatsappRecords = [];
        appointmentsRecords = [];
        bookingServicesRecords = [];
        subscribersRecords = [];
        campaignsRecords = [];
        pagesRecords = [];
        dashboardResponse = null;
    }

    function clearTrafficRecords() {
        trafficResponse = null;
        isLoadingTraffic = false;
    }

    async function hydrateSelectedWebsiteRecord() {
        if (
            !ApiClient.isAdminSuperuser()
            || !websitesCollection?.id
            || !selectedWebsiteId
            || !websites.some((website) => website?.id === selectedWebsiteId)
        ) {
            return;
        }

        try {
            const fullWebsiteRecord = await ApiClient.collection(websitesCollection.id).getOne(selectedWebsiteId, {
                requestKey: `nuvio_reports_website_${selectedWebsiteId}`,
            });

            if (!fullWebsiteRecord?.id) {
                return;
            }

            websites = websites.map((website) => (
                website?.id === fullWebsiteRecord.id
                    ? fullWebsiteRecord
                    : website
            ));
        } catch (err) {
            ApiClient.error(err, false);
        }
    }

    async function loadWebsites() {
        if (!websitesCollection?.id) {
            websites = [];
            selectedWebsiteId = "";
            return;
        }

        isLoadingWebsites = true;

        try {
            websites = await ApiClient.getBackofficeWebsites({
                requestKey: "nuvio_reports_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                clearDataRecords();
                clearTrafficRecords();
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            }

            await hydrateSelectedWebsiteRecord();
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            clearDataRecords();
            clearTrafficRecords();
            reportsLoadError = "Unable to load reports right now.";
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadDashboardData() {
        if (!selectedWebsiteId) {
            clearDataRecords();
            reportsLoadError = "";
            return;
        }

        const currentToken = ++dataLoadToken;
        isLoadingData = true;
        reportsLoadError = "";

        try {
            const response = await ApiClient.getReportsDashboard({
                websiteId: selectedWebsiteId,
                period: selectedPeriod,
                requestKey: `nuvio_reports_dashboard_${selectedWebsiteId}`,
            });

            if (currentToken !== dataLoadToken) {
                return;
            }

            const datasets = response?.datasets && typeof response.datasets === "object"
                ? response.datasets
                : {};

            dashboardResponse = response && typeof response === "object"
                ? response
                : null;
            contactsRecords = Array.isArray(datasets.contacts) ? datasets.contacts : [];
            whatsappRecords = Array.isArray(datasets.whatsapp) ? datasets.whatsapp : [];
            appointmentsRecords = Array.isArray(datasets.appointments) ? datasets.appointments : [];
            bookingServicesRecords = Array.isArray(datasets.bookingServices) ? datasets.bookingServices : [];
            subscribersRecords = Array.isArray(datasets.subscribers) ? datasets.subscribers : [];
            campaignsRecords = Array.isArray(datasets.campaigns) ? datasets.campaigns : [];
            pagesRecords = Array.isArray(datasets.pages) ? datasets.pages : [];
        } catch (err) {
            if (currentToken !== dataLoadToken) {
                return;
            }

            dashboardResponse = null;
            reportsLoadError = "Unable to load reports right now.";
            ApiClient.error(err);
        } finally {
            if (currentToken === dataLoadToken) {
                isLoadingData = false;
            }
        }
    }

    async function loadTrafficData() {
        if (!selectedWebsiteId || !reportsFeatureAvailable) {
            clearTrafficRecords();
            return;
        }

        const currentToken = ++trafficLoadToken;
        isLoadingTraffic = true;

        try {
            const response = await ApiClient.send("/api/nuvio/reports/traffic", {
                method: "GET",
                query: {
                    websiteId: selectedWebsiteId,
                    period: selectedTrafficPeriod,
                },
                requestKey: `nuvio_reports_traffic_${selectedWebsiteId}_${selectedTrafficPeriod}`,
            });

            if (currentToken !== trafficLoadToken) {
                return;
            }

            trafficResponse = response && typeof response === "object"
                ? response
                : {
                    state: "provider_error",
                    period: {
                        key: selectedTrafficPeriod,
                        label: selectedPeriodLabel,
                    },
                };
        } catch (err) {
            if (currentToken !== trafficLoadToken) {
                return;
            }

            trafficResponse = {
                state: "provider_error",
                period: {
                    key: selectedTrafficPeriod,
                    label: selectedPeriodLabel,
                },
            };
            ApiClient.error(err, false);
        } finally {
            if (currentToken === trafficLoadToken) {
                isLoadingTraffic = false;
            }
        }
    }

    async function saveTrafficAnalyticsSetup() {
        trafficAnalyticsSetupError = "";
        trafficAnalyticsSetupSuccess = "";

        if (!selectedWebsite?.id || !websitesCollection?.id || !websiteSettingsField) {
            trafficAnalyticsSetupError = "Unable to save traffic analytics settings for this website.";
            return;
        }

        const nextSiteId = normalizeString(trafficAnalyticsSetupDraft.siteId);
        if (trafficAnalyticsSetupDraft.enabled && !nextSiteId) {
            trafficAnalyticsSetupError = "Add the Umami site ID to enable traffic analytics.";
            return;
        }

        isSavingTrafficAnalyticsSetup = true;

        try {
            const fullSettings = normalizeWebsiteSettingsValue(selectedWebsite?.[websiteSettingsField]);
            const reportsSettings = fullSettings?.reports && typeof fullSettings.reports === "object"
                ? fullSettings.reports
                : {};
            const analyticsSettings = reportsSettings?.analytics && typeof reportsSettings.analytics === "object"
                ? reportsSettings.analytics
                : {};
            const analyticsEvents = analyticsSettings?.events && typeof analyticsSettings.events === "object"
                ? analyticsSettings.events
                : {};

            const nextSettings = normalizeWebsiteSettingsValue({
                ...fullSettings,
                reports: {
                    ...reportsSettings,
                    analytics: {
                        ...analyticsSettings,
                        provider: "umami",
                        enabled: !!trafficAnalyticsSetupDraft.enabled,
                        siteId: nextSiteId,
                        scriptEnabled: !!trafficAnalyticsSetupDraft.scriptEnabled,
                        scriptUrl: normalizeString(trafficAnalyticsSetupDraft.scriptUrl),
                        events: {
                            ...analyticsEvents,
                            scrollDepth: !!trafficAnalyticsSetupDraft.scrollDepth,
                        },
                    },
                },
            });

            await ApiClient.collection(websitesCollection.id).update(selectedWebsite.id, {
                [websiteSettingsField]: structuredClone(nextSettings),
            });

            trafficAnalyticsSetupSuccess = "Traffic analytics settings saved.";
            await loadWebsites();
            await loadTrafficData();
        } catch (err) {
            ApiClient.error(err);
            trafficAnalyticsSetupError = err?.response?.message || err?.message || "Failed to save traffic analytics settings.";
        } finally {
            isSavingTrafficAnalyticsSetup = false;
        }
    }

    function toTimestamp(value) {
        if (typeof value === "number") {
            return Number.isFinite(value) && value > 0 ? value : 0;
        }

        const raw = normalizeString(value);
        if (!raw) {
            return 0;
        }

        if (/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
            const parsedDateOnly = new Date(`${raw}T00:00:00`).getTime();
            return Number.isNaN(parsedDateOnly) ? 0 : parsedDateOnly;
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const parsed = new Date(normalized).getTime();
        return Number.isNaN(parsed) ? 0 : parsed;
    }

    function formatDateTime(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return "-";
        }

        const timestamp = toTimestamp(value);
        if (!timestamp) {
            return /^\d+$/.test(raw) ? "-" : raw;
        }

        return new Date(timestamp).toLocaleString();
    }

    function formatAppointmentDateTime(dateValue, timeValue) {
        const dateText = normalizeString(dateValue);
        const timeText = normalizeString(timeValue);
        if (!dateText && !timeText) {
            return "-";
        }

        const combinedTs = toTimestamp(`${dateText} ${timeText}`.trim());
        if (combinedTs) {
            return new Date(combinedTs).toLocaleString([], {
                year: "numeric",
                month: "short",
                day: "2-digit",
                hour: "2-digit",
                minute: "2-digit",
            });
        }

        return `${dateText || "-"}${timeText ? ` ${timeText}` : ""}`.trim();
    }

    function resolveMaxTimestamp(values = []) {
        let maxTimestamp = 0;
        for (const value of values || []) {
            const timestamp = Number(value || 0);
            if (Number.isFinite(timestamp) && timestamp > maxTimestamp) {
                maxTimestamp = timestamp;
            }
        }
        return maxTimestamp;
    }

    function getPeriodBounds(periodKey) {
        const now = new Date();
        const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);

        if (periodKey === "allTime") {
            return { start: 0, end: 0 };
        }

        if (periodKey === "thisMonth") {
            const start = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0).getTime();
            return { start, end: 0 };
        }

        if (periodKey === "lastMonth") {
            const start = new Date(now.getFullYear(), now.getMonth() - 1, 1, 0, 0, 0, 0).getTime();
            const end = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0).getTime();
            return { start, end };
        }

        if (periodKey === "last30Days") {
            const start = new Date(startOfToday.getTime() - (29 * 24 * 60 * 60 * 1000)).getTime();
            return { start, end: 0 };
        }

        return { start: 0, end: 0 };
    }

    function isTimestampInPeriod(timestamp, periodKey) {
        const ts = Number(timestamp || 0);
        if (!ts) {
            return false;
        }

        if (periodKey === "allTime") {
            return true;
        }

        const { start, end } = getPeriodBounds(periodKey);
        if (!start && !end) {
            return true;
        }

        if (start && ts < start) {
            return false;
        }

        if (end && ts >= end) {
            return false;
        }

        return true;
    }

    function normalizeStatusKey(value) {
        const normalized = normalizeLower(value);

        if (archivedStatusAliases.includes(normalized)) {
            return "archived";
        }

        if (normalized === "read") {
            return "read";
        }

        if (normalized === "confirmed") {
            return "confirmed";
        }

        if (normalized === "cancelled") {
            return "cancelled";
        }

        if (normalized === "pending") {
            return "pending";
        }

        return "new";
    }

    function normalizeContactLeads(records, websiteLabelMap) {
        return (records || []).map((record) => {
            const channel = normalizeLower(record?.channel);
            const sourceKey = channel === "booking" ? "booking" : "contact";
            const statusKey = normalizeStatusKey(record?.status);
            const websiteId = normalizeString(record?.website);
            const subject = normalizeString(record?.subject);
            const message = stripHtml(record?.message);

            return {
                key: `contact:${normalizeString(record?.id)}`,
                sourceKey,
                statusKey,
                name: normalizeString(record?.name) || "Unknown lead",
                email: normalizeString(record?.email),
                phone: normalizeString(record?.phone),
                subject,
                message,
                notes: normalizeString(record?.notes),
                created: normalizeString(record?.created),
                createdTs: toTimestamp(record?.created),
                websiteId,
                websiteLabel: websiteLabelMap.get(websiteId) || "",
            };
        });
    }

    function normalizeWhatsAppLeads(records, websiteLabelMap) {
        return (records || []).map((record) => {
            const statusKey = normalizeStatusKey(record?.status);
            const websiteId = normalizeString(record?.website);
            const message = stripHtml(record?.message || record?.defaultMessage || "");
            const sourceDetail = normalizeString(record?.source);

            return {
                key: `whatsapp:${normalizeString(record?.id)}`,
                sourceKey: "whatsapp",
                statusKey,
                name: normalizeString(record?.name) || "WhatsApp interaction",
                email: normalizeString(record?.email),
                phone: normalizeString(record?.phone),
                subject: sourceDetail || "WhatsApp interaction",
                message,
                notes: normalizeString(record?.notes),
                created: normalizeString(record?.created),
                createdTs: toTimestamp(record?.created),
                websiteId,
                websiteLabel: websiteLabelMap.get(websiteId) || "",
            };
        });
    }

    function normalizeRelationId(value) {
        if (Array.isArray(value)) {
            return normalizeString(value[0]);
        }

        if (typeof value === "string") {
            const raw = value.trim();
            if (!raw) {
                return "";
            }

            if (raw.startsWith("[")) {
                try {
                    const parsed = JSON.parse(raw);
                    return normalizeRelationId(parsed);
                } catch (_) {
                    return raw;
                }
            }

            return raw;
        }

        return "";
    }

    function normalizeAppointments(records, serviceMap, websiteLabelMap) {
        return (records || []).map((record) => {
            const serviceId = normalizeRelationId(record?.service);
            const websiteId = normalizeRelationId(record?.website);
            const date = normalizeString(record?.date);
            const time = normalizeString(record?.time);
            const statusKey = normalizeStatusKey(record?.status || "pending");
            const scheduledTs = toTimestamp(`${date} ${time}`.trim());

            return {
                id: normalizeString(record?.id),
                name: normalizeString(record?.name) || "Unnamed customer",
                email: normalizeString(record?.email),
                phone: normalizeString(record?.phone),
                notes: normalizeString(record?.notes),
                statusKey: ["pending", "confirmed", "cancelled"].includes(statusKey) ? statusKey : "pending",
                date,
                time,
                scheduledTs,
                created: normalizeString(record?.created),
                createdTs: toTimestamp(record?.created),
                serviceId,
                serviceLabel: serviceMap.get(serviceId) || "Unknown service",
                websiteId,
                websiteLabel: websiteLabelMap.get(websiteId) || "",
            };
        });
    }

    function normalizeSubscribers(records) {
        return (records || []).map((record) => {
            return {
                id: normalizeString(record?.id),
                email: normalizeString(record?.email),
                statusKey: normalizeLower(record?.status),
                createdTs: toTimestamp(record?.created),
            };
        });
    }

    function normalizeCampaigns(records) {
        return (records || []).map((record) => {
            return {
                id: normalizeString(record?.id),
                subject: normalizeString(record?.subject) || "Untitled campaign",
                statusKey: normalizeLower(record?.status),
                recipientsCount: Number(record?.recipientsCount || 0),
                sentTs: toTimestamp(record?.sentAt),
                updatedTs: toTimestamp(record?.updated),
                createdTs: toTimestamp(record?.created),
            };
        });
    }

    function normalizePages(records, websiteLabelMap) {
        return (records || []).map((record) => {
            const websiteId = normalizeRelationId(record?.website);
            return {
                id: normalizeString(record?.id),
                title: normalizeString(record?.title) || normalizeString(record?.name) || "Untitled page",
                slug: normalizeString(record?.slug || record?.path || record?.url || ""),
                websiteId,
                websiteLabel: websiteLabelMap.get(websiteId) || "",
                seoTitle: normalizeString(record?.seo_title || record?.seoTitle),
                seoDescription: stripHtml(record?.seo_description || record?.seoDescription),
                seoSocialImage: hasFileValue(record?.seo_social_image || record?.seoSocialImage),
                seoNoindex: toBoolean(record?.seo_noindex ?? record?.seoNoindex),
            };
        });
    }

    function hasFileValue(value) {
        if (Array.isArray(value)) {
            return value.some((item) => normalizeString(item));
        }

        if (typeof value === "string") {
            const trimmed = value.trim();
            if (!trimmed) {
                return false;
            }
            if (trimmed.startsWith("[")) {
                try {
                    const parsed = JSON.parse(trimmed);
                    return hasFileValue(parsed);
                } catch (_) {
                    return !!trimmed;
                }
            }
            return true;
        }

        return !!value;
    }

    function toBoolean(value) {
        if (typeof value === "boolean") {
            return value;
        }

        if (typeof value === "number") {
            return value === 1;
        }

        const normalized = normalizeLower(value);
        return ["true", "1", "yes", "on"].includes(normalized);
    }

    function normalizeWebsiteSeo(website) {
        return {
            seoTitle: normalizeString(website?.seoTitle || website?.seo_title),
            seoDescription: stripHtml(website?.seoDescription || website?.seo_description),
            seoSocialImage: toBoolean(website?.seoSocialImage) || hasFileValue(website?.seoImage || website?.seo_image),
        };
    }

    function buildSeoSummary(pages, websiteSeo) {
        const summary = {
            totalPages: pages.length,
            good: 0,
            needsAttention: 0,
            missingBasics: 0,
            missingTitle: 0,
            missingDescription: 0,
            missingSocialImage: 0,
            noindexPages: 0,
            pageRows: [],
        };

        for (const page of pages) {
            const hasTitle = !!normalizeString(page?.seoTitle || websiteSeo?.seoTitle);
            const hasDescription = !!normalizeString(page?.seoDescription || websiteSeo?.seoDescription);
            const hasSocialImage = !!(page?.seoSocialImage || websiteSeo?.seoSocialImage);
            const noindex = !!page?.seoNoindex;

            if (!hasTitle) {
                summary.missingTitle += 1;
            }
            if (!hasDescription) {
                summary.missingDescription += 1;
            }
            if (!hasSocialImage) {
                summary.missingSocialImage += 1;
            }
            if (noindex) {
                summary.noindexPages += 1;
            }

            let healthKey = "good";
            const reasons = [];

            if (!hasTitle || !hasDescription) {
                healthKey = "missing-basics";
                if (!hasTitle) {
                    reasons.push("Missing title");
                }
                if (!hasDescription) {
                    reasons.push("Missing description");
                }
            } else if (!hasSocialImage || noindex) {
                healthKey = "needs-attention";
                if (!hasSocialImage) {
                    reasons.push("Missing social image");
                }
                if (noindex) {
                    reasons.push("Noindex enabled");
                }
            }

            if (healthKey === "good") {
                summary.good += 1;
            } else if (healthKey === "needs-attention") {
                summary.needsAttention += 1;
            } else {
                summary.missingBasics += 1;
            }

            if (healthKey !== "good") {
                summary.pageRows.push({
                    id: page.id,
                    title: page.title,
                    slug: normalizeString(page?.slug),
                    healthKey,
                    reasons,
                });
            }
        }

        summary.pageRows = summary.pageRows.slice(0, 10);

        return summary;
    }

    function buildTopServices(appointments) {
        const countByService = new Map();

        for (const appointment of appointments || []) {
            const key = normalizeString(appointment?.serviceLabel) || "Unknown service";
            countByService.set(key, Number(countByService.get(key) || 0) + 1);
        }

        return [...countByService.entries()]
            .map(([label, count]) => ({ label, count }))
            .sort((a, b) => {
                if (b.count !== a.count) {
                    return b.count - a.count;
                }
                return a.label.localeCompare(b.label);
            })
            .slice(0, 6);
    }

    function buildReportWarnings(payload = {}) {
        const warnings = [];

        if (!payload?.isLoadingTraffic && payload?.trafficState !== "ok") {
            warnings.push("Traffic analytics are not configured yet.");
        }

        if (Number(payload?.leadsSummary?.total || 0) < 1) {
            warnings.push("No leads were received in this period.");
        }

        const seoAttention = Number(payload?.seoSummary?.needsAttention || 0) + Number(payload?.seoSummary?.missingBasics || 0);
        if (seoAttention > 0) {
            warnings.push("Some pages need SEO attention.");
        }

        if (Number(payload?.bookingSummary?.pendingCount || 0) > 0) {
            warnings.push("There are pending booking requests waiting for follow-up.");
        }

        return warnings.slice(0, 5);
    }

    function buildReportSuggestions(payload = {}) {
        const suggestions = [];

        if (!payload?.isLoadingTraffic && payload?.trafficState !== "ok") {
            suggestions.push("Configure Analytics in Website Settings > Reports.");
        }

        if (Number(payload?.leadsSummary?.total || 0) < 1) {
            suggestions.push("Use Lead Capture settings to make contact options clear.");
        }

        if ((Number(payload?.seoSummary?.needsAttention || 0) + Number(payload?.seoSummary?.missingBasics || 0)) > 0) {
            suggestions.push("Review SEO pages with missing basics.");
        }

        if (Number(payload?.newsletterSummary?.activeSubscribers || 0) > 0 && Number(payload?.newsletterSummary?.sentCampaignsPeriod || 0) < 1) {
            suggestions.push("Send at least one newsletter campaign in this period.");
        }

        if (Number(payload?.bookingSummary?.upcomingCount || 0) < 1 && Number(payload?.bookingSummary?.total || 0) > 0) {
            suggestions.push("Follow up confirmed bookings to keep pipeline moving.");
        }

        if (!suggestions.length) {
            suggestions.push("Performance looks stable. Keep optimizing one conversion touchpoint this month.");
        }

        return suggestions.slice(0, 5);
    }

    function resolveReportHealthState(warningsCount = 0) {
        const count = Number(warningsCount || 0);

        if (count < 1) {
            return { label: "Healthy", pillClass: "label-success" };
        }

        if (count < 3) {
            return { label: "Needs attention", pillClass: "label-warning" };
        }

        return { label: "Missing basics", pillClass: "label-danger" };
    }

    function scoreOverviewAttentionMessage(message = "") {
        const normalized = normalizeLower(message);
        if (!normalized) {
            return 0;
        }
        if (normalized.includes("pending booking")) {
            return 100;
        }
        if (normalized.includes("no leads")) {
            return 90;
        }
        if (normalized.includes("seo")) {
            return 80;
        }
        if (normalized.includes("analytics")) {
            return 70;
        }
        if (normalized.includes("newsletter")) {
            return 60;
        }
        return 40;
    }

    function dedupeStrings(values = []) {
        const seen = new Set();
        const result = [];
        for (const value of values || []) {
            const text = normalizeString(value);
            if (!text) {
                continue;
            }
            const key = normalizeLower(text);
            if (seen.has(key)) {
                continue;
            }
            seen.add(key);
            result.push(text);
        }
        return result;
    }

    function normalizeInsightSeverity(value) {
        const normalized = normalizeLower(value);
        if (normalized === "high" || normalized === "medium" || normalized === "low" || normalized === "info") {
            return normalized;
        }
        return "info";
    }

    function normalizeInsightArea(value) {
        const normalized = normalizeLower(value);
        if (normalized === "data" || normalized === "conversions" || normalized === "engagement" || normalized === "traffic") {
            return normalized;
        }
        return "traffic";
    }

    function normalizeInsightConfidence(value) {
        const normalized = normalizeLower(value);
        if (normalized === "high" || normalized === "medium" || normalized === "low") {
            return normalized;
        }
        return "low";
    }

    function normalizeInsightId(rawId, fallbackTitle, area, severity) {
        const candidate = normalizeString(rawId) || `${normalizeString(area)}-${normalizeString(severity)}-${normalizeString(fallbackTitle)}`;
        const slug = normalizeLower(candidate)
            .replace(/[^a-z0-9]+/g, "-")
            .replace(/^-+|-+$/g, "")
            .slice(0, 80);
        return slug || "";
    }

    function normalizeInsightEvidence(value) {
        const evidenceRows = Array.isArray(value) ? value : [value];
        const cleaned = [];
        for (const row of evidenceRows) {
            const text = normalizeString(row);
            if (!text) {
                continue;
            }
            cleaned.push(text);
        }
        return dedupeStrings(cleaned);
    }

    function normalizeInsightTargetRoute(value) {
        const route = normalizeString(value);
        if (!route || !route.startsWith("/")) {
            return "";
        }
        if (route.includes("token=")) {
            return "";
        }
        return route;
    }

    function scoreInsightSeverity(severity) {
        if (severity === "high") {
            return 4;
        }
        if (severity === "medium") {
            return 3;
        }
        if (severity === "low") {
            return 2;
        }
        return 1;
    }

    function scoreInsightArea(area) {
        if (area === "data") {
            return 4;
        }
        if (area === "conversions") {
            return 3;
        }
        if (area === "engagement") {
            return 2;
        }
        return 1;
    }

    function normalizeTrafficInsights(value) {
        if (!Array.isArray(value)) {
            return [];
        }

        const seen = new Set();
        const rows = [];

        for (const rawInsight of value) {
            if (!rawInsight || typeof rawInsight !== "object") {
                continue;
            }

            const severity = normalizeInsightSeverity(rawInsight.severity);
            const area = normalizeInsightArea(rawInsight.area);
            const confidence = normalizeInsightConfidence(rawInsight.confidence);
            const title = normalizeString(rawInsight.title);
            const recommendation = normalizeString(rawInsight.recommendation);
            if (!title) {
                continue;
            }

            const evidence = normalizeInsightEvidence(rawInsight.evidence);
            const targetRoute = normalizeInsightTargetRoute(rawInsight.targetRoute);
            const id = normalizeInsightId(rawInsight.id, title, area, severity);
            if (!id || seen.has(id)) {
                continue;
            }
            seen.add(id);

            rows.push({
                id,
                severity,
                area,
                title,
                evidence,
                recommendation,
                confidence,
                targetRoute,
            });
        }

        return rows.sort((a, b) => {
            const severityDiff = scoreInsightSeverity(b.severity) - scoreInsightSeverity(a.severity);
            if (severityDiff !== 0) {
                return severityDiff;
            }
            const areaDiff = scoreInsightArea(b.area) - scoreInsightArea(a.area);
            if (areaDiff !== 0) {
                return areaDiff;
            }
            return a.id.localeCompare(b.id);
        });
    }

    function resolveInsightSeverityLabel(severity) {
        if (severity === "high") {
            return "High";
        }
        if (severity === "medium") {
            return "Medium";
        }
        if (severity === "low") {
            return "Low";
        }
        return "Info";
    }

    function resolveInsightAreaLabel(area) {
        if (area === "data") {
            return "Data";
        }
        if (area === "conversions") {
            return "Conversions";
        }
        if (area === "engagement") {
            return "Engagement";
        }
        return "Traffic";
    }

    function resolveInsightConfidenceLabel(confidence) {
        if (confidence === "high") {
            return "High confidence";
        }
        if (confidence === "medium") {
            return "Medium confidence";
        }
        return "Low confidence";
    }

    function resolveInsightSeverityPillClass(severity) {
        if (severity === "high") {
            return "label-danger";
        }
        if (severity === "medium") {
            return "label-warning";
        }
        if (severity === "low") {
            return "";
        }
        return "label-success";
    }

    function buildOverviewAttentionItems(payload = {}) {
        if (Array.isArray(payload?.trafficInsights) && payload.trafficInsights.length) {
            const prioritized = payload.trafficInsights
                .filter((item) => item.severity === "high" || item.severity === "medium")
                .slice(0, 6);

            return prioritized.map((item) => ({
                ...item,
                source: "insight",
            }));
        }

        const messages = [];

        for (const warning of payload?.reportWarnings || []) {
            messages.push(warning);
        }

        if (!payload?.isLoadingTraffic && payload?.trafficState !== "ok") {
            messages.push("Traffic analytics are not connected yet.");
        }

        for (const sourceRow of payload?.sourceReadinessRows || []) {
            if (!sourceRow?.ok) {
                messages.push(`${sourceRow.label} source needs setup.`);
            }
        }

        return dedupeStrings(messages)
            .sort((a, b) => scoreOverviewAttentionMessage(b) - scoreOverviewAttentionMessage(a))
            .slice(0, 6)
            .map((message, index) => ({
                id: `fallback-attention-${index + 1}`,
                title: message,
                evidence: [],
                recommendation: "",
                severity: "medium",
                area: "traffic",
                confidence: "medium",
                targetRoute: "",
                source: "fallback",
            }));
    }

    function buildOverviewNextActions(payload = {}) {
        if (Array.isArray(payload?.trafficInsights) && payload.trafficInsights.length) {
            const byRecommendation = new Map();
            for (const item of payload.trafficInsights) {
                const recommendation = normalizeString(item?.recommendation);
                if (!recommendation) {
                    continue;
                }
                const key = normalizeLower(recommendation);
                if (byRecommendation.has(key)) {
                    continue;
                }
                byRecommendation.set(key, {
                    id: `insight-action-${item.id}`,
                    text: recommendation,
                    area: item.area,
                    severity: item.severity,
                    confidence: item.confidence,
                    targetRoute: item.targetRoute,
                    source: "insight",
                });
            }

            const rows = [...byRecommendation.values()].slice(0, 6);
            if (rows.length) {
                return rows;
            }
        }

        const actions = [];

        for (const suggestion of payload?.reportSuggestions || []) {
            actions.push(suggestion);
        }

        if (Number(payload?.bookingSummary?.pendingCount || 0) > 0) {
            actions.push("Review pending booking requests.");
        }

        if (Number(payload?.leadsSummary?.newCount || 0) > 0) {
            actions.push("Follow up with new leads.");
        }

        if (Number(payload?.seoSummary?.missingBasics || 0) > 0) {
            actions.push("Fix SEO title and description on pages missing basics.");
        }

        if (!payload?.isLoadingTraffic && payload?.trafficState !== "ok") {
            actions.push("Configure Analytics in Website Settings > Reports.");
        }

        if (
            Number(payload?.newsletterSummary?.activeSubscribers || 0) > 0
            && Number(payload?.newsletterSummary?.sentCampaignsPeriod || 0) < 1
        ) {
            actions.push("Send a newsletter campaign to active subscribers.");
        }

        const deduped = dedupeStrings(actions).slice(0, 6);
        if (deduped.length) {
            return deduped.map((actionText, index) => ({
                id: `fallback-action-${index + 1}`,
                text: actionText,
                area: "",
                severity: "low",
                confidence: "medium",
                targetRoute: "",
                source: "fallback",
            }));
        }

        return [{
            id: "fallback-action-default",
            text: "Performance looks stable. Keep monitoring this report weekly.",
            area: "",
            severity: "info",
            confidence: "medium",
            targetRoute: "",
            source: "fallback",
        }];
    }

    function buildOverviewDataConfidenceRows(payload = {}) {
        const rows = (payload?.sourceReadinessRows || []).map((sourceRow) => ({
            label: sourceRow?.label || "Source",
            ok: !!sourceRow?.ok,
            status: sourceRow?.ok ? "ready" : "check",
            statusLabel: sourceRow?.ok ? "Ready" : "Check",
            message: normalizeString(sourceRow?.message) || "Status unavailable.",
        }));

        if (payload?.isLoadingTraffic) {
            rows.push({
                label: "Traffic analytics",
                ok: false,
                status: "loading",
                statusLabel: "Checking",
                message: "Checking traffic analytics...",
            });
            return rows;
        }

        const trafficReady = payload?.trafficDisplayState === "ok";
        const trafficLabel = payload?.trafficPeriod?.label || payload?.selectedPeriodLabel || "Selected period";
        const partialProviderMessage = normalizeString(payload?.trafficResponse?.message);
        const trafficIsPartial = trafficReady && !!partialProviderMessage;
        const insightIds = new Set(
            (payload?.trafficInsights || []).map((insight) => normalizeString(insight?.id)).filter(Boolean),
        );
        const hasAnalyticsConfigInsight = insightIds.has("analytics-not-configured");
        const hasTrafficPartialInsight = insightIds.has("traffic-partial");
        let trafficMessage = trafficReady
            ? `${trafficLabel} traffic analytics loaded.`
            : resolveTrafficStateMessage(payload?.trafficDisplayState || "", {
                isAdminViewer: !!payload?.canConfigureTrafficAnalytics,
            });

        if (trafficIsPartial) {
            trafficMessage = "Traffic analytics connected, but some breakdowns are unavailable.";
            if (hasTrafficPartialInsight) {
                trafficMessage = "See attention items for details about partial analytics coverage.";
            }
        } else if (!trafficReady && hasAnalyticsConfigInsight) {
            trafficMessage = "See attention items for analytics setup steps.";
        }

        rows.push({
            label: "Traffic analytics",
            ok: trafficReady && !trafficIsPartial,
            status: trafficReady ? (trafficIsPartial ? "partial" : "ready") : "check",
            statusLabel: trafficReady ? (trafficIsPartial ? "Partial" : "Ready") : "Check",
            message: trafficMessage || "Traffic analytics are not connected yet.",
        });

        return rows;
    }

    function resolveLeadSourceLabel(sourceKey) {
        if (sourceKey === "booking") {
            return "Booking";
        }
        if (sourceKey === "whatsapp") {
            return "WhatsApp";
        }
        return "Contact form";
    }

    function resolveLeadStatusLabel(statusKey) {
        if (statusKey === "archived") {
            return "Archived";
        }
        if (statusKey === "read") {
            return "Read";
        }
        return "New";
    }

    function resolveLeadStatusPillClass(statusKey) {
        if (statusKey === "new") {
            return "label-warning";
        }
        if (statusKey === "read") {
            return "label-success";
        }
        return "";
    }

    function resolveAppointmentStatusLabel(statusKey) {
        if (statusKey === "confirmed") {
            return "Confirmed";
        }
        if (statusKey === "cancelled") {
            return "Cancelled";
        }
        return "Pending";
    }

    function resolveAppointmentStatusPillClass(statusKey) {
        if (statusKey === "confirmed") {
            return "label-success";
        }
        if (statusKey === "pending") {
            return "label-warning";
        }
        return "";
    }

    function resolveCampaignStatusLabel(statusKey) {
        const normalized = normalizeLower(statusKey);
        if (normalized === "sent") {
            return "Submitted";
        }
        if (normalized === "draft") {
            return "Draft";
        }
        if (normalized === "scheduled") {
            return "Scheduled";
        }
        if (["failed", "error", "rejected"].includes(normalized)) {
            return "Failed";
        }
        return "Campaign";
    }

    function resolveCampaignStatusPillClass(statusKey) {
        const normalized = normalizeLower(statusKey);
        if (normalized === "sent") {
            return "label-success";
        }
        if (normalized === "scheduled") {
            return "label-warning";
        }
        if (["failed", "error", "rejected"].includes(normalized)) {
            return "label-danger";
        }
        return "";
    }

    function resolveSeoIssueDescription(issueLabel) {
        const normalized = normalizeLower(issueLabel);
        if (normalized.includes("title")) {
            return "Missing page titles reduce clarity in search results.";
        }
        if (normalized.includes("description")) {
            return "Missing descriptions limit search snippet quality.";
        }
        if (normalized.includes("sharing image")) {
            return "Missing sharing images weaken social preview quality.";
        }
        if (normalized.includes("noindex")) {
            return "Noindex pages are intentionally excluded from search indexing.";
        }
        return "Review this SEO issue to improve visibility health.";
    }

    function resolveSeoIssueSeverityLabel(issueLabel) {
        const normalized = normalizeLower(issueLabel);
        if (normalized.includes("title") || normalized.includes("description")) {
            return "High priority";
        }
        if (normalized.includes("sharing image")) {
            return "Medium priority";
        }
        if (normalized.includes("noindex")) {
            return "Review";
        }
        return "Check";
    }

    function resolveSeoIssueSeverityPillClass(issueLabel) {
        const normalized = normalizeLower(issueLabel);
        if (normalized.includes("title") || normalized.includes("description")) {
            return "label-danger";
        }
        if (normalized.includes("sharing image") || normalized.includes("noindex")) {
            return "label-warning";
        }
        return "";
    }

    function resolveSeoReasonPillClass(reasonLabel) {
        const normalized = normalizeLower(reasonLabel);
        if (normalized.includes("title") || normalized.includes("description")) {
            return "label-danger";
        }
        if (normalized.includes("noindex") || normalized.includes("missing")) {
            return "label-warning";
        }
        return "";
    }

    function resolveSeoHealthLabel(healthKey) {
        if (healthKey === "missing-basics") {
            return "Missing basics";
        }
        if (healthKey === "needs-attention") {
            return "Needs attention";
        }
        return "Good";
    }
</script>

<PageWrapper>
    <section class="operations-head panel reports-head m-b-base">
        <div class="head-main">
            <div class="summary-title-wrap">
                <div class="title-row">
                    <h2 class="m-0">Reports</h2>
                    <RefreshButton class="btn-sm" tooltip={{ text: "Refresh", position: "right" }} on:refresh={reload} />
                </div>
                <p class="head-description txt-sm txt-hint m-b-0">Track business performance for this website.</p>
            </div>

            <div class="head-selector operations-website-select">
                <div class="selector-row selector-row--website">
                    <label class="txt-sm txt-hint selector-label m-b-0" for="reports-website-selector">Website</label>
                    <select
                        id="reports-website-selector"
                        class="input input-sm"
                        bind:value={selectedWebsiteId}
                        disabled={isLoadingWebsites || !websites.length}
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

                <div class="selector-row selector-row--period">
                    <label class="txt-sm txt-hint selector-label m-b-0" for="reports-period-selector">Period</label>
                    <select id="reports-period-selector" class="input input-sm" bind:value={selectedPeriod}>
                        {#each periodOptions as periodOption (periodOption.key)}
                            <option value={periodOption.key}>{periodOption.label}</option>
                        {/each}
                    </select>
                </div>
            </div>
        </div>

        <div class="head-tools">
            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-mail-line" />
                    Leads: {leadsSummary.total}
                </span>
                <span class="summary-pill">
                    <i class="ri-calendar-check-line" />
                    Bookings: {bookingSummary.total}
                </span>
                <span class="summary-pill">
                    <i class="ri-user-follow-line" />
                    Active subscribers: {newsletterSummary.activeSubscribers}
                </span>
                <span class="summary-pill">
                    <i class="ri-send-plane-2-line" />
                    Campaigns sent: {newsletterSummary.sentCampaignsPeriod}
                </span>
            </div>
        </div>
    </section>

    {#if $isCollectionsLoading && !websitesCollection}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading reports...</h1>
        </div>
    {:else if !websitesCollection}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>Website data source not found. Reports cannot load without website data.</div>
        </div>
    {:else if !selectedWebsiteId}
        <div class="placeholder-section m-b-base">
            <h1>Select a website to view reports.</h1>
            <p class="txt-sm txt-hint m-b-0">Once selected, live metrics will load automatically.</p>
        </div>
    {:else if !reportsFeatureAvailable}
        <div class="alert alert-warning m-b-base">
            <div class="icon">
                <i class="ri-information-line" />
            </div>
            <div>Reports are disabled for this website in settings.</div>
        </div>
    {:else if isLoadingData}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading report metrics...</h1>
        </div>
    {:else if reportsLoadError}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>{reportsLoadError}</div>
        </div>
    {:else}
        <section class="panel operations-content-panel reports-body m-b-base">
            <div class="tabs-header compact combined left operations-tabs reports-tabs m-b-sm">
                {#each reportsTabs as tab (tab.key)}
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeTab === tab.key}
                        on:click={() => (activeTab = tab.key)}
                    >
                        <i class={`${tab.icon} tab-icon`} aria-hidden="true" />
                        <span class="tab-label">{tab.label}</span>
                    </button>
                {/each}
            </div>

            {#if activeTab === "overview"}
                <div class="reports-overview-layout">
                    <div class="reports-overview-main">
                        <section class="panel reports-overview-section reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">This period in one view</h5>
                                <p class="txt-sm txt-hint m-b-0">Business summary for {selectedPeriodLabel.toLowerCase()}.</p>
                            </div>
                            <div class="reports-kpi-grid reports-overview-kpi-grid reports-kpi-grid--hero">
                                {#each overviewMetricCards as metric (metric.key)}
                                    <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                        <div class="reports-kpi-top">
                                            <span class="reports-kpi-icon" aria-hidden="true">
                                                <i class={metric.icon} />
                                            </span>
                                            {#if metric.badgeLabel}
                                                <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                            {/if}
                                        </div>
                                        <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                        <div class="reports-kpi-value">{metric.value}</div>
                                        <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                        {#if metric.meta}
                                            <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                        {/if}
                                    </article>
                                {/each}
                            </div>
                        </section>

                        <div class="reports-grid-two reports-overview-pulse-grid">
                            <section class="panel reports-breakdown-card reports-section-shell reports-pulse-card">
                                <div class="section-head report-section-head m-b-sm">
                                    <h5 class="m-0">Lead and booking activity</h5>
                                    <p class="txt-sm txt-hint m-b-0">Pipeline movement in this period.</p>
                                </div>
                                <div class="report-bar-list reports-pulse-list">
                                    {#each overviewLeadBookingRows as metricRow (metricRow.label)}
                                        <div class="report-bar-item reports-pulse-item">
                                            <div class="report-bar-head">
                                                <span class="report-bar-label">{metricRow.label}</span>
                                                <strong class="report-bar-value">{formatMetricNumber(metricRow.count)}</strong>
                                            </div>
                                            {#if metricRow.meta}
                                                <div class="txt-xs txt-hint report-bar-meta">{metricRow.meta}</div>
                                            {/if}
                                            {#if maxLeadBookingRow > 0}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(metricRow.count, maxLeadBookingRow)}%;`} /></div>
                                            {/if}
                                        </div>
                                    {/each}
                                </div>
                            </section>

                            <section class="panel reports-breakdown-card reports-section-shell reports-pulse-card">
                                <div class="section-head report-section-head m-b-sm">
                                    <h5 class="m-0">Newsletter and SEO status</h5>
                                    <p class="txt-sm txt-hint m-b-0">Audience growth and visibility health.</p>
                                </div>
                                <div class="report-bar-list reports-pulse-list">
                                    {#each overviewNewsletterSeoRows as metricRow (metricRow.label)}
                                        <div class="report-bar-item reports-pulse-item">
                                            <div class="report-bar-head">
                                                <span class="report-bar-label">{metricRow.label}</span>
                                                <strong class="report-bar-value">{formatMetricNumber(metricRow.count)}</strong>
                                            </div>
                                            {#if metricRow.meta}
                                                <div class="txt-xs txt-hint report-bar-meta">{metricRow.meta}</div>
                                            {/if}
                                            {#if maxNewsletterSeoRow > 0}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(metricRow.count, maxNewsletterSeoRow)}%;`} /></div>
                                            {/if}
                                        </div>
                                    {/each}
                                </div>
                            </section>
                        </div>

                        {#if trafficState === "ok"}
                            <section class="panel reports-breakdown-card reports-section-shell reports-overview-traffic-card">
                                <div class="section-head report-section-head report-section-head--with-meta m-b-sm">
                                    <div class="report-section-main">
                                        <h5 class="m-0">Traffic analytics pulse</h5>
                                        <p class="txt-sm txt-hint m-b-0">{trafficOverviewMessage || (trafficPeriod?.label || selectedPeriodLabel)}</p>
                                    </div>
                                    <span class={`label label-sm ${resolveMetricStatePillClass(trafficOverviewStatus)}`}>{trafficOverviewStatusLabel}</span>
                                </div>
                                <div class="report-metric-grid reports-overview-traffic-metrics">
                                    <div class="report-metric-row"><span>Visitors</span><strong>{formatMetricNumber(trafficSummary?.visitors)}</strong></div>
                                    <div class="report-metric-row"><span>Pageviews</span><strong>{formatMetricNumber(trafficSummary?.pageviews)}</strong></div>
                                    <div class="report-metric-row"><span>Bounce rate</span><strong>{formatBounceRate(trafficSummary?.bounceRate)}</strong></div>
                                    <div class="report-metric-row"><span>Visit duration</span><strong>{formatDurationSeconds(trafficSummary?.visitDurationSeconds)}</strong></div>
                                </div>
                                {#if overviewTrafficHighlightRows.length}
                                    <div class="reports-list reports-overview-traffic-highlights">
                                        {#each overviewTrafficHighlightRows as highlightRow (highlightRow.label)}
                                            <div class="reports-list-item reports-overview-traffic-highlight">
                                                <div class="reports-list-main">
                                                    <span class="txt-xs txt-hint">{highlightRow.label}</span>
                                                    <span class="reports-list-title">{highlightRow.value}</span>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </section>
                        {:else}
                            <section class="panel reports-breakdown-card reports-section-shell reports-overview-traffic-card reports-overview-traffic-card--disabled">
                                <div class="section-head report-section-head report-section-head--with-meta m-b-sm">
                                    <div class="report-section-main">
                                        <h5 class="m-0">Traffic analytics</h5>
                                        <p class="txt-sm txt-hint m-b-0">Traffic analytics are not connected yet.</p>
                                    </div>
                                    <span class={`label label-sm ${resolveMetricStatePillClass(trafficDisplayState)}`}>{resolveMetricStateLabel(trafficDisplayState)}</span>
                                </div>
                                <div class="report-empty-state reports-overview-traffic-empty">
                                    {resolveTrafficStateMessage(trafficDisplayState, { isAdminViewer: canConfigureTrafficAnalytics })}
                                </div>
                                {#if canConfigureTrafficAnalytics}
                                    <p class="txt-xs txt-hint m-b-0">Admins can configure analytics from the Traffic tab setup card.</p>
                                {/if}
                            </section>
                        {/if}
                    </div>

                    <aside class="reports-overview-rail">
                        <section class="panel report-health-panel reports-rail-card reports-rail-card--attention reports-section-shell">
                            <div class="report-health-head">
                                <div class="report-health-main">
                                    <h5 class="m-0">Attention needed now</h5>
                                    <p class="txt-sm txt-hint m-b-0">Priority items that need follow-up in this period.</p>
                                </div>
                                <div class="report-health-meta">
                                    <span class={`label label-sm ${reportHealthState.pillClass}`}>{reportHealthState.label}</span>
                                    <span class="summary-pill">{overviewAttentionItems.length} item(s)</span>
                                </div>
                            </div>
                            {#if overviewAttentionItems.length}
                                {#each overviewAttentionItems as attentionItem (attentionItem.id)}
                                    <div class={`report-health-item reports-rail-item reports-insight-card ${attentionItem.severity === "high" ? "warning" : ""}`}>
                                        <div class="report-health-item-headline">
                                            <span class={`label label-sm report-health-pill ${resolveInsightSeverityPillClass(attentionItem.severity)}`}>
                                                {resolveInsightSeverityLabel(attentionItem.severity)}
                                            </span>
                                            {#if attentionItem.source === "insight"}
                                                <span class="label label-sm">{resolveInsightAreaLabel(attentionItem.area)}</span>
                                                <span class="label label-sm">{resolveInsightConfidenceLabel(attentionItem.confidence)}</span>
                                            {/if}
                                        </div>
                                        <div class="report-health-item-body">
                                            <span class="report-health-item-title">{attentionItem.title}</span>
                                            {#if attentionItem.evidence?.length}
                                                <span class="txt-xs txt-hint reports-rail-evidence">{truncate(attentionItem.evidence[0], 120)}</span>
                                            {/if}
                                        </div>
                                    </div>
                                {/each}
                            {:else}
                                <p class="txt-sm txt-hint m-b-0 reports-rail-empty">No urgent issues detected for this period. Review recommendations and data confidence for lower-priority suggestions.</p>
                            {/if}
                        </section>

                        <section class="panel report-health-panel reports-rail-card reports-rail-card--actions reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Recommended next actions</h5>
                                <p class="txt-sm txt-hint m-b-0">Suggested business actions based on current report signals.</p>
                            </div>
                            {#if overviewNextActions.length}
                                {#each overviewNextActions as nextAction (nextAction.id)}
                                    <div class="report-health-item reports-rail-item reports-action-card">
                                        <div class="report-health-item-headline">
                                            <span class="label label-sm report-health-pill">Next</span>
                                            {#if nextAction.source === "insight"}
                                                <span class="label label-sm">{resolveInsightAreaLabel(nextAction.area)}</span>
                                                <span class="label label-sm">{resolveInsightConfidenceLabel(nextAction.confidence)}</span>
                                            {/if}
                                        </div>
                                        <div class="report-health-item-body">
                                            <span class="report-health-item-title">{nextAction.text}</span>
                                        </div>
                                    </div>
                                {/each}
                            {:else}
                                <p class="txt-sm txt-hint m-b-0 reports-rail-empty">No recommendations yet.</p>
                            {/if}
                        </section>

                        <section class="panel report-sources-panel reports-rail-card reports-rail-card--confidence reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Data confidence</h5>
                                <p class="txt-sm txt-hint m-b-0">Data source coverage and analytics connection status.</p>
                            </div>
                            <div class="reports-mini-stack reports-confidence-list">
                                {#each overviewDataConfidenceRows as sourceRow}
                                    <div class="report-health-item reports-rail-item reports-rail-item--confidence reports-confidence-row">
                                        <div class="report-health-item-headline">
                                            <span class="reports-mini-row-label">{sourceRow.label}</span>
                                            <span
                                                class={`label label-sm ${
                                                    sourceRow.status === "ready"
                                                        ? "label-success"
                                                        : sourceRow.status === "loading"
                                                            ? ""
                                                            : "label-warning"
                                                }`}
                                            >
                                                {sourceRow.statusLabel || (sourceRow.ok ? "Ready" : "Check")}
                                            </span>
                                        </div>
                                        <div class="report-health-item-body">
                                            <span class="txt-xs txt-hint m-b-0 reports-confidence-note">{sourceRow.message}</span>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        </section>
                    </aside>
                </div>
            {:else if activeTab === "leads"}
                <div class="reports-traffic-layout">
                    <section class="panel reports-overview-section reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Leads summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Lead performance and follow-up status for {selectedPeriodLabel.toLowerCase()}.</p>
                        </div>
                        <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                            {#each leadsHeroCards as metric (metric.key)}
                                <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                    <div class="reports-kpi-top">
                                        <span class="reports-kpi-icon" aria-hidden="true">
                                            <i class={metric.icon} />
                                        </span>
                                        {#if metric.badgeLabel}
                                            <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                        {/if}
                                    </div>
                                    <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                    <div class="reports-kpi-value">{metric.value}</div>
                                    <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                    {#if metric.meta}
                                        <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                    {/if}
                                </article>
                            {/each}
                        </div>
                        {#if !leadsHasRecords}
                            <div class="report-empty-state">No leads recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Lead sources</h5>
                            <p class="txt-sm txt-hint m-b-0">Contact form, WhatsApp, and booking channels.</p>
                        </div>
                        {#if leadSourcesHasData}
                            {@const maxLeadSources = resolveTrafficMaxCount(leadSourceBreakdownRows)}
                            <div class="report-bar-list reports-pulse-list">
                                {#each leadSourceBreakdownRows as sourceRow}
                                    <div class="report-bar-item reports-pulse-item">
                                        <div class="report-bar-head">
                                            <span class="report-bar-label">{sourceRow.label}</span>
                                            <strong class="report-bar-value">{formatMetricNumber(sourceRow.count)}</strong>
                                        </div>
                                        {#if sourceRow.meta}
                                            <div class="txt-xs txt-hint report-bar-meta">{sourceRow.meta}</div>
                                        {/if}
                                        <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(sourceRow.count, maxLeadSources)}%;`} /></div>
                                    </div>
                                {/each}
                            </div>
                        {:else}
                            <div class="report-empty-state">No leads recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Follow-up needed</h5>
                            <p class="txt-sm txt-hint m-b-0">New leads and the latest lead activity.</p>
                        </div>
                        <div class="reports-grid-two reports-operational-grid">
                            <article class="report-breakdown-item reports-operational-card">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Lead status</h6>
                                    <span class={`label label-sm ${leadsSummary.newCount > 0 ? "label-warning" : "label-success"}`}>
                                        {leadsSummary.newCount > 0 ? "Action needed" : "On track"}
                                    </span>
                                </div>
                                <div class="report-metric-grid report-metric-grid--compact">
                                    <div class="report-metric-row">
                                        <span>Total leads</span>
                                        <strong>{formatMetricNumber(leadsSummary.total)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>New leads</span>
                                        <strong>{formatMetricNumber(leadsSummary.newCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Read leads</span>
                                        <strong>{formatMetricNumber(leadsSummary.readCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Archived leads</span>
                                        <strong>{formatMetricNumber(leadsSummary.archivedCount)}</strong>
                                    </div>
                                </div>
                                {#if leadsSummary.newCount > 0}
                                    <p class="txt-xs txt-hint m-b-0">
                                        {formatMetricNumber(leadsSummary.newCount)} new lead{leadsSummary.newCount === 1 ? "" : "s"} waiting for follow-up.
                                    </p>
                                {:else if leadsHasRecords}
                                    <p class="txt-xs txt-hint m-b-0">No new leads need follow-up right now.</p>
                                {:else}
                                    <div class="report-empty-state">No leads recorded for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Recent leads</h6>
                                {#if !sortedRecentLeads.length}
                                    <div class="report-empty-state">No leads recorded for this period.</div>
                                {:else}
                                    <div class="reports-list reports-operational-list">
                                        {#each sortedRecentLeads as lead (lead.key)}
                                            <div class="reports-list-item reports-operational-row">
                                                <div class="reports-list-main reports-operational-main">
                                                    <div class="reports-list-title">{lead.name}</div>
                                                    <div class="reports-operational-chip-row">
                                                        <span class="label label-sm">{resolveLeadSourceLabel(lead.sourceKey)}</span>
                                                        <span class={`label label-sm ${resolveLeadStatusPillClass(lead.statusKey)}`}>
                                                            {resolveLeadStatusLabel(lead.statusKey)}
                                                        </span>
                                                        {#if lead.email || lead.phone}
                                                            <span class="txt-xs txt-hint">{lead.email || lead.phone}</span>
                                                        {/if}
                                                    </div>
                                                    {#if lead.subject || lead.message}
                                                        <div class="txt-sm reports-list-snippet">
                                                            {truncate(lead.subject || lead.message, 120)}
                                                        </div>
                                                    {/if}
                                                </div>
                                                <div class="txt-xs txt-hint reports-list-meta reports-operational-time">{formatDateTime(lead.created)}</div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </article>
                        </div>
                    </section>
                </div>
            {:else if activeTab === "booking"}
                <div class="reports-traffic-layout">
                    <section class="panel reports-overview-section reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Booking summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Booking requests and appointment pipeline for {selectedPeriodLabel.toLowerCase()}.</p>
                        </div>
                        <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                            {#each bookingHeroCards as metric (metric.key)}
                                <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                    <div class="reports-kpi-top">
                                        <span class="reports-kpi-icon" aria-hidden="true">
                                            <i class={metric.icon} />
                                        </span>
                                        {#if metric.badgeLabel}
                                            <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                        {/if}
                                    </div>
                                    <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                    <div class="reports-kpi-value">{metric.value}</div>
                                    <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                    {#if metric.meta}
                                        <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                    {/if}
                                </article>
                            {/each}
                        </div>
                        {#if !bookingHasRequests}
                            <div class="report-empty-state">No booking requests recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Status and service demand</h5>
                            <p class="txt-sm txt-hint m-b-0">Booking statuses and most requested services.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Booking status</h6>
                                {#if bookingStatusHasData}
                                    {@const maxBookingStatus = resolveTrafficMaxCount(bookingStatusBreakdownRows)}
                                    <div class="report-bar-list reports-pulse-list">
                                        {#each bookingStatusBreakdownRows as statusRow}
                                            <div class="report-bar-item reports-pulse-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{statusRow.label}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(statusRow.count)}</strong>
                                                </div>
                                                {#if statusRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{statusRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(statusRow.count, maxBookingStatus)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No booking requests recorded for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Top services</h6>
                                {#if bookingServicesHasData}
                                    {@const maxBookingServices = resolveTrafficMaxCount(bookingServiceBreakdownRows)}
                                    <div class="report-bar-list reports-pulse-list">
                                        {#each bookingServiceBreakdownRows as serviceRow}
                                            <div class="report-bar-item reports-pulse-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{truncate(serviceRow.label, 56)}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(serviceRow.count)}</strong>
                                                </div>
                                                {#if serviceRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{serviceRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(serviceRow.count, maxBookingServices)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No booking service data for this period.</div>
                                {/if}
                            </article>
                        </div>
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Operational attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Pending booking requests and upcoming appointments.</p>
                        </div>
                        <div class="reports-grid-two reports-operational-grid">
                            <article class="report-breakdown-item reports-operational-card">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Pending bookings</h6>
                                    <span class={`label label-sm ${bookingSummary.pendingCount > 0 ? "label-warning" : "label-success"}`}>
                                        {bookingSummary.pendingCount > 0 ? "Action needed" : "On track"}
                                    </span>
                                </div>
                                <div class="report-metric-grid report-metric-grid--compact">
                                    <div class="report-metric-row">
                                        <span>Total requests</span>
                                        <strong>{formatMetricNumber(bookingSummary.total)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Pending requests</span>
                                        <strong>{formatMetricNumber(bookingSummary.pendingCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Upcoming appointments</span>
                                        <strong>{formatMetricNumber(bookingSummary.upcomingCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Confirmed bookings</span>
                                        <strong>{formatMetricNumber(bookingSummary.confirmedCount)}</strong>
                                    </div>
                                </div>
                                {#if bookingSummary.pendingCount > 0}
                                    <p class="txt-xs txt-hint m-b-0">
                                        {formatMetricNumber(bookingSummary.pendingCount)} pending booking request{bookingSummary.pendingCount === 1 ? "" : "s"} {bookingSummary.pendingCount === 1 ? "needs" : "need"} review.
                                    </p>
                                {:else if bookingHasRequests}
                                    <p class="txt-xs txt-hint m-b-0">No pending booking requests right now.</p>
                                {:else}
                                    <div class="report-empty-state">No booking requests recorded for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Upcoming appointments</h6>
                                {#if !upcomingAppointments.length}
                                    <div class="report-empty-state">No upcoming appointments right now.</div>
                                {:else}
                                    <div class="reports-list reports-operational-list">
                                        {#each upcomingAppointments as appointment (appointment.id)}
                                            <div class="reports-list-item reports-operational-row">
                                                <div class="reports-list-main reports-operational-main">
                                                    <div class="reports-list-title">{appointment.name}</div>
                                                    <div class="reports-operational-chip-row">
                                                        <span class="label label-sm">{appointment.serviceLabel}</span>
                                                        <span class={`label label-sm ${resolveAppointmentStatusPillClass(appointment.statusKey)}`}>
                                                            {resolveAppointmentStatusLabel(appointment.statusKey)}
                                                        </span>
                                                        {#if appointment.email || appointment.phone}
                                                            <span class="txt-xs txt-hint">{appointment.email || appointment.phone}</span>
                                                        {/if}
                                                    </div>
                                                </div>
                                                <div class="txt-xs txt-hint reports-list-meta reports-operational-time">{formatAppointmentDateTime(appointment.date, appointment.time)}</div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </article>
                        </div>
                    </section>
                </div>
            {:else if activeTab === "newsletter"}
                <div class="reports-traffic-layout">
                    <section class="panel reports-overview-section reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Audience and campaign summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Newsletter growth and campaign output for {selectedPeriodLabel.toLowerCase()}.</p>
                        </div>
                        <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                            {#each newsletterHeroCards as metric (metric.key)}
                                <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                    <div class="reports-kpi-top">
                                        <span class="reports-kpi-icon" aria-hidden="true">
                                            <i class={metric.icon} />
                                        </span>
                                        {#if metric.badgeLabel}
                                            <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                        {/if}
                                    </div>
                                    <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                    <div class="reports-kpi-value">{metric.value}</div>
                                    <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                    {#if metric.meta}
                                        <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                    {/if}
                                </article>
                            {/each}
                        </div>
                        {#if !newsletterHasSubscribers && !newsletterHasAnyCampaigns}
                            <div class="report-empty-state">No newsletter subscribers recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Campaign output</h5>
                            <p class="txt-sm txt-hint m-b-0">Submission activity and provider-accepted recipients for this period.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Subscriber status</h6>
                                {#if newsletterSubscriberStatusHasData}
                                    <div class="report-bar-list reports-pulse-list">
                                        {#each newsletterSubscriberStatusRows as statusRow}
                                            <div class="report-bar-item reports-pulse-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{statusRow.label}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(statusRow.count)}</strong>
                                                </div>
                                                {#if statusRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{statusRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(statusRow.count, maxNewsletterSubscriberStatus)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No newsletter subscribers recorded for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Campaign status mix</h6>
                                {#if newsletterCampaignStatusHasData}
                                    <div class="report-bar-list reports-pulse-list">
                                        {#each newsletterCampaignStatusRows as statusRow}
                                            <div class="report-bar-item reports-pulse-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{statusRow.label}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(statusRow.count)}</strong>
                                                </div>
                                                {#if statusRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{statusRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(statusRow.count, maxNewsletterCampaignStatus)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else if newsletterHasAnyCampaigns}
                                    <div class="report-empty-state">No campaign output in this period.</div>
                                {:else}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {/if}
                            </article>
                        </div>
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Operational attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Pending confirmations and recent campaign activity.</p>
                        </div>
                        <div class="reports-grid-two reports-operational-grid">
                            <article class="report-breakdown-item reports-operational-card">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Newsletter insights</h6>
                                    <span class={`label label-sm ${pendingSubscribersCount > 0 ? "label-warning" : "label-success"}`}>
                                        {pendingSubscribersCount > 0 ? "Needs follow-up" : "Stable"}
                                    </span>
                                </div>
                                {#if newsletterOperationalInsights.length}
                                    <div class="reports-mini-stack">
                                        {#each newsletterOperationalInsights as insightRow (insightRow.id)}
                                            <div class={`report-health-item reports-rail-item reports-insight-card ${insightRow.severity === "warning" ? "warning" : ""}`}>
                                                <div class="report-health-item-headline">
                                                    <span class={`label label-sm report-health-pill ${
                                                        insightRow.severity === "warning"
                                                            ? "label-warning"
                                                            : insightRow.severity === "success"
                                                                ? "label-success"
                                                                : ""
                                                    }`}
                                                    >
                                                        {insightRow.severity === "warning"
                                                            ? "Attention"
                                                            : insightRow.severity === "success"
                                                                ? "Healthy"
                                                                : "Info"}
                                                    </span>
                                                </div>
                                                <div class="report-health-item-body">
                                                    <span class="report-health-item-title">{insightRow.title}</span>
                                                    <span class="txt-xs txt-hint reports-rail-evidence">{insightRow.detail}</span>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No newsletter operations insights for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Recent campaigns</h6>
                                {#if !recentNewsletterCampaigns.length}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {:else}
                                    <div class="reports-list reports-operational-list">
                                        {#each recentNewsletterCampaigns as campaign (campaign.id)}
                                            <div class="reports-list-item reports-operational-row">
                                                <div class="reports-list-main reports-operational-main">
                                                    <div class="reports-list-title">{truncate(campaign.subject, 80)}</div>
                                                    <div class="reports-operational-chip-row">
                                                        <span class={`label label-sm ${resolveCampaignStatusPillClass(campaign.statusKey)}`}>
                                                            {resolveCampaignStatusLabel(campaign.statusKey)}
                                                        </span>
                                                        {#if Number(campaign.recipientsCount || 0) > 0}
                                                            <span class="txt-xs txt-hint">{formatMetricNumber(campaign.recipientsCount)} recipients submitted</span>
                                                        {/if}
                                                    </div>
                                                </div>
                                                <div class="txt-xs txt-hint reports-list-meta reports-operational-time">{formatDateTime(campaign.sentTs || campaign.updatedTs || campaign.createdTs)}</div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </article>
                        </div>
                    </section>
                </div>
            {:else if activeTab === "seo"}
                <div class="reports-traffic-layout">
                    <section class="panel reports-overview-section reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">SEO summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Search visibility health across this website.</p>
                        </div>
                        <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                            {#each seoHeroCards as metric (metric.key)}
                                <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                    <div class="reports-kpi-top">
                                        <span class="reports-kpi-icon" aria-hidden="true">
                                            <i class={metric.icon} />
                                        </span>
                                        {#if metric.badgeLabel}
                                            <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                        {/if}
                                    </div>
                                    <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                    <div class="reports-kpi-value">{metric.value}</div>
                                    <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                    {#if metric.meta}
                                        <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                    {/if}
                                </article>
                            {/each}
                        </div>
                        {#if seoSummary.totalPages < 1}
                            <div class="report-empty-state">No pages available for SEO analysis yet.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Quick SEO fixes</h5>
                            <p class="txt-sm txt-hint m-b-0">Most common page issues affecting search visibility basics.</p>
                        </div>
                        <div class="reports-grid-two reports-operational-grid">
                            <article class="report-breakdown-item reports-operational-card">
                                <h6 class="m-0 report-breakdown-title">Issue distribution</h6>
                                {#if seoIssueBreakdownHasData}
                                    <div class="report-bar-list reports-pulse-list">
                                        {#each seoIssueAuditRows as issueRow}
                                            <div class="report-bar-item reports-pulse-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{issueRow.label}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(issueRow.count)}</strong>
                                                </div>
                                                <div class="reports-operational-chip-row">
                                                    <span class={`label label-sm ${issueRow.severityClass}`}>{issueRow.severityLabel}</span>
                                                </div>
                                                <div class="txt-xs txt-hint report-bar-meta">{issueRow.description}</div>
                                                {#if issueRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{issueRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(issueRow.count, maxSeoIssueCount)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No SEO issues detected for this website.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item reports-operational-card">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">SEO attention</h6>
                                    <span class={`label label-sm ${seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "label-warning" : "label-success"}`}>
                                        {seoSummary.needsAttention + seoSummary.missingBasics > 0 ? "Action needed" : "Healthy"}
                                    </span>
                                </div>
                                {#if seoAttentionInsights.length}
                                    <div class="reports-mini-stack">
                                        {#each seoAttentionInsights as insightRow (insightRow.id)}
                                            <div class={`report-health-item reports-rail-item reports-insight-card ${insightRow.severity === "warning" ? "warning" : ""}`}>
                                                <div class="report-health-item-headline">
                                                    <span class={`label label-sm report-health-pill ${
                                                        insightRow.severity === "warning"
                                                            ? "label-warning"
                                                            : insightRow.severity === "success"
                                                                ? "label-success"
                                                                : ""
                                                    }`}
                                                    >
                                                        {insightRow.severity === "warning"
                                                            ? "Attention"
                                                            : insightRow.severity === "success"
                                                                ? "Healthy"
                                                                : "Info"}
                                                    </span>
                                                </div>
                                                <div class="report-health-item-body">
                                                    <span class="report-health-item-title">{insightRow.title}</span>
                                                    <span class="txt-xs txt-hint reports-rail-evidence">{insightRow.detail}</span>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No SEO insights available for this period.</div>
                                {/if}
                            </article>
                        </div>
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Pages needing attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Prioritized pages for quick SEO fixes.</p>
                        </div>

                        {#if !prioritizedSeoRows.length}
                            <div class="report-empty-state">No SEO issues detected for this website.</div>
                        {:else}
                            <div class="reports-list reports-operational-list">
                                {#each prioritizedSeoRows as pageRow (pageRow.id)}
                                    <div class="reports-list-item reports-operational-row">
                                        <div class="reports-list-main reports-operational-main">
                                            <div class="reports-list-title">{pageRow.title}</div>
                                            <div class="reports-operational-chip-row">
                                                {#if pageRow.slug}
                                                    <span class="txt-xs txt-hint">{pageRow.slug.startsWith("/") ? pageRow.slug : `/${pageRow.slug}`}</span>
                                                {/if}
                                                {#each pageRow.reasons as reason}
                                                    <span class={`label label-sm ${resolveSeoReasonPillClass(reason)}`}>{reason}</span>
                                                {/each}
                                            </div>
                                        </div>
                                        <span class={`label label-sm ${pageRow.healthKey === "missing-basics" ? "label-danger" : "label-warning"}`}>
                                            {resolveSeoHealthLabel(pageRow.healthKey)}
                                        </span>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </section>
                </div>
            {:else if activeTab === "traffic"}
                {#if isLoadingTraffic}
                    <section class="panel reports-breakdown-card reports-section-shell reports-placeholder-panel">
                        <div class="placeholder-section m-b-0">
                            <span class="loader loader-lg" />
                            <h1>Loading traffic analytics...</h1>
                        </div>
                    </section>
                {:else if trafficState === "ok"}
                    <div class="reports-traffic-layout">
                        <section class="panel reports-overview-section reports-section-shell">
                            <div class="section-head report-section-head report-section-head--with-meta m-b-sm">
                                <div class="report-section-main">
                                    <h5 class="m-0">Traffic overview</h5>
                                    <p class="txt-sm txt-hint m-b-0">{trafficPeriod?.label || selectedPeriodLabel}</p>
                                </div>
                                <span class={`label label-sm ${resolveMetricStatePillClass(trafficOverviewStatus)}`}>{trafficOverviewStatusLabel}</span>
                            </div>
                            <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                                {#each trafficHeroCards as metric (metric.key)}
                                    <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                        <div class="reports-kpi-top">
                                            <span class="reports-kpi-icon" aria-hidden="true">
                                                <i class={metric.icon} />
                                            </span>
                                            {#if metric.badgeLabel}
                                                <span class={`label label-sm ${metric.badgeClass || ""}`}>{metric.badgeLabel}</span>
                                            {/if}
                                        </div>
                                        <span class="txt-xs txt-hint reports-kpi-title">{metric.label}</span>
                                        <div class="reports-kpi-value">{metric.value}</div>
                                        <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">{metric.hint}</p>
                                        {#if metric.meta}
                                            <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{metric.meta}</p>
                                        {/if}
                                    </article>
                                {/each}
                            </div>
                            {#if trafficOverviewMessage}
                                <p class="txt-sm txt-hint m-b-0 report-traffic-state-note">
                                    <span class={`label label-sm ${resolveMetricStatePillClass(trafficOverviewStatus)}`}>
                                        {trafficOverviewStatusLabel}
                                    </span>
                                    {trafficOverviewMessage}
                                </p>
                            {/if}
                            {#if trafficNoDataYet}
                                <div class="report-empty-state">
                                    No traffic data yet for this period. Metrics will appear after the public website receives visits.
                                </div>
                                <p class="txt-xs txt-hint m-b-0">{trafficNoDataMessage}</p>
                            {/if}
                        </section>

                        <section class="panel reports-breakdown-card reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Traffic details</h5>
                                <p class="txt-sm txt-hint m-b-0">Core metrics for this period.</p>
                            </div>
                            <div class="report-metric-grid">
                                <div class="report-metric-row">
                                    <span>Visitors</span>
                                    <strong>{formatMetricNumber(trafficVisitorsCount)}</strong>
                                </div>
                                <div class="report-metric-row">
                                    <span>Visits</span>
                                    <strong>{formatMetricNumber(trafficVisitsCount)}</strong>
                                </div>
                                <div class="report-metric-row">
                                    <span>Pageviews</span>
                                    <strong>{formatMetricNumber(trafficPageviewsCount)}</strong>
                                </div>
                                <div class="report-metric-row">
                                    <span>Bounce / interaction rate</span>
                                    <strong>{formatBounceRate(trafficSummary?.bounceRate)}</strong>
                                </div>
                                <div class="report-metric-row">
                                    <span>Average visit duration</span>
                                    <strong>{formatDurationSeconds(trafficSummary?.visitDurationSeconds)}</strong>
                                </div>
                            </div>
                        </section>

                        <section class="panel reports-breakdown-card reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Content performance</h5>
                                <p class="txt-sm txt-hint m-b-0">Top, entry, and exit pages for this period.</p>
                            </div>
                            <div class="reports-grid-three">
                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Top pages</h6>
                                    {#if trafficTopPageRows.length}
                                        {@const maxTopPages = resolveTrafficMaxCount(trafficTopPageRows)}
                                        <div class="report-bar-list">
                                            {#each trafficTopPageRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label || "/", 56)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    {#if row.meta}
                                                        <div class="txt-xs txt-hint report-bar-meta">{row.meta}</div>
                                                    {/if}
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxTopPages)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No top pages data available for this period.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Entry pages</h6>
                                    {#if trafficEntryPageRows.length}
                                        {@const maxEntryPages = resolveTrafficMaxCount(trafficEntryPageRows)}
                                        <div class="report-bar-list">
                                            {#each trafficEntryPageRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label || "/", 56)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    {#if row.meta}
                                                        <div class="txt-xs txt-hint report-bar-meta">{row.meta}</div>
                                                    {/if}
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxEntryPages)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No entry pages data available for this period.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Exit pages</h6>
                                    {#if trafficExitPageRows.length}
                                        {@const maxExitPages = resolveTrafficMaxCount(trafficExitPageRows)}
                                        <div class="report-bar-list">
                                            {#each trafficExitPageRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label || "/", 56)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    {#if row.meta}
                                                        <div class="txt-xs txt-hint report-bar-meta">{row.meta}</div>
                                                    {/if}
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxExitPages)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No exit pages data available for this period.</div>
                                    {/if}
                                </article>
                            </div>
                        </section>

                        <section class="panel reports-breakdown-card reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Acquisition</h5>
                                <p class="txt-sm txt-hint m-b-0">Sources that drove visits in this period.</p>
                            </div>
                            {#if trafficSourceRows.length}
                                {@const maxSources = resolveTrafficMaxCount(trafficSourceRows)}
                                <div class="report-bar-list">
                                    {#each trafficSourceRows as row}
                                        <div class="report-bar-item">
                                            <div class="report-bar-head">
                                                <span class="report-bar-label">{truncate(row.label, 56)}</span>
                                                <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                            </div>
                                            <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxSources)}%;`} /></div>
                                        </div>
                                    {/each}
                                </div>
                            {:else}
                                <div class="report-empty-state">No source data available for this period.</div>
                            {/if}
                        </section>

                        <section class="panel reports-breakdown-card reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Audience and technology</h5>
                                <p class="txt-sm txt-hint m-b-0">Visitor locations, devices, and technology details.</p>
                            </div>
                            <div class="reports-grid-three">
                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Countries</h6>
                                    {#if trafficCountryRows.length}
                                        {@const maxCountries = resolveTrafficMaxCount(trafficCountryRows)}
                                        <div class="report-bar-list">
                                            {#each trafficCountryRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxCountries)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No country data available.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Regions</h6>
                                    {#if trafficRegionRows.length}
                                        {@const maxRegions = resolveTrafficMaxCount(trafficRegionRows)}
                                        <div class="report-bar-list">
                                            {#each trafficRegionRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxRegions)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No region data available.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Cities</h6>
                                    {#if trafficCityRows.length}
                                        {@const maxCities = resolveTrafficMaxCount(trafficCityRows)}
                                        <div class="report-bar-list">
                                            {#each trafficCityRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxCities)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No city data available.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Devices</h6>
                                    {#if trafficDeviceRows.length}
                                        {@const maxDevices = resolveTrafficMaxCount(trafficDeviceRows)}
                                        <div class="report-bar-list">
                                            {#each trafficDeviceRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxDevices)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No device data available.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Browsers</h6>
                                    {#if trafficBrowserRows.length}
                                        {@const maxBrowsers = resolveTrafficMaxCount(trafficBrowserRows)}
                                        <div class="report-bar-list">
                                            {#each trafficBrowserRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxBrowsers)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No browser data available.</div>
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Operating systems</h6>
                                    {#if trafficOperatingSystemRows.length}
                                        {@const maxOperatingSystems = resolveTrafficMaxCount(trafficOperatingSystemRows)}
                                        <div class="report-bar-list">
                                            {#each trafficOperatingSystemRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 42)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxOperatingSystems)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">No operating system data available.</div>
                                    {/if}
                                </article>
                            </div>
                        </section>

                        <section class="panel reports-breakdown-card reports-section-shell">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Tracked actions and scroll depth</h5>
                                <p class="txt-sm txt-hint m-b-0">Conversion events and engagement depth signals.</p>
                            </div>
                            <div class="reports-grid-two">
                                <article class="report-breakdown-item">
                                    <div class="report-breakdown-head">
                                        <h6 class="m-0 report-breakdown-title">Tracked actions</h6>
                                        <span class={`label label-sm ${resolveMetricStatePillClass(trafficConversionsState)}`}>
                                            {resolveMetricStateLabel(trafficConversionsState)}
                                        </span>
                                    </div>
                                    <p class="txt-xs txt-hint m-b-0">{trafficConversionsMessage}</p>
                                    {#if trafficConversionShowMeasuredSummary}
                                        <div class="report-metric-grid report-metric-grid--compact">
                                            <div class="report-metric-row"><span>Total actions</span><strong>{formatMetricNumber(trafficConversionTotalsAllEvents)}</strong></div>
                                            <div class="report-metric-row"><span>Action types</span><strong>{formatMetricNumber(trafficConversionTotalsUniqueEventTypes)}</strong></div>
                                        </div>
                                    {:else if trafficConversionShowKnownSummary}
                                        <div class="report-metric-grid report-metric-grid--compact">
                                            <div class="report-metric-row"><span>Known actions</span><strong>{formatMetricNumber(trafficConversionTotalsAllEvents)}</strong></div>
                                            <div class="report-metric-row"><span>Known action types</span><strong>{formatMetricNumber(trafficConversionTotalsUniqueEventTypes)}</strong></div>
                                        </div>
                                    {:else}
                                        <div class="report-empty-state">{trafficConversionUnavailableMessage}</div>
                                    {/if}
                                    {#if trafficConversionShouldRenderBreakdowns && trafficConversionByTypeRenderableRows.length}
                                        {@const maxConversionTypes = resolveTrafficMaxCount(trafficConversionByTypeRenderableRows)}
                                        {#if trafficConversionIsPartial}
                                            <div class="report-empty-state">Some conversion event breakdowns are unavailable. Showing directional data only.</div>
                                        {/if}
                                        <div class="report-bar-list">
                                            {#each trafficConversionByTypeRenderableRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 48)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxConversionTypes)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        {#if trafficConversionIsOk}
                                            <div class="report-empty-state">No tracked actions recorded for this period.</div>
                                        {:else if trafficConversionShowKnownSummary}
                                            <div class="report-empty-state">{trafficConversionUnavailableMessage}</div>
                                        {/if}
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <div class="report-breakdown-head">
                                        <h6 class="m-0 report-breakdown-title">Scroll depth</h6>
                                        <span class={`label label-sm ${resolveMetricStatePillClass(trafficScrollDepthState)}`}>
                                            {resolveMetricStateLabel(trafficScrollDepthState)}
                                        </span>
                                    </div>
                                    <p class="txt-xs txt-hint m-b-0">{trafficScrollDepthMessage}</p>
                                    {#if trafficScrollDepthShouldRenderBreakdowns && trafficScrollDepthThresholdRenderableRows.length}
                                        {@const maxScrollThreshold = resolveTrafficMaxCount(trafficScrollDepthThresholdRenderableRows)}
                                        {#if trafficScrollDepthIsPartial}
                                            <div class="report-empty-state">Some scroll depth metrics are unavailable. Showing directional data only.</div>
                                        {/if}
                                        <div class="report-bar-list">
                                            {#each trafficScrollDepthThresholdRenderableRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{row.label}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxScrollThreshold)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        {#if trafficScrollDepthIsOk}
                                            <div class="report-empty-state">No scroll depth events recorded for this period.</div>
                                        {:else}
                                            <div class="report-empty-state">{trafficScrollDepthUnavailableMessage}</div>
                                        {/if}
                                    {/if}
                                </article>
                            </div>

                            <div class="reports-grid-three m-t-sm">
                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Actions by page</h6>
                                    {#if trafficConversionShouldRenderBreakdowns && trafficConversionByPageRenderableRows.length}
                                        {@const maxConversionsByPage = resolveTrafficMaxCount(trafficConversionByPageRenderableRows)}
                                        <div class="report-bar-list">
                                            {#each trafficConversionByPageRenderableRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 40)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxConversionsByPage)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        {#if trafficConversionIsOk}
                                            <div class="report-empty-state">No per-page action data available.</div>
                                        {:else}
                                            <div class="report-empty-state">{trafficConversionUnavailableMessage}</div>
                                        {/if}
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Actions by source block</h6>
                                    {#if trafficConversionShouldRenderBreakdowns && trafficConversionBySourceRenderableRows.length}
                                        {@const maxConversionsBySource = resolveTrafficMaxCount(trafficConversionBySourceRenderableRows)}
                                        <div class="report-bar-list">
                                            {#each trafficConversionBySourceRenderableRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 40)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxConversionsBySource)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        {#if trafficConversionIsOk}
                                            <div class="report-empty-state">No source block action data available.</div>
                                        {:else}
                                            <div class="report-empty-state">{trafficConversionUnavailableMessage}</div>
                                        {/if}
                                    {/if}
                                </article>

                                <article class="report-breakdown-item">
                                    <h6 class="m-0 report-breakdown-title">Actions by CTA type</h6>
                                    {#if trafficConversionShouldRenderBreakdowns && trafficConversionByCtaRenderableRows.length}
                                        {@const maxConversionsByCta = resolveTrafficMaxCount(trafficConversionByCtaRenderableRows)}
                                        <div class="report-bar-list">
                                            {#each trafficConversionByCtaRenderableRows as row}
                                                <div class="report-bar-item">
                                                    <div class="report-bar-head">
                                                        <span class="report-bar-label">{truncate(row.label, 40)}</span>
                                                        <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                    </div>
                                                    <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxConversionsByCta)}%;`} /></div>
                                                </div>
                                            {/each}
                                        </div>
                                    {:else}
                                        {#if trafficConversionIsOk}
                                            <div class="report-empty-state">No CTA action data available.</div>
                                        {:else}
                                            <div class="report-empty-state">{trafficConversionUnavailableMessage}</div>
                                        {/if}
                                    {/if}
                                </article>
                            </div>

                            <article class="report-breakdown-item m-t-sm">
                                <h6 class="m-0 report-breakdown-title">Scroll depth by page</h6>
                                {#if trafficScrollDepthShouldRenderBreakdowns && trafficScrollDepthByPageRenderableRows.length}
                                    {@const maxScrollByPage = resolveTrafficMaxCount(trafficScrollDepthByPageRenderableRows)}
                                    <div class="report-bar-list">
                                        {#each trafficScrollDepthByPageRenderableRows as row}
                                            <div class="report-bar-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{truncate(row.label, 56)}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(row.count)}</strong>
                                                </div>
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(row.count, maxScrollByPage)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    {#if trafficScrollDepthIsOk}
                                        <div class="report-empty-state">No per-page scroll depth data available.</div>
                                    {:else}
                                        <div class="report-empty-state">{trafficScrollDepthUnavailableMessage}</div>
                                    {/if}
                                {/if}
                            </article>
                        </section>
                    </div>
                {:else}
                    <section class="panel reports-breakdown-card reports-section-shell reports-placeholder-panel">
                        <div class="section-head report-section-head report-section-head--with-meta m-b-sm">
                            <div class="report-section-main">
                                <h5 class="m-0">Traffic analytics</h5>
                                <p class="txt-sm txt-hint m-b-0">Analytics status for this website.</p>
                            </div>
                            <span class={`label label-sm ${resolveMetricStatePillClass(trafficDisplayState)}`}>
                                {resolveMetricStateLabel(trafficDisplayState)}
                            </span>
                        </div>
                        <div class="report-empty-state">
                            {resolveTrafficStateMessage(trafficDisplayState, { isAdminViewer: canConfigureTrafficAnalytics })}
                        </div>
                        {#if !canConfigureTrafficAnalytics}
                            <p class="txt-xs txt-hint m-b-0">Traffic analytics will become available once this website is configured.</p>
                        {/if}
                    </section>
                    {#if showTrafficAnalyticsSetup}
                        <section class="panel reports-breakdown-card reports-section-shell reports-analytics-setup-card">
                            <div class="section-head report-section-head m-b-sm">
                                <h5 class="m-0">Traffic analytics setup</h5>
                                <p class="txt-sm txt-hint m-b-0">Configure Umami website settings for this website.</p>
                            </div>
                            {#if trafficAnalyticsSetupMissingReasons.length}
                                <div class="report-empty-state reports-analytics-setup-missing">
                                    {#each trafficAnalyticsSetupMissingReasons as missingReason}
                                        <div>{missingReason}</div>
                                    {/each}
                                </div>
                            {/if}
                            <div class="reports-analytics-setup-grid">
                                <label class="reports-analytics-setup-toggle">
                                    <input type="checkbox" bind:checked={trafficAnalyticsSetupDraft.enabled} />
                                    <span>Enable analytics tracking</span>
                                </label>
                                <label class="reports-analytics-setup-field">
                                    <span class="txt-sm txt-hint">Provider</span>
                                    <input class="input input-sm" type="text" value="Umami" disabled />
                                </label>
                                <label class="reports-analytics-setup-field">
                                    <span class="txt-sm txt-hint">Umami site ID</span>
                                    <input
                                        class="input input-sm"
                                        type="text"
                                        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                                        bind:value={trafficAnalyticsSetupDraft.siteId}
                                    />
                                </label>
                                <label class="reports-analytics-setup-toggle">
                                    <input type="checkbox" bind:checked={trafficAnalyticsSetupDraft.scriptEnabled} />
                                    <span>Inject analytics tracking script</span>
                                </label>
                                <label class="reports-analytics-setup-field">
                                    <span class="txt-sm txt-hint">Analytics script URL</span>
                                    <input
                                        class="input input-sm"
                                        type="text"
                                        placeholder="https://umami.example.com/script.js"
                                        bind:value={trafficAnalyticsSetupDraft.scriptUrl}
                                    />
                                </label>
                                <label class="reports-analytics-setup-toggle">
                                    <input type="checkbox" bind:checked={trafficAnalyticsSetupDraft.scrollDepth} />
                                    <span>Track scroll depth events</span>
                                </label>
                            </div>
                            <div class="settings-section-actions m-t-sm">
                                <button
                                    type="button"
                                    class="btn btn-sm"
                                    disabled={isSavingTrafficAnalyticsSetup}
                                    on:click={saveTrafficAnalyticsSetup}
                                >
                                    {isSavingTrafficAnalyticsSetup ? "Saving..." : "Save traffic analytics settings"}
                                </button>
                            </div>
                            {#if trafficAnalyticsSetupError}
                                <p class="txt-danger m-t-8 m-b-0">{trafficAnalyticsSetupError}</p>
                            {/if}
                            {#if trafficAnalyticsSetupSuccess}
                                <p class="txt-success m-t-8 m-b-0">{trafficAnalyticsSetupSuccess}</p>
                            {/if}
                        </section>
                    {/if}
                {/if}
            {:else if activeTab === "history"}
                <div class="reports-traffic-layout">
                    <section class="panel reports-overview-section reports-section-shell">
                        <div class="section-head report-section-head report-section-head--with-meta m-b-sm">
                            <div class="report-section-main">
                                <h5 class="m-0">Report archive</h5>
                                <p class="txt-sm txt-hint m-b-0">Snapshot history and reporting continuity.</p>
                            </div>
                            <span class={`label label-sm ${historyArchiveStatusClass}`}>{historyArchiveStatusLabel}</span>
                        </div>
                        <div class="reports-kpi-grid reports-kpi-grid--hero reports-tab-kpi-grid">
                            <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                <div class="reports-kpi-top">
                                    <span class="reports-kpi-icon" aria-hidden="true"><i class="ri-history-line" /></span>
                                    <span class={`label label-sm ${historyArchiveStatusClass}`}>{historyArchiveStatusLabel}</span>
                                </div>
                                <span class="txt-xs txt-hint reports-kpi-title">Archive status</span>
                                <div class="reports-kpi-value">Planned</div>
                                <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">Scheduled report snapshots are not enabled yet.</p>
                                <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{historyPlaceholderMessage}</p>
                            </article>
                            <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                <div class="reports-kpi-top">
                                    <span class="reports-kpi-icon" aria-hidden="true"><i class="ri-calendar-event-line" /></span>
                                    <span class="label label-sm">Current</span>
                                </div>
                                <span class="txt-xs txt-hint reports-kpi-title">Selected period</span>
                                <div class="reports-kpi-value">{selectedPeriodLabel}</div>
                                <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">Current reporting window in use.</p>
                                <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">{resolveWebsiteLabel(selectedWebsite) || "Website selected"}</p>
                            </article>
                            <article class="panel reports-kpi-card reports-overview-kpi-card reports-kpi-card--hero">
                                <div class="reports-kpi-top">
                                    <span class="reports-kpi-icon" aria-hidden="true"><i class="ri-database-2-line" /></span>
                                    <span class={`label label-sm ${historyReadySourcesCount > 0 ? "label-success" : ""}`}>{historyReadySourcesCount > 0 ? "Ready" : "Checking"}</span>
                                </div>
                                <span class="txt-xs txt-hint reports-kpi-title">Data sources ready</span>
                                <div class="reports-kpi-value">{formatMetricNumber(historyReadySourcesCount)} / {formatMetricNumber(historyTotalSourcesCount)}</div>
                                <p class="txt-sm txt-hint m-b-0 reports-kpi-hint">Sources currently available to build future snapshots.</p>
                                <p class="txt-xs txt-hint m-b-0 reports-kpi-meta">Traffic state: {resolveMetricStateLabel(trafficDisplayState)}</p>
                            </article>
                        </div>
                        <div class="report-empty-state">{historyPlaceholderMessage}</div>
                    </section>

                    <section class="panel reports-breakdown-card reports-section-shell">
                        <div class="section-head report-section-head m-b-sm">
                            <h5 class="m-0">Recent reporting activity</h5>
                            <p class="txt-sm txt-hint m-b-0">Latest data activity across report areas.</p>
                        </div>
                        {#if historyHasActivityRows}
                            <div class="reports-list reports-operational-list">
                                {#each historyActivityRows as historyRow (historyRow.id)}
                                    <div class="reports-list-item reports-operational-row">
                                        <div class="reports-list-main reports-operational-main">
                                            <div class="reports-list-title">{historyRow.title}</div>
                                            <div class="txt-sm reports-list-snippet">{historyRow.detail}</div>
                                            <div class="reports-operational-chip-row">
                                                <span class={`label label-sm ${historyRow.statusClass || ""}`}>{historyRow.statusLabel}</span>
                                            </div>
                                        </div>
                                        <div class="txt-xs txt-hint reports-list-meta reports-operational-time">{historyRow.timestampLabel}</div>
                                    </div>
                                {/each}
                            </div>
                        {:else}
                            <div class="report-empty-state">
                                No report activity has been captured yet. Activity rows will appear once website data starts updating.
                            </div>
                        {/if}
                    </section>
                </div>
            {/if}
        </section>
    {/if}
</PageWrapper>

<style>
    .reports-head.operations-head .head-main {
        gap: 12px;
    }

    .reports-head.operations-head .head-description {
        max-width: 520px;
    }

    .reports-head.operations-head .head-tools {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .reports-head.operations-head .head-tools > * {
        min-width: 0;
    }

    .reports-head.operations-head .head-selector {
        width: min(100%, 760px);
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(220px, 260px);
        align-items: end;
        gap: 8px;
    }

    .reports-head.operations-head .selector-row {
        display: flex;
        align-items: center;
        gap: 8px;
        min-width: 0;
    }

    .reports-head.operations-head .selector-row .input {
        flex: 1 1 auto;
        min-width: 0;
    }

    .reports-head.operations-head .selector-row--website .selector-label {
        min-width: 60px;
    }

    .reports-head.operations-head .selector-row--period .selector-label {
        min-width: 52px;
    }

    .reports-head.operations-head .summary-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        justify-content: flex-end;
    }

    .reports-body {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .reports-tabs {
        width: fit-content;
        max-width: 100%;
        margin-top: 0;
        flex-wrap: wrap;
    }

    .reports-overview-layout {
        display: grid;
        gap: 12px;
        grid-template-columns: minmax(0, 1fr) minmax(300px, 340px);
        align-items: start;
    }

    .reports-overview-main,
    .reports-overview-rail,
    .reports-traffic-layout,
    .report-health-panel {
        display: grid;
        gap: 12px;
    }

    .reports-kpi-grid {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-kpi-grid--hero {
        align-items: stretch;
    }

    .reports-overview-kpi-grid {
        grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
        gap: 10px;
    }

    .reports-tab-kpi-grid {
        grid-template-columns: repeat(auto-fit, minmax(175px, 1fr));
    }

    .reports-kpi-card {
        display: grid;
        gap: 8px;
    }

    .reports-overview-kpi-card {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 11%, var(--baseColor));
        box-shadow: none;
        padding: 12px;
        gap: 7px;
    }

    .reports-kpi-card--hero {
        align-content: start;
        min-height: 148px;
    }

    .reports-kpi-top {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .reports-kpi-icon {
        width: 30px;
        height: 30px;
        border-radius: 999px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 20%, var(--baseColor));
        color: var(--txtPrimaryColor);
        font-size: 15px;
    }

    .reports-kpi-title {
        font-weight: 600;
        line-height: 1.35;
    }

    .reports-kpi-value {
        font-size: 28px;
        line-height: 1.1;
        font-weight: 700;
        color: var(--txtPrimaryColor);
    }

    .reports-kpi-hint {
        line-height: 1.35;
    }

    .reports-kpi-meta {
        line-height: 1.35;
    }

    .reports-overview-section {
        display: grid;
        gap: 10px;
    }

    .reports-section-shell {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        box-shadow: none;
    }

    .report-section-head {
        display: grid;
        gap: 4px;
    }

    .report-section-head--with-meta {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .report-section-main {
        display: grid;
        gap: 4px;
        min-width: 0;
    }

    .reports-overview-pulse-grid .reports-pulse-card {
        align-content: start;
    }

    .reports-pulse-list {
        gap: 8px;
    }

    .reports-pulse-item {
        gap: 5px;
        padding: 8px 9px;
        border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
    }

    .reports-overview-traffic-card {
        gap: 10px;
    }

    .reports-overview-traffic-metrics {
        grid-template-columns: repeat(4, minmax(0, 1fr));
    }

    .reports-overview-traffic-highlights {
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 8px;
    }

    .reports-overview-traffic-highlight {
        padding: 9px 10px;
        border-style: dashed;
    }

    .reports-overview-traffic-highlight .reports-list-main {
        gap: 3px;
    }

    .reports-overview-traffic-highlight .reports-list-title {
        font-size: var(--smFontSize);
        line-height: 1.35;
    }

    .reports-overview-traffic-card--disabled {
        background: color-mix(in srgb, var(--baseAlt1Color) 7%, var(--baseColor));
    }

    .reports-overview-traffic-empty {
        margin-top: -2px;
    }

    .reports-grid-two,
    .reports-grid-three {
        display: grid;
        gap: 10px;
    }

    .reports-grid-two {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-grid-three {
        grid-template-columns: repeat(3, minmax(0, 1fr));
    }

    .reports-breakdown-card {
        display: grid;
        gap: 10px;
        background: color-mix(in srgb, var(--baseAlt1Color) 9%, var(--baseColor));
    }

    .reports-breakdown-card,
    .reports-rail-card {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        box-shadow: none;
    }

    .reports-operational-grid {
        align-items: stretch;
    }

    .reports-operational-card {
        align-content: start;
        gap: 10px;
    }

    .report-breakdown-item {
        display: grid;
        gap: 8px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 9px 10px;
        min-width: 0;
    }

    .report-breakdown-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .report-breakdown-title {
        font-size: var(--baseFontSize);
        line-height: 1.35;
    }

    .report-metric-grid {
        display: grid;
        gap: 7px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .report-metric-grid--compact {
        grid-template-columns: 1fr;
    }

    .report-metric-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        padding: 7px 9px;
        border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseAlt1Color) 14%, var(--baseColor));
        min-width: 0;
    }

    .report-metric-row > span {
        min-width: 0;
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
        line-height: 1.35;
    }

    .report-metric-row > strong {
        flex: 0 0 auto;
        color: var(--txtPrimaryColor);
        font-size: var(--mdFontSize);
        line-height: 1.2;
    }

    .report-traffic-state-note {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .report-empty-state {
        border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 85%, transparent);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseAlt1Color) 8%, var(--baseColor));
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
        line-height: var(--smLineHeight);
        padding: 8px 10px;
    }

    .report-bar-list {
        display: grid;
        gap: 6px;
    }

    .report-bar-item {
        display: grid;
        gap: 4px;
        min-width: 0;
    }

    .report-bar-head {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
        gap: 8px;
        min-width: 0;
    }

    .report-bar-label {
        min-width: 0;
        color: var(--txtPrimaryColor);
        font-size: var(--smFontSize);
        line-height: 1.35;
    }

    .report-bar-value {
        flex: 0 0 auto;
        color: var(--txtPrimaryColor);
        font-size: var(--smFontSize);
        line-height: 1.2;
    }

    .report-bar-meta {
        line-height: 1.35;
    }

    .report-bar-track {
        width: 100%;
        height: 5px;
        border-radius: 999px;
        background: color-mix(in srgb, var(--baseAlt2Color) 72%, transparent);
        overflow: hidden;
    }

    .report-bar-fill {
        display: block;
        height: 100%;
        border-radius: inherit;
        background: color-mix(in srgb, var(--primaryColor) 58%, var(--txtHintColor));
    }

    .reports-mini-stack,
    .reports-list {
        display: grid;
        gap: 8px;
    }

    .reports-mini-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        border: 1px dashed var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 8px 10px;
        background: var(--baseColor);
    }

    .reports-mini-row-label {
        min-width: 0;
    }

    .reports-mini-row--wrap {
        align-items: flex-start;
        flex-direction: column;
    }

    .reports-confidence-note {
        margin: 0;
        line-height: 1.35;
    }

    .reports-list-item {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
    }

    .reports-list-main {
        display: grid;
        gap: 4px;
    }

    .reports-list-title {
        font-weight: 600;
        line-height: 1.3;
    }

    .reports-list-snippet {
        color: var(--txtPrimaryColor);
    }

    .reports-list-meta {
        white-space: nowrap;
    }

    .reports-operational-list {
        gap: 10px;
    }

    .reports-operational-row {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 82%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 7%, var(--baseColor));
    }

    .reports-operational-main {
        gap: 6px;
    }

    .reports-operational-chip-row {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .reports-operational-time {
        text-align: right;
    }

    .reports-rail-card {
        background: color-mix(in srgb, var(--baseAlt1Color) 10%, var(--baseColor));
        gap: 10px;
    }

    .report-health-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .report-health-main {
        display: grid;
        gap: 4px;
    }

    .report-health-meta {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .report-health-item {
        display: grid;
        gap: 8px;
        padding: 9px 10px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        color: var(--txtPrimaryColor);
    }

    .reports-rail-item {
        gap: 6px;
        padding: 9px 10px;
    }

    .report-health-item-headline {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .report-health-item-body {
        display: grid;
        gap: 4px;
    }

    .report-health-item-title {
        line-height: 1.35;
        font-weight: 600;
    }

    .reports-rail-evidence {
        line-height: 1.4;
    }

    .reports-rail-empty {
        padding: 2px 0;
    }

    .report-health-item.warning {
        border-left: 2px solid color-mix(in srgb, var(--warningColor) 55%, var(--baseAlt2Color));
        padding-left: 9px;
    }

    .report-health-pill {
        min-width: 52px;
        justify-content: center;
    }

    .reports-confidence-list {
        gap: 8px;
    }

    .reports-confidence-list .reports-rail-item {
        border-top: 0;
    }

    .reports-insight-card,
    .reports-action-card,
    .reports-confidence-row {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 84%, transparent);
        box-shadow: none;
    }

    .reports-action-card .report-health-item-title {
        font-weight: 600;
    }

    .reports-confidence-row {
        background: color-mix(in srgb, var(--baseAlt1Color) 8%, var(--baseColor));
    }

    .reports-placeholder-panel {
        min-height: 180px;
        display: grid;
        align-content: start;
        gap: 10px;
    }

    .reports-analytics-setup-card {
        gap: 10px;
    }

    .reports-analytics-setup-missing {
        display: grid;
        gap: 4px;
    }

    .reports-analytics-setup-grid {
        display: grid;
        gap: 8px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-analytics-setup-field {
        display: grid;
        gap: 4px;
        min-width: 0;
    }

    .reports-analytics-setup-toggle {
        display: flex;
        align-items: center;
        gap: 8px;
        min-height: 34px;
        border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseAlt1Color) 12%, var(--baseColor));
        padding: 7px 9px;
    }

    .reports-analytics-setup-toggle > span {
        color: var(--txtPrimaryColor);
        font-size: var(--smFontSize);
        line-height: 1.35;
    }

    .reports-analytics-setup-toggle > input {
        margin: 0;
        flex: 0 0 auto;
    }

    @media (max-width: 1220px) {
        .reports-overview-layout {
            grid-template-columns: 1fr;
        }

        .reports-overview-rail {
            grid-template-columns: 1fr;
        }

        .reports-overview-traffic-metrics,
        .reports-overview-traffic-highlights {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 980px) {
        .reports-head.operations-head .head-tools {
            align-items: stretch;
        }

        .reports-head.operations-head .summary-badges {
            justify-content: flex-start;
        }

        .reports-head.operations-head .head-selector {
            grid-template-columns: 1fr;
            width: 100%;
        }

        .reports-grid-two,
        .reports-grid-three,
        .reports-kpi-grid,
        .reports-overview-rail,
        .reports-analytics-setup-grid,
        .reports-overview-traffic-metrics,
        .reports-overview-traffic-highlights {
            grid-template-columns: 1fr;
        }

        .report-metric-grid {
            grid-template-columns: 1fr;
        }

        .reports-list-item {
            flex-direction: column;
            align-items: flex-start;
        }

        .reports-list-meta {
            white-space: normal;
        }

        .reports-operational-time {
            text-align: left;
        }
    }

    @media (max-width: 760px) {
        .reports-head.operations-head .selector-row {
            min-width: 100%;
        }
    }
</style>
