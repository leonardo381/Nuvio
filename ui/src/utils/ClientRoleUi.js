import ApiClient from "@/utils/ApiClient";

// NUVIO CUSTOM START: Client role collection access and field visibility configuration.
const clientEditableCollectionNames = new Set([
    "assets",
    "blocks",
    "pages",
    "websites",
    "contacts",
    "whatsapp",
    "whatsapp_interactions",
]);

const clientCollectionUiConfig = {
    assets: {
        hiddenGridFields: ["id", "checksum", "size"],
        hiddenFormFields: ["id", "checksum", "size"],
        hiddenPreviewFields: [],
    },

    blocks: {
        hiddenGridFields: ["slot", "id", "component_key", "enabled", "props"],
        hiddenFormFields: ["slot", "id", "component_key", "enabled", "props"],
        hiddenPreviewFields: [],
    },

    pages: {
        hiddenGridFields: ["id", "slug"],
        hiddenFormFields: ["id", "slug"],
        hiddenPreviewFields: [],
    },

    websites: {
        hiddenGridFields: ["id", "slug"],
        hiddenFormFields: ["id", "slug"],
        hiddenPreviewFields: [],
    },

    contacts: {
        hiddenGridFields: ["id"],
        hiddenFormFields: ["id"],
        hiddenPreviewFields: [],
    },

    whatsapp: {
        hiddenGridFields: ["id"],
        hiddenFormFields: ["id"],
        hiddenPreviewFields: [],
    },

    whatsapp_interactions: {
        hiddenGridFields: ["id"],
        hiddenFormFields: ["id"],
        hiddenPreviewFields: [],
    },

};
// NUVIO CUSTOM END: Client role collection access and field visibility configuration.

function normalizeCollectionName(collectionOrName) {
    if (!collectionOrName) {
        return "";
    }

    if (typeof collectionOrName === "string") {
        return collectionOrName.toLowerCase();
    }

    return (collectionOrName?.name || "").toLowerCase();
}

function getCollectionConfig(collectionOrName) {
    return clientCollectionUiConfig[normalizeCollectionName(collectionOrName)] || null;
}

function isClientEditableCollection(collectionOrName) {
    return clientEditableCollectionNames.has(normalizeCollectionName(collectionOrName));
}

function getHiddenFieldNamesForClient(collectionOrName, key) {
    const config = getCollectionConfig(collectionOrName);
    const list = config?.[key];

    if (!Array.isArray(list)) {
        return [];
    }

    return list.map((fieldName) => (fieldName || "").toLowerCase()).filter(Boolean);
}

function isClientFieldHidden(collectionOrName, fieldName, key) {
    if (!ApiClient.isClientSuperuser()) {
        return false;
    }

    const field = (fieldName || "").toLowerCase();
    if (!field) {
        return false;
    }

    return getHiddenFieldNamesForClient(collectionOrName, key).includes(field);
}

export default {
    clientEditableCollectionNames,
    clientCollectionUiConfig,
    isClientEditableCollection,
    getHiddenFieldNamesForClient,
    isClientFieldHidden,
};
