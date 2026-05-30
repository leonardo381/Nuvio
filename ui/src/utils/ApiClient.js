import PocketBase, { LocalAuthStore, isTokenExpired } from "pocketbase";
// ---
import { protectedFilesCollectionsCache } from "@/stores/collections";
import { setErrors } from "@/stores/errors";
import { setSuperuser } from "@/stores/superuser";
import { addErrorToast } from "@/stores/toasts";
import CommonHelper from "@/utils/CommonHelper";
import { replace } from "svelte-spa-router";
import { get } from "svelte/store";

const superuserFileTokenKey = "pb_superuser_file_token";

PocketBase.prototype.isSuperuserAuth = function () {
    return this.authStore.isValid && this.authStore.record?.collectionName === "_superusers";
};

PocketBase.prototype.getSuperuserRole = function () {
    if (!this.isSuperuserAuth()) {
        return "";
    }

    return this.authStore.record?.role === "client" ? "client" : "admin";
};

PocketBase.prototype.isAdminSuperuser = function () {
    return this.getSuperuserRole() === "admin";
};

PocketBase.prototype.isClientSuperuser = function () {
    return this.getSuperuserRole() === "client";
};

/**
 * Loads scoped backoffice website selector options.
 *
 * @param {Object} [options]
 * @param {String} [options.requestKey]
 * @returns {Promise<Array>}
 */
PocketBase.prototype.getBackofficeWebsites = async function (options = {}) {
    const requestKey = `${options?.requestKey || ""}`.trim() || "nuvio_backoffice_websites";
    const response = await this.send("/api/nuvio/backoffice/websites", {
        method: "GET",
        requestKey,
    });

    return Array.isArray(response) ? response : [];
};

