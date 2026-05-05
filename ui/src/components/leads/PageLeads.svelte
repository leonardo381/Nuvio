<script>
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { onMount } from "svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addErrorToast, addSuccessToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { normalizeWebsiteSettingsValue } from "@/utils/WebsiteSettingsSchema";

    // NUVIO CUSTOM START: Unified Leads operations page for purpose-built backoffice workflows.
    $pageTitle = "Leads";

    const sourceFilterOptions = [
        { key: "all", label: "All" },
        { key: "contact", label: "Contact form" },
        { key: "whatsapp", label: "WhatsApp" },
        { key: "booking", label: "Booking" },
    ];
    const statusFilterOptions = [
        { key: "all", label: "All" },
        { key: "new", label: "New" },
        { key: "read", label: "Read" },
    ];
    const archivedStatusAliases = ["archived", "archive"];

    const ALL_WEBSITES_KEY = "all";
    const whatsappCollectionAliases = ["whatsapp", "whatsapp_interactions", "whatsapp_clicks"];
    const desktopMasterDetailMinWidth = 1060;

    let websites = [];
    let selectedWebsiteId = ALL_WEBSITES_KEY;
    let contactsRecords = [];
    let whatsappRecords = [];

    let sourceFilter = "all";
    let statusFilter = "all";
    let searchTerm = "";
    let sortOrder = "newest";
    let selectedLeadKey = "";
    let isLeadDetailsActive = false;
    let isDesktopMasterDetail = false;
    let isUpdatingLeadStatus = false;
    let updatingLeadStatusKey = "";

    let isLoadingWebsites = false;
    let isLoadingLeads = false;
    let leadsError = "";

    let lastWebsitesCollectionId = "";
    let lastLeadsDataKey = "";

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: contactsCollection = resolveCollectionByAliases(["contacts"]);
    $: resolvedWhatsAppCollection = resolveWhatsAppCollection();
    $: whatsappCollection = resolvedWhatsAppCollection?.collection || null;
    $: resolvedWhatsAppAlias = resolvedWhatsAppCollection?.alias || "";

    $: hasAnyLeadCollections = !!contactsCollection?.id || !!whatsappCollection?.id;
    $: hasRequiredCollections = !!contactsCollection?.id && !!whatsappCollection?.id;

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = ALL_WEBSITES_KEY;
        lastWebsitesCollectionId = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: websiteOptionMap = new Map(
        websites.map((website) => [website.id, resolveWebsiteLabel(website)]),
    );
    $: hasMultipleWebsites = websites.length > 1;
    $: singleWebsiteLabel = websites.length === 1 ? resolveWebsiteLabel(websites[0]) : "";

    $: if (selectedWebsiteId !== ALL_WEBSITES_KEY && !websiteOptionMap.has(selectedWebsiteId)) {
        selectedWebsiteId = ALL_WEBSITES_KEY;
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
    $: contactsStatusSupport = resolveStatusSupport(contactsCollection);
    $: whatsappStatusSupport = resolveStatusSupport(whatsappCollection);

    $: normalizedLeads = normalizeLeadRecords({
        contacts: contactsRecords,
        whatsapp: whatsappRecords,
        websiteLabelById: websiteOptionMap,
    });

    $: filteredLeads = filterAndSortLeads(normalizedLeads);
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
    $: selectedLead = normalizedLeads.find((lead) => lead.key === selectedLeadKey) || null;
    $: if (isDesktopMasterDetail && isLeadDetailsActive) {
        isLeadDetailsActive = false;
    }
    $: selectedLeadMailto = buildMailtoLink(selectedLead?.email);
    $: selectedLeadWhatsAppPhone = normalizeString(selectedLead?.whatsappTargetPhone || selectedLead?.phone);
    $: selectedLeadWhatsAppMessage = normalizeString(selectedLead?.whatsappTargetMessage || selectedLead?.message);
    $: selectedLeadWhatsAppLink = buildWhatsAppLink(selectedLeadWhatsAppPhone, selectedLeadWhatsAppMessage);
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
    $: selectedLeadToggleToRead = selectedLeadCurrentStatusNormalized === normalizeLower(selectedLeadStatusSupport?.newValue);
    $: selectedLeadToggleActionLabel = selectedLeadToggleToRead ? "Mark as read" : "Mark as new";
    $: selectedLeadToggleActionTarget = selectedLeadToggleToRead
        ? selectedLeadStatusSupport?.readValue
        : selectedLeadStatusSupport?.newValue;
    $: canArchiveSelectedLead = !!selectedLead?.recordId
        && selectedLeadStatusSupport?.supportsArchive
        && selectedLeadCurrentStatusNormalized !== normalizeLower(selectedLeadStatusSupport?.archiveValue);
    $: selectedLeadPrimaryMessage = normalizeString(selectedLead?.whatsappTargetMessage || selectedLead?.message);
    $: selectedLeadMissingContact = !!selectedLead && !normalizeString(selectedLead?.email) && !normalizeString(selectedLead?.phone);
    $: selectedLeadMissingMessage = !!selectedLead && !selectedLeadPrimaryMessage;
    $: selectedLeadNeedsFollowUp = normalizeLower(selectedLead?.statusKey) === "new";
    $: selectedLeadHealthWarnings = buildLeadHealthWarnings({
        selectedLead,
        selectedLeadMissingContact,
        selectedLeadMissingMessage,
        selectedLeadNeedsFollowUp,
    });
    $: selectedLeadHealthSuggestions = buildLeadHealthSuggestions({
        selectedLead,
        selectedLeadMissingContact,
        selectedLeadMissingMessage,
        selectedLeadNeedsFollowUp,
    });
    $: selectedLeadHealthState = resolveLeadHealthState(selectedLeadHealthWarnings, selectedLeadHealthSuggestions);

    $: totalNewLeads = normalizedLeads.filter((lead) => lead.statusKey === "new").length;
    $: totalThisMonthLeads = normalizedLeads.filter((lead) => isCurrentMonth(lead.created)).length;
    $: totalContactFormLeads = normalizedLeads.filter((lead) => lead.sourceKey === "contact").length;
    $: totalWhatsAppLeads = normalizedLeads.filter((lead) => lead.sourceKey === "whatsapp").length;
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
        if (!Array.isArray(aliases) || !aliases.length) {
            return null;
        }

        const normalizedAliases = aliases.map((alias) => normalizeLower(alias)).filter(Boolean);

        for (const alias of normalizedAliases) {
            const match = $collections.find((collection) => normalizeLower(collection?.name) === alias);
            if (match) {
                return match;
            }
        }

        return null;
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
        return (
            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim()
            || website?.id
            || ""
        );
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

    function isCurrentMonth(value) {
        const timestamp = toTimestamp(value);
        if (!timestamp) {
            return false;
        }

        const date = new Date(timestamp);
        const now = new Date();

        return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
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

    function buildLeadHealthWarnings({
        selectedLead,
        selectedLeadMissingContact,
        selectedLeadMissingMessage,
        selectedLeadNeedsFollowUp,
    }) {
        if (!selectedLead) {
            return [];
        }

        const warnings = [];
        if (selectedLeadMissingContact) {
            warnings.push("Contact details are incomplete for this lead.");
        }
        if (selectedLeadMissingMessage) {
            warnings.push("No message was captured for this lead.");
        }
        if (selectedLeadNeedsFollowUp) {
            warnings.push("This lead is still marked as new.");
        }

        return warnings;
    }

    function buildLeadHealthSuggestions({
        selectedLead,
        selectedLeadMissingContact,
        selectedLeadMissingMessage,
        selectedLeadNeedsFollowUp,
    }) {
        if (!selectedLead) {
            return [];
        }

        const suggestions = [];
        if (!selectedLeadMissingContact) {
            suggestions.push("Use copy actions to follow up quickly.");
        }
        if (!selectedLeadMissingMessage) {
            suggestions.push("Review the message before responding.");
        }
        if (selectedLeadNeedsFollowUp) {
            suggestions.push("Mark this lead as read after handling it.");
        }

        if (!suggestions.length) {
            suggestions.push("Lead details are complete and ready for follow-up.");
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

    function filterAndSortLeads(leads) {
        const next = (leads || []).filter((lead) => {
            const bySource = sourceFilter === "all" || lead.sourceKey === sourceFilter;
            const byStatus = statusFilter === "all" || lead.statusKey === statusFilter;

            if (!bySource || !byStatus) {
                return false;
            }

            if (!normalizedSearchTerm) {
                return true;
            }

            const searchable = [
                lead.identity,
                lead.secondaryIdentity,
                lead.websiteName,
                lead.subject,
                lead.message,
                lead.whatsappTargetMessage,
                lead.preview,
                lead.page,
                lead.originSource,
            ]
                .map((value) => normalizeLower(value))
                .filter(Boolean)
                .join(" ");

            return searchable.includes(normalizedSearchTerm);
        });

        next.sort((a, b) => {
            if (sortOrder === "oldest") {
                return (a.createdTs || 0) - (b.createdTs || 0);
            }

            return (b.createdTs || 0) - (a.createdTs || 0);
        });

        return next;
    }

    function buildWebsiteFilterValue() {
        if (selectedWebsiteId === ALL_WEBSITES_KEY) {
            return "";
        }

        return `website="${selectedWebsiteId}"`;
    }

    async function loadRecordsByCollection(collection, requestKeyPrefix) {
        if (!collection?.id) {
            return [];
        }

        const filter = buildWebsiteFilterValue();
        const requestOptions = {
            sort: "-created",
            requestKey: `${requestKeyPrefix}_${selectedWebsiteId || ALL_WEBSITES_KEY}`,
            expand: "website",
        };

        if (filter) {
            requestOptions.filter = filter;
        }

        try {
            return await ApiClient.collection(collection.id).getFullList(requestOptions);
        } catch (err) {
            ApiClient.error(err, false);
            return [];
        }
    }

    async function loadWebsites() {
        if (!websitesCollection?.id) {
            websites = [];
            return;
        }

        isLoadingWebsites = true;

        try {
            websites = await ApiClient.collection(websitesCollection.id).getFullList({
                sort: resolveWebsitesSort(websitesCollection),
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

        if (!hasAnyLeadCollections) {
            contactsRecords = [];
            whatsappRecords = [];
            return;
        }

        isLoadingLeads = true;

        try {
            const [nextContacts, nextWhatsApp] = await Promise.all([
                loadRecordsByCollection(contactsCollection, "nuvio_leads_contacts"),
                loadRecordsByCollection(whatsappCollection, "nuvio_leads_whatsapp"),
            ]);

            contactsRecords = nextContacts;
            whatsappRecords = nextWhatsApp;
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

    function patchLeadStatus(sourceKey, recordId, statusFieldName, nextStatusValue) {
        if (!recordId || !statusFieldName) {
            return;
        }

        const patch = (records) =>
            (records || []).map((record) =>
                record?.id === recordId
                    ? {
                        ...record,
                        [statusFieldName]: nextStatusValue,
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
        if (!statusSupport?.collectionId || !statusSupport?.fieldName) {
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
            await ApiClient.collection(statusSupport.collectionId).update(recordId, {
                [statusSupport.fieldName]: nextStatusValue,
            });

            patchLeadStatus(sourceKey, recordId, statusSupport.fieldName, nextStatusValue);

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
        searchTerm = "";
        sortOrder = "newest";
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
                </div>
                <p class="head-description txt-sm txt-hint m-b-0">
                    Review contacts and interactions generated by the website.
                </p>
            </div>

            <div class="head-tools">
                <div class="head-context">
                    {#if hasMultipleWebsites}
                        <div class="head-context-item">
                            <label class="txt-xs txt-hint" for="leads-website-filter">Website</label>
                            <select
                                id="leads-website-filter"
                                class="input input-sm"
                                bind:value={selectedWebsiteId}
                                disabled={isLoadingWebsites}
                            >
                                <option value={ALL_WEBSITES_KEY}>All websites</option>
                                {#each websites as website (website.id)}
                                    <option value={website.id}>{resolveWebsiteLabel(website)}</option>
                                {/each}
                            </select>
                        </div>
                    {:else if singleWebsiteLabel}
                        <span class="summary-pill">
                            <i class="ri-global-line" />
                            {singleWebsiteLabel}
                        </span>
                    {/if}
                </div>

                <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={reload} />
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
        </div>
    </section>

    <section class="panel leads-body m-b-base">
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
                                placeholder="Search by name, email, phone, subject, message, or source..."
                                bind:value={searchTerm}
                            />
                        </div>

                        <div class="leads-filter-cell leads-filter-cell--sort">
                            <label class="txt-sm txt-hint" for="leads-sort-filter">Sort</label>
                            <select id="leads-sort-filter" class="input input-sm" bind:value={sortOrder}>
                                <option value="newest">Newest first</option>
                                <option value="oldest">Oldest first</option>
                            </select>
                        </div>

                        <div class="leads-filter-cell leads-filter-cell--reset">
                            <button type="button" class="btn btn-sm btn-outline" on:click={clearFilters}>
                                <span class="txt">Reset filters</span>
                            </button>
                        </div>
                    </div>

                    <div class="leads-filter-segments">
                        <div class="leads-segment-group">
                            <span class="txt-sm txt-hint">Source</span>
                            <div class="tabs-header compact combined left operations-tabs operations-tabs--nested">
                                {#each sourceFilterOptions as option (option.key)}
                                    <button
                                        type="button"
                                        class="tab-item"
                                        class:active={sourceFilter === option.key}
                                        on:click={() => (sourceFilter = option.key)}
                                    >
                                        <span class="txt">{option.label}</span>
                                    </button>
                                {/each}
                            </div>
                        </div>

                        <div class="leads-segment-group">
                            <span class="txt-sm txt-hint">Status</span>
                            <div class="tabs-header compact combined left operations-tabs operations-tabs--nested">
                                {#each statusFilterOptions as option (option.key)}
                                    <button
                                        type="button"
                                        class="tab-item"
                                        class:active={statusFilter === option.key}
                                        on:click={() => (statusFilter = option.key)}
                                    >
                                        <span class="txt">{option.label}</span>
                                    </button>
                                {/each}
                            </div>
                        </div>
                    </div>
                </div>

                {#if !hasAnyLeadCollections}
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

                {#if $isCollectionsLoading || isLoadingLeads}
                    <div class="placeholder-section m-b-0">
                        <span class="loader loader-lg" />
                        <h1>Loading leads...</h1>
                    </div>
                {:else if !normalizedLeads.length}
                    <div class="empty-state m-b-0">
                        No leads yet. Contact form submissions and WhatsApp interactions will appear here.
                    </div>
                {:else if !filteredLeads.length}
                    <div class="empty-state m-b-0">
                        No leads match these filters.
                    </div>
                {:else}
                    <div class="leads-inbox-list" role="list">
                        {#each filteredLeads as lead (lead.key)}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                            <article
                                class="leads-inbox-item"
                                class:selected={selectedLeadKey === lead.key}
                                role="button"
                                tabindex="0"
                                aria-label={`Open ${lead.sourceLabel} lead details`}
                                on:click={() => openLeadDetails(lead)}
                                on:keydown={(event) => handleLeadCardKeyDown(event, lead)}
                            >
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

                                {#if lead.secondaryIdentity}
                                    <div class="txt-sm txt-hint leads-inbox-item-secondary">{lead.secondaryIdentity}</div>
                                {/if}

                                {#if lead.email || lead.phone}
                                    <div class="leads-inbox-item-contact txt-xs txt-hint">
                                        {#if lead.email}
                                            <span><i class="ri-mail-line" /> {lead.email}</span>
                                        {/if}
                                        {#if lead.phone}
                                            <span><i class="ri-phone-line" /> {lead.phone}</span>
                                        {/if}
                                    </div>
                                {/if}

                                <p class="txt-sm m-b-0 leads-inbox-item-preview">
                                    {resolveLeadPreviewText(lead)}
                                </p>

                                <div class="leads-inbox-item-attribution txt-xs txt-hint">
                                    <span>{resolveLeadAttribution(lead)}</span>
                                    {#if resolveLeadLocationHint(lead)}
                                        <span>{resolveLeadLocationHint(lead)}</span>
                                    {/if}
                                </div>
                            </article>
                        {/each}
                    </div>
                {/if}
            </div>

            {#if isDesktopMasterDetail}
                    <aside class="leads-detail-rail" aria-live="polite">
                        {#if selectedLead}
                            <div class="lead-detail-layout">
                                <section class="lead-detail-section lead-rail-block lead-rail-block--summary">
                                    <div class="lead-detail-section-head">
                                        <h5 class="m-0">Lead summary</h5>
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

                                    <div class="lead-detail-grid">
                                        <div class="lead-detail-row">
                                            <span class="txt-xs txt-hint">Created</span>
                                            <span class="txt-sm">{formatDateTime(selectedLead.created)}</span>
                                        </div>
                                        <div class="lead-detail-row">
                                            <span class="txt-xs txt-hint">Contact</span>
                                            <span class="txt-sm">
                                                {#if selectedLead.email || selectedLead.phone}
                                                    {selectedLead.email || "No email"}{selectedLead.phone ? ` - ${selectedLead.phone}` : ""}
                                                {:else}
                                                    No contact details available.
                                                {/if}
                                            </span>
                                        </div>
                                        {#if selectedLead.subject}
                                            <div class="lead-detail-row">
                                                <span class="txt-xs txt-hint">Subject</span>
                                                <span class="txt-sm">{selectedLead.subject}</span>
                                            </div>
                                        {/if}
                                        <div class="lead-detail-row lead-detail-row-block">
                                            <span class="txt-xs txt-hint">Message preview</span>
                                            <p class="txt-sm m-b-0">{resolveLeadPreviewText(selectedLead)}</p>
                                        </div>
                                        <div class="lead-detail-row">
                                            <span class="txt-xs txt-hint">Attribution</span>
                                            <span class="txt-sm">{resolveLeadAttribution(selectedLead)}</span>
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

                                <section class="lead-detail-section lead-rail-block lead-rail-block--actions">
                                    <div class="lead-detail-section-head">
                                        <h5 class="m-0">Actions</h5>
                                    </div>

                                    <div class="lead-detail-actions-block">
                                        <div class="txt-xs txt-hint txt-uppercase">Status actions</div>
                                        <div class="lead-detail-actions">
                                            {#if canToggleSelectedLeadStatus}
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
                                                Status actions are not available for this lead source yet.
                                            </p>
                                        {:else if selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsArchive}
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
                                        </div>
                                    </div>
                                </section>

                                <section class="lead-detail-section lead-rail-block lead-rail-block--health">
                                    <div class="lead-detail-section-head">
                                        <h5 class="m-0">Lead health</h5>
                                    </div>
                                    <p class="txt-sm txt-hint m-b-0">
                                        Review what is missing before follow-up.
                                    </p>
                                    <div class="lead-health-summary">
                                        <span class={`label label-sm ${selectedLeadHealthState.badgeClass}`}>{selectedLeadHealthState.label}</span>
                                        <span class="summary-pill">
                                            Warnings: {selectedLeadHealthWarnings.length}
                                        </span>
                                        <span class="summary-pill">
                                            Suggestions: {selectedLeadHealthSuggestions.length}
                                        </span>
                                    </div>

                                    {#if selectedLeadHealthWarnings.length}
                                        <div class="lead-health-list-wrap">
                                            <div class="txt-xs txt-hint txt-uppercase">Warnings</div>
                                            <ul class="lead-health-list lead-health-list--warnings">
                                                {#each selectedLeadHealthWarnings as warning}
                                                    <li>{warning}</li>
                                                {/each}
                                            </ul>
                                        </div>
                                    {/if}

                                    {#if selectedLeadHealthSuggestions.length}
                                        <div class="lead-health-list-wrap">
                                            <div class="txt-xs txt-hint txt-uppercase">Suggestions</div>
                                            <ul class="lead-health-list">
                                                {#each selectedLeadHealthSuggestions as suggestion}
                                                    <li>{suggestion}</li>
                                                {/each}
                                            </ul>
                                        </div>
                                    {/if}
                                </section>

                                <section class="lead-detail-section lead-rail-block lead-rail-block--notification">
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
                            </div>
                        {:else}
                            <div class="empty-state m-b-0 leads-detail-empty-state">
                                Select a lead to view details.
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
                                {#if canToggleSelectedLeadStatus}
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
                                    Status actions are not available for this lead source yet.
                                </p>
                            {:else if selectedLead.sourceKey !== "whatsapp" && !selectedLeadStatusSupport.supportsArchive}
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
                            </div>
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
        max-width: 640px;
    }

    .leads-body {
        padding: calc(var(--baseSpacing) - 10px);
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .head-tools {
        display: flex;
        align-items: flex-end;
        gap: 10px;
        flex-wrap: wrap;
    }

    .head-context {
        display: inline-flex;
        align-items: flex-end;
        gap: 8px;
        flex-wrap: wrap;
    }

    .head-context-item {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 200px;
    }

    .head-context-item .input {
        min-width: 200px;
    }

    .leads-toolbar-row {
        display: grid;
        grid-template-columns: minmax(300px, 1.8fr) minmax(170px, 0.7fr) auto;
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

    .leads-filter-cell--reset {
        align-items: flex-end;
        justify-content: flex-end;
    }

    .leads-filter-cell--reset .btn {
        white-space: nowrap;
    }

    .leads-filter-segments {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 10px;
    }

    .leads-segment-group {
        display: flex;
        flex-direction: column;
        gap: 6px;
        min-width: 0;
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

    .leads-inbox-item {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px 12px;
        display: flex;
        flex-direction: column;
        gap: 6px;
        cursor: pointer;
        transition: border-color var(--baseAnimationSpeed), box-shadow var(--baseAnimationSpeed), background-color var(--baseAnimationSpeed);
    }

    .leads-inbox-item:hover,
    .leads-inbox-item:focus-visible {
        border-color: var(--baseAlt3Color);
        box-shadow: 0 0 0 2px var(--baseAltColor);
        outline: none;
    }

    .leads-inbox-item.selected {
        border-color: color-mix(in srgb, var(--primaryColor) 40%, var(--baseAlt3Color));
        background: color-mix(in srgb, var(--primaryColor) 6%, var(--baseColor));
        box-shadow: 0 0 0 2px color-mix(in srgb, var(--primaryColor) 20%, transparent);
    }

    .leads-inbox-item-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
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
        gap: 10px;
    }

    .leads-inbox-item-contact span {
        display: inline-flex;
        align-items: center;
        gap: 4px;
    }

    .leads-inbox-item-preview {
        color: var(--txtPrimaryColor);
    }

    .leads-inbox-item-attribution {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        padding-top: 2px;
    }

    .leads-inbox-item-attribution span {
        color: var(--txtHintColor);
    }

    .leads-inbox-item-attribution span:first-child {
        color: var(--txtPrimaryColor);
        font-weight: 500;
    }

    .leads-inbox-item-attribution span + span::before {
        content: "-";
        margin-right: 8px;
        color: var(--txtDisabledColor);
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
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 10px;
        background: color-mix(in srgb, var(--baseColor) 90%, var(--baseAlt1Color));
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

    .lead-detail-section-head {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
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

    .lead-detail-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        padding-top: 2px;
    }

    .lead-notification-badges {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .lead-health-summary {
        display: inline-flex;
        flex-wrap: wrap;
        gap: 6px;
        align-items: center;
    }

    .lead-health-list-wrap {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .lead-health-list {
        margin: 0;
        padding-left: 16px;
        color: var(--txtHintColor);
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .lead-health-list--warnings {
        color: var(--dangerColor);
    }

    @media (max-width: 1240px) {
        .leads-toolbar-row {
            grid-template-columns: minmax(240px, 1.5fr) minmax(150px, 0.8fr) auto;
        }
    }

    @media (min-width: 1480px) {
        .leads-inbox-list {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 1060px) {
        .leads-toolbar-row {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }

        .leads-filter-cell--reset {
            align-items: flex-start;
        }

        .leads-filter-segments {
            grid-template-columns: 1fr;
        }

        .leads-inbox-layout {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 760px) {
        .head-context-item,
        .head-context-item .input {
            min-width: 0;
            width: 100%;
        }

        .leads-toolbar-row {
            grid-template-columns: 1fr;
        }
    }
</style>
