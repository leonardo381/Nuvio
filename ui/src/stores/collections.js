import ApiClient from "@/utils/ApiClient";
import CommonHelper from "@/utils/CommonHelper";
import { get, writable } from "svelte/store";

export const collections = writable([]);
export const activeCollection = writable({});
export const isCollectionsLoading = writable(false);
export const hasCollectionsLoaded = writable(false);
export const collectionsLoadError = writable("");
export const protectedFilesCollectionsCache = writable({});
export const scaffolds = writable({});

let notifyChannel;
let loadCollectionsPromise = null;

function createCollectionStub(name, fields = []) {
    const normalizedName = `${name || ""}`.trim();

    return {
        id: `nuvio_client_${normalizeCollectionLookupKey(normalizedName) || "collection"}`,
        name: normalizedName,
        type: "base",
        system: false,
        fields: (Array.isArray(fields) ? fields : [])
            .map((field) => {
                if (typeof field === "string") {
                    const fieldName = `${field || ""}`.trim();
                    return fieldName ? { name: fieldName, type: "text" } : null;
                }

                const fieldName = `${field?.name || ""}`.trim();
                if (!fieldName) {
                    return null;
                }

                const options = CommonHelper.isObject(field?.options)
                    ? CommonHelper.clone(field.options)
                    : undefined;
                const relationCollectionId = `${field?.collectionId || ""}`.trim();

                return {
                    name: fieldName,
                    type: `${field?.type || "text"}`.trim() || "text",
                    ...(options ? { options } : {}),
                    ...(relationCollectionId ? { collectionId: relationCollectionId } : {}),
                };
            })
            .filter(Boolean),
    };
}

const clientScopedBackofficeCollections = [
    createCollectionStub("websites", [
        "title",
        "name",
        "slug",
        "domain",
        "settings",
        "logo",
        "seoTitle",
        "seoDescription",
        "seoImage",
        "seo_title_template",
        "seo_title_separator",
        "seo_canonical_domain",
        "business_name",
        "business_type",
        "business_primary_category",
        "business_phone",
        "business_email",
        "business_address",
        "business_city",
        "business_postal_code",
        "business_country",
        "business_service_area",
        "business_opening_hours",
        "business_google_place_id",
        "business_social_profiles",
        "business_price_range",
        "created",
        "updated",
    ]),
    createCollectionStub("pages", [
        { name: "website", type: "relation", collectionId: "nuvio_client_websites" },
        "title",
        "name",
        "slug",
        "path",
        "url",
        "status",
        "published",
        "visible",
        "seo_title",
        "seo_description",
        "seo_social_image",
        "seo_canonical_url",
        "seo_noindex",
        "seo_exclude_from_sitemap",
        "seo_focus_keyword",
        "seo_translations",
        "created",
        "updated",
    ]),
    createCollectionStub("blocks", [
        { name: "page", type: "relation", collectionId: "nuvio_client_pages" },
        "website",
        { name: "component", type: "relation", collectionId: "nuvio_client_components" },
        "component_key",
        "variant",
        "slot",
        "region",
        "displayOrder",
        "order",
        "props",
        "translations",
        "enabled",
        "visible",
        "status",
        "created",
        "updated",
    ]),
    createCollectionStub("components", [
        "key",
        "component_key",
        "name",
        "label",
        "title",
        "category",
        "group",
        "variant",
        "defaultVariant",
        "schema",
        "created",
        "updated",
    ]),
    createCollectionStub("Contacts", [
        "website",
        "name",
        "email",
        "phone",
        "subject",
        "message",
        "source",
        "page",
        "notes",
        "lastContactedAt",
        { name: "status", type: "select", options: { values: ["new", "read", "archived"] } },
        "created",
        "updated",
    ]),
    createCollectionStub("Whatsapp", [
        "website",
        "name",
        "email",
        "phone",
        "message",
        "defaultMessage",
        "source",
        "page",
        "notes",
        "lastContactedAt",
        { name: "status", type: "select", options: { values: ["new", "read", "archived"] } },
        "created",
        "updated",
    ]),
    createCollectionStub("Appointments", [
        "website",
        "service",
        "serviceNameSnapshot",
        "serviceDurationMinutesSnapshot",
        "serviceDescriptionSnapshot",
        "name",
        "email",
        "phone",
        "date",
        "time",
        "notes",
        "message",
        "internalNotes",
        "confirmedAt",
        "cancelledAt",
        "rescheduledAt",
        "archivedAt",
        { name: "status", type: "select", options: { values: ["pending", "confirmed", "cancelled"] } },
        "created",
        "updated",
    ]),
    createCollectionStub("BookingServices", [
        "website",
        "name",
        "description",
        "durationMinutes",
        "duration",
        "priority",
        "active",
        "displayOrder",
        "calendarBlockingMode",
        "autoConfirm",
        "created",
        "updated",
    ]),
    createCollectionStub("BookingAvailability", [
        "website",
        "service",
        "dayOfWeek",
        "startTime",
        "endTime",
        "active",
        "enabled",
        "capacity",
        "created",
        "updated",
    ]),
    createCollectionStub("BookingExceptions", [
        "website",
        "service",
        "date",
        "startTime",
        "endTime",
        "type",
        "reason",
        "note",
        "active",
        "enabled",
        "created",
        "updated",
    ]),
    createCollectionStub("Subscribers", [
        "website",
        "email",
        "name",
        "source",
        "groups",
        "confirmedAt",
        "unsubscribedAt",
        { name: "status", type: "select", options: { values: ["pending", "active", "unsubscribed"] } },
        "created",
        "updated",
    ]),
    createCollectionStub("Campaigns", [
        "website",
        "subject",
        "body",
        "recipientsType",
        "recipientsIds",
        "recipientsCount",
        "sentAt",
        { name: "status", type: "select", options: { values: ["draft", "sent"] } },
        "created",
        "updated",
    ]),
    createCollectionStub("SubscriberGroups", [
        "website",
        "name",
        "slug",
        "created",
        "updated",
    ]),
];

