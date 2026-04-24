<script>
    import { querystring } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addSuccessToast } from "@/stores/toasts";

    // NUVIO CUSTOM START: Newsletter V1 dedicated section/page (collection-backed).
    $pageTitle = "Newsletter";

    const initialQueryParams = new URLSearchParams($querystring);

    const subscriberStatuses = ["pending", "active", "unsubscribed"];
    const campaignRecipientsTypeOptions = ["all", "manual"];
    const subscriberSortOptions = [
        { value: "newest", label: "Newest" },
        { value: "oldest", label: "Oldest" },
        { value: "emailAsc", label: "Email A-Z" },
        { value: "emailDesc", label: "Email Z-A" },
        { value: "status", label: "Status" },
    ];
    const subscribersPageSize = 20;
    const newsletterSections = new Set(["subscribers", "campaigns"]);

    let activeSection = newsletterSections.has(initialQueryParams.get("newsletterTab"))
        ? initialQueryParams.get("newsletterTab")
        : "subscribers";
    let websites = [];
    let selectedWebsiteId = initialQueryParams.get("newsletterWebsite") || "";

    let subscribers = [];
    let campaigns = [];

    let isLoadingWebsites = false;
    let isLoadingSubscribers = false;
    let isLoadingCampaigns = false;
    let isCreatingSubscriber = false;
    let isCreatingCampaign = false;
    let isBulkUpdating = false;
    let isSendingCampaign = {};

    let subscriberForm = {
        email: "",
        status: "pending",
    };
    let subscriberFormError = "";

    let campaignForm = {
        subject: "",
        body: "",
        recipientsType: "all",
        recipientsIds: [],
    };
    let campaignFormError = "";
    let campaignPreviewMode = "plain";

    let subscriberSearch = "";
    let subscriberStatusFilter = "all";
    let subscriberSort = "newest";
    let subscribersPage = 1;
    let selectedSubscriberIds = [];

    let pendingSendCampaign = null;

    let subscriberEmailInput;

    let lastWebsitesCollectionId = "";
    let lastDataKey = "";
    let lastSubscribersFilterKey = "";
    let lastPersistedContextKey = "";

    loadCollections();

    $: websitesCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "websites") || null;
    $: subscribersCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "subscribers") || null;
    $: campaignsCollection = $collections.find((c) => (c?.name || "").toLowerCase() === "campaigns") || null;

    $: missingCollectionNames = [];
    $: if (!subscribersCollection?.id) {
        missingCollectionNames.push("Subscribers");
    }
    $: if (!campaignsCollection?.id) {
        missingCollectionNames.push("Campaigns");
    }

    $: hasNewsletterCollections = missingCollectionNames.length === 0;

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        lastWebsitesCollectionId = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: websiteDataKey = `${selectedWebsiteId}:${subscribersCollection?.id || ""}:${campaignsCollection?.id || ""}`;
    $: if (selectedWebsiteId && hasNewsletterCollections && websiteDataKey !== lastDataKey) {
        lastDataKey = websiteDataKey;
        loadSubscribers();
        loadCampaigns();
    }

    $: activeSubscribers = subscribers.filter((subscriber) => normalizeStatus(subscriber?.status) === "active");
    $: sentCampaigns = campaigns.filter((campaign) => normalizeStatus(campaign?.status) === "sent");
    $: draftCampaigns = campaigns.filter((campaign) => normalizeStatus(campaign?.status) === "draft");
    $: newSubscribersLast7Days = subscribers.filter((subscriber) => isWithinLastDays(subscriber?.created, 7)).length;
    $: unsubscribedLast7Days = subscribers.filter((subscriber) => {
        return normalizeStatus(subscriber?.status) === "unsubscribed"
            && isWithinLastDays(subscriber?.updated || subscriber?.created, 7);
    }).length;

    $: normalizedSubscriberFormEmail = normalizeEmail(subscriberForm.email);
    $: subscriberAlreadyExists = subscribers.some((subscriber) => normalizeEmail(subscriber?.email) === normalizedSubscriberFormEmail);
    $: createSubscriberDisabledReason = resolveCreateSubscriberDisabledReason();
    $: createCampaignDisabledReason = resolveCreateCampaignDisabledReason();

    $: normalizedSubscriberSearch = `${subscriberSearch || ""}`.trim().toLowerCase();
    $: filteredSubscribers = sortSubscribers(
        subscribers.filter((subscriber) => {
            const status = normalizeStatus(subscriber?.status);
            const byStatus = subscriberStatusFilter === "all" || status === subscriberStatusFilter;
            const bySearch = !normalizedSubscriberSearch
                || `${subscriber?.email || ""}`.toLowerCase().includes(normalizedSubscriberSearch);

            return byStatus && bySearch;
        }),
    );
    $: subscribersTotalPages = Math.max(1, Math.ceil(filteredSubscribers.length / subscribersPageSize));
    $: if (subscribersPage > subscribersTotalPages) {
        subscribersPage = subscribersTotalPages;
    }
    $: subscribersPageStart = (subscribersPage - 1) * subscribersPageSize;
    $: pagedSubscribers = filteredSubscribers.slice(subscribersPageStart, subscribersPageStart + subscribersPageSize);
    $: visibleSubscriberIds = pagedSubscribers.map((subscriber) => subscriber.id);
    $: areAllVisibleSubscribersSelected = visibleSubscriberIds.length > 0
        && visibleSubscriberIds.every((id) => selectedSubscriberIds.includes(id));
    $: selectedSubscribersCount = selectedSubscriberIds.length;

    $: pendingSendRecipientsCount = resolveCampaignRecipientsCount(pendingSendCampaign);

    $: {
        const nextFilterKey = `${selectedWebsiteId}|${subscriberSearch}|${subscriberStatusFilter}|${subscriberSort}|${subscribers.length}`;
        if (nextFilterKey !== lastSubscribersFilterKey) {
            lastSubscribersFilterKey = nextFilterKey;
            subscribersPage = 1;
        }
    }

    $: {
        const nextSelected = selectedSubscriberIds.filter((id) => filteredSubscribers.some((subscriber) => subscriber.id === id));
        if (nextSelected.length !== selectedSubscriberIds.length) {
            selectedSubscriberIds = nextSelected;
        }
    }

    // Keep context in URL query so refresh/navigation preserves website and active tab.
    $: if (hasNewsletterCollections) {
        const nextContextKey = `${selectedWebsiteId || ""}|${activeSection || "subscribers"}`;
        if (nextContextKey !== lastPersistedContextKey) {
            lastPersistedContextKey = nextContextKey;
            CommonHelper.replaceHashQueryParams({
                newsletterWebsite: selectedWebsiteId || null,
                newsletterTab: activeSection !== "subscribers" ? activeSection : null,
            });
        }
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
            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim() || website?.id || ""
        );
    }

    function normalizeEmail(email) {
        return `${email || ""}`.trim().toLowerCase();
    }

    function normalizeStatus(status) {
        return `${status || ""}`.trim().toLowerCase();
    }

    function isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalizeEmail(email));
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

    function isWithinLastDays(value, days) {
        const ts = toTimestamp(value);
        if (!ts) {
            return false;
        }

        const diff = Date.now() - ts;
        return diff >= 0 && diff <= (days * 24 * 60 * 60 * 1000);
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

    function sortSubscribers(list) {
        const sorted = [...list];

        sorted.sort((a, b) => {
            if (subscriberSort === "oldest") {
                return toTimestamp(a?.created) - toTimestamp(b?.created);
            }

            if (subscriberSort === "emailAsc") {
                return `${a?.email || ""}`.localeCompare(`${b?.email || ""}`);
            }

            if (subscriberSort === "emailDesc") {
                return `${b?.email || ""}`.localeCompare(`${a?.email || ""}`);
            }

            if (subscriberSort === "status") {
                const statusCompare = normalizeStatus(a?.status).localeCompare(normalizeStatus(b?.status));
                if (statusCompare !== 0) {
                    return statusCompare;
                }
            }

            // newest/default
            return toTimestamp(b?.created) - toTimestamp(a?.created);
        });

        return sorted;
    }

    function resolveCreateSubscriberDisabledReason() {
        if (isCreatingSubscriber) {
            return "Adding subscriber...";
        }

        if (!selectedWebsiteId) {
            return "Select a website first.";
        }

        if (!normalizedSubscriberFormEmail) {
            return "Enter subscriber email.";
        }

        if (!isValidEmail(normalizedSubscriberFormEmail)) {
            return "Enter a valid email.";
        }

        if (subscriberAlreadyExists) {
            return "This email already exists for this website.";
        }

        return "";
    }

    function resolveCreateCampaignDisabledReason() {
        if (isCreatingCampaign) {
            return "Creating draft...";
        }

        if (!selectedWebsiteId) {
            return "Select a website first.";
        }

        if (!`${campaignForm.subject || ""}`.trim()) {
            return "Campaign subject is required.";
        }

        if (!`${campaignForm.body || ""}`.trim()) {
            return "Campaign body is required.";
        }

        if (campaignForm.recipientsType === "manual" && !campaignForm.recipientsIds.length) {
            return "Select at least one manual recipient.";
        }

        return "";
    }

    function resolveCampaignRecipientsCount(campaign) {
        if (!campaign) {
            return 0;
        }

        const recipientsType = `${campaign.recipientsType || "all"}`.toLowerCase();
        if (recipientsType === "manual") {
            return Array.isArray(campaign.recipientsIds) ? campaign.recipientsIds.filter(Boolean).length : 0;
        }

        return activeSubscribers.length;
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

        if (resolveCampaignRecipientsCount(campaign) < 1) {
            return "No eligible recipients.";
        }

        return "";
    }

    function clearSubscriberFormError() {
        if (subscriberFormError) {
            subscriberFormError = "";
        }
    }

    function clearCampaignFormError() {
        if (campaignFormError) {
            campaignFormError = "";
        }
    }

    function focusSubscriberEmailInput() {
        subscriberEmailInput?.focus();
    }

    function setActiveSection(section) {
        if (newsletterSections.has(section)) {
            activeSection = section;
        }
    }

    function setSubscribersPage(page) {
        const nextPage = Math.min(Math.max(page, 1), subscribersTotalPages);
        subscribersPage = nextPage;
    }

    function isManualRecipientSelected(subscriberId) {
        return campaignForm.recipientsIds.includes(subscriberId);
    }

    function toggleManualRecipient(subscriberId) {
        clearCampaignFormError();

        if (campaignForm.recipientsIds.includes(subscriberId)) {
            campaignForm.recipientsIds = campaignForm.recipientsIds.filter((id) => id !== subscriberId);
        } else {
            campaignForm.recipientsIds = [...campaignForm.recipientsIds, subscriberId];
        }
    }

    function isSubscriberSelected(subscriberId) {
        return selectedSubscriberIds.includes(subscriberId);
    }

    function toggleSubscriberSelection(subscriberId) {
        if (selectedSubscriberIds.includes(subscriberId)) {
            selectedSubscriberIds = selectedSubscriberIds.filter((id) => id !== subscriberId);
        } else {
            selectedSubscriberIds = [...selectedSubscriberIds, subscriberId];
        }
    }

    function toggleAllVisibleSubscribers() {
        if (areAllVisibleSubscribersSelected) {
            selectedSubscriberIds = selectedSubscriberIds.filter((id) => !visibleSubscriberIds.includes(id));
            return;
        }

        const nextSelectedIds = new Set(selectedSubscriberIds);
        visibleSubscriberIds.forEach((id) => nextSelectedIds.add(id));
        selectedSubscriberIds = [...nextSelectedIds];
    }

    function resetSubscriberSelection() {
        selectedSubscriberIds = [];
    }

    function openSendCampaignModal(campaign) {
        const reason = getSendCampaignDisabledReason(campaign);
        if (reason) {
            return;
        }

        pendingSendCampaign = campaign;
    }

    function closeSendCampaignModal() {
        pendingSendCampaign = null;
    }

    async function confirmSendCampaign() {
        if (!pendingSendCampaign) {
            return;
        }

        const campaign = pendingSendCampaign;
        pendingSendCampaign = null;
        await sendCampaign(campaign);
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
                requestKey: "nuvio_newsletter_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                subscribers = [];
                campaigns = [];
                resetSubscriberSelection();
                return;
            }

            if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            subscribers = [];
            campaigns = [];
            resetSubscriberSelection();
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadSubscribers() {
        if (!hasNewsletterCollections || !selectedWebsiteId) {
            subscribers = [];
            resetSubscriberSelection();
            return;
        }

        isLoadingSubscribers = true;

        try {
            subscribers = await ApiClient.collection(subscribersCollection.id).getFullList({
                filter: `website="${selectedWebsiteId}"`,
                sort: "-created",
                requestKey: "nuvio_newsletter_subscribers_" + selectedWebsiteId,
            });
        } catch (err) {
            subscribers = [];
            ApiClient.error(err);
        }

        isLoadingSubscribers = false;
    }

    async function loadCampaigns() {
        if (!hasNewsletterCollections || !selectedWebsiteId) {
            campaigns = [];
            return;
        }

        isLoadingCampaigns = true;

        try {
            campaigns = await ApiClient.collection(campaignsCollection.id).getFullList({
                filter: `website="${selectedWebsiteId}"`,
                sort: "-created",
                requestKey: "nuvio_newsletter_campaigns_" + selectedWebsiteId,
            });
        } catch (err) {
            campaigns = [];
            ApiClient.error(err);
        }

        isLoadingCampaigns = false;
    }

    function handleWebsiteChange(event) {
        selectedWebsiteId = event.target.value || "";
        subscriberFormError = "";
        campaignFormError = "";
        pendingSendCampaign = null;
        resetSubscriberSelection();
    }

    async function createSubscriber() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingSubscriber) {
            return;
        }

        const email = normalizedSubscriberFormEmail;
        if (!isValidEmail(email)) {
            subscriberFormError = "Please provide a valid subscriber email.";
            return;
        }

        if (subscriberAlreadyExists) {
            subscriberFormError = "This email is already subscribed for this website.";
            return;
        }

        subscriberFormError = "";
        isCreatingSubscriber = true;

        try {
            const payload = {
                website: selectedWebsiteId,
                email,
                status: subscriberForm.status,
            };

            if (subscriberForm.status === "active") {
                payload.confirmedAt = new Date().toISOString();
            }

            await ApiClient.collection(subscribersCollection.id).create(payload);

            subscriberForm = {
                email: "",
                status: "pending",
            };

            await loadSubscribers();
            addSuccessToast("Subscriber added.");
            focusSubscriberEmailInput();
        } catch (err) {
            ApiClient.error(err);
        }

        isCreatingSubscriber = false;
    }

    async function setSubscriberStatus(subscriber, status) {
        if (!subscriber?.id || !hasNewsletterCollections) {
            return;
        }

        try {
            const payload = { status };
            if (status === "active" && !subscriber.confirmedAt) {
                payload.confirmedAt = new Date().toISOString();
            }

            await ApiClient.collection(subscribersCollection.id).update(subscriber.id, payload);
            await loadSubscribers();
            addSuccessToast(`Subscriber marked as ${status}.`);
        } catch (err) {
            ApiClient.error(err);
        }
    }

    async function applyBulkStatus(status) {
        if (!selectedSubscriberIds.length || isBulkUpdating || !hasNewsletterCollections) {
            return;
        }

        isBulkUpdating = true;

        try {
            const updates = selectedSubscriberIds
                .map((subscriberId) => {
                    const subscriber = subscribers.find((item) => item.id === subscriberId);
                    if (!subscriber || normalizeStatus(subscriber.status) === normalizeStatus(status)) {
                        return null;
                    }

                    const payload = { status };
                    if (status === "active" && !subscriber.confirmedAt) {
                        payload.confirmedAt = new Date().toISOString();
                    }

                    return ApiClient.collection(subscribersCollection.id).update(subscriber.id, payload);
                })
                .filter(Boolean);

            if (!updates.length) {
                addSuccessToast("Selected subscribers are already up to date.");
                return;
            }

            await Promise.all(updates);
            resetSubscriberSelection();
            await loadSubscribers();
            addSuccessToast(`Updated ${updates.length} subscriber(s).`);
        } catch (err) {
            ApiClient.error(err);
        }

        isBulkUpdating = false;
    }

    async function createCampaignDraft() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingCampaign) {
            return;
        }

        const validationError = resolveCreateCampaignDisabledReason();
        if (validationError) {
            campaignFormError = validationError;
            return;
        }

        campaignFormError = "";
        isCreatingCampaign = true;

        try {
            await ApiClient.collection(campaignsCollection.id).create({
                website: selectedWebsiteId,
                subject: `${campaignForm.subject || ""}`.trim(),
                body: `${campaignForm.body || ""}`.trim(),
                status: "draft",
                recipientsType: campaignForm.recipientsType,
                recipientsIds: campaignForm.recipientsType === "manual" ? campaignForm.recipientsIds : [],
                recipientsCount: 0,
            });

            campaignForm = {
                subject: "",
                body: "",
                recipientsType: "all",
                recipientsIds: [],
            };
            campaignPreviewMode = "plain";

            await loadCampaigns();
            addSuccessToast("Draft campaign created.");
        } catch (err) {
            ApiClient.error(err);
        }

        isCreatingCampaign = false;
    }

    async function sendCampaign(campaign) {
        if (!campaign?.id || isSendingCampaign[campaign.id]) {
            return false;
        }

        isSendingCampaign[campaign.id] = true;
        isSendingCampaign = { ...isSendingCampaign };

        let sent = false;

        try {
            const response = await ApiClient.send("/api/nuvio/newsletter/campaigns/send", {
                method: "POST",
                body: {
                    campaignId: campaign.id,
                },
                requestKey: "nuvio_newsletter_send_" + campaign.id,
            });

            addSuccessToast(`Campaign sent to ${response?.recipientsCount || 0} recipient(s).`);
            await loadCampaigns();
            sent = true;
        } catch (err) {
            ApiClient.error(err);
        }

        delete isSendingCampaign[campaign.id];
        isSendingCampaign = { ...isSendingCampaign };

        return sent;
    }

    function refreshAll() {
        loadWebsites();
        loadSubscribers();
        loadCampaigns();
    }
    // NUVIO CUSTOM END: Newsletter V1 dedicated section/page (collection-backed).
