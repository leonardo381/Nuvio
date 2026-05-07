// NUVIO CUSTOM START: Static website-level settings schema with role-aware filtering.
const ROLE_ADMIN = "admin";
const ROLE_CLIENT = "client";
const defaultEditableBy = [ROLE_ADMIN];
const websiteFeatureFlagKeys = ["whatsapp", "contactForm", "reviews", "newsletter", "booking", "reports", "i18n"];
const websiteFeatureSectionKeys = new Set(["whatsapp", "contactForm", "reviews", "newsletter", "booking", "reports", "i18n"]);
const hiddenWebsiteSettingsFeatureKeys = new Set(["reviews"]);

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
                {
                    key: "emailNotifications",
                    label: "Email notifications",
                    type: "object",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                    fields: [
                        {
                            key: "enabled",
                            label: "Send email notifications",
                            type: "bool",
                            default: false,
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                        },
                        {
                            key: "to",
                            label: "To recipients",
                            type: "array",
                            itemLabel: "Recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                        {
                            key: "cc",
                            label: "CC recipients",
                            type: "array",
                            itemLabel: "CC recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                    ],
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
                    key: "emailNotifications",
                    label: "Email notifications",
                    type: "object",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                    fields: [
                        {
                            key: "enabled",
                            label: "Send email notifications",
                            type: "bool",
                            default: false,
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                        },
                        {
                            key: "to",
                            label: "To recipients",
                            type: "array",
                            itemLabel: "Recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                        {
                            key: "cc",
                            label: "CC recipients",
                            type: "array",
                            itemLabel: "CC recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                    ],
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
                    key: "enabled",
                    label: "Enabled",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "googlePlaceId",
                    label: "Google Place ID",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "reviewLink",
                    label: "Google review link",
                    type: "text",
                    default: "",
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "displayEnabled",
                    label: "Show reviews on website",
                    type: "bool",
                    default: false,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "maxReviews",
                    label: "Maximum reviews shown",
                    type: "text",
                    default: "6",
                    editableBy: [ROLE_ADMIN],
                    options: {
                        pattern: "^[0-9]+$",
                    },
                },
                {
                    key: "minRating",
                    label: "Minimum rating",
                    type: "select",
                    default: "4",
                    editableBy: [ROLE_ADMIN],
                    options: [
                        { label: "1", value: "1" },
                        { label: "2", value: "2" },
                        { label: "3", value: "3" },
                        { label: "4", value: "4" },
                        { label: "5", value: "5" },
                    ],
                },
                {
                    key: "sortOrder",
                    label: "Sort reviews by",
                    type: "select",
                    default: "newest",
                    editableBy: [ROLE_ADMIN],
                    options: [
                        { label: "Newest first", value: "newest" },
                        { label: "Highest rating", value: "highestRating" },
                    ],
                },
                {
                    key: "showRating",
                    label: "Show rating",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "showDate",
                    label: "Show date",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "showSource",
                    label: "Show source",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "showReviewerPhoto",
                    label: "Show reviewer photo",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
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
            key: "booking",
            label: "Booking",
            type: "object",
            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
            fields: [
                {
                    key: "enabled",
                    label: "Accept booking requests",
                    type: "bool",
                    default: true,
                    editableBy: [ROLE_ADMIN],
                },
                {
                    key: "emailNotifications",
                    label: "Email notifications",
                    type: "object",
                    editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                    fields: [
                        {
                            key: "enabled",
                            label: "Send business notifications",
                            type: "bool",
                            default: true,
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                        },
                        {
                            key: "to",
                            label: "To recipients",
                            type: "array",
                            itemLabel: "Recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                        {
                            key: "cc",
                            label: "CC recipients",
                            type: "array",
                            itemLabel: "CC recipient",
                            default: [],
                            editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            item: {
                                type: "text",
                                label: "Email",
                                default: "",
                                editableBy: [ROLE_ADMIN, ROLE_CLIENT],
                            },
                        },
                    ],
                },
            ],
        },
        {
            key: "reports",
            label: "Reports",
            type: "object",
            editableBy: [ROLE_ADMIN],
            fields: [
                {
                    key: "analytics",
                    label: "Traffic analytics",
                    type: "object",
                    editableBy: [ROLE_ADMIN],
                    fields: [
                        {
                            key: "provider",
                            label: "Analytics provider",
                            type: "select",
                            default: "plausible",
                            editableBy: [ROLE_ADMIN],
                            options: [
                                { label: "Plausible", value: "plausible" },
                            ],
                        },
                        {
                            key: "enabled",
                            label: "Enable traffic analytics",
                            type: "bool",
                            default: false,
                            editableBy: [ROLE_ADMIN],
                        },
                        {
                            key: "siteId",
                            label: "Plausible site ID / domain",
                            type: "text",
                            default: "",
                            editableBy: [ROLE_ADMIN],
                            hint: "Usually the website domain configured in Plausible.",
                        },
                        {
                            key: "scriptEnabled",
                            label: "Inject Plausible tracking script",
                            type: "bool",
                            default: false,
                            editableBy: [ROLE_ADMIN],
                            hint: "Used later by the public website runtime. No script is injected in this phase.",
                        },
                    ],
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

function filterHiddenWebsiteSettingsFields(fields = []) {
    const filtered = [];

    for (const field of fields || []) {
        if (!field) {
            continue;
        }

        if (hiddenWebsiteSettingsFeatureKeys.has(field.key)) {
            continue;
        }

        if (field.key === "featureFlags" && field.type === "object") {
            const nextFeatureFlagFields = (field.fields || []).filter(
                (childField) => !hiddenWebsiteSettingsFeatureKeys.has(childField?.key),
            );

            if (!nextFeatureFlagFields.length) {
                continue;
            }

            filtered.push({
                ...field,
                fields: nextFeatureFlagFields,
            });
            continue;
        }

        filtered.push(field);
    }

    return filtered;
}

function shouldShowFeatureForClient(featureKey, featureFlags) {
    if (!websiteFeatureSectionKeys.has(featureKey)) {
        return true;
    }

    if (featureFlags?.[featureKey] === false) {
        return false;
    }

    return true;
}

function stripClientFeatureAvailabilityField(field) {
    if (!field || !websiteFeatureSectionKeys.has(field?.key) || field.type !== "object") {
        return field;
    }

    const nextFields = (field.fields || []).filter((childField) => childField?.key !== "enabled");
    if (!nextFields.length) {
        return null;
    }

    return {
        ...field,
        fields: nextFields,
    };
}

function filterFieldsByClientFeatureFlags(fields, featureFlags) {
    const result = [];

    for (const field of fields || []) {
        if (!websiteFeatureSectionKeys.has(field?.key)) {
            result.push(field);
            continue;
        }

        if (!shouldShowFeatureForClient(field.key, featureFlags)) {
            continue;
        }

        const nextField = stripClientFeatureAvailabilityField(field);
        if (!nextField) {
            continue;
        }

        result.push(nextField);
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

function normalizeReviewsSettings(reviewsSettings) {
    const source = isPlainObject(reviewsSettings) ? reviewsSettings : {};
    const normalizedMinRating = `${source.minRating ?? ""}`.trim();
    const minRating = ["1", "2", "3", "4", "5"].includes(normalizedMinRating) ? normalizedMinRating : "4";
    const normalizedSortOrder = `${source.sortOrder ?? ""}`.trim();
    const sortOrder = ["newest", "highestRating"].includes(normalizedSortOrder) ? normalizedSortOrder : "newest";
    const rawMaxReviews = `${source.maxReviews ?? ""}`.trim();
    const parsedMaxReviews = Number.parseInt(rawMaxReviews, 10);
    const maxReviews = Number.isFinite(parsedMaxReviews) && parsedMaxReviews > 0
        ? `${Math.min(50, parsedMaxReviews)}`
        : "6";

    return {
        ...source,
        enabled: !!source.enabled,
        googlePlaceId: typeof source.googlePlaceId === "string" ? source.googlePlaceId.trim() : "",
        reviewLink: typeof source.reviewLink === "string" ? source.reviewLink.trim() : "",
        displayEnabled: !!source.displayEnabled,
        maxReviews,
        minRating,
        sortOrder,
        showRating: typeof source.showRating === "boolean" ? source.showRating : true,
        showDate: typeof source.showDate === "boolean" ? source.showDate : true,
        showSource: typeof source.showSource === "boolean" ? source.showSource : true,
        showReviewerPhoto: typeof source.showReviewerPhoto === "boolean" ? source.showReviewerPhoto : true,
    };
}

function normalizeEmailRecipients(value, { preserveEmptyArrayRows = false } = {}) {
    const source = Array.isArray(value) ? value : [value];
    const recipients = [];
    const seen = new Set();

    for (const item of source) {
        const isArrayRowSource = Array.isArray(value);
        const asString =
            typeof item === "string"
                ? item
                : isPlainObject(item)
                    ? [item.email, item.address, item.value].find((candidate) => typeof candidate === "string")
                    : "";

        if (typeof asString !== "string") {
            continue;
        }

        const pieces = isArrayRowSource ? [asString] : asString.split(/[\n,;]+/g);
        for (const piece of pieces) {
            const normalized = piece.trim().toLowerCase();
            if (!normalized) {
                if (preserveEmptyArrayRows && isArrayRowSource) {
                    recipients.push("");
                }
                continue;
            }

            if (seen.has(normalized)) {
                continue;
            }

            recipients.push(normalized);
            seen.add(normalized);
        }
    }

    return recipients;
}

function normalizeEmailNotifications(settingsSection, { legacyDestination = "" } = {}) {
    const source = isPlainObject(settingsSection) ? settingsSection : {};
    const to = normalizeEmailRecipients(source.to, { preserveEmptyArrayRows: true });
    const legacyTo = typeof legacyDestination === "string" ? legacyDestination.trim().toLowerCase() : "";

    if (!to.length && legacyTo) {
        to.push(legacyTo);
    }

    return {
        ...source,
        enabled: !!source.enabled,
        to,
        cc: normalizeEmailRecipients(source.cc, { preserveEmptyArrayRows: true }),
    };
}

function normalizeContactFormSettings(contactFormSettings) {
    const source = isPlainObject(contactFormSettings) ? contactFormSettings : {};

    return {
        ...source,
        enabled: !!source.enabled,
        confirmationMessage:
            typeof source.confirmationMessage === "string" ? source.confirmationMessage : "",
        fields: {
            ...(isPlainObject(source.fields) ? source.fields : {}),
            phone: !!source?.fields?.phone,
        },
        emailNotifications: normalizeEmailNotifications(source.emailNotifications, {
            legacyDestination: typeof source.emailDestination === "string" ? source.emailDestination : "",
        }),
    };
}

function normalizeWhatsappSettings(whatsappSettings) {
    const source = isPlainObject(whatsappSettings) ? whatsappSettings : {};

    return {
        ...source,
        enabled: !!source.enabled,
        phone: typeof source.phone === "string" ? source.phone.trim() : "",
        defaultMessage: typeof source.defaultMessage === "string" ? source.defaultMessage : "",
        showFloatingButton: !!source.showFloatingButton,
        emailNotifications: normalizeEmailNotifications(source.emailNotifications),
    };
}

function normalizeBookingSettings(bookingSettings) {
    const source = isPlainObject(bookingSettings) ? bookingSettings : {};

    return {
        ...source,
        enabled: typeof source.enabled === "boolean" ? source.enabled : true,
        emailNotifications: normalizeEmailNotifications(source.emailNotifications, {
            legacyDestination: typeof source.emailDestination === "string" ? source.emailDestination : "",
        }),
    };
}

export function getWebsiteSettingsSchemaForRole(role = ROLE_ADMIN, rawSettings = null) {
    const normalizedRole = normalizeRole(role);
    const roleFilteredFields = filterHiddenWebsiteSettingsFields(
        filterFieldsByRole(websiteSettingsSchema.fields, normalizedRole),
    );

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

    if (isPlainObject(normalized.reviews)) {
        normalized.reviews = normalizeReviewsSettings(normalized.reviews);
    }

    if (isPlainObject(normalized.contactForm)) {
        normalized.contactForm = normalizeContactFormSettings(normalized.contactForm);
    }

    if (isPlainObject(normalized.whatsapp)) {
        normalized.whatsapp = normalizeWhatsappSettings(normalized.whatsapp);
    }

    if (isPlainObject(normalized.booking)) {
        normalized.booking = normalizeBookingSettings(normalized.booking);
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
