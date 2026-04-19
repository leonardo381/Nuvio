// NUVIO CUSTOM START: Static website-level settings schema with role-aware filtering.
const ROLE_ADMIN = "admin";
const ROLE_CLIENT = "client";
const defaultEditableBy = [ROLE_ADMIN];
const websiteFeatureFlagKeys = ["whatsapp", "contactForm", "reviews", "newsletter", "booking", "reports", "i18n"];
const websiteFeatureSectionKeys = new Set(["whatsapp", "contactForm", "reviews", "newsletter", "i18n"]);

export const websiteSettingsSchema = {
    fields: [
        {
            key: "featureFlags",
            label: "Feature availability",
            type: "object",
            editableBy: [ROLE_ADMIN],
            fields: [
                {
                    key: "whatsapp",
                    label: "WhatsApp",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "contactForm",
                    label: "Contact form",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "reviews",
                    label: "Reviews",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "newsletter",
                    label: "Newsletter",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "booking",
                    label: "Booking",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "reports",
                    label: "Reports",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "i18n",
                    label: "Internationalization",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
            ],
        },
        {
            key: "whatsapp",
            label: "WhatsApp",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "enabled",
                    label: "Enabled",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "phone",
                    label: "Phone",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "defaultMessage",
                    label: "Default message",
                    type: "textarea",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "showFloatingButton",
                    label: "Show floating button",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
            ],
        },
        {
            key: "contactForm",
            label: "Contact form",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "enabled",
                    label: "Enabled",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "emailDestination",
                    label: "Email destination",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "confirmationMessage",
                    label: "Confirmation message",
                    type: "textarea",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "fields",
                    label: "Form fields",
                    type: "object",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                    fields: [
                        {
                            key: "phone",
                            label: "Enable phone field",
                            type: "bool",
                            default: true,
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                        },
                    ],
                },
            ],
        },
        {
            key: "reviews",
            label: "Reviews",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "googlePlaceId",
                    label: "Google Place ID",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "reviewLink",
                    label: "Review link",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
            ],
        },
        {
            key: "newsletter",
            label: "Newsletter",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "doubleOptIn",
                    label: "Double opt-in",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
            ],
        },
        {
            key: "i18n",
            label: "Internationalization",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "enabled",
                    label: "Enabled",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                },
                {
                    key: "languages",
                    label: "Languages",
                    type: "array",
                    itemLabel: "Language",
                    default: [],
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                    item: {
                        fields: [
                            {
                                key: "code",
                                label: "Code",
                                type: "text",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                            {
                                key: "label",
                                label: "Label",
                                type: "text",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        ],
                    },
                },
            ],
        },
    ],
};

function normalizeRole(role) {
    return role === ROLE_CLIENT ? ROLE_CLIENT : ROLE_ADMIN;
}

function isPlainObject(value) {
    return !!value && typeof value === "object" && !Array.isArray(value);
}

function cloneValue(value) {
    if (typeof structuredClone === "function") {
        return structuredClone(value);
    }

    return JSON.parse(JSON.stringify(value));
}

function getEditableBy(field) {
    if (Array.isArray(field?.editableBy) && field.editableBy.length) {
        return field.editableBy;
    }

    return defaultEditableBy;
}

function filterFieldsByRole(fields, role) {
    const filtered = [];

    for (const field of fields || []) {
        if (!getEditableBy(field).includes(role)) {
            continue;
        }

        const nextField = { ...field };

        if (field.type === "object") {
            nextField.fields = filterFieldsByRole(field.fields || [], role);
            if (!nextField.fields.length) {
                continue;
            }
        }

        if (field.type === "array" && field.item?.fields) {
            const filteredItemFields = filterFieldsByRole(field.item.fields, role);
            if (!filteredItemFields.length) {
                continue;
            }

            nextField.item = {
                ...field.item,
                fields: filteredItemFields,
            };
        }

        filtered.push(nextField);
    }

    return filtered;
}

function filterFieldsByClientFeatureFlags(fields, featureFlags) {
    const result = [];

    for (const field of fields || []) {
        if (!websiteFeatureSectionKeys.has(field?.key)) {
            result.push(field);
            continue;
        }

        if (featureFlags?.[field.key] !== false) {
            result.push(field);
        }
    }

    return result;
}

function parseRawSettings(rawSettings) {
    if (isPlainObject(rawSettings)) {
        return cloneValue(rawSettings);
    }

    if (typeof rawSettings === "string") {
        try {
            const parsed = JSON.parse(rawSettings);
            return isPlainObject(parsed) ? parsed : {};
        } catch (_) {
            return {};
        }
    }

    return {};
}

function getDefaultValue(field) {
    if (typeof field.default !== "undefined") {
        return cloneValue(field.default);
    }

    if (field.type === "object") {
        return {};
    }

    if (field.type === "array") {
        return [];
    }

    if (field.type === "bool") {
        return false;
    }

    return "";
}

function normalizeFieldValue(field, value) {
    if (field.type === "object") {
        const source = isPlainObject(value) ? value : {};
        const nextObject = {};
        const knownKeys = new Set((field.fields || []).map((childField) => childField.key));

        for (const key of Object.keys(source)) {
            if (!knownKeys.has(key)) {
                nextObject[key] = cloneValue(source[key]);
            }
        }

        for (const childField of field.fields || []) {
            nextObject[childField.key] = normalizeFieldValue(childField, source[childField.key]);
        }

        return nextObject;
    }

    if (field.type === "array") {
        return Array.isArray(value) ? cloneValue(value) : getDefaultValue(field);
    }

    if (field.type === "bool") {
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

        return getDefaultValue(field);
    }

    if (typeof value === "undefined" || value === null) {
        return getDefaultValue(field);
    }

    return value;
}

export function getWebsiteSettingsSchemaForRole(role = ROLE_ADMIN, rawSettings = null) {
    const normalizedRole = normalizeRole(role);
    const roleFilteredFields = filterFieldsByRole(websiteSettingsSchema.fields, normalizedRole);

    if (normalizedRole !== ROLE_CLIENT) {
        return {
            fields: roleFilteredFields,
        };
    }

    const normalizedSettings = normalizeWebsiteSettingsValue(rawSettings, websiteSettingsSchema.fields);
    const featureFlags = normalizedSettings?.featureFlags || {};

    return {
        fields: filterFieldsByClientFeatureFlags(roleFilteredFields, featureFlags),
    };
}

export function normalizeWebsiteSettingsValue(rawSettings, schemaFields = websiteSettingsSchema.fields) {
    const source = parseRawSettings(rawSettings);
    const normalized = {};
    const knownKeys = new Set((schemaFields || []).map((field) => field.key));

    for (const key of Object.keys(source)) {
        if (!knownKeys.has(key)) {
            normalized[key] = cloneValue(source[key]);
        }
    }

    for (const field of schemaFields || []) {
        normalized[field.key] = normalizeFieldValue(field, source[field.key]);
    }

    return normalized;
}

export function isWebsiteFeatureAvailable(rawSettings, featureKey, fallback = true) {
    if (!websiteFeatureFlagKeys.includes(featureKey)) {
        return fallback;
    }

    const normalizedSettings = normalizeWebsiteSettingsValue(rawSettings, websiteSettingsSchema.fields);
    const value = normalizedSettings?.featureFlags?.[featureKey];

    return typeof value === "boolean" ? value : fallback;
}
// NUVIO CUSTOM END: Static website-level settings schema with role-aware filtering.