if (typeof BroadcastChannel != "undefined") {
    notifyChannel = new BroadcastChannel("collections");

    notifyChannel.onmessage = () => {
        loadCollections(get(activeCollection)?.id);
    };
}

function notifyOtherTabs() {
    notifyChannel?.postMessage("reload");
}

function normalizeCollectionLookupKey(value) {
    return `${value || ""}`
        .trim()
        .toLowerCase()
        .replace(/[\s_-]+/g, "");
}

export function resolveCollectionName(collectionList = [], requestedNames = []) {
    const list = Array.isArray(collectionList) ? collectionList : [];
    const requested = (Array.isArray(requestedNames) ? requestedNames : [requestedNames])
        .map((name) => normalizeCollectionLookupKey(name))
        .filter(Boolean);

    if (!requested.length) {
        return "";
    }

    const namesByLookupKey = new Map();
    for (const collection of list) {
        const collectionName = `${collection?.name || ""}`.trim();
        const lookupKey = normalizeCollectionLookupKey(collectionName);
        if (lookupKey && collectionName && !namesByLookupKey.has(lookupKey)) {
            namesByLookupKey.set(lookupKey, collectionName);
        }
    }

    for (const lookupKey of requested) {
        const resolvedName = namesByLookupKey.get(lookupKey);
        if (resolvedName) {
            return resolvedName;
        }
    }

    return "";
}

export function findCollectionByRequiredNames(collectionList = [], requestedNames = []) {
    const list = Array.isArray(collectionList) ? collectionList : [];
    const resolvedName = resolveCollectionName(list, requestedNames);

    if (!resolvedName) {
        return null;
    }

    return list.find((collection) => `${collection?.name || ""}`.trim() === resolvedName) || null;
}

export function changeActiveCollectionByIdOrName(collectionIdOrName) {
    collections.update((list) => {
        const found = list.find((c) => c.id == collectionIdOrName || c.name == collectionIdOrName);
        if (found) {
            activeCollection.set(found);
        } else if (list.length) {
            activeCollection.set(list.find((c) => !c.system) || list[0]);
        }

        return list;
    });
}

