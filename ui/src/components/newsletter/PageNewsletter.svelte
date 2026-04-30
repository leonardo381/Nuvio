<script>    import { querystring } from "svelte-spa-router";    import ApiClient from "@/utils/ApiClient";    import CommonHelper from "@/utils/CommonHelper";    import PageWrapper from "@/components/base/PageWrapper.svelte";    import RefreshButton from "@/components/base/RefreshButton.svelte";    import { pageTitle } from "@/stores/app";    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";    import { addSuccessToast } from "@/stores/toasts";    // NUVIO CUSTOM START: Newsletter V1 dedicated section/page (collection-backed).
    $pageTitle = "Newsletter";    const initialQueryParams = new URLSearchParams($querystring);    const subscriberStatuses = ["pending", "active", "unsubscribed"];    const subscriberLeadSource = "manual_dashboard";    const subscriberSortOptions = [        { value: "newest", label: "Newest" },        { value: "oldest", label: "Oldest" },        { value: "emailAsc", label: "Email A-Z" },        { value: "emailDesc", label: "Email Z-A" },        { value: "status", label: "Status" },    ];    const subscribersPageSize = 20;    const newsletterSections = new Set(["subscribers", "campaigns"]);    let activeSection = newsletterSections.has(initialQueryParams.get("newsletterTab"))        ? initialQueryParams.get("newsletterTab")        : "subscribers";    let websites = [];    let selectedWebsiteId = initialQueryParams.get("newsletterWebsite") || "";    let subscribers = [];
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
    let campaignForm = {        subject: "",        body: "",        recipientsType: "manual",        recipientsIds: [],    };    let campaignFormError = "";    let campaignWorkspace = "builder";    let campaignBuilderShowEditor = true;    let campaignBuilderShowPreview = false;    let campaignStatusFilter = "all";    let isCampaignPreviewExpanded = false;    let editingCampaignId = "";    let isSavingCampaign = false;    let subscriberSearch = "";
    let subscriberStatusFilter = "all";
    let subscriberGroupFilter = "all";
    let subscriberSort = "newest";
    let subscribersPage = 1;
    let selectedSubscriberIds = [];

    let pendingSendCampaign = null;
    let pendingDeleteSubscriber = null;
    let editingSubscriberId = "";
    let editingSubscriberForm = {
        name: "",
        email: "",
        status: "pending",
        groupIds: [],
    };
    let editingSubscriberError = "";
    let subscriberEmailInput;
    let subscriberGroupCountById = {};
    let lastWebsitesCollectionId = "";    let lastDataKey = "";    let lastSubscribersFilterKey = "";    let lastPersistedContextKey = "";    loadCollections();    $: websitesCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "websites") || null;    $: subscribersCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "subscribers") || null;
    $: campaignsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "campaigns") || null;
    $: subscriberGroupsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "subscribergroups") || null;
    $: missingCollectionNames = [];    $: if (!subscribersCollection?.id) {        missingCollectionNames.push("Subscribers");    }    $: if (!campaignsCollection?.id) {        missingCollectionNames.push("Campaigns");    };    $: hasNewsletterCollections = missingCollectionNames.length === 0;
    $: subscriberFieldKeys = new Set(
        !!subscribersCollection?.id
            ? CommonHelper.getAllCollectionIdentifiers(subscribersCollection).map((field) => `${field || ""}`.trim().toLowerCase())
            : [],
    );
    $: subscribersSupportsGroupsField = subscriberFieldKeys.has("groups");
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
    $: createCampaignDisabledReason = resolveCreateCampaignDisabledReason(
        isCreatingCampaign,
        selectedWebsiteId,
        campaignForm.subject,
        campaignForm.body,
        campaignForm.recipientsIds,
    );
    $: campaignSubjectValue = `${campaignForm.subject || ""}`.trim();
    $: campaignBodyValue = `${campaignForm.body || ""}`.trim();
    $: shouldShowCampaignSubjectValidation = !campaignSubjectValue && (!!campaignBodyValue || !!campaignFormError);
    $: shouldShowCampaignBodyValidation = !campaignBodyValue && (!!campaignSubjectValue || !!campaignFormError);
    $: filteredCampaigns = campaigns.filter((campaign) => {
        if (campaignStatusFilter === "all") {
            return true;
        }
        return normalizeStatus(campaign?.status) === campaignStatusFilter;
    });
    $: editingCampaign = campaigns.find((campaign) => campaign.id === editingCampaignId) || null;
    $: editingCampaignLabel = `${editingCampaign?.subject || ""}`.trim() || "(No subject)";
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
    $: if (hasNewsletterCollections) {        const nextContextKey = `${selectedWebsiteId || ""}|${activeSection || "subscribers"}`;        if (nextContextKey !== lastPersistedContextKey) {            lastPersistedContextKey = nextContextKey;            CommonHelper.replaceHashQueryParams({                newsletterWebsite: selectedWebsiteId || null,                newsletterTab: activeSection !== "subscribers" ? activeSection : null,            });        }    };    function resolveWebsitesSort(collection) {        const preferredSortFields = ["title", "name", "slug"];        const availableFields = new Set(            CommonHelper.getAllCollectionIdentifiers(collection).map((field) => `${field || ""}`.trim().toLowerCase()),        );        const validSortFields = preferredSortFields.filter((field) => availableFields.has(field));        if (!validSortFields.length) {            return "+id";        }        return validSortFields.map((field) => `+${field}`).join(",");    }    function resolveWebsiteLabel(website) {        return (            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim() || website?.id || ""        );    }    function normalizeEmail(email) {        return `${email || ""}`.trim().toLowerCase();    }    function normalizeSubscriberName(name) {
        return `${name || ""}`.trim().replace(/\s+/g, " ");
    }    function resolveSubscriberDisplayName(subscriber) {
        return normalizeSubscriberName(subscriber?.name);
    }    function normalizeStatus(status) {
        return `${status || ""}`.trim().toLowerCase();
    }

    function getSubscriberStatusLabel(status) {
        const normalized = normalizeStatus(status);
        if (normalized === "pending") {
            return "Pending";
        }
        if (normalized === "active") {
            return "Active";
        }
        if (normalized === "unsubscribed") {
            return "Unsubscribed";
        }
        return `${status || ""}`.trim() || "Unknown";
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
        return Array.isArray(subscriber?.groups) ? subscriber.groups.filter(Boolean) : [];
    }

    function hasSubscriberGroup(subscriber, groupId) {
        return getSubscriberGroupIds(subscriber).includes(groupId);
    }

    function getActiveSubscriberIdsForGroup(groupId) {
        if (!groupId) {
            return [];
        }
        return activeSubscribers
            .filter((subscriber) => hasSubscriberGroup(subscriber, groupId))
            .map((subscriber) => subscriber.id);
    }

    function isManualGroupFullySelected(groupId) {
        const groupIds = getActiveSubscriberIdsForGroup(groupId);
        return groupIds.length > 0 && groupIds.every((id) => campaignForm.recipientsIds.includes(id));
    }

    function toggleManualRecipientsForGroup(groupId) {
        clearCampaignFormError();
        const groupIds = getActiveSubscriberIdsForGroup(groupId);
        if (!groupIds.length) {
            return;
        }

        const next = new Set(campaignForm.recipientsIds);
        const shouldUnselect = groupIds.every((id) => next.has(id));

        groupIds.forEach((id) => {
            if (shouldUnselect) {
                next.delete(id);
            } else {
                next.add(id);
            }
        });

        campaignForm.recipientsIds = [...next];
    }

    function selectAllManualRecipients() {
        clearCampaignFormError();
        campaignForm.recipientsIds = activeSubscribers.map((subscriber) => subscriber.id);
    }

    function clearManualRecipients() {
        clearCampaignFormError();
        campaignForm.recipientsIds = [];
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

    function resolveCreateCampaignDisabledReason(isCreating, websiteId, subject, body, recipientsIds) {
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

        if (!Array.isArray(recipientsIds) || !recipientsIds.length) {
            return "Select at least one recipient.";
        }

        return "";
    }
    function resolveCampaignRecipientsCount(campaign) {        if (!campaign) {            return 0;        }        const recipientsType = `${campaign.recipientsType || "all"}`.toLowerCase();        if (recipientsType === "manual") {            return Array.isArray(campaign.recipientsIds) ? campaign.recipientsIds.filter(Boolean).length : 0;        }        return activeSubscribers.length;    }    function getSendCampaignDisabledReason(campaign) {        if (!campaign?.id) {            return "Invalid campaign.";        }        if (isSendingCampaign[campaign.id]) {            return "Sending campaign...";        }        if (normalizeStatus(campaign.status) === "sent") {            return "Campaign already sent.";        }        if (resolveCampaignRecipientsCount(campaign) < 1) {            return "No eligible recipients.";        }        return "";    }    function clearSubscriberFormError() {
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
    function focusSubscriberEmailInput() {
        subscriberEmailInput?.focus();
    }

    function toggleSubscriberFormGroup(groupId) {
        if (!groupId || !hasSubscriberGroupsFeature) {
            return;
        }

        clearSubscriberFormError();

        const currentGroupIds = Array.isArray(subscriberForm.groupIds) ? subscriberForm.groupIds : [];
        const nextGroupIds = currentGroupIds.includes(groupId)
            ? currentGroupIds.filter((id) => id !== groupId)
            : [...currentGroupIds, groupId];

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

        const currentGroupIds = Array.isArray(editingSubscriberForm.groupIds) ? editingSubscriberForm.groupIds : [];
        const nextGroupIds = currentGroupIds.includes(groupId)
            ? currentGroupIds.filter((id) => id !== groupId)
            : [...currentGroupIds, groupId];

        editingSubscriberForm = { ...editingSubscriberForm, groupIds: nextGroupIds };
    }
    function setActiveSection(section) {        if (newsletterSections.has(section)) {            activeSection = section;            if (section === "campaigns") {                campaignWorkspace = "builder";                campaignBuilderShowEditor = true;                campaignBuilderShowPreview = false;            }        }    }    function setSubscribersPage(page) {        const nextPage = Math.min(Math.max(page, 1), subscribersTotalPages);        subscribersPage = nextPage;    }    function isManualRecipientSelected(subscriberId) {        return campaignForm.recipientsIds.includes(subscriberId);    }    function toggleManualRecipient(subscriberId) {        clearCampaignFormError();        if (campaignForm.recipientsIds.includes(subscriberId)) {            campaignForm.recipientsIds = campaignForm.recipientsIds.filter((id) => id !== subscriberId);        } else {            campaignForm.recipientsIds = [...campaignForm.recipientsIds, subscriberId];        }    }    function isSubscriberSelected(subscriberId) {        return selectedSubscriberIds.includes(subscriberId);    }    function toggleSubscriberSelection(subscriberId) {        if (selectedSubscriberIds.includes(subscriberId)) {            selectedSubscriberIds = selectedSubscriberIds.filter((id) => id !== subscriberId);        } else {            selectedSubscriberIds = [...selectedSubscriberIds, subscriberId];        }    }    function toggleAllVisibleSubscribers() {        if (areAllVisibleSubscribersSelected) {            selectedSubscriberIds = selectedSubscriberIds.filter((id) => !visibleSubscriberIds.includes(id));            return;        }        const nextSelectedIds = new Set(selectedSubscriberIds);        visibleSubscriberIds.forEach((id) => nextSelectedIds.add(id));        selectedSubscriberIds = [...nextSelectedIds];    }    function resetSubscriberSelection() {        selectedSubscriberIds = [];    }    function openSendCampaignModal(campaign) {        const reason = getSendCampaignDisabledReason(campaign);        if (reason) {            return;        }        pendingSendCampaign = campaign;    }    function closeSendCampaignModal() {
        pendingSendCampaign = null;
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
                status: subscriberForm.status,
            };
            if (subscribersSupportsNameField) {
                payload.name = normalizeSubscriberName(subscriberForm.name);
            }
            if (subscribersSupportsSourceField) {
                payload.source = subscriberLeadSource;
            }
            if (hasSubscriberGroupsFeature && Array.isArray(subscriberForm.groupIds) && subscriberForm.groupIds.length) {
                payload.groups = subscriberForm.groupIds;
            }

            if (subscriberForm.status === "active") {
                payload.confirmedAt = new Date().toISOString();
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
                status: normalizeStatus(editingSubscriberForm.status) || "pending",
            };
            if (subscribersSupportsNameField) {
                payload.name = normalizeSubscriberName(editingSubscriberForm.name);
            }

            if (hasSubscriberGroupsFeature) {
                payload.groups = Array.isArray(editingSubscriberForm.groupIds) ? editingSubscriberForm.groupIds : [];
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
            await ApiClient.collection(subscriberGroupsCollection.id).create({
                website: selectedWebsiteId,
                name,
                slug: slugifyGroupName(name),
            });

            subscriberGroupForm = { name: "" };
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
            await Promise.all(updates);            resetSubscriberSelection();            await loadSubscribers();            addSuccessToast(`Updated ${updates.length} subscriber(s).`);        } catch (err) {            ApiClient.error(err);        }        isBulkUpdating = false;    }    async function sendCampaign(campaign) {        if (!campaign?.id || isSendingCampaign[campaign.id]) {            return false;        }        isSendingCampaign[campaign.id] = true;        isSendingCampaign = { ...isSendingCampaign };        let sent = false;        try {            const response = await ApiClient.send("/api/nuvio/newsletter/campaigns/send", {                method: "POST",                body: {                    campaignId: campaign.id,                },                requestKey: "nuvio_newsletter_send_" + campaign.id,            });            addSuccessToast(`Campaign sent to ${response?.recipientsCount || 0} recipient(s).`);            await loadCampaigns();            sent = true;        } catch (err) {            ApiClient.error(err);        }        delete isSendingCampaign[campaign.id];        isSendingCampaign = { ...isSendingCampaign };        return sent;    }    function resetCampaignComposer() {
        campaignForm = {
            subject: "",
            body: "",
            recipientsType: "manual",
            recipientsIds: [],
        };
        campaignFormError = "";
        campaignBuilderShowEditor = true;
        campaignBuilderShowPreview = false;
        closeCampaignPreviewModal();
        editingCampaignId = "";
    }

    function startEditCampaign(campaign) {
        if (!campaign?.id) {
            return;
        }

        const previousRecipientsType = `${campaign.recipientsType || "all"}`.toLowerCase();
        const resolvedRecipientsIds = previousRecipientsType === "manual"
            ? (Array.isArray(campaign.recipientsIds) ? campaign.recipientsIds.filter(Boolean) : [])
            : activeSubscribers.map((subscriber) => subscriber.id);

        editingCampaignId = campaign.id;
        campaignForm = {
            subject: `${campaign.subject || ""}`,
            body: `${campaign.body || ""}`,
            recipientsType: "manual",
            recipientsIds: [...new Set(resolvedRecipientsIds)],
        };
        campaignBuilderShowEditor = true;
        campaignBuilderShowPreview = false;
        closeCampaignPreviewModal();
        campaignFormError = "";
    }

    async function saveCampaignDraftFromComposer() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingCampaign || isSavingCampaign) {
            return;
        }

        const validationError = resolveCreateCampaignDisabledReason(
            false,
            selectedWebsiteId,
            campaignForm.subject,
            campaignForm.body,
            campaignForm.recipientsIds,
        );

        if (validationError) {
            campaignFormError = validationError;
            return;
        }

        campaignFormError = "";
        const payload = {
            website: selectedWebsiteId,
            subject: `${campaignForm.subject || ""}`.trim(),
            body: `${campaignForm.body || ""}`.trim(),
            status: "draft",
            recipientsType: "manual",
            recipientsIds: Array.isArray(campaignForm.recipientsIds) ? campaignForm.recipientsIds : [],
        };

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


    function refreshAll() {
        loadWebsites();
        loadSubscribers();
        loadCampaigns();
        loadSubscriberGroups();
    }
    // NUVIO CUSTOM END: Newsletter V1 dedicated section/page (collection-backed).
