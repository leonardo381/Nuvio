<script>
    import { querystring } from "svelte-spa-router";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import SchemaForm from "@/components/base/nuvio/schema/SchemaForm.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addSuccessToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { getWebsiteSettingsSchemaForRole, normalizeWebsiteSettingsValue } from "@/utils/WebsiteSettingsSchema";

    $pageTitle = "Website Content";
    loadCollections();

    const initialQueryParams = new URLSearchParams($querystring);
    const cmsTabPagesKey = "pages";
    const cmsTabSettingsKey = "settings";
    const clientSettingsRole = "client";
    const visibleClientSettingsKeys = new Set(["whatsapp", "contactForm", "newsletter", "i18n"]);

    let websites = [];
    let pages = [];
    let blocks = [];
    let components = [];

    let selectedWebsiteId = initialQueryParams.get("cmsWebsite") || "";
    let selectedPageId = initialQueryParams.get("cmsPage") || "";
    let activeCmsTab = initialQueryParams.get("cmsTab") === cmsTabSettingsKey ? cmsTabSettingsKey : cmsTabPagesKey;
    let openSectionId = "";

    let isLoadingWebsites = false;
    let isLoadingPages = false;
    let isLoadingBlocks = false;
    let isLoadingComponents = false;
    let isSavingPage = false;
    let isSavingWebsiteSettings = false;
    let isSavingWebsiteIdentitySeo = false;

    let pageError = "";
    let websiteSettingsError = "";
    let websiteIdentitySeoError = "";
    let pageEditForm = {
        title: "",
        enabled: true,
        seoTitle: "",
        seoDescription: "",
    };

    let websiteSettingsFullDraft = {};
    let websiteSettingsDraft = {};
    let websiteIdentitySeoDraft = {
        seoTitle: "",
        seoDescription: "",
        logoCurrent: "",
        seoImageCurrent: "",
        logoFile: null,
        seoImageFile: null,
    };

    let sectionPropsDraftById = {};
    let sectionErrorById = {};
    let isSavingSectionById = {};
    let pagePreviewReloadToken = 0;

    let lastCollectionsKey = "";
    let lastPersistedContextKey = "";
    let lastPageSeedId = "";
    let lastSectionsSeedKey = "";
    let lastWebsiteSettingsSeedKey = "";

    $: websitesCollection = findCollection("websites");
    $: pagesCollection = findCollection("pages");
    $: blocksCollection = findCollection("blocks");
    $: componentsCollection = findCollection("components");

    $: hasCmsCollections = !!websitesCollection?.id && !!pagesCollection?.id && !!blocksCollection?.id;

    $: missingCollections = [];
    $: if (!websitesCollection?.id) {
        missingCollections.push("websites");
    }
    $: if (!pagesCollection?.id) {
        missingCollections.push("pages");
    }
    $: if (!blocksCollection?.id) {
        missingCollections.push("blocks");
    }

    $: websiteNameField = resolveFieldName(websitesCollection, ["name", "title", "label"]);
    $: websiteSlugField = resolveFieldName(websitesCollection, ["slug"]);
    $: websiteDomainField = resolveFieldName(websitesCollection, ["domain", "url", "host"]);
    $: websiteSettingsField = resolveFieldName(websitesCollection, ["settings"]);
    $: hasWebsiteSettingsField = !!websiteSettingsField;

    $: pageTitleField = resolveFieldName(pagesCollection, ["title", "name", "label"]);
    $: pageSlugField = resolveFieldName(pagesCollection, ["slug"]);
    $: pageEnabledField = resolveFieldName(pagesCollection, ["enabled", "published", "active"]);
    $: pageSeoTitleField = resolveFieldName(pagesCollection, ["seo_title", "seoTitle"]);
    $: pageSeoDescriptionField = resolveFieldName(pagesCollection, ["seo_description", "seoDescription"]);

    $: websiteLogoField = resolveFieldName(websitesCollection, ["logo"]);
    $: websiteSeoTitleField = resolveFieldName(websitesCollection, ["seoTitle", "seo_title"]);
    $: websiteSeoDescriptionField = resolveFieldName(websitesCollection, ["seoDescription", "seo_description"]);
    $: websiteSeoImageField = resolveFieldName(websitesCollection, ["seoImage", "seo_image"]);
    $: hasWebsiteIdentitySeoFields = !!(
        websiteLogoField ||
        websiteSeoTitleField ||
        websiteSeoDescriptionField ||
        websiteSeoImageField
    );

    $: blockTitleField = resolveFieldName(blocksCollection, ["title", "name", "label"]);
    $: blockPropsField = resolveFieldName(blocksCollection, ["props"]);
    $: blockComponentKeyField = resolveFieldName(blocksCollection, ["component_key", "componentKey"]);
    $: blockComponentRelationField = resolveRelationFieldName(blocksCollection, componentsCollection?.id, ["component"]);
    $: blockPageRelationField = resolveRelationFieldName(blocksCollection, pagesCollection?.id, ["page"]);

    $: componentKeyField = resolveFieldName(componentsCollection, ["key", "component_key"]);
    $: componentNameField = resolveFieldName(componentsCollection, ["name", "title", "label"]);
    $: componentSchemaField = resolveFieldName(componentsCollection, ["schema"]);

    $: pageWebsiteRelationField = resolveRelationFieldName(pagesCollection, websitesCollection?.id, ["website"]);

    $: selectedWebsite = websites.find((record) => record.id === selectedWebsiteId) || null;
    $: selectedPage = pages.find((record) => record.id === selectedPageId) || null;
    $: roleScopedSettingsFields = getWebsiteSettingsSchemaForRole(clientSettingsRole, websiteSettingsFullDraft).fields;
    $: clientWebsiteSettingsFields = filterClientWebsiteSettingsFields(roleScopedSettingsFields);

    $: websitePublicUrl = getWebsitePublicUrl(selectedWebsite);
    $: selectedPagePath = getPagePath(selectedPage);
    $: selectedWebsiteSlug = normalizeString(websiteSlugField ? selectedWebsite?.[websiteSlugField] : "");
    $: selectedPageSlug = normalizeString(pageSlugField ? selectedPage?.[pageSlugField] : "");
    $: pagePreviewUrl = buildPagePreviewUrl(selectedWebsiteSlug, selectedPageSlug);
    $: pagePreviewIframeSrc = buildPreviewIframeSrc(pagePreviewUrl, pagePreviewReloadToken);

    $: componentsById = new Map(components.map((record) => [record.id, record]));
    $: componentsByKey = new Map(
        components
            .map((record) => {
                const key = normalizeString(componentKeyField ? record?.[componentKeyField] : "");
                return [key, record];
            })
            .filter(([key]) => !!key),
    );

    $: if (selectedPage?.id && selectedPage.id !== lastPageSeedId) {
        lastPageSeedId = selectedPage.id;
        pageEditForm = {
            title: pageTitleField ? `${selectedPage?.[pageTitleField] || ""}` : "",
            enabled: pageEnabledField ? !!selectedPage?.[pageEnabledField] : true,
            seoTitle: pageSeoTitleField ? `${selectedPage?.[pageSeoTitleField] || ""}` : "",
            seoDescription: pageSeoDescriptionField ? `${selectedPage?.[pageSeoDescriptionField] || ""}` : "",
        };
        pageError = "";
    }

    $: if (!selectedPage?.id) {
        lastPageSeedId = "";
    }

    $: if (!hasCmsCollections) {
        websites = [];
        pages = [];
        blocks = [];
        components = [];
        selectedWebsiteId = "";
        selectedPageId = "";
        openSectionId = "";
        lastCollectionsKey = "";
    } else {
        const nextKey = `${websitesCollection?.id || ""}|${pagesCollection?.id || ""}|${blocksCollection?.id || ""}|${componentsCollection?.id || ""}`;
        if (nextKey !== lastCollectionsKey) {
            lastCollectionsKey = nextKey;
            reload();
        }
    }

    $: if (hasCmsCollections) {
        const nextContextKey = `${selectedWebsiteId || ""}|${selectedPageId || ""}|${activeCmsTab || ""}`;
        if (nextContextKey !== lastPersistedContextKey) {
            lastPersistedContextKey = nextContextKey;
            CommonHelper.replaceHashQueryParams({
                cmsWebsite: selectedWebsiteId || null,
                cmsPage: selectedPageId || null,
                cmsTab: activeCmsTab === cmsTabSettingsKey ? cmsTabSettingsKey : null,
            });
        }
    }

    $: {
        const nextSeedKey = `${blockPropsField || ""}|${blocks.map((block) => `${block.id}:${block?.updated || ""}`).join("|")}`;
        if (nextSeedKey !== lastSectionsSeedKey) {
            lastSectionsSeedKey = nextSeedKey;

            const nextDraft = {};
            const nextErrors = {};
            const nextSaving = {};

            for (const block of blocks) {
                nextDraft[block.id] = toPropsObject(blockPropsField ? block?.[blockPropsField] : {});
                nextErrors[block.id] = "";
                nextSaving[block.id] = false;
            }

            sectionPropsDraftById = nextDraft;
            sectionErrorById = nextErrors;
            isSavingSectionById = nextSaving;

            if (!openSectionId || !blocks.some((block) => block.id === openSectionId)) {
                openSectionId = blocks[0]?.id || "";
            }
        }
    }

    $: {
        const nextWebsiteSettingsSeedKey = `${selectedWebsite?.id || ""}|${selectedWebsite?.updated || ""}|${websiteSettingsField || ""}`;
        if (nextWebsiteSettingsSeedKey !== lastWebsiteSettingsSeedKey) {
            lastWebsiteSettingsSeedKey = nextWebsiteSettingsSeedKey;
            initializeWebsiteSettingsDraft();
            initializeWebsiteIdentitySeoDraft();
        }
    }

    function findCollection(name) {
        return $collections.find((item) => `${item?.name || ""}`.toLowerCase() === `${name || ""}`.toLowerCase()) || null;
    }

    function getCollectionFieldNames(collection) {
        return new Set((collection?.fields || []).map((field) => `${field?.name || ""}`.toLowerCase()));
    }

    function resolveFieldName(collection, candidates = []) {
        const fieldNames = getCollectionFieldNames(collection);
        for (const candidate of candidates) {
            if (fieldNames.has(`${candidate || ""}`.toLowerCase())) {
                return candidate;
            }
        }
        return "";
    }

    function resolveRelationFieldName(collection, targetCollectionId = "", fallbackCandidates = []) {
        const fields = collection?.fields || [];
        const relationByTarget = fields.find(
            (field) => field?.type === "relation" && `${field.collectionId || ""}` === `${targetCollectionId || ""}`,
        );

        if (relationByTarget?.name) {
            return relationByTarget.name;
        }

        const normalizedFallback = fallbackCandidates.map((item) => `${item || ""}`.toLowerCase());
        const relationByName = fields.find(
            (field) => field?.type === "relation" && normalizedFallback.includes(`${field?.name || ""}`.toLowerCase()),
        );

        return relationByName?.name || "";
    }

    function resolveSort(collection, candidates = []) {
        const fieldNames = getCollectionFieldNames(collection);
        const valid = candidates.filter((candidate) => fieldNames.has(`${candidate || ""}`.toLowerCase()));
        if (!valid.length) {
            return "+id";
        }
        return valid.map((field) => `+${field}`).join(",");
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function toPropsObject(value) {
        if (value && typeof value === "object" && !Array.isArray(value)) {
            return structuredClone(value);
        }

        if (typeof value === "string" && value.trim()) {
            try {
                const parsed = JSON.parse(value);
                if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
                    return parsed;
                }
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function isPlainObject(value) {
        return !!value && typeof value === "object" && !Array.isArray(value);
    }

    function mergeSettingsObjects(baseValue, patchValue) {
        if (!isPlainObject(baseValue) || !isPlainObject(patchValue)) {
            return structuredClone(patchValue);
        }

        const merged = { ...baseValue };

        for (const [key, value] of Object.entries(patchValue)) {
            if (isPlainObject(value) && isPlainObject(baseValue[key])) {
                merged[key] = mergeSettingsObjects(baseValue[key], value);
            } else {
                merged[key] = structuredClone(value);
            }
        }

        return merged;
    }

    function parseSchemaFields(rawSchema) {
        let schema = rawSchema;
        if (typeof schema === "string") {
            try {
                schema = JSON.parse(schema);
            } catch (_) {
                return [];
            }
        }

        if (!schema || typeof schema !== "object") {
            return [];
        }

        return Array.isArray(schema.fields) ? schema.fields : [];
    }

    function setPayloadField(payload, fieldName, value) {
        if (!fieldName) {
            return;
        }
        payload[fieldName] = value;
    }

    function getWebsiteLabel(record) {
        return `${CommonHelper.displayValue(record || {}, [websiteNameField, websiteSlugField, websiteDomainField], "") || ""}`.trim() || record?.id || "-";
    }

    function getPageLabel(record) {
        return `${CommonHelper.displayValue(record || {}, [pageTitleField, pageSlugField], "") || ""}`.trim() || record?.id || "-";
    }

    function getPagePath(record) {
        const slug = normalizeString(pageSlugField ? record?.[pageSlugField] : "");
        if (!slug) {
            return "/";
        }
        return slug.startsWith("/") ? slug : `/${slug}`;
    }

    function getWebsitePublicUrl(record) {
        const domain = normalizeString(websiteDomainField ? record?.[websiteDomainField] : "");
        if (!domain) {
            return "";
        }
        if (/^https?:\/\//i.test(domain)) {
            return domain;
        }
        return `https://${domain}`;
    }

    function getConfiguredPublicBaseUrl() {
        const explicitPublicBase = normalizeString(import.meta.env?.VITE_PUBLIC_SITE_BASE_URL);
        if (explicitPublicBase) {
            return explicitPublicBase;
        }

        return normalizeString(import.meta.env?.VITE_SITE_BASE_URL);
    }

    function normalizeBaseUrl(value, options = {}) {
        const { allowSingleLabelHost = true } = options;
        const input = normalizeString(value);
        if (!input) {
            return "";
        }

        const candidate = /^https?:\/\//i.test(input) ? input : `https://${input}`;
        try {
            const parsed = new URL(candidate);
            if (!/^https?:$/i.test(parsed.protocol)) {
                return "";
            }

            const hostname = normalizeString(parsed.hostname).toLowerCase();
            const isLocalHost =
                hostname === "localhost" ||
                hostname === "127.0.0.1" ||
                hostname === "::1" ||
                hostname === "[::1]";
            const isIpv4 = /^\d{1,3}(?:\.\d{1,3}){3}$/.test(hostname);
            const isSingleLabelHost = !!hostname && !hostname.includes(".") && !isLocalHost && !isIpv4;
            const isKnownPlaceholder = new Set(["test", "example", "invalid", "placeholder", "your-domain", "domain"]).has(
                hostname,
            );

            if (!allowSingleLabelHost && (isSingleLabelHost || isKnownPlaceholder)) {
                return "";
            }

            return parsed.origin;
        } catch (_) {
            return "";
        }
    }

    function buildPagePreviewUrl(websiteSlug, pageSlug) {
        const normalizedWebsiteSlug = normalizeString(websiteSlug);
        const normalizedPageSlug = normalizeString(pageSlug);

        if (!normalizedWebsiteSlug || !normalizedPageSlug) {
            return "";
        }

        const configuredBase = normalizeBaseUrl(getConfiguredPublicBaseUrl());
        const websiteBase = normalizeBaseUrl(getWebsitePublicUrl(selectedWebsite), { allowSingleLabelHost: false });
        const baseUrl = configuredBase || websiteBase;

        if (!baseUrl) {
            return "";
        }

        return `${baseUrl}/site/${encodeURIComponent(normalizedWebsiteSlug)}/${encodeURIComponent(normalizedPageSlug)}`;
    }

    function buildPreviewIframeSrc(url, reloadToken) {
        const previewUrl = normalizeString(url);
        if (!previewUrl) {
            return "";
        }

        try {
            const parsed = new URL(previewUrl);
            if (reloadToken) {
                parsed.searchParams.set("_cmsPreview", `${reloadToken}`);
            }
            return parsed.toString();
        } catch (_) {
            return "";
        }
    }

    function refreshPagePreview() {
        pagePreviewReloadToken = Date.now();
    }

    function getExpandedComponent(block) {
        if (!blockComponentRelationField) {
            return null;
        }

        const expanded = block?.expand?.[blockComponentRelationField];
        if (Array.isArray(expanded)) {
            return expanded[0] || null;
        }
        return expanded || null;
    }

    function getBlockComponentId(block) {
        if (!blockComponentRelationField) {
            return "";
        }

        const relationValue = block?.[blockComponentRelationField];
        if (Array.isArray(relationValue)) {
            return normalizeString(relationValue[0]);
        }
        if (typeof relationValue === "string") {
            return normalizeString(relationValue);
        }

        const expanded = getExpandedComponent(block);
        return normalizeString(expanded?.id);
    }

    function getBlockComponentKey(block) {
        if (blockComponentKeyField) {
            const rawKey = normalizeString(block?.[blockComponentKeyField]);
            if (rawKey) {
                return rawKey;
            }
        }

        const expanded = getExpandedComponent(block);
        if (expanded && componentKeyField) {
            const expandedKey = normalizeString(expanded?.[componentKeyField]);
            if (expandedKey) {
                return expandedKey;
            }
        }

        const componentId = getBlockComponentId(block);
        if (componentId && componentsById.has(componentId) && componentKeyField) {
            return normalizeString(componentsById.get(componentId)?.[componentKeyField]);
        }

        return "";
    }

    function getComponentForBlock(block) {
        const key = getBlockComponentKey(block);
        if (key && componentsByKey.has(key)) {
            return componentsByKey.get(key);
        }

        const componentId = getBlockComponentId(block);
        if (componentId && componentsById.has(componentId)) {
            return componentsById.get(componentId);
        }

        return null;
    }

    function getSectionSchemaFields(block) {
        const component = getComponentForBlock(block);
        if (!component || !componentSchemaField) {
            return [];
        }
        return parseSchemaFields(component?.[componentSchemaField]);
    }

    function getSectionTitle(block, index) {
        const title = normalizeString(blockTitleField ? block?.[blockTitleField] : "");
        if (title) {
            return title;
        }

        const component = getComponentForBlock(block);
        const componentName = normalizeString(componentNameField ? component?.[componentNameField] : "");
        if (componentName) {
            return componentName;
        }

        return `Section ${index + 1}`;
    }

    function getSectionDescription(block) {
        const component = getComponentForBlock(block);
        const componentName = normalizeString(componentNameField ? component?.[componentNameField] : "");
        if (componentName) {
            return componentName;
        }
        return "Edit section content.";
    }

    function toggleSection(blockId) {
        if (openSectionId === blockId) {
            openSectionId = "";
            return;
        }
        openSectionId = blockId;
    }

    function updateSectionDraft(blockId, nextValue) {
        sectionPropsDraftById = {
            ...sectionPropsDraftById,
            [blockId]: toPropsObject(nextValue),
        };
    }

    function filterClientWebsiteSettingsFields(fields = []) {
        return (fields || []).filter((field) => visibleClientSettingsKeys.has(field?.key));
    }

    function toSingleFileName(value) {
        if (Array.isArray(value)) {
            return normalizeString(value[0]);
        }
        return normalizeString(value);
    }

    function initializeWebsiteSettingsDraft() {
        websiteSettingsError = "";

        if (!selectedWebsite || !hasWebsiteSettingsField) {
            websiteSettingsFullDraft = {};
            websiteSettingsDraft = {};
            return;
        }

        const normalizedFullSettings = normalizeWebsiteSettingsValue(selectedWebsite?.[websiteSettingsField]);
        const roleScopedFields = getWebsiteSettingsSchemaForRole(clientSettingsRole, normalizedFullSettings).fields;
        const visibleFields = filterClientWebsiteSettingsFields(roleScopedFields);

        websiteSettingsFullDraft = normalizedFullSettings;
        websiteSettingsDraft = normalizeWebsiteSettingsValue(websiteSettingsFullDraft, visibleFields);
    }

    function initializeWebsiteIdentitySeoDraft() {
        websiteIdentitySeoError = "";

        if (!selectedWebsite) {
            websiteIdentitySeoDraft = {
                seoTitle: "",
                seoDescription: "",
                logoCurrent: "",
                seoImageCurrent: "",
                logoFile: null,
                seoImageFile: null,
            };
            return;
        }

        websiteIdentitySeoDraft = {
            seoTitle: websiteSeoTitleField ? `${selectedWebsite?.[websiteSeoTitleField] || ""}` : "",
            seoDescription: websiteSeoDescriptionField ? `${selectedWebsite?.[websiteSeoDescriptionField] || ""}` : "",
            logoCurrent: websiteLogoField ? toSingleFileName(selectedWebsite?.[websiteLogoField]) : "",
            seoImageCurrent: websiteSeoImageField ? toSingleFileName(selectedWebsite?.[websiteSeoImageField]) : "",
            logoFile: null,
            seoImageFile: null,
        };
    }

    function handleWebsiteSettingsChange(event) {
        if (!hasWebsiteSettingsField) {
            return;
        }

        const nextValue = event.detail?.value ?? event.detail ?? {};
        const roleScopedFields = getWebsiteSettingsSchemaForRole(clientSettingsRole, websiteSettingsFullDraft).fields;
        const visibleFields = filterClientWebsiteSettingsFields(roleScopedFields);
        const normalizedScopedChanges = normalizeWebsiteSettingsValue(nextValue, visibleFields);
        const normalizedFullSettings = normalizeWebsiteSettingsValue(
            mergeSettingsObjects(websiteSettingsFullDraft, normalizedScopedChanges),
        );
        const nextRoleScopedFields = getWebsiteSettingsSchemaForRole(clientSettingsRole, normalizedFullSettings).fields;
        const nextVisibleFields = filterClientWebsiteSettingsFields(nextRoleScopedFields);

        websiteSettingsFullDraft = normalizedFullSettings;
        websiteSettingsDraft = normalizeWebsiteSettingsValue(websiteSettingsFullDraft, nextVisibleFields);
    }

    function handleWebsiteSeoFileChange(type, event) {
        const file = event.currentTarget?.files?.[0] || null;

        if (type === "logo") {
            websiteIdentitySeoDraft = {
                ...websiteIdentitySeoDraft,
                logoFile: file,
            };
        } else if (type === "seoImage") {
            websiteIdentitySeoDraft = {
                ...websiteIdentitySeoDraft,
                seoImageFile: file,
            };
        }

        if (event.currentTarget) {
            event.currentTarget.value = "";
        }
    }

    async function reload() {
        if (!hasCmsCollections) {
            return;
        }

        await loadWebsites();
        await loadComponents();
        await loadPages();
        await loadBlocks();
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
                sort: resolveSort(websitesCollection, ["name", "title", "slug", "domain"]),
                requestKey: "nuvio_cms_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
                selectedPageId = "";
            } else if (!websites.some((record) => record.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
                selectedPageId = "";
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            selectedPageId = "";
            ApiClient.error(err);
        }

        isLoadingWebsites = false;
    }

    async function loadComponents() {
        if (!componentsCollection?.id) {
            components = [];
            return;
        }

        isLoadingComponents = true;

        try {
            components = await ApiClient.collection(componentsCollection.id).getFullList({
                sort: resolveSort(componentsCollection, ["name", "title", "key"]),
                requestKey: "nuvio_cms_components",
            });
        } catch (err) {
            components = [];
            ApiClient.error(err);
        }

        isLoadingComponents = false;
    }

    async function loadPages() {
        if (!pagesCollection?.id || !selectedWebsiteId || !pageWebsiteRelationField) {
            pages = [];
            selectedPageId = "";
            return;
        }

        isLoadingPages = true;

        try {
            pages = await ApiClient.collection(pagesCollection.id).getFullList({
                filter: `${pageWebsiteRelationField}="${selectedWebsiteId}"`,
                sort: resolveSort(pagesCollection, ["title", "name", "slug"]),
                requestKey: "nuvio_cms_pages_" + selectedWebsiteId,
            });

            if (!pages.length) {
                selectedPageId = "";
            } else if (!pages.some((record) => record.id === selectedPageId)) {
                selectedPageId = pages[0].id;
            }
        } catch (err) {
            pages = [];
            selectedPageId = "";
            ApiClient.error(err);
        }

        isLoadingPages = false;
    }

    async function loadBlocks() {
        if (!blocksCollection?.id || !selectedPageId || !blockPageRelationField) {
            blocks = [];
            openSectionId = "";
            return;
        }

        isLoadingBlocks = true;

        try {
            const query = {
                filter: `${blockPageRelationField}="${selectedPageId}"`,
                sort: resolveSort(blocksCollection, ["title", "slot", "variant"]),
                requestKey: "nuvio_cms_blocks_" + selectedPageId,
            };

            if (blockComponentRelationField) {
                query.expand = blockComponentRelationField;
            }

            blocks = await ApiClient.collection(blocksCollection.id).getFullList(query);
            if (!openSectionId || !blocks.some((block) => block.id === openSectionId)) {
                openSectionId = blocks[0]?.id || "";
            }
        } catch (err) {
            blocks = [];
            openSectionId = "";
            ApiClient.error(err);
        }

        isLoadingBlocks = false;
    }

    async function selectWebsite(websiteId) {
        if (`${websiteId || ""}` === `${selectedWebsiteId || ""}`) {
            return;
        }

        selectedWebsiteId = `${websiteId || ""}`;
        selectedPageId = "";
        openSectionId = "";

        await loadPages();
        await loadBlocks();
    }

    async function selectPage(pageId) {
        if (`${pageId || ""}` === `${selectedPageId || ""}`) {
            return;
        }

        selectedPageId = `${pageId || ""}`;
        openSectionId = "";

        await loadBlocks();
    }

    async function savePage() {
        pageError = "";

        if (!selectedPage?.id || !pagesCollection?.id || !selectedWebsiteId || !pageWebsiteRelationField) {
            pageError = "Select a page first.";
            return;
        }

        const payload = {};
        setPayloadField(payload, pageWebsiteRelationField, selectedWebsiteId);

        if (pageTitleField) {
            setPayloadField(payload, pageTitleField, normalizeString(pageEditForm.title));
        }
        if (pageEnabledField) {
            setPayloadField(payload, pageEnabledField, !!pageEditForm.enabled);
        }
        if (pageSeoTitleField) {
            setPayloadField(payload, pageSeoTitleField, normalizeString(pageEditForm.seoTitle));
        }
        if (pageSeoDescriptionField) {
            setPayloadField(payload, pageSeoDescriptionField, `${pageEditForm.seoDescription || ""}`);
        }

        isSavingPage = true;

        try {
            await ApiClient.collection(pagesCollection.id).update(selectedPage.id, payload);
            addSuccessToast("Page updated.");
            await loadPages();
            refreshPagePreview();
        } catch (err) {
            ApiClient.error(err);
            pageError = err?.response?.message || err?.message || "Failed to save page.";
        }

        isSavingPage = false;
    }

    function setActiveCmsTab(nextTab) {
        if (nextTab === cmsTabSettingsKey || nextTab === cmsTabPagesKey) {
            activeCmsTab = nextTab;
        }
    }

    async function saveWebsiteIdentitySeo() {
        websiteIdentitySeoError = "";

        if (!selectedWebsite?.id || !websitesCollection?.id || !hasWebsiteIdentitySeoFields) {
            websiteIdentitySeoError = "Could not save identity and global SEO.";
            return;
        }

        const payload = {};

        if (websiteSeoTitleField) {
            setPayloadField(payload, websiteSeoTitleField, normalizeString(websiteIdentitySeoDraft.seoTitle));
        }

        if (websiteSeoDescriptionField) {
            setPayloadField(payload, websiteSeoDescriptionField, `${websiteIdentitySeoDraft.seoDescription || ""}`);
        }

        if (websiteLogoField && websiteIdentitySeoDraft.logoFile) {
            setPayloadField(payload, websiteLogoField, websiteIdentitySeoDraft.logoFile);
        }

        if (websiteSeoImageField && websiteIdentitySeoDraft.seoImageFile) {
            setPayloadField(payload, websiteSeoImageField, websiteIdentitySeoDraft.seoImageFile);
        }

        if (!Object.keys(payload).length) {
            websiteIdentitySeoError = "There are no available fields to save.";
            return;
        }

        isSavingWebsiteIdentitySeo = true;

        try {
            await ApiClient.collection(websitesCollection.id).update(selectedWebsite.id, payload);
            addSuccessToast("Identity and global SEO updated.");
            await loadWebsites();
        } catch (err) {
            ApiClient.error(err);
            websiteIdentitySeoError = err?.response?.message || err?.message || "Failed to save identity and global SEO.";
        }

        isSavingWebsiteIdentitySeo = false;
    }

    async function saveWebsiteSettings() {
        websiteSettingsError = "";

        if (!selectedWebsite?.id || !websitesCollection?.id || !hasWebsiteSettingsField) {
            websiteSettingsError = "Could not save this website settings.";
            return;
        }

        isSavingWebsiteSettings = true;

        try {
            const payload = {};
            setPayloadField(payload, websiteSettingsField, structuredClone(websiteSettingsFullDraft));
            await ApiClient.collection(websitesCollection.id).update(selectedWebsite.id, payload);
            addSuccessToast("Website settings updated.");
            await loadWebsites();
        } catch (err) {
            ApiClient.error(err);
            websiteSettingsError = err?.response?.message || err?.message || "Failed to save website settings.";
        }

        isSavingWebsiteSettings = false;
    }

    async function saveSection(block) {
        const blockId = `${block?.id || ""}`;
        if (!blockId) {
            return;
        }

        sectionErrorById = { ...sectionErrorById, [blockId]: "" };

        if (!blocksCollection?.id || !blockPropsField) {
            sectionErrorById = {
                ...sectionErrorById,
                [blockId]: "Section cannot be saved because props field is missing.",
            };
            return;
        }

        const payload = {};
        setPayloadField(payload, blockPropsField, toPropsObject(sectionPropsDraftById?.[blockId]));

        isSavingSectionById = {
            ...isSavingSectionById,
            [blockId]: true,
        };

        try {
            await ApiClient.collection(blocksCollection.id).update(blockId, payload);
            addSuccessToast("Section updated.");
            await loadBlocks();
            refreshPagePreview();
        } catch (err) {
            ApiClient.error(err);
            sectionErrorById = {
                ...sectionErrorById,
                [blockId]: err?.response?.message || err?.message || "Failed to save section.",
            };
        }

        isSavingSectionById = {
            ...isSavingSectionById,
            [blockId]: false,
        };
    }

    async function handleWebsiteChange(event) {
        await selectWebsite(event.currentTarget.value);
    }
</script>

<PageWrapper>
    {#if $isCollectionsLoading && !hasCmsCollections}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading website content...</h1>
        </div>
    {:else if !hasCmsCollections}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>
                Missing required collections: {missingCollections.join(", ")}.<br />
                Keep using Collections until those collections are available.
            </div>
        </div>
    {:else}
        <section class="cms-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <div class="title-row">
                        <h2 class="m-0">Website Content</h2>
                        <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={reload} />
                    </div>
                    <p class="txt-sm txt-hint m-b-0 head-description">Edit your website pages and sections.</p>
                </div>

                <div class="head-selector">
                    <div class="selector-label-row">
                        <label class="txt-sm txt-hint selector-label m-b-0" for="cms-website">Website</label>
                    </div>
                    <div class="selector-controls">
                        <select
                            id="cms-website"
                            class="input input-sm"
                            value={selectedWebsiteId}
                            disabled={isLoadingWebsites || !websites.length}
                            on:change={handleWebsiteChange}
                        >
                            <option value="">Select website</option>
                            {#each websites as website}
                                <option value={website.id}>{getWebsiteLabel(website)}</option>
                            {/each}
                        </select>

                        {#if websitePublicUrl}
                            <a href={websitePublicUrl} target="_blank" rel="noreferrer noopener" class="btn btn-sm btn-outline">
                                View website
                            </a>
                        {/if}
                    </div>
                </div>
            </div>

            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-file-list-3-line" />
                    {pages.length} pages
                </span>
                <span class="summary-pill">
                    <i class="ri-layout-grid-line" />
                    {blocks.length} sections
                </span>
                {#if !componentsCollection?.id}
                    <span class="summary-pill warning">
                        <i class="ri-alert-line" />
                        components missing
                    </span>
                {/if}
            </div>

            <div class="tabs-header compact combined left operations-tabs cms-top-tabs">
                <button
                    type="button"
                    class="tab-item"
                    class:active={activeCmsTab === cmsTabPagesKey}
                    on:click={() => setActiveCmsTab(cmsTabPagesKey)}
                >
                    <i class="ri-file-list-3-line tab-icon" aria-hidden="true" />
                    <span class="tab-label">Pages</span>
                </button>
                <button
                    type="button"
                    class="tab-item"
                    class:active={activeCmsTab === cmsTabSettingsKey}
                    on:click={() => setActiveCmsTab(cmsTabSettingsKey)}
                >
                    <i class="ri-settings-3-line tab-icon" aria-hidden="true" />
                    <span class="tab-label">Website Settings</span>
                </button>
            </div>
        </section>

        <section class="panel cms-section-panel m-b-base">
            {#if activeCmsTab === cmsTabPagesKey}
                <div class="content-workspace-grid">
                    <aside class="pages-list-panel">
                        <div class="pages-list-head">
                            <h4 class="m-0">Pages</h4>
                            <span class="txt-sm txt-hint">{pages.length} total</span>
                        </div>

                        {#if !selectedWebsiteId}
                            <div class="txt-hint m-t-sm">Select a website first.</div>
                        {:else if !pageWebsiteRelationField}
                            <div class="txt-hint txt-danger m-t-sm">Pages relation to websites was not found.</div>
                        {:else if isLoadingPages}
                            <div class="txt-hint m-t-sm">Loading pages...</div>
                        {:else if !pages.length}
                            <div class="txt-hint m-t-sm">There are no pages for this website.</div>
                        {:else}
                            <div class="pages-list-body m-t-sm">
                                {#each pages as page}
                                    <button
                                        type="button"
                                        class="page-row"
                                        class:active={page.id === selectedPageId}
                                        on:click={() => selectPage(page.id)}
                                    >
                                        <span class="page-row-title">{getPageLabel(page)}</span>
                                        <span class="page-row-path">{getPagePath(page)}</span>
                                    </button>
                                {/each}
                            </div>
                        {/if}
                    </aside>

                    <div class="page-editor-panel">
                        {#if selectedPage}
                            <div class="page-editor-head">
                                <div>
                                    <h4 class="m-0">Page Details</h4>
                                    <p class="txt-sm txt-hint m-b-0 m-t-6">URL: {selectedPagePath}</p>
                                </div>
                            </div>

                            <div class="form-grid two-col m-t-sm">
                                {#if pageTitleField}
                                    <div class="form-field">
                                        <label class="txt-sm txt-hint block m-b-5" for="cms-page-title-content">Title</label>
                                        <input id="cms-page-title-content" class="input input-sm" bind:value={pageEditForm.title} />
                                    </div>
                                {/if}

                                <div class="form-field read-only-field">
                                    <label class="txt-sm txt-hint block m-b-5">URL</label>
                                    <div class="read-only-value">{selectedPagePath}</div>
                                </div>
                            </div>

                            {#if pageEnabledField}
                                <label class="checkbox-field m-t-8">
                                    <input type="checkbox" bind:checked={pageEditForm.enabled} />
                                    <span>Page active</span>
                                </label>
                            {/if}

                            {#if pageSeoTitleField || pageSeoDescriptionField}
                                <div class="seo-page-wrap m-t-base">
                                    <div class="sections-head">
                                        <h5 class="m-0">Page SEO</h5>
                                    </div>

                                    <div class="form-grid m-t-sm">
                                        {#if pageSeoTitleField}
                                            <div class="form-field">
                                                <label class="txt-sm txt-hint block m-b-5" for="cms-page-seo-title-content">
                                                    SEO Title
                                                </label>
                                                <input
                                                    id="cms-page-seo-title-content"
                                                    class="input input-sm"
                                                    bind:value={pageEditForm.seoTitle}
                                                />
                                            </div>
                                        {/if}

                                        {#if pageSeoDescriptionField}
                                            <div class="form-field">
                                                <label
                                                    class="txt-sm txt-hint block m-b-5"
                                                    for="cms-page-seo-description-content"
                                                >
                                                    SEO Description
                                                </label>
                                                <textarea
                                                    id="cms-page-seo-description-content"
                                                    class="input input-sm textarea-input"
                                                    rows="4"
                                                    bind:value={pageEditForm.seoDescription}
                                                />
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            {/if}

                            <div class="form-actions m-t-sm">
                                <button type="button" class="btn btn-sm btn-strong" disabled={isSavingPage} on:click={savePage}>
                                    {isSavingPage ? "Saving..." : "Save page"}
                                </button>
                            </div>

                            {#if pageError}
                                <p class="txt-danger m-t-8 m-b-0">{pageError}</p>
                            {/if}

                            <div class="page-preview-wrap m-t-base">
                                <div class="sections-head page-preview-head">
                                    <div>
                                        <h5 class="m-0">Page Preview</h5>
                                        <p class="txt-sm txt-hint m-b-0 m-t-6">Preview shows saved content only.</p>
                                    </div>
                                    <div class="page-preview-actions">
                                        <button
                                            type="button"
                                            class="btn btn-sm btn-outline"
                                            disabled={!pagePreviewUrl}
                                            on:click={refreshPagePreview}
                                        >
                                            Refresh preview
                                        </button>
                                        {#if pagePreviewUrl}
                                            <a href={pagePreviewUrl} target="_blank" rel="noreferrer noopener" class="btn btn-sm btn-outline">
                                                Open in new tab
                                            </a>
                                        {/if}
                                    </div>
                                </div>

                                {#if !selectedPageSlug}
                                    <div class="preview-empty-state m-t-sm">
                                        This page has no slug yet. Add a page slug to enable preview.
                                    </div>
                                {:else if !pagePreviewUrl}
                                    <div class="preview-empty-state m-t-sm">
                                        Preview base URL is not configured. Set <code>VITE_PUBLIC_SITE_BASE_URL</code> in the admin UI
                                        environment (for local dev: <code>VITE_PUBLIC_SITE_BASE_URL=http://localhost:5173</code>).
                                    </div>
                                {:else}
                                    <div class="page-preview-iframe-wrap m-t-sm">
                                        <iframe
                                            class="page-preview-iframe"
                                            src={pagePreviewIframeSrc}
                                            title={`Preview: ${selectedPagePath}`}
                                            loading="lazy"
                                        ></iframe>
                                    </div>
                                {/if}
                            </div>

                            <div class="sections-wrap m-t-base">
                                <div class="sections-head">
                                    <h5 class="m-0">Sections on this page</h5>
                                    <span class="txt-sm txt-hint">{blocks.length} total</span>
                                </div>

                                {#if !blockPageRelationField}
                                    <p class="txt-sm txt-danger m-t-8 m-b-0">Sections relation to pages was not found.</p>
                                {:else if !blockPropsField}
                                    <p class="txt-sm txt-danger m-t-8 m-b-0">Sections props field was not found.</p>
                                {:else if isLoadingBlocks || isLoadingComponents}
                                    <p class="txt-sm txt-hint m-t-8 m-b-0">Loading sections...</p>
                                {:else if !blocks.length}
                                    <p class="txt-sm txt-hint m-t-8 m-b-0">There are no sections linked to this page.</p>
                                {:else}
                                    <div class="sections-list m-t-sm">
                                        {#each blocks as block, index}
                                            <article class="section-card" class:open={openSectionId === block.id}>
                                                <button type="button" class="section-toggle" on:click={() => toggleSection(block.id)}>
                                                    <span class="section-toggle-content">
                                                        <strong>{getSectionTitle(block, index)}</strong>
                                                        <small>{getSectionDescription(block)}</small>
                                                    </span>
                                                    <i class={openSectionId === block.id ? "ri-arrow-up-s-line" : "ri-arrow-down-s-line"} />
                                                </button>

                                                {#if openSectionId === block.id}
                                                    <div class="section-body">
                                                        {#if getSectionSchemaFields(block).length}
                                                            <SchemaForm
                                                                fields={getSectionSchemaFields(block)}
                                                                value={sectionPropsDraftById[block.id] || {}}
                                                                showImport={false}
                                                                path={`sections.${block.id}`}
                                                                on:propsChange={(event) => updateSectionDraft(block.id, event.detail)}
                                                            />
                                                        {:else}
                                                            <p class="txt-sm txt-hint m-b-0">
                                                                This section has no editable fields in schema yet.
                                                            </p>
                                                        {/if}

                                                        <div class="form-actions m-t-sm">
                                                            <button
                                                                type="button"
                                                                class="btn btn-sm btn-strong"
                                                                disabled={!!isSavingSectionById[block.id] || !blockPropsField}
                                                                on:click={() => saveSection(block)}
                                                            >
                                                                {isSavingSectionById[block.id] ? "Saving..." : "Save section"}
                                                            </button>
                                                        </div>

                                                        {#if sectionErrorById[block.id]}
                                                            <p class="txt-danger m-t-8 m-b-0">{sectionErrorById[block.id]}</p>
                                                        {/if}
                                                    </div>
                                                {/if}
                                            </article>
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        {:else}
                            <div class="txt-hint">Select a page to edit.</div>
                        {/if}
                    </div>
                </div>
            {:else if activeCmsTab === cmsTabSettingsKey}
                <div class="settings-workspace">
                    <div class="settings-head">
                        <div>
                            <h4 class="m-0">Website Settings</h4>
                            <p class="txt-sm txt-hint m-b-0 m-t-6">Edit general settings for the selected website.</p>
                        </div>
                    </div>

                    {#if !selectedWebsiteId}
                        <p class="txt-hint m-b-0">Select a website to edit settings.</p>
                    {:else}
                        {#if hasWebsiteIdentitySeoFields}
                            <div class="settings-form-wrap m-t-sm">
                                <div class="settings-subhead">
                                    <h5 class="m-0">Identity & Global SEO</h5>
                                    <p class="txt-sm txt-hint m-b-0 m-t-6">Edit website logo and global metadata.</p>
                                </div>

                                <div class="form-grid two-col m-t-sm">
                                    {#if websiteSeoTitleField}
                                        <div class="form-field">
                                            <label class="txt-sm txt-hint block m-b-5" for="cms-website-seo-title">SEO Title global</label>
                                            <input
                                                id="cms-website-seo-title"
                                                class="input input-sm"
                                                bind:value={websiteIdentitySeoDraft.seoTitle}
                                            />
                                        </div>
                                    {/if}

                                    {#if websiteSeoDescriptionField}
                                        <div class="form-field">
                                            <label class="txt-sm txt-hint block m-b-5" for="cms-website-seo-description">
                                                SEO Description global
                                            </label>
                                            <textarea
                                                id="cms-website-seo-description"
                                                class="input input-sm textarea-input"
                                                rows="4"
                                                bind:value={websiteIdentitySeoDraft.seoDescription}
                                            />
                                        </div>
                                    {/if}

                                    {#if websiteLogoField}
                                        <div class="form-field">
                                            <label class="txt-sm txt-hint block m-b-5" for="cms-website-logo-file">Logo</label>
                                            <input
                                                id="cms-website-logo-file"
                                                class="input input-sm file-input"
                                                type="file"
                                                on:change={(event) => handleWebsiteSeoFileChange("logo", event)}
                                            />
                                            <div class="file-field-hint m-t-6">
                                                {#if websiteIdentitySeoDraft.logoFile}
                                                    <span>New: {websiteIdentitySeoDraft.logoFile.name}</span>
                                                {:else if websiteIdentitySeoDraft.logoCurrent}
                                                    <span>Current: {websiteIdentitySeoDraft.logoCurrent}</span>
                                                {:else}
                                                    <span>No current file.</span>
                                                {/if}
                                            </div>
                                        </div>
                                    {/if}

                                    {#if websiteSeoImageField}
                                        <div class="form-field">
                                            <label class="txt-sm txt-hint block m-b-5" for="cms-website-seo-image-file">
                                                Global SEO Image
                                            </label>
                                            <input
                                                id="cms-website-seo-image-file"
                                                class="input input-sm file-input"
                                                type="file"
                                                on:change={(event) => handleWebsiteSeoFileChange("seoImage", event)}
                                            />
                                            <div class="file-field-hint m-t-6">
                                                {#if websiteIdentitySeoDraft.seoImageFile}
                                                    <span>New: {websiteIdentitySeoDraft.seoImageFile.name}</span>
                                                {:else if websiteIdentitySeoDraft.seoImageCurrent}
                                                    <span>Current: {websiteIdentitySeoDraft.seoImageCurrent}</span>
                                                {:else}
                                                    <span>No current file.</span>
                                                {/if}
                                            </div>
                                        </div>
                                    {/if}
                                </div>

                                <div class="form-actions m-t-sm">
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-strong"
                                        disabled={isSavingWebsiteIdentitySeo}
                                        on:click={saveWebsiteIdentitySeo}
                                    >
                                        {isSavingWebsiteIdentitySeo ? "Saving..." : "Save identity & SEO"}
                                    </button>
                                </div>

                                {#if websiteIdentitySeoError}
                                    <p class="txt-danger m-t-8 m-b-0">{websiteIdentitySeoError}</p>
                                {/if}
                            </div>
                        {/if}

                        {#if !hasWebsiteSettingsField}
                            <p class="txt-danger m-b-0 m-t-sm">Settings field was not found in websites collection.</p>
                        {:else if !clientWebsiteSettingsFields.length}
                            <p class="txt-hint m-b-0 m-t-sm">No settings available for this profile.</p>
                        {:else}
                            <div class="settings-form-wrap m-t-sm">
                                <SchemaForm
                                    fields={clientWebsiteSettingsFields}
                                    value={websiteSettingsDraft}
                                    showImport={false}
                                    path={`websites.${selectedWebsiteId}.settings`}
                                    on:change={handleWebsiteSettingsChange}
                                />
                            </div>

                            <div class="form-actions m-t-sm">
                                <button
                                    type="button"
                                    class="btn btn-sm btn-strong"
                                    disabled={isSavingWebsiteSettings}
                                    on:click={saveWebsiteSettings}
                                >
                                    {isSavingWebsiteSettings ? "Saving..." : "Save settings"}
                                </button>
                            </div>

                            {#if websiteSettingsError}
                                <p class="txt-danger m-t-8 m-b-0">{websiteSettingsError}</p>
                            {/if}
                        {/if}
                    {/if}
                </div>
            {/if}
        </section>
    {/if}
</PageWrapper>

<style>
    .cms-head {
        display: flex;
        flex-direction: column;
        gap: 8px;
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

    .head-description {
        max-width: 520px;
    }

    .head-selector {
        width: min(100%, 620px);
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .selector-controls {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .selector-controls .input {
        flex: 1 1 auto;
        min-width: 260px;
    }

    .summary-badges {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 6px;
        justify-content: flex-end;
        margin-left: auto;
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
        padding: 5px 9px;
        white-space: nowrap;
    }

    .summary-pill i {
        color: var(--txtPrimaryColor);
        opacity: 0.85;
        font-size: 13px;
    }

    .summary-pill.warning {
        color: color-mix(in srgb, var(--dangerColor) 65%, var(--txtHintColor));
    }

    .cms-top-tabs {
        margin-top: 2px;
    }

    .cms-section-panel {
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .content-workspace-grid {
        display: grid;
        grid-template-columns: minmax(260px, 0.85fr) minmax(520px, 1.8fr);
        gap: 12px;
        align-items: stretch;
    }

    .pages-list-panel,
    .page-editor-panel {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        min-height: 460px;
        overflow: hidden;
        padding: 12px;
    }

    .pages-list-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .pages-list-body {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 620px;
        overflow: auto;
    }

    .page-row {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 10px;
        text-align: left;
        display: flex;
        flex-direction: column;
        gap: 4px;
        cursor: pointer;
    }

    .page-row:hover {
        background: color-mix(in srgb, var(--baseAlt1Color) 72%, var(--baseColor));
    }

    .page-row.active {
        border-color: color-mix(in srgb, var(--primaryColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 8%, var(--baseColor));
        box-shadow: inset 2px 0 0 color-mix(in srgb, var(--primaryColor) 60%, transparent);
    }

    .page-row-title {
        color: var(--txtPrimaryColor);
        font-weight: 600;
    }

    .page-row-path {
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
    }

    .page-editor-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .form-grid {
        display: grid;
        gap: 10px;
    }

    .form-grid.two-col {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .form-field {
        min-width: 0;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        padding: 8px 10px 10px;
    }

    .form-field .input {
        width: 100%;
        background: var(--baseColor);
    }

    .read-only-field .read-only-value {
        min-height: 32px;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        display: flex;
        align-items: center;
        padding: 0 10px;
        color: var(--txtHintColor);
    }

    .checkbox-field {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
    }

    .form-actions {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
    }

    .page-preview-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .page-preview-head {
        align-items: flex-start;
    }

    .page-preview-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .preview-empty-state {
        border: 1px dashed var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        color: var(--txtHintColor);
        font-size: var(--smFontSize);
        padding: 12px;
    }

    .preview-empty-state code {
        font-size: 12px;
    }

    .page-preview-iframe-wrap {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        overflow: hidden;
        background: var(--baseColor);
        min-height: 480px;
    }

    .page-preview-iframe {
        width: 100%;
        height: 70vh;
        min-height: 480px;
        border: 0;
        display: block;
        background: white;
    }

    .sections-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .sections-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .sections-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .section-card {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseAlt1Color);
        overflow: hidden;
    }

    .section-toggle {
        width: 100%;
        border: 0;
        background: transparent;
        text-align: left;
        padding: 10px 12px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        cursor: pointer;
    }

    .section-toggle:hover {
        background: color-mix(in srgb, var(--baseAlt1Color) 72%, var(--baseColor));
    }

    .section-toggle-content {
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
    }

    .section-toggle-content strong {
        color: var(--txtPrimaryColor);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .section-toggle-content small {
        color: var(--txtHintColor);
        font-size: 12px;
    }

    .section-body {
        border-top: 1px solid var(--baseAlt2Color);
        background: var(--baseColor);
        padding: 12px;
    }

    .seo-page-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .settings-workspace {
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 12px;
        min-height: 460px;
    }

    .settings-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
    }

    .settings-form-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 12px;
    }

    .settings-subhead {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .textarea-input {
        width: 100%;
        min-height: 108px;
        resize: vertical;
    }

    .file-input {
        width: 100%;
        padding: 6px 8px;
    }

    .file-field-hint {
        font-size: var(--smFontSize);
        color: var(--txtHintColor);
    }

    @media (max-width: 1200px) {
        .content-workspace-grid {
            grid-template-columns: 1fr;
        }

        .pages-list-panel,
        .page-editor-panel {
            min-height: auto;
        }
    }

    @media (max-width: 840px) {
        .cms-head,
        .cms-section-panel {
            padding: calc(var(--baseSpacing) - 12px) calc(var(--baseSpacing) - 10px);
        }

        .selector-controls {
            flex-direction: column;
            align-items: stretch;
        }

        .selector-controls .input {
            min-width: 0;
        }

        .summary-badges {
            justify-content: flex-start;
            margin-left: 0;
        }

        .cms-top-tabs {
            width: 100%;
        }

        .form-grid.two-col {
            grid-template-columns: 1fr;
        }

        .page-preview-head {
            flex-direction: column;
            align-items: stretch;
        }

        .page-preview-actions {
            width: 100%;
        }

        .page-preview-actions .btn {
            flex: 1 1 auto;
            justify-content: center;
        }

        .page-preview-iframe {
            height: 65vh;
            min-height: 420px;
        }
    }
</style>
