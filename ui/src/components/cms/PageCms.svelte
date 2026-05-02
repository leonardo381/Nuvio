<script>
    import { onMount } from "svelte";
    import { querystring } from "svelte-spa-router";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
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
    const pageEditorTabContentKey = "content";
    const pageEditorTabSeoKey = "seo";
    const pageStatusFilterAllKey = "all";
    const pageStatusFilterActiveKey = "active";
    const pageStatusFilterInactiveKey = "inactive";
    const clientSettingsRole = "client";
    const visibleClientSettingsKeys = new Set(["whatsapp", "contactForm", "newsletter", "i18n"]);

    let websites = [];
    let pages = [];
    let blocks = [];
    let components = [];

    let selectedWebsiteId = initialQueryParams.get("cmsWebsite") || "";
    let selectedPageId = initialQueryParams.get("cmsPage") || "";
    let activeCmsTab = initialQueryParams.get("cmsTab") === cmsTabSettingsKey ? cmsTabSettingsKey : cmsTabPagesKey;
    let activePageEditorTab = pageEditorTabContentKey;
    let pageStatusFilter = pageStatusFilterAllKey;
    let pageSearch = "";
    let focusedBlockId = "";
    let editingSectionId = "";
    let sectionEditorPanel;

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
    $: blockVariantField = resolveFieldName(blocksCollection, ["variant"]);
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
    $: selectedWebsiteSlug = normalizeString(websiteSlugField ? selectedWebsite?.[websiteSlugField] : "");
    $: selectedPageSlug = normalizeString(pageSlugField ? selectedPage?.[pageSlugField] : "");
    $: pagePreviewUrl = buildPagePreviewUrl(selectedWebsiteSlug, selectedPageSlug);
    $: pagePreviewFocusedUrl = buildPagePreviewFocusedUrl(pagePreviewUrl, focusedBlockId);
    $: pagePreviewIframeSrc = buildPreviewIframeSrc(pagePreviewFocusedUrl, pagePreviewReloadToken);
    $: normalizedPageSearch = normalizeString(pageSearch).toLowerCase();
    $: activePagesCount = pageEnabledField ? pages.filter((record) => isPageActive(record)).length : 0;
    $: inactivePagesCount = pageEnabledField ? Math.max(0, pages.length - activePagesCount) : 0;
    $: filteredPages = pages.filter((record) => {
        if (pageEnabledField) {
            if (pageStatusFilter === pageStatusFilterActiveKey && !isPageActive(record)) {
                return false;
            }
            if (pageStatusFilter === pageStatusFilterInactiveKey && isPageActive(record)) {
                return false;
            }
        }

        if (!normalizedPageSearch) {
            return true;
        }

        const pageLabel = normalizeString(getPageLabel(record)).toLowerCase();
        const pageSlug = normalizeString(pageSlugField ? record?.[pageSlugField] : "").toLowerCase();
        return pageLabel.includes(normalizedPageSearch) || pageSlug.includes(normalizedPageSearch);
    });
    $: if (!pageEnabledField && pageStatusFilter !== pageStatusFilterAllKey) {
        pageStatusFilter = pageStatusFilterAllKey;
    }

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
        focusedBlockId = "";
        editingSectionId = "";
        activePageEditorTab = pageEditorTabContentKey;
        pageEditForm = {
            seoTitle: pageSeoTitleField ? `${selectedPage?.[pageSeoTitleField] || ""}` : "",
            seoDescription: pageSeoDescriptionField ? `${selectedPage?.[pageSeoDescriptionField] || ""}` : "",
        };
        pageError = "";
    }

    $: if (!selectedPage?.id) {
        lastPageSeedId = "";
        focusedBlockId = "";
        editingSectionId = "";
        activePageEditorTab = pageEditorTabContentKey;
    }

    $: if (activePageEditorTab !== pageEditorTabContentKey && activePageEditorTab !== pageEditorTabSeoKey) {
        activePageEditorTab = pageEditorTabContentKey;
    }

    $: if (!hasCmsCollections) {
        websites = [];
        pages = [];
        blocks = [];
        components = [];
        selectedWebsiteId = "";
        selectedPageId = "";
        lastCollectionsKey = "";
        editingSectionId = "";
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

            if (editingSectionId && !blocks.some((block) => block.id === editingSectionId)) {
                editingSectionId = "";
            }
        }
    }

    $: if (focusedBlockId && !blocks.some((block) => `${block?.id || ""}` === `${focusedBlockId}`)) {
        focusedBlockId = "";
    }

    $: selectedEditingSection = blocks.find((block) => `${block?.id || ""}` === `${editingSectionId || ""}`) || null;
    $: selectedEditingSectionFields = selectedEditingSection ? getSectionSchemaFields(selectedEditingSection) : [];
    $: selectedEditingSectionIndex = selectedEditingSection
        ? blocks.findIndex((block) => `${block?.id || ""}` === `${selectedEditingSection?.id || ""}`)
        : -1;
    $: selectedEditingSectionSubtitle = selectedEditingSection
        ? getSectionDescription(selectedEditingSection, Math.max(selectedEditingSectionIndex, 0))
        : "";
    $: selectedEditingSectionSummaryPills = selectedEditingSection
        ? getSectionSummaryPills(selectedEditingSection, selectedEditingSectionFields)
        : [];

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
        const title = normalizeString(pageTitleField ? record?.[pageTitleField] : "");
        if (title) {
            return title;
        }

        const slug = normalizeString(pageSlugField ? record?.[pageSlugField] : "");
        if (slug) {
            return slug;
        }

        return record?.id || "-";
    }

    function formatPageListDate(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return "";
        }

        const parsed = new Date(raw);
        if (Number.isNaN(parsed.getTime())) {
            return "";
        }

        return parsed.toLocaleString([], {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function getPageListMeta(record) {
        const updated = formatPageListDate(record?.updated || record?.created);
        if (updated) {
            return `Updated ${updated}`;
        }

        return "Click to edit content, SEO, and preview.";
    }

    function isPageActive(record) {
        if (!pageEnabledField) {
            return false;
        }
        return !!record?.[pageEnabledField];
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

    function buildPagePreviewFocusedUrl(pagePreviewUrl, blockId) {
        const normalizedPreviewUrl = normalizeString(pagePreviewUrl);
        if (!normalizedPreviewUrl) {
            return "";
        }

        try {
            const parsed = new URL(normalizedPreviewUrl);
            parsed.searchParams.set("cmsPreview", "1");
            const normalizedBlockId = normalizeString(blockId);
            if (normalizedBlockId) {
                parsed.searchParams.set("focusBlock", normalizedBlockId);
            } else {
                parsed.searchParams.delete("focusBlock");
            }
            return parsed.toString();
        } catch (_) {
            return "";
        }
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

    function areSameLabel(left, right) {
        return normalizeString(left).toLowerCase() === normalizeString(right).toLowerCase();
    }

    function toHumanLabel(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return "";
        }

        const readable = raw
            .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
            .replace(/[_-]+/g, " ")
            .replace(/\s+/g, " ")
            .trim();

        if (!readable) {
            return "";
        }

        return readable
            .split(" ")
            .map((token) => token.charAt(0).toUpperCase() + token.slice(1))
            .join(" ");
    }

    function formatCount(count, singular, plural = `${singular}s`) {
        return `${count} ${count === 1 ? singular : plural}`;
    }

    function getSectionVariantLabel(block) {
        if (!blockVariantField) {
            return "";
        }

        const rawVariant = normalizeString(block?.[blockVariantField]);
        if (!rawVariant) {
            return "";
        }

        return toHumanLabel(rawVariant);
    }

    function getSectionDescription(block, index) {
        const title = getSectionTitle(block, index);
        const component = getComponentForBlock(block);
        const componentName = normalizeString(componentNameField ? component?.[componentNameField] : "");
        if (componentName && !areSameLabel(componentName, title)) {
            return componentName;
        }

        const variantLabel = getSectionVariantLabel(block);
        if (variantLabel && !areSameLabel(variantLabel, title)) {
            return `Variant: ${variantLabel}`;
        }

        return "Edit section content.";
    }

    function getSectionDraftProps(block) {
        const blockId = normalizeString(block?.id);
        const draftValue = blockId ? sectionPropsDraftById?.[blockId] : null;

        if (isPlainObject(draftValue)) {
            return draftValue;
        }

        return toPropsObject(blockPropsField ? block?.[blockPropsField] : {});
    }

    function getSectionSummary(block, schemaFields = getSectionSchemaFields(block)) {
        const summary = {
            editableFields: schemaFields.length,
            itemsCount: 0,
            hasMedia: false,
            variantLabel: getSectionVariantLabel(block),
        };

        if (!schemaFields.length) {
            return summary;
        }

        const propsValue = getSectionDraftProps(block);

        for (const field of schemaFields) {
            const fieldType = normalizeString(field?.type).toLowerCase();
            const fieldKey = normalizeString(field?.key);

            if (fieldType === "file") {
                summary.hasMedia = true;
            }

            if (fieldType === "array" && fieldKey && Array.isArray(propsValue?.[fieldKey])) {
                summary.itemsCount += propsValue[fieldKey].length;
            }
        }

        return summary;
    }

    function getSectionSummaryPills(block, schemaFields = getSectionSchemaFields(block)) {
        const sectionSummary = getSectionSummary(block, schemaFields);
        const pills = [formatCount(sectionSummary.editableFields, "field")];

        if (sectionSummary.itemsCount > 0) {
            pills.push(formatCount(sectionSummary.itemsCount, "item"));
        }

        if (sectionSummary.hasMedia) {
            pills.push("Media");
        }

        if (sectionSummary.variantLabel) {
            pills.push(`Variant: ${sectionSummary.variantLabel}`);
        }

        return pills;
    }

    function openSectionEditor(blockId, options = {}) {
        const { forceRefresh = false } = options;
        const normalizedBlockId = normalizeString(blockId);
        if (!normalizedBlockId) {
            return;
        }

        if (!blocks.some((block) => `${block?.id || ""}` === normalizedBlockId)) {
            return;
        }

        activePageEditorTab = pageEditorTabContentKey;

        if (focusedBlockId !== normalizedBlockId) {
            focusedBlockId = normalizedBlockId;
            refreshPagePreview();
        } else if (forceRefresh) {
            refreshPagePreview();
        }

        editingSectionId = normalizedBlockId;
    }

    function closeSectionEditor() {
        editingSectionId = "";
    }

    function clearFocusedPreview() {
        if (!focusedBlockId) {
            return;
        }

        focusedBlockId = "";
        refreshPagePreview();
    }

    function getOriginFromUrl(value) {
        try {
            return normalizeString(new URL(value).origin);
        } catch (_) {
            return "";
        }
    }

    function getAllowedPreviewOrigins() {
        const allowedOrigins = new Set();
        const configuredBase = normalizeBaseUrl(getConfiguredPublicBaseUrl(), { allowSingleLabelHost: false });
        const websiteBase = normalizeBaseUrl(getWebsitePublicUrl(selectedWebsite), { allowSingleLabelHost: false });

        for (const candidate of [configuredBase, pagePreviewUrl, websiteBase]) {
            const origin = getOriginFromUrl(candidate);
            if (origin) {
                allowedOrigins.add(origin);
            }
        }

        return allowedOrigins;
    }

    function isPreviewMessageOriginAllowed(origin) {
        const normalizedOrigin = normalizeString(origin);
        if (!normalizedOrigin) {
            return false;
        }

        const allowedOrigins = getAllowedPreviewOrigins();
        if (allowedOrigins.has(normalizedOrigin)) {
            return true;
        }

        if (import.meta.env.DEV) {
            return /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/i.test(normalizedOrigin);
        }

        return false;
    }

    function parsePreviewMessageData(data) {
        if (data && typeof data === "object") {
            return data;
        }

        if (typeof data === "string") {
            try {
                const parsed = JSON.parse(data);
                return parsed && typeof parsed === "object" ? parsed : null;
            } catch (_) {
                return null;
            }
        }

        return null;
    }

    function handlePreviewIframeEditMessage(event) {
        if (activeCmsTab !== cmsTabPagesKey || !selectedPageId) {
            return;
        }

        const message = parsePreviewMessageData(event?.data);
        if (!message || message.source !== "nuvio-preview" || message.type !== "edit-block") {
            return;
        }

        if (!isPreviewMessageOriginAllowed(event?.origin || "")) {
            return;
        }

        const nextBlockId = normalizeString(message?.blockId);
        if (!nextBlockId) {
            return;
        }

        if (!blocks.some((block) => `${block?.id || ""}` === nextBlockId)) {
            return;
        }

        openSectionEditor(nextBlockId, { forceRefresh: true });
    }

    onMount(() => {
        if (typeof window === "undefined") {
            return undefined;
        }

        const handleWindowMessage = (event) => {
            handlePreviewIframeEditMessage(event);
        };

        window.addEventListener("message", handleWindowMessage);

        return () => {
            window.removeEventListener("message", handleWindowMessage);
        };
    });

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
            editingSectionId = "";
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
            if (editingSectionId && !blocks.some((block) => block.id === editingSectionId)) {
                editingSectionId = "";
            }
        } catch (err) {
            blocks = [];
            editingSectionId = "";
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
        pageSearch = "";
        pageStatusFilter = pageStatusFilterAllKey;
        focusedBlockId = "";
        editingSectionId = "";
        activePageEditorTab = pageEditorTabContentKey;

        await loadPages();
        await loadBlocks();
    }

    async function selectPage(pageId) {
        if (`${pageId || ""}` === `${selectedPageId || ""}`) {
            return;
        }

        selectedPageId = `${pageId || ""}`;
        focusedBlockId = "";
        editingSectionId = "";
        activePageEditorTab = pageEditorTabContentKey;

        await loadBlocks();
    }

    function setActivePageEditorTab(nextTab) {
        if (nextTab === pageEditorTabContentKey || nextTab === pageEditorTabSeoKey) {
            activePageEditorTab = nextTab;
        }
    }

    async function savePageSeo() {
        pageError = "";

        if (!selectedPage?.id || !pagesCollection?.id) {
            pageError = "Select a page first.";
            return;
        }

        const payload = {};
        if (pageSeoTitleField) {
            setPayloadField(payload, pageSeoTitleField, normalizeString(pageEditForm.seoTitle));
        }
        if (pageSeoDescriptionField) {
            setPayloadField(payload, pageSeoDescriptionField, `${pageEditForm.seoDescription || ""}`);
        }

        if (!Object.keys(payload).length) {
            pageError = "SEO fields are not available for this page collection.";
            return;
        }

        isSavingPage = true;

        try {
            await ApiClient.collection(pagesCollection.id).update(selectedPage.id, payload);
            addSuccessToast("Page SEO updated.");
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
        <section class="cms-head operations-head panel m-b-base">
            <div class="head-main">
                <div class="summary-title-wrap">
                    <div class="title-row">
                        <h2 class="m-0">Website Content</h2>
                        <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={reload} />
                    </div>
                    <p class="txt-sm txt-hint m-b-0 head-description">Edit your website pages and sections.</p>
                </div>

                <div class="head-selector">
                    <div class="selector-row">
                        <label class="txt-sm txt-hint selector-label m-b-0" for="cms-website">Website</label>
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

            <div class="head-tools">
                <div class="tabs-header compact combined left operations-tabs">
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
            </div>
        </section>

        <section class="panel cms-section-panel m-b-base" class:is-pages-workspace={activeCmsTab === cmsTabPagesKey}>
            {#if activeCmsTab === cmsTabPagesKey}
                <div class="content-workspace-grid">
                    <aside class="panel pages-list-panel">
                        <div class="pages-list-head">
                            <div class="pages-list-title-wrap">
                                <h4 class="m-0">Pages</h4>
                                <p class="txt-sm txt-hint m-b-0 m-t-6">Find and open a page to edit content and SEO.</p>
                            </div>
                            <span class="txt-sm txt-hint pages-list-totals">
                                {#if pageEnabledField}
                                    {pages.length} total | {activePagesCount} active | {inactivePagesCount} inactive
                                {:else}
                                    {pages.length} total
                                {/if}
                            </span>
                        </div>

                        <div class="page-filter-chips m-t-sm" role="toolbar" aria-label="Filter pages by status">
                            <button
                                type="button"
                                class="btn btn-xs btn-outline page-filter-chip"
                                class:is-active={pageStatusFilter === pageStatusFilterAllKey}
                                on:click={() => (pageStatusFilter = pageStatusFilterAllKey)}
                            >
                                All ({pages.length})
                            </button>
                            {#if pageEnabledField}
                                <button
                                    type="button"
                                    class="btn btn-xs btn-outline page-filter-chip"
                                    class:is-active={pageStatusFilter === pageStatusFilterActiveKey}
                                    on:click={() => (pageStatusFilter = pageStatusFilterActiveKey)}
                                >
                                    Active ({activePagesCount})
                                </button>
                                <button
                                    type="button"
                                    class="btn btn-xs btn-outline page-filter-chip"
                                    class:is-active={pageStatusFilter === pageStatusFilterInactiveKey}
                                    on:click={() => (pageStatusFilter = pageStatusFilterInactiveKey)}
                                >
                                    Inactive ({inactivePagesCount})
                                </button>
                            {/if}
                        </div>

                        <div class="m-t-sm">
                            <label class="txt-sm txt-hint block m-b-5" for="cms-pages-search">Search</label>
                            <input
                                id="cms-pages-search"
                                class="input input-sm"
                                type="search"
                                placeholder="Search by page title..."
                                bind:value={pageSearch}
                            />
                        </div>

                        {#if !selectedWebsiteId}
                            <div class="txt-hint m-t-sm">Select a website first.</div>
                        {:else if !pageWebsiteRelationField}
                            <div class="txt-hint txt-danger m-t-sm">Pages relation to websites was not found.</div>
                        {:else if isLoadingPages}
                            <div class="txt-hint m-t-sm">Loading pages...</div>
                        {:else if !pages.length}
                            <div class="txt-hint m-t-sm">There are no pages for this website.</div>
                        {:else if !filteredPages.length}
                            <div class="txt-hint m-t-sm">No pages match the current filters.</div>
                        {:else}
                            <div class="pages-list-body m-t-sm">
                                {#each filteredPages as page}
                                    <button
                                        type="button"
                                        class="page-row"
                                        class:active={page.id === selectedPageId}
                                        on:click={() => selectPage(page.id)}
                                    >
                                        <div class="page-row-main">
                                            <span class="page-row-title">{getPageLabel(page)}</span>
                                            {#if pageEnabledField}
                                                <span class="label label-sm page-list-status" class:is-active={isPageActive(page)}>
                                                    {isPageActive(page) ? "Active" : "Inactive"}
                                                </span>
                                            {/if}
                                        </div>
                                        <div class="page-row-meta">
                                            <span class="txt-xs txt-hint">{getPageListMeta(page)}</span>
                                        </div>
                                    </button>
                                {/each}
                            </div>
                        {/if}
                    </aside>

                    <div class="panel page-editor-panel">
                        {#if selectedPage}
                            <div class="page-editor-head">
                                <div class="page-context-main">
                                    <h4 class="m-0">{getPageLabel(selectedPage)}</h4>
                                </div>
                                <div class="page-context-meta">
                                    {#if pageEnabledField}
                                        <span class="label label-sm page-status-pill" class:is-active={isPageActive(selectedPage)}>
                                            {isPageActive(selectedPage) ? "Active" : "Inactive"}
                                        </span>
                                    {/if}
                                    <span class="label label-sm page-status-pill page-count-pill">{blocks.length} sections</span>
                                </div>
                            </div>

                            <div class="page-editor-tabs-row">
                                <div class="tabs-header compact combined left operations-tabs page-editor-tabs">
                                    <button
                                        type="button"
                                        class="tab-item"
                                        class:active={activePageEditorTab === pageEditorTabContentKey}
                                        on:click={() => setActivePageEditorTab(pageEditorTabContentKey)}
                                    >
                                        <i class="ri-layout-grid-line tab-icon" aria-hidden="true" />
                                        <span class="tab-label">Content</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="tab-item"
                                        class:active={activePageEditorTab === pageEditorTabSeoKey}
                                        on:click={() => setActivePageEditorTab(pageEditorTabSeoKey)}
                                    >
                                        <i class="ri-search-eye-line tab-icon" aria-hidden="true" />
                                        <span class="tab-label">SEO</span>
                                    </button>
                                </div>
                            </div>

                            {#if activePageEditorTab === pageEditorTabContentKey}
                                <div class="content-preview-first-wrap m-t-sm">
                                    <div class="sections-head page-preview-head">
                                        <div class="page-preview-head-left">
                                            <h5 class="m-0">Page preview</h5>
                                            <span class="txt-sm txt-hint page-preview-helper">
                                                Click Edit section in the preview to edit content. Preview shows saved content only.
                                            </span>
                                            {#if focusedBlockId}
                                                <div class="preview-focus-hint">
                                                    <span>Highlighting selected section.</span>
                                                    <button type="button" class="btn-link" on:click={clearFocusedPreview}>
                                                        Clear highlight
                                                    </button>
                                                </div>
                                            {/if}
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
                                                <a
                                                    href={pagePreviewFocusedUrl || pagePreviewUrl}
                                                    target="_blank"
                                                    rel="noreferrer noopener"
                                                    class="btn btn-sm btn-outline"
                                                >
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
                                            Preview unavailable. Configure the public preview URL to edit visually.
                                        </div>

                                        {#if !blockPageRelationField}
                                            <p class="txt-sm txt-danger m-t-sm m-b-0">Sections relation to pages was not found.</p>
                                        {:else if !blockPropsField}
                                            <p class="txt-sm txt-danger m-t-sm m-b-0">Sections props field was not found.</p>
                                        {:else if isLoadingBlocks || isLoadingComponents}
                                            <p class="txt-sm txt-hint m-t-sm m-b-0">Loading sections...</p>
                                        {:else if !blocks.length}
                                            <p class="txt-sm txt-hint m-t-sm m-b-0">There are no sections linked to this page.</p>
                                        {:else}
                                            <div class="fallback-sections-list m-t-sm">
                                                {#each blocks as block, index}
                                                    {@const sectionFields = getSectionSchemaFields(block)}
                                                    {@const sectionSubtitle = getSectionDescription(block, index)}
                                                    <button
                                                        type="button"
                                                        class="fallback-section-item"
                                                        class:selected={editingSectionId === block.id}
                                                        on:click={() => openSectionEditor(block.id)}
                                                    >
                                                        <span class="fallback-section-title">{getSectionTitle(block, index)}</span>
                                                        <span class="fallback-section-subtitle">{sectionSubtitle}</span>
                                                        <span class="fallback-section-meta">
                                                            {index + 1} of {blocks.length} · {sectionFields.length} fields
                                                        </span>
                                                    </button>
                                                {/each}
                                            </div>
                                        {/if}
                                    {:else}
                                        <div class="content-preview-workspace m-t-sm">
                                            <div class="page-preview-iframe-wrap content-preview-iframe-wrap">
                                                <iframe
                                                    class="page-preview-iframe content-preview-iframe"
                                                    src={pagePreviewIframeSrc}
                                                    title={`Preview: ${getPageLabel(selectedPage)}`}
                                                    loading="lazy"
                                                ></iframe>
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                            {:else if activePageEditorTab === pageEditorTabSeoKey}
                                <div class="seo-page-wrap m-t-sm">
                                    <div class="sections-head">
                                        <div>
                                            <h5 class="m-0">Page SEO</h5>
                                            <p class="txt-sm txt-hint m-b-0 m-t-6">Control how this page appears in search results.</p>
                                        </div>
                                    </div>

                                    {#if pageSeoTitleField || pageSeoDescriptionField}
                                        <div class="form-grid m-t-sm">
                                            {#if pageSeoTitleField}
                                                <div class="form-field">
                                                    <label for="cms-page-seo-title-content">
                                                        SEO Title
                                                    </label>
                                                    <input
                                                        id="cms-page-seo-title-content"
                                                        class="input"
                                                        bind:value={pageEditForm.seoTitle}
                                                    />
                                                </div>
                                            {/if}

                                            {#if pageSeoDescriptionField}
                                                <div class="form-field">
                                                    <label for="cms-page-seo-description-content">
                                                        SEO Description
                                                    </label>
                                                    <textarea
                                                        id="cms-page-seo-description-content"
                                                        class="input textarea-input"
                                                        rows="4"
                                                        bind:value={pageEditForm.seoDescription}
                                                    />
                                                </div>
                                            {/if}
                                        </div>

                                        <div class="form-actions m-t-sm">
                                            <button
                                                type="button"
                                                class="btn btn-sm"
                                                disabled={isSavingPage}
                                                on:click={savePageSeo}
                                            >
                                                {isSavingPage ? "Saving..." : "Save SEO"}
                                            </button>
                                        </div>

                                        {#if pageError}
                                            <p class="txt-danger m-t-8 m-b-0">{pageError}</p>
                                        {/if}
                                    {:else}
                                        <p class="txt-sm txt-hint m-t-8 m-b-0">SEO fields are not available for this page collection.</p>
                                    {/if}
                                </div>
                            {/if}
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
                                            <label for="cms-website-seo-title">SEO Title global</label>
                                            <input
                                                id="cms-website-seo-title"
                                                class="input"
                                                bind:value={websiteIdentitySeoDraft.seoTitle}
                                            />
                                        </div>
                                    {/if}

                                    {#if websiteSeoDescriptionField}
                                        <div class="form-field">
                                            <label for="cms-website-seo-description">
                                                SEO Description global
                                            </label>
                                            <textarea
                                                id="cms-website-seo-description"
                                                class="input textarea-input"
                                                rows="4"
                                                bind:value={websiteIdentitySeoDraft.seoDescription}
                                            />
                                        </div>
                                    {/if}

                                    {#if websiteLogoField}
                                        <div class="form-field">
                                            <label for="cms-website-logo-file">Logo</label>
                                            <input
                                                id="cms-website-logo-file"
                                                class="input file-input"
                                                type="file"
                                                on:change={(event) => handleWebsiteSeoFileChange("logo", event)}
                                            />
                                            <div class="help-block file-field-hint m-t-6">
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
                                            <label for="cms-website-seo-image-file">
                                                Global SEO Image
                                            </label>
                                            <input
                                                id="cms-website-seo-image-file"
                                                class="input file-input"
                                                type="file"
                                                on:change={(event) => handleWebsiteSeoFileChange("seoImage", event)}
                                            />
                                            <div class="help-block file-field-hint m-t-6">
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
                                        class="btn btn-sm"
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
                                    class="btn btn-sm"
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

        {#if activeCmsTab === cmsTabPagesKey && selectedEditingSection}
            <OverlayPanel
                bind:this={sectionEditorPanel}
                class="overlay-panel-lg cms-section-editor-panel"
                active={true}
                btnClose={false}
                escClose={false}
                overlayClose={true}
                on:hide={closeSectionEditor}
            >
                <svelte:fragment slot="header">
                    <div class="section-drawer-head">
                        <div class="section-drawer-head-top">
                            <span class="txt-xs txt-hint txt-uppercase txt-bold">Edit section</span>
                        </div>
                        <strong class="section-drawer-name">{getSectionTitle(selectedEditingSection, Math.max(selectedEditingSectionIndex, 0))}</strong>
                        <div class="section-drawer-support-row">
                            <span class="txt-sm txt-hint section-drawer-helper">{selectedEditingSectionSubtitle || "Edit section content."}</span>
                            <div class="section-drawer-meta">
                                <span class="label label-sm section-summary-pill">{Math.max(selectedEditingSectionIndex, 0) + 1} of {blocks.length}</span>
                                {#each selectedEditingSectionSummaryPills as summaryPill}
                                    <span class="label label-sm section-summary-pill">{summaryPill}</span>
                                {/each}
                                {#if selectedPage}
                                    <span class="label label-sm section-summary-pill">Page: {getPageLabel(selectedPage)}</span>
                                {/if}
                            </div>
                        </div>
                    </div>
                    <button
                        type="button"
                        class="btn btn-sm btn-outline section-drawer-close-btn"
                        on:click={() => sectionEditorPanel?.hide()}
                    >
                        Close
                    </button>
                </svelte:fragment>

                <div class="section-drawer-body">
                    {#if selectedEditingSectionFields.length}
                        <SchemaForm
                            fields={selectedEditingSectionFields}
                            value={sectionPropsDraftById[selectedEditingSection.id] || {}}
                            showImport={false}
                            path={`sections.${selectedEditingSection.id}`}
                            on:propsChange={(event) => updateSectionDraft(selectedEditingSection.id, event.detail)}
                        />
                    {:else}
                        <p class="txt-sm txt-hint m-b-0">This section has no editable fields.</p>
                    {/if}

                    {#if sectionErrorById[selectedEditingSection.id]}
                        <p class="txt-danger m-t-8 m-b-0">{sectionErrorById[selectedEditingSection.id]}</p>
                    {/if}
                </div>

                <svelte:fragment slot="footer">
                    <button type="button" class="btn btn-sm btn-outline" on:click={() => sectionEditorPanel?.hide()}>Cancel</button>
                    <button
                        type="button"
                        class="btn btn-sm"
                        disabled={!!isSavingSectionById[selectedEditingSection.id] || !blockPropsField}
                        on:click={() => saveSection(selectedEditingSection)}
                    >
                        {isSavingSectionById[selectedEditingSection.id] ? "Saving..." : "Save changes"}
                    </button>
                </svelte:fragment>
            </OverlayPanel>
        {/if}
    {/if}
</PageWrapper>

<style>
    .page-editor-tabs-row {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        width: auto;
        margin-top: 4px;
    }

    .cms-section-panel {
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .cms-section-panel.is-pages-workspace {
        border: 0;
        background: transparent;
        box-shadow: none;
        padding: 0;
    }

    .content-workspace-grid {
        display: grid;
        grid-template-columns: minmax(280px, 340px) minmax(0, 1fr);
        gap: 14px;
        align-items: start;
    }

    .pages-list-panel,
    .page-editor-panel {
        min-height: 460px;
        overflow: hidden;
        padding: calc(var(--baseSpacing) - 10px) calc(var(--baseSpacing) - 8px);
    }

    .pages-list-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .pages-list-title-wrap {
        min-width: 0;
    }

    .pages-list-totals {
        text-align: right;
        white-space: nowrap;
    }

    .page-filter-chips {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
    }

    .page-filter-chip {
        min-height: 28px;
    }

    .page-filter-chip.is-active {
        background: var(--baseAlt2Color);
        border-color: var(--baseAlt2Color);
    }

    .pages-list-body {
        display: flex;
        flex-direction: column;
        gap: 0;
        max-height: 620px;
        overflow: auto;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
    }

    .page-row {
        border: 0;
        border-bottom: 1px solid var(--baseAlt2Color);
        border-radius: 0;
        background: transparent;
        padding: 10px 12px;
        text-align: left;
        display: flex;
        flex-direction: column;
        gap: 6px;
        cursor: pointer;
    }

    .page-row:last-child {
        border-bottom: 0;
    }

    .page-row:nth-child(odd) {
        background: var(--baseColor);
    }

    .page-row:nth-child(even) {
        background: var(--baseAlt1Color);
    }

    .page-row:hover {
        background: color-mix(in srgb, var(--primaryColor) 4%, var(--baseColor));
    }

    .page-row.active {
        background: color-mix(in srgb, var(--primaryColor) 8%, var(--baseColor));
        box-shadow: inset 3px 0 0 color-mix(in srgb, var(--primaryColor) 60%, transparent);
    }

    .page-row-main {
        width: 100%;
        min-width: 0;
        display: inline-flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .page-row-title {
        color: var(--txtPrimaryColor);
        font-weight: 600;
        min-width: 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .page-row-meta {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 6px;
        min-width: 0;
    }

    .page-list-status,
    .page-status-pill {
        font-weight: 600;
    }

    .page-list-status.is-active,
    .page-status-pill.is-active {
        color: color-mix(in srgb, var(--successColor) 80%, var(--txtHintColor));
        border-color: color-mix(in srgb, var(--successColor) 42%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--successColor) 10%, var(--baseColor));
    }

    .page-count-pill {
        color: var(--txtHintColor);
    }

    .page-editor-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
        padding-bottom: 2px;
    }

    .page-context-main {
        min-width: 0;
        display: inline-flex;
        align-items: baseline;
        gap: 8px;
        flex-wrap: wrap;
    }

    .page-context-meta {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .page-editor-tabs {
        display: inline-flex;
        align-items: center;
        align-self: flex-start;
        width: fit-content !important;
        max-width: 100%;
        flex: 0 0 auto;
        flex-wrap: wrap;
    }

    .form-grid {
        display: grid;
        gap: 10px 12px;
    }

    .form-grid.two-col {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .form-grid .form-field {
        margin-bottom: 0;
    }

    .form-field {
        min-width: 0;
    }

    .form-actions {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
        justify-content: flex-end;
        align-items: center;
    }

    .page-preview-head {
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
    }

    .page-preview-head-left {
        min-width: 0;
        flex: 1 1 auto;
        display: flex;
        align-items: center;
        gap: 10px;
        flex-wrap: wrap;
    }

    .page-preview-helper {
        min-width: 0;
        line-height: 1.35;
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

    .content-preview-first-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 10px;
    }

    .content-preview-workspace {
        display: block;
    }

    .content-preview-iframe-wrap {
        flex: 1 1 auto;
        min-height: clamp(560px, calc(100vh - 300px), 780px);
    }

    .content-preview-iframe {
        height: 100%;
        min-height: clamp(560px, calc(100vh - 300px), 780px);
    }

    :global(.overlay-panel.cms-section-editor-panel) {
        width: min(90vw, 800px);
    }

    :global(.cms-section-editor-panel .panel-header) {
        flex-wrap: nowrap;
        align-items: flex-start;
        justify-content: space-between;
        column-gap: 12px;
        row-gap: 4px;
        padding: 8px 14px;
    }

    :global(.cms-section-editor-panel .panel-header > :first-child) {
        flex: 1 1 auto;
        min-width: 0;
    }

    :global(.cms-section-editor-panel .panel-header > :last-child) {
        flex: 0 0 auto;
        margin-left: 8px;
    }

    .section-drawer-head {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .section-drawer-head-top {
        min-width: 0;
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .section-drawer-head-top > .txt-xs {
        flex: 0 0 auto;
        white-space: nowrap;
    }

    .section-drawer-name {
        min-width: 0;
        display: block;
        white-space: normal;
        overflow: hidden;
        text-overflow: ellipsis;
        font-size: 18px;
        line-height: 1.2;
        color: var(--txtPrimaryColor);
    }

    .section-drawer-support-row {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        min-width: 0;
    }

    .section-drawer-helper {
        line-height: 1.3;
    }

    .section-drawer-meta {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        flex-wrap: wrap;
        min-width: 0;
    }

    .section-drawer-close-btn {
        min-height: 30px;
        padding: 0 10px;
        align-self: flex-start;
    }

    .section-drawer-body {
        flex: 1 1 auto;
        min-height: 0;
        overflow: auto;
        padding: 10px 12px 10px;
    }

    .fallback-sections-list {
        display: flex;
        flex-direction: column;
        border: 1px solid var(--baseAlt2Color);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        overflow: hidden;
    }

    .fallback-section-item {
        border: 0;
        border-bottom: 1px solid var(--baseAlt2Color);
        background: transparent;
        text-align: left;
        padding: 10px 12px;
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .fallback-section-item:last-child {
        border-bottom: 0;
    }

    .fallback-section-item:nth-child(odd) {
        background: var(--baseColor);
    }

    .fallback-section-item:nth-child(even) {
        background: var(--baseAlt1Color);
    }

    .fallback-section-item:hover {
        background: color-mix(in srgb, var(--primaryColor) 4%, var(--baseColor));
    }

    .fallback-section-item.selected {
        background: color-mix(in srgb, var(--primaryColor) 9%, var(--baseColor));
        box-shadow: inset 3px 0 0 color-mix(in srgb, var(--primaryColor) 60%, transparent);
    }

    .fallback-section-title {
        font-size: 13px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .fallback-section-subtitle,
    .fallback-section-meta {
        font-size: 11px;
        color: var(--txtHintColor);
    }

    .sections-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .section-summary-pill {
        --labelHPadding: 7px;
        min-height: 18px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        line-height: 1;
        color: var(--txtHintColor);
        background: var(--baseColor);
    }

    .preview-focus-hint {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        font-size: var(--smFontSize);
        color: var(--txtHintColor);
        min-width: 0;
        white-space: normal;
    }

    .section-drawer-body :global(.pb-field) {
        margin: 10px 0;
        padding: 8px 10px 10px;
        border-radius: 8px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        background: transparent;
    }

    .section-drawer-body :global(.pb-label) {
        margin-bottom: 7px;
        font-size: 11.5px;
    }

    .section-drawer-body :global(.array-field),
    .section-drawer-body :global(.object-field) {
        margin-top: 10px !important;
        margin-bottom: 10px !important;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent) !important;
        background: var(--baseColor) !important;
        border-radius: 8px !important;
    }

    .section-drawer-body :global(.array-field__header),
    .section-drawer-body :global(.object-field__header),
    .section-drawer-body :global(.array-item__header) {
        background: color-mix(in srgb, var(--baseAlt1Color) 58%, var(--baseColor)) !important;
        border-bottom-color: color-mix(in srgb, var(--baseAlt2Color) 82%, transparent) !important;
    }

    .section-drawer-body :global(.array-field__items),
    .section-drawer-body :global(.object-field__body),
    .section-drawer-body :global(.array-item__body) {
        padding: 10px !important;
    }

    .section-drawer-body :global(.array-item) {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent) !important;
        border-radius: 8px !important;
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
    }

    .file-field-hint {
        font-size: var(--smFontSize);
        color: var(--txtHintColor);
    }

    @media (max-width: 1200px) {
        .content-workspace-grid {
            grid-template-columns: 1fr;
        }

        .content-preview-iframe-wrap {
            min-height: clamp(480px, 62vh, 700px);
        }

        .content-preview-iframe {
            min-height: clamp(480px, 62vh, 700px);
        }

        .pages-list-panel,
        .page-editor-panel {
            min-height: auto;
        }
    }

    @media (max-width: 840px) {
        .cms-section-panel {
            padding: calc(var(--baseSpacing) - 12px) calc(var(--baseSpacing) - 10px);
        }

        .form-grid.two-col {
            grid-template-columns: 1fr;
        }

        .page-preview-head {
            flex-direction: column;
            align-items: stretch;
            gap: 8px;
        }

        .page-preview-head-left {
            width: 100%;
            align-items: flex-start;
            flex-wrap: wrap;
            gap: 6px;
        }

        .page-preview-helper {
            white-space: normal;
            overflow: visible;
            text-overflow: unset;
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

        .content-preview-iframe-wrap {
            min-height: clamp(420px, 58vh, 620px);
        }

        .content-preview-iframe {
            min-height: clamp(420px, 58vh, 620px);
        }

        .pages-list-totals {
            text-align: left;
            white-space: normal;
        }
    }
</style>
