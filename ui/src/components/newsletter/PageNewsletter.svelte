<script>    import { querystring } from "svelte-spa-router";    import ApiClient from "@/utils/ApiClient";    import CommonHelper from "@/utils/CommonHelper";    import PageWrapper from "@/components/base/PageWrapper.svelte";    import RefreshButton from "@/components/base/RefreshButton.svelte";    import OverlayPanel from "@/components/base/OverlayPanel.svelte";    import TinyMCE from "@/components/base/TinyMCE.svelte";    import { pageTitle } from "@/stores/app";    import { collections, collectionsLoadError, findCollectionByRequiredNames, hasCollectionsLoaded, isCollectionsLoading } from "@/stores/collections";    import { addErrorToast, addSuccessToast, addWarningToast } from "@/stores/toasts";    // NUVIO CUSTOM START: Newsletter V1 dedicated section/page (collection-backed).
    $pageTitle = "Newsletter";    const initialQueryParams = new URLSearchParams($querystring);    const subscriberStatuses = ["active", "pending", "unsubscribed"];    const subscriberLeadSource = "manual_dashboard";    const subscriberSortOptions = [        { value: "newest", label: "Newest" },        { value: "oldest", label: "Oldest" },        { value: "emailAsc", label: "Email A-Z" },        { value: "emailDesc", label: "Email Z-A" },        { value: "status", label: "Status" },    ];    const subscribersPageSize = 20;    const audienceInitialVisibleCount = 30;    const audienceVisibleIncrement = 30;    const newsletterSections = new Set(["subscribers", "campaigns"]);    const subscriberGroupsFieldAliases = ["groups", "groupIds", "subscriberGroups", "subscriber_groups"];    const campaignRecipientsTypeFieldAliases = ["recipientsType", "recipientType", "recipients_type"];    const campaignRecipientsIdsFieldAliases = ["recipientsIds", "recipientIds", "recipients_ids"];    let activeSection = newsletterSections.has(initialQueryParams.get("newsletterTab"))        ? initialQueryParams.get("newsletterTab")        : "subscribers";    let websites = [];    let selectedWebsiteId = initialQueryParams.get("newsletterWebsite") || "";    let subscribers = [];
    let campaigns = [];
    let subscriberGroups = [];
    let isLoadingWebsites = false;    let isLoadingSubscribers = false;
    let isLoadingCampaigns = false;
    let isLoadingSubscriberGroups = false;
    let isCreatingSubscriber = false;
    let isCreatingSubscriberGroup = false;
    let isCreatingCampaign = false;
    let isBulkUpdating = false;
    let isSavingSubscriber = false;
    let isSendingCampaign = {};
    let deletingSubscriberId = "";
    let isSubscriberCreateOpen = false;
    let subscriberForm = {
        name: "",
        email: "",
        status: "pending",
        groupIds: [],
    };
    let subscriberFormError = "";
    let subscriberGroupForm = {
        name: "",
    };
    let subscriberGroupFormError = "";
    let campaignForm = {
        subject: "",
        body: "",
        recipientsType: "manual",
        recipientsIds: [],
    };
    let campaignFormError = "";
    let campaignWorkspace = "builder";
    let campaignBuilderShowEditor = true;
    let campaignBuilderShowPreview = false;
    let campaignStatusFilter = "all";
    let isCampaignPreviewExpanded = false;
    const campaignPreviewModes = [
        { key: "desktop", label: "Desktop", icon: "ri-computer-line", width: 720 },
        { key: "tablet", label: "Tablet", icon: "ri-tablet-line", width: 600 },
        { key: "mobile", label: "Mobile", icon: "ri-smartphone-line", width: 390 },
    ];
    let campaignPreviewMode = "desktop";
    let editingCampaignId = "";
    let viewingSentCampaignId = "";
    let isSavingCampaign = false;
    let isDuplicatingCampaign = {};
    let audienceRecipientSearch = "";
    let audienceRecipientVisibilityFilter = "all";
    let visibleAudienceLimit = audienceInitialVisibleCount;
    let subscriberSearch = "";
    let subscriberStatusFilter = "all";
    let subscriberGroupFilter = "all";
    let subscriberSort = "newest";
    let subscribersPage = 1;
    let selectedSubscriberIds = [];

    let pendingSendCampaign = null;
    let pendingDeleteCampaign = null;
    let pendingDeleteSubscriber = null;
    let editingSubscriberId = "";
    let deletingCampaignId = "";
    let editingSubscriberForm = {
        name: "",
        email: "",
        status: "pending",
        groupIds: [],
    };
    let editingSubscriberError = "";
    let subscriberEmailInput;
    let subscriberGroupCountById = {};
    let activeSubscriberIds = [];
    let activeSubscriberIdsSet = new Set();
    let normalizedCampaignManualRecipientIds = [];
    let normalizedCampaignManualRecipientIdsSet = new Set();
    let normalizedCampaignManualRecipientsCount = 0;
    let activeSubscriberIdsByGroupId = new Map();
    let groupSelectionMetaById = new Map();
    let lastWebsitesCollectionId = "";    let lastDataKey = "";    let lastSubscribersFilterKey = "";    let lastAudienceRecipientsResetKey = "";    let lastPersistedContextKey = "";    $: websitesCollection = findCollectionByRequiredNames($collections, ["websites", "Websites"]);    $: subscribersCollection = findCollectionByRequiredNames($collections, ["Subscribers", "subscribers"]); 
    $: campaignsCollection = findCollectionByRequiredNames($collections, ["Campaigns", "campaigns"]);
    $: subscriberGroupsCollection = findCollectionByRequiredNames($collections, ["SubscriberGroups", "subscribergroups"]);
    $: missingCollectionNames = [];
    $: if (!subscribersCollection) {
        missingCollectionNames.push("Subscribers");
    }
    $: if (!campaignsCollection) {
        missingCollectionNames.push("Campaigns");
    }
    $: hasNewsletterCollections = missingCollectionNames.length === 0;
    $: subscriberFieldKeys = new Set(
        Array.isArray(subscribersCollection?.fields)
            ? subscribersCollection.fields.map((field) => `${field?.name || ""}`.trim().toLowerCase()).filter(Boolean)
            : [],
    );
    $: subscriberGroupsFieldName = resolveCollectionFieldNameForMultiSelectAlias(subscribersCollection, subscriberGroupsFieldAliases);
    $: campaignRecipientsTypeFieldName = resolveCollectionFieldNameByAliasPriority(campaignsCollection, campaignRecipientsTypeFieldAliases) || "recipientsType";
    $: campaignRecipientsIdsFieldName = resolveCollectionFieldNameByAliasPriority(campaignsCollection, campaignRecipientsIdsFieldAliases) || "recipientsIds";
    $: subscribersSupportsGroupsField = !!subscriberGroupsFieldName;
    $: subscribersSupportsNameField = subscriberFieldKeys.has("name");
    $: subscribersSupportsSourceField = subscriberFieldKeys.has("source");
    $: hasSubscriberGroupsFeature = !!subscriberGroupsCollection?.id && subscribersSupportsGroupsField;
    $: if (!websitesCollection?.id) {        websites = [];        selectedWebsiteId = "";        lastWebsitesCollectionId = "";    } else if (websitesCollection.id !== lastWebsitesCollectionId) {        lastWebsitesCollectionId = websitesCollection.id;        loadWebsites();    };    $: websiteDataKey = `${selectedWebsiteId}:${subscribersCollection?.id || ""}:${campaignsCollection?.id || ""}:${subscriberGroupsCollection?.id || ""}`;
    $: if (selectedWebsiteId && hasNewsletterCollections && websiteDataKey !== lastDataKey) {
        lastDataKey = websiteDataKey;
        loadSubscribers();
        loadCampaigns();
        loadSubscriberGroups();
    }

    $: if (!selectedWebsiteId || !hasSubscriberGroupsFeature) {
        subscriberGroups = [];
    }
    $: activeSubscribers = subscribers.filter((subscriber) => normalizeStatus(subscriber?.status) === "active");    $: sentCampaigns = campaigns.filter((campaign) => normalizeStatus(campaign?.status) === "sent");    $: draftCampaigns = campaigns.filter((campaign) => normalizeStatus(campaign?.status) === "draft");    $: newSubscribersLast7Days = subscribers.filter((subscriber) => isWithinLastDays(subscriber?.created, 7)).length;    $: unsubscribedLast7Days = subscribers.filter((subscriber) => {
        return normalizeStatus(subscriber?.status) === "unsubscribed"
            && isWithinLastDays(subscriber?.updated || subscriber?.created, 7);
    }).length;
    $: if (!isLoadingSubscribers && !subscribers.length) {
        isSubscriberCreateOpen = true;
    }
    $: subscriberGroupsById = new Map(subscriberGroups.map((group) => [group.id, group]));
    $: subscriberGroupCountById = subscriberGroups.reduce((acc, group) => {
        acc[group.id] = 0;
        return acc;
    }, {});
    $: subscribers.forEach((subscriber) => {
        getSubscriberGroupIds(subscriber).forEach((groupId) => {
            if (subscriberGroupCountById[groupId] !== undefined) {
                subscriberGroupCountById[groupId] += 1;
            }
        });
    });
    $: normalizedSubscriberFormEmail = normalizeEmail(subscriberForm.email);    $: subscriberAlreadyExists = !!normalizedSubscriberFormEmail
        && subscribers.some((subscriber) => normalizeEmail(subscriber?.email) === normalizedSubscriberFormEmail);
    $: createSubscriberDisabledReason = resolveCreateSubscriberDisabledReason(
        isCreatingSubscriber,
        selectedWebsiteId,
        subscriberForm.email,
    );
    $: normalizedSubscriberGroupName = normalizeGroupName(subscriberGroupForm.name).toLowerCase();
    $: subscriberGroupNameAlreadyExists = !!normalizedSubscriberGroupName
        && subscriberGroups.some((group) => normalizeGroupName(group?.name).toLowerCase() === normalizedSubscriberGroupName);
    $: createSubscriberGroupDisabledReason = resolveCreateSubscriberGroupDisabledReason(
        isCreatingSubscriberGroup,
        selectedWebsiteId,
        subscriberGroupForm.name,
    );
    $: createCampaignDisabledReason = (editingCampaignId
        && viewingSentCampaignId === editingCampaignId
        && normalizeStatus(editingCampaign?.status) === "sent")
        ? "Sent campaigns are kept as history. Duplicate to edit and send again."
        : resolveCreateCampaignDisabledReason(
            isCreatingCampaign,
            selectedWebsiteId,
            campaignForm.subject,
            campaignForm.body,
            campaignForm.recipientsType,
            campaignForm.recipientsIds,
        );
    $: campaignSubjectValue = `${campaignForm.subject || ""}`.trim();
    $: campaignBodyValue = `${campaignForm.body || ""}`.trim();
    $: activeCampaignPreviewMode = campaignPreviewModes.find((mode) => mode.key === campaignPreviewMode) || campaignPreviewModes[0];
    $: campaignPreviewFrameWidth = activeCampaignPreviewMode?.width || 720;
    $: campaignPreviewFrameStyle = `--campaign-preview-frame-width: ${campaignPreviewFrameWidth}px;`;
    $: campaignPreviewHasBody = !!campaignBodyValue;
    $: campaignPreviewDocumentHtml = buildCampaignPreviewDocument(campaignSubjectValue, campaignForm.body);
    $: campaignBodyEditorConfig = {        ...CommonHelper.defaultEditorOptions(),        convert_urls: false,        relative_urls: false,        min_height: 320,        height: 320,    };
    $: shouldShowCampaignSubjectValidation = !campaignSubjectValue && (!!campaignBodyValue || !!campaignFormError);
    $: shouldShowCampaignBodyValidation = !campaignBodyValue && (!!campaignSubjectValue || !!campaignFormError);
    $: filteredCampaigns = campaigns.filter((campaign) => {
        if (campaignStatusFilter === "all") {
            return true;
        }
        return normalizeStatus(campaign?.status) === campaignStatusFilter;
    });
    $: editingCampaign = campaigns.find((campaign) => campaign.id === editingCampaignId) || null;
    $: isViewingSentCampaign = !!editingCampaign
        && viewingSentCampaignId === editingCampaign.id
        && normalizeStatus(editingCampaign?.status) === "sent";
    $: editingCampaignLabel = `${editingCampaign?.subject || ""}`.trim() || "(No subject)";
    $: if (viewingSentCampaignId && viewingSentCampaignId !== editingCampaignId) {
        viewingSentCampaignId = "";
    }
    $: normalizedSubscriberSearch = `${subscriberSearch || ""}`.trim().toLowerCase();    $: filteredSubscribers = sortSubscribers(
        subscribers.filter((subscriber) => {
            const status = normalizeStatus(subscriber?.status);
            const groupIds = getSubscriberGroupIds(subscriber);
            const byStatus = subscriberStatusFilter === "all" || status === subscriberStatusFilter;
            const byGroup = subscriberGroupFilter === "all" || groupIds.includes(subscriberGroupFilter);
            const bySearch = !normalizedSubscriberSearch
                || `${subscriber?.email || ""}`.toLowerCase().includes(normalizedSubscriberSearch)
                || `${subscriber?.name || ""}`.toLowerCase().includes(normalizedSubscriberSearch);

            return byStatus && byGroup && bySearch;
        }),
    );
    $: activeSubscriberIds = activeSubscribers.map((subscriber) => subscriber.id);
    $: activeSubscriberIdsSet = new Set(activeSubscriberIds);
    $: normalizedCampaignManualRecipientIds = normalizeManualRecipientIds(campaignForm.recipientsIds);
    $: normalizedCampaignManualRecipientIdsSet = new Set(normalizedCampaignManualRecipientIds);
    $: normalizedCampaignManualRecipientsCount = normalizedCampaignManualRecipientIds.length;
    $: activeSubscriberIdsByGroupId = new Map(
        subscriberGroups.map((group) => {
            const ids = activeSubscribers
                .filter((subscriber) => normalizeIdList(subscriber?.[subscriberGroupsFieldName]).includes(group.id))
                .map((subscriber) => subscriber.id);
            return [group.id, ids];
        }),
    );
    $: groupSelectionMetaById = new Map(
        subscriberGroups.map((group) => {
            const groupActiveRecipientIds = [...(activeSubscriberIdsByGroupId.get(group.id) || [])];
            const totalCount = groupActiveRecipientIds.length;
            let selectedCount = 0;
            groupActiveRecipientIds.forEach((id) => {
                if (normalizedCampaignManualRecipientIdsSet.has(id)) {
                    selectedCount += 1;
                }
            });
            let state = "none";
            if (totalCount > 0) {
                if (selectedCount === totalCount) {
                    state = "full";
                } else if (selectedCount > 0) {
                    state = "partial";
                }
            }
            return [group.id, { state, selectedCount, totalCount }];
        }),
    );
    $: normalizedAudienceRecipientSearch = `${audienceRecipientSearch || ""}`.trim().toLowerCase();
    $: filteredAudienceRecipients = activeSubscribers.filter((subscriber) => {
        const isSelected = normalizedCampaignManualRecipientIdsSet.has(subscriber.id);
        if (audienceRecipientVisibilityFilter === "selected" && !isSelected) {
            return false;
        }
        if (audienceRecipientVisibilityFilter === "unselected" && isSelected) {
            return false;
        }
        if (!normalizedAudienceRecipientSearch) {
            return true;
        }
        const groupsLabel = hasSubscriberGroupsFeature
            ? getSubscriberGroupIds(subscriber)
                .map((groupId) => subscriberGroupsById.get(groupId)?.name)
                .filter(Boolean)
                .join(" ")
            : "";
        const searchable = `${resolveSubscriberDisplayName(subscriber)} ${subscriber?.email || ""} ${groupsLabel}`.toLowerCase();
        return searchable.includes(normalizedAudienceRecipientSearch);
    });
    $: audienceRecipientsVisibilityContext = audienceRecipientVisibilityFilter === "all"
        ? "all"
        : `${normalizedCampaignManualRecipientsCount}`;
    $: audienceRecipientsResetKey = [
        selectedWebsiteId,
        activeSection,
        campaignWorkspace,
        editingCampaignId,
        viewingSentCampaignId,
        normalizedAudienceRecipientSearch,
        audienceRecipientVisibilityFilter,
        audienceRecipientsVisibilityContext,
    ].join("|");
    $: if (audienceRecipientsResetKey !== lastAudienceRecipientsResetKey) {
        lastAudienceRecipientsResetKey = audienceRecipientsResetKey;
        visibleAudienceLimit = audienceInitialVisibleCount;
    }
    $: visibleAudienceRecipients = filteredAudienceRecipients.slice(0, visibleAudienceLimit);
    $: visibleAudienceRecipientsCount = visibleAudienceRecipients.length;
    $: canLoadMoreAudienceRecipients = visibleAudienceRecipientsCount < filteredAudienceRecipients.length;
    $: audienceRecipientsSummary = `${resolveManualAudienceRecipientCountForMode(campaignForm.recipientsType, campaignForm.recipientsIds)} selected / ${activeSubscribers.length} active`;
    $: audienceSummaryStatus = isViewingSentCampaign
        ? "Sent campaign (read-only)"
        : (editingCampaignId ? "Editing existing draft" : "New draft");
    $: audienceSubjectReady = !!campaignSubjectValue;
    $: audienceBodyReady = !!campaignBodyValue;
    $: audienceRecipientsReady = normalizedCampaignManualRecipientsCount > 0;
    $: audienceHasUnsavedChanges = !!editingCampaign && campaignHasUnsavedComposerChanges(editingCampaign);
    $: audienceSendDisabledReason = !editingCampaign
        ? "Save the draft before sending this campaign."
        : getSendCampaignDisabledReason(editingCampaign);
    $: audienceBlockingWarnings = [];
    $: if (!audienceSubjectReady) {
        audienceBlockingWarnings.push("Add a subject before sending.");
    }
    $: if (!audienceBodyReady) {
        audienceBlockingWarnings.push("Add email content before sending.");
    }
    $: if (!audienceRecipientsReady) {
        audienceBlockingWarnings.push("Select at least one active recipient before sending.");
    }
    $: if (!editingCampaignId) {
        audienceBlockingWarnings.push("Save the draft before sending this campaign.");
    }
    $: if (editingCampaign && normalizeStatus(editingCampaign.status) === "sent") {
        audienceBlockingWarnings.push("Campaign already sent.");
    }
    $: audienceAttentionWarnings = [];
    $: if (audienceHasUnsavedChanges) {
        audienceAttentionWarnings.push("Update the draft before sending these selected recipients.");
    }
    $: if (
        !!editingCampaign
        && !!audienceSendDisabledReason
        && audienceSendDisabledReason !== "Update the draft before sending these selected recipients."
        && !audienceBlockingWarnings.includes(audienceSendDisabledReason)
    ) {
        audienceAttentionWarnings.push(audienceSendDisabledReason);
    }
    $: audienceWarnings = [...audienceBlockingWarnings, ...audienceAttentionWarnings];
    $: audienceSuggestions = [];
    $: if (hasSubscriberGroupsFeature) {
        audienceSuggestions.push("Use groups to select recipients faster.");
    }
    $: audienceSuggestions.push("Review your audience before sending.");
    $: audienceHasMissingBasics = audienceBlockingWarnings.length > 0;
    $: audienceNeedsAttention = !audienceHasMissingBasics && audienceAttentionWarnings.length > 0;
    $: audienceCanSendNow = !!editingCampaign && !audienceSendDisabledReason;
    $: audienceHealthStatus = resolveAudienceHealthStatus({
        hasMissingBasics: audienceHasMissingBasics,
        needsAttention: audienceNeedsAttention,
        canSend: audienceCanSendNow,
    });
    $: audienceHealthPillClass = resolveAudienceHealthPillClass(audienceHealthStatus);
    $: subscribersTotalPages = Math.max(1, Math.ceil(filteredSubscribers.length / subscribersPageSize));    $: if (subscribersPage > subscribersTotalPages) {        subscribersPage = subscribersTotalPages;    };    $: subscribersPageStart = (subscribersPage - 1) * subscribersPageSize;    $: pagedSubscribers = filteredSubscribers.slice(subscribersPageStart, subscribersPageStart + subscribersPageSize);    $: visibleSubscriberIds = pagedSubscribers.map((subscriber) => subscriber.id);    $: areAllVisibleSubscribersSelected = visibleSubscriberIds.length > 0        && visibleSubscriberIds.every((id) => selectedSubscriberIds.includes(id));    $: selectedSubscribersCount = selectedSubscriberIds.length;    $: pendingSendRecipientsCount = resolveCampaignRecipientsCount(pendingSendCampaign);    $: {
        const nextFilterKey = `${selectedWebsiteId}|${subscriberSearch}|${subscriberStatusFilter}|${subscriberGroupFilter}|${subscriberSort}|${subscribers.length}`;
        if (nextFilterKey !== lastSubscribersFilterKey) {
            lastSubscribersFilterKey = nextFilterKey;
            subscribersPage = 1;
        }
    }

    $: {
        const validGroupIds = new Set(subscriberGroups.map((group) => group.id));
        const currentGroupIds = Array.isArray(subscriberForm.groupIds) ? subscriberForm.groupIds : [];
        const nextGroupIds = currentGroupIds.filter((groupId) => validGroupIds.has(groupId));
        if (nextGroupIds.length !== currentGroupIds.length) {
            subscriberForm = { ...subscriberForm, groupIds: nextGroupIds };
        }
    }
    $: {
        const nextSelected = selectedSubscriberIds.filter((id) => filteredSubscribers.some((subscriber) => subscriber.id === id));
        if (nextSelected.length !== selectedSubscriberIds.length) {
            selectedSubscriberIds = nextSelected;
        }
    }

    $: if (editingSubscriberId && !subscribers.some((subscriber) => subscriber.id === editingSubscriberId)) {
        cancelEditSubscriber();
    }
    // Keep context in URL query so refresh/navigation preserves website and active tab.
    $: if (hasNewsletterCollections) {        const nextContextKey = `${selectedWebsiteId || ""}|${activeSection || "subscribers"}`;        if (nextContextKey !== lastPersistedContextKey) {            lastPersistedContextKey = nextContextKey;            CommonHelper.replaceHashQueryParams({                newsletterWebsite: selectedWebsiteId || null,                newsletterTab: activeSection !== "subscribers" ? activeSection : null,            });        }    };    function resolveWebsitesSort(collection) {        const preferredSortFields = ["title", "name", "slug"];        const availableFields = new Set(            CommonHelper.getAllCollectionIdentifiers(collection).map((field) => `${field || ""}`.trim().toLowerCase()),        );        const validSortFields = preferredSortFields.filter((field) => availableFields.has(field));        if (!validSortFields.length) {            return "+id";        }        return validSortFields.map((field) => `+${field}`).join(",");    }    function resolveWebsiteLabel(website) {        return CommonHelper.websiteDisplayLabel(website, { missingValue: "" });    }    function normalizeEmail(email) {        return `${email || ""}`.trim().toLowerCase();    }    function resolveCollectionFieldName(collection, aliases = []) {        if (!collection || !Array.isArray(collection.fields)) {            return "";        }        const normalizedAliases = aliases.map((alias) => `${alias || ""}`.trim().toLowerCase()).filter(Boolean);        for (const field of collection.fields) {            const fieldName = `${field?.name || ""}`.trim();            if (!fieldName) {                continue;            }            if (normalizedAliases.includes(fieldName.toLowerCase())) {                return fieldName;            }        }        return "";    }    function normalizeIdList(value) {        if (Array.isArray(value)) {            return [...new Set(                value                    .map((item) => `${item || ""}`.trim())                    .filter(Boolean),            )];        }        if (typeof value === "string") {            const trimmed = value.trim();            if (!trimmed) {                return [];            }            if (trimmed.startsWith("[")) {                try {                    const parsed = JSON.parse(trimmed);                    return normalizeIdList(parsed);                } catch (err) {                    return [];                }            }            return [trimmed];        }        return [];    }    function getCampaignRecipientsType(campaign) {        const rawType = campaign?.recipientsType            ?? campaign?.[campaignRecipientsTypeFieldName]            ?? campaign?.recipientType            ?? campaign?.recipients_type            ?? "all";        return normalizeStatus(rawType || "all");    }    function getCampaignRecipientIds(campaign) {        const rawIds = campaign?.recipientsIds            ?? campaign?.[campaignRecipientsIdsFieldName]            ?? campaign?.recipientIds            ?? campaign?.recipients_ids;        return normalizeIdList(rawIds);    }    function normalizeSubscriberName(name) {
        return `${name || ""}`.trim().replace(/\s+/g, " ");
    }    function resolveSubscriberDisplayName(subscriber) {
        return normalizeSubscriberName(subscriber?.name);
    }    function normalizeStatus(status) {
        return `${status || ""}`.trim().toLowerCase();
    }

    function getSubscriberStatusLabel(status) {
        const normalized = normalizeStatus(status);
        if (normalized === "pending") {
            return "Pending confirmation";
        }
        if (normalized === "active") {
            return "Active";
        }
        if (normalized === "unsubscribed") {
            return "Unsubscribed";
        }
        return `${status || ""}`.trim() || "Unknown";
    }

    function getSubscriberStatusLabelClass(status) {
        const normalized = normalizeStatus(status);
        if (normalized === "active") {
            return "label-success";
        }
        if (normalized === "pending") {
            return "label-warning";
        }
        if (normalized === "unsubscribed") {
            return "label-danger";
        }
        return "label-info";
    }

    function resolveSubscriberSourceLabel(source) {
        const normalized = `${source || ""}`.trim();
        if (!normalized) {
            return "";
        }

        return CommonHelper.sentenize(normalized.replace(/_/g, " "), false);
    }
    function normalizeGroupName(value) {
        return `${value || ""}`.trim().replace(/\s+/g, " ");
    }

    function slugifyGroupName(value) {
        return normalizeGroupName(value)
            .toLowerCase()
            .replace(/[^a-z0-9]+/g, "-")
            .replace(/^-+|-+$/g, "")
            .slice(0, 80);
    }

    function isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalizeEmail(email));
    }

    function getSubscriberGroupIds(subscriber) {
        if (!subscriberGroupsFieldName) {
            return [];
        }
        return normalizeIdList(subscriber?.[subscriberGroupsFieldName]);
    }

    function hasSubscriberGroup(subscriber, groupId) {
        return getSubscriberGroupIds(subscriber).includes(groupId);
    }

    function getActiveSubscriberIds() {
        return activeSubscribers.map((subscriber) => subscriber.id);
    }

    function setCampaignRecipientsType(nextType) {
        void nextType;
        campaignForm = { ...campaignForm, recipientsType: "manual" };
    }

    function normalizeManualRecipientIds(ids) {
        const activeIds = getActiveSubscriberIds();
        if (!activeIds.length) {
            return [];
        }

        const selectedIdsSet = new Set(normalizeIdList(ids));
        return activeIds.filter((id) => selectedIdsSet.has(id));
    }

    function areIdListsEqual(left = [], right = []) {
        if (left.length !== right.length) {
            return false;
        }
        return left.every((value, index) => value === right[index]);
    }

    function setManualRecipientIds(ids, options = {}) {
        const { clearError = true } = options;
        if (clearError) {
            clearCampaignFormError();
        }
        const normalizedIds = normalizeManualRecipientIds(ids);
        campaignForm = { ...campaignForm, recipientsType: "manual", recipientsIds: normalizedIds };
        return normalizedIds;
    }

    function getGroupActiveRecipientIds(groupId) {
        if (!groupId) {
            return [];
        }
        return [...(activeSubscriberIdsByGroupId?.get(groupId) || [])];
    }

    function getGroupSelectionMeta(groupId) {
        if (!groupId) {
            return { state: "none", selectedCount: 0, totalCount: 0 };
        }
        return groupSelectionMetaById?.get(groupId) || { state: "none", selectedCount: 0, totalCount: 0 };
    }

    function toggleGroupRecipients(groupId) {
        const groupActiveRecipientIds = getGroupActiveRecipientIds(groupId);
        if (!groupActiveRecipientIds.length) {
            return;
        }

        if (normalizeStatus(campaignForm.recipientsType) !== "manual") {
            setCampaignRecipientsType("manual");
        }

        const nextSelection = new Set(normalizedCampaignManualRecipientIds);
        const groupState = getGroupSelectionMeta(groupId).state;
        if (groupState === "full") {
            groupActiveRecipientIds.forEach((id) => nextSelection.delete(id));
        } else {
            groupActiveRecipientIds.forEach((id) => nextSelection.add(id));
        }
        setManualRecipientIds([...nextSelection]);
    }

    function selectAllActiveRecipients() {
        if (normalizeStatus(campaignForm.recipientsType) !== "manual") {
            setCampaignRecipientsType("manual");
        }
        setManualRecipientIds(getActiveSubscriberIds());
    }

    function clearManualRecipients() {
        if (normalizeStatus(campaignForm.recipientsType) === "all") {
            setCampaignRecipientsType("manual");
        }
        setManualRecipientIds([]);
    }

    function loadMoreAudienceRecipients() {
        visibleAudienceLimit += audienceVisibleIncrement;
    }

    function normalizeManualAudienceRecipientIds(ids) {
        return normalizeManualRecipientIds(ids);
    }

    function resolveManualAudienceRecipientCountForMode(recipientsType, recipientsIds) {
        void recipientsType;
        return normalizeManualAudienceRecipientIds(recipientsIds).length;
    }

    function toTimestamp(value) {        const raw = `${value || ""}`.trim();        if (!raw) {            return 0;        }        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");        const parsed = new Date(normalized).getTime();        return Number.isNaN(parsed) ? 0 : parsed;    }    function isWithinLastDays(value, days) {        const ts = toTimestamp(value);        if (!ts) {            return false;        }        const diff = Date.now() - ts;        return diff >= 0 && diff <= (days * 24 * 60 * 60 * 1000);    }    function formatDateTime(value) {        const raw = `${value || ""}`.trim();        if (!raw) {            return "-";        }        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");        const parsed = new Date(normalized);        if (Number.isNaN(parsed.getTime())) {            return raw;        }        return parsed.toLocaleString();    }    function sortSubscribers(list) {        const sorted = [...list];        sorted.sort((a, b) => {            if (subscriberSort === "oldest") {                return toTimestamp(a?.created) - toTimestamp(b?.created);            }            if (subscriberSort === "emailAsc") {                return `${a?.email || ""}`.localeCompare(`${b?.email || ""}`);            }            if (subscriberSort === "emailDesc") {                return `${b?.email || ""}`.localeCompare(`${a?.email || ""}`);            }            if (subscriberSort === "status") {                const statusCompare = normalizeStatus(a?.status).localeCompare(normalizeStatus(b?.status));                if (statusCompare !== 0) {                    return statusCompare;                }            }            // newest/default
            return toTimestamp(b?.created) - toTimestamp(a?.created);        });        return sorted;    }    function resolveCreateSubscriberDisabledReason(isCreating, websiteId, email) {
        if (isCreating) {
            return "Adding subscriber...";
        }

        if (!websiteId) {
            return "Select a website first.";
        }

        if (!`${email || ""}`.length) {
            return "Enter subscriber email.";
        }

        return "";
    }

    function resolveCreateSubscriberGroupDisabledReason(isCreating, websiteId, name) {
        if (isCreating) {
            return "Creating group...";
        }

        if (!hasSubscriberGroupsFeature) {
            return "Groups feature is not available yet.";
        }

        if (!websiteId) {
            return "Select a website first.";
        }

        if (!normalizeGroupName(name)) {
            return "Enter a group name.";
        }

        return "";
    }

    function resolveCreateCampaignDisabledReason(isCreating, websiteId, subject, body, recipientsType, recipientsIds) {
        if (isCreating) {
            return "Creating draft...";
        }

        if (!websiteId) {
            return "Select a website first.";
        }

        if (!`${subject || ""}`.trim()) {
            return "Campaign subject is required.";
        }

        if (!`${body || ""}`.trim()) {
            return "Campaign body is required.";
        }

        if (!resolveManualAudienceRecipientCountForMode(recipientsType, recipientsIds)) {
            return "Select at least one recipient.";
        }

        return "";
    }
    function resolveCampaignRecipientsCount(campaign) {
        if (!campaign) {
            return 0;
        }
        const recipientsType = getCampaignRecipientsType(campaign);
        if (recipientsType === "manual") {
            return normalizeManualAudienceRecipientIds(getCampaignRecipientIds(campaign)).length;
        }
        return activeSubscribers.length;
    }
    function campaignHasUnsavedComposerChanges(campaign) {
        if (!campaign?.id || campaign.id !== editingCampaignId) {
            return false;
        }
        if (`${campaignForm.subject || ""}`.trim() !== `${campaign?.subject || ""}`.trim()) {
            return true;
        }
        if (`${campaignForm.body || ""}`.trim() !== `${campaign?.body || ""}`.trim()) {
            return true;
        }
        const composerRecipients = normalizeManualAudienceRecipientIds(campaignForm.recipientsIds);
        const persistedRecipients = getCampaignRecipientsType(campaign) === "all"
            ? normalizeManualAudienceRecipientIds(getActiveSubscriberIds())
            : normalizeManualAudienceRecipientIds(getCampaignRecipientIds(campaign));
        if (composerRecipients.length !== persistedRecipients.length) {
            return true;
        }
        const persistedSet = new Set(persistedRecipients);
        return composerRecipients.some((id) => !persistedSet.has(id));
    }
    function getSendCampaignDisabledReason(campaign) {
        if (!campaign?.id) {
            return "Invalid campaign.";
        }
        if (isSendingCampaign[campaign.id]) {
            return "Sending campaign...";
        }
        if (normalizeStatus(campaign.status) === "sent") {
            return "Campaign already sent.";
        }
        if (campaignHasUnsavedComposerChanges(campaign)) {
            return "Update the draft before sending these selected recipients.";
        }
        if (resolveCampaignRecipientsCount(campaign) < 1) {
            return "Select at least one active recipient before sending.";
        }
        return "";
    }
    function resolveCampaignStatusLabelClass(status) {
        const normalized = normalizeStatus(status);
        if (normalized === "sent") {
            return "label-success";
        }
        if (normalized === "draft") {
            return "label-warning";
        }
        return "label-info";
    }

    function resolveCampaignStatusLabel(status) {
        const normalized = normalizeStatus(status);
        if (normalized === "sent") {
            return "Sent";
        }
        if (normalized === "draft") {
            return "Draft";
        }
        return CommonHelper.sentenize(`${status || ""}`, false) || "Unknown";
    }

    function resolveCampaignAudienceLabel(campaign) {
        void campaign;
        return "Selected recipients";
    }

    function resolveCampaignTimelineLabel(campaign) {
        const sent = normalizeStatus(campaign?.status) === "sent";
        const timestamp = sent
            ? (campaign?.sentAt || campaign?.updated || campaign?.created)
            : (campaign?.updated || campaign?.created);
        return `${sent ? "Sent" : "Updated"}: ${formatDateTime(timestamp)}`;
    }

    function resolveCampaignRecipientsSummary(campaign) {
        return `${resolveCampaignRecipientsCount(campaign)} est.`;
    }

    function resolveCampaignDeliveredSummary(campaign) {
        return `${campaign?.recipientsCount || 0} sent`;
    }

    function resolveCampaignAudienceSummary(campaign) {
        return resolveCampaignAudienceLabel(campaign);
    }

    function resolveCampaignSendPreviewLabel(campaign) {
        return resolveCampaignRecipientsSummary(campaign);
    }

    function resolveCampaignSentDate(campaign) {
        return formatDateTime(campaign?.sentAt || campaign?.updated || campaign?.created);
    }

    function resolveCampaignUpdatedDate(campaign) {
        return formatDateTime(campaign?.updated || campaign?.created);
    }

    function resolveCampaignSendConfirmationCount(campaign) {
        return resolveCampaignRecipientsCount(campaign);
    }

    function resolveCampaignSendActionLabel(campaign) {
        return normalizeStatus(campaign?.status) === "sent" ? "Sent" : "Send";
    }

    function resolveCampaignPrimaryMeta(campaign) {
        return `${resolveCampaignAudienceSummary(campaign)} - ${resolveCampaignRecipientsSummary(campaign)}`;
    }

    function resolveCampaignSecondaryMeta(campaign) {
        return `${resolveCampaignDeliveredSummary(campaign)} - ${resolveCampaignTimelineLabel(campaign)}`;
    }

    function resolveCampaignSentMeta(campaign) {
        return `Sent: ${resolveCampaignSentDate(campaign)}`;
    }

    function resolveCampaignDraftMeta(campaign) {
        return `Updated: ${resolveCampaignUpdatedDate(campaign)}`;
    }

    function resolveCampaignMetaDate(campaign) {
        return normalizeStatus(campaign?.status) === "sent"
            ? resolveCampaignSentMeta(campaign)
            : resolveCampaignDraftMeta(campaign);
    }

    function resolveCampaignMetaLine(campaign) {
        return `${resolveCampaignPrimaryMeta(campaign)} - ${resolveCampaignDeliveredSummary(campaign)} - ${resolveCampaignMetaDate(campaign)}`;
    }

    function resolveCampaignSubject(campaign) {
        return `${campaign?.subject || ""}`.trim() || "(No subject)";
    }

    function resolveCampaignEmptyStateHint() {
        return "Create your first draft in Builder, then select audience and send.";
    }

    function resolveSubscriberEmptyStateHint() {
        return "Add your first subscriber to start building campaigns.";
    }

    function resolveSubscriberFilterEmptyHint() {
        return "Try adjusting search, status, or group filters.";
    }

    function resolveAudienceStepHint() {
        return "Finalize recipients, then save or send from your drafts list.";
    }

    function resolveAudienceHealthStatus({ hasMissingBasics = false, needsAttention = false, canSend = false } = {}) {
        if (hasMissingBasics) {
            return "Missing basics";
        }
        if (needsAttention || !canSend) {
            return "Needs attention";
        }
        return "Ready to send";
    }

    function resolveAudienceHealthPillClass(status) {
        if (status === "Ready to send") {
            return "label-success";
        }
        if (status === "Needs attention") {
            return "label-warning";
        }
        return "label-danger";
    }

    function resolveBuilderStepHint() {
        return "Write your subject and content, then review the preview.";
    }

    function setCampaignPreviewMode(modeKey) {
        if (!campaignPreviewModes.some((mode) => mode.key === modeKey)) {
            return;
        }
        campaignPreviewMode = modeKey;
    }

    function escapePreviewText(value) {
        return `${value || ""}`
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#39;");
    }

    function sanitizePreviewBodyHtml(value) {
        return `${value || ""}`
            .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, "")
            .trim();
    }

    function resolveCampaignPreviewBaseHref() {
        if (typeof window === "undefined") {
            return "/";
        }

        const fallbackBase = `${window.location.origin || ""}/`;
        const configuredBackendUrl = `${import.meta.env.PB_BACKEND_URL || ""}`.trim();

        if (!configuredBackendUrl) {
            return fallbackBase;
        }

        try {
            return new URL(configuredBackendUrl, window.location.href).toString();
        } catch (_) {
            return fallbackBase;
        }
    }

    function normalizeCampaignPreviewMediaUrl(rawUrl, previewBaseHref) {
        const source = `${rawUrl || ""}`.trim();
        if (!source) {
            return source;
        }

        // Keep explicit non-http resource schemes unchanged.
        if (/^(data:|blob:|cid:|mailto:|tel:|javascript:)/i.test(source)) {
            return source;
        }

        try {
            const resolved = new URL(source, previewBaseHref);
            if (typeof window !== "undefined") {
                const currentUrl = new URL(window.location.href);
                const isLoopback = (hostname) => hostname === "localhost" || hostname === "127.0.0.1";
                if (
                    isLoopback(resolved.hostname)
                    && isLoopback(currentUrl.hostname)
                    && resolved.protocol === currentUrl.protocol
                    && resolved.port === currentUrl.port
                ) {
                    resolved.hostname = currentUrl.hostname;
                }
            }
            return resolved.toString();
        } catch (_) {
            return source;
        }
    }

    function normalizeCampaignPreviewBodyHtml(body, previewBaseHref) {
        const sanitizedBody = sanitizePreviewBodyHtml(body);
        if (!sanitizedBody || typeof window === "undefined" || typeof DOMParser === "undefined") {
            return sanitizedBody;
        }

        try {
            const parser = new DOMParser();
            const doc = parser.parseFromString(`<div id="campaign-preview-root">${sanitizedBody}</div>`, "text/html");
            const root = doc.getElementById("campaign-preview-root");

            if (!root) {
                return sanitizedBody;
            }

            root.querySelectorAll("img[src], source[src], video[src], audio[src]").forEach((node) => {
                const rawSrc = node.getAttribute("src");
                if (!rawSrc) {
                    return;
                }
                node.setAttribute("src", normalizeCampaignPreviewMediaUrl(rawSrc, previewBaseHref));
            });

            return root.innerHTML.trim();
        } catch (_) {
            return sanitizedBody;
        }
    }

    function buildCampaignPreviewDocument(subject, body) {
        const trimmedSubject = `${subject || ""}`.trim();
        const escapedTitle = escapePreviewText(trimmedSubject || "Campaign preview");
        const previewBaseHref = resolveCampaignPreviewBaseHref();
        const normalizedBody = normalizeCampaignPreviewBodyHtml(body, previewBaseHref);
        const bodyHtml = normalizedBody || '<p class="preview-empty">Body preview will appear here.</p>';

        return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <base href="${escapePreviewText(previewBaseHref)}" />
  <title>${escapedTitle}</title>
  <style>
    :root { color-scheme: light; }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 24px;
      background: #f3f4f6;
      color: #111827;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      line-height: 1.5;
    }
    .preview-doc {
      max-width: 100%;
      margin: 0 auto;
      background: #ffffff;
      border: 1px solid #e5e7eb;
      border-radius: 10px;
      padding: 20px;
      box-shadow: 0 10px 24px rgba(17, 24, 39, 0.08);
    }
    .preview-subject {
      margin: 0 0 16px;
      font-size: 20px;
      font-weight: 600;
      line-height: 1.3;
    }
    .preview-content p:last-child { margin-bottom: 0; }
    .preview-content img {
      max-width: 100%;
      height: auto;
      display: block;
    }
    .preview-empty {
      margin: 0;
      color: #6b7280;
      font-size: 14px;
      font-style: italic;
    }
  </style>
