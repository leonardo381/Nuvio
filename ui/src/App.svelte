<script>
    import "./scss/main.scss";

    import tooltip from "@/actions/tooltip";
    import Confirmation from "@/components/base/Confirmation.svelte";
    import TinyMCE from "@/components/base/TinyMCE.svelte";
    import Toasts from "@/components/base/Toasts.svelte";
    import Toggler from "@/components/base/Toggler.svelte";
    import { appName, hideControls, pageTitle } from "@/stores/app";
    import { collections, findCollectionByRequiredNames, loadCollections } from "@/stores/collections";
    import { resetConfirmation } from "@/stores/confirmation";
    import { setErrors } from "@/stores/errors";
    import { superuser } from "@/stores/superuser";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { get } from "svelte/store";
    import { onDestroy, onMount } from "svelte";
    import Router, { link, replace } from "svelte-spa-router";
    import active from "svelte-spa-router/active";
    import routes from "./routes";

    let oldLocation = undefined;

    let showAppSidebar = false;
    let isClientSuperuser = false;
    let canAccessAdminAreas = false;
    $: isClientSuperuser = !!$superuser?.id && ApiClient.isClientSuperuser();
    $: canAccessAdminAreas = !!$superuser?.id && !isClientSuperuser && ApiClient.isAdminSuperuser();

    let isTinyMCEPreloaded = false;
    const leadsContactsCollectionAliases = ["contacts", "contact", "Contacts"];
    const leadsWhatsAppCollectionAliases = [
        "whatsapp",
        "Whatsapp",
        "WhatsApp",
        "whatsapp_interactions",
        "whatsappInteractions",
        "whatsapp_clicks",
    ];
    const bookingAppointmentsCollectionAliases = ["Appointments", "appointments"];
    const sidebarBadgeRefreshEventName = "nuvio:sidebar-badges-refresh";

    let leadsSidebarBadgeCount = 0;
    let bookingSidebarBadgeCount = 0;
    let sidebarBadgeRequestRunId = 0;

    $: if ($superuser?.id) {
        loadSettings();
    }

    function handleRouteLoading(e) {
        if (e?.detail?.location === oldLocation) {
            return; // not an actual change
        }

        showAppSidebar = !!e?.detail?.userData?.showAppSidebar && ApiClient.isSuperuserAuth();
        oldLocation = e?.detail?.location;

        // resets
        $pageTitle = "";
        setErrors({});
        resetConfirmation();

        if (showAppSidebar) {
            refreshSidebarNavBadges();
        }
    }

    function handleRouteFailure() {
        replace("/");
    }

    async function loadSettings() {
        if (!$superuser?.id) {
            return;
        }

        try {
            const settings = await ApiClient.settings.getAll({
                $cancelKey: "initialAppSettings",
            });
            $appName = settings?.meta?.appName || "";
            $hideControls = !!settings?.meta?.hideControls;
        } catch (err) {
            if (!err?.isAbort) {
                console.warn("Failed to load app settings.", err);
            }
        }
    }

    function logout() {
        ApiClient.logout();
    }

    function normalizeLookupKey(value) {
        return `${value || ""}`
            .trim()
            .toLowerCase()
            .replace(/[\s_-]+/g, "");
    }

    function resolveCollectionFieldByAliases(collection, aliases = []) {
        if (!collection || !Array.isArray(collection.fields)) {
            return null;
        }

        const normalizedAliases = (Array.isArray(aliases) ? aliases : [aliases])
            .map((alias) => normalizeLookupKey(alias))
            .filter(Boolean);

        for (const field of collection.fields) {
            const fieldName = `${field?.name || ""}`.trim();
            if (!fieldName) {
                continue;
            }

            if (normalizedAliases.includes(normalizeLookupKey(fieldName))) {
                return field;
            }
        }

        return null;
    }

    function resolveFieldSelectValues(field) {
        const values = field?.options?.values;
        return Array.isArray(values)
            ? values
                .map((value) => `${value || ""}`.trim())
                .filter(Boolean)
            : [];
    }

    function findMatchingSelectValue(values = [], aliases = []) {
        const normalizedAliases = aliases
            .map((value) => normalizeLookupKey(value))
            .filter(Boolean);

        for (const value of values) {
            if (normalizedAliases.includes(normalizeLookupKey(value))) {
                return `${value || ""}`.trim();
            }
        }

        return "";
    }

    function escapeFilterValue(value) {
        return `${value || ""}`
            .replaceAll("\\", "\\\\")
            .replaceAll("\"", "\\\"");
    }

    function resolveCollectionName(collectionList = [], aliases = []) {
        const list = Array.isArray(collectionList) ? collectionList : [];
        const match = findCollectionByRequiredNames(list, aliases);
        return `${match?.name || ""}`.trim();
    }

    function buildStatusCountFilter({ statusFieldName, statusValue, archivedFieldName = "" }) {
        if (!statusFieldName || !statusValue) {
            return "";
        }

        const statusFilter = `${statusFieldName}="${escapeFilterValue(statusValue)}"`;
        if (!archivedFieldName) {
            return statusFilter;
        }

        return `${statusFilter}&&(${archivedFieldName}=""||${archivedFieldName}=null)`;
    }

    async function countCollectionRecords(collectionName, filter, requestKey) {
        if (!collectionName || !filter) {
            return 0;
        }

        try {
            const result = await ApiClient.collection(collectionName).getList(1, 1, {
                filter,
                fields: "id",
                requestKey,
            });

            const total = Number(result?.totalItems || 0);
            return Number.isFinite(total) ? total : 0;
        } catch (err) {
            ApiClient.error(err, false);
            return null;
        }
    }

    async function ensureCollectionsLoadedForBadges() {
        const list = get(collections);
        if (Array.isArray(list) && list.length) {
            return list;
        }

        await loadCollections();
        return get(collections);
    }

    async function countNewLeadsAcrossCollections(collectionList = []) {
        const list = Array.isArray(collectionList) ? collectionList : [];
        const contactsCollection = findCollectionByRequiredNames(list, leadsContactsCollectionAliases);
        const whatsAppCollection = findCollectionByRequiredNames(list, leadsWhatsAppCollectionAliases);

        const contactsStatusField = resolveCollectionFieldByAliases(contactsCollection, ["status"]);
        const whatsAppStatusField = resolveCollectionFieldByAliases(whatsAppCollection, ["status"]);

        const contactsNewStatus = findMatchingSelectValue(resolveFieldSelectValues(contactsStatusField), ["new"]) || "new";
        const whatsAppNewStatus = findMatchingSelectValue(resolveFieldSelectValues(whatsAppStatusField), ["new"]) || "new";

        const contactsFilter = buildStatusCountFilter({
            statusFieldName: contactsStatusField?.name || "status",
            statusValue: contactsNewStatus,
        });
        const whatsAppFilter = buildStatusCountFilter({
            statusFieldName: whatsAppStatusField?.name || "status",
            statusValue: whatsAppNewStatus,
        });

        const contactsCollectionName = resolveCollectionName(list, leadsContactsCollectionAliases);
        const whatsAppCollectionName = resolveCollectionName(list, leadsWhatsAppCollectionAliases);

        const [contactsCount, whatsAppCount] = await Promise.all([
            countCollectionRecords(contactsCollectionName, contactsFilter, "nuvio_sidebar_leads_contacts_badge"),
            countCollectionRecords(whatsAppCollectionName, whatsAppFilter, "nuvio_sidebar_leads_whatsapp_badge"),
        ]);

        if (contactsCount === null && whatsAppCount === null) {
            return null;
        }

        const total = Math.max(0, Number(contactsCount || 0)) + Math.max(0, Number(whatsAppCount || 0));
        return Number.isFinite(total) ? total : 0;
    }

    async function countPendingBookingAcrossCollections(collectionList = []) {
        const list = Array.isArray(collectionList) ? collectionList : [];
        const appointmentsCollection = findCollectionByRequiredNames(list, bookingAppointmentsCollectionAliases);
        if (!appointmentsCollection) {
            return 0;
        }

        const statusField = resolveCollectionFieldByAliases(appointmentsCollection, ["status"]);
        const statusValues = resolveFieldSelectValues(statusField);
        const pendingStatus = findMatchingSelectValue(statusValues, ["pending"]) || "pending";
        const archivedField = resolveCollectionFieldByAliases(appointmentsCollection, ["archivedAt", "archived_at"]);
        const appointmentsCollectionName = resolveCollectionName(list, bookingAppointmentsCollectionAliases);
        const filter = buildStatusCountFilter({
            statusFieldName: statusField?.name || "status",
            statusValue: pendingStatus,
            archivedFieldName: archivedField?.name || "",
        });

        return countCollectionRecords(appointmentsCollectionName, filter, "nuvio_sidebar_booking_badge");
    }

    async function refreshSidebarNavBadges() {
        if (!$superuser?.id || !showAppSidebar || !ApiClient.isSuperuserAuth()) {
            leadsSidebarBadgeCount = 0;
            bookingSidebarBadgeCount = 0;
            return;
        }

        const requestRunId = ++sidebarBadgeRequestRunId;

        try {
            const collectionList = await ensureCollectionsLoadedForBadges();
            const [nextLeadsCount, nextBookingCount] = await Promise.all([
                countNewLeadsAcrossCollections(collectionList),
                countPendingBookingAcrossCollections(collectionList),
            ]);

            if (requestRunId !== sidebarBadgeRequestRunId) {
                return;
            }

            if (Number.isFinite(nextLeadsCount)) {
                leadsSidebarBadgeCount = Math.max(0, nextLeadsCount);
            }

            if (Number.isFinite(nextBookingCount)) {
                bookingSidebarBadgeCount = Math.max(0, nextBookingCount);
            }
        } catch (err) {
            ApiClient.error(err, false);
        }
    }

    function formatSidebarBadgeCount(count) {
        const normalized = Number(count || 0);
        if (!Number.isFinite(normalized) || normalized <= 0) {
            return "";
        }

        if (normalized > 99) {
            return "99+";
        }

        return `${Math.trunc(normalized)}`;
    }

    function shouldShowSidebarBadge(count) {
        return Number(count || 0) > 0;
    }

    function handleSidebarBadgeRefreshEvent() {
        refreshSidebarNavBadges();
    }

    onMount(() => {
        window.addEventListener(sidebarBadgeRefreshEventName, handleSidebarBadgeRefreshEvent);
        if (ApiClient.isSuperuserAuth()) {
            refreshSidebarNavBadges();
        }
    });

    onDestroy(() => {
        window.removeEventListener(sidebarBadgeRefreshEventName, handleSidebarBadgeRefreshEvent);
    });