/**
 * Loads scoped Reports dashboard datasets for a website.
 *
 * @param {Object} params
 * @param {String} params.websiteId
 * @param {String} [params.period]
 * @param {String} [params.requestKey]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.getReportsDashboard = async function (params = {}) {
    const websiteId = `${params?.websiteId || ""}`.trim();
    const period = `${params?.period || ""}`.trim();
    const requestKey = `${params?.requestKey || ""}`.trim() || `nuvio_reports_dashboard_${websiteId || "unknown"}`;

    return this.send("/api/nuvio/reports/dashboard", {
        method: "GET",
        query: {
            websiteId,
            period,
        },
        requestKey,
    });
};

/**
 * Loads scoped Leads dashboard datasets for a website.
 *
 * @param {Object} params
 * @param {String} params.websiteId
 * @param {String} [params.requestKey]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.getLeadsDashboard = async function (params = {}) {
    const websiteId = `${params?.websiteId || ""}`.trim();
    const requestKey = `${params?.requestKey || ""}`.trim() || `nuvio_leads_dashboard_${websiteId || "unknown"}`;

    return this.send("/api/nuvio/leads/dashboard", {
        method: "GET",
        query: {
            websiteId,
        },
        requestKey,
    });
};

/**
 * Updates a contact lead status through scoped backend endpoint.
 *
 * @param {String} id
 * @param {String} status
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateLeadContactStatus = async function (id, status, options = {}) {
    const leadId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/leads/contacts/${encodeURIComponent(leadId)}/status`, {
        method: "PATCH",
        body: {
            status: `${status || ""}`.trim(),
        },
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_lead_contact_status_${leadId || "unknown"}`,
    });
};

/**
 * Updates a WhatsApp lead status through scoped backend endpoint.
 *
 * @param {String} id
 * @param {String} status
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateLeadWhatsappStatus = async function (id, status, options = {}) {
    const leadId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/leads/whatsapp/${encodeURIComponent(leadId)}/status`, {
        method: "PATCH",
        body: {
            status: `${status || ""}`.trim(),
        },
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_lead_whatsapp_status_${leadId || "unknown"}`,
    });
};

/**
 * Updates contact lead follow-up fields through scoped backend endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {String} [payload.notes]
 * @param {String|null} [payload.lastContactedAt]
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateLeadContactFollowUp = async function (id, payload = {}, options = {}) {
    const leadId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/leads/contacts/${encodeURIComponent(leadId)}/follow-up`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_lead_contact_followup_${leadId || "unknown"}`,
    });
};

/**
 * Updates WhatsApp lead follow-up fields through scoped backend endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {String} [payload.notes]
 * @param {String|null} [payload.lastContactedAt]
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateLeadWhatsappFollowUp = async function (id, payload = {}, options = {}) {
    const leadId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/leads/whatsapp/${encodeURIComponent(leadId)}/follow-up`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_lead_whatsapp_followup_${leadId || "unknown"}`,
    });
};

/**
 * Loads scoped Newsletter dashboard datasets for a website.
 *
 * @param {Object} params
 * @param {String} params.websiteId
 * @param {String} [params.requestKey]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.getNewsletterBackofficeDashboard = async function (params = {}) {
    const websiteId = `${params?.websiteId || ""}`.trim();
    const requestKey = `${params?.requestKey || ""}`.trim() || `nuvio_newsletter_dashboard_${websiteId || "unknown"}`;

    return this.send("/api/nuvio/newsletter/backoffice/dashboard", {
        method: "GET",
        query: {
            websiteId,
        },
        requestKey,
    });
};

/**
 * Creates a subscriber through scoped Newsletter backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createNewsletterSubscriber = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/newsletter/backoffice/subscribers", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_subscriber_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates a subscriber through scoped Newsletter backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateNewsletterSubscriber = async function (id, payload = {}, options = {}) {
    const subscriberId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/subscribers/${encodeURIComponent(subscriberId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_subscriber_update_${subscriberId || "unknown"}`,
    });
};

/**
 * Deletes a subscriber through scoped Newsletter backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.deleteNewsletterSubscriber = async function (id, options = {}) {
    const subscriberId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/subscribers/${encodeURIComponent(subscriberId)}`, {
        method: "DELETE",
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_subscriber_delete_${subscriberId || "unknown"}`,
    });
};

/**
 * Creates a subscriber group through scoped Newsletter backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createNewsletterGroup = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/newsletter/backoffice/groups", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_group_create_${websiteId || "unknown"}`,
    });
};

/**
 * Creates a campaign through scoped Newsletter backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createNewsletterCampaign = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/newsletter/backoffice/campaigns", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_campaign_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates a campaign through scoped Newsletter backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateNewsletterCampaign = async function (id, payload = {}, options = {}) {
    const campaignId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/campaigns/${encodeURIComponent(campaignId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_campaign_update_${campaignId || "unknown"}`,
    });
};

/**
 * Deletes a campaign through scoped Newsletter backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.deleteNewsletterCampaign = async function (id, options = {}) {
    const campaignId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/campaigns/${encodeURIComponent(campaignId)}`, {
        method: "DELETE",
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_campaign_delete_${campaignId || "unknown"}`,
    });
};

/**
 * Duplicates a campaign as draft through scoped Newsletter backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.duplicateNewsletterCampaign = async function (id, options = {}) {
    const campaignId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/campaigns/${encodeURIComponent(campaignId)}/duplicate`, {
        method: "POST",
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_campaign_duplicate_${campaignId || "unknown"}`,
    });
};

/**
 * Resends a newsletter confirmation through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.inviteNewsletterSubscriber = async function (id, options = {}) {
    const subscriberId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/subscribers/${encodeURIComponent(subscriberId)}/invite`, {
        method: "POST",
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_subscriber_invite_${subscriberId || "unknown"}`,
    });
};

/**
 * Sends a newsletter campaign through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.sendNewsletterCampaign = async function (id, options = {}) {
    const campaignId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/newsletter/backoffice/campaigns/${encodeURIComponent(campaignId)}/send`, {
        method: "POST",
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_newsletter_campaign_send_${campaignId || "unknown"}`,
    });
};

/**
 * Loads scoped Booking dashboard datasets for a website.
 *
 * @param {Object} params
 * @param {String} params.websiteId
 * @param {String} [params.requestKey]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.getBookingBackofficeDashboard = async function (params = {}) {
    const websiteId = `${params?.websiteId || ""}`.trim();
    const requestKey = `${params?.requestKey || ""}`.trim() || `nuvio_booking_dashboard_${websiteId || "unknown"}`;

    return this.send("/api/nuvio/booking/backoffice/dashboard", {
        method: "GET",
        query: {
            websiteId,
        },
        requestKey,
    });
};

/**
 * Creates a booking appointment through scoped backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createBookingBackofficeAppointment = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/booking/backoffice/appointments", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_appointment_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates appointment status through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeAppointmentStatus = async function (id, payload = {}, options = {}) {
    const appointmentId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/appointments/${encodeURIComponent(appointmentId)}/status`, {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_appointment_status_${appointmentId || "unknown"}`,
    });
};

/**
 * Reschedules an appointment through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.rescheduleBookingBackofficeAppointment = async function (id, payload = {}, options = {}) {
    const appointmentId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/appointments/${encodeURIComponent(appointmentId)}/reschedule`, {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_appointment_reschedule_${appointmentId || "unknown"}`,
    });
};

/**
 * Updates appointment internal notes through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeAppointmentInternalNotes = async function (id, payload = {}, options = {}) {
    const appointmentId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/appointments/${encodeURIComponent(appointmentId)}/internal-notes`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_appointment_internal_notes_${appointmentId || "unknown"}`,
    });
};

/**
 * Updates appointment archive state through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeAppointmentArchive = async function (id, payload = {}, options = {}) {
    const appointmentId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/appointments/${encodeURIComponent(appointmentId)}/archive`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_appointment_archive_${appointmentId || "unknown"}`,
    });
};

/**
 * Creates a booking service through scoped backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createBookingBackofficeService = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/booking/backoffice/services", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_service_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates a booking service through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeService = async function (id, payload = {}, options = {}) {
    const serviceId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/services/${encodeURIComponent(serviceId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_service_update_${serviceId || "unknown"}`,
    });
};

/**
 * Creates a booking availability window through scoped backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createBookingBackofficeAvailability = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/booking/backoffice/availability", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_availability_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates a booking availability window through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeAvailability = async function (id, payload = {}, options = {}) {
    const availabilityId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/availability/${encodeURIComponent(availabilityId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_availability_update_${availabilityId || "unknown"}`,
    });
};

/**
 * Creates a booking exception through scoped backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.createBookingBackofficeException = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/booking/backoffice/exceptions", {
        method: "POST",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_exception_create_${websiteId || "unknown"}`,
    });
};

/**
 * Updates a booking exception through scoped backoffice endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeException = async function (id, payload = {}, options = {}) {
    const exceptionId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/booking/backoffice/exceptions/${encodeURIComponent(exceptionId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_exception_update_${exceptionId || "unknown"}`,
    });
};

/**
 * Updates booking rules through scoped backoffice endpoint.
 *
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateBookingBackofficeRules = async function (payload = {}, options = {}) {
    const websiteId = `${payload?.websiteId || ""}`.trim();
    return this.send("/api/nuvio/booking/backoffice/settings/rules", {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_booking_backoffice_rules_update_${websiteId || "unknown"}`,
    });
};

/**
 * Loads scoped CMS dashboard data for a website/page context.
 *
 * @param {Object} params
 * @param {String} params.websiteId
 * @param {String} [params.pageId]
 * @param {String} [params.requestKey]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.getCMSDashboard = async function (params = {}) {
    const websiteId = `${params?.websiteId || ""}`.trim();
    const pageId = `${params?.pageId || ""}`.trim();
    const requestKey = `${params?.requestKey || ""}`.trim() || `nuvio_cms_dashboard_${websiteId || "unknown"}_${pageId || "none"}`;

    const query = {
        websiteId,
    };

    if (pageId) {
        query.pageId = pageId;
    }

    return this.send("/api/nuvio/cms/dashboard", {
        method: "GET",
        query,
        requestKey,
    });
};

/**
 * Updates website identity/global SEO fields through scoped CMS endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateCMSWebsiteIdentity = async function (id, payload = {}, options = {}) {
    const websiteId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/cms/websites/${encodeURIComponent(websiteId)}/identity`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_cms_website_identity_${websiteId || "unknown"}`,
    });
};

/**
 * Updates website settings fields through scoped CMS endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateCMSWebsiteSettings = async function (id, payload = {}, options = {}) {
    const websiteId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/cms/websites/${encodeURIComponent(websiteId)}/settings`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_cms_website_settings_${websiteId || "unknown"}`,
    });
};

/**
 * Updates page SEO fields through scoped CMS endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateCMSPageSEO = async function (id, payload = {}, options = {}) {
    const pageId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/cms/pages/${encodeURIComponent(pageId)}/seo`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_cms_page_seo_${pageId || "unknown"}`,
    });
};

/**
 * Updates block props/translations through scoped CMS endpoint.
 *
 * @param {String} id
 * @param {Object} payload
 * @param {Object} [options]
 * @returns {Promise<Object>}
 */
