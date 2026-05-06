<script>
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";

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
    let reportsLoadError = "";

    let lastWebsitesCollectionId = "";
    let lastDataKey = "";
    let dataLoadToken = 0;

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: contactsCollection = resolveCollectionByAliases(["contacts", "contact"]);
    $: whatsappCollection = resolveCollectionByAliases(["whatsapp", "whatsapp_interactions", "whatsapp_clicks"]);
    $: appointmentsCollection = resolveCollectionByAliases(["appointments"]);
    $: bookingServicesCollection = resolveCollectionByAliases(["bookingservices"]);
    $: subscribersCollection = resolveCollectionByAliases(["subscribers"]);
    $: campaignsCollection = resolveCollectionByAliases(["campaigns"]);
    $: pagesCollection = resolveCollectionByAliases(["pages"]);

    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
    $: reportsFeatureAvailable = resolveReportsFeatureAvailable(selectedWebsite);

    $: selectedPeriodLabel = periodOptions.find((option) => option.key === selectedPeriod)?.label || "This month";

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

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        clearDataRecords();
        reportsLoadError = "";
        lastWebsitesCollectionId = "";
        lastDataKey = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: if (selectedWebsiteId && dataKey !== lastDataKey) {
        lastDataKey = dataKey;
        loadDashboardData();
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

    $: recentSentCampaigns = [...normalizedCampaigns]
        .filter((campaign) => campaign.statusKey === "sent")
        .sort((a, b) => (b.sentTs || b.updatedTs || b.createdTs || 0) - (a.sentTs || a.updatedTs || a.createdTs || 0))
        .slice(0, 8);

    $: normalizedPages = normalizePages(pagesRecords, websiteLabelById);
    $: selectedWebsiteSeo = normalizeWebsiteSeo(selectedWebsite);
    $: seoSummary = buildSeoSummary(normalizedPages, selectedWebsiteSeo);

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
    });

    $: reportSuggestions = buildReportSuggestions({
        leadsSummary,
        bookingSummary,
        newsletterSummary,
        seoSummary,
    });

    $: reportHealthState = resolveReportHealthState(reportWarnings.length);

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
            hint: `${newsletterSummary.recipientsReachedPeriod} recipients reached`,
            icon: "ri-send-plane-2-line",
        },
        {
            key: "seoAttention",
            label: "SEO pages needing attention",
            value: seoSummary.needsAttention + seoSummary.missingBasics,
            hint: `${seoSummary.missingTitle} missing title - ${seoSummary.missingDescription} missing description`,
            icon: "ri-search-eye-line",
        },
    ];

    $: trafficPlaceholderMessage = "Traffic analytics are not configured yet.";
    $: historyPlaceholderMessage = "Monthly report history will appear here once automatic snapshots are enabled.";

    export function reload() {
        if (!websitesCollection?.id) {
            return Promise.resolve();
        }

        if (!selectedWebsiteId) {
            return loadWebsites();
        }

        return loadDashboardData();
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
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

    function parseSettings(rawSettings) {
        if (rawSettings && typeof rawSettings === "object" && !Array.isArray(rawSettings)) {
            return rawSettings;
        }

        if (typeof rawSettings === "string") {
            try {
                const parsed = JSON.parse(rawSettings);
                return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? parsed : {};
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
        return (
            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim()
            || website?.id
            || ""
        );
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
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            clearDataRecords();
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

        warnings.push("Traffic analytics are not configured yet.");

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

            <div class="head-selector">
                <div class="selector-row">
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

                <div class="selector-row">
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
        <section class="panel reports-body m-b-base">
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
                        <div class="reports-kpi-grid">
                            {#each overviewMetricCards as metric (metric.key)}
                                <article class="panel reports-kpi-card">
                                    <div class="reports-kpi-head">
                                        <span class="txt-sm txt-hint">{metric.label}</span>
                                        <i class={metric.icon} aria-hidden="true" />
                                    </div>
                                    <div class="reports-kpi-value">{metric.value}</div>
                                    <div class="txt-sm txt-hint">{metric.hint}</div>
                                </article>
                            {/each}
                        </div>

                        <div class="reports-grid-two m-t-sm">
                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Leads by type</h5>
                                    <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Contact form</span><strong>{leadsSummary.contactCount}</strong></div>
                                    <div class="reports-mini-row"><span>WhatsApp</span><strong>{leadsSummary.whatsappCount}</strong></div>
                                    <div class="reports-mini-row"><span>Booking</span><strong>{leadsSummary.bookingCount}</strong></div>
                                </div>
                            </article>

                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Booking status</h5>
                                    <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Pending</span><strong>{bookingSummary.pendingCount}</strong></div>
                                    <div class="reports-mini-row"><span>Confirmed</span><strong>{bookingSummary.confirmedCount}</strong></div>
                                    <div class="reports-mini-row"><span>Cancelled</span><strong>{bookingSummary.cancelledCount}</strong></div>
                                    <div class="reports-mini-row"><span>Upcoming</span><strong>{bookingSummary.upcomingCount}</strong></div>
                                </div>
                            </article>
                        </div>

                        <div class="reports-grid-two m-t-sm">
                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">Newsletter activity</h5>
                                    <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Active subscribers</span><strong>{newsletterSummary.activeSubscribers}</strong></div>
                                    <div class="reports-mini-row"><span>New subscribers</span><strong>{newsletterSummary.newSubscribersPeriod}</strong></div>
                                    <div class="reports-mini-row"><span>Sent campaigns</span><strong>{newsletterSummary.sentCampaignsPeriod}</strong></div>
                                    <div class="reports-mini-row"><span>Recipients reached</span><strong>{newsletterSummary.recipientsReachedPeriod}</strong></div>
                                </div>
                            </article>

                            <article class="panel">
                                <div class="section-head m-b-sm">
                                    <h5 class="m-0">SEO health</h5>
                                    <p class="txt-sm txt-hint m-b-0">Current pages status.</p>
                                </div>
                                <div class="reports-mini-stack">
                                    <div class="reports-mini-row"><span>Good</span><strong>{seoSummary.good}</strong></div>
                                    <div class="reports-mini-row"><span>Needs attention</span><strong>{seoSummary.needsAttention}</strong></div>
                                    <div class="reports-mini-row"><span>Missing basics</span><strong>{seoSummary.missingBasics}</strong></div>
                                    <div class="reports-mini-row"><span>Noindex pages</span><strong>{seoSummary.noindexPages}</strong></div>
                                </div>
                            </article>
                        </div>
                    </div>

                    <aside class="reports-overview-rail">
                        <section class="panel report-health-panel">
                            <div class="report-health-head">
                                <div class="report-health-main">
                                    <h5 class="m-0">Report health</h5>
                                    <p class="txt-sm txt-hint m-b-0">Key warnings and next actions for this period.</p>
                                </div>
                                <div class="report-health-meta">
                                    <span class={`label label-sm ${reportHealthState.pillClass}`}>{reportHealthState.label}</span>
                                    <span class="summary-pill">{reportWarnings.length} warnings - {reportSuggestions.length} suggestions</span>
                                </div>
                            </div>

                            <div class="report-health-group m-t-8">
                                <div class="report-health-group-title">Warnings</div>
                                {#if reportWarnings.length}
                                    {#each reportWarnings as warning}
                                        <div class="report-health-item warning">
                                            <span class="label label-sm report-health-pill warning">Warning</span>
                                            <span>{warning}</span>
                                        </div>
                                    {/each}
                                {:else}
                                    <p class="txt-sm txt-hint m-b-0">No warnings for this period.</p>
                                {/if}
                            </div>

                            <div class="report-health-group m-t-8">
                                <div class="report-health-group-title">Suggestions</div>
                                {#if reportSuggestions.length}
                                    {#each reportSuggestions as suggestion}
                                        <div class="report-health-item">
                                            <span class="label label-sm report-health-pill">Info</span>
                                            <span>{suggestion}</span>
                                        </div>
                                    {/each}
                                {:else}
                                    <p class="txt-sm txt-hint m-b-0">No suggestions yet.</p>
                                {/if}
                            </div>
                        </section>

                        <section class="panel report-sources-panel">
                            <div class="section-head m-b-sm">
                                <h5 class="m-0">Data coverage</h5>
                                <p class="txt-sm txt-hint m-b-0">Availability of the data sources used in this dashboard.</p>
                            </div>
                            <div class="reports-mini-stack">
                                {#each sourceReadinessRows as sourceRow}
                                    <div class="reports-mini-row">
                                        <span>{sourceRow.label}</span>
                                        <span class:txt-success={sourceRow.ok} class:txt-hint={!sourceRow.ok}>{sourceRow.message}</span>
                                    </div>
                                {/each}
                            </div>
                        </section>
                    </aside>
                </div>
            {:else if activeTab === "leads"}
                <div class="reports-tab-grid">
                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Leads summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Total leads</span><strong>{leadsSummary.total}</strong></div>
                            <div class="reports-mini-row"><span>New</span><strong>{leadsSummary.newCount}</strong></div>
                            <div class="reports-mini-row"><span>Read</span><strong>{leadsSummary.readCount}</strong></div>
                            <div class="reports-mini-row"><span>Archived</span><strong>{leadsSummary.archivedCount}</strong></div>
                        </div>
                    </article>

                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Lead types</h5>
                            <p class="txt-sm txt-hint m-b-0">Contact form, WhatsApp, and Booking channels.</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Contact form</span><strong>{leadsSummary.contactCount}</strong></div>
                            <div class="reports-mini-row"><span>WhatsApp</span><strong>{leadsSummary.whatsappCount}</strong></div>
                            <div class="reports-mini-row"><span>Booking</span><strong>{leadsSummary.bookingCount}</strong></div>
                        </div>
                    </article>
                </div>

                <section class="panel m-t-sm">
                    <div class="section-head m-b-sm">
                        <h5 class="m-0">Recent leads ({selectedPeriodLabel})</h5>
                        <p class="txt-sm txt-hint m-b-0">Latest entries from contact and WhatsApp channels.</p>
                    </div>

                    {#if !sortedRecentLeads.length}
                        <div class="empty-state m-b-0">No data available for this period.</div>
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
                </section>
            {:else if activeTab === "booking"}
                <div class="reports-tab-grid">
                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Appointments summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Total appointments</span><strong>{bookingSummary.total}</strong></div>
                            <div class="reports-mini-row"><span>Pending</span><strong>{bookingSummary.pendingCount}</strong></div>
                            <div class="reports-mini-row"><span>Confirmed</span><strong>{bookingSummary.confirmedCount}</strong></div>
                            <div class="reports-mini-row"><span>Cancelled</span><strong>{bookingSummary.cancelledCount}</strong></div>
                            <div class="reports-mini-row"><span>Upcoming</span><strong>{bookingSummary.upcomingCount}</strong></div>
                        </div>
                    </article>

                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Top services</h5>
                            <p class="txt-sm txt-hint m-b-0">Most requested services in the selected period.</p>
                        </div>

                        {#if !topBookingServices.length}
                            <div class="empty-state m-b-0">No booking service data for this period.</div>
                        {:else}
                            <div class="reports-mini-stack">
                                {#each topBookingServices as serviceRow}
                                    <div class="reports-mini-row"><span>{serviceRow.label}</span><strong>{serviceRow.count}</strong></div>
                                {/each}
                            </div>
                        {/if}
                    </article>
                </div>

                <section class="panel m-t-sm">
                    <div class="section-head m-b-sm">
                        <h5 class="m-0">Upcoming appointments</h5>
                        <p class="txt-sm txt-hint m-b-0">Next pending/confirmed requests.</p>
                    </div>

                    {#if !upcomingAppointments.length}
                        <div class="empty-state m-b-0">No upcoming appointments right now.</div>
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
                </section>
            {:else if activeTab === "newsletter"}
                <div class="reports-tab-grid">
                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Newsletter summary</h5>
                            <p class="txt-sm txt-hint m-b-0">{selectedPeriodLabel}</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Active subscribers</span><strong>{newsletterSummary.activeSubscribers}</strong></div>
                            <div class="reports-mini-row"><span>New subscribers</span><strong>{newsletterSummary.newSubscribersPeriod}</strong></div>
                            <div class="reports-mini-row"><span>Sent campaigns</span><strong>{newsletterSummary.sentCampaignsPeriod}</strong></div>
                            <div class="reports-mini-row"><span>Recipients reached</span><strong>{newsletterSummary.recipientsReachedPeriod}</strong></div>
                            <div class="reports-mini-row"><span>Draft campaigns</span><strong>{newsletterSummary.draftCampaigns}</strong></div>
                        </div>
                    </article>

                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Recent sent campaigns</h5>
                            <p class="txt-sm txt-hint m-b-0">Latest sent records for this website.</p>
                        </div>

                        {#if !recentSentCampaigns.length}
                            <div class="empty-state m-b-0">No sent campaigns available.</div>
                        {:else}
                            <div class="reports-mini-stack">
                                {#each recentSentCampaigns as campaign}
                                    <div class="reports-mini-row reports-mini-row--wrap">
                                        <span>{truncate(campaign.subject, 80)}</span>
                                        <span class="txt-sm txt-hint">
                                            {campaign.recipientsCount} recipients - {formatDateTime(campaign.sentTs || campaign.updatedTs || campaign.createdTs)}
                                        </span>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </article>
                </div>
            {:else if activeTab === "seo"}
                <div class="reports-tab-grid">
                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">Page SEO health</h5>
                            <p class="txt-sm txt-hint m-b-0">Current pages status for this website.</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Total pages</span><strong>{seoSummary.totalPages}</strong></div>
                            <div class="reports-mini-row"><span>Good</span><strong>{seoSummary.good}</strong></div>
                            <div class="reports-mini-row"><span>Needs attention</span><strong>{seoSummary.needsAttention}</strong></div>
                            <div class="reports-mini-row"><span>Missing basics</span><strong>{seoSummary.missingBasics}</strong></div>
                        </div>
                    </article>

                    <article class="panel">
                        <div class="section-head m-b-sm">
                            <h5 class="m-0">SEO gaps</h5>
                            <p class="txt-sm txt-hint m-b-0">Most common blockers found in pages.</p>
                        </div>
                        <div class="reports-mini-stack">
                            <div class="reports-mini-row"><span>Missing title</span><strong>{seoSummary.missingTitle}</strong></div>
                            <div class="reports-mini-row"><span>Missing description</span><strong>{seoSummary.missingDescription}</strong></div>
                            <div class="reports-mini-row"><span>Missing social image</span><strong>{seoSummary.missingSocialImage}</strong></div>
                            <div class="reports-mini-row"><span>Noindex pages</span><strong>{seoSummary.noindexPages}</strong></div>
                        </div>
                    </article>
                </div>

                <section class="panel m-t-sm">
                    <div class="section-head m-b-sm">
                        <h5 class="m-0">Pages needing attention</h5>
                        <p class="txt-sm txt-hint m-b-0">Focus on these pages first.</p>
                    </div>

                    {#if !seoSummary.pageRows.length}
                        <div class="empty-state m-b-0">No pages need SEO attention right now.</div>
                    {:else}
                        <div class="reports-list">
                            {#each seoSummary.pageRows as pageRow (pageRow.id)}
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
            {:else if activeTab === "traffic"}
                <section class="panel reports-placeholder-panel">
                    <div class="section-head m-b-sm">
                        <h5 class="m-0">Traffic</h5>
                        <p class="txt-sm txt-hint m-b-0">Plausible integration will be added in a later phase.</p>
                    </div>
                    <div class="empty-state m-b-0">{trafficPlaceholderMessage}</div>
                </section>
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
        gap: 16px;
    }

    .reports-head.operations-head .head-selector {
        display: flex;
        align-items: flex-end;
        gap: 10px;
        flex-wrap: wrap;
    }

    .reports-head.operations-head .selector-row {
        min-width: min(100%, 260px);
        display: grid;
        gap: 6px;
    }

    .reports-head.operations-head .summary-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .reports-body {
        display: grid;
        gap: 12px;
    }

    .reports-tabs {
        flex-wrap: wrap;
    }

    .reports-overview-layout {
        display: grid;
        gap: 12px;
        grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
        align-items: start;
    }

    .reports-overview-main {
        display: grid;
        gap: 10px;
    }

    .reports-overview-rail {
        display: grid;
        gap: 10px;
    }

    .reports-kpi-grid {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-kpi-card {
        display: grid;
        gap: 6px;
    }

    .reports-kpi-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .reports-kpi-value {
        font-size: 26px;
        line-height: 1.1;
        font-weight: 700;
        color: var(--txtPrimaryColor);
    }

    .reports-grid-two,
    .reports-tab-grid {
        display: grid;
        gap: 10px;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .reports-mini-stack {
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

    .reports-mini-row--wrap {
        align-items: flex-start;
        flex-direction: column;
    }

    .reports-list {
        display: grid;
        gap: 8px;
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

    .report-health-panel {
        display: grid;
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

    .report-health-group-title {
        font-size: 11px;
        letter-spacing: 0.06em;
        text-transform: uppercase;
        color: var(--txtHintColor);
        margin-bottom: 6px;
        font-weight: 700;
    }

    .report-health-item {
        display: grid;
        grid-template-columns: auto 1fr;
        gap: 8px;
        align-items: flex-start;
        padding: 8px 0;
        border-top: 1px dashed var(--baseAlt2Color);
        color: var(--txtPrimaryColor);
    }

    .report-health-item.warning {
        color: color-mix(in srgb, var(--warningColor) 80%, var(--txtPrimaryColor));
    }

    .report-health-pill {
        min-width: 56px;
        justify-content: center;
    }

    .report-health-pill.warning {
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--warningColor) 88%, var(--txtPrimaryColor));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .reports-placeholder-panel {
        min-height: 180px;
        display: grid;
        align-content: start;
        gap: 8px;
    }

    @media (max-width: 1220px) {
        .reports-overview-layout {
            grid-template-columns: 1fr;
        }

        .reports-overview-rail {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 980px) {
        .reports-grid-two,
        .reports-tab-grid,
        .reports-kpi-grid,
        .reports-overview-rail {
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
        .reports-head.operations-head .head-selector {
            width: 100%;
        }

        .reports-head.operations-head .selector-row {
            min-width: 100%;
        }
    }
</style>