</head>
<body>
  <main class="preview-doc">
    ${trimmedSubject ? `<h1 class="preview-subject">${escapePreviewText(trimmedSubject)}</h1>` : ""}
    <section class="preview-content">${bodyHtml}</section>
  </main>
</body>
</html>`;
    }

    function resolveCampaignsSectionHint() {
        return "Review drafts and sent campaigns. Sent campaigns are kept as history. Duplicate to edit and send again.";
    }

    function resolveSubscribersSectionHint() {
        return "Add, edit, and organize subscribers for this website.";
    }

    function resolveSubscribersTableConfirmedLabel() {
        return "Lifecycle";
    }

    function resolveSubscriberConfirmedValue(subscriber) {
        const normalizedStatus = normalizeStatus(subscriber?.status);
        if (normalizedStatus === "pending") {
            if (subscriber?.confirmationTokenExpiresAt) {
                return `Confirmation expires: ${formatDateTime(subscriber.confirmationTokenExpiresAt)}`;
            }
            return "Pending confirmation";
        }
        if (normalizedStatus === "unsubscribed") {
            if (subscriber?.unsubscribedAt) {
                return `Unsubscribed: ${formatDateTime(subscriber.unsubscribedAt)}`;
            }
            return "Unsubscribed";
        }
        if (subscriber?.confirmedAt) {
            return `Confirmed: ${formatDateTime(subscriber.confirmedAt)}`;
        }
        return "Active";
    }

    function resolveSubscribersTableAddedLabel() {
        return "Created";
    }

    function resolveSubscriberCreatedValue(subscriber) {
        return formatDateTime(subscriber?.created);
    }

    function resolveSubscriberCreateEmptyStateActionLabel() {
        return "Add subscriber";
    }

    function resolveCampaignListStats() {
        return `${campaigns.length} total - ${draftCampaigns.length} drafts - ${sentCampaigns.length} sent`;
    }

    function resolveSubscribersListStats() {
        return `${filteredSubscribers.length} shown - ${subscribers.length} total`;
    }

    function resolveSendConfirmationRecipients(campaign) {
        return resolveCampaignSendConfirmationCount(campaign);
    }

    function resolveSendConfirmationTitle(campaign) {
        return resolveCampaignSubject(campaign);
    }

    function resolveSendConfirmationHint() {
        return "This action sends the campaign immediately.";
    }

    function resolveCampaignFilterEmptyHint() {
        return "Try another status filter to view more campaigns.";
    }

    function resolveCampaignNoItemsHint() {
        return resolveCampaignEmptyStateHint();
    }

    function clearSubscriberFormError() {
        if (subscriberFormError) {
            subscriberFormError = "";
        }
    }

    function clearSubscriberGroupFormError() {
        if (subscriberGroupFormError) {
            subscriberGroupFormError = "";
        }
    }
    function clearCampaignFormError() {
        if (campaignFormError) {
            campaignFormError = "";
        }
    }

    function clearEditingSubscriberError() {
        if (editingSubscriberError) {
            editingSubscriberError = "";
        }
    }

    function resolveCollectionFieldNameByAliasPriority(collection, aliases = []) {
        if (!collection || !Array.isArray(collection.fields)) {
            return "";
        }

        const normalizedAliases = aliases.map((alias) => `${alias || ""}`.trim().toLowerCase()).filter(Boolean);
        if (!normalizedAliases.length) {
            return "";
        }

        const fieldsByNormalizedName = new Map(
            collection.fields
                .map((field) => {
                    const fieldName = `${field?.name || ""}`.trim();
                    if (!fieldName) {
                        return null;
                    }
                    return [fieldName.toLowerCase(), fieldName];
                })
                .filter(Boolean),
        );

        for (const alias of normalizedAliases) {
            const resolvedName = fieldsByNormalizedName.get(alias);
            if (resolvedName) {
                return resolvedName;
            }
        }

        return "";
    }

    function resolveCollectionFieldNameForMultiSelectAlias(collection, aliases = []) {
        if (!collection || !Array.isArray(collection.fields)) {
            return "";
        }

        const normalizedAliases = aliases.map((alias) => `${alias || ""}`.trim().toLowerCase()).filter(Boolean);
        if (!normalizedAliases.length) {
            return "";
        }

        const aliasCandidates = collection.fields.filter((field) => {
            const fieldName = `${field?.name || ""}`.trim().toLowerCase();
            return fieldName && normalizedAliases.includes(fieldName);
        });

        const multiValueCandidate = aliasCandidates.find((field) => {
            const fieldType = `${field?.type || ""}`.trim().toLowerCase();
            const maxSelect = Number(field?.maxSelect ?? 1);
            return ["relation", "select", "file"].includes(fieldType) && maxSelect !== 1;
        });

        if (multiValueCandidate?.name) {
            return `${multiValueCandidate.name}`.trim();
        }

        return resolveCollectionFieldNameByAliasPriority(collection, aliases);
    }

    function hasSelectedGroup(selectedGroupIds, groupId) {
        const normalizedGroupId = `${groupId || ""}`.trim();
        if (!normalizedGroupId) {
            return false;
        }
        return normalizeIdList(selectedGroupIds).includes(normalizedGroupId);
    }

    function focusSubscriberEmailInput() {
        subscriberEmailInput?.focus();
    }

    function toggleSubscriberFormGroup(groupId) {
        if (!groupId || !hasSubscriberGroupsFeature) {
            return;
        }

        clearSubscriberFormError();

        const normalizedGroupId = `${groupId || ""}`.trim();
        if (!normalizedGroupId) {
            return;
        }

        const currentGroupIds = normalizeIdList(subscriberForm.groupIds);
        const nextGroupIds = currentGroupIds.includes(normalizedGroupId)
            ? currentGroupIds.filter((id) => id !== normalizedGroupId)
            : [...currentGroupIds, normalizedGroupId];

        subscriberForm = { ...subscriberForm, groupIds: nextGroupIds };
    }

    function startEditSubscriber(subscriber) {
        if (!subscriber?.id) {
            return;
        }

        editingSubscriberId = subscriber.id;
        editingSubscriberError = "";
        editingSubscriberForm = {
            name: normalizeSubscriberName(subscriber.name),
            email: `${subscriber.email || ""}`,
            status: normalizeStatus(subscriber.status) || "pending",
            groupIds: [...getSubscriberGroupIds(subscriber)],
        };
    }

    function cancelEditSubscriber() {
        editingSubscriberId = "";
        editingSubscriberError = "";
        editingSubscriberForm = {
            name: "",
            email: "",
            status: "pending",
            groupIds: [],
        };
    }

    function toggleEditingSubscriberGroup(groupId) {
        if (!groupId || !hasSubscriberGroupsFeature || !editingSubscriberId) {
            return;
        }

        clearEditingSubscriberError();

        const normalizedGroupId = `${groupId || ""}`.trim();
        if (!normalizedGroupId) {
            return;
        }

        const currentGroupIds = normalizeIdList(editingSubscriberForm.groupIds);
        const nextGroupIds = currentGroupIds.includes(normalizedGroupId)
            ? currentGroupIds.filter((id) => id !== normalizedGroupId)
            : [...currentGroupIds, normalizedGroupId];

        editingSubscriberForm = { ...editingSubscriberForm, groupIds: nextGroupIds };
    }

    function autoSelectCreatedGroupForCurrentFlow(groupId) {
        if (!groupId || !hasSubscriberGroupsFeature) {
            return;
        }

        if (editingSubscriberId) {
            const nextEditingGroupIds = new Set(normalizeIdList(editingSubscriberForm.groupIds));
            nextEditingGroupIds.add(groupId);
            editingSubscriberForm = { ...editingSubscriberForm, groupIds: [...nextEditingGroupIds] };
            return;
        }

        const nextCreateGroupIds = new Set(normalizeIdList(subscriberForm.groupIds));
        nextCreateGroupIds.add(groupId);
        subscriberForm = { ...subscriberForm, groupIds: [...nextCreateGroupIds] };
    }
    function setActiveSection(section) {        if (newsletterSections.has(section)) {            activeSection = section;            if (section === "campaigns") {                campaignWorkspace = "builder";                campaignBuilderShowEditor = true;                campaignBuilderShowPreview = false;            }        }    }    function setSubscribersPage(page) {        const nextPage = Math.min(Math.max(page, 1), subscribersTotalPages);        subscribersPage = nextPage;    }    function isManualRecipientSelected(subscriberId) {        return normalizedCampaignManualRecipientIdsSet.has(subscriberId);    }    function toggleManualRecipient(subscriberId) {        if (!subscriberId || !activeSubscriberIdsSet.has(subscriberId)) {            return;        }        if (normalizeStatus(campaignForm.recipientsType) !== "manual") {            setCampaignRecipientsType("manual");        }        const nextSelection = new Set(normalizedCampaignManualRecipientIds);        if (nextSelection.has(subscriberId)) {            nextSelection.delete(subscriberId);        } else {            nextSelection.add(subscriberId);        }        setManualRecipientIds([...nextSelection]);    }    function isSubscriberSelected(subscriberId) {        return selectedSubscriberIds.includes(subscriberId);    }    function toggleSubscriberSelection(subscriberId) {        if (selectedSubscriberIds.includes(subscriberId)) {            selectedSubscriberIds = selectedSubscriberIds.filter((id) => id !== subscriberId);        } else {            selectedSubscriberIds = [...selectedSubscriberIds, subscriberId];        }    }    function toggleAllVisibleSubscribers() {        if (areAllVisibleSubscribersSelected) {            selectedSubscriberIds = selectedSubscriberIds.filter((id) => !visibleSubscriberIds.includes(id));            return;        }        const nextSelectedIds = new Set(selectedSubscriberIds);        visibleSubscriberIds.forEach((id) => nextSelectedIds.add(id));        selectedSubscriberIds = [...nextSelectedIds];    }    function resetSubscriberSelection() {        selectedSubscriberIds = [];    }    function openSendCampaignModal(campaign) {        const reason = getSendCampaignDisabledReason(campaign);        if (reason) {            return;        }        pendingSendCampaign = campaign;    }    function closeSendCampaignModal() {
        pendingSendCampaign = null;
    }

    function openDeleteCampaignModal(campaign) {
        if (!campaign?.id || deletingCampaignId || normalizeStatus(campaign?.status) !== "draft") {
            return;
        }

        pendingDeleteCampaign = campaign;
    }

    function closeDeleteCampaignModal() {
        if (deletingCampaignId) {
            return;
        }

        pendingDeleteCampaign = null;
    }

    function openCampaignPreviewModal() {
        isCampaignPreviewExpanded = true;
    }

    function closeCampaignPreviewModal() {
        isCampaignPreviewExpanded = false;
    }

    function toggleCampaignBuilderPanel(panel) {
        const nextShowEditor = panel === "edit" ? !campaignBuilderShowEditor : campaignBuilderShowEditor;
        const nextShowPreview = panel === "preview" ? !campaignBuilderShowPreview : campaignBuilderShowPreview;

        if (!nextShowEditor && !nextShowPreview) {
            return;
        }

        campaignBuilderShowEditor = nextShowEditor;
        campaignBuilderShowPreview = nextShowPreview;
    }

    function openDeleteSubscriberModal(subscriber) {
        if (!subscriber?.id || !!deletingSubscriberId) {
            return;
        }

        pendingDeleteSubscriber = subscriber;
    }

    function closeDeleteSubscriberModal() {
        if (deletingSubscriberId) {
            return;
        }

        pendingDeleteSubscriber = null;
    }
    async function confirmSendCampaign() {
        if (!pendingSendCampaign) {
            return;
        }
        const campaign = pendingSendCampaign;
        pendingSendCampaign = null;
        await sendCampaign(campaign);
    }

    async function confirmDeleteSubscriber() {
        if (!pendingDeleteSubscriber) {
            return;
        }

        const subscriber = pendingDeleteSubscriber;
        pendingDeleteSubscriber = null;
        await deleteSubscriber(subscriber);
    }

    async function confirmDeleteCampaign() {
        if (!pendingDeleteCampaign) {
            return;
        }

        const campaign = pendingDeleteCampaign;
        pendingDeleteCampaign = null;
        await deleteDraftCampaign(campaign);
    }
    async function loadWebsites() {        if (!websitesCollection?.id) {            websites = [];            selectedWebsiteId = "";            return;        }        isLoadingWebsites = true;        try {            websites = await ApiClient.collection(websitesCollection.id).getFullList({                sort: resolveWebsitesSort(websitesCollection),                requestKey: "nuvio_newsletter_websites",            });            if (!websites.length) {
                selectedWebsiteId = "";
                subscribers = [];
                campaigns = [];
                subscriberGroups = [];
                resetSubscriberSelection();
                return;
            }
            if (!websites.find((website) => website.id === selectedWebsiteId)) {                selectedWebsiteId = websites[0].id;            }        } catch (err) {            websites = [];
            selectedWebsiteId = "";
            subscribers = [];
            campaigns = [];
            subscriberGroups = [];
            resetSubscriberSelection();
            ApiClient.error(err);
        }
        isLoadingWebsites = false;    }    async function loadSubscribers() {        if (!hasNewsletterCollections || !selectedWebsiteId) {            subscribers = [];            resetSubscriberSelection();            return;        }        isLoadingSubscribers = true;        try {            subscribers = await ApiClient.collection(subscribersCollection.id).getFullList({                filter: `website="${selectedWebsiteId}"`,                sort: "-created",                requestKey: "nuvio_newsletter_subscribers_" + selectedWebsiteId,            });        } catch (err) {            subscribers = [];            ApiClient.error(err);        }        isLoadingSubscribers = false;    }    async function loadCampaigns() {
        if (!hasNewsletterCollections || !selectedWebsiteId) {
            campaigns = [];
            return;
        }
        isLoadingCampaigns = true;        try {            campaigns = await ApiClient.collection(campaignsCollection.id).getFullList({                filter: `website="${selectedWebsiteId}"`,                sort: "-created",                requestKey: "nuvio_newsletter_campaigns_" + selectedWebsiteId,            });        } catch (err) {            campaigns = [];            ApiClient.error(err);        }        isLoadingCampaigns = false;
    }

    async function loadSubscriberGroups() {
        if (!selectedWebsiteId || !hasSubscriberGroupsFeature) {
            subscriberGroups = [];
            return;
        }

        isLoadingSubscriberGroups = true;

        try {
            subscriberGroups = await ApiClient.collection(subscriberGroupsCollection.id).getFullList({
                filter: `website="${selectedWebsiteId}"`,
                sort: "+name",
                requestKey: "nuvio_newsletter_subscriber_groups_" + selectedWebsiteId,
            });
        } catch (err) {
            subscriberGroups = [];
            ApiClient.error(err);
        }

        isLoadingSubscriberGroups = false;
    }
    function handleWebsiteChange() {
        subscriberFormError = "";
        subscriberGroupFormError = "";
        campaignFormError = "";
        pendingSendCampaign = null;
        pendingDeleteCampaign = null;
        pendingDeleteSubscriber = null;
        subscriberGroupFilter = "all";
        subscriberForm = { ...subscriberForm, name: "", groupIds: [] };
        cancelEditSubscriber();
        resetSubscriberSelection();
        resetCampaignComposer();
    }
    async function createSubscriber() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingSubscriber) {
            return;
        }
        const email = normalizedSubscriberFormEmail;        if (!isValidEmail(email)) {            subscriberFormError = "Please provide a valid subscriber email.";            return;        }        if (subscriberAlreadyExists) {            subscriberFormError = "This email is already subscribed for this website.";            return;        }        subscriberFormError = "";        isCreatingSubscriber = true;        try {            const payload = {
                website: selectedWebsiteId,
                email,
                status: "pending",
            };
            if (subscribersSupportsNameField) {
                payload.name = normalizeSubscriberName(subscriberForm.name);
            }
            if (subscribersSupportsSourceField) {
                payload.source = subscriberLeadSource;
            }
            if (hasSubscriberGroupsFeature && Array.isArray(subscriberForm.groupIds) && subscriberForm.groupIds.length) {
                payload[subscriberGroupsFieldName] = normalizeIdList(subscriberForm.groupIds);
            }
            await ApiClient.collection(subscribersCollection.id).create(payload);            subscriberForm = {
                name: "",
                email: "",
                status: "pending",
                groupIds: [],
            };
            await loadSubscribers();            addSuccessToast("Subscriber added.");            focusSubscriberEmailInput();        } catch (err) {            ApiClient.error(err);        }        isCreatingSubscriber = false;
    }

    async function saveSubscriberEdit(subscriber) {
        if (!subscriber?.id || !editingSubscriberId || editingSubscriberId !== subscriber.id || isSavingSubscriber) {
            return;
        }

        const email = normalizeEmail(editingSubscriberForm.email);
        if (!isValidEmail(email)) {
            editingSubscriberError = "Please provide a valid subscriber email.";
            return;
        }

        const duplicateEmail = subscribers.some((item) => item.id !== subscriber.id && normalizeEmail(item?.email) === email);
        if (duplicateEmail) {
            editingSubscriberError = "This email is already subscribed for this website.";
            return;
        }

        editingSubscriberError = "";
        isSavingSubscriber = true;

        try {
            const payload = {
                email,
                status: normalizeStatus(subscriber?.status) || "pending",
            };
            if (subscribersSupportsNameField) {
                payload.name = normalizeSubscriberName(editingSubscriberForm.name);
            }

            if (hasSubscriberGroupsFeature) {
                payload[subscriberGroupsFieldName] = normalizeIdList(editingSubscriberForm.groupIds);
            }

            if (payload.status === "active" && !subscriber.confirmedAt) {
                payload.confirmedAt = new Date().toISOString();
            }

            await ApiClient.collection(subscribersCollection.id).update(subscriber.id, payload);
            await loadSubscribers();
            cancelEditSubscriber();
            addSuccessToast("Subscriber updated.");
        } catch (err) {
            ApiClient.error(err);
        }

        isSavingSubscriber = false;
    }

    async function deleteSubscriber(subscriber) {
        if (!subscriber?.id || deletingSubscriberId) {
            return;
        }

        deletingSubscriberId = subscriber.id;

        try {
            await ApiClient.collection(subscribersCollection.id).delete(subscriber.id);
            if (editingSubscriberId === subscriber.id) {
                cancelEditSubscriber();
            }
            selectedSubscriberIds = selectedSubscriberIds.filter((id) => id !== subscriber.id);
            await loadSubscribers();
            addSuccessToast("Subscriber deleted.");
        } catch (err) {
            ApiClient.error(err);
        }

        if (pendingDeleteSubscriber?.id === subscriber.id) {
            pendingDeleteSubscriber = null;
        }

        deletingSubscriberId = "";
    }

    async function createSubscriberGroup() {
        if (!hasSubscriberGroupsFeature || !selectedWebsiteId || isCreatingSubscriberGroup) {
            return;
        }

        const name = normalizeGroupName(subscriberGroupForm.name);
        if (!name) {
            subscriberGroupFormError = "Please provide a group name.";
            return;
        }

        if (subscriberGroupNameAlreadyExists) {
            subscriberGroupFormError = "A group with this name already exists for this website.";
            return;
        }

        subscriberGroupFormError = "";
        isCreatingSubscriberGroup = true;

        try {
            const createdGroup = await ApiClient.collection(subscriberGroupsCollection.id).create({
                website: selectedWebsiteId,
                name,
                slug: slugifyGroupName(name),
            });

            subscriberGroupForm = { name: "" };
            autoSelectCreatedGroupForCurrentFlow(createdGroup?.id);
            await loadSubscriberGroups();
            addSuccessToast("Subscriber group added.");
        } catch (err) {
            ApiClient.error(err);
        }

        isCreatingSubscriberGroup = false;
    }

    async function setSubscriberStatus(subscriber, status) {        if (!subscriber?.id || !hasNewsletterCollections) {            return;        }        try {            const payload = { status };            if (status === "active" && !subscriber.confirmedAt) {                payload.confirmedAt = new Date().toISOString();            }            await ApiClient.collection(subscribersCollection.id).update(subscriber.id, payload);            await loadSubscribers();            addSuccessToast(`Subscriber marked as ${status}.`);        } catch (err) {            ApiClient.error(err);        }    }    async function applyBulkStatus(status) {        if (!selectedSubscriberIds.length || isBulkUpdating || !hasNewsletterCollections) {            return;        }        isBulkUpdating = true;        try {            const updates = selectedSubscriberIds                .map((subscriberId) => {                    const subscriber = subscribers.find((item) => item.id === subscriberId);                    if (!subscriber || normalizeStatus(subscriber.status) === normalizeStatus(status)) {                        return null;                    }                    const payload = { status };                    if (status === "active" && !subscriber.confirmedAt) {                        payload.confirmedAt = new Date().toISOString();                    }                    return ApiClient.collection(subscribersCollection.id).update(subscriber.id, payload);                })                .filter(Boolean);            if (!updates.length) {
                addSuccessToast("Selected subscribers are already up to date.");
                isBulkUpdating = false;
                return;
            }
            await Promise.all(updates);            resetSubscriberSelection();            await loadSubscribers();            addSuccessToast(`Updated ${updates.length} subscriber(s).`);        } catch (err) {            ApiClient.error(err);        }        isBulkUpdating = false;    }    async function sendCampaign(campaign) {        if (!campaign?.id || isSendingCampaign[campaign.id]) {            return false;        }        isSendingCampaign[campaign.id] = true;        isSendingCampaign = { ...isSendingCampaign };        let sent = false;        try {            const response = await ApiClient.send("/api/nuvio/newsletter/campaigns/send", {                method: "POST",                body: {                    campaignId: campaign.id,                },                requestKey: "nuvio_newsletter_send_" + campaign.id,            });            const sentCountValue = Number(response?.sentCount);            const failedCountValue = Number(response?.failedCount);            const recipientsCountValue = Number(response?.recipientsCount);            const hasSentCount = Number.isFinite(sentCountValue) && sentCountValue >= 0;            const hasFailedCount = Number.isFinite(failedCountValue) && failedCountValue >= 0;            const hasRecipientsCount = Number.isFinite(recipientsCountValue) && recipientsCountValue >= 0;            if (hasSentCount && hasFailedCount) {                if (sentCountValue > 0 && failedCountValue > 0) {                    addWarningToast(`Campaign sent to ${sentCountValue} subscriber(s). ${failedCountValue} recipient(s) failed.`);                    sent = true;                } else if (sentCountValue > 0) {                    addSuccessToast(`Campaign sent to ${sentCountValue} subscriber(s).`);                    sent = true;                } else if (failedCountValue > 0) {                    addErrorToast("Campaign could not be sent. Please try again or check email configuration.");                } else {                    addSuccessToast("Campaign sent.");                    sent = true;                }            } else if (hasRecipientsCount) {                addSuccessToast(`Campaign sent to ${recipientsCountValue} subscriber(s).`);                sent = true;            } else {                addSuccessToast("Campaign sent.");                sent = true;            }            await loadCampaigns();        } catch (err) {            ApiClient.error(err, false);            addErrorToast("Campaign could not be sent. Please try again or check email configuration.");        }        delete isSendingCampaign[campaign.id];        isSendingCampaign = { ...isSendingCampaign };        return sent;    }    function resetCampaignComposer() {
        campaignForm = {
            subject: "",
            body: "",
            recipientsType: "manual",
            recipientsIds: [],
        };
        audienceRecipientSearch = "";
        audienceRecipientVisibilityFilter = "all";
        campaignFormError = "";
        campaignBuilderShowEditor = true;
        campaignBuilderShowPreview = false;
        closeCampaignPreviewModal();
        editingCampaignId = "";
        viewingSentCampaignId = "";
    }

    function startEditCampaign(campaign) {
        if (!campaign?.id) {
            return;
        }

        const previousRecipientsType = getCampaignRecipientsType(campaign);
        const resolvedRecipientsIds = previousRecipientsType === "manual"
            ? normalizeManualRecipientIds(getCampaignRecipientIds(campaign))
            : getActiveSubscriberIds();

        editingCampaignId = campaign.id;
        viewingSentCampaignId = "";
        campaignForm = {
            subject: `${campaign.subject || ""}`,
            body: `${campaign.body || ""}`,
            recipientsType: "manual",
            recipientsIds: normalizeManualRecipientIds(resolvedRecipientsIds),
        };
        audienceRecipientSearch = "";
        audienceRecipientVisibilityFilter = "all";
        campaignBuilderShowEditor = true;
        campaignBuilderShowPreview = false;
        closeCampaignPreviewModal();
        campaignFormError = "";
    }

    function startViewCampaign(campaign) {
        if (!campaign?.id) {
            return;
        }

        const previousRecipientsType = getCampaignRecipientsType(campaign);
        const resolvedRecipientsIds = previousRecipientsType === "all"
            ? getActiveSubscriberIds()
            : getCampaignRecipientIds(campaign);

        editingCampaignId = campaign.id;
        viewingSentCampaignId = campaign.id;
        campaignForm = {
            subject: `${campaign.subject || ""}`,
            body: `${campaign.body || ""}`,
            recipientsType: "manual",
            recipientsIds: normalizeManualRecipientIds(resolvedRecipientsIds),
        };
        audienceRecipientSearch = "";
        audienceRecipientVisibilityFilter = "all";
        campaignBuilderShowEditor = false;
        campaignBuilderShowPreview = true;
        campaignWorkspace = "builder";
        closeCampaignPreviewModal();
        campaignFormError = "";
    }

    async function saveCampaignDraftFromComposer() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingCampaign || isSavingCampaign) {
            return;
        }

        if (isViewingSentCampaign) {
            campaignFormError = "Sent campaigns are kept as history. Duplicate to edit and send again.";
            return;
        }

        const validationError = resolveCreateCampaignDisabledReason(
            false,
            selectedWebsiteId,
            campaignForm.subject,
            campaignForm.body,
            campaignForm.recipientsType,
            campaignForm.recipientsIds,
        );

        if (validationError) {
            campaignFormError = validationError;
            return;
        }

        campaignFormError = "";
        const normalizedRecipientsType = "manual";
        const normalizedRecipientsIds = normalizeManualRecipientIds(campaignForm.recipientsIds);
        const payload = {
            website: selectedWebsiteId,
            subject: `${campaignForm.subject || ""}`.trim(),
            body: `${campaignForm.body || ""}`.trim(),
            status: "draft",
            recipientsType: normalizedRecipientsType,
            recipientsIds: normalizedRecipientsIds,
        };
        payload[campaignRecipientsTypeFieldName] = normalizedRecipientsType;
        payload[campaignRecipientsIdsFieldName] = normalizedRecipientsIds;

        if (editingCampaignId) {
            isSavingCampaign = true;
            try {
                await ApiClient.collection(campaignsCollection.id).update(editingCampaignId, payload);
                await loadCampaigns();
                addSuccessToast("Campaign draft updated.");
                resetCampaignComposer();
                campaignWorkspace = "audience";
            } catch (err) {
                ApiClient.error(err);
            }
            isSavingCampaign = false;
            return;
        }

        isCreatingCampaign = true;
        try {
            await ApiClient.collection(campaignsCollection.id).create({
                ...payload,
                recipientsCount: 0,
            });
            await loadCampaigns();
            addSuccessToast("Draft campaign created.");
            resetCampaignComposer();
            campaignWorkspace = "audience";
        } catch (err) {
            ApiClient.error(err);
        }
        isCreatingCampaign = false;
    }

    async function duplicateCampaignAsDraft(campaign) {
        if (!campaign?.id || !hasNewsletterCollections || !selectedWebsiteId || isDuplicatingCampaign[campaign.id]) {
            return;
        }

        isDuplicatingCampaign[campaign.id] = true;
        isDuplicatingCampaign = { ...isDuplicatingCampaign };

        const copiedRecipientsType = normalizeStatus(getCampaignRecipientsType(campaign)) === "all" ? "all" : "manual";
        const copiedRecipientsIds = normalizeIdList(getCampaignRecipientIds(campaign));
        const payload = {
            website: campaign.website || selectedWebsiteId,
            subject: `${campaign.subject || ""}`.trim(),
            body: `${campaign.body || ""}`.trim(),
            status: "draft",
            recipientsType: copiedRecipientsType,
            recipientsIds: copiedRecipientsIds,
            recipientsCount: 0,
            sentAt: null,
        };
        payload[campaignRecipientsTypeFieldName] = copiedRecipientsType;
        payload[campaignRecipientsIdsFieldName] = copiedRecipientsIds;

        try {
            const createdCampaign = await ApiClient.collection(campaignsCollection.id).create(payload);
            await loadCampaigns();
            const nextDraft = campaigns.find((item) => item.id === createdCampaign?.id) || createdCampaign;
            startEditCampaign(nextDraft);
            campaignWorkspace = "builder";
            addSuccessToast("Draft created from sent campaign.");
        } catch (err) {
            ApiClient.error(err);
        }

        delete isDuplicatingCampaign[campaign.id];
        isDuplicatingCampaign = { ...isDuplicatingCampaign };
    }

    async function deleteDraftCampaign(campaign) {
        if (
            !campaign?.id
            || !hasNewsletterCollections
            || deletingCampaignId
            || normalizeStatus(campaign?.status) !== "draft"
        ) {
            return;
        }

        deletingCampaignId = campaign.id;

        try {
            await ApiClient.collection(campaignsCollection.id).delete(campaign.id);

            if (editingCampaignId === campaign.id) {
                resetCampaignComposer();
                campaignWorkspace = "builder";
            }

            if (pendingSendCampaign?.id === campaign.id) {
                pendingSendCampaign = null;
            }

            await loadCampaigns();
            addSuccessToast("Draft deleted.");
        } catch (err) {
            console.error("Unable to delete draft campaign:", err);
            addErrorToast("Unable to delete draft right now.");
        }

        if (pendingDeleteCampaign?.id === campaign.id) {
            pendingDeleteCampaign = null;
        }

        deletingCampaignId = "";
    }


    function refreshAll() {
        loadWebsites();
        loadSubscribers();
        loadCampaigns();
        loadSubscriberGroups();
    }
    // NUVIO CUSTOM END: Newsletter V1 dedicated section/page (collection-backed).
</script>

<PageWrapper>
    {#if $isCollectionsLoading || (!$hasCollectionsLoaded && !$collectionsLoadError)}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading Newsletter...</h1>
        </div>
    {:else if $collectionsLoadError}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>
                Could not verify Newsletter collections.<br />
                Refresh the page or check your connection.
            </div>
        </div>
    {:else if $hasCollectionsLoaded && !hasNewsletterCollections}
        <div class="alert alert-warning m-b-base">
            <div class="icon">
                <i class="ri-information-line" />
            </div>
            <div>
                Newsletter collections are missing:
                <strong>{missingCollectionNames.join(", ")}</strong>.
                Run the latest migrations to enable Newsletter V1.
            </div>
        </div>
    {:else}
        <section class="newsletter-head operations-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <div class="title-row">
                        <h2 class="m-0">Newsletter Operations</h2>
                        <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={refreshAll} />
                    </div>
                    <p class="txt-sm txt-hint m-b-0 head-description">Manage subscribers and campaigns by website in one place.</p>
                </div>

                <div class="head-selector operations-website-select">
                    <div class="selector-row">
                        <label class="txt-sm txt-hint selector-label m-b-0" for="newsletter-website">Website</label>
                        <select
                            id="newsletter-website"
                            class="input input-sm"
                            bind:value={selectedWebsiteId}
                            disabled={isLoadingWebsites || !websites.length}
                            on:change={handleWebsiteChange}
                        >
                            {#if !websites.length}                                <option value="">No websites available</option>                            {:else}                                {#each websites as website (website.id)}                                    <option value={website.id}>{resolveWebsiteLabel(website)}</option>                                {/each}
                            {/if}
                        </select>
                    </div>
                </div>
            </div>
            <div class="head-tools">
                <div class="tabs-header compact combined left operations-tabs">
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "subscribers"}
                        on:click={() => setActiveSection("subscribers")}
                    >
                        <i class="ri-user-3-line tab-icon" aria-hidden="true" />
                        <span class="tab-label">Subscribers</span>
                    </button>
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "campaigns"}
                        on:click={() => setActiveSection("campaigns")}
                    >
                        <i class="ri-megaphone-line tab-icon" aria-hidden="true" />
                        <span class="tab-label">Campaigns</span>
                    </button>
                </div>

                <div class="summary-badges">
                    <span class="summary-pill">
                        <i class="ri-user-3-line" />
                        {subscribers.length} subscribers
                    </span>
                    <span class="summary-pill">
                        <i class="ri-user-follow-line" />
                        {activeSubscribers.length} active
                    </span>
                    <span class="summary-pill">
                        <i class="ri-megaphone-line" />
                        {campaigns.length} campaigns
                    </span>
                    <span class="summary-pill">
                        <i class="ri-draft-line" />
                        {draftCampaigns.length} drafts
                    </span>
                    <span class="summary-pill">
                        <i class="ri-send-plane-2-line" />
                        {sentCampaigns.length} sent
                    </span>
                </div>
            </div>
        </section>
        {#if !selectedWebsiteId}
            <div class="placeholder-section m-b-base">
                <h1>Select a website to manage Newsletter.</h1>
                <p class="txt-sm txt-hint m-b-0">Once selected, subscribers and campaigns will be loaded automatically.</p>
            </div>
        {:else}
            <div class="tabs">
                <div class="tabs-content">
                    {#if activeSection === "subscribers"}
                        <section class="panel subscribers-section-panel">
                            <div class="subscribers-panel-header m-b-sm">
                                <div class="section-head section-head-inline m-b-0">
                                    <h4 class="m-0">Subscribers</h4>
                                    <span class="txt-sm txt-hint">{resolveSubscribersSectionHint()}</span>
                                </div>
                                <div class="flex-fill" />
                                <div class="subscribers-panel-header-actions">
                                    <button
                                        type="button"
                                        class="btn btn-sm add-form-toggle-btn"
                                        class:btn-outline={isSubscriberCreateOpen}
                                        on:click={() => (isSubscriberCreateOpen = !isSubscriberCreateOpen)}
                                    >
                                        <i class={isSubscriberCreateOpen ? "ri-eye-off-line" : "ri-add-line"} aria-hidden="true" />
                                        <span class="txt">{isSubscriberCreateOpen ? "Hide form" : "Add form"}</span>
                                    </button>
                                    <span class="txt-sm txt-hint">
                                        {resolveSubscribersListStats()}
                                    </span>
                                </div>
                            </div>

                            {#if isSubscriberCreateOpen}
                                <form class="subscriber-create-form subscriber-inline-create m-b-sm" on:submit|preventDefault={createSubscriber}>
                                    <div class="subscriber-create-row subscriber-create-row-primary" class:no-name={!subscribersSupportsNameField}>
                                        {#if subscribersSupportsNameField}
                                            <div class="create-name-field">
                                                <label class="txt-sm txt-hint block m-b-5" for="subscriber-name">Name (optional)</label>
                                                <input
                                                    id="subscriber-name"
                                                    type="text"
                                                    class="input input-sm create-name-input"
                                                    placeholder="Subscriber name"
                                                    bind:value={subscriberForm.name}
                                                    on:input={clearSubscriberFormError}
                                                />
                                            </div>
                                        {/if}
                                        <div class="create-email-field">
                                            <label class="txt-sm txt-hint block m-b-5" for="subscriber-email">Email</label>
                                            <input
                                                id="subscriber-email"
                                                bind:this={subscriberEmailInput}
                                                type="email"
                                                class="input input-sm"
                                                placeholder="name@example.com"
                                                bind:value={subscriberForm.email}
                                                on:input={clearSubscriberFormError}
                                            />
                                        </div>
                                        <div class="create-action-field">
                                            <button
                                                type="submit"
                                                class="btn add-subscriber-btn"
                                                class:btn-loading={isCreatingSubscriber}
                                                disabled={!!createSubscriberDisabledReason}
                                                title={createSubscriberDisabledReason || null}
                                            >
                                                <span class="txt">Add subscriber</span>
                                            </button>
                                        </div>
                                    </div>
                                    {#if hasSubscriberGroupsFeature}
                                        <div class="subscriber-create-row subscriber-create-row-groups">
                                            <div class="subscriber-groups-select">
                                                <label class="txt-sm txt-hint block m-b-5">Assign groups (optional)</label>
                                                <div class="group-pill-list form-group-pill-list">
                                                    {#if isLoadingSubscriberGroups}
                                                        <span class="txt-sm txt-hint">Loading groups...</span>
                                                    {:else if !subscriberGroups.length}
                                                        <span class="txt-sm txt-hint">No groups yet. Create your first group.</span>
                                                    {:else}
                                                        {#each subscriberGroups as group (group.id)}
                                                            <button
                                                                type="button"
                                                                class="group-pill-btn"
                                                                class:is-selected={hasSelectedGroup(subscriberForm.groupIds, group.id)}
                                                                on:click={() => toggleSubscriberFormGroup(group.id)}
                                                            >
                                                                {group.name}
                                                            </button>
                                                        {/each}
                                                    {/if}
                                                </div>
                                            </div>

                                            <div class="subscriber-groups-input">
                                                <label class="txt-sm txt-hint block m-b-5" for="subscriber-group-name">Create group</label>
                                                <input
                                                    id="subscriber-group-name"
                                                    type="text"
                                                    class="input input-sm"
                                                    placeholder="e.g. VIP Clients"
                                                    bind:value={subscriberGroupForm.name}
                                                    on:input={clearSubscriberGroupFormError}
                                                />
                                            </div>

                                            <div class="subscriber-groups-action">
                                                <button
                                                    type="button"
                                                    class="btn btn-sm btn-outline"
                                                    class:btn-loading={isCreatingSubscriberGroup}
                                                    disabled={!!createSubscriberGroupDisabledReason}
                                                    title={createSubscriberGroupDisabledReason || null}
                                                    on:click={createSubscriberGroup}
                                                >
                                                    <span class="txt">Add group</span>
                                                </button>
                                            </div>
                                        </div>
                                    {/if}
                                {#if subscriberFormError}
                                    <div>
                                        <div class="txt-sm txt-danger">{subscriberFormError}</div>
                                    </div>
                                {/if}
                                {#if subscriberGroupFormError}
                                    <div>
                                        <div class="txt-sm txt-danger">{subscriberGroupFormError}</div>
                                    </div>
                                {/if}
                                </form>
                            {/if}

                            <div class="subscriber-controls m-b-sm">
                                <div class="subscriber-filter-grid" class:group-enabled={hasSubscriberGroupsFeature}>
                                    <div class="control-item">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-search">Search</label>
                                        <input
                                            id="subscriber-search"                                            type="text"                                            class="input input-sm"                                            placeholder="Search by name or email..."                                            bind:value={subscriberSearch}                                        />                                    </div>                                    <div class="control-item">                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-filter-status">Status</label>                                        <select                                            id="subscriber-filter-status"                                            class="input input-sm"                                            bind:value={subscriberStatusFilter}                                        >                                            <option value="all">All</option>                                            {#each subscriberStatuses as status}                                                <option value={status}>{getSubscriberStatusLabel(status)}</option>                                            {/each}                                        </select>                                    </div>                                    <div class="control-item">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-sort">Sort</label>
                                        <select id="subscriber-sort" class="input input-sm" bind:value={subscriberSort}>
                                            {#each subscriberSortOptions as sortOption}
                                                <option value={sortOption.value}>{sortOption.label}</option>
                                            {/each}
                                        </select>
                                    </div>
                                    {#if hasSubscriberGroupsFeature}
                                        <div class="control-item">
                                            <label class="txt-sm txt-hint block m-b-5" for="subscriber-filter-group">Group</label>
                                            <select
                                                id="subscriber-filter-group"
                                                class="input input-sm"
                                                bind:value={subscriberGroupFilter}
                                            >
                                                <option value="all">All groups</option>
                                                {#each subscriberGroups as group (group.id)}
                                                    <option value={group.id}>{group.name}</option>
                                                {/each}
                                            </select>
                                        </div>
                                    {/if}
                                </div>
                            </div>
                            {#if isLoadingSubscribers}
                                <div class="loading-state">
                                    <span class="loader loader-sm" />
                                    <span class="txt-hint">Loading subscribers...</span>
                                </div>
                            {:else if !subscribers.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No subscribers yet for this website.</span>
                                    <span class="txt-sm txt-hint">{resolveSubscriberEmptyStateHint()}</span>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        on:click={() => {
                                            isSubscriberCreateOpen = true;
                                            focusSubscriberEmailInput();
                                        }}
                                    >
                                        <span class="txt">{resolveSubscriberCreateEmptyStateActionLabel()}</span>
                                    </button>
                                </div>
                            {:else if !filteredSubscribers.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No subscribers match the current filters.</span>
                                    <span class="txt-sm txt-hint">{resolveSubscriberFilterEmptyHint()}</span>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        on:click={() => {
                                            subscriberSearch = "";
                                            subscriberStatusFilter = "all";
                                            subscriberGroupFilter = "all";
                                            subscriberSort = "newest";
                                        }}
                                    >
                                        <span class="txt">Clear filters</span>
                                    </button>
                                </div>
                            {:else}
                                <div class="list list-compact subscriber-table-list">
                                    <div class="subscriber-table-head txt-xs txt-hint">
                                        <div class="selection-cell subscriber-col-select subscriber-col-select-head">
                                            <input
                                                type="checkbox"
                                                checked={areAllVisibleSubscribersSelected}
                                                disabled={!pagedSubscribers.length}
                                                on:change={toggleAllVisibleSubscribers}
                                                aria-label="Select visible subscribers"
                                            />
                                        </div>
                                        <div class="subscriber-col-name">Name</div>
                                        <div class="subscriber-col-email">Email</div>
                                        <div class="subscriber-col-status">Status</div>
                                        <div class="subscriber-col-confirmed">{resolveSubscribersTableConfirmedLabel()}</div>
                                        <div class="subscriber-col-added">{resolveSubscribersTableAddedLabel()}</div>
                                        <div class="subscriber-col-groups">Groups</div>
                                        <div class="subscriber-col-actions">Actions</div>
                                    </div>
                                    <div class="list-content">                                        {#each pagedSubscribers as subscriber (subscriber.id)}                                            <div class="list-item newsletter-list-item subscriber-row-item" class:is-editing={editingSubscriberId === subscriber.id} class:bulk-selected={isSubscriberSelected(subscriber.id)}>
                                                <div class="subscriber-row-grid">
                                                    <div class="selection-cell subscriber-col-select">
                                                        <input
                                                            type="checkbox"
                                                            checked={isSubscriberSelected(subscriber.id)}
                                                            aria-label={`Select ${subscriber.email}`}
                                                            on:change={() => toggleSubscriberSelection(subscriber.id)}
                                                        />
                                                    </div>

                                                    {#if editingSubscriberId === subscriber.id}
                                                        <div class="subscriber-edit-wrap">
                                                            <div class="subscriber-edit-grid">
                                                                {#if subscribersSupportsNameField}
                                                                    <div class="subscriber-edit-field">
                                                                        <label class="txt-xs txt-hint block m-b-5">Name (optional)</label>
                                                                        <input
                                                                            type="text"
                                                                            class="input input-sm"
                                                                            bind:value={editingSubscriberForm.name}
                                                                            on:input={clearEditingSubscriberError}
                                                                        />
                                                                    </div>
                                                                {/if}
                                                                <div class="subscriber-edit-field">
                                                                    <label class="txt-xs txt-hint block m-b-5">Email</label>
                                                                    <input
                                                                        type="email"
                                                                        class="input input-sm"
                                                                        bind:value={editingSubscriberForm.email}
                                                                        on:input={clearEditingSubscriberError}
                                                                    />
                                                                </div>
                                                                {#if hasSubscriberGroupsFeature}
                                                                    <div class="subscriber-edit-groups">
                                                                        <label class="txt-xs txt-hint block m-b-5">Groups</label>
                                                                        <div class="group-pill-list form-group-pill-list">
                                                                            {#if !subscriberGroups.length}
                                                                                <span class="txt-xs txt-hint">No groups created yet.</span>
                                                                            {:else}
                                                                                {#each subscriberGroups as group (group.id)}
                                                                                    <button
                                                                                        type="button"
                                                                                        class="group-pill-btn"
                                                                                        class:is-selected={hasSelectedGroup(editingSubscriberForm.groupIds, group.id)}
                                                                                        on:click={() => toggleEditingSubscriberGroup(group.id)}
                                                                                    >
                                                                                        {group.name}
                                                                                    </button>
                                                                                {/each}
                                                                            {/if}
                                                                        </div>
                                                                    </div>
                                                                {/if}
                                                                {#if editingSubscriberError}
                                                                    <div class="txt-sm txt-danger">{editingSubscriberError}</div>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    {:else}
                                                        <div class="subscriber-col-name">
                                                            <span class="txt subscriber-primary-label">
                                                                {resolveSubscriberDisplayName(subscriber) || "-"}
                                                            </span>
                                                            {#if subscribersSupportsSourceField && subscriber.source}
                                                                <span class="txt-xs txt-hint subscriber-secondary-line">
                                                                    Source: {resolveSubscriberSourceLabel(subscriber.source)}
                                                                </span>
                                                            {/if}
                                                        </div>
                                                        <div class="subscriber-col-email">
                                                            <span class="txt subscriber-primary-label">{subscriber.email}</span>
                                                        </div>
                                                        <div class="subscriber-col-status">
                                                            <span class={`label label-sm ${getSubscriberStatusLabelClass(subscriber.status)}`}>
                                                                {getSubscriberStatusLabel(subscriber.status)}
                                                            </span>
                                                        </div>
                                                        <div class="subscriber-col-confirmed txt-xs txt-hint">
                                                            {resolveSubscriberConfirmedValue(subscriber)}
                                                        </div>
                                                        <div class="subscriber-col-added txt-xs txt-hint">
                                                            {resolveSubscriberCreatedValue(subscriber)}
                                                        </div>
                                                        <div class="subscriber-col-groups">
                                                            {#if hasSubscriberGroupsFeature}
                                                                <div class="group-pill-list row-group-pill-list">
                                                                    {#if !getSubscriberGroupIds(subscriber).length}
                                                                        <span class="txt-xs txt-hint">No groups</span>
                                                                    {:else}
                                                                        {#each getSubscriberGroupIds(subscriber) as groupId (groupId)}
                                                                            <span class="group-pill-btn row-group-pill-btn is-selected">
                                                                                {subscriberGroupsById.get(groupId)?.name || "Group"}
                                                                            </span>
                                                                        {/each}
                                                                    {/if}
                                                                </div>
                                                            {:else}
                                                                <span class="txt-xs txt-hint">-</span>
                                                            {/if}
                                                        </div>
                                                    {/if}

                                                    <div class="actions subscriber-col-actions">
                                                        {#if editingSubscriberId === subscriber.id}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm action-btn"
                                                                class:btn-loading={isSavingSubscriber}
                                                                disabled={isSavingSubscriber}
                                                                on:click={() => saveSubscriberEdit(subscriber)}
                                                            >
                                                                <span class="txt">Save</span>
                                                            </button>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline action-btn"
                                                                disabled={isSavingSubscriber}
                                                                on:click={cancelEditSubscriber}
                                                            >
                                                                <span class="txt">Cancel</span>
                                                            </button>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-danger btn-outline action-btn"
                                                                class:btn-loading={deletingSubscriberId === subscriber.id}
                                                                disabled={!!deletingSubscriberId}
                                                                on:click={() => openDeleteSubscriberModal(subscriber)}
                                                            >
                                                                <span class="txt">Delete</span>
                                                            </button>
                                                        {:else}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline action-btn"
                                                                disabled={!!editingSubscriberId}
                                                                on:click={() => startEditSubscriber(subscriber)}
                                                            >
                                                                <span class="txt">Edit</span>
                                                            </button>
                                                            {#if normalizeStatus(subscriber.status) !== "active"}
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-sm btn-outline action-btn"
                                                                    disabled={!!editingSubscriberId}
                                                                    on:click={() => setSubscriberStatus(subscriber, "active")}
                                                                >
                                                                    <span class="txt">Activate</span>
                                                                </button>
                                                            {/if}
                                                            {#if normalizeStatus(subscriber.status) !== "unsubscribed"}
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-sm action-btn"
                                                                    disabled={!!editingSubscriberId}
                                                                    on:click={() => setSubscriberStatus(subscriber, "unsubscribed")}
                                                                >
                                                                    <span class="txt">Unsubscribe</span>
                                                                </button>
                                                            {/if}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-danger btn-outline action-btn"
                                                                class:btn-loading={deletingSubscriberId === subscriber.id}
                                                                disabled={!!deletingSubscriberId || !!editingSubscriberId}
                                                                on:click={() => openDeleteSubscriberModal(subscriber)}
                                                            >
                                                                <span class="txt">Delete</span>
                                                            </button>
                                                        {/if}
                                                    </div>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>                                </div>                                {#if filteredSubscribers.length > subscribersPageSize}                                    <div class="pagination-wrap">                                        <button                                            type="button"                                            class="btn btn-xs btn-outline"                                            disabled={subscribersPage <= 1}                                            on:click={() => setSubscribersPage(subscribersPage - 1)}                                        >                                            <span class="txt">Previous</span>                                        </button>                                        <span class="txt-sm txt-hint">                                            Page {subscribersPage} of {subscribersTotalPages}                                        </span>                                        <button                                            type="button"                                            class="btn btn-xs btn-outline"                                            disabled={subscribersPage >= subscribersTotalPages}                                            on:click={() => setSubscribersPage(subscribersPage + 1)}                                        >                                            <span class="txt">Next</span>                                        </button>                                    </div>                                {/if}
                            {/if}

                            {#if selectedSubscribersCount}
                                <div class="subscriber-selection-popover" role="status" aria-live="polite">
                                    <span class="selection-summary">
                                        Selected {selectedSubscribersCount} record(s)
                                    </span>
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-outline"
                                        disabled={isBulkUpdating}
                                        on:click={resetSubscriberSelection}
                                    >
                                        <span class="txt">Reset</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-outline bulk-action-btn"
                                        disabled={!selectedSubscribersCount || isBulkUpdating}
                                        on:click={() => applyBulkStatus("active")}
                                    >
                                        <span class="txt">Mark selected active</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-danger btn-outline bulk-action-btn"
                                        disabled={!selectedSubscribersCount || isBulkUpdating}
                                        on:click={() => applyBulkStatus("unsubscribed")}
                                    >
                                        <span class="txt">Unsubscribe selected</span>
                                    </button>
                                </div>
                            {/if}
                        </section>
                    {:else}
                        <div class="campaign-layout-grid">
                            <section class="panel campaign-composer-panel">
                                <div class="campaign-top-controls m-b-sm">
                                    <div class="tabs-header compact combined left operations-tabs operations-tabs--nested">
                                        <button
                                            type="button"
                                            class="tab-item"
                                            class:active={campaignWorkspace === "builder"}
                                            on:click={() => (campaignWorkspace = "builder")}
                                        >
                                            <i class="ri-tools-line tab-icon" aria-hidden="true" />
                                            Builder
                                        </button>
                                        <button
                                            type="button"
                                            class="tab-item"
                                            class:active={campaignWorkspace === "audience"}
                                            on:click={() => {
                                                campaignWorkspace = "audience";
                                                closeCampaignPreviewModal();
                                            }}
                                        >
                                            <i class="ri-send-plane-line tab-icon" aria-hidden="true" />
                                            Audience & Send
                                        </button>
                                    </div>
                                </div>

                                {#if editingCampaignId}
                                    <div class="campaign-edit-banner m-b-sm">
                                        <span class="txt-sm">
                                            {#if isViewingSentCampaign}
                                                Viewing: <strong>{editingCampaignLabel}</strong> (read-only)
                                                {#if editingCampaign}
                                                    <span class="txt-xs txt-hint"> | {resolveCampaignSentMeta(editingCampaign)} | {resolveCampaignDeliveredSummary(editingCampaign)}</span>
                                                {/if}
                                            {:else}
                                                Editing: <strong>{editingCampaignLabel}</strong>
                                            {/if}
                                        </span>
                                        {#if isViewingSentCampaign}
                                            <button
                                                type="button"
                                                class="btn btn-xs btn-outline"
                                                class:btn-loading={editingCampaign?.id && isDuplicatingCampaign[editingCampaign.id]}
                                                disabled={!editingCampaign?.id || !!isDuplicatingCampaign[editingCampaign.id]}
                                                on:click={() => duplicateCampaignAsDraft(editingCampaign)}
                                            >
                                                <span class="txt">Duplicate</span>
                                            </button>
                                        {:else}
                                            <button type="button" class="btn btn-xs btn-outline" on:click={resetCampaignComposer}>
                                                <span class="txt">New draft</span>
                                            </button>
                                        {/if}
                                    </div>
                                {/if}

                                {#if campaignWorkspace === "builder"}
                                    <div class="campaign-head-row m-b-sm">
                                        <div class="campaign-head-inline campaign-head-inline--single-line">
                                            <h4 class="m-0">Campaign Builder</h4>
                                            <div class="campaign-step-label txt-xs txt-hint">Step 1 of 2</div>
                                            <p class="txt-sm txt-hint m-b-0 campaign-head-description">
                                                {resolveBuilderStepHint()}
                                            </p>
                                        </div>
                                        <div class="tabs-header compact combined left operations-tabs operations-tabs--nested campaign-builder-view-tabs">
                                            <button
                                                type="button"
                                                class="tab-item"
                                                class:active={campaignBuilderShowEditor}
                                                aria-pressed={campaignBuilderShowEditor}
                                                on:click={() => toggleCampaignBuilderPanel("edit")}
                                            >
                                                <i class="ri-edit-line tab-icon" aria-hidden="true" />
                                                Edit
                                            </button>
                                            <button
                                                type="button"
                                                class="tab-item"
                                                class:active={campaignBuilderShowPreview}
                                                aria-pressed={campaignBuilderShowPreview}
                                                on:click={() => toggleCampaignBuilderPanel("preview")}
                                            >
                                                <i class="ri-eye-line tab-icon" aria-hidden="true" />
                                                Preview
                                            </button>
                                        </div>
                                    </div>

                                    <div
                                        class="campaign-builder-layout"
                                        class:is-split={campaignBuilderShowEditor && campaignBuilderShowPreview}
                                        class:is-edit-only={campaignBuilderShowEditor && !campaignBuilderShowPreview}
                                        class:is-preview-only={!campaignBuilderShowEditor && campaignBuilderShowPreview}
                                    >
                                        {#if campaignBuilderShowEditor}
                                        <div class="campaign-builder-editor">
                                            <label class="txt-sm txt-hint block m-b-5" for="campaign-subject">Subject</label>
                                            <input
                                                id="campaign-subject"
                                                type="text"
                                                class="input"
                                                placeholder="Newsletter subject..."
                                                bind:value={campaignForm.subject}
                                                disabled={isViewingSentCampaign}
                                                on:input={clearCampaignFormError}
                                            />
                                            {#if shouldShowCampaignSubjectValidation}
                                                <div class="txt-xs txt-danger m-t-5">Subject is required.</div>
                                            {/if}

                                            <label class="txt-sm txt-hint block m-b-5 m-t-sm" for="campaign-body-editor">Body</label>
                                            <div class="campaign-body-editor">
                                                <TinyMCE
                                                    id="campaign-body-editor"
                                                    conf={campaignBodyEditorConfig}
                                                    disabled={isViewingSentCampaign}
                                                    bind:value={campaignForm.body}
                                                    on:change={clearCampaignFormError}
                                                    on:input={clearCampaignFormError}
                                                />
                                            </div>
                                            {#if shouldShowCampaignBodyValidation}
                                                <div class="txt-xs txt-danger m-t-5">Body is required.</div>
                                            {/if}

                                            {#if campaignFormError}
                                                <div class="txt-sm txt-danger m-t-sm">{campaignFormError}</div>
                                            {/if}

                                            <div class="campaign-builder-footer m-t-sm">
                                                <span class="txt-sm txt-hint">
                                                    {#if !campaignSubjectValue || !campaignBodyValue}
                                                        Complete subject and body to continue to audience.
                                                    {:else}
                                                        Ready to continue to audience and recipient setup.
                                                    {/if}
                                                </span>
                                                <button
                                                    type="button"
                                                    class="btn action-btn campaign-builder-cta"
                                                    disabled={!campaignSubjectValue || !campaignBodyValue}
                                                    on:click={() => {
                                                        campaignWorkspace = "audience";
                                                        closeCampaignPreviewModal();
                                                    }}
                                                >
                                                    <span class="txt">Continue to Audience</span>
                                                </button>
                                            </div>
                                        </div>
                                        {/if}

                                        {#if campaignBuilderShowPreview}
                                        <div class="campaign-builder-preview-side">
                                            <div class="campaign-preview-header m-b-xs">
                                                <h5 class="m-0">Preview</h5>
                                                <div class="campaign-preview-header-actions">
                                                    <div class="tabs-header compact combined left operations-tabs campaign-preview-device-tabs">
                                                        {#each campaignPreviewModes as mode (mode.key)}
                                                            <button
                                                                type="button"
                                                                class="tab-item"
                                                                class:active={campaignPreviewMode === mode.key}
                                                                aria-pressed={campaignPreviewMode === mode.key}
                                                                on:click={() => setCampaignPreviewMode(mode.key)}
                                                            >
                                                                <i class={`${mode.icon} tab-icon`} aria-hidden="true" />
                                                                <span class="tab-label">{mode.label}</span>
                                                            </button>
                                                        {/each}
                                                    </div>
                                                    <button
                                                        type="button"
                                                        class="btn btn-xs btn-outline campaign-preview-expand-btn"
                                                        on:click={openCampaignPreviewModal}
                                                    >
                                                        <span class="txt">Expand</span>
                                                    </button>
                                                </div>
                                            </div>

                                            <div class="campaign-preview-box campaign-preview-html">
                                                {#if campaignPreviewHasBody}
                                                    <div class="campaign-preview-canvas">
                                                        <div class="campaign-preview-frame-shell" style={campaignPreviewFrameStyle}>
                                                            <iframe
                                                                class="campaign-preview-iframe"
                                                                title="Campaign preview"
                                                                sandbox="allow-same-origin"
                                                                srcdoc={campaignPreviewDocumentHtml}
                                                            />
                                                        </div>
                                                    </div>
                                                {:else}
                                                    <p class="txt-sm txt-hint m-0">Body preview will appear here.</p>
                                                {/if}
                                            </div>
                                        </div>
                                        {/if}
                                    </div>
                                {:else}
                                    <div class="campaign-head-row m-b-sm">
                                        <div class="campaign-head-inline campaign-head-inline--single-line">
                                            <h4 class="m-0">Audience & Send</h4>
                                            <div class="campaign-step-label txt-xs txt-hint">Step 2 of 2</div>
                                            <p class="txt-sm txt-hint m-b-0 campaign-head-description">
                                                {resolveAudienceStepHint()}
                                            </p>
                                        </div>
                                    </div>

                                    <div class="campaign-audience-workspace">
                                        <form id="campaign-audience-form" class="campaign-audience-form" on:submit|preventDefault={saveCampaignDraftFromComposer}>
                                            <section class="campaign-audience-main-section manual-recipients">
                                                <div class="manual-recipients-head section-head-inline">
                                                    <h5 class="m-0">Recipients</h5>
                                                    <span class="txt-sm txt-hint m-b-0 manual-section-helper">
                                                        Choose recipients manually. Group chips are shortcuts for manual selection.
                                                    </span>
                                                </div>

                                                <div class="campaign-audience-toolbar">
                                                    <div class="campaign-audience-toolbar-row campaign-audience-toolbar-row--filters">
                                                        <label class="campaign-audience-toolbar-field campaign-audience-toolbar-field--search">
                                                            <span class="txt-xs txt-hint">Search subscribers</span>
                                                            <input
                                                                class="form-input"
                                                                type="search"
                                                                placeholder="Search by name, email, or group..."
                                                                bind:value={audienceRecipientSearch}
                                                                disabled={isViewingSentCampaign}
                                                            />
                                                        </label>
                                                        <label class="campaign-audience-toolbar-field campaign-audience-toolbar-field--compact">
                                                            <span class="txt-xs txt-hint">View</span>
                                                            <select class="form-select" bind:value={audienceRecipientVisibilityFilter} disabled={isViewingSentCampaign}>
                                                                <option value="all">All</option>
                                                                <option value="selected">Selected only</option>
                                                                <option value="unselected">Unselected only</option>
                                                            </select>
                                                        </label>
                                                    </div>
                                                    <div class="campaign-audience-toolbar-row campaign-audience-toolbar-row--actions">
                                                        <div class="campaign-audience-toolbar-actions-row">
                                                            <span class="manual-recipients-count txt-xs">{audienceRecipientsSummary}</span>
                                                            <button type="button" class="btn btn-xs btn-outline" disabled={isViewingSentCampaign} on:click={selectAllActiveRecipients}>
                                                                <span class="txt">Select all active</span>
                                                            </button>
                                                            <button type="button" class="btn btn-xs btn-outline" disabled={isViewingSentCampaign} on:click={clearManualRecipients}>
                                                                <span class="txt">Clear selection</span>
                                                            </button>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div class="manual-group-tools">
                                                    <div class="section-head-inline manual-group-head">
                                                        <h6 class="m-0">Groups</h6>
                                                        <span class="txt-sm txt-hint m-b-0 manual-section-helper">Choose groups to add their active subscribers to this campaign.</span>
                                                    </div>
                                                    {#if hasSubscriberGroupsFeature && subscriberGroups.length}
                                                        <div class="manual-group-chip-list">
                                                            {#each subscriberGroups as group (group.id)}
                                                                {@const groupSelectionMeta = getGroupSelectionMeta(group.id)}
                                                                <button
                                                                    type="button"
                                                                class="manual-group-chip"
                                                                class:is-active={groupSelectionMeta.state === "full"}
                                                                class:is-partial={groupSelectionMeta.state === "partial"}
                                                                disabled={isViewingSentCampaign}
                                                                on:click={() => toggleGroupRecipients(group.id)}
                                                            >
                                                                    <span class="manual-group-chip-name">{group.name}</span>
                                                                    <span class="manual-group-chip-count">
                                                                        {#if groupSelectionMeta.state === "none"}
                                                                            {groupSelectionMeta.totalCount}
                                                                        {:else}
                                                                            {groupSelectionMeta.selectedCount}/{groupSelectionMeta.totalCount}
                                                                        {/if}
                                                                    </span>
                                                                </button>
                                                            {/each}
                                                        </div>
                                                    {:else if hasSubscriberGroupsFeature}
                                                        <p class="txt-xs txt-hint m-b-0">No groups available yet for this website.</p>
                                                    {:else}
                                                        <p class="txt-xs txt-hint m-b-0">Group shortcuts are not available for this website.</p>
                                                    {/if}
                                                </div>

                                                {#if !activeSubscribers.length}
                                                    <div class="empty-state">No active subscribers available for this audience.</div>
                                                {:else if !filteredAudienceRecipients.length}
                                                    <div class="empty-state">No recipients match the current audience filters.</div>
                                                {:else}
                                                    <div class="manual-recipients-tiles">
                                                        {#each visibleAudienceRecipients as subscriber (subscriber.id)}
                                                            {@const subscriberGroupNames = hasSubscriberGroupsFeature
                                                                ? getSubscriberGroupIds(subscriber)
                                                                    .map((groupId) => subscriberGroupsById.get(groupId)?.name)
                                                                    .filter(Boolean)
                                                                : []}
                                                            <label
                                                                class="manual-recipient-tile"
                                                                class:is-selected={normalizedCampaignManualRecipientIdsSet.has(subscriber.id)}
                                                            >
                                                                <input
                                                                    type="checkbox"
                                                                    class="manual-recipient-check"
                                                                    checked={normalizedCampaignManualRecipientIdsSet.has(subscriber.id)}
                                                                    disabled={isViewingSentCampaign}
                                                                    on:change={() => toggleManualRecipient(subscriber.id)}
                                                                />
                                                                <span class="manual-recipient-content">
                                                                    {#if resolveSubscriberDisplayName(subscriber)}
                                                                        <span class="manual-recipient-title">{resolveSubscriberDisplayName(subscriber)}</span>
                                                                        <span class="manual-recipient-subtitle">{subscriber.email}</span>
                                                                    {:else}
                                                                        <span class="manual-recipient-title">{subscriber.email}</span>
                                                                    {/if}
                                                                    {#if hasSubscriberGroupsFeature}
                                                                        <span class="manual-recipient-groups">
                                                                            {#if subscriberGroupNames.length}
                                                                                {#each subscriberGroupNames as groupName}
                                                                                    <span class="label label-sm manual-recipient-group-pill">{groupName}</span>
                                                                                {/each}
                                                                            {:else}
                                                                                <span class="txt-xs txt-hint manual-recipient-no-groups">No groups</span>
                                                                            {/if}
                                                                        </span>
                                                                    {/if}
                                                                </span>
                                                            </label>
                                                        {/each}
                                                    </div>
                                                    <div class="manual-recipients-footer">
                                                        {#if canLoadMoreAudienceRecipients}
                                                            <span class="txt-sm txt-hint">Showing {visibleAudienceRecipientsCount} of {filteredAudienceRecipients.length} recipients</span>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline"
                                                                on:click={loadMoreAudienceRecipients}
                                                            >
                                                                <span class="txt">Load more</span>
                                                            </button>
                                                        {:else}
                                                            <span class="txt-sm txt-hint">Showing all {visibleAudienceRecipientsCount} recipients</span>
                                                        {/if}
                                                    </div>
                                                {/if}
                                            </section>

                                            {#if campaignFormError}
                                                <div class="txt-sm txt-danger">{campaignFormError}</div>
                                            {/if}
                                        </form>

                                        <aside class="campaign-audience-side">
                                            <section class="campaign-audience-side-section">
                                                <h5 class="m-0">Campaign summary</h5>
                                                <div class="campaign-audience-summary-list">
                                                    <div class="campaign-audience-summary-row">
                                                        <span class="audience-stat-label">Draft</span>
                                                        <span class="audience-stat-value">{campaignSubjectValue || "Untitled draft"}</span>
                                                    </div>
                                                    <div class="campaign-audience-summary-row">
                                                        <span class="audience-stat-label">Status</span>
                                                        <span class="audience-stat-value">{audienceSummaryStatus}</span>
                                                    </div>
                                                    <div class="campaign-audience-summary-row">
                                                        <span class="audience-stat-label">Recipients</span>
                                                        <span class="audience-stat-value">{audienceRecipientsSummary}</span>
                                                    </div>
                                                    <div class="campaign-audience-summary-row">
                                                        <span class="audience-stat-label">Delivery</span>
                                                        <span class="audience-stat-value">Campaigns are sent only to active subscribers.</span>
                                                    </div>
                                                    <div class="campaign-audience-summary-row">
                                                        <span class="audience-stat-label">Groups available</span>
                                                        <span class="audience-stat-value">
                                                            {hasSubscriberGroupsFeature ? `${subscriberGroups.length} available` : "Not enabled"}
                                                        </span>
                                                    </div>
                                                    {#if isViewingSentCampaign && editingCampaign}
                                                        <div class="campaign-audience-summary-row">
                                                            <span class="audience-stat-label">Sent at</span>
                                                            <span class="audience-stat-value">{resolveCampaignSentDate(editingCampaign)}</span>
                                                        </div>
                                                        <div class="campaign-audience-summary-row">
                                                            <span class="audience-stat-label">Delivered</span>
                                                            <span class="audience-stat-value">{resolveCampaignDeliveredSummary(editingCampaign)}</span>
                                                        </div>
                                                        <div class="campaign-audience-summary-row">
                                                            <span class="audience-stat-label">Audience</span>
                                                            <span class="audience-stat-value">{resolveCampaignAudienceSummary(editingCampaign)}</span>
                                                        </div>
                                                    {/if}
                                                </div>
                                            </section>

                                            <section class="campaign-audience-side-section campaign-audience-actions-panel">
                                                <h5 class="m-0">Actions</h5>
                                                <div class="campaign-audience-actions">
                                                    {#if isViewingSentCampaign}
                                                        <p class="txt-sm txt-hint m-b-0">Sent campaigns are kept as history. Duplicate to edit and send again.</p>
                                                        <button
                                                            type="button"
                                                            class="btn btn-sm btn-outline action-btn"
                                                            class:btn-loading={editingCampaign?.id && isDuplicatingCampaign[editingCampaign.id]}
                                                            disabled={!editingCampaign?.id || !!isDuplicatingCampaign[editingCampaign.id]}
                                                            on:click={() => duplicateCampaignAsDraft(editingCampaign)}
                                                        >
                                                            <span class="txt">Duplicate</span>
                                                        </button>
                                                        <button type="button" class="btn btn-sm btn-outline action-btn" on:click={resetCampaignComposer}>
                                                            <span class="txt">Back to drafts</span>
                                                        </button>
                                                    {:else}
                                                        {#if editingCampaignId}
                                                            <button type="button" class="btn btn-sm btn-outline action-btn" on:click={resetCampaignComposer}>
                                                                <span class="txt">Cancel Edit</span>
                                                            </button>
                                                        {/if}
                                                        <button
                                                            type="submit"
                                                            form="campaign-audience-form"
                                                            class="btn btn-sm action-btn"
                                                            class:btn-loading={isCreatingCampaign || isSavingCampaign}
                                                            disabled={!!createCampaignDisabledReason || isCreatingCampaign || isSavingCampaign}
                                                            title={createCampaignDisabledReason || null}
                                                        >
                                                            <span class="txt">{editingCampaignId ? "Update draft" : "Create draft"}</span>
                                                        </button>
                                                        <button
                                                            type="button"
                                                            class="btn btn-sm action-btn"
                                                            class:btn-loading={editingCampaign?.id && isSendingCampaign[editingCampaign.id]}
                                                            disabled={!!audienceSendDisabledReason}
                                                            title={audienceSendDisabledReason || null}
                                                            on:click={() => openSendCampaignModal(editingCampaign)}
                                                        >
                                                            <span class="txt">Send campaign</span>
                                                        </button>
                                                    {/if}
                                                </div>
                                            </section>

                                            <section class="campaign-audience-side-section audience-health-panel">
                                                <div class="campaign-audience-health-head">
                                                    <div class="campaign-audience-health-main">
                                                        <h5 class="m-0">Audience health</h5>
                                                        <p class="txt-sm txt-hint m-b-0 campaign-audience-health-helper">
                                                            Review what is missing before sending.
                                                        </p>
                                                    </div>
                                                    <div class="campaign-audience-health-meta">
                                                        <span class={`label label-sm audience-health-status-pill ${audienceHealthPillClass}`}>
                                                            {audienceHealthStatus}
                                                        </span>
                                                        <span class="summary-pill audience-health-summary-pill" class:warning={audienceWarnings.length > 0}>
                                                            {audienceWarnings.length} warning{audienceWarnings.length === 1 ? "" : "s"} | {audienceSuggestions.length} suggestion{audienceSuggestions.length === 1 ? "" : "s"}
                                                        </span>
                                                    </div>
                                                </div>

                                                {#if audienceWarnings.length}
                                                    <div class="audience-health-group">
                                                        <div class="audience-health-group-title">Warnings</div>
                                                        <div class="audience-health-check-list">
                                                            {#each audienceWarnings as warning}
                                                                <div class="audience-health-check-item warning">
                                                                    <span class="label label-sm audience-health-check-pill warning">Warning</span>
                                                                    <span class="audience-health-check-message">{warning}</span>
                                                                </div>
                                                            {/each}
                                                        </div>
                                                    </div>
                                                {/if}

                                                {#if audienceSuggestions.length}
                                                    <div class="audience-health-group">
                                                        <div class="audience-health-group-title">Suggestions</div>
                                                        <div class="audience-health-check-list">
                                                            {#each audienceSuggestions as suggestion}
                                                                <div class="audience-health-check-item">
                                                                    <span class="label label-sm audience-health-check-pill">Info</span>
                                                                    <span class="audience-health-check-message">{suggestion}</span>
                                                                </div>
                                                            {/each}
                                                        </div>
                                                    </div>
                                                {/if}

                                                {#if !audienceWarnings.length && !audienceSuggestions.length}
                                                    <p class="txt-sm txt-hint m-b-0">No audience issues found for this campaign.</p>
                                                {/if}
                                            </section>
                                        </aside>
                                    </div>
                                {/if}
                            </section>

                            <section class="panel campaign-list-panel">
                                <div class="campaigns-header-row m-b-sm">
                                    <div class="section-head m-b-0">
                                        <h4 class="m-0">Campaigns</h4>
                                        <p class="txt-sm txt-hint m-b-0">{resolveCampaignsSectionHint()}</p>
                                    </div>
                                    <div class="campaign-list-header-actions">
                                        <span class="txt-sm txt-hint">{resolveCampaignListStats()}</span>
                                    </div>
                                </div>

                                <div class="campaign-filter-chips m-b-sm" role="toolbar" aria-label="Filter campaigns">
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline campaign-filter-chip"
                                        class:is-active={campaignStatusFilter === "all"}
                                        aria-pressed={campaignStatusFilter === "all"}
                                        on:click={() => (campaignStatusFilter = "all")}
                                    >
                                        <span class="txt">All ({campaigns.length})</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline campaign-filter-chip"
                                        class:is-active={campaignStatusFilter === "draft"}
                                        aria-pressed={campaignStatusFilter === "draft"}
                                        on:click={() => (campaignStatusFilter = "draft")}
                                    >
                                        <span class="txt">Draft ({draftCampaigns.length})</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline campaign-filter-chip"
                                        class:is-active={campaignStatusFilter === "sent"}
                                        aria-pressed={campaignStatusFilter === "sent"}
                                        on:click={() => (campaignStatusFilter = "sent")}
                                    >
                                        <span class="txt">Sent ({sentCampaigns.length})</span>
                                    </button>
                                </div>

                                {#if isLoadingCampaigns}
                                    <div class="loading-state">
                                        <span class="loader loader-sm" />
                                        <span class="txt-hint">Loading campaigns...</span>
                                    </div>
                                {:else if !campaigns.length}
                                    <div class="empty-state empty-state-stack">
                                        <span>No campaigns yet for this website.</span>
                                        <span class="txt-sm txt-hint">{resolveCampaignNoItemsHint()}</span>
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-outline"
                                            on:click={() => {
                                                campaignWorkspace = "builder";
                                                campaignBuilderShowEditor = true;
                                                campaignBuilderShowPreview = false;
                                            }}
                                        >
                                            <span class="txt">Start draft</span>
                                        </button>
                                    </div>
                                {:else if !filteredCampaigns.length}
                                    <div class="empty-state empty-state-stack">
                                        <span>No campaigns match this filter.</span>
                                        <span class="txt-sm txt-hint">{resolveCampaignFilterEmptyHint()}</span>
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-outline"
                                            on:click={() => (campaignStatusFilter = "all")}
                                        >
                                            <span class="txt">Show all campaigns</span>
                                        </button>
                                    </div>
                                {:else}
                                    <div class="list list-compact">
                                        <div class="list-content campaign-list-scroll">
                                            {#each filteredCampaigns as campaign (campaign.id)}
                                                <div
                                                    class="list-item newsletter-list-item campaign-row-item"
                                                    class:is-editing={editingCampaignId === campaign.id}
                                                >
                                                    <div class="content campaign-row-content">
                                                        <div class="campaign-row-top">
                                                            <span class="txt campaign-row-subject">{resolveCampaignSubject(campaign)}</span>
                                                            <span class={`label label-sm ${resolveCampaignStatusLabelClass(campaign.status)}`}>
                                                                {resolveCampaignStatusLabel(campaign.status)}
                                                            </span>
                                                        </div>
                                                        <div class="txt-xs txt-hint meta-line campaign-row-meta">
                                                            <span><strong>Audience:</strong> {resolveCampaignAudienceLabel(campaign)}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>Est. recipients:</strong> {resolveCampaignRecipientsSummary(campaign)}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>Delivered:</strong> {resolveCampaignDeliveredSummary(campaign)}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>{normalizeStatus(campaign.status) === "sent" ? "Sent" : "Updated"}:</strong> {normalizeStatus(campaign.status) === "sent" ? resolveCampaignSentDate(campaign) : resolveCampaignUpdatedDate(campaign)}</span>
                                                        </div>
                                                    </div>

                                                    <div class="actions campaign-row-actions">
                                                        {#if normalizeStatus(campaign.status) === "draft"}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline action-btn campaign-row-btn"
                                                                on:click={() => startEditCampaign(campaign)}
                                                            >
                                                                <span class="txt">{editingCampaignId === campaign.id ? "Editing" : "Edit"}</span>
                                                            </button>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm action-btn campaign-row-btn"
                                                                class:btn-loading={isSendingCampaign[campaign.id]}
                                                                disabled={!!getSendCampaignDisabledReason(campaign)}
                                                                title={getSendCampaignDisabledReason(campaign) || null}
                                                                on:click={() => openSendCampaignModal(campaign)}
                                                            >
                                                                <span class="txt">Send</span>
                                                            </button>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-danger btn-outline action-btn campaign-row-btn"
                                                                class:btn-loading={deletingCampaignId === campaign.id}
                                                                disabled={deletingCampaignId === campaign.id}
                                                                on:click={() => openDeleteCampaignModal(campaign)}
                                                            >
                                                                <span class="txt">Delete</span>
                                                            </button>
                                                        {:else if normalizeStatus(campaign.status) === "sent"}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline action-btn campaign-row-btn"
                                                                on:click={() => startViewCampaign(campaign)}
                                                            >
                                                                <span class="txt">{isViewingSentCampaign && editingCampaignId === campaign.id ? "Viewing" : "View"}</span>
                                                            </button>
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-outline action-btn campaign-row-btn"
                                                                class:btn-loading={isDuplicatingCampaign[campaign.id]}
                                                                disabled={!!isDuplicatingCampaign[campaign.id]}
                                                                on:click={() => duplicateCampaignAsDraft(campaign)}
                                                            >
                                                                <span class="txt">Duplicate</span>
                                                            </button>
                                                        {/if}
                                                    </div>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}
                            </section>
                        </div>
                    {/if}
                </div>
            </div>
        {/if}

        {#if isCampaignPreviewExpanded}
            <OverlayPanel
                popup
                class="newsletter-campaign-preview"
                active={true}
                overlayClose={true}
                escClose={false}
                btnClose={false}
                on:hide={closeCampaignPreviewModal}
            >
                <div slot="header" class="campaign-preview-modal-header">
                    <div class="campaign-preview-modal-title">
                        <h4 class="m-0">Campaign Preview</h4>
                        <p class="txt-sm txt-hint m-b-0">Previewing current unsaved editor content.</p>
                    </div>
                    <div class="campaign-preview-modal-actions">
                        <div class="tabs-header compact combined left operations-tabs campaign-preview-device-tabs campaign-preview-device-tabs--modal">
                            {#each campaignPreviewModes as mode (mode.key)}
                                <button
                                    type="button"
                                    class="tab-item"
                                    class:active={campaignPreviewMode === mode.key}
                                    aria-pressed={campaignPreviewMode === mode.key}
                                    on:click={() => setCampaignPreviewMode(mode.key)}
                                >
                                    <i class={`${mode.icon} tab-icon`} aria-hidden="true" />
                                    <span class="tab-label">{mode.label}</span>
                                </button>
                            {/each}
                        </div>
                        <button
                            type="button"
                            class="btn btn-xs btn-outline campaign-preview-modal-close-btn"
                            on:click={closeCampaignPreviewModal}
                        >
                            <span class="txt">Close</span>
                        </button>
                    </div>
                </div>
                <div class="campaign-preview-box campaign-preview-html campaign-preview-modal-box">
                    <div class="campaign-preview-modal-meta txt-xs txt-hint">
                        <span>{activeCampaignPreviewMode.label} preview</span>
                        <span aria-hidden="true" class="meta-sep">|</span>
                        <span>{campaignPreviewFrameWidth}px frame</span>
                    </div>
                    {#if campaignPreviewHasBody}
                        <div class="campaign-preview-canvas campaign-preview-canvas--modal">
                            <div class="campaign-preview-frame-shell campaign-preview-frame-shell--modal" style={campaignPreviewFrameStyle}>
                                <iframe
                                    class="campaign-preview-iframe campaign-preview-iframe--modal"
                                    title="Expanded campaign preview"
                                    sandbox="allow-same-origin"
                                    srcdoc={campaignPreviewDocumentHtml}
                                />
                            </div>
                        </div>
                    {:else}
                        <div class="campaign-preview-empty-state">
                            <p class="txt-sm txt-hint m-0">Body preview will appear here.</p>
                        </div>
                    {/if}
                </div>
            </OverlayPanel>
        {/if}

        {#if pendingSendCampaign}
            <OverlayPanel
                popup
                class="newsletter-send-confirm hide-content overlay-panel-sm"
                active={true}
                overlayClose={true}
                escClose={false}
                btnClose={false}
                on:hide={closeSendCampaignModal}
            >
                <div slot="header" class="newsletter-send-confirm-head">
                    <h4 class="m-0">Send campaign now?</h4>
                </div>

                <p class="txt-sm txt-hint m-0">
                    <strong>{pendingSendCampaign.subject}</strong> will be sent to approximately
                    <strong> {pendingSendRecipientsCount}</strong> recipient(s).
                </p>
                <p class="txt-xs txt-hint m-t-xs m-b-0">{resolveSendConfirmationHint()}</p>

                <svelte:fragment slot="footer">
                    <button type="button" class="btn btn-sm btn-outline" on:click={closeSendCampaignModal}>
                        <span class="txt">Cancel</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm"
                        class:btn-loading={isSendingCampaign[pendingSendCampaign.id]}
                        disabled={!!isSendingCampaign[pendingSendCampaign.id]}
                        on:click={confirmSendCampaign}
                    >
                        <span class="txt">Confirm send</span>
                    </button>
                </svelte:fragment>
            </OverlayPanel>
        {/if}

        {#if pendingDeleteSubscriber}
            <OverlayPanel
                popup
                class="newsletter-delete-confirm hide-content overlay-panel-sm"
                active={true}
                overlayClose={!deletingSubscriberId}
                escClose={false}
                btnClose={false}
                on:hide={closeDeleteSubscriberModal}
            >
                <div slot="header" class="newsletter-delete-confirm-head">
                    <h4 class="m-0">Delete subscriber?</h4>
                    <p class="txt-sm txt-hint m-t-5 m-b-0">
                        <strong>{pendingDeleteSubscriber.email}</strong> will be permanently removed.
                    </p>
                </div>

                <svelte:fragment slot="footer">
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!!deletingSubscriberId}
                        on:click={closeDeleteSubscriberModal}
                    >
                        <span class="txt">Cancel</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-danger btn-outline"
                        class:btn-loading={!!deletingSubscriberId}
                        disabled={!!deletingSubscriberId}
                        on:click={confirmDeleteSubscriber}
                    >
                        <span class="txt">Delete</span>
                    </button>
                </svelte:fragment>
            </OverlayPanel>
        {/if}
        {#if pendingDeleteCampaign}
            <OverlayPanel
                popup
                class="newsletter-delete-confirm hide-content overlay-panel-sm"
                active={true}
                overlayClose={!deletingCampaignId}
                escClose={false}
                btnClose={false}
                on:hide={closeDeleteCampaignModal}
            >
                <div slot="header" class="newsletter-delete-confirm-head">
                    <h4 class="m-0">Delete draft?</h4>
                    <p class="txt-sm txt-hint m-t-5 m-b-0">
                        This draft will be permanently removed. This action cannot be undone.
                    </p>
                </div>

                <svelte:fragment slot="footer">
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        disabled={!!deletingCampaignId}
                        on:click={closeDeleteCampaignModal}
                    >
                        <span class="txt">Cancel</span>
                    </button>
                    <button
                        type="button"
                        class="btn btn-sm btn-danger btn-outline"
                        class:btn-loading={!!deletingCampaignId}
                        disabled={!!deletingCampaignId}
                        on:click={confirmDeleteCampaign}
                    >
                        <span class="txt">Delete draft</span>
                    </button>
                </svelte:fragment>
            </OverlayPanel>
        {/if}
    {/if}
</PageWrapper>
<style>
    .newsletter-head.operations-head .head-description {
        max-width: 460px;
    }

    .subscribers-section-panel {
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .subscriber-controls {
        display: flex;
        align-items: flex-end;
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: wrap;
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 8px;
    }

    .subscribers-panel-header {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .subscribers-panel-header-actions {
        display: inline-flex;
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
        justify-content: flex-end;
    }

    .add-form-toggle-btn {
        min-width: 118px;
        justify-content: center;
        gap: 6px;
    }

    .subscriber-create-form {
        display: flex;
        flex-direction: column;
        gap: 9px;
        width: 100%;
    }

    .subscriber-inline-create {
        border: 0;
        border-radius: 0;
        background: transparent;
        padding: 0;
    }

    .subscriber-create-row {
        display: grid;
        gap: 8px 10px;
        align-items: end;
    }

    .subscriber-create-row-primary {
        grid-template-columns: minmax(200px, 1fr) minmax(260px, 1fr) minmax(140px, 170px);
    }

    .subscriber-create-row-primary.no-name {
        grid-template-columns: minmax(340px, 1fr) minmax(140px, 170px);
    }

    .subscriber-create-row-groups {
        grid-template-columns: minmax(0, 1fr) minmax(180px, 260px) auto;
        padding-top: 4px;
    }

    .create-name-field,
    .create-email-field,
    .create-action-field {
        min-width: 0;
    }

    .create-name-field .input,
    .create-email-field .input {
        width: 100%;
    }

    .create-name-input {
        margin-bottom: 0;
    }

    .create-action-field {
        display: flex;
        align-items: flex-end;
        justify-content: flex-end;
    }

    .subscriber-groups-select,
    .subscriber-groups-input,
    .subscriber-groups-action {
        min-width: 0;
    }

    .subscriber-groups-action {
        display: flex;
        justify-content: flex-end;
    }

    .subscriber-groups-action .btn {
        min-height: var(--inputHeight);
        white-space: nowrap;
        min-width: 102px;
    }

    .subscriber-filter-grid {
        display: grid;
        gap: 7px;
        grid-template-columns: repeat(3, minmax(170px, 1fr));
        flex: 1 1 620px;
        min-width: 260px;
    }

    .subscriber-filter-grid.group-enabled {
        grid-template-columns: repeat(4, minmax(150px, 1fr));
    }

    .control-item {
        min-width: 0;
    }

    .group-pill-list {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        min-height: var(--smBtnHeight);
    }

    .group-pill-btn {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseAlt1Color);
        color: var(--txtHintColor);
        font-size: 11px;
        line-height: 1;
        padding: 5px 9px;
        cursor: pointer;
        transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;
    }

    .group-pill-btn:hover {
        background: var(--baseColor);
        border-color: color-mix(in srgb, var(--primaryColor) 32%, var(--baseAlt2Color));
        color: var(--txtPrimaryColor);
    }

    .group-pill-btn.is-selected {
        border-color: color-mix(in srgb, var(--primaryColor) 50%, transparent);
        background: color-mix(in srgb, var(--primaryColor) 10%, transparent);
        color: var(--txtPrimaryColor);
    }

    .group-pill-btn:disabled {
        opacity: 0.62;
        cursor: default;
    }

    .form-group-pill-list {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: transparent;
        padding: 5px 6px;
        min-height: var(--inputHeight);
    }

    .form-group-pill-list .group-pill-btn {
        padding: 5px 10px;
        font-size: 11px;
    }

    .row-group-pill-list {
        margin-top: 7px;
    }

    .row-group-pill-btn {
        padding: 4px 8px;
        font-size: 10px;
    }

    .bulk-action-btn {
        min-width: 145px;
    }

    .subscriber-selection-popover {
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

    .selection-summary {
        color: var(--txtPrimaryColor);
        font-weight: 600;
        font-size: var(--smFontSize);
        padding: 0 3px;
        white-space: nowrap;
    }
    .selection-cell {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding-right: 2px;
    }

    .pagination-wrap {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 8px;
        margin-top: 10px;
    }

    .campaign-preview-box {
        margin: 0;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 12px;
        min-height: 250px;
        max-height: 400px;
        overflow: hidden;
        font-family: inherit;
        font-size: var(--baseFontSize);
        flex: 1 1 auto;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .campaign-preview-html :global(p:last-child) {
        margin-bottom: 0;
    }

    .campaign-preview-canvas {
        width: 100%;
        min-height: 220px;
        padding: 4px;
        display: flex;
        justify-content: center;
        align-items: stretch;
        overflow: auto;
    }

    .campaign-preview-frame-shell {
        width: min(var(--campaign-preview-frame-width, 720px), 100%);
        background: var(--baseColor);
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
        border-radius: 10px;
        box-shadow: 0 10px 24px color-mix(in srgb, var(--txtPrimaryColor) 12%, transparent);
        overflow: hidden;
        flex: 0 0 auto;
    }

    .campaign-preview-iframe {
        width: 100%;
        height: 360px;
        border: 0;
        display: block;
        background: var(--baseColor);
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

    .empty-state-stack {
        flex-direction: column;
    }

    .add-subscriber-btn {
        min-height: var(--inputHeight);
        min-width: 150px;
        width: 100%;
    }
    .campaign-audience-main-section {
        display: flex;
        flex-direction: column;
        gap: 8px;
        min-width: 0;
    }

    .manual-recipients {
        border: 0;
        padding-top: 0;
        gap: 12px;
    }

    .manual-recipients-head,
    .manual-group-head {
        align-items: baseline;
        gap: 8px;
        flex-wrap: wrap;
    }

    .manual-section-helper {
        white-space: normal;
        line-height: 1.35;
    }

    .manual-recipients-count {
        display: inline-flex;
        align-items: center;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        padding: 2px 8px;
        background: var(--baseColor);
        color: var(--txtHintColor);
    }

    .manual-recipients-tiles {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 7px;
    }

    .manual-recipients-footer {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .manual-recipient-tile {
        display: grid;
        grid-template-columns: auto minmax(0, 1fr);
        align-items: flex-start;
        gap: 7px;
        border: 2px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        padding: 8px 10px;
        background: var(--baseColor);
        cursor: pointer;
        min-width: 0;
        position: relative;
        transition: border-color var(--baseAnimationSpeed), background-color var(--baseAnimationSpeed);
    }

    .manual-recipient-tile::before {
        content: "";
        position: absolute;
        left: 0;
        top: 0;
        bottom: 0;
        width: 3px;
        border-radius: var(--baseRadius) 0 0 var(--baseRadius);
        background: color-mix(in srgb, var(--txtPrimaryColor) 38%, transparent);
        opacity: 0;
        transition: opacity 140ms ease;
    }

    .manual-recipient-tile:hover {
        border-color: color-mix(in srgb, var(--txtPrimaryColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 18%, transparent);
    }

    .manual-recipient-tile.is-selected {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 28%, transparent);
    }

    .manual-recipient-tile.is-selected::before {
        opacity: 1;
    }

    .manual-recipient-check {
        margin-top: 1px;
        flex: 0 0 auto;
        width: 15px;
        height: 15px;
        accent-color: var(--primaryColor);
    }

    .manual-recipient-content {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .manual-recipient-title {
        font-size: 13px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-recipient-subtitle {
        font-size: 12px;
        color: color-mix(in srgb, var(--txtHintColor) 92%, var(--txtPrimaryColor));
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-recipient-groups {
        display: inline-flex;
        align-items: center;
        gap: 5px;
        flex-wrap: wrap;
        margin-top: 2px;
    }

    .manual-recipient-group-pill {
        font-size: 11px;
        line-height: 1.25;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 86%, var(--baseColor));
        color: var(--txtHintColor);
    }

    .manual-recipient-tile.is-selected .manual-recipient-group-pill {
        border-color: color-mix(in srgb, var(--primaryColor) 30%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--txtPrimaryColor) 85%, var(--txtHintColor));
        background: color-mix(in srgb, var(--primaryColor) 8%, var(--baseColor));
    }

    .manual-recipient-no-groups {
        display: inline-block;
        padding-top: 1px;
    }

    .manual-recipient-tile input:checked + .manual-recipient-content .manual-recipient-title {
        font-weight: 600;
    }

    .manual-recipient-tile.is-selected .manual-recipient-title {
        font-weight: 600;
    }

    .newsletter-list-item {
        gap: 6px;
        padding: 6px var(--xsSpacing);
    }

    .subscriber-row-item {
        border: 0;
        border-radius: 0;
        background: transparent;
        box-shadow: inset 0 0 0 2px transparent;
        transition: box-shadow var(--baseAnimationSpeed), background-color var(--baseAnimationSpeed);
    }

    .subscriber-row-item:nth-child(odd) {
        background: var(--baseColor);
    }

    .subscriber-row-item:nth-child(even) {
        background: color-mix(in srgb, var(--baseAlt1Color) 78%, var(--baseAlt2Color));
    }

    .subscriber-row-item:hover {
        box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--txtPrimaryColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 18%, transparent);
    }

    .subscriber-row-item.bulk-selected {
        box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--txtPrimaryColor) 55%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1) 22%, transparent);
    }

    .subscriber-row-item.is-editing {
        background: var(--bodyColor);
        box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--txtPrimaryColor) 30%, var(--baseAlt2Color));
    }

    .subscriber-row-item.is-editing:hover,
    .subscriber-row-item.is-editing.bulk-selected {
        background: var(--bodyColor);
        box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--txtPrimaryColor) 30%, var(--baseAlt2Color));
    }

    .subscriber-table-list {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        overflow-x: auto;
        overflow-y: hidden;
    }

    .subscriber-table-head {
        display: grid;
        grid-template-columns: 34px minmax(130px, 0.9fr) minmax(210px, 1.35fr) minmax(110px, 0.8fr) minmax(150px, 0.95fr) minmax(150px, 0.95fr) minmax(170px, 1fr) minmax(340px, 1.35fr);
        gap: 6px;
        align-items: center;
        border-bottom: 1px solid color-mix(in srgb, var(--baseAlt2Color) 92%, transparent);
        background: color-mix(in srgb, var(--baseAlt2Color) 72%, var(--baseAlt1Color));
        padding: 7px var(--xsSpacing);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        font-weight: 700;
        color: color-mix(in srgb, var(--txtPrimaryColor) 84%, var(--txtHintColor));
    }

    .subscriber-row-grid {
        display: grid;
        grid-template-columns: 34px minmax(130px, 0.9fr) minmax(210px, 1.35fr) minmax(110px, 0.8fr) minmax(150px, 0.95fr) minmax(150px, 0.95fr) minmax(170px, 1fr) minmax(340px, 1.35fr);
        gap: 6px;
        align-items: center;
        width: 100%;
    }

    .subscriber-col-select {
        justify-self: center;
    }

    .subscriber-col-select-head {
        justify-self: center;
    }

    .subscriber-col-name,
    .subscriber-col-email,
    .subscriber-col-status,
    .subscriber-col-confirmed,
    .subscriber-col-added,
    .subscriber-col-groups,
    .subscriber-col-actions {
        min-width: 0;
    }

    .subscriber-col-confirmed,
    .subscriber-col-added {
        white-space: nowrap;
    }

    .subscriber-col-email .txt {
        display: inline-block;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .subscriber-primary-label {
        display: inline-block;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .subscriber-secondary-line {
        display: block;
        margin-top: 1px;
    }

    .subscriber-col-groups .group-pill-list {
        margin-top: 0;
    }

    .subscriber-col-actions.actions {
        justify-content: flex-start;
        flex-wrap: wrap;
        gap: 6px;
    }

    .subscriber-col-actions .action-btn {
        min-width: 0;
        min-height: var(--smBtnHeight);
        padding-inline: 10px;
        white-space: nowrap;
        flex: 0 1 auto;
    }

    .subscriber-edit-wrap {
        grid-column: 2 / 8;
    }

    .subscriber-edit-grid {
        display: grid;
        gap: 8px 10px;
        grid-template-columns: minmax(220px, 1fr) minmax(220px, 1fr);
    }

    .subscriber-edit-field {
        min-width: 0;
    }

    .subscriber-edit-groups {
        grid-column: 1 / -1;
    }
    .section-head {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }
    .campaign-head-row {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-top-controls {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .campaign-head-inline {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        row-gap: 4px;
    }

    .campaign-head-inline--single-line {
        flex: 1 1 auto;
        min-width: 0;
        flex-wrap: nowrap;
        align-items: center;
        gap: 10px;
    }

    .campaign-head-inline--single-line h4,
    .campaign-head-inline--single-line .campaign-step-label {
        white-space: nowrap;
        flex: 0 0 auto;
    }

    .campaign-head-inline--single-line .campaign-head-description {
        flex: 1 1 auto;
        min-width: 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .campaign-head-row .campaign-builder-view-tabs {
        margin-left: auto;
    }

    .campaign-head-description {
        flex: 1 1 340px;
        min-width: 220px;
    }

    .campaign-builder-view-tabs .tab-item {
        min-width: 78px;
    }

    .campaign-step-label {
        display: inline-flex;
        align-items: center;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseAlt1Color);
        padding: 1px 8px;
        width: fit-content;
    }

    .campaign-layout-grid {
        display: grid;
        grid-template-columns: minmax(0, 1.7fr) minmax(300px, 0.75fr);
        gap: 12px;
        align-items: start;
    }

    .campaign-composer-panel,
    .campaign-list-panel {
        min-width: 0;
    }

    .campaign-list-header-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        justify-content: flex-end;
    }

    .campaign-list-scroll {
        max-height: 620px;
        overflow: auto;
    }

    .campaign-filter-chips {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .campaign-filter-chip {
        min-height: 28px;
    }

    .campaign-filter-chip.is-active {
        background: var(--baseAlt2Color);
        border-color: var(--baseAlt2Color);
    }

    .campaign-builder-layout {
        display: grid;
        gap: 16px;
        align-items: stretch;
        grid-template-columns: 1fr;
    }

    .campaign-builder-layout.is-split {
        grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
        column-gap: 20px;
    }

    .campaign-builder-layout.is-edit-only,
    .campaign-builder-layout.is-preview-only {
        grid-template-columns: 1fr;
    }

    .campaign-builder-editor,
    .campaign-builder-preview-side {
        min-width: 0;
    }

    .campaign-builder-preview-side {
        border: 0;
        border-radius: 0;
        background: transparent;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 8px;
        min-height: 100%;
    }

    .campaign-builder-layout.is-split .campaign-builder-preview-side {
        border-left: 1px solid var(--baseAlt2Color);
        padding-left: 16px;
    }

    .campaign-preview-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-preview-header-actions {
        display: inline-flex;
        align-items: center;
        justify-content: flex-end;
        gap: 8px;
        flex-wrap: wrap;
        margin-left: auto;
    }

    .campaign-preview-device-tabs .tab-item {
        min-height: 30px;
        min-width: 92px;
        justify-content: center;
        gap: 6px;
    }

    .campaign-preview-device-tabs .tab-label {
        font-size: var(--smFontSize);
    }

    .campaign-preview-expand-btn {
        margin-left: 0;
    }

    .campaign-edit-banner {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 7px 10px;
    }

    .campaign-builder-editor {
        display: flex;
        flex-direction: column;
        min-height: 100%;
    }

    .campaign-builder-footer {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
        justify-content: flex-end;
    }

    .campaign-builder-cta {
        min-height: var(--inputHeight);
        min-width: 178px;
    }

    .campaign-audience-workspace {
        display: grid;
        grid-template-columns: minmax(0, 1fr) clamp(320px, 28vw, 360px);
        gap: 10px;
        align-items: start;
    }

    .campaign-audience-form {
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .campaign-audience-toolbar {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .campaign-audience-toolbar-row {
        display: flex;
        align-items: end;
        gap: 10px;
        min-width: 0;
    }

    .campaign-audience-toolbar-row--filters {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(160px, 220px);
        align-items: end;
    }

    .campaign-audience-toolbar-row--actions {
        justify-content: flex-end;
    }

    .campaign-audience-toolbar-field {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .campaign-audience-toolbar-field--compact {
        max-width: 220px;
    }

    .campaign-audience-toolbar-actions-row {
        display: inline-flex;
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
    }

    .campaign-audience-side {
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .campaign-audience-side-section {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .campaign-audience-summary-list {
        display: grid;
        gap: 6px;
    }

    .campaign-audience-summary-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        min-width: 0;
    }

    .audience-stat-label {
        font-size: 11px;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: var(--txtHintColor);
    }

    .audience-stat-value {
        font-size: 13px;
        color: var(--txtPrimaryColor);
        text-align: right;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .campaign-audience-health-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-audience-health-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .campaign-audience-health-helper {
        font-size: 11px;
        line-height: 1.35;
    }

    .campaign-audience-health-meta {
        display: inline-flex;
        align-items: center;
        justify-content: flex-end;
        gap: 6px;
        flex-wrap: wrap;
    }

    .audience-health-status-pill {
        --labelHPadding: 8px;
        min-height: 20px;
        color: var(--txtHintColor);
        border-color: color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 18%, var(--baseColor));
        font-weight: 600;
    }

    .audience-health-status-pill.label-success {
        color: color-mix(in srgb, var(--successColor) 85%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--successColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--successColor) 12%, var(--baseColor));
    }

    .audience-health-status-pill.label-warning {
        color: color-mix(in srgb, var(--warningColor) 86%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .audience-health-status-pill.label-danger {
        color: color-mix(in srgb, var(--dangerColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--dangerColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--dangerColor) 12%, var(--baseColor));
    }

    .audience-health-summary-pill {
        --labelHPadding: 9px;
        min-height: 20px;
        color: var(--txtHintColor);
        background: color-mix(in srgb, var(--baseAlt1Color) 22%, var(--baseColor));
    }

    .audience-health-summary-pill.warning {
        color: color-mix(in srgb, var(--warningColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .audience-health-group {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .audience-health-group-title {
        color: var(--txtHintColor);
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.02em;
        text-transform: uppercase;
    }

    .audience-health-check-list {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .audience-health-check-item {
        display: flex;
        align-items: flex-start;
        gap: 7px;
        padding: 6px 0;
        font-size: var(--smFontSize);
        line-height: var(--smLineHeight);
        color: var(--txtHintColor);
    }

    .audience-health-check-item + .audience-health-check-item {
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .audience-health-check-item.warning {
        color: color-mix(in srgb, var(--warningColor) 80%, var(--txtPrimaryColor));
    }

    .audience-health-check-message {
        min-width: 0;
    }

    .audience-health-check-pill {
        --labelHPadding: 7px;
        min-height: 18px;
        flex: 0 0 auto;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        color: var(--txtHintColor);
        background: var(--baseColor);
    }

    .audience-health-check-pill.warning {
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--warningColor) 88%, var(--txtPrimaryColor));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .campaign-audience-actions-panel .campaign-audience-actions {
        display: flex;
        flex-direction: column;
        align-items: stretch;
        gap: 8px;
    }

    .campaign-audience-actions .action-btn {
        width: 100%;
        min-width: 0;
        min-height: var(--smBtnHeight);
    }

    .manual-group-tools {
        display: flex;
        flex-direction: column;
        gap: 6px;
        margin: 0;
        padding-top: 2px;
    }

    .manual-group-action-buttons {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .manual-group-chip-list {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
    }

    .manual-group-chip {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseAlt1Color);
        color: var(--txtHintColor);
        min-height: 28px;
        padding: 0 9px;
        transition: border-color 140ms ease, background-color 140ms ease, color 140ms ease;
    }

    .manual-group-chip:hover {
        border-color: color-mix(in srgb, var(--primaryColor) 28%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1Color) 88%, var(--baseColor));
        color: var(--txtPrimaryColor);
    }

    .manual-group-chip.is-active {
        border-color: color-mix(in srgb, var(--primaryColor) 38%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 10%, var(--baseColor));
        color: var(--txtPrimaryColor);
    }

    .manual-group-chip.is-partial {
        border-color: color-mix(in srgb, var(--primaryColor) 32%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 6%, var(--baseColor));
        color: var(--txtPrimaryColor);
    }

    .manual-group-chip-name {
        font-size: 12px;
        max-width: 180px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-group-chip-count {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-width: 20px;
        height: 20px;
        border-radius: 999px;
        border: 1px solid var(--baseAlt2Color);
        background: var(--baseColor);
        color: var(--txtHintColor);
        font-size: 11px;
        font-variant-numeric: tabular-nums;
        padding: 0 6px;
    }

    .manual-group-chip.is-active .manual-group-chip-count {
        border-color: color-mix(in srgb, var(--primaryColor) 38%, var(--baseAlt2Color));
        color: var(--txtPrimaryColor);
    }

    .manual-group-chip.is-partial .manual-group-chip-count {
        border-color: color-mix(in srgb, var(--primaryColor) 32%, var(--baseAlt2Color));
        color: var(--txtPrimaryColor);
    }

    .campaigns-header-row {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .campaign-row-item {
        align-items: flex-start;
        gap: 10px;
        border: 0;
        border-radius: 0;
        background: transparent;
    }

    .campaign-row-item:nth-child(odd) {
        background: var(--baseColor);
    }

    .campaign-row-item:nth-child(even) {
        background: var(--baseAlt1Color);
    }

    .campaign-row-item.is-editing {
        background: color-mix(in srgb, var(--primaryColor) 7%, var(--baseColor));
        box-shadow: inset 3px 0 0 color-mix(in srgb, var(--primaryColor) 45%, transparent);
    }

    .campaign-row-content {
        width: 100%;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 6px;
    }

    .campaign-row-top {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-row-subject {
        font-weight: 600;
    }

    .campaign-row-meta {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
    }

    .campaign-row-actions {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-row-btn {
        min-width: 96px;
    }

    .section-head-inline {
        display: inline-flex;
        flex-direction: row;
        align-items: baseline;
        flex-wrap: wrap;
        gap: 8px;
    }

    .meta-line {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        margin-top: 3px;
    }

    .meta-sep {
        opacity: 0.55;
    }

    .action-btn {
        min-width: 96px;
        min-height: var(--smBtnHeight);
    }
    .campaign-body-editor {
        min-height: 170px;
    }

    :global(.overlay-panel.newsletter-campaign-preview) {
        width: min(100%, 1180px);
        max-height: 94vh;
    }

    :global(.overlay-panel.newsletter-campaign-preview .panel-header .campaign-preview-device-tabs--modal) {
        margin-bottom: 0;
    }


    .campaign-preview-modal-header {
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: nowrap;
    }

    .campaign-preview-modal-title {
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 10px;
        min-width: 0;
        flex: 1 1 auto;
        white-space: nowrap;
    }

    .campaign-preview-modal-title h4,
    .campaign-preview-modal-title p {
        margin: 0;
        line-height: 1.2;
    }

    .campaign-preview-modal-title p {
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .campaign-preview-modal-actions {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 10px;
        flex-wrap: nowrap;
        margin-left: auto;
        flex: 0 0 auto;
        min-height: 34px;
    }

    .campaign-preview-device-tabs--modal {
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 0;
    }

    .campaign-preview-device-tabs--modal .tab-item {
        min-height: 34px;
        justify-content: center;
    }

    .campaign-preview-modal-close-btn {
        min-height: 34px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
    }

    .campaign-preview-modal-meta {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .campaign-preview-modal-box {
        min-height: min(72vh, 760px);
        max-height: min(74vh, 780px);
    }

    .campaign-preview-canvas--modal {
        min-height: min(66vh, 680px);
        padding: 6px;
    }

    .campaign-preview-frame-shell--modal {
        width: min(var(--campaign-preview-frame-width, 720px), 100%);
    }

    .campaign-preview-iframe--modal {
        height: min(64vh, 660px);
    }

    .campaign-preview-empty-state {
        min-height: 220px;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: calc(var(--baseRadius) - 2px);
        background: color-mix(in srgb, var(--baseColor) 88%, var(--baseAlt1Color));
        padding: 16px;
    }

    @media (max-width: 980px) {
        .subscriber-controls {
            align-items: stretch;
        }

        .subscribers-panel-header {
            align-items: flex-start;
            flex-wrap: wrap;
        }

        .subscribers-panel-header-actions {
            width: 100%;
            justify-content: space-between;
        }

        .subscriber-create-row-primary {
            grid-template-columns: minmax(220px, 1fr) minmax(220px, 1fr) minmax(140px, 170px);
        }

        .subscriber-create-row-primary.no-name {
            grid-template-columns: minmax(220px, 1fr) minmax(140px, 170px);
        }

        .subscriber-create-row-groups {
            grid-template-columns: minmax(0, 1fr) minmax(180px, 230px) auto;
        }

        .create-action-field {
            justify-content: flex-end;
        }

        .subscriber-edit-grid {
            grid-template-columns: 1fr;
        }

        .subscriber-table-list {
            overflow-x: auto;
        }

        .subscriber-table-head,
        .subscriber-row-grid {
            min-width: 1360px;
        }

        .campaign-layout-grid {
            grid-template-columns: 1fr;
        }

        .campaign-builder-layout {
            grid-template-columns: 1fr;
        }

        .campaign-top-controls {
            align-items: flex-start;
            justify-content: flex-start;
        }

        .campaign-head-inline--single-line {
            flex-wrap: wrap;
        }

        .campaign-head-inline--single-line .campaign-head-description {
            white-space: normal;
            overflow: visible;
            text-overflow: clip;
        }

        .campaign-builder-layout.is-split .campaign-builder-preview-side {
            border-left: 0;
            padding-left: 0;
            border-top: 1px solid var(--baseAlt2Color);
            padding-top: 12px;
        }

        .campaign-preview-header {
            flex-wrap: wrap;
        }

        .campaign-preview-header-actions {
            width: 100%;
            justify-content: flex-start;
            margin-left: 0;
        }

        .campaign-preview-device-tabs {
            flex: 1 1 auto;
        }

        .campaign-preview-modal-actions {
            width: 100%;
            justify-content: flex-start;
            margin-left: 0;
            flex-wrap: wrap;
        }

        .campaign-preview-modal-header {
            flex-wrap: wrap;
        }

        .campaign-preview-modal-title {
            white-space: normal;
            flex-wrap: wrap;
        }

        .campaign-preview-canvas {
            padding: 8px;
        }

        .campaign-preview-iframe {
            height: 320px;
        }

        .campaign-preview-iframe--modal {
            height: min(58vh, 560px);
        }

        .campaign-preview-expand-btn {
            margin-left: 0;
        }

        .campaign-audience-workspace {
            grid-template-columns: 1fr;
        }

        .manual-recipients-tiles {
            grid-template-columns: 1fr;
        }

        .campaign-audience-toolbar {
            gap: 10px;
        }

        .campaign-audience-toolbar-row--filters {
            grid-template-columns: 1fr;
        }

        .campaign-audience-toolbar-field--compact {
            max-width: none;
        }

        .campaign-audience-toolbar-row--actions {
            justify-content: flex-start;
        }

        .campaign-audience-toolbar-actions-row {
            width: 100%;
        }

        .subscriber-filter-grid {
            grid-template-columns: repeat(2, minmax(180px, 1fr));
            flex-basis: 100%;
        }

        .subscriber-filter-grid.group-enabled {
            grid-template-columns: repeat(2, minmax(180px, 1fr));
        }

        .subscriber-selection-popover {
            left: 10px;
            right: 10px;
            bottom: 12px;
            transform: none;
            border-radius: var(--baseRadius);
            flex-wrap: wrap;
            justify-content: center;
        }
    }

    @media (max-width: 640px) {
        .subscriber-create-row-primary,
        .subscriber-create-row-primary.no-name,
        .subscriber-create-row-groups {
            grid-template-columns: 1fr;
        }

        .subscriber-groups-action {
            justify-content: stretch;
        }

        .create-action-field {
            justify-content: stretch;
        }

        .add-subscriber-btn {
            width: 100%;
            min-width: 0;
        }
        .subscriber-filter-grid {
            grid-template-columns: 1fr;
        }

        .subscriber-filter-grid.group-enabled {
            grid-template-columns: 1fr;
        }

        .subscriber-groups-action .btn {
            width: 100%;
        }

        .subscriber-selection-popover {
            justify-content: flex-start;
        }

        .selection-summary {
            width: 100%;
        }

        .subscriber-selection-popover .btn {
            width: 100%;
        }
        .pagination-wrap {            justify-content: center;        }        .newsletter-list-item {
            padding: 7px;
        }

        .actions {
            width: 100%;
            justify-content: flex-start;
        }

        .subscriber-col-actions.actions {
            width: auto;
        }

        .campaign-builder-footer {
            justify-content: flex-start;
        }

        .campaign-head-row {
            align-items: stretch;
        }

        .campaign-audience-toolbar-actions-row {
            gap: 6px;
            width: 100%;
            justify-content: flex-start;
        }

        .campaign-audience-toolbar-actions-row .btn {
            flex: 1 1 auto;
        }

        .campaign-audience-actions-panel .campaign-audience-actions .action-btn {
            width: 100%;
        }

        .campaign-builder-view-tabs {
            width: fit-content;
            max-width: 100%;
            margin-left: 0;
        }

        .campaign-builder-view-tabs .tab-item {
            flex: 0 0 auto;
            min-width: 78px;
        }

        .campaign-builder-cta,
        .campaign-row-btn {
            width: 100%;
        }
    }
</style>













