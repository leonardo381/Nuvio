import ApiClient from "@/utils/ApiClient";

const clientEditableCollectionNames = new Set(["assets", "blocks", "pages", "websites"]);

const clientCollectionUiConfig = {
    blocks: {
        hiddenGridFields: ["slot"],
        hiddenFormFields: ["slot"],
        hiddenPreviewFields: [],
    },
};

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
