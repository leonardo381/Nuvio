<script>
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { onMount } from "svelte";
    import { pageTitle } from "@/stores/app";
    import {
        collections,
        collectionsLoadError,
        findCollectionByRequiredNames,
        hasCollectionsLoaded,
        isCollectionsLoading,
    } from "@/stores/collections";
    import { addErrorToast, addSuccessToast, addWarningToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { normalizeWebsiteSettingsValue } from "@/utils/WebsiteSettingsSchema";

    // NUVIO CUSTOM START: Unified Leads operations page for purpose-built backoffice workflows.
    $pageTitle = "Leads";

    const sourceFilterOptions = [
        { key: "all", label: "All lead types" },
        { key: "contact", label: "Contact form" },
        { key: "whatsapp", label: "WhatsApp" },
        { key: "booking", label: "Booking" },
    ];
    const statusFilterOptions = [
        { key: "all", label: "All statuses" },
        { key: "new", label: "New" },
        { key: "read", label: "Read" },
    ];
    const periodFilterOptions = [
        { key: "all", label: "All time" },
        { key: "today", label: "Today" },
        { key: "thisWeek", label: "This week" },
        { key: "thisMonth", label: "This month" },
    ];
    const leadsViewOptions = [
        { key: "inbox", label: "Inbox", icon: "ri-inbox-line" },
        { key: "archived", label: "Archived", icon: "ri-archive-line" },
    ];
    const archivedStatusAliases = ["archived", "archive"];
    const leadNotesFieldAliases = ["notes", "note", "internalNotes", "internal_notes"];
    const leadLastContactedFieldAliases = ["lastContactedAt", "last_contacted_at", "lastContacted", "last_contacted"];

    const ALL_WEBSITES_KEY = "all";
    const contactsCollectionAliases = ["contacts", "contact", "Contacts"];
    const whatsappCollectionAliases = [
        "whatsapp",
        "Whatsapp",
        "WhatsApp",
        "whatsapp_interactions",
        "whatsappInteractions",
        "whatsapp_clicks",
    ];
    const desktopMasterDetailMinWidth = 1060;
    const leadsInitialVisibleCount = 24;
    const leadsVisibleIncrement = 24;
    const staleInboxLeadDaysThreshold = 3;

    let websites = [];
    let selectedWebsiteId = "";
    let contactsRecords = [];
    let whatsappRecords = [];

    let sourceFilter = "all";
    let statusFilter = "all";
    let periodFilter = "all";
    let leadsView = "inbox";
    let searchTerm = "";
    let sortOrder = "newest";
    let selectedLeadKey = "";
    let selectedLeadKeys = [];
    let selectedLeadKeysSet = new Set();
    let selectedLeadsCount = 0;
    let visibleLeadLimit = leadsInitialVisibleCount;
    let lastLeadVisibleResetKey = "";
    let isLeadDetailsActive = false;
    let isDesktopMasterDetail = false;
    let isUpdatingLeadStatus = false;
    let isBulkUpdatingLeads = false;
    let bulkLeadActionKey = "";
    let updatingLeadStatusKey = "";
    let isSavingLeadFollowUp = false;
    let leadFollowUpError = "";
    let leadNotesDraft = "";
    let leadNotesDraftKey = "";
    let isLeadFollowUpOpen = false;
    let isLeadUtilitiesOpen = false;
    let isInvitingLeadToNewsletter = false;

    let isLoadingWebsites = false;
    let isLoadingLeads = false;
    let leadsError = "";

    let lastWebsitesCollectionId = "";
    let lastLeadsDataKey = "";

    $: websitesCollection = resolveCollectionByAliases(["websites", "Websites"]);
    $: contactsCollection = resolveCollectionByAliases(contactsCollectionAliases);
    $: resolvedWhatsAppCollection = resolveWhatsAppCollection();
    $: whatsappCollection = resolvedWhatsAppCollection?.collection || null;
    $: resolvedWhatsAppAlias = resolvedWhatsAppCollection?.alias || "";

    $: hasAnyLeadCollections = !!contactsCollection?.id || !!whatsappCollection?.id;
    $: hasRequiredCollections = !!contactsCollection?.id && !!whatsappCollection?.id;

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        lastWebsitesCollectionId = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: websiteOptionMap = new Map(
        websites.map((website) => [website.id, resolveWebsiteLabel(website)]),
    );
    $: if (!websites.length) {
        selectedWebsiteId = "";
    } else if (!websiteOptionMap.has(selectedWebsiteId)) {
        selectedWebsiteId = websites[0].id;
    }

    $: leadsDataKey = [
        selectedWebsiteId,
        contactsCollection?.id || "",
        whatsappCollection?.id || "",
        resolvedWhatsAppAlias,
    ].join(":");

    $: if (hasAnyLeadCollections && leadsDataKey !== lastLeadsDataKey) {
        lastLeadsDataKey = leadsDataKey;
        loadLeads();
    }

    $: normalizedSearchTerm = `${searchTerm || ""}`.trim().toLowerCase();
    $: normalizedLeadsView = leadsView === "archived" ? "archived" : "inbox";
    $: contactsStatusSupport = resolveStatusSupport(contactsCollection);
    $: whatsappStatusSupport = resolveStatusSupport(whatsappCollection);
    $: contactsNotesFieldName = resolveCollectionFieldNameByAliases(contactsCollection, leadNotesFieldAliases);
    $: whatsappNotesFieldName = resolveCollectionFieldNameByAliases(whatsappCollection, leadNotesFieldAliases);
    $: contactsLastContactedFieldName = resolveCollectionFieldNameByAliases(contactsCollection, leadLastContactedFieldAliases);
    $: whatsappLastContactedFieldName = resolveCollectionFieldNameByAliases(whatsappCollection, leadLastContactedFieldAliases);

    $: normalizedLeads = normalizeLeadRecords({
        contacts: contactsRecords,
        whatsapp: whatsappRecords,
        websiteLabelById: websiteOptionMap,
    });

    $: websiteScopedLeads = filterLeadsBySelectedWebsite(normalizedLeads, selectedWebsiteId);
    $: inboxLeadsCount = websiteScopedLeads.filter((lead) => !isArchivedLead(lead)).length;
    $: archivedLeadsCount = websiteScopedLeads.filter((lead) => isArchivedLead(lead)).length;
    $: leadsByView = filterLeadsByView(websiteScopedLeads, normalizedLeadsView);
    $: leadsByType = filterLeadsByType(leadsByView, sourceFilter);
    $: leadsByStatus = filterLeadsByStatus(leadsByType, normalizedLeadsView, statusFilter);
    $: leadsByPeriod = filterLeadsByPeriod(leadsByStatus, periodFilter);
    $: leadsBySearch = filterLeadsBySearch(leadsByPeriod, normalizedSearchTerm);
    $: filteredLeads = sortLeads(leadsBySearch, sortOrder);
    $: leadsVisibleResetKey = [
        selectedWebsiteId,
        normalizedLeadsView,
        normalizedSearchTerm,
        sourceFilter,
        statusFilter,
        periodFilter,
        sortOrder,
    ].join(":");
    $: if (leadsVisibleResetKey !== lastLeadVisibleResetKey) {
        lastLeadVisibleResetKey = leadsVisibleResetKey;
        visibleLeadLimit = leadsInitialVisibleCount;
    }
    $: visibleLeads = filteredLeads.slice(0, visibleLeadLimit);
    $: visibleLeadsCount = visibleLeads.length;
    $: canLoadMoreLeads = visibleLeadsCount < filteredLeads.length;
    $: selectedLeadKeysSet = new Set(
        selectedLeadKeys
            .map((key) => normalizeString(key))
            .filter(Boolean),
    );
    $: selectedLeadsCount = selectedLeadKeysSet.size;
    $: {
        const visibleKeys = new Set(filteredLeads.map((lead) => normalizeString(lead?.key)).filter(Boolean));
        const nextSelected = selectedLeadKeys.filter((key) => visibleKeys.has(normalizeString(key)));
        if (nextSelected.length !== selectedLeadKeys.length) {
            selectedLeadKeys = nextSelected;
        }
    }
    $: if (filteredLeads.length) {
        const isSelectedLeadVisible = selectedLeadKey
            ? filteredLeads.some((lead) => lead.key === selectedLeadKey)
            : false;
        if (!isSelectedLeadVisible) {
            selectedLeadKey = filteredLeads[0].key;
        }
    } else if (selectedLeadKey) {
        selectedLeadKey = "";
        isLeadDetailsActive = false;
    }
    $: selectedLead = websiteScopedLeads.find((lead) => lead.key === selectedLeadKey) || null;
    $: selectedLeads = websiteScopedLeads.filter((lead) => selectedLeadKeysSet.has(normalizeString(lead?.key)));
    $: bulkMarkReadEligibleCount = selectedLeads.filter((lead) => canBulkMarkLeadAsRead(lead)).length;
    $: bulkArchiveEligibleCount = selectedLeads.filter((lead) => canBulkArchiveLead(lead)).length;
    $: bulkMoveInboxEligibleCount = selectedLeads.filter((lead) => canBulkMoveLeadToInbox(lead)).length;
    $: if (isDesktopMasterDetail && isLeadDetailsActive) {
        isLeadDetailsActive = false;
    }
    $: selectedLeadMailto = buildMailtoLink(selectedLead?.email);
    $: selectedLeadWhatsAppPhone = normalizeString(selectedLead?.whatsappTargetPhone || selectedLead?.phone);
    $: selectedLeadWhatsAppMessage = normalizeString(selectedLead?.whatsappTargetMessage || selectedLead?.message);
    $: selectedLeadWhatsAppLink = buildWhatsAppLink(selectedLeadWhatsAppPhone, selectedLeadWhatsAppMessage);
    $: selectedLeadNewsletterInviteEmail = normalizeString(selectedLead?.email);
    $: selectedLeadHasValidInviteEmail = isValidEmailAddress(selectedLeadNewsletterInviteEmail);
    $: selectedLeadNewsletterInviteWebsiteId = normalizeString(selectedLead?.websiteId);
    $: selectedLeadCanInviteToNewsletter = !!selectedLead?.recordId
        && selectedLeadHasValidInviteEmail
        && !!selectedLeadNewsletterInviteWebsiteId
        && !isInvitingLeadToNewsletter;
    $: selectedLeadInviteUnavailableMessage = !selectedLead
        ? ""
        : !selectedLeadHasValidInviteEmail
            ? "This lead does not have a valid email address."
            : !selectedLeadNewsletterInviteWebsiteId
                ? "This lead is missing website context for newsletter invitation."
                : "";
    $: websiteRecordById = new Map(
        websites.map((website) => [normalizeString(website?.id), website]),
    );
    $: selectedLeadWebsiteRecord = selectedLead?.websiteId
        ? websiteRecordById.get(selectedLead.websiteId) || null
        : null;
    $: selectedLeadNotificationSetup = resolveLeadNotificationSetup(
        selectedLead,
        selectedLeadWebsiteRecord,
    );
    $: selectedLeadStatusSupport = resolveLeadStatusSupport(selectedLead);
    $: canToggleSelectedLeadStatus = !!selectedLead?.recordId && selectedLeadStatusSupport?.supportsToggle;
    $: selectedLeadCurrentStatusNormalized = normalizeLower(selectedLead?.statusValue);
    $: isSelectedLeadArchived = normalizeLower(selectedLead?.statusKey) === "archived";
    $: selectedLeadToggleToRead = selectedLeadCurrentStatusNormalized === normalizeLower(selectedLeadStatusSupport?.newValue);
    $: selectedLeadToggleActionLabel = selectedLeadToggleToRead ? "Mark as read" : "Mark as new";
    $: selectedLeadToggleActionTarget = selectedLeadToggleToRead
        ? selectedLeadStatusSupport?.readValue
        : selectedLeadStatusSupport?.newValue;
    $: canMoveSelectedLeadToInbox = !!selectedLead?.recordId
        && selectedLeadStatusSupport?.supportsToggle
        && isSelectedLeadArchived;
    $: canArchiveSelectedLead = !!selectedLead?.recordId
        && selectedLeadStatusSupport?.supportsArchive
        && !isSelectedLeadArchived
        && selectedLeadCurrentStatusNormalized !== normalizeLower(selectedLeadStatusSupport?.archiveValue);
    $: selectedLeadFollowUpSupport = resolveLeadFollowUpSupport(selectedLead);
    $: selectedLeadLastContactedLabel = resolveLastContactedLabel(selectedLead?.lastContactedAt);
    $: selectedLeadLastContactedDisplay = selectedLeadLastContactedLabel || "Not contacted yet";
    $: selectedLeadNotesValue = `${selectedLead?.notes || ""}`;
    $: leadNotesDirty = !!selectedLead && `${leadNotesDraft || ""}` !== selectedLeadNotesValue;
    $: canSaveSelectedLeadNote = !!selectedLead?.recordId
        && !!selectedLeadFollowUpSupport?.collectionId
        && !!selectedLeadFollowUpSupport?.notesFieldName;
    $: canMarkSelectedLeadContacted = !!selectedLead?.recordId
        && !!selectedLeadFollowUpSupport?.collectionId
        && !!selectedLeadFollowUpSupport?.lastContactedFieldName;
    $: if (selectedLead?.key) {
        if (leadNotesDraftKey !== selectedLead.key) {
            leadNotesDraftKey = selectedLead.key;
            leadNotesDraft = `${selectedLead?.notes || ""}`;
            leadFollowUpError = "";
        }
    } else if (leadNotesDraftKey || leadNotesDraft || leadFollowUpError) {
        leadNotesDraftKey = "";
        leadNotesDraft = "";
        leadFollowUpError = "";
    }

    $: nonArchivedLeads = websiteScopedLeads.filter((lead) => !isArchivedLead(lead));
    $: inboxUncontactedLeadsCount = nonArchivedLeads.filter((lead) => (lead?.lastContactedTs || 0) <= 0).length;
    $: inboxOlderUncontactedLeadsCount = nonArchivedLeads.filter((lead) => {
        if ((lead?.lastContactedTs || 0) > 0) {
            return false;
        }

        return isTimestampOlderThanDays(lead?.createdTs, staleInboxLeadDaysThreshold);
    }).length;
    $: inboxMissingContactLeadsCount = nonArchivedLeads.filter((lead) => !normalizeString(lead?.email) && !normalizeString(lead?.phone)).length;
    $: inboxLeadsWithNotesCount = nonArchivedLeads.filter((lead) => !!normalizeString(lead?.notes)).length;
    $: inboxNewsletterInviteEligibleCount = nonArchivedLeads.filter((lead) => isValidEmailAddress(lead?.email)).length;
    $: leadHealthWarnings = buildLeadHealthWarnings({
        totalWebsiteLeads: websiteScopedLeads.length,
        inboxLeadsCount: nonArchivedLeads.length,
        newLeadsCount: nonArchivedLeads.filter((lead) => lead.statusKey === "new").length,
        uncontactedLeadsCount: inboxUncontactedLeadsCount,
        olderUncontactedLeadsCount: inboxOlderUncontactedLeadsCount,
        missingContactLeadsCount: inboxMissingContactLeadsCount,
    });
    $: leadHealthSuggestions = buildLeadHealthSuggestions({
        totalWebsiteLeads: websiteScopedLeads.length,
        inboxLeadsCount: nonArchivedLeads.length,
        archivedLeadsCount,
        newLeadsCount: nonArchivedLeads.filter((lead) => lead.statusKey === "new").length,
        uncontactedLeadsCount: inboxUncontactedLeadsCount,
        olderUncontactedLeadsCount: inboxOlderUncontactedLeadsCount,
        missingContactLeadsCount: inboxMissingContactLeadsCount,
        leadsWithNotesCount: inboxLeadsWithNotesCount,
        newsletterInviteEligibleCount: inboxNewsletterInviteEligibleCount,
    });
    $: leadHealthState = resolveLeadHealthState(leadHealthWarnings, leadHealthSuggestions);
    $: totalNewLeads = nonArchivedLeads.filter((lead) => lead.statusKey === "new").length;
    $: totalThisMonthLeads = nonArchivedLeads.filter((lead) => isCurrentMonth(lead.created)).length;
    $: totalContactFormLeads = nonArchivedLeads.filter((lead) => lead.sourceKey === "contact").length;
    $: totalWhatsAppLeads = nonArchivedLeads.filter((lead) => lead.sourceKey === "whatsapp").length;
    // NUVIO CUSTOM END: Unified Leads operations page for purpose-built backoffice workflows.

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
    }

    function normalizeWebsiteRelationValue(value) {
        if (Array.isArray(value)) {
            return normalizeString(value[0]);
        }
        return normalizeString(value);
    }

    function resolveCollectionByAliases(aliases = []) {
        return findCollectionByRequiredNames($collections, aliases);
    }

    function resolveWhatsAppCollection() {
        for (const alias of whatsappCollectionAliases) {
            const collection = resolveCollectionByAliases([alias]);
            if (collection) {
                return {
                    alias,
                    collection,
                };
            }
        }

        return null;
    }

    function resolveWebsitesSort(collection) {
        const preferredSortFields = ["title", "name", "slug"];
        const availableFields = new Set(
            CommonHelper.getAllCollectionIdentifiers(collection).map((field) => normalizeLower(field)),
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

    function resolveValueByCandidates(record, candidates = []) {
        for (const key of candidates) {
            const value = normalizeString(record?.[key]);
            if (value) {
                return value;
            }
        }

        return "";
    }

    function resolveCollectionField(collection, names = []) {
        const normalizedNames = names.map((name) => normalizeLower(name)).filter(Boolean);
        const fields = Array.isArray(collection?.fields) ? collection.fields : [];

        for (const field of fields) {
            const fieldName = normalizeLower(field?.name);
            if (fieldName && normalizedNames.includes(fieldName)) {
                return field;
            }
        }

        return null;
    }

    function resolveCollectionFieldNameByAliases(collection, aliases = []) {
        const field = resolveCollectionField(collection, aliases);
        return normalizeString(field?.name);
    }

    function resolveFieldSelectValues(field) {
        if (Array.isArray(field?.values)) {
            return field.values.map((value) => normalizeString(value)).filter(Boolean);
        }

        if (Array.isArray(field?.options?.values)) {
            return field.options.values.map((value) => normalizeString(value)).filter(Boolean);
        }

        return [];
    }

    function findMatchingValue(values, aliases = []) {
        const normalizedAliases = aliases.map((alias) => normalizeLower(alias)).filter(Boolean);

        for (const value of values) {
            if (normalizedAliases.includes(normalizeLower(value))) {
                return value;
            }
        }

        return "";
    }

    function resolveStatusSupport(collection) {
        const field = resolveCollectionField(collection, ["status"]);
        const statusValues = resolveFieldSelectValues(field);
        const newValue = findMatchingValue(statusValues, ["new"]);
        const readValue = findMatchingValue(statusValues, ["read"]);
        const archiveValue = findMatchingValue(statusValues, archivedStatusAliases);

        return {
            collectionId: normalizeString(collection?.id),
            collectionName: normalizeString(collection?.name),
            fieldName: normalizeString(field?.name),
            supportsToggle: !!(field?.name && newValue && readValue),
            newValue: newValue || "new",
            readValue: readValue || "read",
            supportsArchive: !!(field?.name && archiveValue),
            archiveValue,
        };
    }

    function resolveStatusKey(statusValue, statusSupport) {
        const normalizedStatus = normalizeLower(statusValue);
        if (!normalizedStatus) {
            return "";
        }

        if (normalizedStatus === normalizeLower(statusSupport?.newValue)) {
            return "new";
        }

        if (normalizedStatus === normalizeLower(statusSupport?.readValue)) {
            return "read";
        }

        if (statusSupport?.supportsArchive && normalizedStatus === normalizeLower(statusSupport?.archiveValue)) {
            return "archived";
        }

        return normalizedStatus;
    }

    function toTimestamp(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return 0;
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const parsed = new Date(normalized).getTime();
        return Number.isNaN(parsed) ? 0 : parsed;
    }

    function getPeriodStartTimestamp(periodKey) {
        const now = new Date();
        const dayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);

        if (periodKey === "today") {
            return dayStart.getTime();
        }

        if (periodKey === "thisWeek") {
            const weekStart = new Date(dayStart);
            const dayIndex = weekStart.getDay();
            const mondayOffset = (dayIndex + 6) % 7;
            weekStart.setDate(weekStart.getDate() - mondayOffset);
            return weekStart.getTime();
        }

        if (periodKey === "thisMonth") {
            return new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0).getTime();
        }

        return 0;
    }

    function isCurrentMonth(value) {
        const timestamp = toTimestamp(value);
        if (!timestamp) {
            return false;
        }

        const date = new Date(timestamp);
        const now = new Date();

        return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
    }

    function resolveLastContactedLabel(value) {
        const timestamp = toTimestamp(value);
        if (!timestamp) {
            return "";
        }

        const nowTs = Date.now();
        const diffMs = Math.max(0, nowTs - timestamp);
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays === 0) {
            return "today";
        }
        if (diffDays === 1) {
            return "yesterday";
        }
        if (diffDays < 7) {
            return `${diffDays} days ago`;
        }

        return formatDateTime(value);
    }

    function formatDateTime(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return "-";
        }

        const parsed = new Date(raw.includes("T") ? raw : raw.replace(" ", "T"));
        if (Number.isNaN(parsed.getTime())) {
            return raw;
        }

        return parsed.toLocaleString();
    }

    function truncate(value, max = 180) {
        const text = normalizeString(value);
        if (!text || text.length <= max) {
            return text;
        }

        return `${text.slice(0, max - 3)}...`;
    }

    function resolveLeadSourceFromContactChannel(channel) {
        const normalizedChannel = normalizeLower(channel);
        if (normalizedChannel === "booking") {
            return "booking";
        }

        return "contact";
    }

    function resolveSourceLabel(sourceKey) {
        if (sourceKey === "whatsapp") {
            return "WhatsApp";
        }
        if (sourceKey === "booking") {
            return "Booking";
        }
        return "Contact form";
    }

    function resolveStatusLabel(statusKey, statusValue = "") {
        if (statusKey === "new") {
            return "New";
        }
        if (statusKey === "read") {
            return "Read";
        }
        if (statusKey === "archived") {
            return "Archived";
        }
        if (statusValue) {
            return CommonHelper.sentenize(statusValue, false);
        }
        return "";
    }

    function resolveSourceBadgeClass(sourceKey) {
        if (sourceKey === "whatsapp") {
            return "label-success";
        }
        if (sourceKey === "booking") {
            return "label-warning";
        }
        return "label-info";
    }

    function resolveStatusBadgeClass(statusKey) {
        if (statusKey === "new") {
            return "label-warning";
        }
        if (statusKey === "read") {
            return "label-success";
        }
        if (statusKey === "archived") {
            return "label-danger";
        }
        return "";
    }

    function resolveWebsiteDisplayName(record, websiteLabelById) {
        const websiteId = normalizeWebsiteRelationValue(record?.website);
        if (!websiteId) {
            return "";
        }

        const expandedWebsite = Array.isArray(record?.expand?.website)
            ? record.expand.website[0]
            : record?.expand?.website;
        const expandedLabel = resolveWebsiteLabel(expandedWebsite);

        if (expandedLabel) {
            return expandedLabel;
        }

        return websiteLabelById.get(websiteId) || "";
    }

    function resolveLeadIdentity(lead) {
        if (lead.name) {
            return lead.name;
        }
        if (lead.email) {
            return lead.email;
        }
        if (lead.phone) {
            return lead.phone;
        }
        if (lead.sourceKey === "whatsapp") {
            return "WhatsApp interaction";
        }

        return "Unknown visitor";
    }

    function resolveSecondaryIdentity(lead) {
        const parts = [];

        if (lead.email && lead.email !== lead.identity) {
            parts.push(lead.email);
        }
        if (lead.phone && lead.phone !== lead.identity) {
            parts.push(lead.phone);
        }

        return parts.join(" - ");
    }

    function resolveLeadPreview(subject, message) {
        const normalizedSubject = normalizeString(subject);
        const normalizedMessage = normalizeString(message);

        if (normalizedSubject && normalizedMessage) {
            return truncate(`${normalizedSubject} - ${normalizedMessage}`);
        }

        return truncate(normalizedSubject || normalizedMessage);
    }

    function resolveLeadAttributionDetail(lead) {
        const sourceLabel = normalizeLower(resolveSourceLabel(lead?.sourceKey));
        const candidates = [lead?.originSource, lead?.page];

        for (const candidate of candidates) {
            const value = normalizeString(candidate);
            if (!value) {
                continue;
            }

            if (normalizeLower(value) === sourceLabel) {
                continue;
            }

            return value;
        }

        return "";
    }

    function resolveLeadAttribution(lead) {
        const sourceLabel = resolveSourceLabel(lead?.sourceKey);
        const detail = resolveLeadAttributionDetail(lead);

        if (!detail) {
            return sourceLabel;
        }

        return `${sourceLabel} · ${detail}`;
    }

    function resolveLeadLocationHint(lead) {
        const parts = [];
        const usedDetail = normalizeLower(resolveLeadAttributionDetail(lead));

        const websiteName = normalizeString(lead?.websiteName);
        if (websiteName) {
            parts.push(websiteName);
        }

        for (const candidate of [lead?.page, lead?.originSource]) {
            const value = normalizeString(candidate);
            if (!value) {
                continue;
            }

            if (normalizeLower(value) === usedDetail) {
                continue;
            }

            if (parts.some((part) => normalizeLower(part) === normalizeLower(value))) {
                continue;
            }

            parts.push(value);
        }

        return parts.join(" · ");
    }

    function resolveLeadPreviewText(lead) {
        const preview = normalizeString(lead?.preview);
        if (preview) {
            return preview;
        }

        if (normalizeLower(lead?.sourceKey) === "whatsapp") {
            return "No message captured for this interaction.";
        }

        return "No message preview available yet.";
    }

    function isTimestampOlderThanDays(timestamp, days) {
        const normalizedDays = Number(days) || 0;
        if (normalizedDays <= 0) {
            return false;
        }

        const safeTimestamp = Number(timestamp) || 0;
        if (!safeTimestamp) {
            return false;
        }

        return safeTimestamp <= (Date.now() - (normalizedDays * 24 * 60 * 60 * 1000));
    }

    function formatCountNoun(count, singular, plural) {
        const safeCount = Number(count) || 0;
        return `${safeCount} ${safeCount === 1 ? singular : plural}`;
    }

    function buildLeadHealthWarnings({
        totalWebsiteLeads,
        inboxLeadsCount,
        newLeadsCount,
        uncontactedLeadsCount,
        olderUncontactedLeadsCount,
        missingContactLeadsCount,
    }) {
        if (!totalWebsiteLeads && !inboxLeadsCount) {
            return [];
        }

        const warnings = [];
        if (newLeadsCount > 0) {
            warnings.push(`${formatCountNoun(newLeadsCount, "new lead", "new leads")} still need review.`);
        }
        if (uncontactedLeadsCount > 0) {
            warnings.push(`${formatCountNoun(uncontactedLeadsCount, "inbox lead", "inbox leads")} have not been contacted yet.`);
        }
        if (olderUncontactedLeadsCount > 0) {
            warnings.push(
                `${formatCountNoun(olderUncontactedLeadsCount, "inbox lead", "inbox leads")} are older than ${staleInboxLeadDaysThreshold} days and still waiting for follow-up.`,
            );
        }
        if (missingContactLeadsCount > 0) {
            warnings.push(`${formatCountNoun(missingContactLeadsCount, "inbox lead", "inbox leads")} are missing usable contact details.`);
        }

        return warnings;
    }

    function buildLeadHealthSuggestions({
        totalWebsiteLeads,
        inboxLeadsCount,
        archivedLeadsCount,
        newLeadsCount,
        uncontactedLeadsCount,
        olderUncontactedLeadsCount,
        missingContactLeadsCount,
        leadsWithNotesCount,
        newsletterInviteEligibleCount,
    }) {
        if (!totalWebsiteLeads && !inboxLeadsCount) {
            return [];
        }

        const suggestions = [];
        if (newLeadsCount > 0) {
            suggestions.push("Review new leads and mark them as read after handling.");
        }
        if (uncontactedLeadsCount > 0) {
            suggestions.push("Follow up with inbox leads that have not been contacted yet.");
        }
        if (olderUncontactedLeadsCount > 0) {
            suggestions.push(`Prioritize inbox leads older than ${staleInboxLeadDaysThreshold} days without follow-up.`);
        }
        if (missingContactLeadsCount > 0) {
            suggestions.push("Use internal notes to capture alternate contact details or next steps.");
        }
        if (inboxLeadsCount > 0 && leadsWithNotesCount < inboxLeadsCount) {
            suggestions.push("Add internal notes to important leads so follow-ups stay consistent.");
        }
        if (inboxLeadsCount > 0 && archivedLeadsCount === 0) {
            suggestions.push("Archive completed or resolved leads to keep Inbox focused.");
        }
        if (newsletterInviteEligibleCount > 0) {
            suggestions.push("Use newsletter invites for qualified contacts when appropriate.");
        }

        if (!suggestions.length) {
            suggestions.push("Lead inbox is healthy for this website.");
        }

        return suggestions;
    }

    function resolveLeadHealthState(warnings = [], suggestions = []) {
        if (warnings.length) {
            return {
                label: "Needs attention",
                badgeClass: "label-warning",
            };
        }

        if (suggestions.length) {
            return {
                label: "Ready for follow-up",
                badgeClass: "label-success",
            };
        }

        return {
            label: "No data",
            badgeClass: "label-info",
        };
    }

    function resolveLeadNotificationSourceKey(lead) {
        const sourceKey = normalizeLower(lead?.sourceKey);
        if (sourceKey === "contact" || sourceKey === "booking") {
            return "contactForm";
        }
        if (sourceKey === "whatsapp") {
            return "whatsapp";
        }

        return "";
    }

    function normalizeRecipientEntries(rawList) {
        if (!Array.isArray(rawList)) {
            return [];
        }

        const normalized = [];
        const seen = new Set();

        const appendEntry = (candidate) => {
            const entry = normalizeLower(candidate);
            if (!entry || seen.has(entry)) {
                return;
            }

            seen.add(entry);
            normalized.push(entry);
        };

        for (const item of rawList) {
            if (typeof item === "string") {
                for (const piece of item.split(/[\n,;]+/g)) {
                    appendEntry(piece);
                }
                continue;
            }

            if (item && typeof item === "object") {
                appendEntry(item?.email);
                appendEntry(item?.address);
                appendEntry(item?.value);
            }
        }

        return normalized;
    }

    function resolveLeadNotificationSetup(lead, websiteRecord) {
        const sourceSettingsKey = resolveLeadNotificationSourceKey(lead);

        if (!sourceSettingsKey) {
            return {
                available: false,
                unavailableMessage: "Notification setup unavailable for this lead.",
            };
        }

        if (!websiteRecord) {
            return {
                available: false,
                unavailableMessage: "Notification setup unavailable for this lead.",
            };
        }

        const settings = normalizeWebsiteSettingsValue(websiteRecord?.settings || {});
        const settingsSection = settings?.[sourceSettingsKey];
        const emailNotifications = settingsSection?.emailNotifications || {};

        const toRecipients = normalizeRecipientEntries(emailNotifications?.to);
        const ccRecipients = normalizeRecipientEntries(emailNotifications?.cc);
        const toCount = toRecipients.length;
        const ccCount = ccRecipients.length;
        const recipientsTotal = toCount + ccCount;
        const notificationsEnabled = emailNotifications?.enabled === true;
        const hasRecipients = recipientsTotal > 0;
        const isContactSource = sourceSettingsKey === "contactForm";
        const featureName = isContactSource ? "Contact Form" : "WhatsApp";
        const featureAvailableValue = settings?.featureFlags?.[sourceSettingsKey];
        const featureAvailable = typeof featureAvailableValue === "boolean"
            ? featureAvailableValue
            : null;

        return {
            available: true,
            notificationsEnabled,
            hasRecipients,
            toCount,
            ccCount,
            recipientsTotal,
            featureAvailable,
            featureAvailabilityLabel: featureAvailable === null
                ? ""
                : featureAvailable
                    ? `${featureName} available`
                    : `${featureName} unavailable`,
        };
    }

    function normalizeLeadRecords({ contacts, whatsapp, websiteLabelById }) {
        const normalized = [];

        for (const record of contacts || []) {
            const sourceKey = resolveLeadSourceFromContactChannel(record?.channel);
            const statusValue = normalizeString(record?.[contactsStatusSupport.fieldName || "status"]);
            const statusKey = resolveStatusKey(statusValue, contactsStatusSupport);
            const subject = normalizeString(record?.subject);
            const message = normalizeString(record?.message);
            const notes = resolveValueByCandidates(record, [contactsNotesFieldName, ...leadNotesFieldAliases]);
            const lastContactedAt = resolveValueByCandidates(record, [contactsLastContactedFieldName, ...leadLastContactedFieldAliases]);
            const lead = {
                key: `contacts:${record?.id || CommonHelper.randomString(8)}`,
                sourceKey,
                sourceLabel: resolveSourceLabel(sourceKey),
                sourceBadgeClass: resolveSourceBadgeClass(sourceKey),
                statusKey,
                statusValue,
                statusLabel: resolveStatusLabel(statusKey, statusValue),
                statusBadgeClass: resolveStatusBadgeClass(statusKey),
                supportsStatusActions: contactsStatusSupport.supportsToggle,
                supportsArchiveActions: contactsStatusSupport.supportsArchive,
                websiteName: resolveWebsiteDisplayName(record, websiteLabelById),
                websiteId: normalizeWebsiteRelationValue(record?.website),
                page: resolveValueByCandidates(record, ["page", "pagePath", "page_path", "pageSlug", "page_slug"]),
                originSource: resolveValueByCandidates(record, ["source", "sourceLabel", "source_label", "origin"]),
                name: normalizeString(record?.name),
                email: normalizeString(record?.email),
                phone: normalizeString(record?.phone),
                subject,
                message,
                whatsappTargetPhone: "",
                whatsappTargetMessage: "",
                notes,
                lastContactedAt,
                lastContactedTs: toTimestamp(lastContactedAt),
                preview: resolveLeadPreview(subject, message),
                created: normalizeString(record?.created),
                createdTs: toTimestamp(record?.created),
                collectionName: normalizeString(contactsCollection?.name),
                collectionId: normalizeString(contactsCollection?.id),
                recordId: normalizeString(record?.id),
            };

            lead.identity = resolveLeadIdentity(lead);
            lead.secondaryIdentity = resolveSecondaryIdentity(lead);
            normalized.push(lead);
        }

        for (const record of whatsapp || []) {
            const sourceKey = "whatsapp";
            const statusValue = normalizeString(record?.[whatsappStatusSupport.fieldName || "status"]);
            const statusKey = resolveStatusKey(statusValue, whatsappStatusSupport);
            const subject = resolveValueByCandidates(record, ["subject", "title"]);
            const message = resolveValueByCandidates(record, ["message", "body", "text", "note"]);
            const notes = resolveValueByCandidates(record, [whatsappNotesFieldName, ...leadNotesFieldAliases]);
            const lastContactedAt = resolveValueByCandidates(record, [whatsappLastContactedFieldName, ...leadLastContactedFieldAliases]);
            const whatsappTargetPhone = resolveValueByCandidates(record, [
                "targetPhone",
                "target_phone",
                "whatsappPhone",
                "whatsapp_phone",
                "phoneNumber",
                "phone_number",
                "to",
            ]);
            const whatsappTargetMessage = resolveValueByCandidates(record, [
                "targetMessage",
                "target_message",
                "prefilledMessage",
                "prefilled_message",
                "messageTemplate",
                "message_template",
                "text",
                "message",
                "body",
            ]);
            const lead = {
                key: `whatsapp:${record?.id || CommonHelper.randomString(8)}`,
                sourceKey,
                sourceLabel: resolveSourceLabel(sourceKey),
                sourceBadgeClass: resolveSourceBadgeClass(sourceKey),
                statusKey,
                statusValue,
                statusLabel: resolveStatusLabel(statusKey, statusValue),
                statusBadgeClass: resolveStatusBadgeClass(statusKey),
                supportsStatusActions: whatsappStatusSupport.supportsToggle,
                supportsArchiveActions: whatsappStatusSupport.supportsArchive,
                websiteName: resolveWebsiteDisplayName(record, websiteLabelById),
                websiteId: normalizeWebsiteRelationValue(record?.website),
                page: resolveValueByCandidates(record, ["page", "pagePath", "page_path", "pageSlug", "page_slug"]),
                originSource: resolveValueByCandidates(record, ["source", "sourceLabel", "source_label", "origin"]),
                name: normalizeString(record?.name),
                email: normalizeString(record?.email),
                phone: normalizeString(record?.phone),
                subject,
                message,
                whatsappTargetPhone,
                whatsappTargetMessage,
                notes,
                lastContactedAt,
                lastContactedTs: toTimestamp(lastContactedAt),
                preview: resolveLeadPreview(subject, message || whatsappTargetMessage),
                created: normalizeString(record?.created),
                createdTs: toTimestamp(record?.created),
                collectionName: normalizeString(whatsappCollection?.name),
                collectionId: normalizeString(whatsappCollection?.id),
                recordId: normalizeString(record?.id),
            };

            lead.identity = resolveLeadIdentity(lead);
            lead.secondaryIdentity = resolveSecondaryIdentity(lead);
            normalized.push(lead);
        }

        return normalized;
    }

    function isArchivedLead(lead) {
        return archivedStatusAliases.includes(normalizeLower(lead?.statusKey));
    }

    function filterLeadsBySelectedWebsite(leads, websiteIdFilter) {
        const next = Array.isArray(leads) ? leads : [];
        if (websiteIdFilter === ALL_WEBSITES_KEY) {
            return [...next];
        }

        const websiteId = normalizeString(websiteIdFilter);
        return next.filter((lead) => normalizeString(lead?.websiteId) === websiteId);
    }

    function filterLeadsByView(leads, viewKey) {
        const next = Array.isArray(leads) ? leads : [];
        return viewKey === "archived"
            ? next.filter((lead) => isArchivedLead(lead))
            : next.filter((lead) => !isArchivedLead(lead));
    }

    function filterLeadsByType(leads, sourceFilterValue) {
        const next = Array.isArray(leads) ? leads : [];
        const normalizedSourceFilter = normalizeLower(sourceFilterValue);
        if (!normalizedSourceFilter || normalizedSourceFilter === "all") {
            return [...next];
        }

        return next.filter((lead) => normalizeLower(lead?.sourceKey) === normalizedSourceFilter);
    }

    function filterLeadsByStatus(leads, viewKey, statusFilterValue) {
        const next = Array.isArray(leads) ? leads : [];
        if (viewKey === "archived") {
            return [...next];
        }

        const normalizedStatusFilter = normalizeLower(statusFilterValue);
        if (!normalizedStatusFilter || normalizedStatusFilter === "all") {
            return [...next];
        }

        return next.filter((lead) => normalizeLower(lead?.statusKey) === normalizedStatusFilter);
    }

    function filterLeadsByPeriod(leads, periodFilterValue) {
        const next = Array.isArray(leads) ? leads : [];
        const periodStartTimestamp = getPeriodStartTimestamp(periodFilterValue);
        if (!periodStartTimestamp) {
            return [...next];
        }

        return next.filter((lead) => (lead?.createdTs || 0) >= periodStartTimestamp);
    }

    function resolveLeadSearchText(lead) {
        return [
            lead?.identity,
            lead?.name,
            lead?.email,
            lead?.phone,
            lead?.secondaryIdentity,
            lead?.websiteName,
            lead?.subject,
            lead?.message,
            lead?.whatsappTargetMessage,
            lead?.notes,
            lead?.preview,
            lead?.page,
            lead?.originSource,
            lead?.sourceLabel,
            lead?.statusLabel,
            resolveLeadAttribution(lead),
            resolveLeadLocationHint(lead),
        ]
            .map((value) => normalizeLower(value))
            .filter(Boolean)
            .join(" ");
    }

    function filterLeadsBySearch(leads, normalizedSearchValue) {
        const next = Array.isArray(leads) ? leads : [];
        if (!normalizedSearchValue) {
            return [...next];
        }

        return next.filter((lead) => resolveLeadSearchText(lead).includes(normalizedSearchValue));
    }

    function sortLeads(leads, sortDirection) {
        const next = [...(Array.isArray(leads) ? leads : [])];
        next.sort((a, b) => {
            if (sortDirection === "oldest") {
                return (a.createdTs || 0) - (b.createdTs || 0);
            }

            return (b.createdTs || 0) - (a.createdTs || 0);
        });

        return next;
    }

    function resolveLeadFollowUpSupport(lead) {
        if (!lead?.sourceKey) {
            return {
                collectionId: "",
                notesFieldName: "",
                lastContactedFieldName: "",
            };
        }

        if (lead.sourceKey === "whatsapp") {
            return {
                collectionId: normalizeString(whatsappCollection?.id),
                notesFieldName: whatsappNotesFieldName,
                lastContactedFieldName: whatsappLastContactedFieldName,
            };
        }

        return {
            collectionId: normalizeString(contactsCollection?.id),
            notesFieldName: contactsNotesFieldName,
            lastContactedFieldName: contactsLastContactedFieldName,
        };
    }

    async function updateLeadStatusBySource(sourceKey, recordId, nextStatusValue) {
        if (normalizeLower(sourceKey) === "whatsapp") {
            return ApiClient.updateLeadWhatsappStatus(recordId, nextStatusValue, {
                requestKey: `nuvio_leads_whatsapp_status_${recordId || "unknown"}`,
            });
        }

        return ApiClient.updateLeadContactStatus(recordId, nextStatusValue, {
            requestKey: `nuvio_leads_contact_status_${recordId || "unknown"}`,
        });
    }

    async function updateLeadFollowUpBySource(sourceKey, recordId, payload = {}) {
        if (normalizeLower(sourceKey) === "whatsapp") {
            return ApiClient.updateLeadWhatsappFollowUp(recordId, payload, {
                requestKey: `nuvio_leads_whatsapp_followup_${recordId || "unknown"}`,
            });
        }

        return ApiClient.updateLeadContactFollowUp(recordId, payload, {
            requestKey: `nuvio_leads_contact_followup_${recordId || "unknown"}`,
        });
    }

    async function loadWebsites() {
        if (!websitesCollection?.id) {
            websites = [];
            return;
        }

        isLoadingWebsites = true;

        try {
            websites = await ApiClient.getBackofficeWebsites({
                requestKey: "nuvio_leads_websites",
            });
        } catch (err) {
            websites = [];
            ApiClient.error(err, false);
        }

        isLoadingWebsites = false;
    }

    async function loadLeads() {
        leadsError = "";

        if (!selectedWebsiteId) {
            contactsRecords = [];
            whatsappRecords = [];
            return;
        }

        isLoadingLeads = true;

        try {
            const response = await ApiClient.getLeadsDashboard({
                websiteId: selectedWebsiteId,
                requestKey: `nuvio_leads_dashboard_${selectedWebsiteId || "unknown"}`,
            });

            contactsRecords = Array.isArray(response?.datasets?.contacts) ? response.datasets.contacts : [];
            whatsappRecords = Array.isArray(response?.datasets?.whatsapp) ? response.datasets.whatsapp : [];
        } catch (err) {
            leadsError = "Unable to load leads right now. Please refresh and try again.";
            ApiClient.error(err, false);
            contactsRecords = [];
            whatsappRecords = [];
        }

        isLoadingLeads = false;
    }

    function buildMailtoLink(email) {
        const value = normalizeString(email);
        if (!value) {
            return "";
        }

        return `mailto:${encodeURIComponent(value)}`;
    }

    function isValidEmailAddress(email) {
        const value = normalizeString(email);
        if (!value) {
            return false;
        }

        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
    }

    function buildWhatsAppLink(phone, message) {
        const normalizedPhone = normalizeString(phone).replace(/[^\d]/g, "");
        if (!normalizedPhone) {
            return "";
        }

        const normalizedMessage = normalizeString(message);
        if (!normalizedMessage) {
            return `https://wa.me/${normalizedPhone}`;
        }

        return `https://wa.me/${normalizedPhone}?text=${encodeURIComponent(normalizedMessage)}`;
    }

    function dispatchSidebarBadgeRefresh() {
        if (typeof window === "undefined") {
            return;
        }

        window.dispatchEvent(new CustomEvent("nuvio:sidebar-badges-refresh"));
    }

    async function copyValue(value, label) {
        const normalizedValue = normalizeString(value);
        if (!normalizedValue) {
            return;
        }

        try {
            if (navigator?.clipboard?.writeText) {
                await navigator.clipboard.writeText(normalizedValue);
            } else {
                const tempInput = document.createElement("textarea");
                tempInput.value = normalizedValue;
                tempInput.style.position = "fixed";
                tempInput.style.opacity = "0";
                document.body.appendChild(tempInput);
                try {
                    tempInput.select();
                    const copied = document.execCommand("copy");
                    if (!copied) {
                        throw new Error("Clipboard copy command failed.");
                    }
                } finally {
                    tempInput.remove();
                }
            }

            addSuccessToast(`${label} copied.`);
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast(`Unable to copy ${label.toLowerCase()}.`);
        }
    }

    function resolveLeadStatusSupport(lead) {
        if (!lead?.sourceKey) {
            return {
                collectionId: "",
                fieldName: "",
                supportsToggle: false,
                newValue: "new",
                readValue: "read",
                supportsArchive: false,
                archiveValue: "",
            };
        }

        if (lead.sourceKey === "whatsapp") {
            return whatsappStatusSupport;
        }

        return contactsStatusSupport;
    }

    function patchLeadRecord(sourceKey, recordId, patchData = {}) {
        if (!recordId || typeof patchData !== "object" || !Object.keys(patchData).length) {
            return;
        }

        const patch = (records) =>
            (records || []).map((record) =>
                record?.id === recordId
                    ? {
                        ...record,
                        ...patchData,
                    }
                    : record,
            );

        if (sourceKey === "whatsapp") {
            whatsappRecords = patch(whatsappRecords);
            return;
        }

        contactsRecords = patch(contactsRecords);
    }

    async function setSelectedLeadStatus(nextStatusValue) {
        if (!selectedLead?.recordId || !nextStatusValue || isUpdatingLeadStatus) {
            return;
        }

        const statusSupport = resolveLeadStatusSupport(selectedLead);
        if (!statusSupport?.supportsToggle && !statusSupport?.supportsArchive) {
            return;
        }

        const normalizedTargetStatus = normalizeLower(nextStatusValue);
        const isToggleStatus = normalizedTargetStatus === normalizeLower(statusSupport.newValue)
            || normalizedTargetStatus === normalizeLower(statusSupport.readValue);
        const isArchiveStatus = normalizedTargetStatus === normalizeLower(statusSupport.archiveValue);
        if ((isToggleStatus && !statusSupport.supportsToggle) || (isArchiveStatus && !statusSupport.supportsArchive)) {
            return;
        }

        isUpdatingLeadStatus = true;
        updatingLeadStatusKey = selectedLead.key;

        const sourceKey = selectedLead.sourceKey;
        const recordId = selectedLead.recordId;

        try {
            await updateLeadStatusBySource(sourceKey, recordId, nextStatusValue);
            patchLeadRecord(sourceKey, recordId, { status: nextStatusValue });
            dispatchSidebarBadgeRefresh();

            const nextLabel = normalizedTargetStatus === normalizeLower(statusSupport.readValue)
                ? "read"
                : normalizedTargetStatus === normalizeLower(statusSupport.newValue)
                    ? "new"
                    : normalizedTargetStatus === normalizeLower(statusSupport.archiveValue)
                        ? "archived"
                    : "updated";
            addSuccessToast(`Lead marked as ${nextLabel}.`);
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to update lead status right now.");
        } finally {
            isUpdatingLeadStatus = false;
            updatingLeadStatusKey = "";
        }
    }

    async function archiveSelectedLead() {
        if (!canArchiveSelectedLead || isUpdatingLeadStatus) {
            return;
        }

        await setSelectedLeadStatus(selectedLeadStatusSupport.archiveValue);
    }

    async function moveSelectedLeadToInbox() {
        if (!canMoveSelectedLeadToInbox || isUpdatingLeadStatus) {
            return;
        }

        await setSelectedLeadStatus(selectedLeadStatusSupport.readValue);
    }

    async function saveSelectedLeadNote() {
        if (!selectedLead?.recordId || !canSaveSelectedLeadNote || isSavingLeadFollowUp) {
            return;
        }

        leadFollowUpError = "";
        isSavingLeadFollowUp = true;
        const patchData = {
            notes: leadNotesDraft,
        };

        try {
            await updateLeadFollowUpBySource(selectedLead.sourceKey, selectedLead.recordId, patchData);
            patchLeadRecord(selectedLead.sourceKey, selectedLead.recordId, { notes: leadNotesDraft });
            addSuccessToast("Lead note saved.");
        } catch (err) {
            ApiClient.error(err, false);
            leadFollowUpError = "Unable to save follow-up note right now.";
            addErrorToast("Unable to save follow-up note right now.");
        } finally {
            isSavingLeadFollowUp = false;
        }
    }

    async function markSelectedLeadContactedNow() {
        if (!selectedLead?.recordId || !canMarkSelectedLeadContacted || isSavingLeadFollowUp) {
            return;
        }

        leadFollowUpError = "";
        isSavingLeadFollowUp = true;
        const patchData = {
            lastContactedAt: new Date().toISOString(),
        };

        if (leadNotesDirty && canSaveSelectedLeadNote) {
            patchData.notes = leadNotesDraft;
        }

        try {
            await updateLeadFollowUpBySource(selectedLead.sourceKey, selectedLead.recordId, patchData);
            patchLeadRecord(selectedLead.sourceKey, selectedLead.recordId, patchData);
            addSuccessToast("Lead marked as contacted.");
        } catch (err) {
            ApiClient.error(err, false);
            leadFollowUpError = "Unable to update follow-up details right now.";
            addErrorToast("Unable to update follow-up details right now.");
        } finally {
            isSavingLeadFollowUp = false;
        }
    }

    async function inviteSelectedLeadToNewsletter() {
        if (isInvitingLeadToNewsletter) {
            return;
        }

        if (!selectedLeadHasValidInviteEmail) {
            addErrorToast("This lead does not have a valid email address.");
            return;
        }

        if (!selectedLeadNewsletterInviteWebsiteId) {
            addErrorToast("This lead is missing website context for newsletter invitation.");
            return;
        }

        isInvitingLeadToNewsletter = true;

        try {
            const response = await ApiClient.send("/api/nuvio/newsletter/backoffice/invite", {
                method: "POST",
                body: {
                    websiteId: selectedLeadNewsletterInviteWebsiteId,
                    email: selectedLeadNewsletterInviteEmail,
                    name: normalizeString(selectedLead?.name),
                    source: normalizeString(selectedLead?.sourceKey),
                },
                requestKey: `nuvio_leads_newsletter_invite_${selectedLead?.recordId || selectedLead?.key || "selected"}`,
            });

            const inviteResult = normalizeLower(response?.result);
            if (inviteResult === "already_active") {
                addWarningToast("This contact is already subscribed.");
                return;
            }

            if (inviteResult === "unsubscribed") {
                addWarningToast("This contact has unsubscribed and was not invited.");
                return;
            }

            if (inviteResult === "resent") {
                addSuccessToast("Confirmation email sent again.");
                return;
            }

            addSuccessToast("Newsletter invitation sent.");
        } catch (err) {
            const statusCode = Number(err?.status) || 0;
            const backendMessage = normalizeLower(err?.data?.message || err?.message);
            if (statusCode === 403) {
                addErrorToast("You do not have permission to invite contacts to the newsletter.");
                return;
            }
            if (backendMessage.includes("valid email")) {
                addErrorToast("This lead does not have a valid email address.");
                return;
            }

            addErrorToast("Unable to send newsletter invitation right now.");
        } finally {
            isInvitingLeadToNewsletter = false;
        }
    }

    function openLeadDetails(lead) {
        if (!lead?.key) {
            return;
        }

        selectedLeadKey = lead.key;
        if (!isDesktopMasterDetail) {
            isLeadDetailsActive = true;
        }
    }

    function closeLeadDetails() {
        isLeadDetailsActive = false;
    }

    function isLeadSelectedForBulk(leadKey) {
        return selectedLeadKeysSet.has(normalizeString(leadKey));
    }

    function toggleLeadBulkSelection(leadKey) {
        const normalizedKey = normalizeString(leadKey);
        if (!normalizedKey) {
            return;
        }

        if (selectedLeadKeysSet.has(normalizedKey)) {
            selectedLeadKeys = selectedLeadKeys.filter((key) => normalizeString(key) !== normalizedKey);
            return;
        }

        selectedLeadKeys = [...selectedLeadKeys, normalizedKey];
    }

    function clearLeadBulkSelection() {
        if (!selectedLeadKeys.length) {
            return;
        }
        selectedLeadKeys = [];
    }

    function resolveBulkLeadStatusTarget(lead, actionKey) {
        if (!lead?.recordId) {
            return null;
        }

        const statusSupport = resolveLeadStatusSupport(lead);

        const normalizedStatusKey = normalizeLower(lead?.statusKey);
        const normalizedCurrentStatus = normalizeLower(lead?.statusValue);

        if (actionKey === "markRead") {
            const nextStatus = normalizeString(statusSupport.readValue);
            if (!statusSupport.supportsToggle || !nextStatus || normalizedStatusKey === "archived") {
                return null;
            }
            if (normalizedCurrentStatus === normalizeLower(nextStatus)) {
                return null;
            }

            return {
                statusSupport,
                nextStatus,
            };
        }

        if (actionKey === "archive") {
            const nextStatus = normalizeString(statusSupport.archiveValue);
            if (!statusSupport.supportsArchive || !nextStatus || normalizedStatusKey === "archived") {
                return null;
            }
            if (normalizedCurrentStatus === normalizeLower(nextStatus)) {
                return null;
            }

            return {
                statusSupport,
                nextStatus,
            };
        }

        if (actionKey === "moveInbox") {
            const nextStatus = normalizeString(statusSupport.readValue);
            if (!statusSupport.supportsToggle || !nextStatus || normalizedStatusKey !== "archived") {
                return null;
            }

            return {
                statusSupport,
                nextStatus,
            };
        }

        return null;
    }

    function canBulkMarkLeadAsRead(lead) {
        return !!resolveBulkLeadStatusTarget(lead, "markRead");
    }

    function canBulkArchiveLead(lead) {
        return !!resolveBulkLeadStatusTarget(lead, "archive");
    }

    function canBulkMoveLeadToInbox(lead) {
        return !!resolveBulkLeadStatusTarget(lead, "moveInbox");
    }

    function resolveBulkLeadSuccessMessage(actionKey, count) {
        const quantityLabel = `${count} lead${count === 1 ? "" : "s"}`;
        if (actionKey === "markRead") {
            return `${quantityLabel} marked as read.`;
        }
        if (actionKey === "archive") {
            return `${quantityLabel} archived.`;
        }
        return `${quantityLabel} moved to inbox.`;
    }

    async function applyBulkLeadStatusUpdate(actionKey) {
        if (!selectedLeadsCount || isBulkUpdatingLeads || isUpdatingLeadStatus) {
            return;
        }

        const selectedSnapshot = [...selectedLeads];
        const operations = selectedSnapshot
            .map((lead) => {
                const target = resolveBulkLeadStatusTarget(lead, actionKey);
                if (!target) {
                    return null;
                }

                return {
                    lead,
                    ...target,
                };
            })
            .filter(Boolean);

        if (!operations.length) {
            addErrorToast("Selected leads do not support this action.");
            return;
        }

        isBulkUpdatingLeads = true;
        bulkLeadActionKey = actionKey;

        try {
            const updateResults = await Promise.allSettled(
                operations.map((operation) =>
                    updateLeadStatusBySource(
                        operation.lead.sourceKey,
                        operation.lead.recordId,
                        operation.nextStatus,
                    )),
            );

            const successfulKeys = [];
            let successCount = 0;
            let failureCount = 0;

            updateResults.forEach((result, index) => {
                const operation = operations[index];
                if (!operation) {
                    return;
                }

                if (result.status === "fulfilled") {
                    patchLeadRecord(operation.lead.sourceKey, operation.lead.recordId, {
                        status: operation.nextStatus,
                    });
                    successCount += 1;
                    successfulKeys.push(normalizeString(operation.lead.key));
                    return;
                }

                failureCount += 1;
                ApiClient.error(result.reason, false);
            });

            if (successfulKeys.length) {
                const successfulKeySet = new Set(successfulKeys);
                selectedLeadKeys = selectedLeadKeys.filter((key) => !successfulKeySet.has(normalizeString(key)));
            }

            if (successCount) {
                addSuccessToast(resolveBulkLeadSuccessMessage(actionKey, successCount));
                dispatchSidebarBadgeRefresh();
            }

            if (failureCount) {
                addErrorToast("Some selected leads could not be updated.");
            }
        } finally {
            isBulkUpdatingLeads = false;
            bulkLeadActionKey = "";
        }
    }

    async function markSelectedLeadsAsRead() {
        await applyBulkLeadStatusUpdate("markRead");
    }

    async function archiveSelectedLeads() {
        await applyBulkLeadStatusUpdate("archive");
    }

    async function moveSelectedLeadsToInbox() {
        await applyBulkLeadStatusUpdate("moveInbox");
    }

    function handleLeadCardKeyDown(event, lead) {
        if (event.key !== "Enter" && event.key !== " ") {
            return;
        }

        event.preventDefault();
        openLeadDetails(lead);
    }

    function clearFilters() {
        sourceFilter = "all";
        statusFilter = "all";
        periodFilter = "all";
        searchTerm = "";
        sortOrder = "newest";
    }

    function setLeadsView(nextView) {
        leadsView = nextView === "archived" ? "archived" : "inbox";
        selectedLeadKeys = [];
        if (leadsView === "archived") {
            statusFilter = "all";
        }
    }

    function loadMoreLeads() {
        visibleLeadLimit += leadsVisibleIncrement;
    }

    function escapeCsvValue(value) {
        const raw = `${value ?? ""}`;
        if (!raw) {
            return "";
        }
        const escaped = raw.replace(/"/g, "\"\"");
        if (/[",\n\r]/.test(escaped)) {
            return `"${escaped}"`;
        }
        return escaped;
    }

    function normalizeFilenamePart(value) {
        return normalizeLower(value)
            .replace(/[^a-z0-9]+/g, "-")
            .replace(/^-+|-+$/g, "")
            || "all-websites";
    }

    function exportFilteredLeadsCsv() {
        const rows = [
            [
                "Created",
                "Last contacted",
                "Status",
                "Type",
                "Name",
                "Email",
                "Phone",
                "Subject",
                "Message",
                "Website",
                "Page/source",
                "Notes",
            ],
        ];

        for (const lead of filteredLeads) {
            rows.push([
                formatDateTime(lead.created),
                lead.lastContactedAt ? formatDateTime(lead.lastContactedAt) : "",
                lead.statusLabel || "",
                lead.sourceLabel || "",
                lead.name || "",
                lead.email || "",
                lead.phone || "",
                lead.subject || "",
                lead.message || lead.whatsappTargetMessage || "",
                lead.websiteName || "",
                resolveLeadLocationHint(lead) || resolveLeadAttribution(lead),
                lead.notes || "",
            ]);
        }

        const csvContent = rows.map((row) => row.map((cell) => escapeCsvValue(cell)).join(",")).join("\r\n");
        const selectedWebsite = websites.find((website) => website.id === selectedWebsiteId);
        const websiteLabel = selectedWebsiteId === ALL_WEBSITES_KEY
            ? "all-websites"
            : resolveWebsiteLabel(selectedWebsite || {});
        const datePart = new Date().toISOString().slice(0, 10);
        const fileName = `leads-${normalizeFilenamePart(websiteLabel)}-${datePart}.csv`;

        const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = fileName;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        URL.revokeObjectURL(url);
    }

    onMount(() => {
        if (typeof window === "undefined" || !window.matchMedia) {
            return undefined;
        }

        const query = window.matchMedia(`(min-width: ${desktopMasterDetailMinWidth}px)`);
        const updateLayoutMode = () => {
            isDesktopMasterDetail = !!query.matches;
        };

        updateLayoutMode();

        if (query.addEventListener) {
            query.addEventListener("change", updateLayoutMode);
        } else if (query.addListener) {
            query.addListener(updateLayoutMode);
        }

        return () => {
            if (query.removeEventListener) {
                query.removeEventListener("change", updateLayoutMode);
            } else if (query.removeListener) {
                query.removeListener(updateLayoutMode);
            }
        };
    });

    export function reload() {
        return loadLeads();
    }
</script>

<PageWrapper>
    <section class="operations-head panel leads-head m-b-base">
        <div class="head-main">
            <div class="summary-title-wrap">
                <div class="title-row">
                    <h2 class="m-0">Leads</h2>
                    <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={reload} />
                </div>
                <p class="head-description txt-sm txt-hint m-b-0">
                    Review contacts and interactions generated by the website.
                </p>
            </div>

            <div class="head-selector operations-website-select">
                <div class="selector-row">
                    <label class="txt-sm txt-hint selector-label m-b-0" for="leads-website-filter">Website</label>
                    <select
                        id="leads-website-filter"
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
            </div>
        </div>

        <div class="head-tools">
            <div class="tabs-header compact combined left operations-tabs leads-view-toggle">
                {#each leadsViewOptions as option (option.key)}
                    <button
                        type="button"
                        class="tab-item"
                        class:active={normalizedLeadsView === option.key}
                        on:click={() => setLeadsView(option.key)}
                    >
                        <i class={`${option.icon} tab-icon`} aria-hidden="true" />
                        <span class="tab-label">
                            {option.label} ({option.key === "archived" ? archivedLeadsCount : inboxLeadsCount})
                        </span>
                    </button>
                {/each}
            </div>
            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-mail-line" />
                    New: {totalNewLeads}
                </span>
                <span class="summary-pill">
                    <i class="ri-calendar-event-line" />
                    Total this month: {totalThisMonthLeads}
                </span>
                <span class="summary-pill">
                    <i class="ri-chat-1-line" />
                    Contact form: {totalContactFormLeads}
                </span>
                <span class="summary-pill">
                    <i class="ri-whatsapp-line" />
                    WhatsApp: {totalWhatsAppLeads}
                </span>
            </div>
        </div>
    </section>

    <section class="panel operations-content-panel leads-body m-b-base">
        <div class="leads-inbox-layout">
            <div class="leads-left-column">
                <div class="leads-controls">
                    <div class="leads-toolbar-row">
                        <div class="leads-filter-cell leads-filter-cell--search">
                            <label class="txt-sm txt-hint" for="leads-search-filter">Search</label>
                            <input
                                id="leads-search-filter"
                                class="input input-sm"
                                type="text"
                                placeholder="Search by name, email, phone, subject, message, notes, or source..."
                                bind:value={searchTerm}
                            />
                        </div>

                        <div class="leads-filter-cell leads-filter-cell--select">
                            <label class="txt-sm txt-hint" for="leads-type-filter">Type</label>
                            <select id="leads-type-filter" class="input input-sm" bind:value={sourceFilter}>
                                {#each sourceFilterOptions as option (option.key)}
                                    <option value={option.key}>{option.label}</option>
                                {/each}
                            </select>
                        </div>

                        {#if normalizedLeadsView === "inbox"}
                            <div class="leads-filter-cell leads-filter-cell--select">
                                <label class="txt-sm txt-hint" for="leads-status-filter">Status</label>
                                <select id="leads-status-filter" class="input input-sm" bind:value={statusFilter}>
                                    {#each statusFilterOptions as option (option.key)}
                                        <option value={option.key}>{option.label}</option>
                                    {/each}
                                </select>
                            </div>
                        {/if}

                        <div class="leads-filter-cell leads-filter-cell--select">
                            <label class="txt-sm txt-hint" for="leads-period-filter">Period</label>
                            <select id="leads-period-filter" class="input input-sm" bind:value={periodFilter}>
                                {#each periodFilterOptions as option (option.key)}
                                    <option value={option.key}>{option.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="leads-filter-cell leads-filter-cell--sort">
                            <label class="txt-sm txt-hint" for="leads-sort-filter">Sort</label>
                            <select id="leads-sort-filter" class="input input-sm" bind:value={sortOrder}>
                                <option value="newest">Newest first</option>
                                <option value="oldest">Oldest first</option>
                            </select>
                        </div>

                        <div class="leads-filter-cell leads-filter-cell--actions">
                            <button type="button" class="btn btn-sm btn-outline" on:click={clearFilters}>
                                <span class="txt">Reset filters</span>
                            </button>
                            <button type="button" class="btn btn-sm btn-outline" on:click={exportFilteredLeadsCsv}>
                                <span class="txt">Export CSV</span>
                            </button>
                        </div>
                    </div>
                </div>

                {#if $isCollectionsLoading || (!$hasCollectionsLoaded && !$collectionsLoadError)}
                    <div class="placeholder-section m-b-0">
                        <span class="loader loader-lg" />
                        <h1>Loading leads...</h1>
                    </div>
                {:else if $collectionsLoadError}
                    <div class="alert alert-danger m-b-0">
                        <div class="icon">
                            <i class="ri-error-warning-line" />
                        </div>
                        <div>
                            Could not verify Leads collections.<br />
                            Refresh the page or check your connection.
                        </div>
                    </div>
                {:else if !hasAnyLeadCollections}
                    <div class="alert alert-warning m-b-0">
                        <div class="icon">
                            <i class="ri-information-line" />
                        </div>
                        <div>
                            Leads data sources were not found. This page expects contact and WhatsApp interactions.
                        </div>
                    </div>
                {:else if !hasRequiredCollections}
                    <div class="alert alert-warning m-b-0">
                        <div class="icon">
                            <i class="ri-information-line" />
                        </div>
                        <div>
                            One leads source is missing. Contact and WhatsApp interactions may be partially available.
                        </div>
                    </div>
                {/if}

                {#if leadsError}
                    <div class="alert alert-danger m-b-0">
                        <div class="icon">
                            <i class="ri-error-warning-line" />
                        </div>
                        <div>{leadsError}</div>
                    </div>
                {/if}

                {#if isLoadingLeads}
                    <div class="placeholder-section m-b-0">
                        <span class="loader loader-lg" />
                        <h1>Loading leads...</h1>
                    </div>
                {:else if !websiteScopedLeads.length}
                    <div class="empty-state m-b-0">
                        {#if normalizedLeadsView === "archived"}
                            No archived leads yet.
                        {:else}
                            No leads yet. Contact form submissions and WhatsApp interactions will appear here.
                        {/if}
                    </div>
                {:else if !filteredLeads.length}
                    <div class="empty-state m-b-0">
                        {#if normalizedLeadsView === "archived" && !archivedLeadsCount}
                            No archived leads yet.
                        {:else}
                            No leads match these filters.
                        {/if}
                    </div>
                {:else}
                    <div class="leads-inbox-list" role="list">
                        {#each visibleLeads as lead (lead.key)}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                            <article
                                class="leads-inbox-item"
                                class:selected={selectedLeadKey === lead.key}
                                class:bulk-selected={selectedLeadKeysSet.has(normalizeString(lead.key))}
                                role="button"
                                tabindex="0"
                                aria-label={`Open ${lead.sourceLabel} lead details`}
                                on:click={() => openLeadDetails(lead)}
                                on:keydown={(event) => handleLeadCardKeyDown(event, lead)}
                            >
                                <div class="leads-inbox-item-main">
                                    <label class="leads-inbox-item-select leads-inbox-item-select--leading" on:click|stopPropagation>
                                        <input
                                            type="checkbox"
                                            checked={selectedLeadKeysSet.has(normalizeString(lead.key))}
                                            aria-label={`Select ${lead.sourceLabel} lead`}
                                            on:click|stopPropagation
                                            on:keydown|stopPropagation
                                            on:change|stopPropagation={() => toggleLeadBulkSelection(lead.key)}
                                        />
                                    </label>

                                    <div class="leads-inbox-item-content">
                                        <div class="leads-inbox-item-head">
                                            <div class="leads-inbox-item-badges">
                                                <span class={`label label-sm ${lead.sourceBadgeClass}`}>{lead.sourceLabel}</span>
                                                {#if lead.statusLabel}
                                                    <span class={`label label-sm ${lead.statusBadgeClass}`}>{lead.statusLabel}</span>
                                                {/if}
                                            </div>
                                            <span class="txt-xs txt-hint leads-inbox-item-date">{formatDateTime(lead.created)}</span>
                                        </div>

                                        <div class="leads-inbox-item-title">{lead.identity}</div>

                                        {#if lead.email || lead.phone}
                                            <div class="leads-inbox-item-contact txt-xs txt-hint">
                                                <span class="leads-inbox-inline-label">Contact:</span>
                                                {#if lead.email}
                                                    <span>{lead.email}</span>
                                                {/if}
                                                {#if lead.email && lead.phone}
                                                    <span aria-hidden="true" class="leads-inbox-item-contact-separator">&middot;</span>
                                                {/if}
                                                {#if lead.phone}
                                                    <span>{lead.phone}</span>
                                                {/if}
                                            </div>
                                        {/if}

                                        <div class="leads-inbox-item-message-row txt-sm">
                                            <span class="leads-inbox-item-message-label leads-inbox-inline-label txt-xs txt-hint">Message:</span>
                                            <span class="leads-inbox-item-preview">{resolveLeadPreviewText(lead)}</span>
                                        </div>

                                        {#if lead.notes}
                                            <div class="leads-inbox-item-note txt-xs">
                                                <span class="leads-inbox-item-note-label">Note:</span>
                                                <span class="leads-inbox-item-note-text">{truncate(lead.notes, 120)}</span>
                                            </div>
                                        {/if}

                                        <div class="leads-inbox-item-footer txt-xs txt-hint">
                                            <div class="leads-inbox-item-attribution">
                                                <span class="leads-inbox-inline-label">Origin:</span>
                                                <span class="leads-inbox-item-attribution-main">{resolveLeadAttribution(lead)}</span>
                                                {#if resolveLeadLocationHint(lead)}
                                                    <span class="leads-inbox-item-attribution-context">{resolveLeadLocationHint(lead)}</span>
                                                {/if}
                                            </div>
                                            <div class="leads-inbox-item-last-contact">
                                                Last contact: {lead.lastContactedAt ? resolveLastContactedLabel(lead.lastContactedAt) : "Not contacted yet"}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </article>
                        {/each}
                    </div>
                    <div class="leads-results-footer">
                        {#if canLoadMoreLeads}
                            <span class="txt-sm txt-hint">Showing {visibleLeadsCount} of {filteredLeads.length} leads</span>
                            <button
                                type="button"
                                class="btn btn-sm btn-outline"
                                on:click={loadMoreLeads}
                            >
                                <span class="txt">Load more</span>
                            </button>
                        {:else}
                            <span class="txt-sm txt-hint">Showing all {visibleLeadsCount} leads</span>
                        {/if}
                    </div>
                {/if}

                {#if selectedLeadsCount > 0}
                    <div class="leads-bulk-popover" role="status" aria-live="polite">
                        <span class="leads-bulk-summary">Selected {selectedLeadsCount} lead(s)</span>
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={isBulkUpdatingLeads}
                            on:click={clearLeadBulkSelection}
                        >
                            <span class="txt">Clear selection</span>
                        </button>
                        {#if normalizedLeadsView === "archived"}
                            <button
                                type="button"
                                class="btn btn-sm btn-outline"
                                class:btn-loading={isBulkUpdatingLeads && bulkLeadActionKey === "moveInbox"}
                                disabled={isBulkUpdatingLeads || !bulkMoveInboxEligibleCount}
                                on:click={moveSelectedLeadsToInbox}
                            >
                                <span class="txt">Move to inbox selected</span>
                            </button>
                        {:else}
                            <button
                                type="button"
                                class="btn btn-sm btn-outline"
                                class:btn-loading={isBulkUpdatingLeads && bulkLeadActionKey === "markRead"}
                                disabled={isBulkUpdatingLeads || !bulkMarkReadEligibleCount}
                                on:click={markSelectedLeadsAsRead}
                            >
                                <span class="txt">Mark selected as read</span>
                            </button>
                            <button
                                type="button"
                                class="btn btn-sm btn-outline"
                                class:btn-loading={isBulkUpdatingLeads && bulkLeadActionKey === "archive"}
                                disabled={isBulkUpdatingLeads || !bulkArchiveEligibleCount}
                                on:click={archiveSelectedLeads}
                            >
                                <span class="txt">Archive selected</span>
                            </button>
                        {/if}
                    </div>
                {/if}
            </div>

            {#if isDesktopMasterDetail}
                    <aside class="leads-detail-rail" aria-live="polite">
                        {#if selectedLead}
                            <div class="lead-detail-layout">
                                <section class="lead-detail-section lead-rail-block lead-rail-block--summary">
                                    <div class="lead-detail-section-head lead-rail-head">
                                        <div class="lead-rail-head-main">
                                            <h5 class="m-0">Lead summary</h5>
                                            <p class="txt-sm txt-hint m-b-0 lead-rail-helper">
                                                Snapshot of this lead and where it came from.
                                            </p>
                                        </div>
                                    </div>
                                    <div class="lead-detail-head">
                                        <div class="lead-detail-badges">
                                            <span class={`label label-sm ${selectedLead.sourceBadgeClass}`}>
                                                {selectedLead.sourceLabel}
                                            </span>
                                            {#if selectedLead.statusLabel}
                                                <span class={`label label-sm ${selectedLead.statusBadgeClass}`}>
                                                    {selectedLead.statusLabel}
                                                </span>
                                            {/if}
                                        </div>
                                        <div class="lead-detail-identity">{selectedLead.identity}</div>
                                    </div>

                                    <div class="lead-summary-v2">
                                        <div class="lead-summary-v2-group lead-summary-v2-group--compact">
                                            <div class="lead-summary-v2-row">
                                                <span class="txt-xs txt-hint lead-summary-v2-key">Created</span>
                                                <span class="txt-sm lead-summary-v2-value">{formatDateTime(selectedLead.created)}</span>
                                            </div>
                                            <div class="lead-summary-v2-row">
                                                <span class="txt-xs txt-hint lead-summary-v2-key">Last contacted</span>
                                                <span class="txt-sm lead-summary-v2-value">{selectedLeadLastContactedDisplay}</span>
                                            </div>
                                            <div class="lead-summary-v2-row">
                                                <span class="txt-xs txt-hint lead-summary-v2-key">Contact</span>
                                                <span class="txt-sm lead-summary-v2-value">
                                                    {#if selectedLead.email || selectedLead.phone}
                                                        {selectedLead.email || "No email"}{selectedLead.phone ? ` - ${selectedLead.phone}` : ""}
                                                    {:else}
                                                        No contact details available.
                                                    {/if}
                                                </span>
                                            </div>
                                            {#if selectedLead.subject}
                                                <div class="lead-summary-v2-row">
                                                    <span class="txt-xs txt-hint lead-summary-v2-key">Subject</span>
                                                    <span class="txt-sm lead-summary-v2-value">{selectedLead.subject}</span>
                                                </div>
                                            {/if}
                                            <div class="lead-summary-v2-row">
                                                <span class="txt-xs txt-hint lead-summary-v2-key">Attribution</span>
                                                <span class="txt-sm lead-summary-v2-value">{resolveLeadAttribution(selectedLead)}</span>
                                            </div>
                                            {#if resolveLeadLocationHint(selectedLead)}
                                                <div class="lead-summary-v2-row">
                                                    <span class="txt-xs txt-hint lead-summary-v2-key">Context</span>
                                                    <span class="txt-sm lead-summary-v2-value">{resolveLeadLocationHint(selectedLead)}</span>
                                                </div>
                                            {/if}
                                            {#if selectedLead.whatsappTargetPhone}
                                                <div class="lead-summary-v2-row">
                                                    <span class="txt-xs txt-hint lead-summary-v2-key">WhatsApp target phone</span>
                                                    <span class="txt-sm lead-summary-v2-value">{selectedLead.whatsappTargetPhone}</span>
                                                </div>
                                            {/if}
                                        </div>

                                        <div class="lead-summary-v2-group lead-summary-v2-group--long">
                                            <div class="lead-summary-v2-long-field">
                                                <span class="txt-xs txt-hint lead-summary-v2-key">Message preview</span>
                                                <p class="txt-sm m-b-0 lead-summary-v2-long-copy">{resolveLeadPreviewText(selectedLead)}</p>
                                            </div>
                                            {#if selectedLead.notes}
                                                <div class="lead-summary-v2-long-field">
                                                    <span class="txt-xs txt-hint lead-summary-v2-key">Notes</span>
                                                    <p class="txt-sm m-b-0 lead-summary-v2-long-copy">{selectedLead.notes}</p>
                                                </div>
                                            {/if}
                                        </div>
                                    </div>
                                    <div class="lead-summary-collapsible">
                                        <button
                                            type="button"
                                            class="lead-summary-collapsible-toggle"
                                            aria-expanded={isLeadFollowUpOpen}
                                            on:click={() => (isLeadFollowUpOpen = !isLeadFollowUpOpen)}
                                        >
                                            <span class="lead-summary-collapsible-heading">
                                                <span class="lead-summary-collapsible-title">Follow-up</span>
                                                <span class="txt-xs txt-hint lead-summary-collapsible-helper">
                                                    Update status, keep notes, and track contact progress.
                                                </span>
                                            </span>
                                            <i class={`lead-summary-collapsible-icon ${isLeadFollowUpOpen ? "ri-arrow-up-s-line" : "ri-arrow-down-s-line"}`} />
                                        </button>

                                        {#if !isLeadFollowUpOpen}
                                            <p class="txt-xs txt-hint m-b-0 lead-summary-collapsible-preview">
                                                {#if leadNotesDirty}
                                                    Unsaved note changes.
                                                {:else}
                                                    Last contacted: {selectedLeadLastContactedDisplay}.
                                                {/if}
                                            </p>
                                        {:else}
                                            <div class="lead-summary-collapsible-content">
                                                <div class="lead-detail-actions-block lead-actions-group">
                                                    <div class="txt-xs txt-hint txt-uppercase lead-actions-title">Status actions</div>
                                                    <div class="lead-detail-actions">
                                                        {#if canMoveSelectedLeadToInbox}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm"
                                                                class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                                                disabled={isUpdatingLeadStatus}
                                                                on:click={moveSelectedLeadToInbox}
                                                            >
                                                                <span class="txt">Move to inbox</span>
                                                            </button>
                                                        {:else if canToggleSelectedLeadStatus}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm"
                                                                class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                                                disabled={isUpdatingLeadStatus}
                                                                on:click={() => setSelectedLeadStatus(selectedLeadToggleActionTarget)}
                                                            >
                                                                <span class="txt">{selectedLeadToggleActionLabel}</span>
                                                            </button>
                                                        {/if}
                                                        {#if canArchiveSelectedLead}
                                                            <button
                                                                type="button"
                                                                class="btn btn-outline btn-sm"
                                                                class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                                                disabled={isUpdatingLeadStatus}
                                                                on:click={archiveSelectedLead}
                                                            >
                                                                <span class="txt">Archive</span>
                                                            </button>
                                                        {/if}
                                                    </div>

                                                    {#if selectedLead.sourceKey === "whatsapp" && !selectedLeadStatusSupport.supportsToggle}
                                                        <p class="txt-xs txt-hint m-b-0 lead-action-note">
                                                            Status actions are not available for WhatsApp interactions yet.
                                                        </p>
                                                    {:else if selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsToggle}
                                                        <p class="txt-xs txt-hint m-b-0 lead-action-note">
                                                            {#if isSelectedLeadArchived}
                                                                Move to inbox is not available for this lead source yet.
                                                            {:else}
                                                                Status actions are not available for this lead source yet.
                                                            {/if}
                                                        </p>
                                                    {:else if !isSelectedLeadArchived && selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsArchive}
                                                        <p class="txt-xs txt-hint m-b-0 lead-action-note">
                                                            Archive is not available for this source yet.
                                                        </p>
                                                    {/if}
                                                </div>

                                                <div class="lead-detail-actions-block lead-actions-group">
                                                    <div class="txt-xs txt-hint txt-uppercase lead-actions-title">Notes</div>
                                                    {#if canSaveSelectedLeadNote || canMarkSelectedLeadContacted}
                                                        <div class="lead-followup-stack">
                                                            <textarea
                                                                id="lead-followup-notes"
                                                                class="input lead-followup-notes"
                                                                rows="4"
                                                                placeholder="Add notes about next steps, context, or follow-up outcomes..."
                                                                aria-label="Notes"
                                                                bind:value={leadNotesDraft}
                                                                disabled={isSavingLeadFollowUp || !canSaveSelectedLeadNote}
                                                            />

                                                            {#if leadFollowUpError}
                                                                <p class="txt-xs txt-danger m-b-0">{leadFollowUpError}</p>
                                                            {/if}

                                                            <div class="lead-detail-actions">
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-sm"
                                                                    class:btn-loading={isSavingLeadFollowUp}
                                                                    disabled={!canSaveSelectedLeadNote || isSavingLeadFollowUp || !leadNotesDirty}
                                                                    on:click={saveSelectedLeadNote}
                                                                >
                                                                    <span class="txt">Save note</span>
                                                                </button>
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-outline btn-sm"
                                                                    class:btn-loading={isSavingLeadFollowUp}
                                                                    disabled={!canMarkSelectedLeadContacted || isSavingLeadFollowUp}
                                                                    on:click={markSelectedLeadContactedNow}
                                                                >
                                                                    <span class="txt">Mark contacted now</span>
                                                                </button>
                                                            </div>
                                                        </div>
                                                    {:else}
                                                        <p class="txt-sm txt-hint m-b-0">
                                                            Follow-up fields are not available for this lead source yet.
                                                        </p>
                                                    {/if}

                                                    <p class="txt-xs txt-hint m-b-0">Last contacted: {selectedLeadLastContactedDisplay}</p>
                                                </div>
                                            </div>
                                        {/if}
                                    </div>

                                    <div class="lead-summary-collapsible">
                                        <button
                                            type="button"
                                            class="lead-summary-collapsible-toggle"
                                            aria-expanded={isLeadUtilitiesOpen}
                                            on:click={() => (isLeadUtilitiesOpen = !isLeadUtilitiesOpen)}
                                        >
                                            <span class="lead-summary-collapsible-heading">
                                                <span class="lead-summary-collapsible-title">Utilities</span>
                                                <span class="txt-xs txt-hint lead-summary-collapsible-helper">
                                                    Copy and open quick communication actions.
                                                </span>
                                            </span>
                                            <i class={`lead-summary-collapsible-icon ${isLeadUtilitiesOpen ? "ri-arrow-up-s-line" : "ri-arrow-down-s-line"}`} />
                                        </button>

                                        {#if !isLeadUtilitiesOpen}
                                            <p class="txt-xs txt-hint m-b-0 lead-summary-collapsible-preview">
                                                {#if selectedLead.email || selectedLead.phone || selectedLead.message || selectedLead.whatsappTargetMessage || selectedLeadMailto || selectedLeadWhatsAppLink}
                                                    Quick actions are available for this lead.
                                                {:else}
                                                    No utility actions available for this lead.
                                                {/if}
                                            </p>
                                        {:else}
                                            <div class="lead-summary-collapsible-content">
                                                <div class="lead-detail-actions-block lead-actions-group">
                                                    <div class="txt-xs txt-hint txt-uppercase lead-actions-title">Utilities</div>
                                                    <div class="lead-detail-actions">
                                                        {#if selectedLead.email}
                                                            <button type="button" class="btn btn-outline btn-sm" on:click={() => copyValue(selectedLead.email, "Email")}>
                                                                <span class="txt">Copy email</span>
                                                            </button>
                                                        {/if}
                                                        {#if selectedLead.phone}
                                                            <button type="button" class="btn btn-outline btn-sm" on:click={() => copyValue(selectedLead.phone, "Phone")}>
                                                                <span class="txt">Copy phone</span>
                                                            </button>
                                                        {/if}
                                                        {#if selectedLead.message || selectedLead.whatsappTargetMessage}
                                                            <button
                                                                type="button"
                                                                class="btn btn-outline btn-sm"
                                                                on:click={() => copyValue(selectedLead.whatsappTargetMessage || selectedLead.message, "Message")}
                                                            >
                                                                <span class="txt">Copy message</span>
                                                            </button>
                                                        {/if}
                                                        {#if selectedLeadMailto}
                                                            <a href={selectedLeadMailto} class="btn btn-sm">
                                                                <span class="txt">Open email</span>
                                                            </a>
                                                        {/if}
                                                        {#if selectedLeadWhatsAppLink}
                                                            <a href={selectedLeadWhatsAppLink} target="_blank" rel="noopener noreferrer" class="btn btn-sm">
                                                                <span class="txt">Open WhatsApp</span>
                                                            </a>
                                                        {/if}
                                                        <button
                                                            type="button"
                                                            class="btn btn-outline btn-sm"
                                                            class:btn-loading={isInvitingLeadToNewsletter}
                                                            disabled={!selectedLeadCanInviteToNewsletter}
                                                            on:click={inviteSelectedLeadToNewsletter}
                                                        >
                                                            <span class="txt">Invite to newsletter</span>
                                                        </button>
                                                    </div>
                                                    {#if selectedLeadInviteUnavailableMessage}
                                                        <p class="txt-xs txt-hint m-b-0 lead-action-note">{selectedLeadInviteUnavailableMessage}</p>
                                                    {/if}
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                </section>

                                <section class="lead-detail-section lead-rail-block lead-rail-block--health">
                                    <div class="lead-health-head">
                                        <div class="lead-health-main">
                                            <h5 class="m-0">Lead health</h5>
                                            <p class="txt-sm txt-hint m-b-0 lead-rail-helper">
                                                Review lead inbox readiness, follow-up gaps, and contact quality.
                                            </p>
                                        </div>
                                        <div class="lead-health-meta">
                                            <span class={`label label-sm lead-health-status-pill ${leadHealthState.badgeClass}`}>
                                                {leadHealthState.label}
                                            </span>
                                            <span class="summary-pill lead-health-summary-pill" class:warning={leadHealthWarnings.length > 0}>
                                                {leadHealthWarnings.length} warnings | {leadHealthSuggestions.length} suggestions
                                            </span>
                                        </div>
                                    </div>

                                    {#if leadHealthWarnings.length}
                                        <div class="lead-health-group">
                                            <div class="lead-health-group-title">Warnings</div>
                                            <div class="lead-health-check-list">
                                                {#each leadHealthWarnings as warning}
                                                    <div class="lead-health-check-item warning">
                                                        <span class="label label-sm lead-health-check-pill warning">Warning</span>
                                                        <span class="lead-health-check-message">{warning}</span>
                                                    </div>
                                                {/each}
                                            </div>
                                        </div>
                                    {/if}

                                    {#if leadHealthSuggestions.length}
                                        <div class="lead-health-group">
                                            <div class="lead-health-group-title">Suggestions</div>
                                            <div class="lead-health-check-list">
                                                {#each leadHealthSuggestions as suggestion}
                                                    <div class="lead-health-check-item">
                                                        <span class="label label-sm lead-health-check-pill">Info</span>
                                                        <span class="lead-health-check-message">{suggestion}</span>
                                                    </div>
                                                {/each}
                                            </div>
                                        </div>
                                    {/if}

                                    {#if !leadHealthWarnings.length && !leadHealthSuggestions.length}
                                        <p class="txt-sm txt-hint m-b-0">No lead data available for this website yet.</p>
                                    {/if}
                                </section>

                            </div>
                        {:else}
                            <div class="lead-detail-layout">
                                <div class="empty-state m-b-0 leads-detail-empty-state">
                                    Select a lead to view details.
                                </div>

                                <section class="lead-detail-section lead-rail-block lead-rail-block--health">
                                    <div class="lead-health-head">
                                        <div class="lead-health-main">
                                            <h5 class="m-0">Lead health</h5>
                                            <p class="txt-sm txt-hint m-b-0 lead-rail-helper">
                                                Review lead inbox readiness, follow-up gaps, and contact quality.
                                            </p>
                                        </div>
                                        <div class="lead-health-meta">
                                            <span class={`label label-sm lead-health-status-pill ${leadHealthState.badgeClass}`}>
                                                {leadHealthState.label}
                                            </span>
                                            <span class="summary-pill lead-health-summary-pill" class:warning={leadHealthWarnings.length > 0}>
                                                {leadHealthWarnings.length} warnings | {leadHealthSuggestions.length} suggestions
                                            </span>
                                        </div>
                                    </div>

                                    {#if leadHealthWarnings.length}
                                        <div class="lead-health-group">
                                            <div class="lead-health-group-title">Warnings</div>
                                            <div class="lead-health-check-list">
                                                {#each leadHealthWarnings as warning}
                                                    <div class="lead-health-check-item warning">
                                                        <span class="label label-sm lead-health-check-pill warning">Warning</span>
                                                        <span class="lead-health-check-message">{warning}</span>
                                                    </div>
                                                {/each}
                                            </div>
                                        </div>
                                    {/if}

                                    {#if leadHealthSuggestions.length}
                                        <div class="lead-health-group">
                                            <div class="lead-health-group-title">Suggestions</div>
                                            <div class="lead-health-check-list">
                                                {#each leadHealthSuggestions as suggestion}
                                                    <div class="lead-health-check-item">
                                                        <span class="label label-sm lead-health-check-pill">Info</span>
                                                        <span class="lead-health-check-message">{suggestion}</span>
                                                    </div>
                                                {/each}
                                            </div>
                                        </div>
                                    {/if}

                                    {#if !leadHealthWarnings.length && !leadHealthSuggestions.length}
                                        <p class="txt-sm txt-hint m-b-0">No lead data available for this website yet.</p>
                                    {/if}
                                </section>
                            </div>
                        {/if}
                    </aside>
            {/if}
        </div>
    </section>

    {#if !isDesktopMasterDetail}
        <OverlayPanel
            bind:active={isLeadDetailsActive}
            class="overlay-panel-lg leads-detail-panel"
            overlayClose={true}
            escClose={true}
            on:hide={closeLeadDetails}
        >
            <svelte:fragment slot="header">
                <h4>Lead details</h4>
            </svelte:fragment>

            {#if selectedLead}
                <div class="lead-detail-layout">
                    <div class="lead-detail-head">
                        <div class="lead-detail-badges">
                            <span class={`label label-sm ${selectedLead.sourceBadgeClass}`}>
                                {selectedLead.sourceLabel}
                            </span>
                            {#if selectedLead.statusLabel}
                                <span class={`label label-sm ${selectedLead.statusBadgeClass}`}>
                                    {selectedLead.statusLabel}
                                </span>
                            {/if}
                        </div>
                        <div class="lead-detail-identity">{selectedLead.identity}</div>
                    </div>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Contact</h5>
                        </div>
                        <div class="lead-detail-grid">
                            <div class="lead-detail-row">
                                <span class="txt-xs txt-hint">Identity</span>
                                <span class="txt-sm">{selectedLead.identity}</span>
                            </div>
                            {#if selectedLead.name && selectedLead.name !== selectedLead.identity}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Name</span>
                                    <span class="txt-sm">{selectedLead.name}</span>
                                </div>
                            {/if}
                            {#if selectedLead.email}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Email</span>
                                    <span class="txt-sm">{selectedLead.email}</span>
                                </div>
                            {/if}
                            {#if selectedLead.phone}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Phone</span>
                                    <span class="txt-sm">{selectedLead.phone}</span>
                                </div>
                            {/if}
                        </div>
                    </section>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Message</h5>
                        </div>
                        <div class="lead-detail-grid">
                            {#if selectedLead.subject}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Subject</span>
                                    <span class="txt-sm">{selectedLead.subject}</span>
                                </div>
                            {/if}
                            {#if selectedLead.message}
                                <div class="lead-detail-row lead-detail-row-block">
                                    <span class="txt-xs txt-hint">Message</span>
                                    <p class="txt-sm m-b-0">{selectedLead.message}</p>
                                </div>
                            {/if}
                            {#if selectedLead.whatsappTargetMessage}
                                <div class="lead-detail-row lead-detail-row-block">
                                    <span class="txt-xs txt-hint">WhatsApp target message</span>
                                    <p class="txt-sm m-b-0">{selectedLead.whatsappTargetMessage}</p>
                                </div>
                            {/if}
                            {#if !selectedLead.subject && !selectedLead.message && !selectedLead.whatsappTargetMessage}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Message</span>
                                    <span class="txt-sm txt-hint">No message details are available for this lead yet.</span>
                                </div>
                            {/if}
                        </div>
                    </section>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Source</h5>
                        </div>
                        <div class="lead-detail-grid">
                            <div class="lead-detail-row">
                                <span class="txt-xs txt-hint">Attribution</span>
                                <span class="txt-sm">{resolveLeadAttribution(selectedLead)}</span>
                            </div>
                            <div class="lead-detail-row">
                                <span class="txt-xs txt-hint">Created</span>
                                <span class="txt-sm">{formatDateTime(selectedLead.created)}</span>
                            </div>
                            <div class="lead-detail-row">
                                <span class="txt-xs txt-hint">Last contacted</span>
                                <span class="txt-sm">{selectedLeadLastContactedDisplay}</span>
                            </div>
                            {#if resolveLeadLocationHint(selectedLead)}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Context</span>
                                    <span class="txt-sm">{resolveLeadLocationHint(selectedLead)}</span>
                                </div>
                            {/if}
                            {#if selectedLead.whatsappTargetPhone}
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">WhatsApp target phone</span>
                                    <span class="txt-sm">{selectedLead.whatsappTargetPhone}</span>
                                </div>
                            {/if}
                        </div>
                    </section>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Follow-up</h5>
                        </div>
                        <p class="txt-sm txt-hint m-b-0">Keep internal notes about this lead.</p>

                        {#if canSaveSelectedLeadNote || canMarkSelectedLeadContacted}
                            <div class="lead-followup-stack">
                                <label class="txt-xs txt-hint m-b-0" for="lead-followup-notes-mobile">Notes</label>
                                <textarea
                                    id="lead-followup-notes-mobile"
                                    class="input lead-followup-notes"
                                    rows="4"
                                    placeholder="Add notes about next steps, context, or follow-up outcomes..."
                                    bind:value={leadNotesDraft}
                                    disabled={isSavingLeadFollowUp || !canSaveSelectedLeadNote}
                                />
                                {#if leadFollowUpError}
                                    <p class="txt-xs txt-danger m-b-0">{leadFollowUpError}</p>
                                {/if}
                                <div class="lead-detail-actions">
                                    <button
                                        type="button"
                                        class="btn btn-sm"
                                        class:btn-loading={isSavingLeadFollowUp}
                                        disabled={!canSaveSelectedLeadNote || isSavingLeadFollowUp || !leadNotesDirty}
                                        on:click={saveSelectedLeadNote}
                                    >
                                        <span class="txt">Save note</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        class:btn-loading={isSavingLeadFollowUp}
                                        disabled={!canMarkSelectedLeadContacted || isSavingLeadFollowUp}
                                        on:click={markSelectedLeadContactedNow}
                                    >
                                        <span class="txt">Mark contacted now</span>
                                    </button>
                                </div>
                                <p class="txt-xs txt-hint m-b-0">Last contacted: {selectedLeadLastContactedDisplay}</p>
                            </div>
                        {:else}
                            <p class="txt-sm txt-hint m-b-0">
                                Follow-up fields are not available for this lead source yet.
                            </p>
                        {/if}
                    </section>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Notification setup</h5>
                        </div>
                        <p class="txt-sm txt-hint m-b-0">
                            Based on current Website Settings. Per-lead delivery status is not stored yet.
                        </p>

                        {#if selectedLeadNotificationSetup.available}
                            <div class="lead-detail-grid lead-notification-grid">
                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Current setup</span>
                                    <div class="lead-notification-badges">
                                        <span class={`label label-sm ${selectedLeadNotificationSetup.notificationsEnabled ? "label-success" : "label-danger"}`}>
                                            Email notifications: {selectedLeadNotificationSetup.notificationsEnabled ? "Enabled" : "Disabled"}
                                        </span>
                                        <span class={`label label-sm ${selectedLeadNotificationSetup.hasRecipients ? "label-info" : "label-warning"}`}>
                                            {selectedLeadNotificationSetup.hasRecipients ? "Recipients configured" : "Missing recipients"}
                                        </span>
                                    </div>
                                </div>

                                <div class="lead-detail-row">
                                    <span class="txt-xs txt-hint">Recipients</span>
                                    <span class="txt-sm">
                                        {selectedLeadNotificationSetup.recipientsTotal} total ({selectedLeadNotificationSetup.toCount} To, {selectedLeadNotificationSetup.ccCount} CC)
                                    </span>
                                </div>

                                {#if selectedLeadNotificationSetup.featureAvailabilityLabel}
                                    <div class="lead-detail-row">
                                        <span class="txt-xs txt-hint">Feature availability</span>
                                        <span class={`label label-sm ${selectedLeadNotificationSetup.featureAvailable ? "label-success" : "label-warning"}`}>
                                            {selectedLeadNotificationSetup.featureAvailabilityLabel}
                                        </span>
                                    </div>
                                {/if}
                            </div>
                        {:else}
                            <p class="txt-sm txt-hint m-b-0">
                                {selectedLeadNotificationSetup.unavailableMessage}
                            </p>
                        {/if}
                    </section>

                    <section class="lead-detail-section">
                        <div class="lead-detail-section-head">
                            <h5 class="m-0">Actions</h5>
                        </div>

                        <div class="lead-detail-actions-block">
                            <div class="txt-xs txt-hint txt-uppercase">Status actions</div>
                            <div class="lead-detail-actions">
                                {#if canMoveSelectedLeadToInbox}
                                    <button
                                        type="button"
                                        class="btn btn-sm"
                                        class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                        disabled={isUpdatingLeadStatus}
                                        on:click={moveSelectedLeadToInbox}
                                    >
                                        <span class="txt">Move to inbox</span>
                                    </button>
                                {:else if canToggleSelectedLeadStatus}
                                    <button
                                        type="button"
                                        class="btn btn-sm"
                                        class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                        disabled={isUpdatingLeadStatus}
                                        on:click={() => setSelectedLeadStatus(selectedLeadToggleActionTarget)}
                                    >
                                        <span class="txt">{selectedLeadToggleActionLabel}</span>
                                    </button>
                                {/if}
                                {#if canArchiveSelectedLead}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        class:btn-loading={isUpdatingLeadStatus && updatingLeadStatusKey === selectedLead.key}
                                        disabled={isUpdatingLeadStatus}
                                        on:click={archiveSelectedLead}
                                    >
                                        <span class="txt">Archive</span>
                                    </button>
                                {/if}
                            </div>

                            {#if selectedLead.sourceKey === "whatsapp" && !selectedLeadStatusSupport.supportsToggle}
                                <p class="txt-sm txt-hint m-b-0">
                                    Status actions are not available for WhatsApp interactions yet.
                                </p>
                            {:else if selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsToggle}
                                <p class="txt-sm txt-hint m-b-0">
                                    {#if isSelectedLeadArchived}
                                        Move to inbox is not available for this lead source yet.
                                    {:else}
                                        Status actions are not available for this lead source yet.
                                    {/if}
                                </p>
                            {:else if !isSelectedLeadArchived && selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsArchive}
                                <p class="txt-sm txt-hint m-b-0">
                                    Archive is not available for this source yet.
                                </p>
                            {/if}
                        </div>

                        <div class="lead-detail-actions-block">
                            <div class="txt-xs txt-hint txt-uppercase">Utilities</div>
                            <div class="lead-detail-actions">
                                {#if selectedLead.email}
                                    <button type="button" class="btn btn-outline btn-sm" on:click={() => copyValue(selectedLead.email, "Email")}>
                                        <span class="txt">Copy email</span>
                                    </button>
                                {/if}
                                {#if selectedLead.phone}
                                    <button type="button" class="btn btn-outline btn-sm" on:click={() => copyValue(selectedLead.phone, "Phone")}>
                                        <span class="txt">Copy phone</span>
                                    </button>
                                {/if}
                                {#if selectedLead.message || selectedLead.whatsappTargetMessage}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        on:click={() => copyValue(selectedLead.whatsappTargetMessage || selectedLead.message, "Message")}
                                    >
                                        <span class="txt">Copy message</span>
                                    </button>
                                {/if}
                                {#if selectedLeadMailto}
                                    <a href={selectedLeadMailto} class="btn btn-sm">
                                        <span class="txt">Open email</span>
                                    </a>
                                {/if}
                                {#if selectedLeadWhatsAppLink}
                                    <a href={selectedLeadWhatsAppLink} target="_blank" rel="noopener noreferrer" class="btn btn-sm">
                                        <span class="txt">Open WhatsApp</span>
                                    </a>
                                {/if}
                                <button
                                    type="button"
                                    class="btn btn-outline btn-sm"
                                    class:btn-loading={isInvitingLeadToNewsletter}
                                    disabled={!selectedLeadCanInviteToNewsletter}
                                    on:click={inviteSelectedLeadToNewsletter}
                                >
                                    <span class="txt">Invite to newsletter</span>
                                </button>
                            </div>
                            {#if selectedLeadInviteUnavailableMessage}
                                <p class="txt-xs txt-hint m-b-0 lead-action-note">{selectedLeadInviteUnavailableMessage}</p>
                            {/if}
                        </div>
                    </section>
                </div>
            {:else}
                <div class="empty-state m-b-0">
                    Select a lead to view details.
                </div>
            {/if}

            <svelte:fragment slot="footer">
                <button type="button" class="btn btn-outline btn-sm" on:click={closeLeadDetails}>
                    <span class="txt">Close</span>
                </button>
            </svelte:fragment>
        </OverlayPanel>
    {/if}
</PageWrapper>

<style>
    .leads-head.operations-head .head-description {
        max-width: 460px;
    }

    .leads-head.operations-head .leads-view-toggle {
        flex: 0 0 auto;
    }

    .leads-head.operations-head .summary-badges {
        justify-content: flex-end;
    }

    .leads-body {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .head-tools {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
        justify-content: space-between;
    }

    .leads-toolbar-row {
        display: grid;
        grid-template-columns: minmax(180px, 2.3fr) repeat(4, minmax(78px, 0.9fr)) auto;
        gap: 10px;
        align-items: end;
    }

    .leads-left-column {
        display: flex;
        flex-direction: column;
        gap: 10px;
        min-width: 0;
    }

    .leads-controls {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .leads-filter-cell {
        display: flex;
        flex-direction: column;
        gap: 5px;
        min-width: 0;
    }

    .leads-filter-cell .input {
        width: 100%;
        min-width: 0;
    }

    .leads-filter-cell--search {
        min-width: 0;
    }

    .leads-filter-cell--select,
    .leads-filter-cell--sort {
        min-width: 78px;
    }

    .leads-filter-cell--actions {
        display: inline-flex;
        flex-direction: row;
        align-items: flex-end;
        justify-content: flex-end;
        gap: 8px;
        flex-wrap: nowrap;
        align-self: end;
    }

    .leads-filter-cell--actions .btn {
        white-space: nowrap;
    }

    .leads-inbox-layout {
        display: grid;
        grid-template-columns: minmax(0, 1fr) clamp(280px, 24vw, 350px);
        gap: 12px;
        align-items: start;
    }

    .leads-inbox-list {
        display: grid;
        grid-template-columns: 1fr;
        gap: 8px;
        min-width: 0;
    }

    .leads-results-footer {
        margin-top: 10px;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .leads-inbox-item {
        border: 2px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 9px 11px;
        display: flex;
        flex-direction: column;
        gap: 5px;
        cursor: pointer;
        transition: border-color var(--baseAnimationSpeed), background-color var(--baseAnimationSpeed);
    }

    .leads-inbox-item:hover,
    .leads-inbox-item:focus-visible {
        border-color: color-mix(in srgb, var(--txtPrimaryColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 18%, transparent);
        outline: none;
    }

    .leads-inbox-item:active {
        border-color: var(--txtPrimaryColor);
    }

    .leads-inbox-item.selected {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 28%, transparent);
    }

    .leads-inbox-item.bulk-selected {
        border-color: color-mix(in srgb, var(--txtPrimaryColor) 55%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 22%, transparent);
    }

    .leads-inbox-item.selected.bulk-selected {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 34%, transparent);
    }

    .leads-inbox-item-main {
        min-width: 0;
        width: 100%;
        display: inline-flex;
        align-items: flex-start;
        gap: 10px;
    }

    .leads-inbox-item-content {
        min-width: 0;
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 5px;
    }

    .leads-inbox-item-select {
        display: inline-flex;
        align-items: center;
        justify-content: center;
    }

    .leads-inbox-item-select input {
        margin: 0;
    }

    .leads-inbox-item-select--leading {
        align-self: flex-start;
        margin-top: 1px;
    }

    .leads-inbox-item-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 6px;
    }

    .leads-inbox-item-badges {
        display: inline-flex;
        align-items: center;
        gap: 5px;
        flex-wrap: wrap;
    }

    .leads-inbox-item-date {
        white-space: nowrap;
    }

    .leads-inbox-item-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .leads-inbox-item-secondary {
        margin-top: -2px;
    }

    .leads-inbox-item-contact {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        min-width: 0;
    }

    .leads-inbox-item-contact span {
        min-width: 0;
    }

    .leads-inbox-inline-label {
        flex: 0 0 auto;
        color: var(--txtHintColor);
        font-weight: 600;
        letter-spacing: 0.01em;
    }

    .leads-inbox-item-preview {
        color: var(--txtPrimaryColor);
        min-width: 0;
        flex: 1 1 auto;
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
        overflow: hidden;
    }

    .leads-inbox-item-message-row {
        display: flex;
        align-items: baseline;
        gap: 6px;
        min-width: 0;
    }

    .leads-inbox-item-message-label {
        white-space: nowrap;
    }

    .leads-inbox-item-note {
        display: flex;
        align-items: flex-start;
        gap: 6px;
        margin-top: 0;
    }

    .leads-inbox-item-note-label {
        flex: 0 0 auto;
        color: var(--txtHintColor);
        font-weight: 600;
        letter-spacing: 0.01em;
    }

    .leads-inbox-item-note-text {
        min-width: 0;
        color: color-mix(in srgb, var(--txtHintColor) 92%, var(--txtPrimaryColor));
        line-height: 1.3;
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 1;
        overflow: hidden;
    }

    .leads-inbox-item-footer {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        margin-top: 1px;
    }

    .leads-inbox-item-attribution {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        min-width: 0;
        flex: 1 1 auto;
    }

    .leads-inbox-item-attribution-main {
        color: var(--txtPrimaryColor);
        font-weight: 500;
        min-width: 0;
    }

    .leads-inbox-item-attribution-context {
        min-width: 0;
    }

    .leads-inbox-item-attribution-context::before {
        content: "-";
        margin-right: 6px;
        color: var(--txtDisabledColor);
    }

    .leads-inbox-item-last-contact {
        margin-top: 0;
        white-space: nowrap;
        flex: 0 0 auto;
    }

    .leads-inbox-item-contact-separator {
        color: var(--txtDisabledColor);
        flex: 0 0 auto;
    }

    .leads-bulk-popover {
        position: fixed;
        left: 50%;
        bottom: 18px;
        transform: translateX(-50%);
        z-index: 55;
        display: inline-flex;
        align-items: center;
        gap: 8px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseColor);
        box-shadow: 0 12px 30px rgba(0, 0, 0, 0.16);
        padding: 8px 10px;
    }

    .leads-bulk-summary {
        color: var(--txtPrimaryColor);
        font-weight: 600;
        font-size: var(--smFontSize);
        padding: 0 3px;
        white-space: nowrap;
    }

    .leads-detail-rail {
        border-left: 0;
        padding-left: 0;
        min-width: 0;
    }

    .leads-detail-empty-state {
        min-height: 280px;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    :global(.leads-detail-panel .panel-content) {
        padding: calc(var(--baseSpacing) - 8px);
    }

    .lead-detail-layout {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .lead-rail-block {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        padding: 10px;
        background: var(--baseColor);
    }

    .lead-detail-head {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-detail-identity {
        font-size: 15px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .lead-detail-badges {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .lead-detail-section {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-detail-section + .lead-detail-section {
        padding-top: 0;
        border-top: 0;
    }

    .lead-detail-section.lead-rail-block + .lead-detail-section.lead-rail-block {
        padding-top: 10px;
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
    }

    .lead-detail-section-head {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
    }

    .lead-rail-head {
        align-items: flex-start;
    }

    .lead-rail-head-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .lead-rail-helper {
        font-size: 11px;
        line-height: 1.35;
    }

    .lead-summary-v2 {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-summary-v2-group {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .lead-summary-v2-group + .lead-summary-v2-group {
        padding-top: 8px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .lead-summary-v2-row {
        display: grid;
        grid-template-columns: 112px minmax(0, 1fr);
        align-items: start;
        gap: 10px;
    }

    .lead-summary-v2-key {
        line-height: 1.35;
        white-space: nowrap;
    }

    .lead-summary-v2-value {
        min-width: 0;
        color: var(--txtPrimaryColor);
        line-height: 1.35;
        word-break: break-word;
    }

    .lead-summary-v2-long-field {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .lead-summary-v2-long-copy {
        color: var(--txtPrimaryColor);
        line-height: 1.4;
        white-space: pre-wrap;
        word-break: break-word;
    }

    .lead-summary-collapsible {
        margin-top: 2px;
        padding-top: 9px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-summary-collapsible-toggle {
        width: 100%;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseAlt1) 14%, transparent);
        color: inherit;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        padding: 8px 10px;
        text-align: left;
        cursor: pointer;
        transition: border-color 0.15s ease, background-color 0.15s ease;
    }

    .lead-summary-collapsible-toggle:hover {
        border-color: color-mix(in srgb, var(--txtPrimaryColor) 38%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 20%, transparent);
    }

    .lead-summary-collapsible-heading {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .lead-summary-collapsible-title {
        color: var(--txtPrimaryColor);
        font-size: var(--smFontSize);
        line-height: 1.3;
        font-weight: 600;
    }

    .lead-summary-collapsible-helper {
        display: block;
        line-height: 1.3;
    }

    .lead-summary-collapsible-icon {
        flex: 0 0 auto;
        font-size: 1rem;
        color: var(--txtHintColor);
    }

    .lead-summary-collapsible-content {
        display: flex;
        flex-direction: column;
        gap: 7px;
    }

    .lead-summary-collapsible-preview {
        margin-top: 2px;
        padding: 8px 10px;
        border: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseAlt1) 10%, transparent);
        word-break: break-word;
        display: -webkit-box;
        -webkit-line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .lead-detail-grid {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-detail-row {
        display: flex;
        flex-direction: column;
        gap: 4px;
        padding: 8px 10px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: color-mix(in srgb, var(--baseColor) 92%, var(--baseAlt1Color));
    }

    .lead-detail-row-block p {
        white-space: pre-wrap;
    }

    .lead-detail-actions-block {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .lead-detail-actions-block + .lead-detail-actions-block {
        padding-top: 8px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .lead-actions-title {
        letter-spacing: 0.02em;
    }

    .lead-detail-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        padding-top: 2px;
    }

    .lead-action-note {
        line-height: 1.35;
    }

    .lead-notification-badges {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .lead-notification-stack {
        display: flex;
        flex-direction: column;
        gap: 7px;
    }

    .lead-notification-row {
        display: flex;
        flex-direction: column;
        gap: 4px;
        padding-top: 7px;
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .lead-notification-row:first-child {
        border-top: 0;
        padding-top: 0;
    }

    .lead-notification-meta {
        color: var(--txtHintColor);
    }

    .lead-followup-stack {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .lead-followup-notes {
        min-height: 90px;
        resize: vertical;
    }

    .lead-health-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .lead-health-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .lead-health-meta {
        display: inline-flex;
        align-items: center;
        justify-content: flex-end;
        gap: 6px;
        flex-wrap: wrap;
    }

    .lead-health-status-pill {
        --labelHPadding: 8px;
        min-height: 20px;
        color: var(--txtHintColor);
        border-color: color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 18%, var(--baseColor));
        font-weight: 600;
    }

    .lead-health-status-pill.label-success {
        color: color-mix(in srgb, var(--successColor) 85%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--successColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--successColor) 12%, var(--baseColor));
    }

    .lead-health-status-pill.label-warning {
        color: color-mix(in srgb, var(--warningColor) 86%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .lead-health-status-pill.label-info {
        color: color-mix(in srgb, var(--primaryColor) 82%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--primaryColor) 30%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 10%, var(--baseColor));
    }

    .lead-health-summary-pill {
        --labelHPadding: 9px;
        min-height: 20px;
        color: var(--txtHintColor);
        background: color-mix(in srgb, var(--baseAlt1Color) 22%, var(--baseColor));
    }

    .lead-health-summary-pill.warning {
        color: color-mix(in srgb, var(--warningColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .lead-health-group {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .lead-health-group-title {
        color: var(--txtHintColor);
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.02em;
        text-transform: uppercase;
    }

    .lead-health-check-list {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .lead-health-check-item {
        display: flex;
        align-items: flex-start;
        gap: 7px;
        padding: 6px 0;
        font-size: var(--smFontSize);
        line-height: var(--smLineHeight);
        color: var(--txtHintColor);
    }

    .lead-health-check-item + .lead-health-check-item {
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .lead-health-check-item.warning {
        color: color-mix(in srgb, var(--warningColor) 80%, var(--txtPrimaryColor));
    }

    .lead-health-check-message {
        min-width: 0;
    }

    .lead-health-check-pill {
        --labelHPadding: 7px;
        min-height: 18px;
        flex: 0 0 auto;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        color: var(--txtHintColor);
        background: var(--baseColor);
    }

    .lead-health-check-pill.warning {
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--warningColor) 88%, var(--txtPrimaryColor));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    @media (min-width: 1480px) {
        .leads-inbox-list {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 1060px) {
        .leads-toolbar-row {
            grid-template-columns: 1fr;
        }

        .leads-filter-cell--search,
        .leads-filter-cell--select,
        .leads-filter-cell--sort {
            min-width: 0;
        }

        .leads-filter-cell--actions {
            grid-column: auto;
            width: 100%;
        }

        .leads-inbox-layout {
            grid-template-columns: 1fr;
        }

        .leads-bulk-popover {
            left: 10px;
            right: 10px;
            bottom: 12px;
            transform: none;
            border-radius: var(--baseRadius);
            flex-wrap: wrap;
            justify-content: center;
        }
    }

    @media (max-width: 760px) {
        .leads-head.operations-head .head-selector,
        .leads-head.operations-head .head-selector .input {
            min-width: 0;
            width: 100%;
        }

        .leads-head.operations-head .head-tools,
        .leads-head.operations-head .summary-badges {
            justify-content: flex-start;
        }

        .leads-filter-cell--actions {
            width: 100%;
            justify-content: flex-start;
        }
    }
</style>