PocketBase.prototype.updateCMSBlock = async function (id, payload = {}, options = {}) {
    const blockId = `${id || ""}`.trim();
    return this.send(`/api/nuvio/cms/blocks/${encodeURIComponent(blockId)}`, {
        method: "PATCH",
        body: payload,
        requestKey: `${options?.requestKey || ""}`.trim() || `nuvio_cms_block_${blockId || "unknown"}`,
    });
};

/**
 * Clears the authorized state and redirects to the login page.
 *
 * @param {Boolean} [redirect] Whether to redirect to the login page.
 */
PocketBase.prototype.logout = function (redirect = true) {
    this.authStore.clear();

    if (redirect) {
        replace("/login");
    }
};

/**
 * Generic API error response handler.
 *
 * @param  {Error}   err        The API error itself.
 * @param  {Boolean} notify     Whether to add a toast notification.
 * @param  {String}  defaultMsg Default toast notification message if the error doesn't have one.
 */
PocketBase.prototype.error = function (err, notify = true, defaultMsg = "") {
    if (!err || !(err instanceof Error) || err.isAbort) {
        return;
    }

    const statusCode = (err?.status << 0) || 400;
    const responseData = err?.data || {};
    const msg = responseData.message || err.message || defaultMsg;

    // add toast error notification
    if (notify && msg) {
        addErrorToast(msg);
    }

    // populate form field errors
    if (!CommonHelper.isEmpty(responseData.data)) {
        setErrors(responseData.data);
    }

    // unauthorized
    if (statusCode === 401) {
        this.cancelAllRequests();
        return this.logout();
    }

    // forbidden
    if (statusCode === 403) {
        this.cancelAllRequests();
        return replace("/");
    }
};