</script>

<svelte:head>
    <title>{CommonHelper.joinNonEmpty([$pageTitle, $appName, "PocketBase"], " - ", false)}</title>

    {#if window.location.protocol == "https:"}
        <link
            rel="shortcut icon"
            type="image/png"
            href="{import.meta.env.BASE_URL}images/favicon/favicon_prod.png"
        />
    {/if}
</svelte:head>

<div class="app-layout">
    {#if $superuser?.id && showAppSidebar}
        <aside class="app-sidebar">
            <a href="/" class="logo logo-sm" use:link>
                <img
                    src="{import.meta.env.BASE_URL}images/logo.svg"
                    alt="PocketBase logo"
                    width="40"
                    height="40"
                />
            </a>

            <nav class="main-menu">
                <!-- NUVIO CUSTOM START: Dedicated CMS section entry in app sidebar. -->
                <a
                    href="/cms"
                    class="menu-item"
                    aria-label="CMS"
                    use:link
                    use:active={{ path: "/cms/?.*", className: "current-route" }}
                    use:tooltip={{ text: "CMS", position: "right" }}
                >
                    <i class="ri-layout-grid-line" />
                </a>
                <!-- NUVIO CUSTOM END: Dedicated CMS section entry in app sidebar. -->
                <!-- NUVIO CUSTOM START: Dedicated Leads section entry in app sidebar. -->
                <a
                    href="/leads"
                    class="menu-item"
                    aria-label="Leads"
                    use:link
                    use:active={{ path: "/leads/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Leads", position: "right" }}
                >
                    <i class="ri-mail-line" />
                    {#if shouldShowSidebarBadge(leadsSidebarBadgeCount)}
                        <span class="menu-item-badge">{formatSidebarBadgeCount(leadsSidebarBadgeCount)}</span>
                    {/if}
                </a>
                <!-- NUVIO CUSTOM END: Dedicated Leads section entry in app sidebar. -->
                <!-- NUVIO CUSTOM START: Dedicated Booking section entry in app sidebar. -->
                <a
                    href="/booking"
                    class="menu-item"
                    aria-label="Booking"
                    use:link
                    use:active={{ path: "/booking/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Booking", position: "right" }}
                >
                    <i class="ri-calendar-line" />
                    {#if shouldShowSidebarBadge(bookingSidebarBadgeCount)}
                        <span class="menu-item-badge">{formatSidebarBadgeCount(bookingSidebarBadgeCount)}</span>
                    {/if}
                </a>
                <!-- NUVIO CUSTOM END: Dedicated Booking section entry in app sidebar. -->
                <!-- NUVIO CUSTOM START: Dedicated Newsletter section entry in app sidebar. -->
                <a
                    href="/newsletter"
                    class="menu-item"
                    aria-label="Newsletter"
                    use:link
                    use:active={{ path: "/newsletter/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Newsletter", position: "right" }}
                >
                    <i class="ri-megaphone-line" />
                </a>
                <!-- NUVIO CUSTOM END: Dedicated Newsletter section entry in app sidebar. -->
                <!-- NUVIO CUSTOM START: Dedicated Reports section entry in app sidebar. -->
                <a
                    href="/reports"
                    class="menu-item"
                    aria-label="Reports"
                    use:link
                    use:active={{ path: "/reports/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Reports", position: "right" }}
                >
                    <i class="ri-bar-chart-grouped-line" />
                </a>
                <!-- NUVIO CUSTOM END: Dedicated Reports section entry in app sidebar. -->
                {#if canAccessAdminAreas}
                    <a
                        href="/collections"
                        class="menu-item"
                        aria-label="Collections"
                        use:link
                        use:active={{ path: "/collections/?.*", className: "current-route" }}
                        use:tooltip={{ text: "Collections", position: "right" }}
                    >
                        <i class="ri-database-2-line" />
                    </a>
                    <a
                        href="/logs"
                        class="menu-item"
                        aria-label="Logs"
                        use:link
                        use:active={{ path: "/logs/?.*", className: "current-route" }}
                        use:tooltip={{ text: "Logs", position: "right" }}
                    >
                        <i class="ri-line-chart-line" />
                    </a>
                    <a
                        href="/settings"
                        class="menu-item"
                        aria-label="Settings"
                        use:link
                        use:active={{ path: "/settings/?.*", className: "current-route" }}
                        use:tooltip={{ text: "Settings", position: "right" }}
                    >
                        <i class="ri-tools-line" />
                    </a>
                {/if}
            </nav>

            <div
                tabindex="0"
                role="button"
                aria-label="Logged superuser menu"
                class="thumb thumb-circle link-hint"
                title={$superuser.email}
            >
                <span class="initials">{CommonHelper.getInitials($superuser.email)}</span>
                <Toggler class="dropdown dropdown-nowrap dropdown-upside dropdown-left">
                    <div class="txt-ellipsis current-superuser" title={$superuser.email}>
                        {$superuser.email}
                    </div>
                    <hr />
                    {#if canAccessAdminAreas}
                        <a
                            href="/collections?collection=_superusers"
                            class="dropdown-item closable"
                            role="menuitem"
                            use:link
                        >
                            <i class="ri-shield-user-line" aria-hidden="true" />
                            <span class="txt">Manage superusers</span>
                        </a>
                    {/if}
                    <button type="button" class="dropdown-item closable" role="menuitem" on:click={logout}>
                        <i class="ri-logout-circle-line" aria-hidden="true" />
                        <span class="txt">Logout</span>
                    </button>
                </Toggler>
            </div>
        </aside>
    {/if}

    <div class="app-body">
        <Router {routes} on:routeLoading={handleRouteLoading} on:conditionsFailed={handleRouteFailure} />

        <Toasts />
    </div>
</div>

<Confirmation />

{#if $superuser?.id && showAppSidebar && !isTinyMCEPreloaded}
    <div class="tinymce-preloader hidden">
        <TinyMCE
            conf={CommonHelper.defaultEditorOptions()}
            on:init={() => {
                isTinyMCEPreloaded = true;
            }}
        />
    </div>
{/if}

<style>
    .current-superuser {
        padding: 10px;
        max-width: 200px;
        color: var(--txtHintColor);
    }

    .menu-item-badge {
        position: absolute;
        top: 3px;
        right: 3px;
        min-width: 18px;
        height: 18px;
        padding: 0 5px;
        border-radius: 999px;
        border: 1px solid color-mix(in srgb, var(--dangerColor) 42%, var(--baseColor));
        display: inline-flex;
        align-items: center;
        justify-content: center;
        background: color-mix(in srgb, var(--dangerAltColor) 72%, var(--baseColor));
        color: color-mix(in srgb, var(--dangerColor) 82%, var(--txtPrimaryColor));
        font-size: 10px;
        line-height: 1;
        font-weight: 700;
        pointer-events: none;
    }
</style>
