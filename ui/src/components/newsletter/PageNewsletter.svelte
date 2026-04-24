<script>
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addSuccessToast } from "@/stores/toasts";

    // NUVIO CUSTOM START: Newsletter V1 dedicated section/page (collection-backed).
    $pageTitle = "Newsletter";

    const subscriberStatuses = ["pending", "active", "unsubscribed"];
    const campaignRecipientsTypeOptions = ["all", "manual"];

    let activeSection = "subscribers";
    let websites = [];
    let selectedWebsiteId = "";

    let subscribers = [];
    let campaigns = [];

    let isLoadingWebsites = false;
    let isLoadingSubscribers = false;
    let isLoadingCampaigns = false;
    let isCreatingSubscriber = false;
    let isCreatingCampaign = false;
    let isSendingCampaign = {};

    let subscriberForm = {
        email: "",
        status: "pending",
    };

    let campaignForm = {
        subject: "",
        body: "",
        recipientsType: "all",
        recipientsIds: [],
    };

    let lastWebsitesCollectionId = "";
    let lastDataKey = "";

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

    $: activeSubscribers = subscribers.filter((subscriber) => {
        return `${subscriber?.status || ""}`.toLowerCase() === "active";
    });
    $: sentCampaigns = campaigns.filter((campaign) => `${campaign?.status || ""}`.toLowerCase() === "sent");

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

    function isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalizeEmail(email));
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

    function isManualRecipientSelected(subscriberId) {
        return campaignForm.recipientsIds.includes(subscriberId);
    }

    function toggleManualRecipient(subscriberId) {
        if (campaignForm.recipientsIds.includes(subscriberId)) {
            campaignForm.recipientsIds = campaignForm.recipientsIds.filter((id) => id !== subscriberId);
        } else {
            campaignForm.recipientsIds = [...campaignForm.recipientsIds, subscriberId];
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
            websites = await ApiClient.collection(websitesCollection.id).getFullList({
                sort: resolveWebsitesSort(websitesCollection),
                requestKey: "nuvio_newsletter_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                subscribers = [];
                campaigns = [];
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
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadSubscribers() {
        if (!hasNewsletterCollections || !selectedWebsiteId) {
            subscribers = [];
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

    async function createSubscriber() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingSubscriber) {
            return;
        }

        const email = normalizeEmail(subscriberForm.email);
        if (!isValidEmail(email)) {
            ApiClient.error(new Error("Please provide a valid subscriber email."));
            return;
        }

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
            const payload = {
                status,
            };

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

    async function createCampaignDraft() {
        if (!hasNewsletterCollections || !selectedWebsiteId || isCreatingCampaign) {
            return;
        }

        const subject = `${campaignForm.subject || ""}`.trim();
        const body = `${campaignForm.body || ""}`.trim();

        if (!subject || !body) {
            ApiClient.error(new Error("Campaign subject and body are required."));
            return;
        }

        if (campaignForm.recipientsType === "manual" && !campaignForm.recipientsIds.length) {
            ApiClient.error(new Error("Select at least one manual recipient."));
            return;
        }

        isCreatingCampaign = true;

        try {
            await ApiClient.collection(campaignsCollection.id).create({
                website: selectedWebsiteId,
                subject,
                body,
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

            await loadCampaigns();
            addSuccessToast("Draft campaign created.");
        } catch (err) {
            ApiClient.error(err);
        }

        isCreatingCampaign = false;
    }

    async function sendCampaign(campaign) {
        if (!campaign?.id || isSendingCampaign[campaign.id]) {
            return;
        }

        isSendingCampaign[campaign.id] = true;
        isSendingCampaign = { ...isSendingCampaign };

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
        } catch (err) {
            ApiClient.error(err);
        }

        delete isSendingCampaign[campaign.id];
        isSendingCampaign = { ...isSendingCampaign };
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
                            on:change={(e) => {
                                selectedWebsiteId = e.target.value || "";
                            }}
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
                    <i class="ri-megaphone-line" />
                    {campaigns.length} campaigns
                </span>
                <span class="summary-pill">
                    <i class="ri-send-plane-2-line" />
                    {sentCampaigns.length} sent
                </span>
                <span class="summary-pill summary-hint-pill">Select a website to scope all actions</span>
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
            </div>
        {:else}
            <div class="tabs">
                <div class="tabs-header compact combined left">
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "subscribers"}
                        on:click={() => (activeSection = "subscribers")}
                    >
                        Subscribers
                    </button>
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeSection === "campaigns"}
                        on:click={() => (activeSection = "campaigns")}
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
                                        type="email"
                                        class="input"
                                        placeholder="name@example.com"
                                        bind:value={subscriberForm.email}
                                    />
                                </div>
                                <div class="col-md-3">
                                    <label class="txt-sm txt-hint block m-b-5" for="subscriber-status">Status</label>
                                    <select
                                        id="subscriber-status"
                                        class="input"
                                        bind:value={subscriberForm.status}
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
                                        disabled={isCreatingSubscriber}
                                    >
                                        <span class="txt">Add subscriber</span>
                                    </button>
                                </div>
                            </form>
                        </section>

                        <section class="panel">
                            <div class="flex m-b-sm">
                                <h4 class="m-0">Subscribers</h4>
                                <div class="flex-fill" />
                                <span class="txt-sm txt-hint">{subscribers.length} total | {activeSubscribers.length} active</span>
                            </div>

                            {#if isLoadingSubscribers}
                                <div class="loading-state">
                                    <span class="loader loader-sm" />
                                    <span class="txt-hint">Loading subscribers...</span>
                                </div>
                            {:else if !subscribers.length}
                                <div class="empty-state">No subscribers yet for this website.</div>
                            {:else}
                                <div class="list list-compact">
                                    <div class="list-content">
                                        {#each subscribers as subscriber (subscriber.id)}
                                            <div class="list-item newsletter-list-item">
                                                <div class="content">
                                                    <div class="subscriber-title">
                                                        <span class="txt">{subscriber.email}</span>
                                                        <span
                                                            class="status-chip"
                                                            class:is-active={subscriber.status === "active"}
                                                            class:is-pending={subscriber.status === "pending"}
                                                            class:is-unsubscribed={subscriber.status === "unsubscribed"}
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
                                                    {#if subscriber.status !== "active"}
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs btn-outline action-btn"
                                                            on:click={() => setSubscriberStatus(subscriber, "active")}
                                                        >
                                                            <span class="txt">Mark active</span>
                                                        </button>
                                                    {/if}
                                                    {#if subscriber.status !== "unsubscribed"}
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
                            {/if}
                        </section>
                    {:else}
                        <section class="panel m-b-base">
                            <div class="section-head m-b-sm">
                                <h4 class="m-0">Create campaign draft</h4>
                                <p class="txt-sm txt-hint m-b-0">Build your message, choose recipients, then send when ready.</p>
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
                                        disabled={isCreatingCampaign}
                                    >
                                        <span class="txt">Create draft</span>
                                    </button>
                                </div>
                            </form>

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
                                <div class="empty-state">No campaigns yet for this website.</div>
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
                                                            class:is-active={campaign.status === "sent"}
                                                            class:is-pending={campaign.status === "draft"}
                                                            class:is-unsubscribed={campaign.status !== "sent" && campaign.status !== "draft"}
                                                        >
                                                            {campaign.status}
                                                        </span>
                                                    </div>
                                                    <div class="txt-xs txt-hint meta-line">
                                                        Recipients: {campaign.recipientsType}
                                                        <span class="meta-sep">|</span>
                                                        Sent count: {campaign.recipientsCount || 0}
                                                        <span class="meta-sep">|</span>
                                                        Sent at: {formatDateTime(campaign.sentAt)}
                                                        <span class="meta-sep">|</span>
                                                        Created: {formatDateTime(campaign.created)}
                                                    </div>
                                                </div>
                                                <div class="actions">
                                                    {#if campaign.status !== "sent"}
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs btn-strong"
                                                            class:btn-loading={isSendingCampaign[campaign.id]}
                                                            disabled={!!isSendingCampaign[campaign.id]}
                                                            on:click={() => sendCampaign(campaign)}
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

        .newsletter-list-item {
            padding: 10px;
        }
    }
</style>