/**
 * @return {Promise<String>}
 */
PocketBase.prototype.getSuperuserFileToken = async function (collectionId = "") {
    let needToken = true;

    if (collectionId) {
        const protectedCollections = get(protectedFilesCollectionsCache);
        needToken = typeof protectedCollections[collectionId] !== "undefined"
            ? protectedCollections[collectionId]
            : true;
    }

    if (!needToken) {
        return "";
    }

    let token = localStorage.getItem(superuserFileTokenKey) || "";

    // request a new token only if the previous one is missing or will expire soon
    if (!token || isTokenExpired(token, 10)) {
        // remove previously stored token (if any)
        token && localStorage.removeItem(superuserFileTokenKey);

        if (!this._superuserFileTokenRequest) {
            this._superuserFileTokenRequest = this.files.getToken();
        }

        token = await this._superuserFileTokenRequest;
        localStorage.setItem(superuserFileTokenKey, token);
        this._superuserFileTokenRequest = null;
    }

    return token;
}

// Custom auth store to sync the svelte superuser store state with the authorized superuser instance.
class AppAuthStore extends LocalAuthStore {
    /**
     * @inheritdoc
     */
    constructor(storageKey = "__pb_superuser_auth__") {
        super(storageKey);

        this.save(this.token, this.record);
    }

    /**
     * @inheritdoc
     */
    save(token, record) {
        super.save(token, record);

        if (record?.collectionName == "_superusers") {
            setSuperuser(record);
        }
    }

    /**
     * @inheritdoc
     */
    clear() {
        super.clear();

        setSuperuser(null);
    }
}

const pb = new PocketBase(import.meta.env.PB_BACKEND_URL, new AppAuthStore());

pb._authReadyPromise = Promise.resolve();

if (pb.authStore.isValid) {
    pb._authReadyPromise = pb.collection(pb.authStore.record.collectionName || "_superusers")
        .authRefresh()
        .catch((err) => {
            console.warn("Failed to refresh the existing auth token:", err);

            // clear the store only on invalidated/expired token
            const status = err?.status << 0;
            if (status == 401 || status == 403) {
                pb.authStore.clear();
            }
        });
}

PocketBase.prototype.whenAuthReady = function () {
    return this._authReadyPromise || Promise.resolve();
};

export default pb;