// add or update collection
export function addCollection(collection) {
    activeCollection.update((current) => {
        return CommonHelper.isEmpty(current?.id) || current.id === collection.id ? collection : current;
    });

    collections.update((list) => {
        CommonHelper.pushOrReplaceByKey(list, collection, "id");

        refreshProtectedFilesCollectionsCache();

        notifyOtherTabs();

        return CommonHelper.sortCollections(list);
    });
}

export function removeCollection(collection) {
    collections.update((list) => {
        CommonHelper.removeByKey(list, "id", collection.id);

        activeCollection.update((current) => {
            if (current.id === collection.id) {
                return list.find((c) => !c.system) || list[0];
            }
            return current;
        });

        refreshProtectedFilesCollectionsCache();

        notifyOtherTabs();

        return list;
    });
}

export async function refreshScaffolds() {
    if (ApiClient.isSuperuserAuth() && ApiClient.isClientSuperuser()) {
        scaffolds.set({});
        return {};
    }

    scaffolds.set(await ApiClient.collections.getScaffolds());
}

// load all collections
export async function loadCollections(activeIdOrName = null) {
    if (loadCollectionsPromise) {
        return loadCollectionsPromise;
    }

    isCollectionsLoading.set(true);
    collectionsLoadError.set("");

    loadCollectionsPromise = (async () => {
        try {
            if (typeof ApiClient.whenAuthReady === "function") {
                await ApiClient.whenAuthReady();
            }

            if (ApiClient.isSuperuserAuth() && ApiClient.isClientSuperuser()) {
                const resultCollections = CommonHelper.sortCollections(
                    CommonHelper.clone(clientScopedBackofficeCollections),
                );

                scaffolds.set({});
                collections.set(resultCollections);

                const found = activeIdOrName && resultCollections.find((c) => c.id == activeIdOrName || c.name == activeIdOrName);
                if (found) {
                    activeCollection.set(found);
                } else if (resultCollections.length) {
                    activeCollection.set(resultCollections.find((c) => !c.system) || resultCollections[0]);
                }

                refreshProtectedFilesCollectionsCache();
                hasCollectionsLoaded.set(true);
                return;
            }

            let [resultScaffolds, resultCollections] = await Promise.all([
                ApiClient.collections.getScaffolds(),
                ApiClient.collections.getFullList(),
            ]);

            scaffolds.set(resultScaffolds);

            resultCollections = CommonHelper.sortCollections(
                Array.isArray(resultCollections) ? resultCollections : [],
            );

            // Some refresh paths can briefly return an empty list even though collections exist.
            // Retry once before flagging the module as unavailable.
            if (!resultCollections.length) {
                await new Promise((resolve) => setTimeout(resolve, 150));
                const retryCollections = await ApiClient.collections.getFullList();
                resultCollections = CommonHelper.sortCollections(
                    Array.isArray(retryCollections) ? retryCollections : [],
                );
            }

            if (!resultCollections.length) {
                throw new Error("Unable to verify collections right now. Empty collections response.");
            }

            collections.set(resultCollections);

            const found = activeIdOrName && resultCollections.find((c) => c.id == activeIdOrName || c.name == activeIdOrName);
            if (found) {
                activeCollection.set(found);
            } else if (resultCollections.length) {
                activeCollection.set(resultCollections.find((c) => !c.system) || resultCollections[0]);
            }

            refreshProtectedFilesCollectionsCache();
            hasCollectionsLoaded.set(true);
        } catch (err) {
            hasCollectionsLoaded.set(false);
            collectionsLoadError.set(err?.message || "Unable to verify collections right now.");
            ApiClient.error(err);
        } finally {
            isCollectionsLoading.set(false);
            loadCollectionsPromise = null;
        }
    })();

    return loadCollectionsPromise;
}

function refreshProtectedFilesCollectionsCache() {
    protectedFilesCollectionsCache.update((cache) => {
        collections.update((current) => {
            for (let c of current) {
                cache[c.id] = !!c.fields?.find((f) => f.type == "file" && f.protected);
            }

            return current;
        });

        return cache;
    });
}