</script>

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Newsletter</div>
        </nav>

        <RefreshButton on:refresh={refreshAll} />
    </header>

    {#if !hasNewsletterCollections}
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
        <section class="newsletter-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <h3 class="m-0">Newsletter Operations</h3>
                    <p class="txt-sm txt-hint m-b-0">Manage subscribers and campaigns by website in one place.</p>
                </div>

                <div class="head-selector">
                    <label class="txt-sm txt-hint block m-b-5" for="newsletter-website">Website</label>
                    <div class="selector-row">
                        <select
                            id="newsletter-website"
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
                        <button
                            type="button"
                            class="btn btn-sm btn-outline"
                            disabled={isLoadingWebsites}
                            on:click={loadWebsites}
                        >
                            <i class="ri-refresh-line" />
                            <span class="txt">Reload sites</span>
                        </button>
                    </div>
                </div>
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
                    <i class="ri-user-add-line" />
                    {newSubscribersLast7Days} new / 7d
                </span>
                <span class="summary-pill">
                    <i class="ri-user-unfollow-line" />
                    {unsubscribedLast7Days} unsubscribed / 7d
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
                <span class="summary-pill summary-hint-pill">Scope and actions are tied to selected website</span>
            </div>
        </section>

        {#if $isCollectionsLoading && !selectedWebsiteId}
            <div class="placeholder-section m-b-base">
                <span class="loader loader-lg" />
                <h1>Loading Newsletter...</h1>
            </div>
        {:else if !selectedWebsiteId}
            <div class="placeholder-section m-b-base">
                <h1>Select a website to manage Newsletter.</h1>
                <p class="txt-sm txt-hint m-b-0">Once selected, subscribers and campaigns will be loaded automatically.</p>
            </div>
        {:else}
            <div class="tabs">
                <div class="tabs-header compact combined left">
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "subscribers"}
                        on:click={() => setActiveSection("subscribers")}
                    >
                        Subscribers
                    </button>
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "campaigns"}
                        on:click={() => setActiveSection("campaigns")}
                    >
                        Campaigns
                    </button>
                </div>

                <div class="tabs-content">
                    {#if activeSection === "subscribers"}
                        <section class="panel m-b-base">
                            <div class="section-head m-b-sm">
                                <h4 class="m-0">Add subscriber</h4>
                                <p class="txt-sm txt-hint m-b-0">Add contacts manually and set their initial lifecycle status.</p>
                            </div>
                            <form class="grid" on:submit|preventDefault={createSubscriber}>
                                <div class="col-md-7">
                                    <label class="txt-sm txt-hint block m-b-5" for="subscriber-email">Email</label>
                                    <input
                                        id="subscriber-email"
                                        bind:this={subscriberEmailInput}
                                        type="email"
                                        class="input"
                                        placeholder="name@example.com"
                                        bind:value={subscriberForm.email}
                                        on:input={clearSubscriberFormError}
                                    />
                                </div>
                                <div class="col-md-3">
                                    <label class="txt-sm txt-hint block m-b-5" for="subscriber-status">Status</label>
                                    <select
                                        id="subscriber-status"
                                        class="input"
                                        bind:value={subscriberForm.status}
                                        on:change={clearSubscriberFormError}
                                    >
                                        {#each subscriberStatuses as status}
                                            <option value={status}>{status}</option>
                                        {/each}
                                    </select>
                                </div>
                                <div class="col-md-2 align-end">
                                    <button
                                        type="submit"
                                        class="btn btn-block btn-strong"
                                        class:btn-loading={isCreatingSubscriber}
                                        disabled={!!createSubscriberDisabledReason}
                                        title={createSubscriberDisabledReason || null}
                                    >
                                        <span class="txt">Add subscriber</span>
                                    </button>
                                </div>
                                {#if subscriberFormError}
                                    <div class="col-12">
                                        <div class="txt-sm txt-danger">{subscriberFormError}</div>
                                    </div>
                                {/if}
                            </form>
                        </section>

                        <section class="panel">
                            <div class="flex m-b-sm">
                                <h4 class="m-0">Subscribers</h4>
                                <div class="flex-fill" />
                                <span class="txt-sm txt-hint">
                                    {filteredSubscribers.length} shown | {subscribers.length} total
                                </span>
                            </div>

                            <div class="subscriber-controls m-b-sm">
                                <div class="subscriber-filter-grid">
                                    <div class="control-item">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-search">Search</label>
                                        <input
                                            id="subscriber-search"
                                            type="text"
                                            class="input input-sm"
                                            placeholder="Search by email..."
                                            bind:value={subscriberSearch}
                                        />
                                    </div>
                                    <div class="control-item">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-filter-status">Status</label>
                                        <select
                                            id="subscriber-filter-status"
                                            class="input input-sm"
                                            bind:value={subscriberStatusFilter}
                                        >
                                            <option value="all">All statuses</option>
                                            {#each subscriberStatuses as status}
                                                <option value={status}>{status}</option>
                                            {/each}
                                        </select>
                                    </div>
                                    <div class="control-item">
                                        <label class="txt-sm txt-hint block m-b-5" for="subscriber-sort">Sort</label>
                                        <select id="subscriber-sort" class="input input-sm" bind:value={subscriberSort}>
                                            {#each subscriberSortOptions as sortOption}
                                                <option value={sortOption.value}>{sortOption.label}</option>
                                            {/each}
                                        </select>
                                    </div>
                                </div>

                                <div class="bulk-controls">
                                    <label class="bulk-select-all">
                                        <input
                                            type="checkbox"
                                            checked={areAllVisibleSubscribersSelected}
                                            disabled={!pagedSubscribers.length}
                                            on:change={toggleAllVisibleSubscribers}
                                        />
                                        <span class="txt-sm txt-hint">Select page</span>
                                    </label>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        disabled={!selectedSubscribersCount || isBulkUpdating}
                                        on:click={() => applyBulkStatus("active")}
                                        title={!selectedSubscribersCount ? "Select at least one subscriber." : null}
                                    >
                                        <span class="txt">Mark selected active</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        disabled={!selectedSubscribersCount || isBulkUpdating}
                                        on:click={() => applyBulkStatus("unsubscribed")}
                                        title={!selectedSubscribersCount ? "Select at least one subscriber." : null}
                                    >
                                        <span class="txt">Unsubscribe selected</span>
                                    </button>
                                </div>
                            </div>

                            {#if selectedSubscribersCount}
                                <div class="txt-sm txt-hint m-b-sm">
                                    {selectedSubscribersCount} subscriber(s) selected.
                                </div>
                            {/if}

                            {#if isLoadingSubscribers}
                                <div class="loading-state">
                                    <span class="loader loader-sm" />
                                    <span class="txt-hint">Loading subscribers...</span>
                                </div>
                            {:else if !subscribers.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No subscribers yet for this website.</span>
                                    <span class="txt-sm txt-hint">Use “Add subscriber” above to create your first contact.</span>
                                </div>
                            {:else if !filteredSubscribers.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No subscribers match the current filters.</span>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline"
                                        on:click={() => {
                                            subscriberSearch = "";
                                            subscriberStatusFilter = "all";
                                            subscriberSort = "newest";
                                        }}
                                    >
                                        <span class="txt">Clear filters</span>
                                    </button>
                                </div>
                            {:else}
                                <div class="list list-compact">
                                    <div class="list-content">
                                        {#each pagedSubscribers as subscriber (subscriber.id)}
                                            <div class="list-item newsletter-list-item">
                                                <div class="selection-cell">
                                                    <input
                                                        type="checkbox"
                                                        checked={isSubscriberSelected(subscriber.id)}
                                                        aria-label={`Select ${subscriber.email}`}
                                                        on:change={() => toggleSubscriberSelection(subscriber.id)}
                                                    />
                                                </div>
                                                <div class="content">
                                                    <div class="subscriber-title">
                                                        <span class="txt">{subscriber.email}</span>
                                                        <span
                                                            class="status-chip"
                                                            class:is-active={normalizeStatus(subscriber.status) === "active"}
                                                            class:is-pending={normalizeStatus(subscriber.status) === "pending"}
                                                            class:is-unsubscribed={normalizeStatus(subscriber.status) === "unsubscribed"}
                                                        >
                                                            {subscriber.status}
                                                        </span>
                                                    </div>
                                                    <div class="txt-xs txt-hint meta-line">
                                                        {#if subscriber.confirmedAt}
                                                            Confirmed: {formatDateTime(subscriber.confirmedAt)}
                                                            <span class="meta-sep">|</span>
                                                        {/if}
                                                        Added: {formatDateTime(subscriber.created)}
                                                    </div>
                                                </div>
                                                <div class="actions">
                                                    {#if normalizeStatus(subscriber.status) !== "active"}
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs btn-outline action-btn"
                                                            on:click={() => setSubscriberStatus(subscriber, "active")}
                                                        >
                                                            <span class="txt">Mark active</span>
                                                        </button>
                                                    {/if}
                                                    {#if normalizeStatus(subscriber.status) !== "unsubscribed"}
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs btn-outline action-btn"
                                                            on:click={() => setSubscriberStatus(subscriber, "unsubscribed")}
                                                        >
                                                            <span class="txt">Unsubscribe</span>
                                                        </button>
                                                    {/if}
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                </div>

                                {#if filteredSubscribers.length > subscribersPageSize}
                                    <div class="pagination-wrap">
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-outline"
                                            disabled={subscribersPage <= 1}
                                            on:click={() => setSubscribersPage(subscribersPage - 1)}
                                        >
                                            <span class="txt">Previous</span>
                                        </button>
                                        <span class="txt-sm txt-hint">
                                            Page {subscribersPage} of {subscribersTotalPages}
                                        </span>
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-outline"
                                            disabled={subscribersPage >= subscribersTotalPages}
                                            on:click={() => setSubscribersPage(subscribersPage + 1)}
                                        >
                                            <span class="txt">Next</span>
                                        </button>
                                    </div>
                                {/if}
                            {/if}
                        </section>
                    {:else}
                        <section class="panel m-b-base">
                            <div class="section-head m-b-sm">
                                <h4 class="m-0">Create campaign draft</h4>
                                <p class="txt-sm txt-hint m-b-0">Build your message, preview, choose recipients, then send when ready.</p>
                            </div>
                            <form class="grid" on:submit|preventDefault={createCampaignDraft}>
                                <div class="col-12">
                                    <label class="txt-sm txt-hint block m-b-5" for="campaign-subject">Subject</label>
                                    <input
                                        id="campaign-subject"
                                        type="text"
                                        class="input"
                                        placeholder="Newsletter subject..."
                                        bind:value={campaignForm.subject}
                                        on:input={clearCampaignFormError}
                                    />
                                </div>
                                <div class="col-12">
                                    <label class="txt-sm txt-hint block m-b-5" for="campaign-body">Body</label>
                                    <textarea
                                        id="campaign-body"
                                        class="input campaign-body-input"
                                        rows="7"
                                        placeholder="Campaign HTML or plain text content..."
                                        bind:value={campaignForm.body}
                                        on:input={clearCampaignFormError}
                                    />
                                </div>
                                <div class="col-md-4">
                                    <label class="txt-sm txt-hint block m-b-5" for="campaign-recipients-type">
                                        Recipients
                                    </label>
                                    <select
                                        id="campaign-recipients-type"
                                        class="input"
                                        bind:value={campaignForm.recipientsType}
                                        on:change={clearCampaignFormError}
                                    >
                                        {#each campaignRecipientsTypeOptions as recipientsType}
                                            <option value={recipientsType}>{recipientsType}</option>
                                        {/each}
                                    </select>
                                </div>
                                <div class="col-md-8 align-end">
                                    <button
                                        type="submit"
                                        class="btn btn-strong"
                                        class:btn-loading={isCreatingCampaign}
                                        disabled={!!createCampaignDisabledReason}
                                        title={createCampaignDisabledReason || null}
                                    >
                                        <span class="txt">Create draft</span>
                                    </button>
                                </div>
                                {#if campaignFormError}
                                    <div class="col-12">
                                        <div class="txt-sm txt-danger">{campaignFormError}</div>
                                    </div>
                                {/if}
                            </form>

                            <div class="campaign-preview m-t-sm">
                                <div class="flex m-b-xs">
                                    <h5 class="m-0">Preview</h5>
                                    <div class="flex-fill" />
                                    <div class="tabs-header compact combined left preview-tabs">
                                        <button
                                            type="button"
                                            class="tab-item"
                                            class:active={campaignPreviewMode === "plain"}
                                            on:click={() => (campaignPreviewMode = "plain")}
                                        >
                                            Plain
                                        </button>
                                        <button
                                            type="button"
                                            class="tab-item"
                                            class:active={campaignPreviewMode === "html"}
                                            on:click={() => (campaignPreviewMode = "html")}
                                        >
                                            HTML
                                        </button>
                                    </div>
                                </div>
                                <div class="txt-sm txt-hint m-b-xs">
                                    Subject: {`${campaignForm.subject || ""}`.trim() || "(No subject yet)"}
                                </div>
                                {#if campaignPreviewMode === "plain"}
                                    <pre class="campaign-preview-box">{`${campaignForm.body || ""}`.trim() || "Body preview will appear here."}</pre>
                                {:else}
                                    <div class="campaign-preview-box campaign-preview-html">
                                        {#if `${campaignForm.body || ""}`.trim()}
                                            {@html campaignForm.body}
                                        {:else}
                                            <p class="txt-sm txt-hint m-0">Body preview will appear here.</p>
                                        {/if}
                                    </div>
                                {/if}
                            </div>

                            {#if campaignForm.recipientsType === "manual"}
                                <div class="manual-recipients m-t-sm">
                                    <h5 class="m-t-0 m-b-xs">Manual recipients ({activeSubscribers.length} active)</h5>
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
                                                    <span>{subscriber.email}</span>
                                                </label>
                                            {/each}
                                        </div>
                                    {/if}
                                </div>
                            {/if}
                        </section>

                        <section class="panel">
                            <div class="flex m-b-sm">
                                <h4 class="m-0">Campaigns</h4>
                                <div class="flex-fill" />
                                <span class="txt-sm txt-hint">{campaigns.length} total | {sentCampaigns.length} sent</span>
                            </div>

                            {#if isLoadingCampaigns}
                                <div class="loading-state">
                                    <span class="loader loader-sm" />
                                    <span class="txt-hint">Loading campaigns...</span>
                                </div>
                            {:else if !campaigns.length}
                                <div class="empty-state empty-state-stack">
                                    <span>No campaigns yet for this website.</span>
                                    <span class="txt-sm txt-hint">Create your first draft above to start sending newsletters.</span>
                                </div>
                            {:else}
                                <div class="list list-compact">
                                    <div class="list-content">
                                        {#each campaigns as campaign (campaign.id)}
                                            <div class="list-item newsletter-list-item">
                                                <div class="content">
                                                    <div class="subscriber-title">
                                                        <span class="txt">{campaign.subject}</span>
                                                        <span
                                                            class="status-chip"
                                                            class:is-active={normalizeStatus(campaign.status) === "sent"}
                                                            class:is-pending={normalizeStatus(campaign.status) === "draft"}
                                                            class:is-unsubscribed={normalizeStatus(campaign.status) !== "sent" && normalizeStatus(campaign.status) !== "draft"}
                                                        >
                                                            {campaign.status}
                                                        </span>
                                                    </div>
                                                    <div class="txt-xs txt-hint meta-line">
                                                        Recipients type: {campaign.recipientsType}
                                                        <span class="meta-sep">|</span>
                                                        Estimated: {resolveCampaignRecipientsCount(campaign)}
                                                        <span class="meta-sep">|</span>
                                                        Sent count: {campaign.recipientsCount || 0}
                                                        <span class="meta-sep">|</span>
                                                        Sent at: {formatDateTime(campaign.sentAt)}
                                                        <span class="meta-sep">|</span>
                                                        Created: {formatDateTime(campaign.created)}
                                                    </div>
                                                </div>
                                                <div class="actions">
                                                    {#if normalizeStatus(campaign.status) !== "sent"}
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs btn-strong"
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
                    {/if}
                </div>
            </div>
        {/if}

        {#if pendingSendCampaign}
            <div class="newsletter-modal-wrap" role="dialog" aria-modal="true" aria-label="Confirm send campaign">
                <button
                    type="button"
                    aria-label="Close send confirmation"
                    class="newsletter-modal-overlay"
                    on:click={closeSendCampaignModal}
                />
                <div class="newsletter-modal panel" on:click|stopPropagation>
                    <h4 class="m-t-0 m-b-xs">Send campaign now?</h4>
                    <p class="txt-sm txt-hint m-b-sm">
                        <strong>{pendingSendCampaign.subject}</strong> will be sent to approximately
                        <strong> {pendingSendRecipientsCount}</strong> recipient(s).
                    </p>
                    <div class="flex gap-5">
                        <button type="button" class="btn btn-sm btn-outline" on:click={closeSendCampaignModal}>
                            <span class="txt">Cancel</span>
                        </button>
                        <button
                            type="button"
                            class="btn btn-sm btn-strong"
                            class:btn-loading={isSendingCampaign[pendingSendCampaign.id]}
                            disabled={!!isSendingCampaign[pendingSendCampaign.id]}
                            on:click={confirmSendCampaign}
                        >
                            <span class="txt">Confirm send</span>
                        </button>
                    </div>
                </div>
            </div>
        {/if}
    {/if}
</PageWrapper>

<style>
    .newsletter-head {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 14px 16px;
    }

    .head-main {
        display: flex;
        align-items: flex-end;
        justify-content: space-between;
        gap: 18px;
        flex-wrap: wrap;
    }

    .summary-title-wrap {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 260px;
    }

    .head-selector {
        width: min(100%, 520px);
    }

    .selector-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .selector-row .input {
        flex: 1 1 auto;
        min-width: 260px;
    }

    .selector-row .btn {
        flex: 0 0 auto;
    }

    .summary-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .summary-pill {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: 999px;
        background: var(--baseAlt1Color);
        color: var(--txtHintColor);
        font-size: 12px;
        padding: 6px 10px;
        white-space: nowrap;
    }

    .summary-hint-pill {
        border-style: dashed;
    }

    .summary-pill i {
        color: var(--txtPrimaryColor);
        opacity: 0.85;
        font-size: 13px;
    }

    .subscriber-controls {
        display: flex;
        align-items: flex-end;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .subscriber-filter-grid {
        display: grid;
        gap: 8px;
        grid-template-columns: repeat(3, minmax(170px, 1fr));
        flex: 1 1 620px;
        min-width: 260px;
    }

    .control-item {
        min-width: 0;
    }

    .bulk-controls {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .bulk-select-all {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        margin-right: 4px;
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

    .campaign-preview {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 12px;
    }

    .preview-tabs .tab-item {
        min-width: 78px;
    }

    .campaign-preview-box {
        margin: 0;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 10px 12px;
        min-height: 110px;
        max-height: 300px;
        overflow: auto;
        white-space: pre-wrap;
        word-break: break-word;
        font-family: inherit;
        font-size: var(--baseFontSize);
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

    .align-end {
        display: flex;
        align-items: flex-end;
    }

    .manual-recipients {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 12px;
    }

    .manual-recipients-grid {
        display: grid;
        gap: 8px;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    }

    .manual-recipient-item {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        padding: 8px 10px;
        background: var(--baseAlt1Color);
    }

    .newsletter-list-item {
        gap: 10px;
        border-radius: var(--baseRadius);
        padding: 10px 12px;
        background: var(--baseAlt1Color);
    }

    .section-head {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .subscriber-title {
        display: inline-flex;
        flex-wrap: wrap;
        align-items: center;
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
        min-width: 110px;
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

    :global(.newsletter-head input:focus),
    :global(.newsletter-head select:focus),
    :global(.newsletter-head textarea:focus),
    :global(.tabs-content input:focus),
    :global(.tabs-content select:focus),
    :global(.tabs-content textarea:focus) {
        box-shadow: 0 0 0 2px color-mix(in srgb, var(--primaryColor) 22%, transparent);
    }

    @media (max-width: 980px) {
        .subscriber-controls {
            align-items: stretch;
        }

        .subscriber-filter-grid {
            grid-template-columns: repeat(2, minmax(180px, 1fr));
            flex-basis: 100%;
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

        .selector-row .input {
            min-width: 0;
        }

        .selector-row .btn {
            width: 100%;
        }

        .summary-pill {
            font-size: 11px;
            padding: 5px 9px;
        }

        .subscriber-filter-grid {
            grid-template-columns: 1fr;
        }

        .bulk-controls {
            width: 100%;
        }

        .bulk-controls .btn {
            width: 100%;
        }

        .pagination-wrap {
            justify-content: center;
        }

        .newsletter-list-item {
            padding: 10px;
        }

        .actions {
            width: 100%;
        }
    }
</style>