</script>

<PageWrapper>
    {#if !hasNewsletterCollections}
        <div class="alert alert-warning m-b-base">
            <div class="icon">
                <i class="ri-information-line" />
            </div>
            <div>                Newsletter collections are missing:                <strong>{missingCollectionNames.join(", ")}</strong>.                Run the latest migrations to enable Newsletter V1.            </div>        </div>    {:else}
        <section class="newsletter-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <div class="title-row">
                        <h2 class="m-0">Newsletter Operations</h2>
                        <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={refreshAll} />
                    </div>
                    <p class="txt-sm txt-hint m-b-0 head-description">Manage subscribers and campaigns by website in one place.</p>
                </div>

                <div class="head-selector">
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
        {#if $isCollectionsLoading && !selectedWebsiteId}            <div class="placeholder-section m-b-base">                <span class="loader loader-lg" />                <h1>Loading Newsletter...</h1>            </div>        {:else if !selectedWebsiteId}            <div class="placeholder-section m-b-base">                <h1>Select a website to manage Newsletter.</h1>                <p class="txt-sm txt-hint m-b-0">Once selected, subscribers and campaigns will be loaded automatically.</p>            </div>        {:else}
            <div class="tabs">
                <div class="tabs-content">
                    {#if activeSection === "subscribers"}
                        <section class="panel subscribers-section-panel">
                            <div class="subscribers-panel-header m-b-sm">
                                <div class="section-head section-head-inline m-b-0">
                                    <h4 class="m-0">Subscribers</h4>
                                    <span class="txt-sm txt-hint">Create and manage subscribers directly in the table.</span>
                                </div>
                                <div class="flex-fill" />
                                <div class="subscribers-panel-header-actions">
                                    <button
                                        type="button"
                                        class="btn btn-sm add-form-toggle-btn"
                                        class:btn-strong={!isSubscriberCreateOpen}
                                        class:btn-outline={isSubscriberCreateOpen}
                                        on:click={() => (isSubscriberCreateOpen = !isSubscriberCreateOpen)}
                                    >
                                        <i class={isSubscriberCreateOpen ? "ri-eye-off-line" : "ri-add-line"} aria-hidden="true" />
                                        <span class="txt">{isSubscriberCreateOpen ? "Hide form" : "Add form"}</span>
                                    </button>
                                    <span class="txt-sm txt-hint">
                                        {filteredSubscribers.length} shown | {subscribers.length} total
                                    </span>
                                </div>
                            </div>

                            {#if isSubscriberCreateOpen}
                                <form class="subscriber-create-form subscriber-inline-create m-b-sm" on:submit|preventDefault={createSubscriber}>
                                <div class="subscriber-create-top">
                                    <div class="create-email-field" class:with-name={subscribersSupportsNameField}>
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
                                        <div class="create-email-inline-field">
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
                                    </div>
                                    <div class="create-status-field">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-status">Status</label>
                                        <select
                                            id="subscriber-status"
                                            class="input input-sm"
                                            bind:value={subscriberForm.status}
                                            on:change={clearSubscriberFormError}
                                        >
                                            {#each subscriberStatuses as status}
                                                <option value={status}>{status}</option>
                                            {/each}
                                        </select>
                                    </div>
                                    <div class="create-action-field">
                                        <button
                                            type="submit"
                                            class="btn btn-strong add-subscriber-btn"
                                            class:btn-loading={isCreatingSubscriber}
                                            disabled={!!createSubscriberDisabledReason}
                                            title={createSubscriberDisabledReason || null}
                                        >
                                            <span class="txt">Add subscriber</span>
                                        </button>
                                    </div>
                                </div>
                                {#if hasSubscriberGroupsFeature}
                                    <div class="subscriber-groups-row">
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
                                                            class:is-selected={subscriberForm.groupIds.includes(group.id)}
                                                            on:click={() => toggleSubscriberFormGroup(group.id)}
                                                        >
                                                            {group.name}
                                                        </button>
                                                    {/each}
                                                {/if}
                                            </div>
                                        </div>

                                        <div class="subscriber-groups-create">
                                            <label class="txt-sm txt-hint block m-b-5" for="subscriber-group-name">Create group</label>
                                            <div class="group-create-row">
                                                <input
                                                    id="subscriber-group-name"
                                                    type="text"
                                                    class="input input-sm"
                                                    placeholder="e.g. VIP Clients"
                                                    bind:value={subscriberGroupForm.name}
                                                    on:input={clearSubscriberGroupFormError}
                                                />
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
                                            id="subscriber-search"                                            type="text"                                            class="input input-sm"                                            placeholder="Search by name or email..."                                            bind:value={subscriberSearch}                                        />                                    </div>                                    <div class="control-item">                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-filter-status">Status</label>                                        <select                                            id="subscriber-filter-status"                                            class="input input-sm"                                            bind:value={subscriberStatusFilter}                                        >                                            <option value="all">All statuses</option>                                            {#each subscriberStatuses as status}                                                <option value={status}>{status}</option>                                            {/each}                                        </select>                                    </div>                                    <div class="control-item">
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
                            {#if isLoadingSubscribers}                                <div class="loading-state">                                    <span class="loader loader-sm" />                                    <span class="txt-hint">Loading subscribers...</span>                                </div>                            {:else if !subscribers.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No subscribers yet for this website.</span>
                                    <span class="txt-sm txt-hint">Use “Add subscriber” to create your first contact.</span>
                                </div>
                            {:else if !filteredSubscribers.length}                                <div class="empty-state empty-state-stack">                                    <span>No subscribers match the current filters.</span>                                    <button                                        type="button"                                        class="btn btn-xs btn-outline"                                        on:click={() => {
                                            subscriberSearch = "";
                                            subscriberStatusFilter = "all";
                                            subscriberGroupFilter = "all";
                                            subscriberSort = "newest";
                                        }}
                                    >
                                        <span class="txt">Clear filters</span>                                    </button>                                </div>                            {:else}                                <div class="list list-compact subscriber-table-list">
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
                                        <div class="subscriber-col-confirmed">Confirmed</div>
                                        <div class="subscriber-col-added">Added</div>
                                        <div class="subscriber-col-groups">Groups</div>
                                        <div class="subscriber-col-actions">Actions</div>
                                    </div>
                                    <div class="list-content">                                        {#each pagedSubscribers as subscriber (subscriber.id)}                                            <div class="list-item newsletter-list-item subscriber-row-item" class:is-editing={editingSubscriberId === subscriber.id}>
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
                                                                <div class="subscriber-edit-field">
                                                                    <label class="txt-xs txt-hint block m-b-5">Status</label>
                                                                    <select
                                                                        class="input input-sm"
                                                                        bind:value={editingSubscriberForm.status}
                                                                        on:change={clearEditingSubscriberError}
                                                                    >
                                                                        {#each subscriberStatuses as status}
                                                                            <option value={status}>{status}</option>
                                                                        {/each}
                                                                    </select>
                                                                </div>
                                                                {#if hasSubscriberGroupsFeature}
                                                                    <div class="subscriber-edit-groups">
                                                                        <label class="txt-xs txt-hint block m-b-5">Groups</label>
                                                                        <div class="group-pill-list row-group-pill-list">
                                                                            {#if !subscriberGroups.length}
                                                                                <span class="txt-xs txt-hint">No groups created yet.</span>
                                                                            {:else}
                                                                                {#each subscriberGroups as group (group.id)}
                                                                                    <button
                                                                                        type="button"
                                                                                        class="group-pill-btn row-group-pill-btn"
                                                                                        class:is-selected={editingSubscriberForm.groupIds.includes(group.id)}
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
                                                        </div>
                                                        <div class="subscriber-col-email">
                                                            <span class="txt subscriber-primary-label">{subscriber.email}</span>
                                                        </div>
                                                        <div class="subscriber-col-status">
                                                            <span
                                                                class="status-chip"
                                                                class:is-active={normalizeStatus(subscriber.status) === "active"}
                                                                class:is-pending={normalizeStatus(subscriber.status) === "pending"}
                                                                class:is-unsubscribed={normalizeStatus(subscriber.status) === "unsubscribed"}
                                                            >
                                                                {getSubscriberStatusLabel(subscriber.status)}
                                                            </span>
                                                        </div>
                                                        <div class="subscriber-col-confirmed txt-xs txt-hint">
                                                            {#if subscriber.confirmedAt}
                                                                {formatDateTime(subscriber.confirmedAt)}
                                                            {:else}
                                                                -
                                                            {/if}
                                                        </div>
                                                        <div class="subscriber-col-added txt-xs txt-hint">
                                                            {formatDateTime(subscriber.created)}
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
                                                                class="btn btn-sm btn-strong action-btn"
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
                                                                    title="Testing status override"
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
                                                                    title="Testing status override"
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
                                    <div class="tabs-header compact combined left campaign-workspace-tabs">
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
                                        <span class="txt-sm">Editing: <strong>{editingCampaignLabel}</strong></span>
                                        <button type="button" class="btn btn-xs btn-outline" on:click={resetCampaignComposer}>
                                            <span class="txt">New draft</span>
                                        </button>
                                    </div>
                                {/if}

                                {#if campaignWorkspace === "builder"}
                                    <div class="campaign-head-row m-b-sm">
                                        <div class="campaign-head-inline">
                                            <h4 class="m-0">Campaign Builder</h4>
                                            <div class="campaign-step-label txt-xs txt-hint">Step 1 of 2</div>
                                            <p class="txt-sm txt-hint m-b-0 campaign-head-description">
                                                Write your campaign and review the preview before moving to audience.
                                            </p>
                                        </div>
                                        <div class="tabs-header compact combined left campaign-builder-view-tabs">
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
                                            <div class="campaign-editor-header m-b-xs">
                                                <h5 class="m-0">Editor</h5>
                                            </div>
                                            <label class="txt-sm txt-hint block m-b-5" for="campaign-subject">Subject</label>
                                            <input
                                                id="campaign-subject"
                                                type="text"
                                                class="input"
                                                placeholder="Newsletter subject..."
                                                bind:value={campaignForm.subject}
                                                on:input={clearCampaignFormError}
                                            />
                                            {#if shouldShowCampaignSubjectValidation}
                                                <div class="txt-xs txt-danger m-t-5">Subject is required.</div>
                                            {/if}

                                            <label class="txt-sm txt-hint block m-b-5 m-t-sm" for="campaign-body">Body</label>
                                            <textarea
                                                id="campaign-body"
                                                class="input campaign-body-input"
                                                rows="10"
                                                placeholder="Campaign HTML or plain text content..."
                                                bind:value={campaignForm.body}
                                                on:input={clearCampaignFormError}
                                            />
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
                                                    class="btn btn-strong action-btn campaign-builder-cta"
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
                                                <button
                                                    type="button"
                                                    class="btn btn-xs btn-outline campaign-preview-expand-btn"
                                                    on:click={openCampaignPreviewModal}
                                                >
                                                    <span class="txt">Expand</span>
                                                </button>
                                            </div>

                                            <div class="campaign-preview-box campaign-preview-html">
                                                {#if campaignBodyValue}
                                                    {@html campaignForm.body}
                                                {:else}
                                                    <p class="txt-sm txt-hint m-0">Body preview will appear here.</p>
                                                {/if}
                                            </div>
                                        </div>
                                        {/if}
                                    </div>
                                {:else}
                                    <div class="campaign-head-row m-b-sm">
                                        <div class="campaign-head-inline">
                                            <h4 class="m-0">Audience & Send</h4>
                                            <div class="campaign-step-label txt-xs txt-hint">Step 2 of 2</div>
                                            <p class="txt-sm txt-hint m-b-0 campaign-head-description">
                                                Finalize the audience and save the draft ready to send.
                                            </p>
                                        </div>
                                        <div class="campaign-audience-actions-wrap">
                                            <div class="campaign-audience-actions">
                                                <button type="button" class="btn btn-sm btn-outline action-btn" on:click={() => (campaignWorkspace = "builder")}>
                                                    <span class="txt">Back to Builder</span>
                                                </button>
                                                {#if editingCampaignId}
                                                    <button type="button" class="btn btn-sm btn-outline action-btn" on:click={resetCampaignComposer}>
                                                        <span class="txt">Cancel Edit</span>
                                                    </button>
                                                {/if}
                                                <button
                                                    type="submit"
                                                    form="campaign-audience-form"
                                                    class="btn btn-sm btn-strong action-btn"
                                                    class:btn-loading={isCreatingCampaign || isSavingCampaign}
                                                    disabled={!!createCampaignDisabledReason || isCreatingCampaign || isSavingCampaign}
                                                    title={createCampaignDisabledReason || null}
                                                >
                                                    <span class="txt">{editingCampaignId ? "Update draft" : "Create draft"}</span>
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    <div class="campaign-audience-summary m-b-sm">
                                        <div class="campaign-audience-stat">
                                            <span class="audience-stat-label">Draft</span>
                                            <span class="audience-stat-value">{campaignSubjectValue || "Untitled draft"}</span>
                                        </div>
                                        <div class="campaign-audience-stat">
                                            <span class="audience-stat-label">Status</span>
                                            <span class="audience-stat-value">{editingCampaignId ? "Editing existing draft" : "New draft"}</span>
                                        </div>
                                        <div class="campaign-audience-stat">
                                            <span class="audience-stat-label">Audience mode</span>
                                            <span class="audience-stat-value">Manual selection</span>
                                        </div>
                                        <div class="campaign-audience-stat">
                                            <span class="audience-stat-label">Recipients</span>
                                            <span class="audience-stat-value">
                                                {campaignForm.recipientsIds.length} selected / {activeSubscribers.length} active
                                            </span>
                                        </div>
                                        <div class="campaign-audience-stat">
                                            <span class="audience-stat-label">Groups</span>
                                            <span class="audience-stat-value">
                                                {hasSubscriberGroupsFeature ? `${subscriberGroups.length} available` : "Not enabled"}
                                            </span>
                                        </div>
                                    </div>

                                    <form id="campaign-audience-form" class="grid campaign-audience-form" on:submit|preventDefault={saveCampaignDraftFromComposer}>
                                        <div class="col-12 manual-recipients m-t-sm">
                                            <div class="manual-recipients-head">
                                                <h5 class="m-0">Recipients</h5>
                                                <span class="manual-recipients-count txt-xs">
                                                    {campaignForm.recipientsIds.length} selected / {activeSubscribers.length} active
                                                </span>
                                            </div>

                                            <div class="manual-group-tools m-b-sm">
                                                <div class="manual-group-tools-head">
                                                    <span class="txt-sm txt-hint">Choose individual recipients, mark everyone, or select by groups.</span>
                                                    <div class="manual-group-action-buttons">
                                                        <button type="button" class="btn btn-xs btn-outline" on:click={selectAllManualRecipients}>
                                                            <span class="txt">Mark everyone</span>
                                                        </button>
                                                        <button type="button" class="btn btn-xs btn-outline" on:click={clearManualRecipients}>
                                                            <span class="txt">Clear selection</span>
                                                        </button>
                                                    </div>
                                                </div>
                                                {#if hasSubscriberGroupsFeature && subscriberGroups.length}
                                                    <div class="manual-group-chip-list">
                                                        {#each subscriberGroups as group (group.id)}
                                                            <button
                                                                type="button"
                                                                class="manual-group-chip"
                                                                class:is-active={isManualGroupFullySelected(group.id)}
                                                                on:click={() => toggleManualRecipientsForGroup(group.id)}
                                                            >
                                                                <span class="manual-group-chip-name">{group.name}</span>
                                                                <span class="manual-group-chip-count">{getActiveSubscriberIdsForGroup(group.id).length}</span>
                                                            </button>
                                                        {/each}
                                                    </div>
                                                {:else if hasSubscriberGroupsFeature}
                                                    <p class="txt-xs txt-hint m-b-0">No groups available yet for this website.</p>
                                                {/if}
                                            </div>

                                            {#if !activeSubscribers.length}
                                                <div class="empty-state">No active subscribers available for manual selection.</div>
                                            {:else}
                                                <div class="manual-recipients-grid">
                                                    {#each activeSubscribers as subscriber (subscriber.id)}
                                                        <label class="manual-recipient-item">
                                                            <input
                                                                type="checkbox"
                                                                checked={isManualRecipientSelected(subscriber.id)}
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
                                                                    <span class="manual-recipient-groups txt-xs txt-hint">
                                                                        {#if getSubscriberGroupIds(subscriber).length}
                                                                            {getSubscriberGroupIds(subscriber)
                                                                                .map((groupId) => subscriberGroupsById.get(groupId)?.name)
                                                                                .filter(Boolean)
                                                                                .join(", ")}
                                                                        {:else}
                                                                            No groups
                                                                        {/if}
                                                                    </span>
                                                                {/if}
                                                            </span>
                                                        </label>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        {#if campaignFormError}
                                            <div class="col-12">
                                                <div class="txt-sm txt-danger">{campaignFormError}</div>
                                            </div>
                                        {/if}
                                    </form>
                                {/if}
                            </section>

                            <section class="panel campaign-list-panel">
                                <div class="campaigns-header-row m-b-sm">
                                    <div class="section-head m-b-0">
                                        <h4 class="m-0">Campaigns</h4>
                                        <p class="txt-sm txt-hint m-b-0">Edit drafts from here, then finalize audience and send.</p>
                                    </div>
                                    <div class="campaign-list-header-actions">
                                        <span class="txt-sm txt-hint">{campaigns.length} total | {draftCampaigns.length} drafts | {sentCampaigns.length} sent</span>
                                        {#if editingCampaignId}
                                            <button type="button" class="btn btn-xs btn-outline" on:click={resetCampaignComposer}>
                                                <span class="txt">Clear Edit</span>
                                            </button>
                                        {/if}
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
                                        <span class="txt-sm txt-hint">Create your first draft in Builder to start sending newsletters.</span>
                                    </div>
                                {:else if !filteredCampaigns.length}
                                    <div class="empty-state empty-state-stack">
                                        <span>No campaigns match this filter.</span>
                                        <span class="txt-sm txt-hint">Try another status filter to view more campaigns.</span>
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
                                                            <span class="txt campaign-row-subject">{campaign.subject || "(No subject)"}</span>
                                                            <span
                                                                class="status-chip"
                                                                class:is-active={normalizeStatus(campaign.status) === "sent"}
                                                                class:is-pending={normalizeStatus(campaign.status) === "draft"}
                                                                class:is-unsubscribed={normalizeStatus(campaign.status) !== "sent" && normalizeStatus(campaign.status) !== "draft"}
                                                            >
                                                                {campaign.status}
                                                            </span>
                                                        </div>
                                                        <div class="txt-xs txt-hint meta-line campaign-row-meta">
                                                            <span><strong>Recipients:</strong> {campaign.recipientsType || "all"}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>Est.:</strong> {resolveCampaignRecipientsCount(campaign)}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>Sent:</strong> {campaign.recipientsCount || 0}</span>
                                                            <span class="meta-sep">|</span>
                                                            <span><strong>Updated:</strong> {formatDateTime(campaign.updated || campaign.created)}</span>
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
                                                        {/if}

                                                        {#if normalizeStatus(campaign.status) !== "sent"}
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-strong action-btn campaign-row-btn"
                                                                class:btn-loading={isSendingCampaign[campaign.id]}
                                                                disabled={!!getSendCampaignDisabledReason(campaign)}
                                                                title={getSendCampaignDisabledReason(campaign) || null}
                                                                on:click={() => openSendCampaignModal(campaign)}
                                                            >
                                                                <span class="txt">Send</span>
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
            <div class="newsletter-modal-wrap" role="dialog" aria-modal="true" aria-label="Campaign preview">
                <button
                    type="button"
                    aria-label="Close campaign preview"
                    class="newsletter-modal-overlay"
                    on:click={closeCampaignPreviewModal}
                />
                <div class="newsletter-modal panel campaign-preview-modal" on:click|stopPropagation>
                    <div class="campaign-preview-modal-header m-b-sm">
                        <h4 class="m-0">Campaign Preview</h4>
                        <button type="button" class="btn btn-xs btn-outline" on:click={closeCampaignPreviewModal}>
                            <span class="txt">Close</span>
                        </button>
                    </div>
                    <div class="campaign-preview-box campaign-preview-html campaign-preview-modal-box">
                        {#if campaignBodyValue}
                            {@html campaignForm.body}
                        {:else}
                            <p class="txt-sm txt-hint m-0">Body preview will appear here.</p>
                        {/if}
                    </div>
                </div>
            </div>
        {/if}

        {#if pendingSendCampaign}
            <div class="newsletter-modal-wrap" role="dialog" aria-modal="true" aria-label="Confirm send campaign">
                <button
                    type="button"
                    aria-label="Close send confirmation"                    class="newsletter-modal-overlay"                    on:click={closeSendCampaignModal}                />                <div class="newsletter-modal panel" on:click|stopPropagation>                    <h4 class="m-t-0 m-b-xs">Send campaign now?</h4>                    <p class="txt-sm txt-hint m-b-sm">                        <strong>{pendingSendCampaign.subject}</strong> will be sent to approximately                        <strong> {pendingSendRecipientsCount}</strong> recipient(s).                    </p>                    <div class="flex gap-5">                        <button type="button" class="btn btn-sm btn-outline" on:click={closeSendCampaignModal}>                            <span class="txt">Cancel</span>                        </button>                        <button                            type="button"                            class="btn btn-sm btn-strong"                            class:btn-loading={isSendingCampaign[pendingSendCampaign.id]}                            disabled={!!isSendingCampaign[pendingSendCampaign.id]}                            on:click={confirmSendCampaign}                        >                            <span class="txt">Confirm send</span>                        </button>                    </div>                </div>
            </div>
        {/if}

        {#if pendingDeleteSubscriber}
            <div class="newsletter-modal-wrap" role="dialog" aria-modal="true" aria-label="Confirm delete subscriber">
                <button
                    type="button"
                    aria-label="Close delete confirmation"
                    class="newsletter-modal-overlay"
                    on:click={closeDeleteSubscriberModal}
                />
                <div class="newsletter-modal panel" on:click|stopPropagation>
                    <h4 class="m-t-0 m-b-xs">Delete subscriber?</h4>
                    <p class="txt-sm txt-hint m-b-sm">
                        <strong>{pendingDeleteSubscriber.email}</strong> will be permanently removed.
                    </p>
                    <div class="flex gap-5">
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
                    </div>
                </div>
            </div>
        {/if}
    {/if}
</PageWrapper>
<style>    .newsletter-head {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .subscribers-section-panel {
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .head-main {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 12px;
        flex-wrap: wrap;
    }

    .summary-title-wrap {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 260px;
    }

    .title-row {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 0;
    }

    .head-selector {
        width: min(100%, 560px);
    }

    .selector-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .selector-label {
        white-space: nowrap;
        min-width: 52px;
    }

    .selector-row .input {
        flex: 1 1 auto;
        min-width: 250px;
    }
    .summary-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        justify-content: flex-end;
        margin-left: auto;
        flex: 1 1 auto;
        min-width: 0;
        overflow: visible;
    }

    .campaign-workspace-tabs,
    .campaign-builder-view-tabs {
        margin: 0;
        display: inline-flex;
        align-items: center;
        align-self: flex-start;
        width: fit-content !important;
        max-width: 100%;
        flex: 0 0 auto;
        flex-wrap: wrap;
        gap: 2px;
        padding: 2px;
        border: 0 !important;
        border-radius: calc(var(--baseRadius) + 2px);
        background: var(--baseAlt1Color);
        overflow: hidden;
        box-shadow: none !important;
    }

    .operations-tabs {
        margin: 0;
        display: inline-flex;
        align-items: center;
        flex: 0 0 auto;
        gap: 2px;
        padding: 2px;
        border: 0 !important;
        border-radius: calc(var(--baseRadius) + 2px);
        background: var(--baseAlt1Color);
        overflow: hidden;
        box-shadow: none !important;
    }

    .operations-tabs .tab-item {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        min-height: 34px;
        padding: 0 16px;
        border: 0 !important;
        border-radius: calc(var(--baseRadius) - 1px);
        background: transparent;
        font-weight: 500;
        color: color-mix(in srgb, var(--txtPrimaryColor) 76%, var(--txtHintColor));
        transition: background-color 140ms ease, color 140ms ease, box-shadow 140ms ease;
    }

    .operations-tabs .tab-item + .tab-item {
        box-shadow: none;
    }

    .operations-tabs .tab-item .tab-icon {
        font-size: 13px;
        opacity: 0.72;
        transition: opacity 140ms ease, color 140ms ease;
    }

    .operations-tabs .tab-item:hover {
        background: color-mix(in srgb, var(--baseColor) 75%, var(--baseAlt1Color));
        color: var(--txtPrimaryColor);
    }

    .operations-tabs .tab-item:focus-visible {
        outline: none;
        box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--primaryColor) 50%, transparent);
    }

    .operations-tabs .tab-item.active {
        background: color-mix(in srgb, var(--baseColor) 96%, var(--baseAlt1Color));
        color: var(--txtPrimaryColor);
        font-weight: 600;
        box-shadow: none;
    }

    .operations-tabs .tab-item.active .tab-icon {
        opacity: 0.95;
        color: color-mix(in srgb, var(--txtPrimaryColor) 86%, var(--txtHintColor));
    }

    .campaign-workspace-tabs .tab-item,
    .campaign-builder-view-tabs .tab-item {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        min-height: 34px;
        border: 0 !important;
        border-radius: calc(var(--baseRadius) - 1px);
        background: transparent;
        color: color-mix(in srgb, var(--txtPrimaryColor) 76%, var(--txtHintColor));
        padding: 0 14px;
        font-weight: 500;
        transition: background-color 140ms ease, color 140ms ease, box-shadow 140ms ease;
    }

    .campaign-workspace-tabs .tab-item:hover,
    .campaign-builder-view-tabs .tab-item:hover {
        background: color-mix(in srgb, var(--baseColor) 75%, var(--baseAlt1Color));
        color: var(--txtPrimaryColor);
    }

    .campaign-workspace-tabs .tab-item:focus-visible,
    .campaign-builder-view-tabs .tab-item:focus-visible {
        outline: none;
        box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--primaryColor) 50%, transparent);
    }

    .campaign-workspace-tabs .tab-item.active,
    .campaign-builder-view-tabs .tab-item.active {
        background: color-mix(in srgb, var(--baseColor) 96%, var(--baseAlt1Color));
        color: var(--txtPrimaryColor);
        font-weight: 600;
        box-shadow: none;
    }

    .campaign-workspace-tabs .tab-item .tab-icon,
    .campaign-builder-view-tabs .tab-item .tab-icon {
        font-size: 13px;
        opacity: 0.72;
        transition: opacity 140ms ease, color 140ms ease;
    }

    .campaign-workspace-tabs .tab-item.active .tab-icon,
    .campaign-builder-view-tabs .tab-item.active .tab-icon {
        opacity: 0.95;
        color: color-mix(in srgb, var(--txtPrimaryColor) 86%, var(--txtHintColor));
    }

    .head-tools {
        display: grid;
        grid-template-columns: auto minmax(0, 1fr);
        align-items: center;
        gap: 8px;
    }

    .summary-pill {
        display: inline-flex;
        align-items: center;
        gap: 6px;        border: 1px solid var(--baseAlt2Color);        border-radius: 999px;        background: var(--baseAlt1Color);        color: var(--txtHintColor);        font-size: 12px;
        padding: 5px 9px;
        white-space: nowrap;
        flex: 0 0 auto;
    }

    .head-description {
        max-width: 460px;
    }
    .summary-pill i {
        color: var(--txtPrimaryColor);
        opacity: 0.85;
        font-size: 13px;
    }
.subscriber-controls {
        display: flex;
        align-items: flex-end;
        justify-content: flex-start;
        gap: 10px;
        flex-wrap: wrap;
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
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
        gap: 10px;
    }

    .subscriber-inline-create {
        border: 0;
        border-radius: 0;
        background: transparent;
        padding: 0;
    }

    .subscriber-create-top {
        display: grid;
        grid-template-columns: minmax(260px, 1fr) minmax(160px, 220px) minmax(140px, 170px);
        gap: 10px 12px;
        align-items: end;
    }

    .create-email-field,
    .create-status-field,
    .create-action-field {
        min-width: 0;
    }

    .create-email-field {
        display: grid;
        grid-template-columns: minmax(0, 1fr);
        gap: 10px;
        align-items: end;
    }

    .create-email-field.with-name {
        grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    }

    .create-name-field,
    .create-email-inline-field {
        min-width: 0;
    }

    .create-name-field .input,
    .create-email-inline-field .input {
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

    .subscriber-groups-row {
        display: grid;
        grid-template-columns: minmax(260px, 1fr) minmax(260px, 340px);
        gap: 10px 12px;
        align-items: end;
        padding-top: 4px;
    }

    .subscriber-groups-select,
    .subscriber-groups-create {
        min-width: 0;
    }

    .subscriber-filter-grid {
        display: grid;
        gap: 8px;
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

    .group-create-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .group-create-row .input {
        flex: 1 1 auto;
        min-width: 0;
    }

    .group-create-row .btn {
        flex: 0 0 auto;
        min-height: var(--inputHeight);
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
    }

    .form-group-pill-list .group-pill-btn {
        padding: 4px 8px;
        font-size: 10px;
    }

    .row-group-pill-list {
        margin-top: 7px;
    }

    .row-group-pill-btn {
        padding: 4px 8px;
        font-size: 10px;
    }

    .bulk-action-btn {
        min-height: var(--inputHeight);
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
        max-height: 380px;
        overflow: auto;
        white-space: pre-wrap;
        word-break: break-word;
        font-family: inherit;
        font-size: var(--baseFontSize);
        flex: 1 1 auto;
    }

    .campaign-preview-html :global(p:last-child) {
        margin-bottom: 0;
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
    .manual-recipients {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .manual-recipients-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
        margin-bottom: 8px;
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

    .manual-recipients-grid {
        display: grid;
        gap: 8px;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    }

    .manual-recipient-item {
        display: flex;
        align-items: flex-start;
        gap: 8px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 8px 10px;
        background: var(--baseAlt1Color);
        cursor: pointer;
        min-width: 0;
        transition: border-color 140ms ease, background-color 140ms ease;
    }

    .manual-recipient-item:hover {
        border-color: color-mix(in srgb, var(--primaryColor) 28%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--baseAlt1Color) 78%, var(--baseAlt2Color));
    }

    .manual-recipient-item input {
        margin-top: 2px;
        flex: 0 0 auto;
    }

    .manual-recipient-content {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .manual-recipient-title {
        font-size: 13px;
        font-weight: 500;
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-recipient-subtitle {
        font-size: 12px;
        color: var(--txtHintColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-recipient-groups {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .manual-recipient-item input:checked + .manual-recipient-content .manual-recipient-title {
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
    }

    .subscriber-row-item:nth-child(odd) {
        background: var(--baseColor);
    }

    .subscriber-row-item:nth-child(even) {
        background: color-mix(in srgb, var(--baseAlt1Color) 78%, var(--baseAlt2Color));
    }

    .subscriber-row-item.is-editing {
        background: var(--bodyColor);
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
        grid-template-columns: minmax(220px, 1fr) minmax(140px, 190px);
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

    .campaign-editor-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        min-height: 30px;
    }

    .campaign-preview-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .campaign-preview-expand-btn {
        margin-left: auto;
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
        min-width: 178px;
    }

    .campaign-audience-summary {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
        gap: 8px;
    }

    .campaign-audience-stat {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 8px 10px;
        min-width: 0;
    }

    .audience-stat-label {
        display: block;
        font-size: 11px;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: var(--txtHintColor);
        margin-bottom: 3px;
    }

    .audience-stat-value {
        display: block;
        font-size: 13px;
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .campaign-audience-form {
        row-gap: 10px;
    }

    .campaign-audience-actions-wrap {
        margin-left: auto;
        display: flex;
        align-items: center;
    }

    .campaign-audience-actions {
        display: inline-flex;
        align-items: center;
        justify-content: flex-end;
        gap: 8px;
        flex-wrap: wrap;
    }

    .campaign-audience-actions .action-btn {
        min-width: 112px;
    }

    .manual-group-tools {
        display: flex;
        flex-direction: column;
        gap: 7px;
    }

    .manual-group-tools-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
        margin-bottom: 8px;
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
        color: var(--txtPrimaryColor);
    }

    .manual-group-chip.is-active {
        border-color: color-mix(in srgb, var(--primaryColor) 38%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 10%, var(--baseColor));
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
        padding: 0 4px;
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

    .status-chip {
        display: inline-flex;
        align-items: center;
        border: 1px solid transparent;
        border-radius: 999px;
        font-size: 11px;
        line-height: 1;
        text-transform: capitalize;
        padding: 4px 8px;
    }

    .status-chip.is-active {
        background: color-mix(in srgb, var(--successColor) 12%, transparent);
        border-color: color-mix(in srgb, var(--successColor) 40%, transparent);
        color: var(--successColor);
    }

    .status-chip.is-pending {
        background: color-mix(in srgb, var(--warningColor) 14%, transparent);
        border-color: color-mix(in srgb, var(--warningColor) 40%, transparent);
        color: var(--warningColor);
    }

    .status-chip.is-unsubscribed {
        background: color-mix(in srgb, var(--dangerColor) 12%, transparent);
        border-color: color-mix(in srgb, var(--dangerColor) 38%, transparent);
        color: var(--dangerColor);
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
        min-height: var(--inputHeight);
    }

    .btn-strong {
        font-weight: 600;
    }

    .campaign-body-input {
        min-height: 170px;
        resize: vertical;
    }

    .newsletter-modal-wrap {
        position: fixed;
        inset: 0;
        z-index: 60;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 16px;
    }

    .newsletter-modal-overlay {
        position: absolute;
        inset: 0;
        border: 0;
        background: rgba(0, 0, 0, 0.42);
        cursor: default;
    }

    .newsletter-modal {
        position: relative;
        z-index: 1;
        width: min(100%, 460px);
    }

    .campaign-preview-modal {
        width: min(100%, 980px);
        max-height: 90vh;
        overflow: auto;
    }

    .campaign-preview-modal-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .campaign-preview-modal-box {
        min-height: min(66vh, 620px);
        max-height: min(68vh, 640px);
    }

    @media (max-width: 980px) {
        .head-tools {
            align-items: stretch;
            grid-template-columns: 1fr;
        }

        .summary-badges {
            justify-content: flex-start;
            margin-left: 0;
        }

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

        .subscriber-create-top {
            grid-template-columns: minmax(220px, 1fr) minmax(150px, 200px);
        }

        .create-email-field {
            grid-template-columns: 1fr;
        }

        .create-email-field.with-name {
            grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
        }

        .create-action-field {
            grid-column: 1 / -1;
            justify-content: flex-end;
        }

        .subscriber-groups-row {
            grid-template-columns: 1fr;
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

        .campaign-builder-layout.is-split .campaign-builder-preview-side {
            border-left: 0;
            padding-left: 0;
            border-top: 1px solid var(--baseAlt2Color);
            padding-top: 12px;
        }

        .campaign-preview-header {
            flex-wrap: wrap;
        }

        .campaign-preview-expand-btn {
            margin-left: 0;
        }

        .campaign-audience-actions-wrap {
            margin-left: 0;
            width: 100%;
            justify-content: flex-start;
        }

        .campaign-audience-actions {
            justify-content: flex-start;
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
        .head-main {
            align-items: stretch;
        }

        .head-selector {
            width: 100%;
        }

        .selector-row {
            flex-direction: column;
            align-items: stretch;
        }

        .selector-label {
            min-width: 0;
        }

        .selector-row .input {
            min-width: 0;
        }
        .subscriber-create-top {
            grid-template-columns: 1fr;
        }

        .create-email-field,
        .create-email-field.with-name {
            grid-template-columns: 1fr;
        }

        .create-action-field {
            justify-content: stretch;
        }

        .add-subscriber-btn {
            width: 100%;
            min-width: 0;
        }
        .summary-pill {            font-size: 11px;            padding: 5px 9px;        }        .subscriber-filter-grid {
            grid-template-columns: 1fr;
        }

        .subscriber-filter-grid.group-enabled {
            grid-template-columns: 1fr;
        }

        .group-create-row {
            flex-direction: column;
            align-items: stretch;
        }

        .group-create-row .btn {
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

        .campaign-audience-summary {
            grid-template-columns: 1fr;
        }

        .campaign-audience-actions {
            width: 100%;
            justify-content: stretch;
        }

        .campaign-audience-actions .action-btn {
            width: 100%;
            min-width: 0;
        }

        .manual-recipients-head {
            align-items: flex-start;
        }

        .manual-group-tools-head {
            align-items: flex-start;
        }

        .manual-group-action-buttons {
            width: 100%;
        }

        .manual-group-action-buttons .btn {
            flex: 1 1 auto;
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





















