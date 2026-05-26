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

    let isLoadingWebsites = false;
    let isLoadingData = false;
    let isLoadingTraffic = false;
    let reportsLoadError = "";

    let lastWebsitesCollectionId = "";
    let lastDataKey = "";
    let lastTrafficKey = "";
    let dataLoadToken = 0;
    let trafficLoadToken = 0;
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
    $: contactsCollection = resolveCollectionByAliases(["contacts", "contact"]);
    $: whatsappCollection = resolveCollectionByAliases(["whatsapp", "whatsapp_interactions", "whatsapp_clicks"]);
    $: appointmentsCollection = resolveCollectionByAliases(["appointments"]);
    $: bookingServicesCollection = resolveCollectionByAliases(["bookingservices"]);
    $: subscribersCollection = resolveCollectionByAliases(["subscribers"]);
    $: campaignsCollection = resolveCollectionByAliases(["campaigns"]);
    $: pagesCollection = resolveCollectionByAliases(["pages"]);
    $: websiteSettingsField = resolveWebsiteSettingsField(websitesCollection);

    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
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

    $: dataKey = [
        selectedWebsiteId,
        contactsCollection?.id || "",
        whatsappCollection?.id || "",
        appointmentsCollection?.id || "",
        bookingServicesCollection?.id || "",
        subscribersCollection?.id || "",
        campaignsCollection?.id || "",
        pagesCollection?.id || "",
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
    $: leadsSummaryMetricRows = [
        { label: "Total leads", count: leadsSummary.total },
        { label: "New leads", count: leadsSummary.newCount },
        { label: "Read leads", count: leadsSummary.readCount },
        { label: "Archived leads", count: leadsSummary.archivedCount },
    ];
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
    $: bookingSummaryMetricRows = [
        { label: "Total requests", count: bookingSummary.total },
        { label: "Pending bookings", count: bookingSummary.pendingCount },
        { label: "Confirmed bookings", count: bookingSummary.confirmedCount },
        { label: "Cancelled bookings", count: bookingSummary.cancelledCount },
        { label: "Upcoming appointments", count: bookingSummary.upcomingCount },
    ];
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
    $: newsletterAudienceMetricRows = [
        { label: "Active subscribers", count: newsletterSummary.activeSubscribers },
        { label: "Pending confirmations", count: pendingSubscribersCount },
        { label: "Unsubscribed", count: unsubscribedSubscribersCount },
        { label: `New subscribers (${selectedPeriodLabel})`, count: newsletterSummary.newSubscribersPeriod },
    ];
    $: newsletterCampaignMetricRows = [
        { label: "Campaigns submitted", count: newsletterSummary.sentCampaignsPeriod },
        { label: "Recipients submitted", count: newsletterSummary.recipientsReachedPeriod },
        { label: "Draft campaigns", count: newsletterSummary.draftCampaigns },
        { label: "Failed submissions", count: failedCampaignSubmissionsPeriod },
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
    $: newsletterCampaignStatusHasData = hasPositiveTrafficRows(newsletterCampaignStatusRows);

    $: recentSentCampaigns = [...normalizedCampaigns]
        .filter((campaign) => campaign.statusKey === "sent")
        .sort((a, b) => (b.sentTs || b.updatedTs || b.createdTs || 0) - (a.sentTs || a.updatedTs || a.createdTs || 0))
        .slice(0, 8);

    $: normalizedPages = normalizePages(pagesRecords, websiteLabelById);
    $: selectedWebsiteSeo = normalizeWebsiteSeo(selectedWebsite);
    $: seoSummary = buildSeoSummary(normalizedPages, selectedWebsiteSeo);
    $: seoSummaryMetricRows = [
        { label: "Total pages", count: seoSummary.totalPages },
        { label: "Good pages", count: seoSummary.good },
        { label: "Pages needing attention", count: seoSummary.needsAttention },
        { label: "Missing basics", count: seoSummary.missingBasics },
        { label: "Noindex pages", count: seoSummary.noindexPages },
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
    $: seoIssueBreakdownHasData = hasPositiveTrafficRows(seoIssueBreakdownRows);
    $: prioritizedSeoRows = [...seoSummary.pageRows].sort((a, b) => {
        const aPriority = a?.healthKey === "missing-basics" ? 0 : 1;
        const bPriority = b?.healthKey === "missing-basics" ? 0 : 1;
        if (aPriority !== bPriority) {
            return aPriority - bPriority;
        }
        return normalizeString(a?.title).localeCompare(normalizeString(b?.title));
    });

    $: sourceReadinessRows = [
        {
            label: "Lead data",
            ok: !!contactsCollection?.id || !!whatsappCollection?.id,
            message: (!!contactsCollection?.id || !!whatsappCollection?.id)
                ? "Data available."
                : "Lead data source unavailable.",
        },
        {
            label: "Booking data",
            ok: !!appointmentsCollection?.id,
            message: appointmentsCollection?.id
                ? "Appointments data available."
                : "Booking appointments data source unavailable.",
        },
        {
            label: "Newsletter data",
            ok: !!subscribersCollection?.id || !!campaignsCollection?.id,
            message: (!!subscribersCollection?.id || !!campaignsCollection?.id)
                ? "Subscribers and campaigns data available."
                : "Newsletter data sources unavailable.",
        },
        {
            label: "SEO pages",
            ok: !!pagesCollection?.id,
            message: pagesCollection?.id
                ? `${normalizedPages.length} page(s) loaded.`
                : "Pages data source unavailable.",
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
            value: leadsSummary.total,
            hint: `${leadsSummary.contactCount} contact - ${leadsSummary.whatsappCount} WhatsApp - ${leadsSummary.bookingCount} booking`,
            icon: "ri-mail-line",
        },
        {
            key: "bookings",
            label: "Booking requests",
            value: bookingSummary.total,
            hint: `${bookingSummary.pendingCount} pending - ${bookingSummary.confirmedCount} confirmed`,
            icon: "ri-calendar-check-line",
        },
        {
            key: "subscribers",
            label: "Active subscribers",
            value: newsletterSummary.activeSubscribers,
            hint: `${newsletterSummary.newSubscribersPeriod} new in ${selectedPeriodLabel.toLowerCase()}`,
            icon: "ri-user-follow-line",
        },
        {
            key: "campaigns",
            label: "Campaigns sent",
            value: newsletterSummary.sentCampaignsPeriod,
            hint: `${newsletterSummary.recipientsReachedPeriod} recipients submitted`,
            icon: "ri-send-plane-2-line",
        },
        {
            key: "seoAttention",
            label: "SEO pages needing attention",
            value: seoSummary.needsAttention + seoSummary.missingBasics,
            hint: `${seoSummary.missingTitle} missing title - ${seoSummary.missingDescription} missing description`,
            icon: "ri-search-eye-line",
        },
        ...(trafficState === "ok" && trafficSummary
            ? [{
                key: "trafficVisitors",
                label: "Visitors",
                value: formatMetricNumber(trafficSummary.visitors),
                hint: `${formatMetricNumber(trafficSummary.pageviews)} pageviews`,
                icon: "ri-line-chart-line",
            }]
            : []),
    ];
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

    $: historyPlaceholderMessage = "Monthly report history will appear here once automatic snapshots are enabled.";

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
    }

    function clearTrafficRecords() {
        trafficResponse = null;
        isLoadingTraffic = false;
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
                clearDataRecords();
                clearTrafficRecords();
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            }
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
            clearDataRecords();
            reportsLoadError = "";
            return;
        }

        const currentToken = ++dataLoadToken;
        isLoadingData = true;
        reportsLoadError = "";

        try {
            const [
                nextContacts,
                nextWhatsApp,
                nextAppointments,
                nextBookingServices,
                nextSubscribers,
                nextCampaigns,
                nextPages,
            ] = await Promise.all([
                loadRecordsByWebsite(contactsCollection, "nuvio_reports_contacts"),
                loadRecordsByWebsite(whatsappCollection, "nuvio_reports_whatsapp"),
                loadRecordsByWebsite(appointmentsCollection, "nuvio_reports_appointments"),
                loadRecordsByWebsite(bookingServicesCollection, "nuvio_reports_booking_services", "+name"),
                loadRecordsByWebsite(subscribersCollection, "nuvio_reports_subscribers"),
                loadRecordsByWebsite(campaignsCollection, "nuvio_reports_campaigns"),
                loadRecordsByWebsite(pagesCollection, "nuvio_reports_pages", "-updated"),
            ]);

            if (currentToken !== dataLoadToken) {
                return;
            }

            contactsRecords = nextContacts;
            whatsappRecords = nextWhatsApp;
            appointmentsRecords = nextAppointments;
            bookingServicesRecords = nextBookingServices;
            subscribersRecords = nextSubscribers;
            campaignsRecords = nextCampaigns;
            pagesRecords = nextPages;
        } catch (err) {
            if (currentToken !== dataLoadToken) {
                return;
            }

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
            seoSocialImage: hasFileValue(website?.seoImage || website?.seo_image),
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

    function resolveAppointmentStatusLabel(statusKey) {
        if (statusKey === "confirmed") {
            return "Confirmed";
        }
        if (statusKey === "cancelled") {
            return "Cancelled";
        }
        return "Pending";
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
                        <section class="panel reports-overview-section">
                            <div class="section-head m-b-sm">
                                <h5 class="m-0">This period in one view</h5>
                                <p class="txt-sm txt-hint m-b-0">Business summary for {selectedPeriodLabel.toLowerCase()}.</p>
                            </div>
                            <div class="reports-kpi-grid reports-overview-kpi-grid">
                                {#each overviewMetricCards as metric (metric.key)}
                                    <article class="panel reports-kpi-card reports-overview-kpi-card">
                                        <div class="reports-kpi-head">
                                            <span class="txt-sm txt-hint reports-kpi-label">{metric.label}</span>
                                            <i class={metric.icon} aria-hidden="true" />
                                        </div>
                                        <div class="reports-kpi-value">{metric.value}</div>
                                        <div class="txt-sm txt-hint reports-kpi-hint">{metric.hint}</div>
                                    </article>
                                {/each}
                            </div>
                        </section>

                        <div class="reports-grid-two">
                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Lead and booking activity</h5>
                                    <p class="txt-sm txt-hint m-b-0">Pipeline movement in this period.</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Contact form leads</span><strong>{leadsSummary.contactCount}</strong></div>
                                    <div class="reports-mini-row"><span>WhatsApp leads</span><strong>{leadsSummary.whatsappCount}</strong></div>
                                    <div class="reports-mini-row"><span>Booking leads</span><strong>{leadsSummary.bookingCount}</strong></div>
                                    <div class="reports-mini-row"><span>Pending booking requests</span><strong>{bookingSummary.pendingCount}</strong></div>
                                    <div class="reports-mini-row"><span>Confirmed bookings</span><strong>{bookingSummary.confirmedCount}</strong></div>
                                </div>
                            </article>

                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Newsletter and SEO status</h5>
                                    <p class="txt-sm txt-hint m-b-0">Audience growth and visibility health.</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Active subscribers</span><strong>{newsletterSummary.activeSubscribers}</strong></div>
                                    <div class="reports-mini-row"><span>New subscribers</span><strong>{newsletterSummary.newSubscribersPeriod}</strong></div>
                                    <div class="reports-mini-row"><span>Campaigns sent</span><strong>{newsletterSummary.sentCampaignsPeriod}</strong></div>
                                    <div class="reports-mini-row"><span>Pages needing SEO attention</span><strong>{seoSummary.needsAttention}</strong></div>
                                    <div class="reports-mini-row"><span>Pages missing SEO basics</span><strong>{seoSummary.missingBasics}</strong></div>
                                </div>
                            </article>
                        </div>

                        {#if trafficState === "ok"}
                            <section class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Traffic highlights</h5>
                                    <p class="txt-sm txt-hint m-b-0">{trafficPeriod?.label || selectedPeriodLabel}</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Visitors</span><strong>{formatMetricNumber(trafficSummary?.visitors)}</strong></div>
                                    <div class="reports-mini-row"><span>Pageviews</span><strong>{formatMetricNumber(trafficSummary?.pageviews)}</strong></div>
                                    <div class="reports-mini-row"><span>Bounce rate</span><strong>{formatBounceRate(trafficSummary?.bounceRate)}</strong></div>
                                    <div class="reports-mini-row"><span>Visit duration</span><strong>{formatDurationSeconds(trafficSummary?.visitDurationSeconds)}</strong></div>
                                    <div class="reports-mini-row reports-mini-row--wrap">
                                        <span>Top page</span>
                                        <span class="txt-sm txt-hint">
                                            {#if topTrafficPage}
                                                {topTrafficPage.page || "/"} - {formatMetricNumber(topTrafficPage.visitors)} visitors
                                            {:else}
                                                No page data available.
                                            {/if}
                                        </span>
                                    </div>
                                    <div class="reports-mini-row reports-mini-row--wrap">
                                        <span>Top source</span>
                                        <span class="txt-sm txt-hint">
                                            {#if topTrafficSource}
                                                {topTrafficSource.source || "Direct"} - {formatMetricNumber(topTrafficSource.visitors)} visitors
                                            {:else}
                                                No source data available.
                                            {/if}
                                        </span>
                                    </div>
                                    <div class="reports-mini-row reports-mini-row--wrap">
                                        <span>Top device</span>
                                        <span class="txt-sm txt-hint">
                                            {#if topTrafficDevice}
                                                {topTrafficDevice.device || "Unknown"} - {formatMetricNumber(topTrafficDevice.visitors)} visitors
                                            {:else}
                                                No device data available.
                                            {/if}
                                        </span>
                                    </div>
                                </div>
                            </section>
                        {:else}
                            <section class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Traffic analytics</h5>
                                    <p class="txt-sm txt-hint m-b-0">Traffic analytics are not connected yet.</p>
                                </div>
                                <div class="empty-state m-b-0">
                                    {resolveTrafficStateMessage(trafficDisplayState, { isAdminViewer: canConfigureTrafficAnalytics })}
                                </div>
                            </section>
                        {/if}
                    </div>

                    <aside class="reports-overview-rail">
                        <section class="panel report-health-panel reports-rail-card reports-rail-card--attention">
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
                                    <div class={`report-health-item reports-rail-item ${attentionItem.severity === "high" ? "warning" : ""}`}>
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

                        <section class="panel report-health-panel reports-rail-card reports-rail-card--actions">
                            <div class="section-head m-b-sm">
                                <h5 class="m-0">Recommended next actions</h5>
                                <p class="txt-sm txt-hint m-b-0">Suggested business actions based on current report signals.</p>
                            </div>
                            {#if overviewNextActions.length}
                                {#each overviewNextActions as nextAction (nextAction.id)}
                                    <div class="report-health-item reports-rail-item">
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

                        <section class="panel report-sources-panel reports-rail-card reports-rail-card--confidence">
                            <div class="section-head m-b-sm">
                                <h5 class="m-0">Data confidence</h5>
                                <p class="txt-sm txt-hint m-b-0">Data source coverage and analytics connection status.</p>
                            </div>
                            <div class="reports-mini-stack reports-confidence-list">
                                {#each overviewDataConfidenceRows as sourceRow}
                                    <div class="report-health-item reports-rail-item reports-rail-item--confidence">
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
                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Leads summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="report-metric-grid">
                            {#each leadsSummaryMetricRows as metricRow (metricRow.label)}
                                <div class="report-metric-row">
                                    <span>{metricRow.label}</span>
                                    <strong>{formatMetricNumber(metricRow.count)}</strong>
                                </div>
                            {/each}
                        </div>
                        {#if !leadsHasRecords}
                            <div class="report-empty-state">No leads recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Lead sources</h5>
                            <p class="txt-sm txt-hint m-b-0">Contact form, WhatsApp, and booking channels.</p>
                        </div>
                        {#if leadSourcesHasData}
                            {@const maxLeadSources = resolveTrafficMaxCount(leadSourceBreakdownRows)}
                            <div class="report-bar-list">
                                {#each leadSourceBreakdownRows as sourceRow}
                                    <div class="report-bar-item">
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

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Follow-up needed</h5>
                            <p class="txt-sm txt-hint m-b-0">New leads and the latest lead activity.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Lead status</h6>
                                    <span class={`label label-sm ${leadsSummary.newCount > 0 ? "label-warning" : "label-success"}`}>
                                        {leadsSummary.newCount > 0 ? "Action needed" : "On track"}
                                    </span>
                                </div>
                                <div class="report-metric-grid report-metric-grid--compact">
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

                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Recent leads</h6>
                                {#if !sortedRecentLeads.length}
                                    <div class="report-empty-state">No leads recorded for this period.</div>
                                {:else}
                                    <div class="reports-list">
                                        {#each sortedRecentLeads as lead (lead.key)}
                                            <div class="reports-list-item">
                                                <div class="reports-list-main">
                                                    <div class="reports-list-title">{lead.name}</div>
                                                    <div class="txt-sm txt-hint">
                                                        {resolveLeadSourceLabel(lead.sourceKey)} - {resolveLeadStatusLabel(lead.statusKey)}
                                                        {#if lead.email || lead.phone}
                                                            - {lead.email || lead.phone}
                                                        {/if}
                                                    </div>
                                                    {#if lead.subject || lead.message}
                                                        <div class="txt-sm reports-list-snippet">
                                                            {truncate(lead.subject || lead.message, 120)}
                                                        </div>
                                                    {/if}
                                                </div>
                                                <div class="txt-xs txt-hint reports-list-meta">{formatDateTime(lead.created)}</div>
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
                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Booking summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="report-metric-grid">
                            {#each bookingSummaryMetricRows as metricRow (metricRow.label)}
                                <div class="report-metric-row">
                                    <span>{metricRow.label}</span>
                                    <strong>{formatMetricNumber(metricRow.count)}</strong>
                                </div>
                            {/each}
                        </div>
                        {#if !bookingHasRequests}
                            <div class="report-empty-state">No booking requests recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Status and service demand</h5>
                            <p class="txt-sm txt-hint m-b-0">Booking statuses and most requested services.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Booking status</h6>
                                {#if bookingStatusHasData}
                                    {@const maxBookingStatus = resolveTrafficMaxCount(bookingStatusBreakdownRows)}
                                    <div class="report-bar-list">
                                        {#each bookingStatusBreakdownRows as statusRow}
                                            <div class="report-bar-item">
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

                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Top services</h6>
                                {#if bookingServicesHasData}
                                    {@const maxBookingServices = resolveTrafficMaxCount(bookingServiceBreakdownRows)}
                                    <div class="report-bar-list">
                                        {#each bookingServiceBreakdownRows as serviceRow}
                                            <div class="report-bar-item">
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

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Operational attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Pending booking requests and upcoming appointments.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Pending bookings</h6>
                                    <span class={`label label-sm ${bookingSummary.pendingCount > 0 ? "label-warning" : "label-success"}`}>
                                        {bookingSummary.pendingCount > 0 ? "Action needed" : "On track"}
                                    </span>
                                </div>
                                <div class="report-metric-grid report-metric-grid--compact">
                                    <div class="report-metric-row">
                                        <span>Pending requests</span>
                                        <strong>{formatMetricNumber(bookingSummary.pendingCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Upcoming appointments</span>
                                        <strong>{formatMetricNumber(bookingSummary.upcomingCount)}</strong>
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

                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Upcoming appointments</h6>
                                {#if !upcomingAppointments.length}
                                    <div class="report-empty-state">No upcoming appointments right now.</div>
                                {:else}
                                    <div class="reports-list">
                                        {#each upcomingAppointments as appointment (appointment.id)}
                                            <div class="reports-list-item">
                                                <div class="reports-list-main">
                                                    <div class="reports-list-title">{appointment.name}</div>
                                                    <div class="txt-sm txt-hint">
                                                        {appointment.serviceLabel} - {resolveAppointmentStatusLabel(appointment.statusKey)}
                                                        {#if appointment.email || appointment.phone}
                                                            - {appointment.email || appointment.phone}
                                                        {/if}
                                                    </div>
                                                </div>
                                                <div class="txt-xs txt-hint reports-list-meta">{formatAppointmentDateTime(appointment.date, appointment.time)}</div>
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
                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Audience summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="report-metric-grid">
                            {#each newsletterAudienceMetricRows as metricRow (metricRow.label)}
                                <div class="report-metric-row">
                                    <span>{metricRow.label}</span>
                                    <strong>{formatMetricNumber(metricRow.count)}</strong>
                                </div>
                            {/each}
                        </div>
                        {#if !newsletterHasSubscribers}
                            <div class="report-empty-state">No newsletter subscribers recorded for this period.</div>
                        {/if}
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Campaign output</h5>
                            <p class="txt-sm txt-hint m-b-0">Submission activity and provider-accepted recipients for this period.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Campaign metrics</h6>
                                <div class="report-metric-grid report-metric-grid--compact">
                                    {#each newsletterCampaignMetricRows as metricRow (metricRow.label)}
                                        <div class="report-metric-row">
                                            <span>{metricRow.label}</span>
                                            <strong>{formatMetricNumber(metricRow.count)}</strong>
                                        </div>
                                    {/each}
                                </div>
                                {#if !newsletterHasCampaignOutput}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Campaign status mix</h6>
                                {#if newsletterCampaignStatusHasData}
                                    {@const maxNewsletterStatus = resolveTrafficMaxCount(newsletterCampaignStatusRows)}
                                    <div class="report-bar-list">
                                        {#each newsletterCampaignStatusRows as statusRow}
                                            <div class="report-bar-item">
                                                <div class="report-bar-head">
                                                    <span class="report-bar-label">{statusRow.label}</span>
                                                    <strong class="report-bar-value">{formatMetricNumber(statusRow.count)}</strong>
                                                </div>
                                                {#if statusRow.meta}
                                                    <div class="txt-xs txt-hint report-bar-meta">{statusRow.meta}</div>
                                                {/if}
                                                <div class="report-bar-track"><span class="report-bar-fill" style={`width: ${resolveTrafficBarWidth(statusRow.count, maxNewsletterStatus)}%;`} /></div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {/if}
                            </article>
                        </div>
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Operational attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Pending confirmations and recent campaign activity.</p>
                        </div>
                        <div class="reports-grid-two">
                            <article class="report-breakdown-item">
                                <div class="report-breakdown-head">
                                    <h6 class="m-0 report-breakdown-title">Pending confirmations</h6>
                                    <span class={`label label-sm ${pendingSubscribersCount > 0 ? "label-warning" : "label-success"}`}>
                                        {pendingSubscribersCount > 0 ? "Action needed" : "On track"}
                                    </span>
                                </div>
                                <div class="report-metric-grid report-metric-grid--compact">
                                    <div class="report-metric-row">
                                        <span>Pending confirmations</span>
                                        <strong>{formatMetricNumber(pendingSubscribersCount)}</strong>
                                    </div>
                                    <div class="report-metric-row">
                                        <span>Campaigns submitted</span>
                                        <strong>{formatMetricNumber(newsletterSummary.sentCampaignsPeriod)}</strong>
                                    </div>
                                </div>
                                {#if pendingSubscribersCount > 0}
                                    <p class="txt-xs txt-hint m-b-0">
                                        {formatMetricNumber(pendingSubscribersCount)} subscriber confirmation{pendingSubscribersCount === 1 ? "" : "s"} still pending.
                                    </p>
                                {:else if newsletterHasSubscribers}
                                    <p class="txt-xs txt-hint m-b-0">No pending confirmations right now.</p>
                                {:else}
                                    <div class="report-empty-state">No newsletter subscribers recorded for this period.</div>
                                {/if}

                                {#if newsletterSummary.sentCampaignsPeriod < 1}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {/if}
                            </article>

                            <article class="report-breakdown-item">
                                <h6 class="m-0 report-breakdown-title">Recent submitted campaigns</h6>
                                {#if !recentSentCampaigns.length}
                                    <div class="report-empty-state">No campaigns submitted for this period.</div>
                                {:else}
                                    <div class="reports-list">
                                        {#each recentSentCampaigns as campaign}
                                            <div class="reports-list-item">
                                                <div class="reports-list-main">
                                                    <div class="reports-list-title">{truncate(campaign.subject, 80)}</div>
                                                    <div class="txt-sm txt-hint">
                                                        {formatMetricNumber(campaign.recipientsCount)} recipients submitted - {formatDateTime(campaign.sentTs || campaign.updatedTs || campaign.createdTs)}
                                                    </div>
                                                </div>
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
                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">SEO summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Search visibility basics across this website.</p>
                        </div>
                        <div class="report-metric-grid">
                            {#each seoSummaryMetricRows as metricRow (metricRow.label)}
                                <div class="report-metric-row">
                                    <span>{metricRow.label}</span>
                                    <strong>{formatMetricNumber(metricRow.count)}</strong>
                                </div>
                            {/each}
                        </div>
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Quick SEO fixes</h5>
                            <p class="txt-sm txt-hint m-b-0">Most common page issues affecting search visibility basics.</p>
                        </div>
                        {#if seoIssueBreakdownHasData}
                            {@const maxSeoIssueCount = resolveTrafficMaxCount(seoIssueBreakdownRows)}
                            <div class="report-bar-list">
                                {#each seoIssueBreakdownRows as issueRow}
                                    <div class="report-bar-item">
                                        <div class="report-bar-head">
                                            <span class="report-bar-label">{issueRow.label}</span>
                                            <strong class="report-bar-value">{formatMetricNumber(issueRow.count)}</strong>
                                        </div>
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
                    </section>

                    <section class="panel reports-breakdown-card">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Pages needing attention</h5>
                            <p class="txt-sm txt-hint m-b-0">Prioritized pages for quick SEO fixes.</p>
                        </div>

                        {#if !prioritizedSeoRows.length}
                            <div class="report-empty-state">No SEO issues detected for this website.</div>
                        {:else}
                            <div class="reports-list">
                                {#each prioritizedSeoRows as pageRow (pageRow.id)}
                                    <div class="reports-list-item">
                                        <div class="reports-list-main">
                                            <div class="reports-list-title">{pageRow.title}</div>
                                            <div class="txt-sm txt-hint">{pageRow.reasons.join(" - ")}</div>
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
                    <section class="panel reports-placeholder-panel">
                        <div class="placeholder-section m-b-0">
                            <span class="loader loader-lg" />
                            <h1>Loading traffic analytics...</h1>
                        </div>
                    </section>
                {:else if trafficState === "ok"}
                    <div class="reports-traffic-layout">
                        <section class="panel reports-breakdown-card">
                            <div class="section-head m-b-sm">
                                <h5 class="m-0">Traffic overview</h5>
                                <p class="txt-sm txt-hint m-b-0">{trafficPeriod?.label || selectedPeriodLabel}</p>
                            </div>
                            <div class="report-metric-grid">
                                <div class="report-metric-row">
                                    <span>Visitors</span>
                                    <strong>{formatMetricNumber(trafficSummary?.visitors)}</strong>
                                </div>
                                <div class="report-metric-row">
                                    <span>Pageviews</span>
                                    <strong>{formatMetricNumber(trafficSummary?.pageviews)}</strong>
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
                            {#if trafficOverviewMessage}
                                <p class="txt-sm txt-hint m-b-0 report-traffic-state-note">
                                    <span class={`label label-sm ${resolveMetricStatePillClass(trafficOverviewStatus)}`}>
                                        {trafficOverviewStatusLabel}
                                    </span>
                                    {trafficOverviewMessage}
                                </p>
                            {/if}
                        </section>

                        <section class="panel reports-breakdown-card">
                            <div class="section-head m-b-sm">
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

                        <section class="panel reports-breakdown-card">
                            <div class="section-head m-b-sm">
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

                        <section class="panel reports-breakdown-card">
                            <div class="section-head m-b-sm">
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

                        <section class="panel reports-breakdown-card">
                            <div class="section-head m-b-sm">
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
                    <section class="panel reports-placeholder-panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Traffic</h5>
                            <p class="txt-sm txt-hint m-b-0">Analytics status for this website.</p>
                        </div>
                        <div class="empty-state m-b-0">
                            {resolveTrafficStateMessage(trafficDisplayState, { isAdminViewer: canConfigureTrafficAnalytics })}
                        </div>
                    </section>
                    {#if showTrafficAnalyticsSetup}
                        <section class="panel reports-breakdown-card reports-analytics-setup-card">
                            <div class="section-head m-b-sm">
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
                <section class="panel reports-placeholder-panel">
                    <div class="section-head m-b-sm">
                        <h5 class="m-0">History</h5>
                        <p class="txt-sm txt-hint m-b-0">Monthly snapshots are planned for a later phase.</p>
                    </div>
                    <div class="empty-state m-b-0">{historyPlaceholderMessage}</div>
                </section>
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
        gap: 10px;
        grid-template-columns: minmax(0, 1fr) minmax(300px, 340px);
        align-items: start;
    }

    .reports-overview-main,
    .reports-overview-rail,
    .reports-traffic-layout,
    .report-health-panel {
        display: grid;
        gap: 10px;
    }

    .reports-kpi-grid {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-overview-kpi-grid {
        grid-template-columns: repeat(auto-fit, minmax(175px, 1fr));
        gap: 8px;
    }

    .reports-kpi-card {
        display: grid;
        gap: 6px;
    }

    .reports-overview-kpi-card {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 14%, var(--baseColor));
        box-shadow: none;
        padding: 10px;
        gap: 5px;
    }

    .reports-kpi-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .reports-kpi-label {
        font-weight: 600;
    }

    .reports-kpi-value {
        font-size: 24px;
        line-height: 1.1;
        font-weight: 700;
        color: var(--txtPrimaryColor);
    }

    .reports-kpi-hint {
        line-height: 1.35;
    }

    .reports-overview-section {
        display: grid;
        gap: 8px;
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

    .reports-rail-card {
        background: color-mix(in srgb, var(--baseAlt1Color) 10%, var(--baseColor));
        gap: 8px;
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
        padding: 8px 0;
        border-top: 1px dashed var(--baseAlt2Color);
        color: var(--txtPrimaryColor);
    }

    .reports-rail-item {
        gap: 6px;
        padding: 7px 0;
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
        padding: 2px 0 1px;
    }

    .report-health-panel .report-health-item:first-of-type {
        border-top: 0;
        padding-top: 0;
    }

    .report-health-item.warning {
        border-left: 2px solid color-mix(in srgb, var(--warningColor) 55%, var(--baseAlt2Color));
        padding-left: 8px;
    }

    .report-health-pill {
        min-width: 52px;
        justify-content: center;
    }

    .reports-confidence-list {
        gap: 0;
    }

    .reports-confidence-list .reports-rail-item {
        border-top: 1px dashed var(--baseAlt2Color);
    }

    .reports-confidence-list .reports-rail-item:first-of-type {
        border-top: 0;
        padding-top: 0;
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
        .reports-analytics-setup-grid {
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
    }

    @media (max-width: 760px) {
        .reports-head.operations-head .selector-row {
            min-width: 100%;
        }
    }
</style>
