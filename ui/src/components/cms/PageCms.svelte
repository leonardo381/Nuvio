<script>
    import { onMount } from "svelte";
    import { querystring } from "svelte-spa-router";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import SchemaForm from "@/components/base/nuvio/schema/SchemaForm.svelte";
    import { pageTitle } from "@/stores/app";
    import {
        collections,
        collectionsLoadError,
        findCollectionByRequiredNames,
        hasCollectionsLoaded,
        isCollectionsLoading,
    } from "@/stores/collections";
    import { addSuccessToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { getWebsiteSettingsSchemaForRole, normalizeWebsiteSettingsValue } from "@/utils/WebsiteSettingsSchema";

    $pageTitle = "Website Content";

    const initialQueryParams = new URLSearchParams($querystring);
    const cmsTabPagesKey = "pages";
    const cmsTabSettingsKey = "settings";
    const pageEditorTabContentKey = "content";
    const pageEditorTabSeoKey = "seo";
    const pageSeoFilterAllKey = "all";
    const pageSeoFilterGoodKey = "good";
    const pageSeoFilterNeedsAttentionKey = "needs-attention";
    const pageSeoFilterMissingBasicsKey = "missing-basics";
    const pageSeoTabBasicKey = "basic";
    const pageSeoTabAdvancedKey = "advanced";
    const websiteIdentitySeoTabBasicKey = "basic";
    const websiteIdentitySeoTabLocalBusinessKey = "local-business";
    const websiteIdentitySeoTabAdvancedKey = "advanced";
    const sectionDefaultLanguageKey = "default";
    const sectionDiscardIntentCloseKey = "close";
    const sectionDiscardIntentSwitchLanguageKey = "switch-language";
    const clientSettingsRole = "client";
    const visibleClientSettingsKeys = new Set(["whatsapp", "contactForm", "newsletter", "booking", "reports", "i18n"]);
    const websiteSettingsAreaIdentitySeoKey = "identity-seo";
    const websiteSettingsAreaFeaturesKey = "features";
    const websiteSettingsFeatureOrder = ["whatsapp", "contactForm", "newsletter", "booking", "reports", "i18n"];
    const websiteSettingsLeadsFeatureKey = "leads";
    const websiteSettingsLeadsChannelKeys = ["contactForm", "whatsapp"];
    const websiteSettingsLeadsHelperText = "Configure how visitors can contact the business and become leads.";
    const websiteSettingsFeatureLabelByKey = {
        whatsapp: "WhatsApp",
        contactForm: "Contact form",
        newsletter: "Newsletter",
        booking: "Booking",
        reports: "Reports",
        i18n: "Internationalization",
    };
    const websiteSettingsAreaIconByKey = {
        [websiteSettingsAreaIdentitySeoKey]: "ri-profile-line",
        [websiteSettingsAreaFeaturesKey]: "ri-settings-4-line",
    };
    const websiteSettingsFeatureIconByKey = {
        whatsapp: "ri-whatsapp-line",
        contactForm: "ri-mail-line",
        newsletter: "ri-megaphone-line",
        booking: "ri-calendar-line",
        reports: "ri-bar-chart-2-line",
        i18n: "ri-global-line",
    };

    let websites = [];
    let pages = [];
    let blocks = [];
    let components = [];

    let selectedWebsiteId = initialQueryParams.get("cmsWebsite") || "";
    let selectedPageId = initialQueryParams.get("cmsPage") || "";
    let activeCmsTab = initialQueryParams.get("cmsTab") === cmsTabSettingsKey ? cmsTabSettingsKey : cmsTabPagesKey;
    let activePageEditorTab = pageEditorTabContentKey;
    let pageSeoFilter = pageSeoFilterAllKey;
    let activePageSeoTab = pageSeoTabBasicKey;
    let activePageSeoLanguageKey = sectionDefaultLanguageKey;
    let pageSearch = "";
    let focusedBlockId = "";
    let editingSectionId = "";
    let sectionEditorPanel;
    let isSectionDiscardConfirmOpen = false;
    let bypassSectionCloseConfirm = false;
    let sectionDiscardIntent = sectionDiscardIntentCloseKey;
    let pendingSectionLanguageKey = "";
    let activeSectionLanguageKey = sectionDefaultLanguageKey;

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
        seoSocialImageCurrent: "",
        seoSocialImageFile: null,
        seoCanonicalUrl: "",
        seoNoindex: false,
        seoExcludeFromSitemap: false,
        seoFocusKeyword: "",
    };

    let websiteSettingsFullDraft = {};
    let websiteSettingsDraft = {};
    let websiteIdentitySeoDraft = {
        seoTitle: "",
        seoDescription: "",
        seoTitleTemplate: "",
        seoTitleSeparator: "",
        seoCanonicalDomain: "",
        businessName: "",
        businessType: "",
        businessPrimaryCategory: "",
        businessPhone: "",
        businessEmail: "",
        businessAddress: "",
        businessCity: "",
        businessPostalCode: "",
        businessCountry: "",
        businessServiceArea: "",
        businessOpeningHours: "",
        businessGooglePlaceId: "",
        businessSocialProfiles: "",
        businessPriceRange: "",
        logoCurrent: "",
        seoImageCurrent: "",
        logoFile: null,
        seoImageFile: null,
    };
    let activeWebsiteSettingsArea = websiteSettingsAreaIdentitySeoKey;
    let activeWebsiteIdentitySeoTab = websiteIdentitySeoTabBasicKey;
    let activeWebsiteSettingsFeatureKey = "";

    let sectionPropsDraftById = {};
    let sectionTranslationsDraftById = {};
    let pageSeoTranslationsDraftByLanguage = {};
    let selectedEditingSectionDraftFormProps = {};
    let selectedEditingSectionDraftFormPropsSeed = "";
    let sectionErrorById = {};
    let isSavingSectionById = {};
    let pagePreviewReloadToken = 0;
    let pageSeoStatusById = new Map();
    let pageSeoCoverageCounts = {
        total: 0,
        good: 0,
        needsAttention: 0,
        missingBasics: 0,
    };

    let lastCollectionsKey = "";
    let lastPersistedContextKey = "";
    let lastPageSeedId = "";
    let lastSectionsSeedKey = "";
    let lastWebsiteSettingsSeedKey = "";

    const seoTitleLongThreshold = 65;
    const seoDescriptionLongThreshold = 170;
    const seoDescriptionShortThreshold = 70;
    const seoSeparatorLongThreshold = 3;
    const defaultSeoTitleSeparator = "|";
    const faqQuestionSummaryKeys = ["question", "title", "heading", "label", "name"];
    const faqAnswerSummaryKeys = ["answer", "description", "body", "content", "text", "html"];

    $: websitesCollection = findCollection(["websites", "Websites"]);
    $: pagesCollection = findCollection(["pages", "Pages"]);
    $: blocksCollection = findCollection(["blocks", "Blocks"]);
    $: componentsCollection = findCollection(["components", "Components"]);

    $: hasCmsCollections = !!websitesCollection && !!pagesCollection && !!blocksCollection;

    $: missingCollections = [];
    $: if (!websitesCollection) {
        missingCollections.push("websites");
    }
    $: if (!pagesCollection) {
        missingCollections.push("pages");
    }
    $: if (!blocksCollection) {
        missingCollections.push("blocks");
    }

    $: websiteNameField = resolveFieldName(websitesCollection, ["name", "title", "label"]);
    $: websiteSlugField = resolveFieldName(websitesCollection, ["slug"]);
    $: websiteDomainField = resolveFieldName(websitesCollection, ["domain", "url", "host"]);
    $: websiteSettingsField = resolveFieldName(websitesCollection, ["settings"]);
    $: websitesIncludeSettingsKey = websites.some((website) => hasOwnObjectKey(website, "settings"));
    $: resolvedWebsiteSettingsField = websiteSettingsField || (websitesIncludeSettingsKey ? "settings" : "");
    $: hasWebsiteSettingsField = !!resolvedWebsiteSettingsField;

    $: pageTitleField = resolveFieldName(pagesCollection, ["title", "name", "label"]);
    $: pageSlugField = resolveFieldName(pagesCollection, ["slug"]);
    $: pageEnabledField = resolveFieldName(pagesCollection, ["enabled", "published", "active"]);
    $: pageSeoTitleField = resolveFieldName(pagesCollection, ["seo_title", "seoTitle"]);
    $: pageSeoDescriptionField = resolveFieldName(pagesCollection, ["seo_description", "seoDescription"]);
    $: pageSeoTranslationsField = resolveFieldName(pagesCollection, ["seo_translations", "seoTranslations"]);
    $: pageSeoSocialImageField = resolveFieldName(pagesCollection, ["seo_social_image", "seoSocialImage"]);
    $: pageSeoCanonicalUrlField = resolveFieldName(pagesCollection, ["seo_canonical_url", "seoCanonicalUrl"]);
    $: pageSeoNoindexField = resolveFieldName(pagesCollection, ["seo_noindex", "seoNoindex"]);
    $: pageSeoExcludeFromSitemapField = resolveFieldName(
        pagesCollection,
        ["seo_exclude_from_sitemap", "seoExcludeFromSitemap"],
    );
    $: pageSeoFocusKeywordField = resolveFieldName(pagesCollection, ["seo_focus_keyword", "seoFocusKeyword"]);
    $: selectedPageIncludesSeoTranslationsKey = hasOwnObjectKey(selectedPage, "seo_translations");
    $: hasPageSeoTranslationsFieldEvidence = !!pageSeoTranslationsField
        || selectedPageIncludesSeoTranslationsKey;
    $: effectivePageSeoTranslationsField = hasPageSeoTranslationsFieldEvidence
        ? (pageSeoTranslationsField || "seo_translations")
        : "";

    $: websiteLogoField = resolveFieldName(websitesCollection, ["logo"]);
    $: websiteSeoTitleField = resolveFieldName(websitesCollection, ["seoTitle", "seo_title"]);
    $: websiteSeoDescriptionField = resolveFieldName(websitesCollection, ["seoDescription", "seo_description"]);
    $: websiteSeoImageField = resolveFieldName(websitesCollection, ["seoImage", "seo_image"]);
    $: websiteSeoTitleTemplateField = resolveFieldName(websitesCollection, ["seo_title_template", "seoTitleTemplate"]);
    $: websiteSeoTitleSeparatorField = resolveFieldName(websitesCollection, ["seo_title_separator", "seoTitleSeparator"]);
    $: websiteSeoCanonicalDomainField = resolveFieldName(websitesCollection, ["seo_canonical_domain", "seoCanonicalDomain"]);
    $: websiteBusinessNameField = resolveFieldName(websitesCollection, ["business_name", "businessName"]);
    $: websiteBusinessTypeField = resolveFieldName(websitesCollection, ["business_type", "businessType"]);
    $: websiteBusinessPrimaryCategoryField = resolveFieldName(websitesCollection, ["business_primary_category", "businessPrimaryCategory"]);
    $: websiteBusinessPhoneField = resolveFieldName(websitesCollection, ["business_phone", "businessPhone"]);
    $: websiteBusinessEmailField = resolveFieldName(websitesCollection, ["business_email", "businessEmail"]);
    $: websiteBusinessAddressField = resolveFieldName(websitesCollection, ["business_address", "businessAddress"]);
    $: websiteBusinessCityField = resolveFieldName(websitesCollection, ["business_city", "businessCity"]);
    $: websiteBusinessPostalCodeField = resolveFieldName(websitesCollection, ["business_postal_code", "businessPostalCode"]);
    $: websiteBusinessCountryField = resolveFieldName(websitesCollection, ["business_country", "businessCountry"]);
    $: websiteBusinessServiceAreaField = resolveFieldName(websitesCollection, ["business_service_area", "businessServiceArea"]);
    $: websiteBusinessOpeningHoursField = resolveFieldName(websitesCollection, ["business_opening_hours", "businessOpeningHours"]);
    $: websiteBusinessGooglePlaceIdField = resolveFieldName(websitesCollection, ["business_google_place_id", "businessGooglePlaceId"]);
    $: websiteBusinessSocialProfilesField = resolveFieldName(websitesCollection, ["business_social_profiles", "businessSocialProfiles"]);
    $: websiteBusinessPriceRangeField = resolveFieldName(websitesCollection, ["business_price_range", "businessPriceRange"]);
    $: hasWebsiteIdentitySeoFields = !!(
        websiteLogoField ||
        websiteSeoTitleField ||
        websiteSeoDescriptionField ||
        websiteSeoImageField ||
        websiteSeoTitleTemplateField ||
        websiteSeoTitleSeparatorField ||
        websiteSeoCanonicalDomainField ||
        websiteBusinessNameField ||
        websiteBusinessTypeField ||
        websiteBusinessPrimaryCategoryField ||
        websiteBusinessPhoneField ||
        websiteBusinessEmailField ||
        websiteBusinessAddressField ||
        websiteBusinessCityField ||
        websiteBusinessPostalCodeField ||
        websiteBusinessCountryField ||
        websiteBusinessServiceAreaField ||
        websiteBusinessOpeningHoursField ||
        websiteBusinessGooglePlaceIdField ||
        websiteBusinessSocialProfilesField ||
        websiteBusinessPriceRangeField
    );

    $: blockTitleField = resolveFieldName(blocksCollection, ["title", "name", "label"]);
    $: blockPropsField = resolveFieldName(blocksCollection, ["props"]);
    $: blockTranslationsField = resolveFieldName(blocksCollection, ["translations"]);
    $: blocksIncludeTranslationsKey = blocks.some((block) => hasOwnObjectKey(block, "translations"));
    $: hasBlockTranslationsFieldEvidence = !!blockTranslationsField
        || blocksIncludeTranslationsKey;
    $: resolvedBlockTranslationsField = hasBlockTranslationsFieldEvidence
        ? (blockTranslationsField || "translations")
        : "";
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
    $: sectionEditorSettingsSource = (() => {
        if (isPlainObject(websiteSettingsFullDraft) && Object.keys(websiteSettingsFullDraft).length > 0) {
            return websiteSettingsFullDraft;
        }

        const fallbackSettingsValue = resolvedWebsiteSettingsField
            ? selectedWebsite?.[resolvedWebsiteSettingsField]
            : (hasOwnObjectKey(selectedWebsite, "settings") ? selectedWebsite?.settings : {});

        return normalizeWebsiteSettingsValue(fallbackSettingsValue);
    })();
    $: sectionEditorConfiguredLanguages = getSectionEditorConfiguredLanguages(sectionEditorSettingsSource);
    $: sectionEditorDefaultLanguage = sectionEditorConfiguredLanguages[0] || null;
    $: sectionEditorDefaultLanguageLabel = sectionEditorDefaultLanguage?.label
        || (sectionEditorDefaultLanguage?.code ? sectionEditorDefaultLanguage.code.toUpperCase() : "Primary");
    $: sectionEditorTranslationLanguages = sectionEditorConfiguredLanguages.slice(1);
    $: pageSeoConfiguredLanguages = sectionEditorConfiguredLanguages;
    $: pageSeoDefaultLanguage = pageSeoConfiguredLanguages[0] || null;
    $: pageSeoDefaultLanguageLabel = pageSeoDefaultLanguage?.label
        || (pageSeoDefaultLanguage?.code ? pageSeoDefaultLanguage.code.toUpperCase() : "Primary");
    $: pageSeoTranslationLanguages = pageSeoConfiguredLanguages.slice(1);
    $: pageSeoSupportsTranslations = !!effectivePageSeoTranslationsField && pageSeoTranslationLanguages.length > 0;
    $: effectiveBlockTranslationsField = resolvedBlockTranslationsField;
    $: sectionEditorSupportsTranslations = !!effectiveBlockTranslationsField && sectionEditorTranslationLanguages.length > 0;
    $: roleScopedSettingsFields = getWebsiteSettingsSchemaForRole(clientSettingsRole, websiteSettingsFullDraft).fields;
    $: clientWebsiteSettingsFields = filterClientWebsiteSettingsFields(roleScopedSettingsFields);
    $: websiteSettingsFieldsByKey = new Map(
        (clientWebsiteSettingsFields || []).map((field) => [normalizeString(field?.key), field]),
    );
    $: availableWebsiteSettingsFeatures = (() => {
        const isFeatureAvailable = (key) => websiteSettingsFullDraft?.featureFlags?.[key] !== false;
        const features = [];

        const leadsChannels = websiteSettingsLeadsChannelKeys
            .map((key) => {
                if (!isFeatureAvailable(key)) {
                    return null;
                }

                const field = websiteSettingsFieldsByKey.get(key) || null;
                return {
                    key,
                    label: field?.label || websiteSettingsFeatureLabelByKey?.[key] || "Channel",
                    icon: websiteSettingsFeatureIconByKey?.[key] || "ri-settings-3-line",
                    field,
                    hasEditableFields: !!field,
                };
            })
            .filter(Boolean);

        if (leadsChannels.length) {
            features.push({
                key: websiteSettingsLeadsFeatureKey,
                label: "Leads",
                icon: "ri-user-line",
                field: null,
                hasEditableFields: leadsChannels.some((channel) => channel.hasEditableFields),
                groupedFeatures: leadsChannels,
            });
        }

        for (const key of websiteSettingsFeatureOrder) {
            if (websiteSettingsLeadsChannelKeys.includes(key)) {
                continue;
            }

            if (!isFeatureAvailable(key)) {
                continue;
            }

            const field = websiteSettingsFieldsByKey.get(key) || null;
            features.push({
                key,
                label: field?.label || websiteSettingsFeatureLabelByKey?.[key] || "Feature",
                icon: websiteSettingsFeatureIconByKey?.[key] || "ri-settings-3-line",
                field,
                hasEditableFields: !!field,
            });
        }

        return features;
    })();
    $: activeWebsiteSettingsFeature = availableWebsiteSettingsFeatures.find(
        (feature) => feature.key === activeWebsiteSettingsFeatureKey,
    ) || null;
    $: activeWebsiteSettingsFeatureField = activeWebsiteSettingsFeature?.field || null;
    $: activeWebsiteSettingsFeatureValue = activeWebsiteSettingsFeatureField
        ? buildWebsiteSettingsFeatureFormValue(activeWebsiteSettingsFeatureField.key)
        : {};
    $: if (!activeWebsiteSettingsFeatureKey && availableWebsiteSettingsFeatures.length) {
        activeWebsiteSettingsFeatureKey = getDefaultWebsiteSettingsFeatureKey();
    } else if (
        activeWebsiteSettingsFeatureKey &&
        !availableWebsiteSettingsFeatures.some((feature) => feature.key === activeWebsiteSettingsFeatureKey)
    ) {
        activeWebsiteSettingsFeatureKey = getDefaultWebsiteSettingsFeatureKey();
    } else if (!availableWebsiteSettingsFeatures.length && activeWebsiteSettingsFeatureKey) {
        activeWebsiteSettingsFeatureKey = "";
    }

    $: websitePublicUrl = getWebsitePublicUrl(selectedWebsite);
    $: selectedWebsiteSlug = normalizeString(websiteSlugField ? selectedWebsite?.[websiteSlugField] : "");
    $: selectedPageSlug = normalizeString(pageSlugField ? selectedPage?.[pageSlugField] : "");
    $: pagePreviewUrl = buildPagePreviewUrl(
        selectedWebsiteSlug,
        selectedPageSlug,
        activeSectionTranslationLanguageCode,
    );
    $: pagePreviewFocusedUrl = buildPagePreviewFocusedUrl(pagePreviewUrl, focusedBlockId);
    $: pagePreviewIframeSrc = buildPreviewIframeSrc(pagePreviewFocusedUrl, pagePreviewReloadToken);
    $: activePageSeoTranslationLanguageCode = (
        activePageSeoLanguageKey !== sectionDefaultLanguageKey && pageSeoSupportsTranslations
    )
        ? activePageSeoLanguageKey
        : "";
    $: activePageSeoTranslationDraft = activePageSeoTranslationLanguageCode
        ? getPageSeoTranslationDraft(activePageSeoTranslationLanguageCode)
        : { title: "", description: "" };
    $: pageSeoTitleInputValue = activePageSeoTranslationLanguageCode
        ? `${activePageSeoTranslationDraft?.title || ""}`
        : `${pageEditForm?.seoTitle || ""}`;
    $: pageSeoDescriptionInputValue = activePageSeoTranslationLanguageCode
        ? `${activePageSeoTranslationDraft?.description || ""}`
        : `${pageEditForm?.seoDescription || ""}`;
    $: pageSeoTitleText = normalizeString(pageSeoTitleInputValue);
    $: pageSeoDescriptionText = toSeoPlainText(pageSeoDescriptionInputValue);
    $: pageSeoTitleLength = pageSeoTitleText.length;
    $: pageSeoDescriptionLength = pageSeoDescriptionText.length;
    $: pageSeoFocusKeywordText = normalizeString(pageEditForm?.seoFocusKeyword);
    $: pageSeoNoindexValue = toBooleanValue(pageEditForm?.seoNoindex);
    $: pageSeoExcludeFromSitemapValue = toBooleanValue(pageEditForm?.seoExcludeFromSitemap);
    $: pageSeoCanonicalUrlText = normalizeString(pageEditForm?.seoCanonicalUrl);
    $: pageSeoHasSocialImage = !!normalizeString(pageEditForm?.seoSocialImageCurrent) || !!pageEditForm?.seoSocialImageFile;
    $: pageSeoHasGlobalSocialImage = !!normalizeString(websiteIdentitySeoDraft?.seoImageCurrent) || !!websiteIdentitySeoDraft?.seoImageFile;
    $: pageSeoSocialImagePreviewUrl = getCollectionFileUrl(selectedPage, pageEditForm?.seoSocialImageCurrent);
    $: websiteLogoPreviewUrl = getCollectionFileUrl(selectedWebsite, websiteIdentitySeoDraft?.logoCurrent);
    $: websiteSeoImagePreviewUrl = getCollectionFileUrl(selectedWebsite, websiteIdentitySeoDraft?.seoImageCurrent);
    $: globalSeoTitleText = normalizeString(websiteIdentitySeoDraft?.seoTitle);
    $: globalSeoDescriptionText = toSeoPlainText(websiteIdentitySeoDraft?.seoDescription);
    $: globalSeoTitleTemplateText = normalizeString(websiteIdentitySeoDraft?.seoTitleTemplate);
    $: globalSeoTitleSeparatorText = normalizeString(websiteIdentitySeoDraft?.seoTitleSeparator);
    $: globalSeoCanonicalDomainText = normalizeString(websiteIdentitySeoDraft?.seoCanonicalDomain);
    $: globalSeoTitleLength = globalSeoTitleText.length;
    $: globalSeoDescriptionLength = globalSeoDescriptionText.length;
    $: pageSeoPreviewTitle = pageSeoTitleText || getPageTitleText(selectedPage) || globalSeoTitleText || getWebsiteLabel(selectedWebsite);
    $: pageSeoPreviewPath = getPageSeoPreviewPath(pagePreviewUrl, selectedWebsiteSlug, selectedPageSlug);
    $: pageSeoPreviewDescription = pageSeoDescriptionText || globalSeoDescriptionText || "No SEO description provided yet.";
    $: pageSeoHasTitleFallback = !!getPageSeoTitleFallbackSource(selectedPage, selectedWebsite, globalSeoTitleText);
    $: pageSeoHasDescriptionFallback = !!globalSeoDescriptionText;
    $: pageSeoHasFaqStructuredData = hasFaqStructuredDataCandidate(blocks);
    $: globalSeoSiteName = globalSeoTitleText || getWebsiteNameText(selectedWebsite) || getWebsiteLabel(selectedWebsite) || "Site name";
    $: globalSeoPreviewTitle = buildGlobalSeoPreviewTitle({
        template: globalSeoTitleTemplateText,
        separator: globalSeoTitleSeparatorText,
        siteName: globalSeoSiteName,
        pageName: "Sample page",
    });
    $: globalSeoPreviewDescription = globalSeoDescriptionText || "No global SEO description provided yet.";
    $: globalSeoPreviewUrl = getWebsiteSeoPreviewUrl(websitePublicUrl, selectedWebsiteSlug);
    $: localBusinessNameText = normalizeString(websiteIdentitySeoDraft?.businessName);
    $: localBusinessTypeText = normalizeString(websiteIdentitySeoDraft?.businessType);
    $: localBusinessPrimaryCategoryText = normalizeString(websiteIdentitySeoDraft?.businessPrimaryCategory);
    $: localBusinessPhoneText = normalizeString(websiteIdentitySeoDraft?.businessPhone);
    $: localBusinessEmailText = normalizeString(websiteIdentitySeoDraft?.businessEmail);
    $: localBusinessAddressText = normalizeString(websiteIdentitySeoDraft?.businessAddress);
    $: localBusinessCityText = normalizeString(websiteIdentitySeoDraft?.businessCity);
    $: localBusinessPostalCodeText = normalizeString(websiteIdentitySeoDraft?.businessPostalCode);
    $: localBusinessCountryText = normalizeString(websiteIdentitySeoDraft?.businessCountry);
    $: localBusinessServiceAreaText = normalizeString(websiteIdentitySeoDraft?.businessServiceArea);
    $: localBusinessOpeningHoursText = normalizeString(websiteIdentitySeoDraft?.businessOpeningHours);
    $: localBusinessGooglePlaceIdText = normalizeString(websiteIdentitySeoDraft?.businessGooglePlaceId);
    $: localBusinessSocialProfilesText = normalizeString(websiteIdentitySeoDraft?.businessSocialProfiles);
    $: localBusinessPriceRangeText = normalizeString(websiteIdentitySeoDraft?.businessPriceRange);
    $: localBusinessReviewsExpected = websiteSettingsFullDraft?.featureFlags?.reviews !== false;
    $: localBusinessHasContactInput = !!websiteBusinessPhoneField || !!websiteBusinessEmailField;
    $: localBusinessHasLocationInput =
        (!!websiteBusinessAddressField && !!websiteBusinessCityField && !!websiteBusinessCountryField) || !!websiteBusinessServiceAreaField;
    $: pageSeoChecks = buildPageSeoChecks({
        titleText: pageSeoTitleText,
        descriptionText: pageSeoDescriptionText,
        titleLength: pageSeoTitleLength,
        descriptionLength: pageSeoDescriptionLength,
        hasTitle: !!pageSeoTitleText,
        hasDescription: !!pageSeoDescriptionText,
        hasTitleFallback: pageSeoHasTitleFallback,
        hasDescriptionFallback: pageSeoHasDescriptionFallback,
        hasSocialImage: pageSeoHasSocialImage,
        hasGlobalSocialImage: pageSeoHasGlobalSocialImage,
        focusKeyword: pageSeoFocusKeywordText,
        noindex: pageSeoNoindexValue,
        excludeFromSitemap: pageSeoExcludeFromSitemapValue,
        canonicalUrl: pageSeoCanonicalUrlText,
        hasCanonicalField: !!pageSeoCanonicalUrlField,
        hasFaqStructuredData: pageSeoHasFaqStructuredData,
    });
    $: globalSeoChecks = buildGlobalSeoChecks({
        titleLength: globalSeoTitleLength,
        descriptionLength: globalSeoDescriptionLength,
        hasTitle: !!globalSeoTitleText,
        hasDescription: !!globalSeoDescriptionText,
        hasSocialImageField: !!websiteSeoImageField,
        hasSocialImage: pageSeoHasGlobalSocialImage,
        hasTitleTemplateField: !!websiteSeoTitleTemplateField,
        titleTemplate: globalSeoTitleTemplateText,
        hasTitleSeparatorField: !!websiteSeoTitleSeparatorField,
        titleSeparator: globalSeoTitleSeparatorText,
        hasCanonicalDomainField: !!websiteSeoCanonicalDomainField,
        canonicalDomain: globalSeoCanonicalDomainText,
    });
    $: localBusinessSeoChecks = buildLocalBusinessSeoChecks({
        hasBusinessNameField: !!websiteBusinessNameField,
        businessName: localBusinessNameText,
        hasPrimaryCategoryField: !!websiteBusinessPrimaryCategoryField,
        primaryCategory: localBusinessPrimaryCategoryText,
        hasPhoneField: !!websiteBusinessPhoneField,
        phone: localBusinessPhoneText,
        hasEmailField: !!websiteBusinessEmailField,
        email: localBusinessEmailText,
        hasAddressField: !!websiteBusinessAddressField,
        address: localBusinessAddressText,
        hasCityField: !!websiteBusinessCityField,
        city: localBusinessCityText,
        hasCountryField: !!websiteBusinessCountryField,
        country: localBusinessCountryText,
        hasServiceAreaField: !!websiteBusinessServiceAreaField,
        serviceArea: localBusinessServiceAreaText,
        hasOpeningHoursField: !!websiteBusinessOpeningHoursField,
        openingHours: localBusinessOpeningHoursText,
        hasGooglePlaceIdField: !!websiteBusinessGooglePlaceIdField,
        googlePlaceId: localBusinessGooglePlaceIdText,
        hasSocialProfilesField: !!websiteBusinessSocialProfilesField,
        socialProfiles: localBusinessSocialProfilesText,
        hasPriceRangeField: !!websiteBusinessPriceRangeField,
        priceRange: localBusinessPriceRangeText,
        expectsGooglePlaceId: localBusinessReviewsExpected,
    });
    $: pageSeoCheckCounts = getSeoCheckCounts(pageSeoChecks);
    $: globalSeoCheckCounts = getSeoCheckCounts(globalSeoChecks);
    $: localBusinessSeoCheckCounts = getSeoCheckCounts(localBusinessSeoChecks);
    $: pageSeoWarningChecks = (pageSeoChecks || []).filter((check) => `${check?.level || ""}` === "warning");
    $: pageSeoSuggestionChecks = (pageSeoChecks || []).filter((check) => `${check?.level || ""}` === "info");
    $: globalSeoWarningChecks = (globalSeoChecks || []).filter((check) => `${check?.level || ""}` === "warning");
    $: globalSeoSuggestionChecks = (globalSeoChecks || []).filter((check) => `${check?.level || ""}` === "info");
    $: localBusinessSeoWarningChecks = (localBusinessSeoChecks || []).filter((check) => `${check?.level || ""}` === "warning");
    $: localBusinessSeoSuggestionChecks = (localBusinessSeoChecks || []).filter((check) => `${check?.level || ""}` === "info");
    $: pageSeoHealthCompactSummary = getSeoHealthCompactSummary(pageSeoCheckCounts);
    $: globalSeoHealthCompactSummary = getSeoHealthCompactSummary(globalSeoCheckCounts);
    $: localBusinessSeoHealthCompactSummary = getSeoHealthCompactSummary(localBusinessSeoCheckCounts);
    $: pageSeoHealthStatus = getPageSeoHealthStatus({
        hasTitle: !!pageSeoTitleText,
        hasDescription: !!pageSeoDescriptionText,
        hasTitleFallback: pageSeoHasTitleFallback,
        hasDescriptionFallback: pageSeoHasDescriptionFallback,
        warningCount: pageSeoCheckCounts.warnings,
    });
    $: globalSeoHealthStatus = getGlobalSeoHealthStatus({
        hasTitle: !!globalSeoTitleText,
        hasDescription: !!globalSeoDescriptionText,
        warningCount: globalSeoCheckCounts.warnings,
    });
    $: localBusinessSeoHealthStatus = getLocalBusinessSeoHealthStatus({
        hasBusinessNameField: !!websiteBusinessNameField,
        hasBusinessName: !!localBusinessNameText,
        hasContactInputFields: localBusinessHasContactInput,
        hasContactSignal: !!localBusinessPhoneText || !!localBusinessEmailText,
        hasLocationInputFields: localBusinessHasLocationInput,
        hasLocationSignal: (!!localBusinessAddressText && !!localBusinessCityText && !!localBusinessCountryText) || !!localBusinessServiceAreaText,
        warningCount: localBusinessSeoCheckCounts.warnings,
        infoCount: localBusinessSeoCheckCounts.infos,
    });
    $: globalSeoCheckSummary = getSeoCheckSummaryText(globalSeoCheckCounts);
    $: localBusinessSeoCheckSummary = getSeoCheckSummaryText(localBusinessSeoCheckCounts);
    $: websiteSeoAdvancedImpactChecks = buildWebsiteSeoAdvancedImpactChecks({
        hasTitleTemplateField: !!websiteSeoTitleTemplateField,
        titleTemplate: globalSeoTitleTemplateText,
        hasTitleSeparatorField: !!websiteSeoTitleSeparatorField,
        titleSeparator: globalSeoTitleSeparatorText,
        hasCanonicalDomainField: !!websiteSeoCanonicalDomainField,
        canonicalDomain: globalSeoCanonicalDomainText,
    });
    $: websiteSeoAdvancedImpactWarningChecks = (websiteSeoAdvancedImpactChecks || []).filter(
        (check) => `${check?.level || ""}` === "warning",
    );
    $: websiteSeoAdvancedImpactInfoChecks = (websiteSeoAdvancedImpactChecks || []).filter(
        (check) => `${check?.level || ""}` === "info",
    );
    $: normalizedPageSearch = normalizeString(pageSearch).toLowerCase();
    $: activePagesCount = pageEnabledField ? pages.filter((record) => isPageActive(record)).length : 0;
    $: inactivePagesCount = pageEnabledField ? Math.max(0, pages.length - activePagesCount) : 0;
    $: pageSeoStatusById = new Map((pages || []).map((record) => [record?.id || "", getPageSeoCoverageStatus(record)]));
    $: pageSeoCoverageCounts = getPageSeoCoverageCounts(pages, pageSeoStatusById);
    $: filteredPages = pages.filter((record) => {
        if (pageSeoFilter !== pageSeoFilterAllKey) {
            const seoStatus = pageSeoStatusById.get(record?.id || "") || getPageSeoCoverageStatus(record);
            if (seoStatus.key !== pageSeoFilter) {
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
    $: if (
        pageSeoFilter !== pageSeoFilterAllKey &&
        pageSeoFilter !== pageSeoFilterGoodKey &&
        pageSeoFilter !== pageSeoFilterNeedsAttentionKey &&
        pageSeoFilter !== pageSeoFilterMissingBasicsKey
    ) {
        pageSeoFilter = pageSeoFilterAllKey;
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
        activePageSeoLanguageKey = sectionDefaultLanguageKey;
        pageEditForm = {
            seoTitle: pageSeoTitleField ? `${selectedPage?.[pageSeoTitleField] || ""}` : "",
            seoDescription: pageSeoDescriptionField ? `${selectedPage?.[pageSeoDescriptionField] || ""}` : "",
            seoSocialImageCurrent: pageSeoSocialImageField ? toSingleFileName(selectedPage?.[pageSeoSocialImageField]) : "",
            seoSocialImageFile: null,
            seoCanonicalUrl: pageSeoCanonicalUrlField ? `${selectedPage?.[pageSeoCanonicalUrlField] || ""}` : "",
            seoNoindex: pageSeoNoindexField ? toBooleanValue(selectedPage?.[pageSeoNoindexField]) : false,
            seoExcludeFromSitemap: pageSeoExcludeFromSitemapField
                ? toBooleanValue(selectedPage?.[pageSeoExcludeFromSitemapField])
                : false,
            seoFocusKeyword: pageSeoFocusKeywordField ? `${selectedPage?.[pageSeoFocusKeywordField] || ""}` : "",
        };
        pageSeoTranslationsDraftByLanguage = toPageSeoTranslationsDraftByLanguage(
            effectivePageSeoTranslationsField
                ? selectedPage?.[effectivePageSeoTranslationsField]
                : (hasOwnObjectKey(selectedPage, "seo_translations") ? selectedPage?.seo_translations : {}),
        );
        pageError = "";
    }

    $: if (!selectedPage?.id) {
        lastPageSeedId = "";
        focusedBlockId = "";
        editingSectionId = "";
        activePageEditorTab = pageEditorTabContentKey;
        activePageSeoLanguageKey = sectionDefaultLanguageKey;
        pageSeoTranslationsDraftByLanguage = {};
    }

    $: if (activePageEditorTab !== pageEditorTabContentKey && activePageEditorTab !== pageEditorTabSeoKey) {
        activePageEditorTab = pageEditorTabContentKey;
    }
    $: if (!pageSeoSupportsTranslations && activePageSeoLanguageKey !== sectionDefaultLanguageKey) {
        activePageSeoLanguageKey = sectionDefaultLanguageKey;
    }
    $: if (
        activePageSeoLanguageKey !== sectionDefaultLanguageKey
        && !pageSeoTranslationLanguages.some((language) => language.code === activePageSeoLanguageKey)
    ) {
        activePageSeoLanguageKey = sectionDefaultLanguageKey;
    }
    $: if (activePageSeoTranslationLanguageCode && activePageSeoTab === pageSeoTabAdvancedKey) {
        activePageSeoTab = pageSeoTabBasicKey;
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
        const nextSeedKey = `${blockPropsField || ""}|${effectiveBlockTranslationsField || ""}|${blocks.map((block) => `${block.id}:${block?.updated || ""}`).join("|")}`;
        if (nextSeedKey !== lastSectionsSeedKey) {
            lastSectionsSeedKey = nextSeedKey;

            const nextDraft = {};
            const nextTranslationDraft = {};
            const nextErrors = {};
            const nextSaving = {};

            for (const block of blocks) {
                nextDraft[block.id] = toPropsObject(blockPropsField ? block?.[blockPropsField] : {});
                nextTranslationDraft[block.id] = toTranslationsDraftByLanguage(
                    effectiveBlockTranslationsField ? block?.[effectiveBlockTranslationsField] : {},
                );
                nextErrors[block.id] = "";
                nextSaving[block.id] = false;
            }

            sectionPropsDraftById = nextDraft;
            sectionTranslationsDraftById = nextTranslationDraft;
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
    $: selectedEditingSectionSchemaFieldKeys = new Set(
        selectedEditingSectionFields
            .map((field) => normalizeString(field?.key).toLowerCase())
            .filter(Boolean),
    );
    $: selectedEditingSectionSchemaFieldKeysFingerprint = Array.from(selectedEditingSectionSchemaFieldKeys)
        .sort()
        .join("|");
    $: selectedEditingSectionIndex = selectedEditingSection
        ? blocks.findIndex((block) => `${block?.id || ""}` === `${selectedEditingSection?.id || ""}`)
        : -1;
    $: selectedEditingSectionSubtitle = selectedEditingSection
        ? getSectionDescription(selectedEditingSection, Math.max(selectedEditingSectionIndex, 0))
        : "";
    $: selectedEditingSectionSummaryPills = selectedEditingSection
        ? getSectionSummaryPills(selectedEditingSection, selectedEditingSectionFields)
        : [];
    $: activeSectionTranslationLanguageCode = (
        activeSectionLanguageKey !== sectionDefaultLanguageKey && sectionEditorSupportsTranslations
    )
        ? activeSectionLanguageKey
        : "";
    $: selectedEditingSectionOriginalProps = selectedEditingSection
        ? toPropsObject(blockPropsField ? selectedEditingSection?.[blockPropsField] : {})
        : {};
    $: selectedEditingSectionDraftProps = selectedEditingSection
        ? toPropsObject(sectionPropsDraftById?.[selectedEditingSection.id])
        : {};
    $: selectedEditingSectionOriginalTranslationProps = (
        selectedEditingSection && activeSectionTranslationLanguageCode
    )
        ? getTranslationPropsByLanguage(
            toTranslationsRecordObject(
                effectiveBlockTranslationsField ? selectedEditingSection?.[effectiveBlockTranslationsField] : {},
            ),
            activeSectionTranslationLanguageCode,
        )
        : {};
    $: selectedEditingSectionDraftTranslationProps = (
        selectedEditingSection && activeSectionTranslationLanguageCode
    )
        ? getSectionTranslationDraftProps(selectedEditingSection.id, activeSectionTranslationLanguageCode)
        : {};
    $: selectedEditingSectionOriginalActiveProps = activeSectionTranslationLanguageCode
        ? selectedEditingSectionOriginalTranslationProps
        : selectedEditingSectionOriginalProps;
    $: selectedEditingSectionDraftActiveProps = activeSectionTranslationLanguageCode
        ? selectedEditingSectionDraftTranslationProps
        : selectedEditingSectionDraftProps;
    $: {
        if (!selectedEditingSection) {
            selectedEditingSectionDraftFormPropsSeed = "";
            selectedEditingSectionDraftFormProps = {};
        } else {
            const draftSeed = `${selectedEditingSection?.id || ""}|${activeSectionTranslationLanguageCode || sectionDefaultLanguageKey}|${selectedEditingSectionSchemaFieldKeysFingerprint}|${stableSerializeForDirtyCheck(selectedEditingSectionDraftActiveProps)}`;
            if (draftSeed !== selectedEditingSectionDraftFormPropsSeed) {
                selectedEditingSectionDraftFormPropsSeed = draftSeed;
                selectedEditingSectionDraftFormProps = sanitizeSectionDraftPropsForForm(
                    selectedEditingSectionDraftActiveProps,
                    selectedEditingSectionSchemaFieldKeys,
                );
            }
        }
    }
    $: isSectionEditorDirty = !!selectedEditingSection
        && stableSerializeForDirtyCheck(selectedEditingSectionDraftActiveProps)
            !== stableSerializeForDirtyCheck(selectedEditingSectionOriginalActiveProps);
    $: if (!sectionEditorSupportsTranslations && activeSectionLanguageKey !== sectionDefaultLanguageKey) {
        activeSectionLanguageKey = sectionDefaultLanguageKey;
    }
    $: if (
        activeSectionLanguageKey !== sectionDefaultLanguageKey
        && !sectionEditorTranslationLanguages.some((language) => language.code === activeSectionLanguageKey)
    ) {
        activeSectionLanguageKey = sectionDefaultLanguageKey;
    }
    $: if (!selectedEditingSection && isSectionDiscardConfirmOpen) {
        isSectionDiscardConfirmOpen = false;
        bypassSectionCloseConfirm = false;
        sectionDiscardIntent = sectionDiscardIntentCloseKey;
        pendingSectionLanguageKey = "";
        activeSectionLanguageKey = sectionDefaultLanguageKey;
    }

    $: {
        const nextWebsiteSettingsSeedKey = `${selectedWebsite?.id || ""}|${selectedWebsite?.updated || ""}|${resolvedWebsiteSettingsField || ""}`;
        if (nextWebsiteSettingsSeedKey !== lastWebsiteSettingsSeedKey) {
            lastWebsiteSettingsSeedKey = nextWebsiteSettingsSeedKey;
            initializeWebsiteSettingsDraft();
            initializeWebsiteIdentitySeoDraft();
            activeWebsiteSettingsArea = websiteSettingsAreaIdentitySeoKey;
            activeWebsiteSettingsFeatureKey = "";
        }
    }

    function findCollection(nameOrAliases) {
        const aliases = Array.isArray(nameOrAliases) ? nameOrAliases : [nameOrAliases];
        return findCollectionByRequiredNames($collections, aliases);
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

    function toBooleanValue(value) {
        if (typeof value === "boolean") {
            return value;
        }

        if (typeof value === "string") {
            if (value.toLowerCase() === "true") {
                return true;
            }
            if (value.toLowerCase() === "false") {
                return false;
            }
        }

        return !!value;
    }

    function decodeBasicHtmlEntities(value) {
        return `${value || ""}`
            .replace(/&nbsp;/gi, " ")
            .replace(/&amp;/gi, "&")
            .replace(/&quot;/gi, "\"")
            .replace(/&#39;/gi, "'")
            .replace(/&lt;/gi, "<")
            .replace(/&gt;/gi, ">");
    }

    function toSeoPlainText(value) {
        const raw = `${value ?? ""}`;
        if (!raw) {
            return "";
        }

        const withoutMarkup = raw
            .replace(/<style[\s\S]*?<\/style>/gi, " ")
            .replace(/<script[\s\S]*?<\/script>/gi, " ")
            .replace(/<\/(p|div|h[1-6]|li|blockquote|section|article)>/gi, " ")
            .replace(/<br\s*\/?>/gi, " ")
            .replace(/<[^>]+>/g, " ");

        return normalizeString(
            decodeBasicHtmlEntities(withoutMarkup).replace(/\s+/g, " "),
        );
    }

    function textContainsKeyword(text, keyword) {
        const normalizedText = normalizeString(text).toLowerCase();
        const normalizedKeyword = normalizeString(keyword).toLowerCase();
        if (!normalizedText || !normalizedKeyword) {
            return false;
        }
        return normalizedText.includes(normalizedKeyword);
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

    function normalizeLanguageCode(value) {
        const normalized = normalizeString(value).toLowerCase();
        if (!normalized) {
            return "";
        }

        if (!/^[a-z]{2,3}(?:-[a-z0-9]{2,8})*$/i.test(normalized)) {
            return "";
        }

        return normalized;
    }

    function toTranslationsRecordObject(value) {
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

    function getSectionEditorConfiguredLanguages(settingsValue) {
        const i18nSettings = isPlainObject(settingsValue?.i18n) ? settingsValue.i18n : {};
        if (i18nSettings?.enabled !== true) {
            return [];
        }

        const sourceLanguages = Array.isArray(i18nSettings?.languages) ? i18nSettings.languages : [];
        const seen = new Set();
        const languages = [];

        for (const entry of sourceLanguages) {
            const code = normalizeLanguageCode(
                isPlainObject(entry)
                    ? entry?.code
                    : entry,
            );
            if (!code || seen.has(code)) {
                continue;
            }

            seen.add(code);
            const labelSource = isPlainObject(entry) ? entry?.label : "";
            const label = normalizeString(labelSource) || code.toUpperCase();
            languages.push({ code, label });
        }

        return languages;
    }

    function findTranslationLanguageKey(translationsValue, languageCode) {
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedLanguage || !isPlainObject(translationsValue)) {
            return "";
        }

        for (const key of Object.keys(translationsValue)) {
            if (normalizeLanguageCode(key) === normalizedLanguage) {
                return key;
            }
        }

        return "";
    }

    function getTranslationPropsByLanguage(translationsValue, languageCode) {
        const key = findTranslationLanguageKey(translationsValue, languageCode);
        if (!key) {
            return {};
        }

        return toPropsObject(translationsValue?.[key]);
    }

    function toTranslationsDraftByLanguage(translationsValue) {
        const source = toTranslationsRecordObject(translationsValue);
        const nextValue = {};

        for (const [rawLanguageCode, rawProps] of Object.entries(source)) {
            const languageCode = normalizeLanguageCode(rawLanguageCode);
            if (!languageCode || !isPlainObject(rawProps)) {
                continue;
            }

            nextValue[languageCode] = toPropsObject(rawProps);
        }

        return nextValue;
    }

    function toPageSeoTranslationDraft(value) {
        const source = toPropsObject(value);
        return {
            title: normalizeString(source?.title),
            description: `${source?.description || ""}`,
        };
    }

    function toPageSeoTranslationsDraftByLanguage(value) {
        const source = toTranslationsRecordObject(value);
        const nextValue = {};

        for (const [rawLanguageCode, rawTranslationValue] of Object.entries(source)) {
            const languageCode = normalizeLanguageCode(rawLanguageCode);
            if (!languageCode || !isPlainObject(rawTranslationValue)) {
                continue;
            }

            const nextDraft = toPageSeoTranslationDraft(rawTranslationValue);
            if (!isPageSeoTranslationDraftEmpty(nextDraft)) {
                nextValue[languageCode] = nextDraft;
            }
        }

        return nextValue;
    }

    function isPageSeoTranslationDraftEmpty(value) {
        const draft = toPageSeoTranslationDraft(value);
        return !normalizeString(draft?.title) && !normalizeString(draft?.description);
    }

    function getPageSeoTranslationDraft(languageCode) {
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedLanguage) {
            return { title: "", description: "" };
        }

        return toPageSeoTranslationDraft(pageSeoTranslationsDraftByLanguage?.[normalizedLanguage]);
    }

    function updatePageSeoTranslationDraft(languageCode, nextValue) {
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedLanguage) {
            return;
        }

        const nextDraft = toPageSeoTranslationDraft(nextValue);
        const nextTranslationsDraftByLanguage = {
            ...(pageSeoTranslationsDraftByLanguage || {}),
        };

        if (isPageSeoTranslationDraftEmpty(nextDraft)) {
            delete nextTranslationsDraftByLanguage[normalizedLanguage];
        } else {
            nextTranslationsDraftByLanguage[normalizedLanguage] = nextDraft;
        }

        pageSeoTranslationsDraftByLanguage = nextTranslationsDraftByLanguage;
    }

    function isPropsObjectEmpty(value) {
        return stableSerializeForDirtyCheck(toPropsObject(value)) === "{}";
    }

    function getSectionTranslationDraftProps(blockId, languageCode) {
        const normalizedBlockId = normalizeString(blockId);
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedBlockId || !normalizedLanguage) {
            return {};
        }

        return toPropsObject(sectionTranslationsDraftById?.[normalizedBlockId]?.[normalizedLanguage]);
    }

    function resetSectionTranslationDraft(block, languageCode) {
        const blockId = normalizeString(block?.id);
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!blockId || !normalizedLanguage) {
            return;
        }

        const persistedTranslations = toTranslationsRecordObject(
            effectiveBlockTranslationsField ? block?.[effectiveBlockTranslationsField] : {},
        );
        const nextLanguageValue = getTranslationPropsByLanguage(persistedTranslations, normalizedLanguage);
        const nextBlockDraft = {
            ...(sectionTranslationsDraftById?.[blockId] || {}),
        };

        if (isPropsObjectEmpty(nextLanguageValue)) {
            delete nextBlockDraft[normalizedLanguage];
        } else {
            nextBlockDraft[normalizedLanguage] = nextLanguageValue;
        }

        sectionTranslationsDraftById = {
            ...sectionTranslationsDraftById,
            [blockId]: nextBlockDraft,
        };
    }

    function updateSectionTranslationDraft(blockId, languageCode, nextValue) {
        const normalizedBlockId = normalizeString(blockId);
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedBlockId || !normalizedLanguage) {
            return;
        }

        const nextBlockDraft = {
            ...(sectionTranslationsDraftById?.[normalizedBlockId] || {}),
        };
        const nextProps = toPropsObject(nextValue);

        if (isPropsObjectEmpty(nextProps)) {
            delete nextBlockDraft[normalizedLanguage];
        } else {
            nextBlockDraft[normalizedLanguage] = nextProps;
        }

        sectionTranslationsDraftById = {
            ...sectionTranslationsDraftById,
            [normalizedBlockId]: nextBlockDraft,
        };
    }

    function removeLanguageTranslationKey(translationsObject, languageCode) {
        const normalizedLanguage = normalizeLanguageCode(languageCode);
        if (!normalizedLanguage || !isPlainObject(translationsObject)) {
            return;
        }

        for (const key of Object.keys(translationsObject)) {
            if (normalizeLanguageCode(key) === normalizedLanguage) {
                delete translationsObject[key];
            }
        }
    }

    function isPlainObject(value) {
        return !!value && typeof value === "object" && !Array.isArray(value);
    }

    function hasOwnObjectKey(value, key) {
        const normalizedKey = normalizeString(key);
        return !!normalizedKey && !!value && typeof value === "object" && Object.prototype.hasOwnProperty.call(value, normalizedKey);
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
        return CommonHelper.websiteDisplayLabel(record, {
            preferredFields: [websiteNameField],
            slugField: websiteSlugField || "slug",
            missingValue: record?.id || "-",
        });
    }

    function getWebsiteNameText(record) {
        return normalizeString(websiteNameField ? record?.[websiteNameField] : "");
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

    function getPageTitleText(record) {
        return normalizeString(pageTitleField ? record?.[pageTitleField] : "");
    }

    function getPageSeoTitleFallbackSource(pageRecord, websiteRecord, globalTitleText) {
        const pageTitle = getPageTitleText(pageRecord);
        if (pageTitle) {
            return pageTitle;
        }

        const globalTitle = normalizeString(globalTitleText);
        if (globalTitle) {
            return globalTitle;
        }

        const websiteName = getWebsiteNameText(websiteRecord);
        if (websiteName) {
            return websiteName;
        }

        return normalizeString(websiteSlugField ? websiteRecord?.[websiteSlugField] : "");
    }

    function getPageSeoCoverageStatus(record) {
        const pageId = normalizeString(record?.id);
        const isEditingPage = !!pageId && pageId === normalizeString(selectedPageId);
        const seoTitleRaw = isEditingPage ? pageEditForm?.seoTitle : (pageSeoTitleField ? record?.[pageSeoTitleField] : "");
        const seoDescriptionRaw = isEditingPage ? pageEditForm?.seoDescription : (pageSeoDescriptionField ? record?.[pageSeoDescriptionField] : "");
        const seoFocusKeywordRaw = isEditingPage ? pageEditForm?.seoFocusKeyword : (pageSeoFocusKeywordField ? record?.[pageSeoFocusKeywordField] : "");
        const seoCanonicalRaw = isEditingPage ? pageEditForm?.seoCanonicalUrl : (pageSeoCanonicalUrlField ? record?.[pageSeoCanonicalUrlField] : "");
        const seoNoindexRaw = isEditingPage ? pageEditForm?.seoNoindex : (pageSeoNoindexField ? record?.[pageSeoNoindexField] : false);
        const seoExcludeFromSitemapRaw = isEditingPage
            ? pageEditForm?.seoExcludeFromSitemap
            : (pageSeoExcludeFromSitemapField ? record?.[pageSeoExcludeFromSitemapField] : false);
        const hasPageSocialImage = pageSeoSocialImageField
            ? isEditingPage
                ? (!!normalizeString(pageEditForm?.seoSocialImageCurrent) || !!pageEditForm?.seoSocialImageFile)
                : !!toSingleFileName(record?.[pageSeoSocialImageField])
            : false;
        const hasGlobalSocialImage = !!normalizeString(websiteIdentitySeoDraft?.seoImageCurrent) || !!websiteIdentitySeoDraft?.seoImageFile;

        const titleText = normalizeString(seoTitleRaw);
        const descriptionText = toSeoPlainText(seoDescriptionRaw);
        const focusKeyword = normalizeString(seoFocusKeywordRaw);
        const canonicalUrl = normalizeString(seoCanonicalRaw);
        const hasTitleFallback = !!getPageSeoTitleFallbackSource(record, selectedWebsite, globalSeoTitleText);
        const hasDescriptionFallback = !!globalSeoDescriptionText;

        const checks = buildPageSeoChecks({
            titleText,
            descriptionText,
            titleLength: titleText.length,
            descriptionLength: descriptionText.length,
            hasTitle: !!titleText,
            hasDescription: !!descriptionText,
            hasTitleFallback,
            hasDescriptionFallback,
            hasSocialImage: hasPageSocialImage,
            hasGlobalSocialImage,
            focusKeyword,
            noindex: pageSeoNoindexField ? toBooleanValue(seoNoindexRaw) : false,
            excludeFromSitemap: pageSeoExcludeFromSitemapField ? toBooleanValue(seoExcludeFromSitemapRaw) : false,
            canonicalUrl,
            hasCanonicalField: !!pageSeoCanonicalUrlField,
            hasFaqStructuredData: false,
        });
        const checkCounts = getSeoCheckCounts(checks);

        return getPageSeoHealthStatus({
            hasTitle: !!titleText,
            hasDescription: !!descriptionText,
            hasTitleFallback,
            hasDescriptionFallback,
            warningCount: checkCounts.warnings,
        });
    }

    function getPageSeoCoverageCounts(records = [], statusById = new Map()) {
        const counts = {
            total: Array.isArray(records) ? records.length : 0,
            good: 0,
            needsAttention: 0,
            missingBasics: 0,
        };

        for (const record of records || []) {
            const key = statusById?.get(record?.id || "")?.key || pageSeoFilterNeedsAttentionKey;
            if (key === pageSeoFilterGoodKey) {
                counts.good += 1;
            } else if (key === pageSeoFilterMissingBasicsKey) {
                counts.missingBasics += 1;
            } else {
                counts.needsAttention += 1;
            }
        }

        return counts;
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

    function buildPagePreviewUrl(websiteSlug, pageSlug, languageCode = "") {
        const normalizedWebsiteSlug = normalizeString(websiteSlug);
        const normalizedPageSlug = normalizeString(pageSlug);
        const normalizedLanguageCode = normalizeLanguageCode(languageCode);

        if (!normalizedWebsiteSlug || !normalizedPageSlug) {
            return "";
        }

        const configuredBase = normalizeBaseUrl(getConfiguredPublicBaseUrl());
        const websiteBase = normalizeBaseUrl(getWebsitePublicUrl(selectedWebsite), { allowSingleLabelHost: false });
        const baseUrl = configuredBase || websiteBase;

        if (!baseUrl) {
            return "";
        }

        try {
            const previewUrl = new URL(
                `${baseUrl}/site/${encodeURIComponent(normalizedWebsiteSlug)}/${encodeURIComponent(normalizedPageSlug)}`,
            );

            if (normalizedLanguageCode) {
                previewUrl.searchParams.set("lang", normalizedLanguageCode);
            } else {
                previewUrl.searchParams.delete("lang");
            }

            return previewUrl.toString();
        } catch (_) {
            return "";
        }
    }

    function getPageSeoPreviewPath(previewUrl, websiteSlug, pageSlug) {
        const normalizedPreviewUrl = normalizeString(previewUrl);
        if (normalizedPreviewUrl) {
            try {
                return new URL(normalizedPreviewUrl).pathname || "/";
            } catch (_) {
                // no-op
            }
        }

        const normalizedWebsiteSlug = normalizeString(websiteSlug);
        const normalizedPageSlug = normalizeString(pageSlug);
        if (normalizedWebsiteSlug && normalizedPageSlug) {
            return `/site/${normalizedWebsiteSlug}/${normalizedPageSlug}`;
        }

        return "/site/{website}/{page}";
    }

    function getWebsiteSeoPreviewUrl(websiteUrl, websiteSlug) {
        const normalizedWebsiteUrl = normalizeString(websiteUrl);
        if (normalizedWebsiteUrl) {
            try {
                const parsed = new URL(normalizedWebsiteUrl);
                return `${parsed.host}${parsed.pathname === "/" ? "" : parsed.pathname}`;
            } catch (_) {
                return normalizedWebsiteUrl.replace(/^https?:\/\//i, "");
            }
        }

        const normalizedWebsiteSlug = normalizeString(websiteSlug);
        if (normalizedWebsiteSlug) {
            return `/site/${normalizedWebsiteSlug}`;
        }

        return "Website URL not available";
    }

    function buildGlobalSeoPreviewTitle({ template, separator, siteName, pageName = "Sample page" }) {
        const normalizedTemplate = normalizeString(template);
        const normalizedSiteName = normalizeString(siteName) || "Site name";
        const normalizedPageName = normalizeString(pageName) || "Sample page";

        if (normalizedTemplate) {
            const rendered = normalizedTemplate
                .replace(/\{page\}/gi, normalizedPageName)
                .replace(/\{site\}/gi, normalizedSiteName);
            return normalizeString(rendered) || normalizedPageName;
        }

        const normalizedSeparator = normalizeString(separator) || defaultSeoTitleSeparator;
        if (!normalizedSiteName) {
            return normalizedPageName;
        }

        return `${normalizedPageName} ${normalizedSeparator} ${normalizedSiteName}`;
    }

    function isLikelyCanonicalDomain(value) {
        const normalized = normalizeString(value);
        if (!normalized) {
            return false;
        }

        if (!/^https?:\/\//i.test(normalized)) {
            return false;
        }

        try {
            const parsed = new URL(normalized);
            return /^https?:$/i.test(parsed.protocol) && !!normalizeString(parsed.hostname);
        } catch (_) {
            return false;
        }
    }

    function getFirstFaqKeyText(source, keys = []) {
        if (!isPlainObject(source)) {
            return "";
        }

        for (const key of keys) {
            const rawValue = source?.[key];
            if (typeof rawValue === "string" || typeof rawValue === "number") {
                const text = toSeoPlainText(`${rawValue}`);
                if (text) {
                    return text;
                }
            }
        }

        return "";
    }

    function hasFaqQuestionAnswerPair(value, depth = 0) {
        if (depth > 5 || value == null) {
            return false;
        }

        if (Array.isArray(value)) {
            return value.some((entry) => hasFaqQuestionAnswerPair(entry, depth + 1));
        }

        if (!isPlainObject(value)) {
            return false;
        }

        const question = getFirstFaqKeyText(value, faqQuestionSummaryKeys);
        const answer = getFirstFaqKeyText(value, faqAnswerSummaryKeys);
        if (question && answer) {
            return true;
        }

        return Object.values(value).some((nestedValue) => hasFaqQuestionAnswerPair(nestedValue, depth + 1));
    }

    function hasFaqStructuredDataCandidate(blocksList = []) {
        if (!Array.isArray(blocksList) || !blocksList.length) {
            return false;
        }

        for (const [index, block] of blocksList.entries()) {
            const keyText = normalizeString(getBlockComponentKey(block)).toLowerCase();
            const titleText = normalizeString(getSectionTitle(block, index)).toLowerCase();
            const descriptionText = normalizeString(getSectionDescription(block, index)).toLowerCase();
            const looksLikeFaq = /faq|accordion/.test(`${keyText} ${titleText} ${descriptionText}`);

            if (!looksLikeFaq) {
                continue;
            }

            const propsValue = getSectionDraftProps(block);
            if (hasFaqQuestionAnswerPair(propsValue)) {
                return true;
            }
        }

        return false;
    }

    function buildPageSeoChecks({
        titleText,
        descriptionText,
        titleLength,
        descriptionLength,
        hasTitle,
        hasDescription,
        hasTitleFallback,
        hasDescriptionFallback,
        hasSocialImage,
        hasGlobalSocialImage,
        focusKeyword,
        noindex,
        excludeFromSitemap,
        canonicalUrl,
        hasCanonicalField,
        hasFaqStructuredData,
    }) {
        const checks = [];
        const addCheck = (level, message) => {
            checks.push({
                level,
                message,
            });
        };
        const normalizedCanonicalUrl = normalizeString(canonicalUrl);

        if (hasTitle) {
            addCheck("pass", "SEO title is set for this page.");
        } else if (hasTitleFallback) {
            addCheck("info", "SEO title is missing on this page. Runtime will use title fallback.");
        } else {
            addCheck("warning", "SEO title is missing and no fallback title was detected.");
        }

        if (hasDescription) {
            addCheck("pass", "SEO description is set for this page.");
        } else if (hasDescriptionFallback) {
            addCheck("info", "SEO description is missing on this page. Runtime will use global fallback description.");
        } else {
            addCheck("warning", "SEO description is missing and no fallback description was detected.");
        }

        if (hasTitle && titleLength > seoTitleLongThreshold) {
            addCheck("warning", "SEO title is long and may be truncated in search results.");
        } else if (hasTitle) {
            addCheck("pass", "SEO title length is within a healthy range.");
        }

        if (hasDescription) {
            if (descriptionLength < seoDescriptionShortThreshold) {
                addCheck("info", "SEO description is short. Consider adding more useful context.");
            } else if (descriptionLength > seoDescriptionLongThreshold) {
                addCheck("warning", "SEO description is long and may be truncated in search results.");
            } else {
                addCheck("pass", "SEO description length is within a healthy range.");
            }
        }

        if (hasSocialImage) {
            addCheck("pass", "Page social image is configured.");
        } else if (hasGlobalSocialImage) {
            addCheck("info", "Page social image is missing. Runtime will use the global SEO image fallback.");
        } else {
            addCheck("warning", "No page or global social image was detected for sharing previews.");
        }

        if (focusKeyword) {
            const titleHasKeyword = textContainsKeyword(titleText, focusKeyword);
            const descriptionHasKeyword = textContainsKeyword(descriptionText, focusKeyword);
            if (titleHasKeyword || descriptionHasKeyword) {
                addCheck("pass", "Focus keyword appears in SEO title or SEO description.");
            } else {
                addCheck("warning", "Focus keyword is not present in SEO title or SEO description.");
            }
        } else {
            addCheck("info", "Focus keyword is optional and currently not set.");
        }

        if (noindex) {
            addCheck("warning", "This page is marked as noindex and will not be indexed.");
        } else {
            addCheck("pass", "This page is indexable (index,follow).");
        }

        if (noindex && !excludeFromSitemap) {
            addCheck("warning", "Noindex pages are usually excluded from sitemap.");
        } else if (noindex && excludeFromSitemap) {
            addCheck("pass", "Noindex and sitemap exclusion settings are aligned.");
        }

        if (!hasCanonicalField) {
            addCheck("info", "Canonical URL field is not available on this page collection.");
        } else if (!normalizedCanonicalUrl) {
            addCheck("info", "Canonical URL is not set. Runtime will apply canonical fallback rules.");
        } else if (!/^https?:\/\//i.test(normalizedCanonicalUrl) || !isLikelyCanonicalDomain(normalizedCanonicalUrl)) {
            addCheck("warning", "Canonical URL should start with http:// or https:// and include a valid host.");
        } else {
            addCheck("pass", "Canonical URL is configured.");
        }

        addCheck("pass", "BreadcrumbList structured data will be generated for canonical site pages.");

        if (hasFaqStructuredData) {
            addCheck("pass", "FAQPage structured data will be generated from detected FAQ/accordion content.");
        } else {
            addCheck("info", "FAQPage structured data will be generated when valid FAQ/accordion content exists.");
        }

        addCheck("info", "LocalBusiness structured data comes from Website Settings, not page SEO fields.");

        return checks;
    }

    function buildGlobalSeoChecks({
        titleLength,
        descriptionLength,
        hasTitle,
        hasDescription,
        hasSocialImageField,
        hasSocialImage,
        hasTitleTemplateField,
        titleTemplate,
        hasTitleSeparatorField,
        titleSeparator,
        hasCanonicalDomainField,
        canonicalDomain,
    }) {
        const checks = [];
        const normalizedTitleTemplate = normalizeString(titleTemplate);
        const normalizedTitleSeparator = normalizeString(titleSeparator);
        const normalizedCanonicalDomain = normalizeString(canonicalDomain);

        if (!hasTitle) {
            checks.push({
                level: "warning",
                message: "Global SEO title is missing. Runtime fallback is weaker for pages without page SEO titles.",
            });
        } else if (titleLength > seoTitleLongThreshold) {
            checks.push({
                level: "warning",
                message: "Global SEO title is long and may be truncated in search results.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Global SEO title is configured.",
            });
        }

        if (!hasDescription) {
            checks.push({
                level: "warning",
                message: "Global SEO description is missing. Page fallback descriptions may be incomplete.",
            });
        } else if (descriptionLength > seoDescriptionLongThreshold) {
            checks.push({
                level: "warning",
                message: "Global SEO description is long and may be truncated in search results.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Global SEO description is configured.",
            });
        }

        if (!hasSocialImageField) {
            checks.push({
                level: "info",
                message: "Global SEO image field is not available on this website collection.",
            });
        } else if (hasSocialImage) {
            checks.push({
                level: "pass",
                message: "Global SEO image is configured for social sharing fallbacks.",
            });
        } else {
            checks.push({
                level: "warning",
                message: "Global SEO image is missing. Shared links may show inconsistent previews.",
            });
        }

        if (!hasTitleTemplateField) {
            checks.push({
                level: "info",
                message: "Title template field is not available on this website collection.",
            });
        } else if (normalizedTitleTemplate) {
            if (!/\{page\}/i.test(normalizedTitleTemplate)) {
                checks.push({
                    level: "warning",
                    message: "Title template should include {page} to represent each page title.",
                });
            } else {
                checks.push({
                    level: "pass",
                    message: "Title template includes {page}.",
                });
            }

            if (!/\{site\}/i.test(normalizedTitleTemplate)) {
                checks.push({
                    level: "info",
                    message: "Consider including {site} in the title template for clearer branding.",
                });
            } else {
                checks.push({
                    level: "pass",
                    message: "Title template includes {site}.",
                });
            }
        } else {
            checks.push({
                level: "info",
                message: "Title template is not set. Runtime fallback uses the title separator pattern.",
            });
        }

        if (!hasTitleSeparatorField) {
            checks.push({
                level: "info",
                message: "Title separator field is not available on this website collection.",
            });
        } else if (!normalizedTitleSeparator) {
            checks.push({
                level: "info",
                message: "Title separator is empty. Runtime fallback will use the default separator.",
            });
        } else if (normalizedTitleSeparator.length > seoSeparatorLongThreshold) {
            checks.push({
                level: "warning",
                message: "Title separator should stay short (usually 1 to 3 characters).",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Title separator length is readable.",
            });
        }

        if (!hasCanonicalDomainField) {
            checks.push({
                level: "info",
                message: "Canonical domain field is not available on this website collection.",
            });
        } else if (!normalizedCanonicalDomain) {
            checks.push({
                level: "warning",
                message: "Canonical domain is missing. Runtime fallback will use website/request host rules.",
            });
        } else if (!isLikelyCanonicalDomain(normalizedCanonicalDomain)) {
            checks.push({
                level: "warning",
                message: "Canonical domain should start with http:// or https:// and include a valid host.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Canonical domain is configured and looks valid.",
            });
        }

        return checks;
    }

    function buildWebsiteSeoAdvancedImpactChecks({
        hasTitleTemplateField,
        titleTemplate,
        hasTitleSeparatorField,
        titleSeparator,
        hasCanonicalDomainField,
        canonicalDomain,
    }) {
        const checks = [];
        const normalizedTitleTemplate = normalizeString(titleTemplate);
        const normalizedTitleSeparator = normalizeString(titleSeparator);
        const normalizedCanonicalDomain = normalizeString(canonicalDomain);

        if (!hasTitleTemplateField) {
            checks.push({
                level: "info",
                message: "Title template field is not available on this website collection.",
            });
        } else if (!normalizedTitleTemplate) {
            checks.push({
                level: "info",
                message: "Title template is not set. Runtime fallback uses the title separator pattern.",
            });
        } else {
            if (!/\{page\}/i.test(normalizedTitleTemplate)) {
                checks.push({
                    level: "warning",
                    message: "Title template should include {page} to represent each page title.",
                });
            }

            if (!/\{site\}/i.test(normalizedTitleTemplate)) {
                checks.push({
                    level: "info",
                    message: "Consider including {site} in the title template for clearer branding.",
                });
            }
        }

        if (!hasTitleSeparatorField) {
            checks.push({
                level: "info",
                message: "Title separator field is not available on this website collection.",
            });
        } else if (!normalizedTitleSeparator) {
            checks.push({
                level: "info",
                message: "Title separator is empty. Runtime fallback will use the default separator.",
            });
        } else if (normalizedTitleSeparator.length > seoSeparatorLongThreshold) {
            checks.push({
                level: "warning",
                message: "Title separator should stay short (usually 1 to 3 characters).",
            });
        }

        if (!hasCanonicalDomainField) {
            checks.push({
                level: "info",
                message: "Canonical domain field is not available on this website collection.",
            });
        } else if (!normalizedCanonicalDomain) {
            checks.push({
                level: "info",
                message: "Canonical domain is not set. Runtime fallback will use website/request host rules.",
            });
        } else if (!isLikelyCanonicalDomain(normalizedCanonicalDomain)) {
            checks.push({
                level: "warning",
                message: "Canonical domain should start with http:// or https:// and include a valid host.",
            });
        }

        return checks;
    }

    function buildLocalBusinessSeoChecks({
        hasBusinessNameField,
        businessName,
        hasPrimaryCategoryField,
        primaryCategory,
        hasPhoneField,
        phone,
        hasEmailField,
        email,
        hasAddressField,
        address,
        hasCityField,
        city,
        hasCountryField,
        country,
        hasServiceAreaField,
        serviceArea,
        hasOpeningHoursField,
        openingHours,
        hasGooglePlaceIdField,
        googlePlaceId,
        hasSocialProfilesField,
        socialProfiles,
        hasPriceRangeField,
        priceRange,
        expectsGooglePlaceId,
    }) {
        const checks = [];
        const hasCoreAddressFields = hasAddressField && hasCityField && hasCountryField;
        const hasCoreAddress = hasCoreAddressFields && !!address && !!city && !!country;
        const hasServiceArea = hasServiceAreaField && !!serviceArea;
        const hasLocationSignal = hasCoreAddress || hasServiceArea;

        if (!hasBusinessNameField) {
            checks.push({
                level: "info",
                message: "Business name field is not available on this website collection.",
            });
        } else if (!businessName) {
            checks.push({
                level: "warning",
                message: "Business name is missing. Local search listings are stronger with a clear business identity.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Business name is configured.",
            });
        }

        if (!hasPrimaryCategoryField) {
            checks.push({
                level: "info",
                message: "Primary category field is not available on this website collection.",
            });
        } else if (!primaryCategory) {
            checks.push({
                level: "warning",
                message: "Primary category is missing. Add a clear category such as Dental clinic or Gym.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Primary category is configured.",
            });
        }

        if (!hasPhoneField) {
            checks.push({
                level: "info",
                message: "Business phone field is not available on this website collection.",
            });
        } else if (!phone) {
            checks.push({
                level: "info",
                message: "Business phone is missing. Contact details help local SEO confidence.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Business phone is configured.",
            });
        }

        if (!hasEmailField) {
            checks.push({
                level: "info",
                message: "Business email field is not available on this website collection.",
            });
        } else if (!email) {
            checks.push({
                level: "info",
                message: "Business email is missing. Add one to strengthen trust/contact signals.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Business email is configured.",
            });
        }

        if (!hasCoreAddressFields && !hasServiceAreaField) {
            checks.push({
                level: "info",
                message: "Location fields are not available on this website collection.",
            });
        } else if (!hasLocationSignal) {
            checks.push({
                level: "warning",
                message: "Add address/city/country or a service area to provide local geography signals.",
            });
        } else if (!hasCoreAddress && hasServiceArea) {
            checks.push({
                level: "info",
                message: "Service area is configured. Add full address details for stronger local business context.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Location details are configured.",
            });
        }

        if (!hasOpeningHoursField) {
            checks.push({
                level: "info",
                message: "Opening hours field is not available on this website collection.",
            });
        } else if (!openingHours) {
            checks.push({
                level: "info",
                message: "Opening hours are missing. Search engines often use business hours in local listings.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Opening hours are configured.",
            });
        }

        if (hasGooglePlaceIdField && expectsGooglePlaceId && !googlePlaceId) {
            checks.push({
                level: "info",
                message: "Google Place ID is missing. Add it if you plan to use local integrations.",
            });
        } else if (hasGooglePlaceIdField && expectsGooglePlaceId && googlePlaceId) {
            checks.push({
                level: "pass",
                message: "Google Place ID is configured.",
            });
        } else if (!hasGooglePlaceIdField && expectsGooglePlaceId) {
            checks.push({
                level: "info",
                message: "Google Place ID field is not available on this website collection.",
            });
        }

        if (!hasSocialProfilesField) {
            checks.push({
                level: "info",
                message: "Social profiles field is not available on this website collection.",
            });
        } else if (!socialProfiles) {
            checks.push({
                level: "info",
                message: "Social profiles are missing. Add them for future sameAs structured data coverage.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Social profiles are configured.",
            });
        }

        if (!hasPriceRangeField) {
            checks.push({
                level: "info",
                message: "Price range field is not available on this website collection.",
            });
        } else if (!priceRange) {
            checks.push({
                level: "info",
                message: "Price range is optional but useful for richer local listing context.",
            });
        } else {
            checks.push({
                level: "pass",
                message: "Price range is configured.",
            });
        }

        return checks;
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

    function sanitizeSectionDraftPropsForForm(draftValue, schemaFieldKeys = new Set()) {
        const sanitized = toPropsObject(draftValue);
        const normalizedTranslationsField = normalizeString(effectiveBlockTranslationsField).toLowerCase();

        if (!normalizedTranslationsField) {
            return sanitized;
        }

        // Keep legitimate component fields, but strip leaked block-level translations metadata.
        if (schemaFieldKeys?.has?.(normalizedTranslationsField)) {
            return sanitized;
        }

        if (hasOwnObjectKey(sanitized, normalizedTranslationsField)) {
            delete sanitized[normalizedTranslationsField];
        }

        return sanitized;
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

    function getSeoCheckCounts(checks = []) {
        let warnings = 0;
        let infos = 0;
        let passes = 0;

        for (const check of checks || []) {
            const level = `${check?.level || ""}`;
            if (level === "warning") {
                warnings += 1;
            } else if (level === "pass") {
                passes += 1;
            } else {
                infos += 1;
            }
        }

        return {
            warnings,
            infos,
            passes,
            total: warnings + infos + passes,
        };
    }

    function getSeoCheckSummaryText(counts) {
        const warnings = Number(counts?.warnings || 0);
        const infos = Number(counts?.infos || 0);

        if (!warnings && !infos) {
            return "No issues";
        }

        const parts = [];
        if (warnings) {
            parts.push(formatCount(warnings, "warning"));
        }
        if (infos) {
            parts.push(formatCount(infos, "info"));
        }

        return parts.join(" · ");
    }

    function getSeoHealthCompactSummary(counts) {
        const warnings = Number(counts?.warnings || 0);
        const suggestions = Number(counts?.infos || 0);

        if (!warnings && !suggestions) {
            return "No warnings";
        }

        const parts = [];
        if (warnings) {
            parts.push(formatCount(warnings, "warning"));
        }
        if (suggestions) {
            parts.push(formatCount(suggestions, "suggestion"));
        }

        return parts.join(" · ");
    }

    function getPageSeoHealthStatus({
        hasTitle,
        hasDescription,
        hasTitleFallback,
        hasDescriptionFallback,
        warningCount,
    } = {}) {
        const missingWithoutFallback = (!hasTitle && !hasTitleFallback) || (!hasDescription && !hasDescriptionFallback);
        if (missingWithoutFallback) {
            return {
                key: "missing-basics",
                label: "Missing basics",
            };
        }

        const usingFallbackBasics = (!hasTitle && hasTitleFallback) || (!hasDescription && hasDescriptionFallback);
        if (usingFallbackBasics) {
            return {
                key: "needs-attention",
                label: "Needs attention",
            };
        }

        if (Number(warningCount || 0) > 0) {
            return {
                key: "needs-attention",
                label: "Needs attention",
            };
        }

        return {
            key: "good",
            label: "Good",
        };
    }

    function getGlobalSeoHealthStatus({
        hasTitle,
        hasDescription,
        warningCount,
    } = {}) {
        if (!hasTitle || !hasDescription) {
            return {
                key: "missing-basics",
                label: "Missing basics",
            };
        }

        if (Number(warningCount || 0) > 0) {
            return {
                key: "needs-attention",
                label: "Needs attention",
            };
        }

        return {
            key: "good",
            label: "Good",
        };
    }

    function getLocalBusinessSeoHealthStatus({
        hasBusinessNameField,
        hasBusinessName,
        hasContactInputFields,
        hasContactSignal,
        hasLocationInputFields,
        hasLocationSignal,
        warningCount,
        infoCount,
    } = {}) {
        const requiresBusinessName = !!hasBusinessNameField;
        const hasAnyContactOrLocationInputs = !!hasContactInputFields || !!hasLocationInputFields;
        const missingName = requiresBusinessName && !hasBusinessName;
        const missingAllSignals = hasAnyContactOrLocationInputs && !hasContactSignal && !hasLocationSignal;

        if (missingName || missingAllSignals) {
            return {
                key: "missing-basics",
                label: "Missing basics",
            };
        }

        if (Number(warningCount || 0) > 0 || Number(infoCount || 0) > 0) {
            return {
                key: "needs-attention",
                label: "Needs attention",
            };
        }

        return {
            key: "good",
            label: "Good",
        };
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
        activeSectionLanguageKey = sectionDefaultLanguageKey;
        pendingSectionLanguageKey = "";
        sectionDiscardIntent = sectionDiscardIntentCloseKey;
    }

    function closeSectionEditor() {
        editingSectionId = "";
        isSectionDiscardConfirmOpen = false;
        bypassSectionCloseConfirm = false;
        sectionDiscardIntent = sectionDiscardIntentCloseKey;
        pendingSectionLanguageKey = "";
        activeSectionLanguageKey = sectionDefaultLanguageKey;
    }

    function closeSectionDiscardConfirm() {
        isSectionDiscardConfirmOpen = false;
        sectionDiscardIntent = sectionDiscardIntentCloseKey;
        pendingSectionLanguageKey = "";
    }

    function discardSectionEditorChanges() {
        const blockId = normalizeString(selectedEditingSection?.id);
        if (blockId && activeSectionTranslationLanguageCode) {
            resetSectionTranslationDraft(selectedEditingSection, activeSectionTranslationLanguageCode);
        } else if (blockId) {
            sectionPropsDraftById = {
                ...sectionPropsDraftById,
                [blockId]: toPropsObject(blockPropsField ? selectedEditingSection?.[blockPropsField] : {}),
            };
        }

        if (
            sectionDiscardIntent === sectionDiscardIntentSwitchLanguageKey
            && pendingSectionLanguageKey
        ) {
            const nextLanguage = pendingSectionLanguageKey;
            isSectionDiscardConfirmOpen = false;
            sectionDiscardIntent = sectionDiscardIntentCloseKey;
            pendingSectionLanguageKey = "";
            activeSectionLanguageKey = nextLanguage;
            return;
        }

        isSectionDiscardConfirmOpen = false;
        sectionDiscardIntent = sectionDiscardIntentCloseKey;
        pendingSectionLanguageKey = "";
        bypassSectionCloseConfirm = true;
        sectionEditorPanel?.hide();
    }

    function shouldCloseSectionEditor() {
        if (bypassSectionCloseConfirm) {
            bypassSectionCloseConfirm = false;
            return true;
        }

        if (!isSectionEditorDirty) {
            return true;
        }

        sectionDiscardIntent = sectionDiscardIntentCloseKey;
        pendingSectionLanguageKey = "";
        isSectionDiscardConfirmOpen = true;
        return false;
    }

    function setActiveSectionLanguage(nextLanguageKey) {
        const normalizedLanguageKey = normalizeString(nextLanguageKey) || sectionDefaultLanguageKey;
        const normalizedTarget = normalizedLanguageKey === sectionDefaultLanguageKey
            ? sectionDefaultLanguageKey
            : normalizeLanguageCode(normalizedLanguageKey);

        if (!normalizedTarget) {
            return;
        }

        if (normalizedTarget === activeSectionLanguageKey) {
            return;
        }

        if (normalizedTarget !== sectionDefaultLanguageKey) {
            const isVisibleLanguage = sectionEditorSupportsTranslations
                && sectionEditorTranslationLanguages.some((language) => language.code === normalizedTarget);
            if (!isVisibleLanguage) {
                return;
            }
        }

        if (isSectionEditorDirty) {
            pendingSectionLanguageKey = normalizedTarget;
            sectionDiscardIntent = sectionDiscardIntentSwitchLanguageKey;
            isSectionDiscardConfirmOpen = true;
            return;
        }

        activeSectionLanguageKey = normalizedTarget;
    }

    function normalizeComparableForDirtyCheck(value) {
        if (Array.isArray(value)) {
            return value.map((item) => normalizeComparableForDirtyCheck(item));
        }

        if (isPlainObject(value)) {
            const normalizedObject = {};
            for (const key of Object.keys(value).sort()) {
                const nextValue = value[key];
                if (typeof nextValue === "undefined") {
                    continue;
                }

                normalizedObject[key] = normalizeComparableForDirtyCheck(nextValue);
            }
            return normalizedObject;
        }

        return value;
    }

    function stableSerializeForDirtyCheck(value) {
        try {
            return JSON.stringify(normalizeComparableForDirtyCheck(value));
        } catch (_) {
            return JSON.stringify({});
        }
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
        if (activeSectionTranslationLanguageCode) {
            updateSectionTranslationDraft(blockId, activeSectionTranslationLanguageCode, nextValue);
            return;
        }

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

    function getCollectionFileUrl(record, fileName) {
        const normalizedFile = toSingleFileName(fileName);
        if (!record || !normalizedFile) {
            return "";
        }

        try {
            return ApiClient.files.getURL?.(record, normalizedFile) || "";
        } catch (_) {
            return "";
        }
    }

    function initializeWebsiteSettingsDraft() {
        websiteSettingsError = "";

        if (!selectedWebsite || !hasWebsiteSettingsField) {
            websiteSettingsFullDraft = {};
            websiteSettingsDraft = {};
            return;
        }

        const normalizedFullSettings = normalizeWebsiteSettingsValue(selectedWebsite?.[resolvedWebsiteSettingsField]);
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
                seoTitleTemplate: "",
                seoTitleSeparator: "",
                seoCanonicalDomain: "",
                businessName: "",
                businessType: "",
                businessPrimaryCategory: "",
                businessPhone: "",
                businessEmail: "",
                businessAddress: "",
                businessCity: "",
                businessPostalCode: "",
                businessCountry: "",
                businessServiceArea: "",
                businessOpeningHours: "",
                businessGooglePlaceId: "",
                businessSocialProfiles: "",
                businessPriceRange: "",
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
            seoTitleTemplate: websiteSeoTitleTemplateField ? `${selectedWebsite?.[websiteSeoTitleTemplateField] || ""}` : "",
            seoTitleSeparator: websiteSeoTitleSeparatorField ? `${selectedWebsite?.[websiteSeoTitleSeparatorField] || ""}` : "",
            seoCanonicalDomain: websiteSeoCanonicalDomainField ? `${selectedWebsite?.[websiteSeoCanonicalDomainField] || ""}` : "",
            businessName: websiteBusinessNameField ? `${selectedWebsite?.[websiteBusinessNameField] || ""}` : "",
            businessType: websiteBusinessTypeField ? `${selectedWebsite?.[websiteBusinessTypeField] || ""}` : "",
            businessPrimaryCategory: websiteBusinessPrimaryCategoryField ? `${selectedWebsite?.[websiteBusinessPrimaryCategoryField] || ""}` : "",
            businessPhone: websiteBusinessPhoneField ? `${selectedWebsite?.[websiteBusinessPhoneField] || ""}` : "",
            businessEmail: websiteBusinessEmailField ? `${selectedWebsite?.[websiteBusinessEmailField] || ""}` : "",
            businessAddress: websiteBusinessAddressField ? `${selectedWebsite?.[websiteBusinessAddressField] || ""}` : "",
            businessCity: websiteBusinessCityField ? `${selectedWebsite?.[websiteBusinessCityField] || ""}` : "",
            businessPostalCode: websiteBusinessPostalCodeField ? `${selectedWebsite?.[websiteBusinessPostalCodeField] || ""}` : "",
            businessCountry: websiteBusinessCountryField ? `${selectedWebsite?.[websiteBusinessCountryField] || ""}` : "",
            businessServiceArea: websiteBusinessServiceAreaField ? `${selectedWebsite?.[websiteBusinessServiceAreaField] || ""}` : "",
            businessOpeningHours: websiteBusinessOpeningHoursField ? `${selectedWebsite?.[websiteBusinessOpeningHoursField] || ""}` : "",
            businessGooglePlaceId: websiteBusinessGooglePlaceIdField ? `${selectedWebsite?.[websiteBusinessGooglePlaceIdField] || ""}` : "",
            businessSocialProfiles: websiteBusinessSocialProfilesField ? `${selectedWebsite?.[websiteBusinessSocialProfilesField] || ""}` : "",
            businessPriceRange: websiteBusinessPriceRangeField ? `${selectedWebsite?.[websiteBusinessPriceRangeField] || ""}` : "",
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

    function handlePageSeoFileChange(event) {
        const file = event.currentTarget?.files?.[0] || null;

        pageEditForm = {
            ...pageEditForm,
            seoSocialImageFile: file,
        };

        if (event.currentTarget) {
            event.currentTarget.value = "";
        }
    }

    function handlePageSeoTitleInput(event) {
        const nextValue = `${event.currentTarget?.value || ""}`;

        if (activePageSeoTranslationLanguageCode) {
            updatePageSeoTranslationDraft(activePageSeoTranslationLanguageCode, {
                ...getPageSeoTranslationDraft(activePageSeoTranslationLanguageCode),
                title: nextValue,
            });
            return;
        }

        pageEditForm = {
            ...pageEditForm,
            seoTitle: nextValue,
        };
    }

    function handlePageSeoDescriptionInput(event) {
        const nextValue = `${event.currentTarget?.value || ""}`;

        if (activePageSeoTranslationLanguageCode) {
            updatePageSeoTranslationDraft(activePageSeoTranslationLanguageCode, {
                ...getPageSeoTranslationDraft(activePageSeoTranslationLanguageCode),
                description: nextValue,
            });
            return;
        }

        pageEditForm = {
            ...pageEditForm,
            seoDescription: nextValue,
        };
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
        pageSeoFilter = pageSeoFilterAllKey;
        activePageSeoTab = pageSeoTabBasicKey;
        activeWebsiteIdentitySeoTab = websiteIdentitySeoTabBasicKey;
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
        activePageSeoTab = pageSeoTabBasicKey;

        await loadBlocks();
    }

    function setActivePageEditorTab(nextTab) {
        if (nextTab === pageEditorTabContentKey || nextTab === pageEditorTabSeoKey) {
            activePageEditorTab = nextTab;
        }
    }

    function setActivePageSeoLanguage(nextLanguageKey) {
        const normalizedLanguageKey = normalizeString(nextLanguageKey) || sectionDefaultLanguageKey;
        const normalizedTarget = normalizedLanguageKey === sectionDefaultLanguageKey
            ? sectionDefaultLanguageKey
            : normalizeLanguageCode(normalizedLanguageKey);

        if (!normalizedTarget || normalizedTarget === activePageSeoLanguageKey) {
            return;
        }

        if (normalizedTarget !== sectionDefaultLanguageKey) {
            const isVisibleLanguage = pageSeoSupportsTranslations
                && pageSeoTranslationLanguages.some((language) => language.code === normalizedTarget);
            if (!isVisibleLanguage) {
                return;
            }
        }

        activePageSeoLanguageKey = normalizedTarget;
        if (normalizedTarget !== sectionDefaultLanguageKey && activePageSeoTab === pageSeoTabAdvancedKey) {
            activePageSeoTab = pageSeoTabBasicKey;
        }
    }

    function setActiveWebsiteSettingsArea(nextArea) {
        if (nextArea === websiteSettingsAreaIdentitySeoKey) {
            activeWebsiteSettingsArea = nextArea;
            return;
        }

        if (nextArea === websiteSettingsAreaFeaturesKey) {
            activeWebsiteSettingsFeatureKey = getDefaultWebsiteSettingsFeatureKey();
            activeWebsiteSettingsArea = nextArea;
        }
    }

    function getDefaultWebsiteSettingsFeatureKey() {
        if (!availableWebsiteSettingsFeatures.length) {
            return "";
        }

        const leadsFeature = availableWebsiteSettingsFeatures.find(
            (feature) => feature.key === websiteSettingsLeadsFeatureKey,
        );

        if (leadsFeature?.key) {
            return leadsFeature.key;
        }

        return availableWebsiteSettingsFeatures[0]?.key || "";
    }

    function setActiveWebsiteIdentitySeoTab(nextTab) {
        if (
            nextTab === websiteIdentitySeoTabBasicKey ||
            nextTab === websiteIdentitySeoTabLocalBusinessKey ||
            nextTab === websiteIdentitySeoTabAdvancedKey
        ) {
            activeWebsiteIdentitySeoTab = nextTab;
        }
    }

    function setActiveWebsiteSettingsFeature(nextFeatureKey) {
        const normalizedKey = normalizeString(nextFeatureKey);
        if (!normalizedKey) {
            return;
        }

        if (!availableWebsiteSettingsFeatures.some((feature) => feature.key === normalizedKey)) {
            return;
        }

        activeWebsiteSettingsFeatureKey = normalizedKey;
    }

    function handleWebsiteSettingsFeatureGroupChange(featureKey, event) {
        const normalizedFeatureKey = normalizeString(featureKey);
        if (!normalizedFeatureKey) {
            return;
        }

        const nextGroupValue = event?.detail?.value?.[normalizedFeatureKey];
        const nextScopedValue = {
            ...websiteSettingsDraft,
            [normalizedFeatureKey]: typeof nextGroupValue === "undefined"
                ? structuredClone(websiteSettingsDraft?.[normalizedFeatureKey] ?? {})
                : structuredClone(nextGroupValue),
        };

        handleWebsiteSettingsChange({
            detail: { value: nextScopedValue },
        });
    }

    function buildWebsiteSettingsFeatureFormValue(featureKey) {
        const normalizedFeatureKey = normalizeString(featureKey);
        if (!normalizedFeatureKey) {
            return {};
        }

        return {
            [normalizedFeatureKey]: structuredClone(websiteSettingsDraft?.[normalizedFeatureKey] ?? {}),
        };
    }

    async function savePageSeo() {
        pageError = "";

        if (!selectedPage?.id || !pagesCollection?.id) {
            pageError = "Select a page first.";
            return;
        }

        const payload = {};
        if (activePageSeoTranslationLanguageCode && activePageSeoTab !== pageSeoTabAdvancedKey) {
            if (!effectivePageSeoTranslationsField) {
                pageError = "SEO translations field is not available for this page collection.";
                return;
            }

            const currentTranslations = toTranslationsRecordObject(
                effectivePageSeoTranslationsField
                    ? selectedPage?.[effectivePageSeoTranslationsField]
                    : (hasOwnObjectKey(selectedPage, "seo_translations") ? selectedPage?.seo_translations : {}),
            );
            const draftTranslation = getPageSeoTranslationDraft(activePageSeoTranslationLanguageCode);

            removeLanguageTranslationKey(currentTranslations, activePageSeoTranslationLanguageCode);
            if (!isPageSeoTranslationDraftEmpty(draftTranslation)) {
                currentTranslations[activePageSeoTranslationLanguageCode] = {
                    title: normalizeString(draftTranslation?.title),
                    description: `${draftTranslation?.description || ""}`,
                };
            }

            setPayloadField(payload, effectivePageSeoTranslationsField, currentTranslations);
        } else {
            if (!activePageSeoTranslationLanguageCode && pageSeoTitleField) {
                setPayloadField(payload, pageSeoTitleField, normalizeString(pageEditForm.seoTitle));
            }
            if (!activePageSeoTranslationLanguageCode && pageSeoDescriptionField) {
                setPayloadField(payload, pageSeoDescriptionField, `${pageEditForm.seoDescription || ""}`);
            }
            if (pageSeoCanonicalUrlField) {
                setPayloadField(payload, pageSeoCanonicalUrlField, normalizeString(pageEditForm.seoCanonicalUrl));
            }
            if (pageSeoNoindexField) {
                setPayloadField(payload, pageSeoNoindexField, toBooleanValue(pageEditForm.seoNoindex));
            }
            if (pageSeoExcludeFromSitemapField) {
                setPayloadField(
                    payload,
                    pageSeoExcludeFromSitemapField,
                    toBooleanValue(pageEditForm.seoExcludeFromSitemap),
                );
            }
            if (pageSeoFocusKeywordField) {
                setPayloadField(payload, pageSeoFocusKeywordField, normalizeString(pageEditForm.seoFocusKeyword));
            }
            if (pageSeoSocialImageField && pageEditForm.seoSocialImageFile) {
                setPayloadField(payload, pageSeoSocialImageField, pageEditForm.seoSocialImageFile);
            }
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

        if (websiteSeoTitleTemplateField) {
            setPayloadField(payload, websiteSeoTitleTemplateField, normalizeString(websiteIdentitySeoDraft.seoTitleTemplate));
        }

        if (websiteSeoTitleSeparatorField) {
            setPayloadField(payload, websiteSeoTitleSeparatorField, normalizeString(websiteIdentitySeoDraft.seoTitleSeparator));
        }

        if (websiteSeoCanonicalDomainField) {
            setPayloadField(payload, websiteSeoCanonicalDomainField, normalizeString(websiteIdentitySeoDraft.seoCanonicalDomain));
        }

        if (websiteBusinessNameField) {
            setPayloadField(payload, websiteBusinessNameField, normalizeString(websiteIdentitySeoDraft.businessName));
        }

        if (websiteBusinessTypeField) {
            setPayloadField(payload, websiteBusinessTypeField, normalizeString(websiteIdentitySeoDraft.businessType));
        }

        if (websiteBusinessPrimaryCategoryField) {
            setPayloadField(payload, websiteBusinessPrimaryCategoryField, normalizeString(websiteIdentitySeoDraft.businessPrimaryCategory));
        }

        if (websiteBusinessPhoneField) {
            setPayloadField(payload, websiteBusinessPhoneField, normalizeString(websiteIdentitySeoDraft.businessPhone));
        }

        if (websiteBusinessEmailField) {
            setPayloadField(payload, websiteBusinessEmailField, normalizeString(websiteIdentitySeoDraft.businessEmail));
        }

        if (websiteBusinessAddressField) {
            setPayloadField(payload, websiteBusinessAddressField, normalizeString(websiteIdentitySeoDraft.businessAddress));
        }

        if (websiteBusinessCityField) {
            setPayloadField(payload, websiteBusinessCityField, normalizeString(websiteIdentitySeoDraft.businessCity));
        }

        if (websiteBusinessPostalCodeField) {
            setPayloadField(payload, websiteBusinessPostalCodeField, normalizeString(websiteIdentitySeoDraft.businessPostalCode));
        }

        if (websiteBusinessCountryField) {
            setPayloadField(payload, websiteBusinessCountryField, normalizeString(websiteIdentitySeoDraft.businessCountry));
        }

        if (websiteBusinessServiceAreaField) {
            setPayloadField(payload, websiteBusinessServiceAreaField, normalizeString(websiteIdentitySeoDraft.businessServiceArea));
        }

        if (websiteBusinessOpeningHoursField) {
            setPayloadField(payload, websiteBusinessOpeningHoursField, `${websiteIdentitySeoDraft.businessOpeningHours || ""}`);
        }

        if (websiteBusinessGooglePlaceIdField) {
            setPayloadField(payload, websiteBusinessGooglePlaceIdField, normalizeString(websiteIdentitySeoDraft.businessGooglePlaceId));
        }

        if (websiteBusinessSocialProfilesField) {
            setPayloadField(payload, websiteBusinessSocialProfilesField, `${websiteIdentitySeoDraft.businessSocialProfiles || ""}`);
        }

        if (websiteBusinessPriceRangeField) {
            setPayloadField(payload, websiteBusinessPriceRangeField, normalizeString(websiteIdentitySeoDraft.businessPriceRange));
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
            setPayloadField(payload, resolvedWebsiteSettingsField, structuredClone(websiteSettingsFullDraft));
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

        if (!blocksCollection?.id) {
            sectionErrorById = {
                ...sectionErrorById,
                [blockId]: "Section cannot be saved because blocks collection is missing.",
            };
            return;
        }

        const payload = {};
        if (activeSectionTranslationLanguageCode) {
            if (!effectiveBlockTranslationsField) {
                sectionErrorById = {
                    ...sectionErrorById,
                    [blockId]: "Section translations cannot be saved because translations field is missing.",
                };
                return;
            }

            const currentTranslations = toTranslationsRecordObject(
                effectiveBlockTranslationsField ? block?.[effectiveBlockTranslationsField] : {},
            );
            const draftTranslationValue = getSectionTranslationDraftProps(blockId, activeSectionTranslationLanguageCode);

            removeLanguageTranslationKey(currentTranslations, activeSectionTranslationLanguageCode);
            if (!isPropsObjectEmpty(draftTranslationValue)) {
                currentTranslations[activeSectionTranslationLanguageCode] = toPropsObject(draftTranslationValue);
            }

            setPayloadField(payload, effectiveBlockTranslationsField, currentTranslations);
        } else {
            if (!blockPropsField) {
                sectionErrorById = {
                    ...sectionErrorById,
                    [blockId]: "Section cannot be saved because props field is missing.",
                };
                return;
            }

            setPayloadField(payload, blockPropsField, toPropsObject(sectionPropsDraftById?.[blockId]));
        }

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
    {#if $isCollectionsLoading || (!$hasCollectionsLoaded && !$collectionsLoadError)}
        <div class="placeholder-section m-b-base">
            <span class="loader loader-lg" />
            <h1>Loading website content...</h1>
        </div>
    {:else if $collectionsLoadError}
        <div class="alert alert-danger m-b-base">
            <div class="icon">
                <i class="ri-error-warning-line" />
            </div>
            <div>
                Could not verify CMS collections.<br />
                Refresh the page or check your connection.
            </div>
        </div>
    {:else if $hasCollectionsLoaded && !hasCmsCollections}
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

                <div class="head-selector operations-website-select">
                    <div class="selector-row">
                        <label class="txt-sm txt-hint selector-label m-b-0" for="cms-website">Website</label>
                        <select
                            id="cms-website"
                            class="input input-sm"
                            value={selectedWebsiteId}
                            disabled={isLoadingWebsites || !websites.length}
                            on:change={handleWebsiteChange}
                        >
                            {#if !websites.length}
                                <option value="">No websites available</option>
                            {:else}
                                {#each websites as website}
                                    <option value={website.id}>{getWebsiteLabel(website)}</option>
                                {/each}
                            {/if}
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

        <section class="panel operations-content-panel cms-section-panel m-b-base" class:is-pages-workspace={activeCmsTab === cmsTabPagesKey}>
            {#if activeCmsTab === cmsTabPagesKey}
                <div class="content-workspace-grid">
                    <aside class="panel pages-list-panel">
                        <div class="pages-list-head">
                            <div class="pages-list-title-wrap">
                                <div class="pages-list-title-head">
                                    <h4 class="m-0">Pages</h4>
                                    <p class="txt-sm txt-hint m-b-0 pages-list-subtitle">Edit content & SEO.</p>
                                </div>
                                <p class="txt-sm txt-hint m-b-0 pages-list-totals">
                                    {#if pageEnabledField}
                                        {pages.length} pages · {activePagesCount} active · {inactivePagesCount} inactive
                                    {:else}
                                        {pages.length} pages
                                    {/if}
                                </p>
                            </div>
                        </div>

                        <div class="pages-search-row">
                            <input
                                id="cms-pages-search"
                                class="input input-sm"
                                type="search"
                                placeholder="Search by page title..."
                                bind:value={pageSearch}
                            />
                        </div>

                        <div class="pages-filter-toolbar">
                            <div class="pages-filter-group pages-filter-group-seo" role="toolbar" aria-label="Filter pages by SEO status">
                                <div class="page-filter-chips">
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline page-filter-chip page-seo-filter-chip"
                                        class:is-active={pageSeoFilter === pageSeoFilterAllKey}
                                        on:click={() => (pageSeoFilter = pageSeoFilterAllKey)}
                                    >
                                        All SEO ({pageSeoCoverageCounts.total})
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline page-filter-chip page-seo-filter-chip"
                                        class:is-active={pageSeoFilter === pageSeoFilterGoodKey}
                                        on:click={() => (pageSeoFilter = pageSeoFilterGoodKey)}
                                    >
                                        Good ({pageSeoCoverageCounts.good})
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline page-filter-chip page-seo-filter-chip"
                                        class:is-active={pageSeoFilter === pageSeoFilterNeedsAttentionKey}
                                        on:click={() => (pageSeoFilter = pageSeoFilterNeedsAttentionKey)}
                                    >
                                        Needs attention ({pageSeoCoverageCounts.needsAttention})
                                    </button>
                                    <button
                                        type="button"
                                        class="btn btn-xs btn-outline page-filter-chip page-seo-filter-chip"
                                        class:is-active={pageSeoFilter === pageSeoFilterMissingBasicsKey}
                                        on:click={() => (pageSeoFilter = pageSeoFilterMissingBasicsKey)}
                                    >
                                        Missing basics ({pageSeoCoverageCounts.missingBasics})
                                    </button>
                                </div>
                            </div>
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
                                    {@const pageSeoStatus = pageSeoStatusById.get(page.id) || getPageSeoCoverageStatus(page)}
                                    <button
                                        type="button"
                                        class="page-row"
                                        class:active={page.id === selectedPageId}
                                        on:click={() => selectPage(page.id)}
                                    >
                                        <div class="page-row-main">
                                            <span class="page-row-title">{getPageLabel(page)}</span>
                                            <span class="page-row-badges">
                                                {#if pageEnabledField}
                                                    <span class="label label-sm page-list-status" class:is-active={isPageActive(page)}>
                                                        {isPageActive(page) ? "Active" : "Inactive"}
                                                    </span>
                                                {/if}
                                                <span
                                                    class="label label-sm page-seo-status-pill"
                                                    class:good={pageSeoStatus.key === pageSeoFilterGoodKey}
                                                    class:needs-attention={pageSeoStatus.key === pageSeoFilterNeedsAttentionKey}
                                                    class:missing-basics={pageSeoStatus.key === pageSeoFilterMissingBasicsKey}
                                                >
                                                    {pageSeoStatus.label}
                                                </span>
                                            </span>
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
                                    <div class="seo-page-head">
                                        <h5 class="m-0">Google Search & Sharing</h5>
                                        <span class="txt-sm txt-hint seo-page-head-helper">
                                            Control how this page appears in Google and when shared.
                                        </span>
                                    </div>

                                    {#if pageSeoTitleField || pageSeoDescriptionField || pageSeoSocialImageField || pageSeoCanonicalUrlField || pageSeoNoindexField || pageSeoExcludeFromSitemapField || pageSeoFocusKeywordField}
                                        <div class="page-seo-tabs-row page-seo-tabs-row--compact">
                                            <div class="tabs-header compact combined left operations-tabs operations-tabs--nested page-seo-tabs">
                                                <button
                                                    type="button"
                                                    class="tab-item"
                                                    class:active={activePageSeoTab === pageSeoTabBasicKey}
                                                    on:click={() => (activePageSeoTab = pageSeoTabBasicKey)}
                                                >
                                                    <i class="ri-layout-left-line tab-icon" aria-hidden="true" />
                                                    <span class="tab-label">Basic</span>
                                                </button>
                                                {#if !activePageSeoTranslationLanguageCode}
                                                    <button
                                                        type="button"
                                                        class="tab-item"
                                                        class:active={activePageSeoTab === pageSeoTabAdvancedKey}
                                                        on:click={() => (activePageSeoTab = pageSeoTabAdvancedKey)}
                                                    >
                                                        <i class="ri-tools-line tab-icon" aria-hidden="true" />
                                                        <span class="tab-label">Advanced</span>
                                                    </button>
                                                {/if}
                                            </div>
                                        </div>

                                        {#if pageSeoSupportsTranslations && activePageSeoTab === pageSeoTabBasicKey}
                                            <div class="page-seo-tabs-row page-seo-tabs-row--compact">
                                                <div class="tabs-header compact combined left operations-tabs operations-tabs--nested section-language-tabs page-seo-language-tabs">
                                                    <button
                                                        type="button"
                                                        class="tab-item"
                                                        class:active={activePageSeoLanguageKey === sectionDefaultLanguageKey}
                                                        on:click={() => setActivePageSeoLanguage(sectionDefaultLanguageKey)}
                                                    >
                                                        <span>{pageSeoDefaultLanguageLabel}</span>
                                                    </button>
                                                    {#each pageSeoTranslationLanguages as language}
                                                        <button
                                                            type="button"
                                                            class="tab-item"
                                                            class:active={activePageSeoLanguageKey === language.code}
                                                            on:click={() => setActivePageSeoLanguage(language.code)}
                                                        >
                                                            <span>{language.label}</span>
                                                        </button>
                                                    {/each}
                                                </div>
                                            </div>
                                        {/if}

                                        {#if activePageSeoTab === pageSeoTabBasicKey}
                                            <div class="seo-editor-grid m-t-sm">
                                                <div class="seo-editor-main">
                                                    <div class="form-grid">
                                                        {#if pageSeoTitleField}
                                                            <div class="form-field seo-field">
                                                                <label for="cms-page-seo-title-content">
                                                                    Page title for Google
                                                                </label>
                                                                <input
                                                                    id="cms-page-seo-title-content"
                                                                    class="input form-input"
                                                                    value={pageSeoTitleInputValue}
                                                                    on:input={handlePageSeoTitleInput}
                                                                />
                                                                <div class="help-block m-t-6 seo-field-helper">
                                                                    <span class="label label-sm seo-count-pill">{pageSeoTitleLength} characters</span>
                                                                    {#if activePageSeoTranslationLanguageCode}
                                                                        <span>Shown in search results for this language. If empty, default SEO title is used.</span>
                                                                    {:else}
                                                                        <span>Shown in search results. If empty, the page title is used.</span>
                                                                    {/if}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if pageSeoDescriptionField}
                                                            <div class="form-field seo-field">
                                                                <label for="cms-page-seo-description-content">
                                                                    Page description for Google
                                                                </label>
                                                                <textarea
                                                                    id="cms-page-seo-description-content"
                                                                    class="input form-textarea textarea-input"
                                                                    rows="4"
                                                                    value={pageSeoDescriptionInputValue}
                                                                    on:input={handlePageSeoDescriptionInput}
                                                                />
                                                                <div class="help-block m-t-6 seo-field-helper">
                                                                    <span class="label label-sm seo-count-pill">{pageSeoDescriptionLength} characters</span>
                                                                    {#if activePageSeoTranslationLanguageCode}
                                                                        <span>Short summary shown in search results for this language.</span>
                                                                    {:else}
                                                                        <span>Short summary shown in search results.</span>
                                                                    {/if}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if !activePageSeoTranslationLanguageCode && pageSeoFocusKeywordField}
                                                            <div class="form-field seo-field">
                                                                <label for="cms-page-seo-focus-keyword">SEO focus keyword</label>
                                                                <input
                                                                    id="cms-page-seo-focus-keyword"
                                                                    class="input form-input"
                                                                    bind:value={pageEditForm.seoFocusKeyword}
                                                                />
                                                                <div class="help-block m-t-6">
                                                                    Used internally for SEO guidance. This is not rendered as meta keywords.
                                                                </div>
                                                            </div>
                                                        {/if}
                                                    </div>

                                                    {#if !activePageSeoTranslationLanguageCode && pageSeoSocialImageField}
                                                        <div class="form-field seo-field m-t-sm">
                                                            <label for="cms-page-seo-social-image-file">Image used when sharing</label>
                                                            <div class="page-seo-file-control">
                                                                {#if pageSeoSocialImagePreviewUrl}
                                                                    <a
                                                                        class="page-seo-file-thumb"
                                                                        href={pageSeoSocialImagePreviewUrl}
                                                                        target="_blank"
                                                                        rel="noreferrer noopener"
                                                                        title="Open current image"
                                                                    >
                                                                        <img src={pageSeoSocialImagePreviewUrl} alt="Current social image preview" loading="lazy" />
                                                                    </a>
                                                                {/if}

                                                                <div class="settings-file-row">
                                                                    <input
                                                                        id="cms-page-seo-social-image-file"
                                                                        class="input form-input file-input page-seo-file-input"
                                                                        type="file"
                                                                        on:change={handlePageSeoFileChange}
                                                                    />
                                                                </div>
                                                            </div>
                                                            <div class="help-block file-field-hint m-t-6">
                                                                {#if pageEditForm.seoSocialImageFile}
                                                                    <span class="label label-sm settings-file-state">New file: {pageEditForm.seoSocialImageFile.name}</span>
                                                                {:else if pageEditForm.seoSocialImageCurrent}
                                                                    <span class="label label-sm settings-file-state">Current file: {pageEditForm.seoSocialImageCurrent}</span>
                                                                {:else}
                                                                    <span class="label label-sm settings-file-state">No current file</span>
                                                                {/if}
                                                            </div>
                                                            <div class="help-block m-t-6">
                                                                Used when this page is shared. If empty, the global SEO image is used.
                                                            </div>
                                                        </div>
                                                    {:else if !activePageSeoTranslationLanguageCode}
                                                        <p class="txt-sm txt-hint m-b-0">Page social image field is not available for this page collection.</p>
                                                    {/if}
                                                </div>

                                                <aside class="seo-editor-side">
                                                    <div class="seo-preview-card seo-search-preview-card">
                                                        <div class="seo-preview-label">Search Preview</div>
                                                        <div class="seo-preview-title">{pageSeoPreviewTitle}</div>
                                                        <div class="seo-preview-hint">{pageSeoPreviewPath}</div>
                                                        <div class="seo-preview-description">
                                                            {pageSeoPreviewDescription}
                                                        </div>
                                                    </div>

                                                    <div class="seo-checklist-panel seo-health-panel m-t-sm">
                                                        <div class="seo-checklist-head">
                                                            <div class="seo-health-main">
                                                                <h6 class="m-0 seo-checklist-title">SEO health</h6>
                                                                <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                    Estimated from current draft values and runtime SEO rules.
                                                                </p>
                                                            </div>
                                                            <div class="seo-health-meta">
                                                                <span
                                                                    class="label label-sm seo-health-status-pill"
                                                                    class:good={pageSeoHealthStatus.key === "good"}
                                                                    class:needs-attention={pageSeoHealthStatus.key === "needs-attention"}
                                                                    class:missing-basics={pageSeoHealthStatus.key === "missing-basics"}
                                                                >
                                                                    {pageSeoHealthStatus.label}
                                                                </span>
                                                                <span
                                                                    class="summary-pill seo-check-summary-pill"
                                                                    class:warning={pageSeoCheckCounts.warnings > 0}
                                                                >
                                                                    {pageSeoHealthCompactSummary}
                                                                </span>
                                                            </div>
                                                        </div>

                                                        {#if pageSeoWarningChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Warnings</div>
                                                                <div class="seo-check-list">
                                                                    {#each pageSeoWarningChecks as check}
                                                                        <div class="seo-check-item warning">
                                                                            <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if pageSeoSuggestionChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Suggestions</div>
                                                                <div class="seo-check-list">
                                                                    {#each pageSeoSuggestionChecks as check}
                                                                        <div class="seo-check-item">
                                                                            <span class="label label-sm seo-check-pill">Info</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if !pageSeoWarningChecks.length && !pageSeoSuggestionChecks.length}
                                                            <p class="txt-sm txt-hint m-t-8 m-b-0">No SEO issues found in this section.</p>
                                                        {/if}
                                                    </div>
                                                </aside>
                                            </div>
                                        {:else if activePageSeoTab === pageSeoTabAdvancedKey && !activePageSeoTranslationLanguageCode}
                                            {#if pageSeoCanonicalUrlField || pageSeoNoindexField || pageSeoExcludeFromSitemapField}
                                                <div class="seo-editor-grid m-t-sm">
                                                    <div class="seo-advanced-main">
                                                        <div class="seo-advanced-pane">
                                                            {#if pageSeoCanonicalUrlField}
                                                                <div class="seo-advanced-section">
                                                                    <h6 class="m-0 seo-advanced-section-title">Canonical</h6>
                                                                    <div class="form-grid seo-advanced-grid m-t-8">
                                                                        <div class="form-field seo-field local-seo-full-width">
                                                                            <label for="cms-page-seo-canonical-url">Canonical URL</label>
                                                                            <input
                                                                                id="cms-page-seo-canonical-url"
                                                                                class="input form-input"
                                                                                type="url"
                                                                                placeholder="https://example.com/canonical-path"
                                                                                bind:value={pageEditForm.seoCanonicalUrl}
                                                                            />
                                                                            <div class="help-block m-t-6">
                                                                                If empty, runtime will apply canonical fallback rules.
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            {#if pageSeoNoindexField || pageSeoExcludeFromSitemapField}
                                                                <div class="seo-advanced-section">
                                                                    <h6 class="m-0 seo-advanced-section-title">Indexing</h6>
                                                                    <div class="form-grid seo-advanced-grid m-t-8">
                                                                        {#if pageSeoNoindexField}
                                                                            <div class="form-field form-field-toggle seo-toggle-field">
                                                                                <input
                                                                                    id="cms-page-seo-noindex"
                                                                                    type="checkbox"
                                                                                    bind:checked={pageEditForm.seoNoindex}
                                                                                />
                                                                                <label for="cms-page-seo-noindex">Hide this page from Google</label>
                                                                            </div>
                                                                        {/if}

                                                                        {#if pageSeoExcludeFromSitemapField}
                                                                            <div class="form-field form-field-toggle seo-toggle-field">
                                                                                <input
                                                                                    id="cms-page-seo-exclude-from-sitemap"
                                                                                    type="checkbox"
                                                                                    bind:checked={pageEditForm.seoExcludeFromSitemap}
                                                                                />
                                                                                <label for="cms-page-seo-exclude-from-sitemap">Remove from sitemap</label>
                                                                            </div>
                                                                        {/if}
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>
                                                    </div>

                                                    <aside class="seo-editor-side">
                                                        <div class="seo-checklist-panel seo-impact-panel">
                                                            <div class="seo-checklist-head">
                                                                <div class="seo-health-main">
                                                                    <h6 class="m-0 seo-checklist-title">Current impact</h6>
                                                                    <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                        Runtime behavior based on current canonical and indexing settings.
                                                                    </p>
                                                                </div>
                                                            </div>

                                                            <div class="seo-check-list m-t-8">
                                                                {#if pageSeoNoindexField}
                                                                    {#if pageSeoNoindexValue}
                                                                        <div class="seo-check-item warning">
                                                                            <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                            <span class="seo-check-message">This page is marked as noindex and will not be indexed.</span>
                                                                        </div>
                                                                    {:else}
                                                                        <div class="seo-check-item pass">
                                                                            <span class="label label-sm seo-check-pill pass">Pass</span>
                                                                            <span class="seo-check-message">This page is indexable (index,follow).</span>
                                                                        </div>
                                                                    {/if}
                                                                {/if}

                                                                {#if pageSeoNoindexField && pageSeoExcludeFromSitemapField && pageSeoNoindexValue && !pageSeoExcludeFromSitemapValue}
                                                                    <div class="seo-check-item warning">
                                                                        <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                        <span class="seo-check-message">Noindex pages are usually excluded from sitemap.</span>
                                                                    </div>
                                                                {/if}

                                                                {#if pageSeoCanonicalUrlField}
                                                                    {#if !pageSeoCanonicalUrlText}
                                                                        <div class="seo-check-item">
                                                                            <span class="label label-sm seo-check-pill">Info</span>
                                                                            <span class="seo-check-message">Canonical URL is not set. Runtime will apply canonical fallback rules.</span>
                                                                        </div>
                                                                    {:else if !/^https?:\/\//i.test(pageSeoCanonicalUrlText) || !isLikelyCanonicalDomain(pageSeoCanonicalUrlText)}
                                                                        <div class="seo-check-item warning">
                                                                            <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                            <span class="seo-check-message">Canonical URL should start with http:// or https:// and include a valid host.</span>
                                                                        </div>
                                                                    {:else}
                                                                        <div class="seo-check-item pass">
                                                                            <span class="label label-sm seo-check-pill pass">Pass</span>
                                                                            <span class="seo-check-message">Canonical URL is configured.</span>
                                                                        </div>
                                                                    {/if}
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    </aside>
                                                </div>
                                            {:else}
                                                <p class="txt-sm txt-hint m-t-8 m-b-0">Advanced SEO fields are not available for this page collection.</p>
                                            {/if}
                                        {/if}

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
                        <h4 class="m-0">Website Settings</h4>
                        <p class="txt-sm txt-hint m-b-0 settings-head-helper">Edit general settings for the selected website.</p>
                    </div>

                    {#if !selectedWebsiteId}
                        <p class="txt-hint m-b-0">Select a website to edit settings.</p>
                    {:else}
                        <div class="settings-nav-row settings-nav-row--compact">
                            <div class="tabs-header compact combined left operations-tabs settings-nav-tabs">
                                <button
                                    type="button"
                                    class="tab-item"
                                    class:active={activeWebsiteSettingsArea === websiteSettingsAreaIdentitySeoKey}
                                    on:click={() => setActiveWebsiteSettingsArea(websiteSettingsAreaIdentitySeoKey)}
                                >
                                    <i class={`${websiteSettingsAreaIconByKey[websiteSettingsAreaIdentitySeoKey]} tab-icon`} aria-hidden="true" />
                                    <span class="tab-label">Identity & SEO</span>
                                </button>
                                <button
                                    type="button"
                                    class="tab-item"
                                    class:active={activeWebsiteSettingsArea === websiteSettingsAreaFeaturesKey}
                                    on:click={() => setActiveWebsiteSettingsArea(websiteSettingsAreaFeaturesKey)}
                                >
                                    <i class={`${websiteSettingsAreaIconByKey[websiteSettingsAreaFeaturesKey]} tab-icon`} aria-hidden="true" />
                                    <span class="tab-label">Features</span>
                                </button>
                            </div>
                        </div>

                        {#if activeWebsiteSettingsArea === websiteSettingsAreaIdentitySeoKey}
                            <div class="settings-sections m-t-sm">
                                {#if hasWebsiteIdentitySeoFields}
                                    <div class="settings-pane settings-identity-pane">
                                        <div class="seo-page-head">
                                            <h5 class="m-0">Identity & SEO</h5>
                                            <p class="txt-sm txt-hint m-b-0 seo-page-head-helper">Manage global fallback metadata and local business SEO signals.</p>
                                        </div>

                                        <div class="page-seo-tabs-row page-seo-tabs-row--compact">
                                            <div class="tabs-header compact combined left operations-tabs operations-tabs--nested page-seo-tabs settings-identity-seo-tabs">
                                                <button
                                                    type="button"
                                                    class="tab-item"
                                                    class:active={activeWebsiteIdentitySeoTab === websiteIdentitySeoTabBasicKey}
                                                    on:click={() => setActiveWebsiteIdentitySeoTab(websiteIdentitySeoTabBasicKey)}
                                                >
                                                    <i class="ri-earth-line tab-icon" aria-hidden="true" />
                                                    <span class="tab-label">Basic</span>
                                                </button>
                                                <button
                                                    type="button"
                                                    class="tab-item"
                                                    class:active={activeWebsiteIdentitySeoTab === websiteIdentitySeoTabLocalBusinessKey}
                                                    on:click={() => setActiveWebsiteIdentitySeoTab(websiteIdentitySeoTabLocalBusinessKey)}
                                                >
                                                    <i class="ri-store-3-line tab-icon" aria-hidden="true" />
                                                    <span class="tab-label">Local Business</span>
                                                </button>
                                                <button
                                                    type="button"
                                                    class="tab-item"
                                                    class:active={activeWebsiteIdentitySeoTab === websiteIdentitySeoTabAdvancedKey}
                                                    on:click={() => setActiveWebsiteIdentitySeoTab(websiteIdentitySeoTabAdvancedKey)}
                                                >
                                                    <i class="ri-tools-line tab-icon" aria-hidden="true" />
                                                    <span class="tab-label">Advanced</span>
                                                </button>
                                            </div>
                                        </div>

                                        {#if activeWebsiteIdentitySeoTab === websiteIdentitySeoTabBasicKey}
                                            <div class="seo-editor-grid m-t-sm">
                                                <div class="seo-editor-main">
                                                    <div class="settings-identity-section">
                                                        <div class="settings-subhead">
                                                            <h6 class="m-0">Identity</h6>
                                                            <p class="txt-sm txt-hint m-b-0 settings-subhead-helper">Fallback defaults are used when a page does not define its own SEO.</p>
                                                        </div>

                                                        {#if websiteLogoField}
                                                            <div class="form-field seo-field m-t-8">
                                                                <label for="cms-website-logo-file">Logo</label>
                                                                <div class="page-seo-file-control">
                                                                    {#if websiteLogoPreviewUrl}
                                                                        <a
                                                                            class="page-seo-file-thumb"
                                                                            href={websiteLogoPreviewUrl}
                                                                            target="_blank"
                                                                            rel="noreferrer noopener"
                                                                            title="Open current logo"
                                                                        >
                                                                            <img src={websiteLogoPreviewUrl} alt="Current logo preview" loading="lazy" />
                                                                        </a>
                                                                    {/if}
                                                                    <div class="settings-file-row">
                                                                        <input
                                                                            id="cms-website-logo-file"
                                                                            class="input form-input file-input page-seo-file-input"
                                                                            type="file"
                                                                            on:change={(event) => handleWebsiteSeoFileChange("logo", event)}
                                                                        />
                                                                    </div>
                                                                </div>
                                                                <div class="help-block file-field-hint m-t-6">
                                                                    {#if websiteIdentitySeoDraft.logoFile}
                                                                        <span class="label label-sm settings-file-state">New file: {websiteIdentitySeoDraft.logoFile.name}</span>
                                                                    {:else if websiteIdentitySeoDraft.logoCurrent}
                                                                        <span class="label label-sm settings-file-state">Current file: {websiteIdentitySeoDraft.logoCurrent}</span>
                                                                    {:else}
                                                                        <span class="label label-sm settings-file-state">No current file</span>
                                                                    {/if}
                                                                </div>
                                                            </div>
                                                        {:else}
                                                            <p class="txt-sm txt-hint m-b-0 m-t-8">Logo field is not available for this website.</p>
                                                        {/if}
                                                    </div>

                                                    <div class="settings-identity-section m-t-sm">
                                                        <div class="settings-subhead">
                                                            <h6 class="m-0">Global SEO</h6>
                                                            <p class="txt-sm txt-hint m-b-0 settings-subhead-helper">Fallback metadata for pages without page-level SEO values.</p>
                                                        </div>

                                                        <div class="form-grid m-t-8">
                                                            {#if websiteSeoTitleField}
                                                                <div class="form-field seo-field">
                                                                    <label for="cms-website-seo-title">Global title for Google</label>
                                                                    <input
                                                                        id="cms-website-seo-title"
                                                                        class="input form-input"
                                                                        bind:value={websiteIdentitySeoDraft.seoTitle}
                                                                    />
                                                                    <div class="help-block m-t-6 seo-field-helper">
                                                                        <span class="label label-sm seo-count-pill">{globalSeoTitleLength} characters</span>
                                                                        <span>Used as fallback when page SEO title is empty.</span>
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            {#if websiteSeoDescriptionField}
                                                                <div class="form-field seo-field">
                                                                    <label for="cms-website-seo-description">Global description for Google</label>
                                                                    <textarea
                                                                        id="cms-website-seo-description"
                                                                        class="input form-textarea textarea-input"
                                                                        rows="4"
                                                                        bind:value={websiteIdentitySeoDraft.seoDescription}
                                                                    />
                                                                    <div class="help-block m-t-6 seo-field-helper">
                                                                        <span class="label label-sm seo-count-pill">{globalSeoDescriptionLength} characters</span>
                                                                        <span>Used as fallback when page SEO description is empty.</span>
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>

                                                        {#if websiteSeoImageField}
                                                            <div class="form-field seo-field m-t-sm">
                                                                <label for="cms-website-seo-image-file">Default image used when sharing</label>
                                                                <div class="page-seo-file-control">
                                                                    {#if websiteSeoImagePreviewUrl}
                                                                        <a
                                                                            class="page-seo-file-thumb"
                                                                            href={websiteSeoImagePreviewUrl}
                                                                            target="_blank"
                                                                            rel="noreferrer noopener"
                                                                            title="Open current SEO image"
                                                                        >
                                                                            <img src={websiteSeoImagePreviewUrl} alt="Current global SEO image preview" loading="lazy" />
                                                                        </a>
                                                                    {/if}
                                                                    <div class="settings-file-row">
                                                                        <input
                                                                            id="cms-website-seo-image-file"
                                                                            class="input form-input file-input page-seo-file-input"
                                                                            type="file"
                                                                            on:change={(event) => handleWebsiteSeoFileChange("seoImage", event)}
                                                                        />
                                                                    </div>
                                                                </div>
                                                                <div class="help-block file-field-hint m-t-6">
                                                                    {#if websiteIdentitySeoDraft.seoImageFile}
                                                                        <span class="label label-sm settings-file-state">New file: {websiteIdentitySeoDraft.seoImageFile.name}</span>
                                                                    {:else if websiteIdentitySeoDraft.seoImageCurrent}
                                                                        <span class="label label-sm settings-file-state">Current file: {websiteIdentitySeoDraft.seoImageCurrent}</span>
                                                                    {:else}
                                                                        <span class="label label-sm settings-file-state">No current file</span>
                                                                    {/if}
                                                                </div>
                                                                <div class="help-block m-t-6">
                                                                    Used when pages are shared. If empty, runtime fallback applies.
                                                                </div>
                                                            </div>
                                                        {:else}
                                                            <p class="txt-sm txt-hint m-b-0 m-t-sm">Global SEO image field is not available for this website.</p>
                                                        {/if}
                                                    </div>
                                                </div>

                                                <aside class="seo-editor-side">
                                                    <div class="seo-preview-card seo-search-preview-card">
                                                        <div class="seo-preview-label">Global Search Preview</div>
                                                        <div class="seo-preview-title">{globalSeoPreviewTitle}</div>
                                                        <div class="seo-preview-hint">{globalSeoPreviewUrl}</div>
                                                        <div class="seo-preview-description">{globalSeoPreviewDescription}</div>
                                                    </div>

                                                    <div class="seo-checklist-panel seo-health-panel m-t-sm">
                                                        <div class="seo-checklist-head">
                                                            <div class="seo-health-main">
                                                                <h6 class="m-0 seo-checklist-title">Global SEO health</h6>
                                                                <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                    Estimated from current draft values and runtime SEO defaults.
                                                                </p>
                                                            </div>
                                                            <div class="seo-health-meta">
                                                                <span
                                                                    class="label label-sm seo-health-status-pill"
                                                                    class:good={globalSeoHealthStatus.key === "good"}
                                                                    class:needs-attention={globalSeoHealthStatus.key === "needs-attention"}
                                                                    class:missing-basics={globalSeoHealthStatus.key === "missing-basics"}
                                                                >
                                                                    {globalSeoHealthStatus.label}
                                                                </span>
                                                                <span
                                                                    class="summary-pill seo-check-summary-pill"
                                                                    class:warning={globalSeoCheckCounts.warnings > 0}
                                                                >
                                                                    {globalSeoHealthCompactSummary}
                                                                </span>
                                                            </div>
                                                        </div>

                                                        {#if globalSeoWarningChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Warnings</div>
                                                                <div class="seo-check-list">
                                                                    {#each globalSeoWarningChecks as check}
                                                                        <div class="seo-check-item warning">
                                                                            <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if globalSeoSuggestionChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Suggestions</div>
                                                                <div class="seo-check-list">
                                                                    {#each globalSeoSuggestionChecks as check}
                                                                        <div class="seo-check-item">
                                                                            <span class="label label-sm seo-check-pill">Info</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if !globalSeoWarningChecks.length && !globalSeoSuggestionChecks.length}
                                                            <p class="txt-sm txt-hint m-t-8 m-b-0">No SEO issues found in this section.</p>
                                                        {/if}
                                                    </div>
                                                </aside>
                                            </div>
                                        {:else if activeWebsiteIdentitySeoTab === websiteIdentitySeoTabLocalBusinessKey}
                                            <div class="seo-editor-grid m-t-sm">
                                                <div class="seo-editor-main">
                                                    <div class="local-seo-groups">
                                                        <div class="local-seo-group">
                                                            <div class="local-seo-group-title">Business Identity</div>
                                                            <div class="settings-form-grid two-col m-t-8">
                                                                {#if websiteBusinessNameField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-name">Business Name</label>
                                                                        <input
                                                                            id="cms-website-business-name"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessName}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessTypeField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-type">Business Type</label>
                                                                        <input
                                                                            id="cms-website-business-type"
                                                                            class="input form-input"
                                                                            placeholder="LocalBusiness"
                                                                            bind:value={websiteIdentitySeoDraft.businessType}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Example values: LocalBusiness, Dentist, HealthClub, Restaurant, ProfessionalService.
                                                                        </div>
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessPrimaryCategoryField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-primary-category">Primary Category</label>
                                                                        <input
                                                                            id="cms-website-business-primary-category"
                                                                            class="input form-input"
                                                                            placeholder="Dental clinic"
                                                                            bind:value={websiteIdentitySeoDraft.businessPrimaryCategory}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessPriceRangeField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-price-range">Price Range</label>
                                                                        <input
                                                                            id="cms-website-business-price-range"
                                                                            class="input form-input"
                                                                            placeholder="€€"
                                                                            bind:value={websiteIdentitySeoDraft.businessPriceRange}
                                                                        />
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                        </div>

                                                        <div class="local-seo-group">
                                                            <div class="local-seo-group-title">Contact & Location</div>
                                                            <div class="settings-form-grid two-col m-t-8">
                                                                {#if websiteBusinessPhoneField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-phone">Phone</label>
                                                                        <input
                                                                            id="cms-website-business-phone"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessPhone}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessEmailField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-email">Email</label>
                                                                        <input
                                                                            id="cms-website-business-email"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessEmail}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessAddressField}
                                                                    <div class="form-field local-seo-full-width">
                                                                        <label for="cms-website-business-address">Address</label>
                                                                        <input
                                                                            id="cms-website-business-address"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessAddress}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessCityField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-city">City</label>
                                                                        <input
                                                                            id="cms-website-business-city"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessCity}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessPostalCodeField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-postal-code">Postal Code</label>
                                                                        <input
                                                                            id="cms-website-business-postal-code"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessPostalCode}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessCountryField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-country">Country</label>
                                                                        <input
                                                                            id="cms-website-business-country"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessCountry}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessServiceAreaField}
                                                                    <div class="form-field local-seo-full-width">
                                                                        <label for="cms-website-business-service-area">Service Area</label>
                                                                        <input
                                                                            id="cms-website-business-service-area"
                                                                            class="input form-input"
                                                                            placeholder="Setúbal, Lisbon District, Almada"
                                                                            bind:value={websiteIdentitySeoDraft.businessServiceArea}
                                                                        />
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                        </div>

                                                        <div class="local-seo-group">
                                                            <div class="local-seo-group-title">Local SEO Details</div>
                                                            <div class="settings-form-grid two-col m-t-8">
                                                                {#if websiteBusinessGooglePlaceIdField}
                                                                    <div class="form-field">
                                                                        <label for="cms-website-business-google-place-id">Google Place ID</label>
                                                                        <input
                                                                            id="cms-website-business-google-place-id"
                                                                            class="input form-input"
                                                                            bind:value={websiteIdentitySeoDraft.businessGooglePlaceId}
                                                                        />
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessOpeningHoursField}
                                                                    <div class="form-field local-seo-full-width">
                                                                        <label for="cms-website-business-opening-hours">Opening Hours</label>
                                                                        <textarea
                                                                            id="cms-website-business-opening-hours"
                                                                            class="input form-textarea textarea-input"
                                                                            rows="3"
                                                                            bind:value={websiteIdentitySeoDraft.businessOpeningHours}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Use plain text or a JSON-like schedule format.
                                                                        </div>
                                                                    </div>
                                                                {/if}

                                                                {#if websiteBusinessSocialProfilesField}
                                                                    <div class="form-field local-seo-full-width">
                                                                        <label for="cms-website-business-social-profiles">Social Profiles</label>
                                                                        <textarea
                                                                            id="cms-website-business-social-profiles"
                                                                            class="input form-textarea textarea-input"
                                                                            rows="3"
                                                                            bind:value={websiteIdentitySeoDraft.businessSocialProfiles}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Add profile URLs (one per line) or JSON data for future sameAs structured data.
                                                                        </div>
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <aside class="seo-editor-side">
                                                    <div class="seo-checklist-panel seo-health-panel">
                                                        <div class="seo-checklist-head">
                                                            <div class="seo-health-main">
                                                                <h6 class="m-0 seo-checklist-title">Local Business SEO health</h6>
                                                                <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                    Estimated from current draft values and runtime SEO defaults.
                                                                </p>
                                                            </div>
                                                            <div class="seo-health-meta">
                                                                <span
                                                                    class="label label-sm seo-health-status-pill"
                                                                    class:good={localBusinessSeoHealthStatus.key === "good"}
                                                                    class:needs-attention={localBusinessSeoHealthStatus.key === "needs-attention"}
                                                                    class:missing-basics={localBusinessSeoHealthStatus.key === "missing-basics"}
                                                                >
                                                                    {localBusinessSeoHealthStatus.label}
                                                                </span>
                                                                <span
                                                                    class="summary-pill seo-check-summary-pill"
                                                                    class:warning={localBusinessSeoCheckCounts.warnings > 0}
                                                                >
                                                                    {localBusinessSeoHealthCompactSummary}
                                                                </span>
                                                            </div>
                                                        </div>

                                                        {#if localBusinessSeoWarningChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Warnings</div>
                                                                <div class="seo-check-list">
                                                                    {#each localBusinessSeoWarningChecks as check}
                                                                        <div class="seo-check-item warning">
                                                                            <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if localBusinessSeoSuggestionChecks.length}
                                                            <div class="seo-health-group m-t-8">
                                                                <div class="seo-health-group-title">Suggestions</div>
                                                                <div class="seo-check-list">
                                                                    {#each localBusinessSeoSuggestionChecks as check}
                                                                        <div class="seo-check-item">
                                                                            <span class="label label-sm seo-check-pill">Info</span>
                                                                            <span class="seo-check-message">{check.message}</span>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            </div>
                                                        {/if}

                                                        {#if !localBusinessSeoWarningChecks.length && !localBusinessSeoSuggestionChecks.length}
                                                            <p class="txt-sm txt-hint m-t-8 m-b-0">No SEO issues found in this section.</p>
                                                        {/if}
                                                    </div>

                                                    <div class="seo-checklist-panel seo-impact-panel m-t-sm">
                                                        <div class="seo-checklist-head">
                                                            <div class="seo-health-main">
                                                                <h6 class="m-0 seo-checklist-title">Structured data note</h6>
                                                                <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                    LocalBusiness structured data is generated only when enough business data exists.
                                                                </p>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </aside>
                                            </div>
                                        {:else if activeWebsiteIdentitySeoTab === websiteIdentitySeoTabAdvancedKey}
                                            {#if websiteSeoTitleTemplateField || websiteSeoTitleSeparatorField || websiteSeoCanonicalDomainField}
                                                <div class="seo-editor-grid m-t-sm">
                                                    <div class="seo-advanced-main">
                                                        <div class="seo-advanced-pane">
                                                            <div class="form-grid">
                                                                {#if websiteSeoTitleTemplateField}
                                                                    <div class="form-field seo-field">
                                                                        <label for="cms-website-seo-title-template">Title Template</label>
                                                                        <input
                                                                            id="cms-website-seo-title-template"
                                                                            class="input form-input"
                                                                            placeholder={"{page} | {site}"}
                                                                            bind:value={websiteIdentitySeoDraft.seoTitleTemplate}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Controls how page titles are combined with the website name. Use {`{page}`} and {`{site}`}.
                                                                        </div>
                                                                    </div>
                                                                {/if}

                                                                {#if websiteSeoTitleSeparatorField}
                                                                    <div class="form-field seo-field">
                                                                        <label for="cms-website-seo-title-separator">Title Separator</label>
                                                                        <input
                                                                            id="cms-website-seo-title-separator"
                                                                            class="input form-input"
                                                                            placeholder="|"
                                                                            bind:value={websiteIdentitySeoDraft.seoTitleSeparator}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Used when no title template is provided.
                                                                        </div>
                                                                    </div>
                                                                {/if}

                                                                {#if websiteSeoCanonicalDomainField}
                                                                    <div class="form-field seo-field">
                                                                        <label for="cms-website-seo-canonical-domain">Canonical Domain</label>
                                                                        <input
                                                                            id="cms-website-seo-canonical-domain"
                                                                            class="input form-input"
                                                                            placeholder="https://example.com"
                                                                            bind:value={websiteIdentitySeoDraft.seoCanonicalDomain}
                                                                        />
                                                                        <div class="help-block m-t-6">
                                                                            Used later for canonical URLs and sitemap generation. Example: https://example.com
                                                                        </div>
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <aside class="seo-editor-side">
                                                        <div class="seo-checklist-panel seo-impact-panel">
                                                            <div class="seo-checklist-head">
                                                                <div class="seo-health-main">
                                                                    <h6 class="m-0 seo-checklist-title">Current impact</h6>
                                                                    <p class="txt-sm txt-hint m-b-0 seo-health-helper">
                                                                        Runtime behavior based on current title defaults and canonical domain settings.
                                                                    </p>
                                                                </div>
                                                            </div>

                                                            {#if websiteSeoAdvancedImpactWarningChecks.length}
                                                                <div class="seo-health-group m-t-8">
                                                                    <div class="seo-health-group-title">Warnings</div>
                                                                    <div class="seo-check-list">
                                                                        {#each websiteSeoAdvancedImpactWarningChecks as check}
                                                                            <div class="seo-check-item warning">
                                                                                <span class="label label-sm seo-check-pill warning">Warning</span>
                                                                                <span class="seo-check-message">{check.message}</span>
                                                                            </div>
                                                                        {/each}
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            {#if websiteSeoAdvancedImpactInfoChecks.length}
                                                                <div class="seo-health-group m-t-8">
                                                                    <div class="seo-health-group-title">Notes</div>
                                                                    <div class="seo-check-list">
                                                                        {#each websiteSeoAdvancedImpactInfoChecks as check}
                                                                            <div class="seo-check-item">
                                                                                <span class="label label-sm seo-check-pill">Info</span>
                                                                                <span class="seo-check-message">{check.message}</span>
                                                                            </div>
                                                                        {/each}
                                                                    </div>
                                                                </div>
                                                            {/if}

                                                            {#if !websiteSeoAdvancedImpactWarningChecks.length && !websiteSeoAdvancedImpactInfoChecks.length}
                                                                <p class="txt-sm txt-hint m-t-8 m-b-0">Advanced SEO defaults look healthy.</p>
                                                            {/if}
                                                        </div>
                                                    </aside>
                                                </div>
                                            {:else}
                                                <p class="txt-sm txt-hint m-t-8 m-b-0">Advanced SEO fields are not available for this website collection.</p>
                                            {/if}
                                        {/if}

                                        <div class="settings-section-actions m-t-sm">
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
                                {:else}
                                    <div class="settings-pane">
                                        <p class="txt-sm txt-hint m-b-0">Identity and global SEO fields are not available for this website.</p>
                                    </div>
                                {/if}
                            </div>
                        {:else}
                            <div class="settings-sections m-t-sm">
                                <div class="settings-pane">
                                    <div class="settings-subhead">
                                        <h5 class="m-0">Website Features & Functional Settings</h5>
                                        <p class="txt-sm txt-hint m-b-0 settings-subhead-helper">Configure website features such as WhatsApp, contact forms, newsletter, and languages.</p>
                                    </div>

                                    {#if !hasWebsiteSettingsField}
                                        <p class="txt-danger m-b-0 m-t-sm">Settings field was not found in websites collection.</p>
                                    {:else if !availableWebsiteSettingsFeatures.length}
                                        <p class="txt-hint m-b-0 m-t-sm">No feature settings available for this website.</p>
                                    {:else}
                                        <div class="tabs-header compact combined left operations-tabs settings-feature-tabs m-t-sm">
                                            {#each availableWebsiteSettingsFeatures as featureTab}
                                                <button
                                                    type="button"
                                                    class="tab-item"
                                                    class:active={featureTab.key === activeWebsiteSettingsFeatureKey}
                                                    on:click={() => setActiveWebsiteSettingsFeature(featureTab.key)}
                                                >
                                                    <i class={`${featureTab.icon} tab-icon`} aria-hidden="true" />
                                                    <span class="tab-label">{featureTab.label}</span>
                                                </button>
                                            {/each}
                                        </div>

                                        {#if activeWebsiteSettingsFeature?.key === websiteSettingsLeadsFeatureKey}
                                            <div class="settings-form-wrap m-t-sm">
                                                <div class="settings-subhead">
                                                    <h6 class="m-0">Leads</h6>
                                                    <p class="txt-sm txt-hint m-b-0 settings-subhead-helper">{websiteSettingsLeadsHelperText}</p>
                                                </div>

                                                {#if activeWebsiteSettingsFeature.groupedFeatures?.length}
                                                    <div class="leads-settings-groups m-t-sm">
                                                        {#each activeWebsiteSettingsFeature.groupedFeatures as leadChannel}
                                                            <section class="leads-settings-group">
                                                                <div class="settings-subhead">
                                                                    <h6 class="m-0">{leadChannel.label}</h6>
                                                                </div>

                                                                {#if leadChannel.field}
                                                                    <SchemaForm
                                                                        fields={[leadChannel.field]}
                                                                        value={buildWebsiteSettingsFeatureFormValue(leadChannel.field.key)}
                                                                        showImport={false}
                                                                        path={`websites.${selectedWebsiteId}.settings.${leadChannel.field.key}`}
                                                                        on:change={(event) => handleWebsiteSettingsFeatureGroupChange(leadChannel.field.key, event)}
                                                                    />
                                                                {:else}
                                                                    <p class="txt-sm txt-hint m-b-0">No client-configurable settings are available for this channel yet.</p>
                                                                {/if}
                                                            </section>
                                                        {/each}
                                                    </div>
                                                {:else}
                                                    <p class="txt-sm txt-hint m-b-0 m-t-sm">No lead channels are currently available for this website.</p>
                                                {/if}
                                            </div>
                                        {:else if activeWebsiteSettingsFeatureField}
                                            <div class="settings-form-wrap m-t-sm">
                                                <SchemaForm
                                                    fields={[activeWebsiteSettingsFeatureField]}
                                                    value={activeWebsiteSettingsFeatureValue}
                                                    showImport={false}
                                                    path={`websites.${selectedWebsiteId}.settings.${activeWebsiteSettingsFeatureField.key}`}
                                                    on:change={(event) => handleWebsiteSettingsFeatureGroupChange(activeWebsiteSettingsFeatureField.key, event)}
                                                />
                                            </div>
                                        {:else if activeWebsiteSettingsFeature}
                                            <div class="settings-form-wrap m-t-sm">
                                                <p class="txt-sm txt-hint m-b-0">No client-configurable settings are available for this feature yet.</p>
                                            </div>
                                        {/if}

                                        <div class="settings-section-actions m-t-sm">
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
                                </div>
                            </div>
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
                escClose={false}
                overlayClose={true}
                beforeHide={shouldCloseSectionEditor}
                on:hide={closeSectionEditor}
            >
                <svelte:fragment slot="header">
                    <div class="section-drawer-head">
                        <div class="section-drawer-copy">
                            <strong class="section-drawer-name">{getSectionTitle(selectedEditingSection, Math.max(selectedEditingSectionIndex, 0))}</strong>
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
                </svelte:fragment>

                <div class="section-drawer-body">
                    {#if sectionEditorSupportsTranslations}
                        <div class="section-language-switcher">
                            <div class="tabs-header compact combined left operations-tabs operations-tabs--nested section-language-tabs">
                                <button
                                    type="button"
                                    class="tab-item"
                                    class:active={activeSectionLanguageKey === sectionDefaultLanguageKey}
                                    on:click={() => setActiveSectionLanguage(sectionDefaultLanguageKey)}
                                >
                                    <span>{sectionEditorDefaultLanguageLabel}</span>
                                </button>
                                {#each sectionEditorTranslationLanguages as language}
                                    <button
                                        type="button"
                                        class="tab-item"
                                        class:active={activeSectionLanguageKey === language.code}
                                        on:click={() => setActiveSectionLanguage(language.code)}
                                    >
                                        <span>{language.label}</span>
                                    </button>
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if selectedEditingSectionFields.length}
                        {#key `${selectedEditingSection.id}:${activeSectionLanguageKey}`}
                            <SchemaForm
                                fields={selectedEditingSectionFields}
                                value={selectedEditingSectionDraftFormProps}
                                showImport={false}
                                path={`sections.${selectedEditingSection.id}`}
                                on:propsChange={(event) => updateSectionDraft(selectedEditingSection.id, event.detail)}
                            />
                        {/key}
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
                        disabled={
                            !!isSavingSectionById[selectedEditingSection.id]
                            || (!activeSectionTranslationLanguageCode && !blockPropsField)
                            || (!!activeSectionTranslationLanguageCode && !effectiveBlockTranslationsField)
                        }
                        on:click={() => saveSection(selectedEditingSection)}
                    >
                        {isSavingSectionById[selectedEditingSection.id] ? "Saving..." : "Save changes"}
                    </button>
                </svelte:fragment>
            </OverlayPanel>

            {#if isSectionDiscardConfirmOpen}
                <OverlayPanel
                    active={true}
                    popup={true}
                    class="overlay-panel-sm hide-content cms-drawer-discard-confirm"
                    btnClose={false}
                    overlayClose={true}
                    escClose={true}
                    on:hide={closeSectionDiscardConfirm}
                >
                    <div slot="header" class="drawer-discard-confirm-head">
                        <h4 class="m-0">Discard changes?</h4>
                        <p class="txt-sm txt-hint m-b-0">You have unsaved changes. If you close this panel, your changes will be lost.</p>
                    </div>
                    <svelte:fragment slot="footer">
                        <button type="button" class="btn btn-sm btn-outline" on:click={closeSectionDiscardConfirm}>
                            <span class="txt">Keep editing</span>
                        </button>
                        <button type="button" class="btn btn-sm btn-danger btn-outline" on:click={discardSectionEditorChanges}>
                            <span class="txt">Discard changes</span>
                        </button>
                    </svelte:fragment>
                </OverlayPanel>
            {/if}
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
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: wrap;
    }

    .pages-list-title-wrap {
        min-width: 0;
        width: 100%;
    }

    .pages-list-title-head {
        display: flex;
        align-items: baseline;
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: nowrap;
        min-width: 0;
    }

    .pages-list-subtitle {
        flex: 0 1 auto;
        min-width: 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .pages-list-totals {
        font-size: var(--smFontSize);
        line-height: 1.35;
        margin-top: 4px;
    }

    .pages-search-row {
        margin-top: 10px;
    }

    .pages-search-row .input {
        width: 100%;
    }

    .pages-filter-toolbar {
        margin-top: 9px;
        display: grid;
        grid-template-columns: minmax(0, 1fr);
        gap: 6px;
    }

    .pages-filter-group {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .page-filter-chips {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .page-filter-chip {
        min-height: 28px;
    }

    .page-filter-chip.is-active {
        background: var(--baseAlt2Color);
        border-color: var(--baseAlt2Color);
    }

    .page-seo-filter-chip.is-active {
        border-color: color-mix(in srgb, var(--primaryColor) 28%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--primaryColor) 8%, var(--baseColor));
        color: color-mix(in srgb, var(--primaryColor) 70%, var(--txtPrimaryColor));
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

    .page-row-badges {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
        flex: 0 0 auto;
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

    .page-seo-status-pill {
        font-weight: 600;
        color: var(--txtHintColor);
        border-color: color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 18%, var(--baseColor));
    }

    .page-seo-status-pill.good {
        color: color-mix(in srgb, var(--successColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--successColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--successColor) 12%, var(--baseColor));
    }

    .page-seo-status-pill.needs-attention {
        color: color-mix(in srgb, var(--warningColor) 86%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .page-seo-status-pill.missing-basics {
        color: color-mix(in srgb, var(--dangerColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--dangerColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--dangerColor) 12%, var(--baseColor));
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
        column-gap: 12px;
        row-gap: 4px;
        padding: 8px 14px;
    }

    .section-drawer-head {
        width: 100%;
        min-width: 0;
        display: flex;
        align-items: flex-start;
    }

    .section-drawer-copy {
        min-width: 0;
        flex: 1 1 auto;
        display: flex;
        flex-direction: column;
        gap: 4px;
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

    .section-language-switcher {
        display: flex;
        align-items: flex-start;
        flex-direction: column;
        gap: 6px;
        margin: 0 0 10px;
    }

    .section-language-tabs {
        width: auto;
        max-width: 100%;
        flex: 0 0 auto;
        margin-top: 0;
        display: inline-flex;
        flex-wrap: wrap;
        row-gap: 4px;
        overflow: visible;
    }

    .section-language-tabs .tab-item {
        display: inline-flex;
        flex: 0 0 auto;
        width: auto;
        align-items: center;
        gap: 6px;
        white-space: nowrap;
    }

    .drawer-discard-confirm-head {
        display: flex;
        flex-direction: column;
        gap: 6px;
        text-align: left;
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
        padding: 0;
        border: 0;
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
    }

    .seo-page-wrap {
        border-top: 1px solid var(--baseAlt2Color);
        padding-top: 8px;
    }

    .seo-page-head {
        display: flex;
        align-items: baseline;
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: wrap;
    }

    .seo-page-head-helper {
        flex: 1 1 340px;
        min-width: 240px;
    }

    .seo-field.form-field {
        margin-bottom: 0;
    }

    .page-seo-tabs-row {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        width: auto;
    }

    .page-seo-tabs-row--compact {
        margin-top: 4px;
    }

    .page-seo-tabs {
        display: inline-flex;
        align-items: center;
        width: fit-content !important;
        max-width: 100%;
        flex-wrap: wrap;
    }

    .seo-editor-grid {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
        gap: 12px;
        align-items: start;
    }

    .seo-editor-main {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .seo-editor-side {
        position: sticky;
        top: 10px;
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .page-seo-file-control {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 10px;
    }

    .page-seo-file-control .settings-file-row {
        flex: 1 1 260px;
        min-width: 0;
    }

    .page-seo-file-thumb {
        width: 58px;
        height: 58px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        border-radius: var(--baseRadius);
        overflow: hidden;
        background: var(--baseAlt1Color);
        display: inline-flex;
        align-items: center;
        justify-content: center;
    }

    .page-seo-file-thumb img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }

    .page-seo-file-input {
        width: 100%;
        min-height: 44px;
        padding: 8px 10px;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 78%, transparent);
    }

    .page-seo-file-input::file-selector-button {
        height: 28px;
        padding: 0 12px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 76%, transparent);
        border-radius: 8px;
        background: color-mix(in srgb, var(--baseAlt1Color) 16%, var(--baseColor));
        color: var(--txtPrimaryColor);
        font-size: var(--smFontSize);
        font-weight: 500;
        margin-right: 8px;
        cursor: pointer;
        transition: background-color var(--baseAnimationSpeed), border-color var(--baseAnimationSpeed);
    }

    .page-seo-file-input:hover::file-selector-button,
    .page-seo-file-input:focus-visible::file-selector-button {
        border-color: color-mix(in srgb, var(--baseAlt3Color) 90%, transparent);
        background: color-mix(in srgb, var(--baseAlt2Color) 60%, var(--baseColor));
    }

    .seo-advanced-pane {
        border-top: 0;
        padding-top: 0;
    }

    .seo-advanced-main {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .seo-advanced-section {
        border-top: 0;
        padding-top: 0;
    }

    .seo-advanced-section + .seo-advanced-section {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        padding-top: 8px;
    }

    .seo-advanced-section-title {
        font-size: 12px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .seo-advanced-grid {
        align-items: start;
    }

    .seo-toggle-field {
        margin-bottom: 0;
        min-height: 24px;
    }

    .seo-impact-panel {
        border-color: color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 6%, var(--baseColor));
    }

    .seo-health-group-title {
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: var(--txtHintColor);
    }

    .settings-workspace {
        border: 0;
        border-radius: 0;
        background: transparent;
        padding: 0;
        min-height: 0;
    }

    .settings-head {
        display: flex;
        align-items: baseline;
        justify-content: flex-start;
        gap: 10px;
        flex-wrap: wrap;
        min-width: 0;
    }

    .settings-head-helper {
        flex: 1 1 340px;
        min-width: 240px;
    }

    .settings-nav-row {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 8px;
        flex-wrap: wrap;
    }

    .settings-nav-row--compact {
        margin-top: 4px;
    }

    .settings-nav-tabs,
    .settings-feature-tabs {
        display: inline-flex;
        align-items: center;
        width: fit-content;
        max-width: 100%;
        flex-wrap: wrap;
    }

    .settings-sections {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .settings-pane {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .settings-identity-pane {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        padding-top: 8px;
    }

    .settings-identity-seo-tabs {
        width: fit-content;
    }

    .settings-identity-section {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .settings-identity-section + .settings-identity-section {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 86%, transparent);
        padding-top: 10px;
    }

    .settings-form-wrap {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        padding-top: 10px;
    }

    .settings-subhead {
        display: flex;
        align-items: baseline;
        justify-content: flex-start;
        gap: 9px;
        flex-wrap: wrap;
        min-width: 0;
    }

    .settings-subhead-helper {
        flex: 1 1 320px;
        min-width: 220px;
    }

    .leads-settings-groups {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .leads-settings-group {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        padding-top: 10px;
    }

    .leads-settings-group:first-child {
        border-top: 0;
        padding-top: 0;
    }

    .settings-form-grid {
        display: grid;
        gap: 10px 12px;
    }

    .settings-form-grid.two-col {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .settings-form-grid.one-col {
        grid-template-columns: 1fr;
    }

    .local-seo-groups {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .local-seo-group {
        border-top: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        padding-top: 8px;
    }

    .local-seo-group:first-child {
        border-top: 0;
        padding-top: 0;
    }

    .local-seo-group-title {
        margin: 0;
        font-size: 12px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .local-seo-full-width {
        grid-column: 1 / -1;
    }

    .settings-section-actions {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 8px;
        flex-wrap: wrap;
    }

    .settings-file-row {
        display: block;
    }

    .settings-file-state {
        --labelHPadding: 8px;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        color: var(--txtHintColor);
        background: color-mix(in srgb, var(--baseAlt1Color) 26%, var(--baseColor));
    }

    .seo-preview-card {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 92%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 9px 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .seo-preview-label {
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: var(--txtHintColor);
    }

    .seo-preview-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
        line-height: 1.25;
    }

    .seo-preview-description {
        font-size: var(--smFontSize);
        color: var(--txtHintColor);
        line-height: var(--smLineHeight);
    }

    .seo-preview-hint {
        font-size: 11px;
        color: color-mix(in srgb, var(--txtHintColor) 92%, var(--txtPrimaryColor));
    }

    .seo-field-helper {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .seo-count-pill {
        --labelHPadding: 7px;
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 26%, var(--baseColor));
        color: var(--txtHintColor);
    }

    .seo-checklist-panel {
        border: 1px solid color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        border-radius: var(--baseRadius);
        background: var(--baseColor);
        padding: 8px 10px;
    }

    .seo-checklist-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
    }

    .seo-checklist-title {
        font-size: 13px;
        font-weight: 600;
        color: var(--txtPrimaryColor);
    }

    .seo-health-main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .seo-health-helper {
        font-size: 11px;
        line-height: 1.35;
    }

    .seo-health-meta {
        display: inline-flex;
        align-items: center;
        justify-content: flex-end;
        gap: 6px;
        flex-wrap: wrap;
    }

    .seo-health-status-pill {
        --labelHPadding: 8px;
        min-height: 20px;
        color: var(--txtHintColor);
        border-color: color-mix(in srgb, var(--baseAlt2Color) 88%, transparent);
        background: color-mix(in srgb, var(--baseAlt1Color) 18%, var(--baseColor));
        font-weight: 600;
    }

    .seo-health-status-pill.good {
        color: color-mix(in srgb, var(--successColor) 85%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--successColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--successColor) 12%, var(--baseColor));
    }

    .seo-health-status-pill.needs-attention {
        color: color-mix(in srgb, var(--warningColor) 86%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .seo-health-status-pill.missing-basics {
        color: color-mix(in srgb, var(--dangerColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--dangerColor) 40%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--dangerColor) 12%, var(--baseColor));
    }

    .seo-check-summary-pill {
        --labelHPadding: 9px;
        min-height: 20px;
        color: var(--txtHintColor);
        background: color-mix(in srgb, var(--baseAlt1Color) 22%, var(--baseColor));
    }

    .seo-check-summary-pill.warning {
        color: color-mix(in srgb, var(--warningColor) 84%, var(--txtPrimaryColor));
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .seo-check-list {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .seo-check-item {
        display: flex;
        align-items: flex-start;
        gap: 7px;
        padding: 6px 0;
        font-size: var(--smFontSize);
        line-height: var(--smLineHeight);
        color: var(--txtHintColor);
    }

    .seo-check-item + .seo-check-item {
        border-top: 1px dashed color-mix(in srgb, var(--baseAlt2Color) 80%, transparent);
    }

    .seo-check-item.warning {
        color: color-mix(in srgb, var(--warningColor) 80%, var(--txtPrimaryColor));
    }

    .seo-check-item.pass {
        color: color-mix(in srgb, var(--successColor) 82%, var(--txtPrimaryColor));
    }

    .seo-check-pill {
        --labelHPadding: 7px;
        min-height: 18px;
        flex: 0 0 auto;
        border-color: color-mix(in srgb, var(--baseAlt2Color) 90%, transparent);
        color: var(--txtHintColor);
        background: var(--baseColor);
    }

    .seo-check-pill.warning {
        border-color: color-mix(in srgb, var(--warningColor) 45%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--warningColor) 88%, var(--txtPrimaryColor));
        background: color-mix(in srgb, var(--warningColor) 14%, var(--baseColor));
    }

    .seo-check-pill.pass {
        border-color: color-mix(in srgb, var(--successColor) 42%, var(--baseAlt2Color));
        color: color-mix(in srgb, var(--successColor) 86%, var(--txtPrimaryColor));
        background: color-mix(in srgb, var(--successColor) 12%, var(--baseColor));
    }

    .seo-check-message {
        flex: 1 1 auto;
        min-width: 0;
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

        .seo-editor-grid {
            grid-template-columns: 1fr;
        }

        .seo-editor-side {
            position: static;
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

        .pages-filter-group {
            align-items: flex-start;
            gap: 6px;
        }

        .seo-page-head {
            align-items: flex-start;
        }

        .page-seo-file-control {
            align-items: flex-start;
        }

        .settings-head,
        .settings-subhead {
            align-items: flex-start;
        }

        .settings-head-helper,
        .settings-subhead-helper {
            min-width: 0;
        }

        .settings-form-grid.two-col {
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
