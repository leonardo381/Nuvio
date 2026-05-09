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
