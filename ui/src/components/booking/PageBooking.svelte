<script>
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import { pageTitle } from "@/stores/app";
    import { collections, isCollectionsLoading, loadCollections } from "@/stores/collections";
    import { addErrorToast, addSuccessToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";

    // NUVIO CUSTOM START: Booking MVP Phase 2 purpose-built backoffice module.
    $pageTitle = "Booking";

    const bookingTabs = [
        { key: "appointments", label: "Appointments", icon: "ri-calendar-check-line" },
        { key: "services", label: "Services", icon: "ri-briefcase-4-line" },
        { key: "availability", label: "Availability", icon: "ri-time-line" },
    ];

    const appointmentStatusOptions = [
        { key: "all", label: "All" },
        { key: "pending", label: "Pending" },
        { key: "confirmed", label: "Confirmed" },
        { key: "cancelled", label: "Cancelled" },
    ];

    const appointmentDateOptions = [
        { key: "all", label: "All" },
        { key: "today", label: "Today" },
        { key: "thisWeek", label: "This week" },
        { key: "thisMonth", label: "This month" },
        { key: "upcoming", label: "Upcoming" },
    ];

    const appointmentSortOptions = [
        { key: "newest", label: "Newest" },
        { key: "oldest", label: "Oldest" },
        { key: "upcoming", label: "Upcoming first" },
    ];

    const availabilityDays = [
        { key: "mon", label: "Monday" },
        { key: "tue", label: "Tuesday" },
        { key: "wed", label: "Wednesday" },
        { key: "thu", label: "Thursday" },
        { key: "fri", label: "Friday" },
        { key: "sat", label: "Saturday" },
        { key: "sun", label: "Sunday" },
    ];

    const appointmentStatusFieldAliases = ["status"];
    const appointmentCustomerNotesFieldAliases = ["notes", "note"];
    const appointmentInternalNotesFieldAliases = ["internalNotes", "internal_notes"];
    const bookingDatePattern = /^\d{4}-\d{2}-\d{2}$/;
    const bookingTimePattern = /^([01]\d|2[0-3]):[0-5]\d$/;
    const bookingEmailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    let activeTab = "appointments";

    let websites = [];
    let selectedWebsiteId = "";

    let servicesRecords = [];
    let availabilityRecords = [];
    let appointmentRecords = [];

    let isLoadingWebsites = false;
    let isLoadingBookingData = false;
    let bookingLoadError = "";

    let lastWebsitesCollectionId = "";
    let lastBookingDataKey = "";

    let appointmentStatusFilter = "all";
    let appointmentDateFilter = "all";
    let appointmentServiceFilter = "all";
    let appointmentSearch = "";
    let appointmentSort = "newest";
    let selectedAppointmentId = "";
    let isUpdatingAppointmentStatus = false;
    let updatingAppointmentId = "";
    let appointmentInternalNotesDraft = "";
    let appointmentInternalNotesDraftId = "";
    let isSavingAppointmentInternalNotes = false;
    let isManualAppointmentPanelOpen = false;
    let isCreatingManualAppointment = false;
    let isLoadingManualAppointmentSlots = false;
    let manualAppointmentSlotsError = "";
    let manualAppointmentFormError = "";
    let manualAppointmentAvailableSlots = [];
    let lastManualSlotsQueryKey = "";
    let manualAppointmentForm = createDefaultManualAppointmentForm();

    let selectedServiceId = "";
    let isSavingService = false;
    let serviceSearch = "";
    let serviceStatusFilter = "all";
    let serviceForm = {
        id: "",
        name: "",
        durationMinutes: "30",
        active: true,
    };
    let serviceFormError = "";

    let availabilityRows = createDefaultAvailabilityRows();
    let isSavingAvailability = {};
    let isSavingAllAvailability = false;
    let slotPreviewServiceId = "";
    let slotPreviewDate = "";
    let slotPreviewSlots = [];
    let slotPreviewError = "";
    let isLoadingSlotPreview = false;
    let lastSlotPreviewQueryKey = "";

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: bookingServicesCollection = resolveCollectionByAliases(["bookingservices"]);
    $: bookingAvailabilityCollection = resolveCollectionByAliases(["bookingavailability"]);
    $: appointmentsCollection = resolveCollectionByAliases(["appointments"]);

    $: hasBookingCollections = !!bookingServicesCollection?.id
        && !!bookingAvailabilityCollection?.id
        && !!appointmentsCollection?.id;

    $: appointmentStatusFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentStatusFieldAliases) || "status";
    $: appointmentCustomerNotesFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentCustomerNotesFieldAliases) || "notes";
    $: appointmentInternalNotesFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentInternalNotesFieldAliases) || "internalNotes";

    $: if (!websitesCollection?.id) {
        websites = [];
        selectedWebsiteId = "";
        lastWebsitesCollectionId = "";
    } else if (websitesCollection.id !== lastWebsitesCollectionId) {
        lastWebsitesCollectionId = websitesCollection.id;
        loadWebsites();
    }

    $: if (websites.length && !websites.find((website) => website.id === selectedWebsiteId)) {
        selectedWebsiteId = websites[0].id;
    }

    $: bookingDataKey = [
        selectedWebsiteId,
        bookingServicesCollection?.id || "",
        bookingAvailabilityCollection?.id || "",
        appointmentsCollection?.id || "",
    ].join(":");

    $: if (hasBookingCollections && selectedWebsiteId && bookingDataKey !== lastBookingDataKey) {
        lastBookingDataKey = bookingDataKey;
        loadBookingData();
    }

    $: serviceLabelById = new Map(
        servicesRecords.map((service) => [service.id, normalizeString(service?.name) || "Untitled service"]),
    );

    $: selectedWebsite = websites.find((website) => website.id === selectedWebsiteId) || null;
    $: selectedWebsiteBookingSettings = resolveWebsiteBookingSettings(selectedWebsite?.settings);

    $: normalizedAppointments = appointmentRecords.map((record) => normalizeAppointment(record));

    $: pendingAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "pending").length;
    $: confirmedAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "confirmed").length;
    $: thisWeekAppointmentsCount = normalizedAppointments.filter((appointment) => isInCurrentWeek(appointment.date)).length;
    $: activeServicesCount = servicesRecords.filter((service) => !!service?.active).length;
    $: activeAvailabilityDaysCount = availabilityRows.filter((row) => !!row?.active).length;
    $: bookingReadinessWarnings = buildBookingReadinessWarnings({
        activeServicesCount,
        activeAvailabilityDaysCount,
        bookingSettings: selectedWebsiteBookingSettings,
    });
    $: bookingReadinessSuggestions = buildBookingReadinessSuggestions({
        activeServicesCount,
        activeAvailabilityDaysCount,
        bookingSettings: selectedWebsiteBookingSettings,
    });
    $: bookingReadinessState = resolveBookingReadinessState(bookingReadinessWarnings.length);
    $: availabilityValidationIssuesCount = availabilityRows.filter((row) => !!row?.active && !!validateAvailabilityRow(row)).length;
    $: availabilityHealthWarnings = buildAvailabilityHealthWarnings({
        bookingWarnings: bookingReadinessWarnings,
        availabilityValidationIssuesCount,
        dirtyAvailabilityRowsCount,
    });
    $: availabilityHealthSuggestions = buildAvailabilityHealthSuggestions({
        bookingSuggestions: bookingReadinessSuggestions,
        dirtyAvailabilityRowsCount,
        availabilityValidationIssuesCount,
    });
    $: availabilityHealthState = resolveBookingReadinessState(availabilityHealthWarnings.length);
    $: normalizedServiceSearch = normalizeLower(serviceSearch);
    $: filteredServices = servicesRecords.filter((service) => {
        const isActive = !!service?.active;

        if (serviceStatusFilter === "active" && !isActive) {
            return false;
        }

        if (serviceStatusFilter === "inactive" && isActive) {
            return false;
        }

        if (normalizedServiceSearch) {
            const duration = `${service?.durationMinutes || ""}`.trim();
            const searchable = [
                normalizeString(service?.name),
                duration,
                duration ? `${duration} minutes` : "",
            ]
                .filter(Boolean)
                .join(" ")
                .toLowerCase();

            if (!searchable.includes(normalizedServiceSearch)) {
                return false;
            }
        }

        return true;
    });
    $: dirtyAvailabilityRowsCount = availabilityRows.filter((row) => !!row?.dirty).length;
    $: hasSavingAvailabilityRows = Object.values(isSavingAvailability || {}).some((value) => !!value);

    $: appointmentServiceOptions = servicesRecords
        .map((service) => ({
            id: normalizeString(service?.id),
            label: normalizeString(service?.name) || "Untitled service",
        }))
        .filter((service) => service.id)
        .sort((a, b) => a.label.localeCompare(b.label));
    $: manualAppointmentServiceOptions = appointmentServiceOptions.filter((service) => {
        const record = servicesRecords.find((item) => normalizeString(item?.id) === service.id);
        return !!record?.active;
    });
    $: slotPreviewServiceOptions = manualAppointmentServiceOptions;
    $: manualAppointmentSlotsQueryKey = isManualAppointmentPanelOpen
        ? `${selectedWebsiteId}:${manualAppointmentForm.serviceId}:${manualAppointmentForm.date}`
        : "";
    $: slotPreviewQueryKey = activeTab === "availability"
        ? `${selectedWebsiteId}:${slotPreviewServiceId}:${slotPreviewDate}`
        : "";
    $: if (
        isManualAppointmentPanelOpen
        && manualAppointmentServiceOptions.length
        && !manualAppointmentServiceOptions.some((service) => service.id === manualAppointmentForm.serviceId)
    ) {
        manualAppointmentForm = {
            ...manualAppointmentForm,
            serviceId: manualAppointmentServiceOptions[0].id,
            time: "",
        };
    }
    $: if (manualAppointmentSlotsQueryKey !== lastManualSlotsQueryKey) {
        lastManualSlotsQueryKey = manualAppointmentSlotsQueryKey;
        loadManualAppointmentSlots();
    }
    $: if (activeTab === "availability" && !slotPreviewDate) {
        slotPreviewDate = getDefaultSlotPreviewDate();
    }
    $: if (
        activeTab === "availability"
        && slotPreviewServiceOptions.length
        && !slotPreviewServiceOptions.some((service) => service.id === slotPreviewServiceId)
    ) {
        slotPreviewServiceId = slotPreviewServiceOptions[0].id;
    }
    $: if (activeTab === "availability" && !slotPreviewServiceOptions.length) {
        slotPreviewServiceId = "";
        slotPreviewSlots = [];
        slotPreviewError = "";
    }
    $: if (slotPreviewQueryKey !== lastSlotPreviewQueryKey) {
        lastSlotPreviewQueryKey = slotPreviewQueryKey;
        loadSlotPreview();
    }

    $: normalizedAppointmentSearch = normalizeLower(appointmentSearch);

    $: filteredAppointments = sortAppointments(
        normalizedAppointments.filter((appointment) => {
            if (appointmentStatusFilter !== "all" && appointment.statusKey !== appointmentStatusFilter) {
                return false;
            }

            if (appointmentServiceFilter !== "all" && appointment.serviceId !== appointmentServiceFilter) {
                return false;
            }

            if (!matchAppointmentDateFilter(appointment, appointmentDateFilter)) {
                return false;
            }

            if (normalizedAppointmentSearch) {
                const searchable = [
                    appointment.name,
                    appointment.email,
                    appointment.phone,
                    appointment.serviceLabel,
                    appointment.customerNotes,
                    appointment.internalNotes,
                ]
                    .filter(Boolean)
                    .join(" ")
                    .toLowerCase();

                if (!searchable.includes(normalizedAppointmentSearch)) {
                    return false;
                }
            }

            return true;
        }),
        appointmentSort,
    );

    $: if (filteredAppointments.length) {
        const hasSelected = selectedAppointmentId
            && filteredAppointments.some((appointment) => appointment.id === selectedAppointmentId);

        if (!hasSelected) {
            selectedAppointmentId = filteredAppointments[0].id;
        }
    } else if (selectedAppointmentId) {
        selectedAppointmentId = "";
    }

    $: selectedAppointment = filteredAppointments.find((appointment) => appointment.id === selectedAppointmentId) || null;
    $: selectedAppointmentStatusKey = selectedAppointment?.statusKey || "";

    $: canSetSelectedAppointmentPending = !!selectedAppointment
        && selectedAppointmentStatusKey !== "pending";
    $: canSetSelectedAppointmentConfirmed = !!selectedAppointment
        && selectedAppointmentStatusKey === "pending";
    $: canSetSelectedAppointmentCancelled = !!selectedAppointment
        && (selectedAppointmentStatusKey === "pending" || selectedAppointmentStatusKey === "confirmed");

    $: if (selectedAppointment?.id) {
        if (appointmentInternalNotesDraftId !== selectedAppointment.id) {
            appointmentInternalNotesDraftId = selectedAppointment.id;
            appointmentInternalNotesDraft = selectedAppointment.internalNotes || "";
        }
    } else if (appointmentInternalNotesDraftId || appointmentInternalNotesDraft) {
        appointmentInternalNotesDraftId = "";
        appointmentInternalNotesDraft = "";
    }

    $: selectedService = servicesRecords.find((service) => service.id === selectedServiceId) || null;

    $: if (servicesRecords.length && !selectedServiceId && !serviceForm.id) {
        selectedServiceId = servicesRecords[0].id;
    }

    $: if (selectedService) {
        if (serviceForm.id !== selectedService.id) {
            setServiceFormFromRecord(selectedService);
        }
    }

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
    }

    function isPlainObject(value) {
        return !!value && typeof value === "object" && !Array.isArray(value);
    }

    function parseSettingsObject(rawSettings) {
        if (isPlainObject(rawSettings)) {
            return rawSettings;
        }

        if (typeof rawSettings === "string") {
            const normalized = rawSettings.trim();
            if (!normalized) {
                return {};
            }

            try {
                const parsed = JSON.parse(normalized);
                return isPlainObject(parsed) ? parsed : {};
            } catch (_) {
                return {};
            }
        }

        return {};
    }

    function readObject(value) {
        return isPlainObject(value) ? value : {};
    }

    function readBoolean(value, fallback = false) {
        if (typeof value === "boolean") {
            return value;
        }

        if (typeof value === "string") {
            const normalized = value.trim().toLowerCase();
            if (normalized === "true") {
                return true;
            }
            if (normalized === "false") {
                return false;
            }
        }

        return fallback;
    }

    function normalizeEmailList(rawValue) {
        const source = Array.isArray(rawValue) ? rawValue : [rawValue];
        const normalized = [];
        const seen = new Set();

        const appendCandidate = (candidateValue) => {
            const textValue = normalizeLower(candidateValue);
            if (!textValue) {
                return;
            }

            const chunks = textValue
                .replace(/[\n;]+/g, ",")
                .split(",")
                .map((chunk) => chunk.trim())
                .filter(Boolean);

            for (const chunk of chunks) {
                if (!seen.has(chunk)) {
                    seen.add(chunk);
                    normalized.push(chunk);
                }
            }
        };

        for (const entry of source) {
            if (typeof entry === "string") {
                appendCandidate(entry);
                continue;
            }

            if (isPlainObject(entry)) {
                appendCandidate(entry.email);
                appendCandidate(entry.address);
                appendCandidate(entry.value);
            }
        }

        return normalized;
    }

    function parseEmailNotifications(rawNotifications, legacyDestination = "") {
        const hasExplicitConfig = isPlainObject(rawNotifications);
        const source = readObject(rawNotifications);
        const toRecipients = normalizeEmailList(source.to);
        const legacyRecipient = normalizeLower(legacyDestination);
        let enabled = readBoolean(source.enabled, false);

        if (!toRecipients.length && legacyRecipient) {
            toRecipients.push(legacyRecipient);
            if (!hasExplicitConfig) {
                enabled = true;
            }
        }

        return {
            enabled,
            to: toRecipients,
            cc: normalizeEmailList(source.cc),
        };
    }

    function resolveWebsiteBookingSettings(rawSettings) {
        const settings = parseSettingsObject(rawSettings);
        const featureFlags = readObject(settings.featureFlags);
        const booking = readObject(settings.booking);
        const contactForm = readObject(settings.contactForm);

        const bookingFeatureAvailable = readBoolean(featureFlags.booking, true);
        const bookingEnabled = readBoolean(booking.enabled, true);

        const bookingNotifications = parseEmailNotifications(
            booking.emailNotifications,
            normalizeString(booking.emailDestination),
        );
        const contactNotifications = parseEmailNotifications(
            contactForm.emailNotifications,
            normalizeString(contactForm.emailDestination),
        );

        const effectiveNotifications = {
            enabled: bookingNotifications.enabled,
            to: [...bookingNotifications.to],
            cc: [...bookingNotifications.cc],
        };

        let usingContactFormFallback = false;
        if (!effectiveNotifications.to.length && contactNotifications.to.length) {
            usingContactFormFallback = true;
            effectiveNotifications.to = [...contactNotifications.to];
            effectiveNotifications.cc = [...contactNotifications.cc];
            if (contactNotifications.enabled) {
                effectiveNotifications.enabled = true;
            }
        }

        const hasBusinessRecipients = effectiveNotifications.to.length > 0;
        const businessNotificationsReady = effectiveNotifications.enabled && hasBusinessRecipients;

        return {
            featureAvailable: bookingFeatureAvailable,
            enabled: bookingEnabled,
            bookingNotifications,
            contactNotifications,
            effectiveNotifications,
            usingContactFormFallback,
            businessNotificationsReady,
        };
    }

    function getDefaultSlotPreviewDate() {
        const date = new Date();
        date.setDate(date.getDate() + 1);
        const year = date.getFullYear();
        const month = `${date.getMonth() + 1}`.padStart(2, "0");
        const day = `${date.getDate()}`.padStart(2, "0");
        return `${year}-${month}-${day}`;
    }

    function setSlotPreviewService(serviceId) {
        slotPreviewServiceId = normalizeString(serviceId);
        slotPreviewSlots = [];
        slotPreviewError = "";
    }

    function setSlotPreviewDate(dateValue) {
        slotPreviewDate = normalizeString(dateValue);
        slotPreviewSlots = [];
        slotPreviewError = "";
    }

    function isValidEmail(value) {
        return bookingEmailPattern.test(normalizeString(value).toLowerCase());
    }

    function createDefaultManualAppointmentForm() {
        return {
            serviceId: "",
            date: "",
            time: "",
            name: "",
            email: "",
            phone: "",
            notes: "",
            internalNotes: "",
            status: "confirmed",
        };
    }

    function resetManualAppointmentForm() {
        manualAppointmentForm = createDefaultManualAppointmentForm();
        manualAppointmentFormError = "";
        manualAppointmentSlotsError = "";
        manualAppointmentAvailableSlots = [];
        lastManualSlotsQueryKey = "";
    }

    function openManualAppointmentPanel() {
        if (!selectedWebsiteId) {
            addErrorToast("Select a website before creating an appointment.");
            return;
        }

        resetManualAppointmentForm();
        if (manualAppointmentServiceOptions.length) {
            manualAppointmentForm = {
                ...manualAppointmentForm,
                serviceId: manualAppointmentServiceOptions[0].id,
            };
        }
        isManualAppointmentPanelOpen = true;
    }

    function closeManualAppointmentPanel() {
        isManualAppointmentPanelOpen = false;
        isLoadingManualAppointmentSlots = false;
        isCreatingManualAppointment = false;
        resetManualAppointmentForm();
    }

    function setManualAppointmentService(serviceId) {
        manualAppointmentForm = {
            ...manualAppointmentForm,
            serviceId: normalizeString(serviceId),
            time: "",
        };
        manualAppointmentFormError = "";
        manualAppointmentSlotsError = "";
    }

    function setManualAppointmentDate(dateValue) {
        manualAppointmentForm = {
            ...manualAppointmentForm,
            date: normalizeString(dateValue),
            time: "",
        };
        manualAppointmentFormError = "";
        manualAppointmentSlotsError = "";
    }

    function selectManualAppointmentSlot(slot) {
        manualAppointmentForm = {
            ...manualAppointmentForm,
            time: normalizeString(slot),
        };
        manualAppointmentFormError = "";
    }

    function normalizeRelationId(value) {
        if (Array.isArray(value)) {
            return normalizeString(value[0]);
        }
        return normalizeString(value);
    }

    function normalizeStatus(value) {
        const normalized = normalizeLower(value);
        if (!normalized) {
            return "pending";
        }
        if (["pending", "confirmed", "cancelled"].includes(normalized)) {
            return normalized;
        }
        return "pending";
    }

    function resolveCollectionByAliases(aliases = []) {
        const normalizedAliases = aliases.map((alias) => normalizeLower(alias)).filter(Boolean);

        for (const alias of normalizedAliases) {
            const match = $collections.find((collection) => normalizeLower(collection?.name) === alias);
            if (match) {
                return match;
            }
        }

        return null;
    }

    function resolveCollectionFieldNameByAliases(collection, aliases = []) {
        const fields = Array.isArray(collection?.fields) ? collection.fields : [];
        const normalizedAliases = aliases.map((alias) => normalizeLower(alias)).filter(Boolean);

        for (const field of fields) {
            const fieldName = normalizeString(field?.name);
            if (fieldName && normalizedAliases.includes(normalizeLower(fieldName))) {
                return fieldName;
            }
        }

        return "";
    }

    function resolveWebsitesSort(collection) {
        const preferredSortFields = ["title", "name", "slug"];
        const availableFields = new Set(
            CommonHelper.getAllCollectionIdentifiers(collection).map((field) => normalizeLower(field)),
        );
        const validSortFields = preferredSortFields.filter((field) => availableFields.has(field));

        if (!validSortFields.length) {
            return "+id";
        }

        return validSortFields.map((field) => `+${field}`).join(",");
    }

    function resolveWebsiteLabel(website) {
        return (
            `${CommonHelper.displayValue(website || {}, ["title", "name", "slug"]) || ""}`.trim()
            || website?.id
            || ""
        );
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
                sort: resolveWebsitesSort(websitesCollection),
                requestKey: "nuvio_booking_websites",
            });

            if (!websites.length) {
                selectedWebsiteId = "";
            } else if (!websites.find((website) => website.id === selectedWebsiteId)) {
                selectedWebsiteId = websites[0].id;
            }
        } catch (err) {
            websites = [];
            selectedWebsiteId = "";
            ApiClient.error(err, false);
            addErrorToast("Unable to load websites right now.");
        }

        isLoadingWebsites = false;
    }

    async function loadBookingData() {
        if (!selectedWebsiteId || !hasBookingCollections) {
            servicesRecords = [];
            availabilityRecords = [];
            appointmentRecords = [];
            availabilityRows = createDefaultAvailabilityRows();
            return;
        }

        isLoadingBookingData = true;
        bookingLoadError = "";

        try {
            const filter = `website="${selectedWebsiteId}"`;
            const [services, availability, appointments] = await Promise.all([
                ApiClient.collection(bookingServicesCollection.id).getFullList({
                    filter,
                    sort: "+name",
                    requestKey: `nuvio_booking_services_${selectedWebsiteId}`,
                }),
                ApiClient.collection(bookingAvailabilityCollection.id).getFullList({
                    filter,
                    sort: "-created",
                    requestKey: `nuvio_booking_availability_${selectedWebsiteId}`,
                }),
                ApiClient.collection(appointmentsCollection.id).getFullList({
                    filter,
                    sort: "-created",
                    expand: "service",
                    requestKey: `nuvio_booking_appointments_${selectedWebsiteId}`,
                }),
            ]);

            servicesRecords = services;
            availabilityRecords = availability;
            appointmentRecords = appointments;
            availabilityRows = createAvailabilityRowsFromRecords(availabilityRecords);
            isSavingAvailability = {};

            const hasSelectedService = !!selectedServiceId
                && services.some((service) => service.id === selectedServiceId);
            const nextSelectedServiceId = hasSelectedService ? selectedServiceId : (services[0]?.id || "");
            if (nextSelectedServiceId !== selectedServiceId) {
                selectedServiceId = nextSelectedServiceId;
            }

            if (!nextSelectedServiceId) {
                resetServiceForm();
            }
        } catch (err) {
            ApiClient.error(err, false);
            bookingLoadError = "Unable to load booking data right now. Please refresh and try again.";
            addErrorToast("Unable to load booking data right now.");
            servicesRecords = [];
            availabilityRecords = [];
            appointmentRecords = [];
            availabilityRows = createDefaultAvailabilityRows();
        }

        isLoadingBookingData = false;
    }

    async function loadSlotPreview() {
        if (activeTab !== "availability") {
            return;
        }

        const serviceId = normalizeString(slotPreviewServiceId);
        const dateValue = normalizeString(slotPreviewDate);
        const hasValidService = slotPreviewServiceOptions.some((service) => service.id === serviceId);
        const requestScopeKey = `${selectedWebsiteId}:${serviceId}:${dateValue}`;

        if (!selectedWebsiteId || !serviceId || !hasValidService || !bookingDatePattern.test(dateValue)) {
            slotPreviewSlots = [];
            slotPreviewError = "";
            isLoadingSlotPreview = false;
            return;
        }

        isLoadingSlotPreview = true;
        slotPreviewError = "";

        try {
            const query = new URLSearchParams({
                websiteId: selectedWebsiteId,
                serviceId,
                date: dateValue,
            });

            const response = await ApiClient.send(`/api/nuvio/booking/slots?${query.toString()}`, {
                method: "GET",
                requestKey: `nuvio_booking_slot_preview_${selectedWebsiteId}_${serviceId}_${dateValue}`,
            });

            const slots = Array.isArray(response?.slots)
                ? response.slots.map((slot) => normalizeString(slot)).filter((slot) => bookingTimePattern.test(slot))
                : [];

            if (slotPreviewQueryKey !== requestScopeKey) {
                return;
            }

            slotPreviewSlots = slots;
        } catch (err) {
            ApiClient.error(err, false);
            if (slotPreviewQueryKey !== requestScopeKey) {
                return;
            }

            slotPreviewSlots = [];
            slotPreviewError = "Unable to load available times right now.";
        } finally {
            if (slotPreviewQueryKey === requestScopeKey) {
                isLoadingSlotPreview = false;
            }
        }
    }

    async function loadManualAppointmentSlots() {
        if (!isManualAppointmentPanelOpen) {
            manualAppointmentAvailableSlots = [];
            manualAppointmentSlotsError = "";
            return;
        }

        const serviceId = normalizeString(manualAppointmentForm.serviceId);
        const dateValue = normalizeString(manualAppointmentForm.date);
        const requestScopeKey = `${selectedWebsiteId}:${serviceId}:${dateValue}`;

        if (!selectedWebsiteId || !serviceId || !bookingDatePattern.test(dateValue)) {
            manualAppointmentAvailableSlots = [];
            manualAppointmentSlotsError = "";
            isLoadingManualAppointmentSlots = false;
            return;
        }

        isLoadingManualAppointmentSlots = true;
        manualAppointmentSlotsError = "";

        try {
            const query = new URLSearchParams({
                websiteId: selectedWebsiteId,
                serviceId,
                date: dateValue,
            });

            const response = await ApiClient.send(`/api/nuvio/booking/slots?${query.toString()}`, {
                method: "GET",
                requestKey: `nuvio_booking_manual_slots_${selectedWebsiteId}_${serviceId}_${dateValue}`,
            });

            const slots = Array.isArray(response?.slots)
                ? response.slots.map((slot) => normalizeString(slot)).filter((slot) => bookingTimePattern.test(slot))
                : [];

            if (!isManualAppointmentPanelOpen || manualAppointmentSlotsQueryKey !== requestScopeKey) {
                return;
            }

            manualAppointmentAvailableSlots = slots;

            if (!slots.includes(normalizeString(manualAppointmentForm.time))) {
                manualAppointmentForm = {
                    ...manualAppointmentForm,
                    time: "",
                };
            }
        } catch (err) {
            ApiClient.error(err, false);

            if (!isManualAppointmentPanelOpen || manualAppointmentSlotsQueryKey !== requestScopeKey) {
                return;
            }

            manualAppointmentAvailableSlots = [];
            manualAppointmentSlotsError = "Unable to load available times right now.";
            manualAppointmentForm = {
                ...manualAppointmentForm,
                time: "",
            };
        } finally {
            isLoadingManualAppointmentSlots = false;
        }
    }

    function validateManualAppointmentForm() {
        const serviceId = normalizeString(manualAppointmentForm.serviceId);
        const dateValue = normalizeString(manualAppointmentForm.date);
        const timeValue = normalizeString(manualAppointmentForm.time);
        const nameValue = normalizeString(manualAppointmentForm.name);
        const emailValue = normalizeString(manualAppointmentForm.email);
        const statusValue = normalizeLower(manualAppointmentForm.status);

        if (!selectedWebsiteId) {
            return "Select a website before creating an appointment.";
        }

        if (!serviceId) {
            return "Service is required.";
        }

        if (!bookingDatePattern.test(dateValue)) {
            return "Date is required.";
        }

        if (!timeValue || !bookingTimePattern.test(timeValue)) {
            return "Select an available time.";
        }

        if (
            Array.isArray(manualAppointmentAvailableSlots)
            && manualAppointmentAvailableSlots.length
            && !manualAppointmentAvailableSlots.includes(timeValue)
        ) {
            return "Select an available time.";
        }

        if (!nameValue) {
            return "Customer name is required.";
        }

        if (!isValidEmail(emailValue)) {
            return "A valid customer email is required.";
        }

        if (!["pending", "confirmed"].includes(statusValue)) {
            return "Status must be pending or confirmed.";
        }

        return "";
    }

    async function createManualAppointment() {
        if (isCreatingManualAppointment || isLoadingManualAppointmentSlots) {
            return;
        }

        const validationError = validateManualAppointmentForm();
        if (validationError) {
            manualAppointmentFormError = validationError;
            return;
        }

        isCreatingManualAppointment = true;
        manualAppointmentFormError = "";

        const payload = {
            websiteId: selectedWebsiteId,
            serviceId: normalizeString(manualAppointmentForm.serviceId),
            date: normalizeString(manualAppointmentForm.date),
            time: normalizeString(manualAppointmentForm.time),
            name: normalizeString(manualAppointmentForm.name),
            email: normalizeString(manualAppointmentForm.email).toLowerCase(),
            phone: normalizeString(manualAppointmentForm.phone),
            notes: normalizeString(manualAppointmentForm.notes),
            internalNotes: normalizeString(manualAppointmentForm.internalNotes),
            status: normalizeLower(manualAppointmentForm.status) || "confirmed",
            createContact: true,
            sendConfirmationEmail: false,
        };

        try {
            const response = await ApiClient.send("/api/nuvio/booking/admin/appointments", {
                method: "POST",
                body: payload,
                requestKey: `nuvio_booking_manual_create_${selectedWebsiteId}`,
            });

            await loadBookingData();

            const createdAppointmentId = normalizeString(response?.appointmentId);
            if (createdAppointmentId) {
                selectedAppointmentId = createdAppointmentId;
            }

            closeManualAppointmentPanel();
            addSuccessToast("Appointment created.");

            if (normalizeString(response?.warning)) {
                addErrorToast(normalizeString(response.warning));
            }
        } catch (err) {
            ApiClient.error(err, false);

            const statusCode = (err?.status << 0) || 0;
            const conflictMessage = "This time is no longer available. Please choose another time.";
            if (statusCode === 409) {
                manualAppointmentFormError = conflictMessage;
                addErrorToast(conflictMessage);
                await loadManualAppointmentSlots();
            } else {
                addErrorToast("Unable to create appointment right now.");
            }
        } finally {
            isCreatingManualAppointment = false;
        }
    }

    function toLocalDateParts(dateValue) {
        const raw = normalizeString(dateValue);
        const match = raw.match(/^(\d{4})-(\d{2})-(\d{2})$/);
        if (!match) {
            return null;
        }

        const year = Number(match[1]);
        const month = Number(match[2]);
        const day = Number(match[3]);

        if (!Number.isFinite(year) || !Number.isFinite(month) || !Number.isFinite(day)) {
            return null;
        }

        return { year, month, day };
    }

    function toLocalDate(dateValue) {
        const parts = toLocalDateParts(dateValue);
        if (!parts) {
            return null;
        }

        return new Date(parts.year, parts.month - 1, parts.day);
    }

    function toAppointmentTimestamp(dateValue, timeValue) {
        const dateParts = toLocalDateParts(dateValue);
        const timeMatch = normalizeString(timeValue).match(/^(\d{2}):(\d{2})$/);

        if (!dateParts || !timeMatch) {
            return 0;
        }

        const hour = Number(timeMatch[1]);
        const minute = Number(timeMatch[2]);

        if (!Number.isFinite(hour) || !Number.isFinite(minute)) {
            return 0;
        }

        const parsed = new Date(
            dateParts.year,
            dateParts.month - 1,
            dateParts.day,
            hour,
            minute,
            0,
            0,
        ).getTime();

        return Number.isNaN(parsed) ? 0 : parsed;
    }

    function toTimestamp(value) {
        const raw = normalizeString(value);
        if (!raw) {
            return 0;
        }

        const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
        const timestamp = new Date(normalized).getTime();
        return Number.isNaN(timestamp) ? 0 : timestamp;
    }

    function formatDateTime(value) {
        const timestamp = toTimestamp(value);
        if (!timestamp) {
            return "-";
        }

        return new Date(timestamp).toLocaleString();
    }

    function formatDate(value) {
        const date = toLocalDate(value);
        if (!date) {
            return "-";
        }

        return date.toLocaleDateString();
    }

    function formatAppointmentDateTime(dateValue, timeValue) {
        const dateLabel = formatDate(dateValue);
        const timeLabel = normalizeString(timeValue) || "--:--";
        return `${dateLabel} at ${timeLabel}`;
    }

    function getStatusMeta(statusKey) {
        if (statusKey === "confirmed") {
            return { label: "Confirmed", className: "label-success" };
        }

        if (statusKey === "cancelled") {
            return { label: "Cancelled", className: "label-danger" };
        }

        return { label: "Pending", className: "label-warning" };
    }

    function resolveBookingReadinessState(warningsCount) {
        if (warningsCount >= 2) {
            return {
                label: "Missing basics",
                badgeClass: "label-danger",
            };
        }

        if (warningsCount === 1) {
            return {
                label: "Needs attention",
                badgeClass: "label-warning",
            };
        }

        return {
            label: "Ready",
            badgeClass: "label-success",
        };
    }

    function buildBookingReadinessWarnings({
        activeServicesCount = 0,
        activeAvailabilityDaysCount = 0,
        bookingSettings = {},
    } = {}) {
        const warnings = [];
        const settings = isPlainObject(bookingSettings) ? bookingSettings : {};
        const featureAvailable = settings.featureAvailable !== false;
        const bookingEnabled = settings.enabled !== false;
        const notificationsReady = !!settings.businessNotificationsReady;

        if (activeServicesCount <= 0) {
            warnings.push("Add at least one active service.");
        }

        if (activeAvailabilityDaysCount <= 0) {
            warnings.push("Set at least one active availability day.");
        }

        if (!featureAvailable) {
            warnings.push("Booking feature availability is turned off for this website.");
        }

        if (featureAvailable && !bookingEnabled) {
            warnings.push("Booking request intake is disabled in Website Settings.");
        }

        if (featureAvailable && bookingEnabled && !notificationsReady) {
            warnings.push("Business email notifications are missing or disabled for Booking requests.");
        }

        return warnings;
    }

    function buildBookingReadinessSuggestions({
        activeServicesCount = 0,
        activeAvailabilityDaysCount = 0,
        bookingSettings = {},
    } = {}) {
        const settings = isPlainObject(bookingSettings) ? bookingSettings : {};
        const featureAvailable = settings.featureAvailable !== false;
        const bookingEnabled = settings.enabled !== false;
        const notificationsReady = !!settings.businessNotificationsReady;
        const usingContactFormFallback = !!settings.usingContactFormFallback;
        const suggestions = [
            "Publish a page with the Booking block in CMS to start receiving requests.",
        ];

        if (!featureAvailable) {
            suggestions.unshift("Ask an admin to enable Booking in Website Settings feature availability.");
        } else if (!bookingEnabled) {
            suggestions.unshift("Enable Booking in Website Settings > Features > Booking to accept new requests.");
        }

        if (featureAvailable && bookingEnabled && !notificationsReady) {
            suggestions.push("Configure Booking email notifications so your team is alerted when new requests arrive.");
        }

        if (usingContactFormFallback) {
            suggestions.push("Booking notifications are currently using Contact Form recipients. Add Booking recipients when ready.");
        }

        if (featureAvailable && bookingEnabled && activeServicesCount > 0 && activeAvailabilityDaysCount > 0) {
            suggestions.unshift("Booking setup basics are ready. New requests can arrive once the public block is live.");
        } else {
            suggestions.push("Keep service durations and weekly schedule aligned with your team capacity.");
        }

        return suggestions;
    }

    function buildAvailabilityHealthWarnings({
        bookingWarnings = [],
        availabilityValidationIssuesCount = 0,
        dirtyAvailabilityRowsCount = 0,
    } = {}) {
        const warnings = Array.isArray(bookingWarnings) ? [...bookingWarnings] : [];

        if (availabilityValidationIssuesCount > 0) {
            warnings.push(
                `${availabilityValidationIssuesCount} active day${availabilityValidationIssuesCount === 1 ? "" : "s"} need valid start and end times.`,
            );
        }

        if (dirtyAvailabilityRowsCount > 0) {
            warnings.push("There are unsaved availability changes.");
        }

        return [...new Set(warnings)];
    }

    function buildAvailabilityHealthSuggestions({
        bookingSuggestions = [],
        dirtyAvailabilityRowsCount = 0,
        availabilityValidationIssuesCount = 0,
    } = {}) {
        const suggestions = Array.isArray(bookingSuggestions) ? [...bookingSuggestions] : [];

        if (dirtyAvailabilityRowsCount > 0) {
            suggestions.unshift("Save changes after updating schedule rows or using presets.");
        }

        if (availabilityValidationIssuesCount > 0) {
            suggestions.unshift("Fix active day time ranges where end time is not after start time.");
        }

        if (!dirtyAvailabilityRowsCount && !availabilityValidationIssuesCount) {
            suggestions.unshift("Use Slot preview to verify what visitors can book for each service.");
        }

        return [...new Set(suggestions)];
    }

    function normalizeAppointment(record) {
        const id = normalizeString(record?.id);
        const serviceId = normalizeRelationId(record?.service);
        const expandedService = Array.isArray(record?.expand?.service)
            ? record.expand.service[0]
            : record?.expand?.service;

        const serviceLabel = normalizeString(expandedService?.name)
            || serviceLabelById.get(serviceId)
            || "Service not found";

        const statusKey = normalizeStatus(record?.[appointmentStatusFieldName]);
        const statusMeta = getStatusMeta(statusKey);

        return {
            ...record,
            id,
            serviceId,
            serviceLabel,
            name: normalizeString(record?.name) || "Unnamed customer",
            email: normalizeString(record?.email),
            phone: normalizeString(record?.phone),
            date: normalizeString(record?.date),
            time: normalizeString(record?.time),
            customerNotes: normalizeString(record?.[appointmentCustomerNotesFieldName]),
            internalNotes: normalizeString(record?.[appointmentInternalNotesFieldName]),
            statusKey,
            statusLabel: statusMeta.label,
            statusClassName: statusMeta.className,
        };
    }

    function getStartOfToday() {
        const now = new Date();
        return new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0).getTime();
    }

    function getCurrentWeekRange() {
        const todayStart = getStartOfToday();
        const date = new Date(todayStart);
        const dayIndex = date.getDay();
        const mondayOffset = (dayIndex + 6) % 7;

        const start = new Date(date);
        start.setDate(date.getDate() - mondayOffset);

        const end = new Date(start);
        end.setDate(start.getDate() + 7);

        return {
            start: start.getTime(),
            end: end.getTime(),
        };
    }

    function isInCurrentWeek(dateValue) {
        const date = toLocalDate(dateValue);
        if (!date) {
            return false;
        }

        const range = getCurrentWeekRange();
        const timestamp = date.getTime();
        return timestamp >= range.start && timestamp < range.end;
    }

    function matchAppointmentDateFilter(appointment, filterKey) {
        if (filterKey === "all") {
            return true;
        }

        const date = toLocalDate(appointment.date);
        if (!date) {
            return false;
        }

        const timestamp = date.getTime();
        const todayStart = getStartOfToday();

        if (filterKey === "today") {
            return timestamp === todayStart;
        }

        if (filterKey === "thisWeek") {
            const weekRange = getCurrentWeekRange();
            return timestamp >= weekRange.start && timestamp < weekRange.end;
        }

        if (filterKey === "thisMonth") {
            const now = new Date();
            return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
        }

        if (filterKey === "upcoming") {
            const appointmentTimestamp = toAppointmentTimestamp(appointment.date, appointment.time);
            const fallbackTimestamp = timestamp;
            const referenceTimestamp = appointmentTimestamp || fallbackTimestamp;
            return referenceTimestamp >= Date.now();
        }

        return true;
    }

    function sortAppointments(list, sortKey) {
        const items = [...list];

        if (sortKey === "oldest") {
            return items.sort((a, b) => toTimestamp(a.created) - toTimestamp(b.created));
        }

        if (sortKey === "upcoming") {
            return items.sort((a, b) => {
                const first = toAppointmentTimestamp(a.date, a.time) || toTimestamp(a.created);
                const second = toAppointmentTimestamp(b.date, b.time) || toTimestamp(b.created);
                return first - second;
            });
        }

        return items.sort((a, b) => toTimestamp(b.created) - toTimestamp(a.created));
    }

    function patchAppointmentRecord(appointmentId, patchData = {}) {
        appointmentRecords = (appointmentRecords || []).map((record) => {
            if (normalizeString(record?.id) !== normalizeString(appointmentId)) {
                return record;
            }

            return {
                ...record,
                ...patchData,
            };
        });
    }

    async function setSelectedAppointmentStatus(nextStatus) {
        if (!selectedAppointment?.id || !nextStatus || isUpdatingAppointmentStatus || !appointmentsCollection?.id) {
            return;
        }

        isUpdatingAppointmentStatus = true;
        updatingAppointmentId = selectedAppointment.id;

        try {
            await ApiClient.collection(appointmentsCollection.id).update(selectedAppointment.id, {
                [appointmentStatusFieldName]: nextStatus,
            });

            patchAppointmentRecord(selectedAppointment.id, {
                [appointmentStatusFieldName]: nextStatus,
            });

            addSuccessToast(`Appointment marked as ${nextStatus}.`);
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to update appointment status right now.");
        } finally {
            isUpdatingAppointmentStatus = false;
            updatingAppointmentId = "";
        }
    }

    async function saveSelectedAppointmentInternalNotes() {
        if (!selectedAppointment?.id || !appointmentsCollection?.id || isSavingAppointmentInternalNotes) {
            return;
        }

        isSavingAppointmentInternalNotes = true;

        try {
            await ApiClient.collection(appointmentsCollection.id).update(selectedAppointment.id, {
                [appointmentInternalNotesFieldName]: appointmentInternalNotesDraft,
            });

            patchAppointmentRecord(selectedAppointment.id, {
                [appointmentInternalNotesFieldName]: appointmentInternalNotesDraft,
            });

            addSuccessToast("Internal notes saved.");
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to save internal notes right now.");
        } finally {
            isSavingAppointmentInternalNotes = false;
        }
    }

    function resetServiceForm() {
        serviceForm = {
            id: "",
            name: "",
            durationMinutes: "30",
            active: true,
        };
        serviceFormError = "";
    }

    function setServiceFormFromRecord(service) {
        serviceForm = {
            id: normalizeString(service?.id),
            name: normalizeString(service?.name),
            durationMinutes: `${service?.durationMinutes || "30"}`,
            active: !!service?.active,
        };
        serviceFormError = "";
    }

    function createNewService() {
        selectedServiceId = "";
        resetServiceForm();
    }

    function selectService(service) {
        selectedServiceId = normalizeString(service?.id);
    }

    async function toggleServiceActive(service, nextActive) {
        const serviceId = normalizeString(service?.id);
        if (!serviceId || !bookingServicesCollection?.id) {
            return;
        }

        try {
            await ApiClient.collection(bookingServicesCollection.id).update(serviceId, {
                active: !!nextActive,
            });

            servicesRecords = servicesRecords.map((record) =>
                normalizeString(record?.id) === serviceId
                    ? { ...record, active: !!nextActive }
                    : record,
            );

            addSuccessToast(`Service ${nextActive ? "activated" : "deactivated"}.`);
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to update service status right now.");
        }
    }

    async function saveService() {
        if (!selectedWebsiteId || !bookingServicesCollection?.id || isSavingService) {
            return;
        }

        const normalizedName = normalizeString(serviceForm.name);
        const duration = Number.parseInt(`${serviceForm.durationMinutes || ""}`.trim(), 10);

        if (!normalizedName) {
            serviceFormError = "Service name is required.";
            return;
        }

        if (!Number.isFinite(duration) || duration < 5 || duration > 480) {
            serviceFormError = "Duration must be an integer between 5 and 480 minutes.";
            return;
        }

        serviceFormError = "";
        isSavingService = true;

        const payload = {
            name: normalizedName,
            durationMinutes: duration,
            active: !!serviceForm.active,
        };

        try {
            if (serviceForm.id) {
                const updated = await ApiClient.collection(bookingServicesCollection.id).update(serviceForm.id, payload);
                servicesRecords = servicesRecords.map((record) =>
                    normalizeString(record?.id) === normalizeString(updated?.id)
                        ? { ...record, ...updated }
                        : record,
                );
                addSuccessToast("Service updated.");
            } else {
                const created = await ApiClient.collection(bookingServicesCollection.id).create({
                    website: selectedWebsiteId,
                    ...payload,
                });
                servicesRecords = [created, ...servicesRecords];
                selectedServiceId = created.id;
                setServiceFormFromRecord(created);
                addSuccessToast("Service created.");
            }
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to save service right now.");
        } finally {
            isSavingService = false;
        }
    }

    function createDefaultAvailabilityRows() {
        return availabilityDays.map((day) => ({
            dayOfWeek: day.key,
            label: day.label,
            recordId: "",
            active: false,
            startTime: "09:00",
            endTime: "17:00",
            initialActive: false,
            initialStartTime: "09:00",
            initialEndTime: "17:00",
            dirty: false,
            error: "",
        }));
    }

    function createAvailabilityRowsFromRecords(records = []) {
        const byDay = new Map();

        for (const record of records) {
            const dayKey = normalizeLower(record?.dayOfWeek);
            if (!dayKey || byDay.has(dayKey)) {
                continue;
            }
            byDay.set(dayKey, record);
        }

        return availabilityDays.map((day) => {
            const record = byDay.get(day.key);
            return {
                dayOfWeek: day.key,
                label: day.label,
                recordId: normalizeString(record?.id),
                active: !!record?.active,
                startTime: normalizeString(record?.startTime) || "09:00",
                endTime: normalizeString(record?.endTime) || "17:00",
                initialActive: !!record?.active,
                initialStartTime: normalizeString(record?.startTime) || "09:00",
                initialEndTime: normalizeString(record?.endTime) || "17:00",
                dirty: false,
                error: "",
            };
        });
    }

    function isAvailabilityRowDirty(row) {
        if (!row) {
            return false;
        }

        return !!row.active !== !!row.initialActive
            || normalizeString(row.startTime) !== normalizeString(row.initialStartTime)
            || normalizeString(row.endTime) !== normalizeString(row.initialEndTime);
    }

    function updateAvailabilityRow(dayKey, patchData = {}) {
        availabilityRows = availabilityRows.map((row) =>
            row.dayOfWeek === dayKey
                ? (() => {
                    const next = {
                        ...row,
                        ...patchData,
                    };

                    return {
                        ...next,
                        dirty: isAvailabilityRowDirty(next),
                    };
                })()
                : row,
        );
    }

    function parseTimeToMinutes(value) {
        const match = normalizeString(value).match(/^(\d{2}):(\d{2})$/);
        if (!match) {
            return -1;
        }

        const hour = Number(match[1]);
        const minute = Number(match[2]);
        if (hour < 0 || hour > 23 || minute < 0 || minute > 59) {
            return -1;
        }

        return hour * 60 + minute;
    }

    function validateAvailabilityRow(row) {
        if (!row.active) {
            return "";
        }

        const timePattern = /^([01]\d|2[0-3]):[0-5]\d$/;
        if (!timePattern.test(normalizeString(row.startTime)) || !timePattern.test(normalizeString(row.endTime))) {
            return "Start and end time must use HH:mm format.";
        }

        const startMinutes = parseTimeToMinutes(row.startTime);
        const endMinutes = parseTimeToMinutes(row.endTime);
        if (startMinutes < 0 || endMinutes < 0 || endMinutes <= startMinutes) {
            return "End time must be after start time.";
        }

        return "";
    }

    async function saveAvailabilityRow(row, options = {}) {
        if (!row || !selectedWebsiteId || !bookingAvailabilityCollection?.id) {
            return false;
        }

        const validationError = validateAvailabilityRow(row);
        if (validationError) {
            updateAvailabilityRow(row.dayOfWeek, { error: validationError });
            return false;
        }

        updateAvailabilityRow(row.dayOfWeek, { error: "" });
        isSavingAvailability = { ...isSavingAvailability, [row.dayOfWeek]: true };

        const normalizedStartTime = normalizeString(row.startTime) || "09:00";
        const normalizedEndTime = normalizeString(row.endTime) || "17:00";
        const normalizedActive = !!row.active;

        const payload = {
            website: selectedWebsiteId,
            dayOfWeek: row.dayOfWeek,
            startTime: normalizedStartTime,
            endTime: normalizedEndTime,
            active: normalizedActive,
        };

        try {
            if (row.recordId) {
                const updated = await ApiClient.collection(bookingAvailabilityCollection.id).update(row.recordId, payload);
                availabilityRecords = availabilityRecords.map((record) =>
                    normalizeString(record?.id) === normalizeString(updated?.id)
                        ? { ...record, ...updated }
                        : record,
                );
                updateAvailabilityRow(row.dayOfWeek, {
                    recordId: updated.id,
                    initialActive: normalizedActive,
                    initialStartTime: normalizedStartTime,
                    initialEndTime: normalizedEndTime,
                    error: "",
                });
                if (!options?.silent) {
                    addSuccessToast(`${row.label} availability updated.`);
                }
            } else {
                const created = await ApiClient.collection(bookingAvailabilityCollection.id).create(payload);
                availabilityRecords = [created, ...availabilityRecords];
                updateAvailabilityRow(row.dayOfWeek, {
                    recordId: created.id,
                    initialActive: normalizedActive,
                    initialStartTime: normalizedStartTime,
                    initialEndTime: normalizedEndTime,
                    error: "",
                });
                if (!options?.silent) {
                    addSuccessToast(`${row.label} availability saved.`);
                }
            }
            return true;
        } catch (err) {
            ApiClient.error(err, false);
            updateAvailabilityRow(row.dayOfWeek, { error: "Unable to save this day right now." });
            if (!options?.silent) {
                addErrorToast("Unable to save availability right now.");
            }
            return false;
        } finally {
            isSavingAvailability = { ...isSavingAvailability, [row.dayOfWeek]: false };
        }
    }

    function isDaySaving(dayKey) {
        return !!isSavingAvailability?.[dayKey];
    }

    function applyMondayToWeekdays() {
        const mondayRow = availabilityRows.find((row) => row.dayOfWeek === "mon");
        if (!mondayRow) {
            return;
        }

        const weekdayKeys = ["tue", "wed", "thu", "fri"];

        for (const dayKey of weekdayKeys) {
            updateAvailabilityRow(dayKey, {
                active: !!mondayRow.active,
                startTime: mondayRow.startTime,
                endTime: mondayRow.endTime,
                error: "",
            });
        }
    }

    function setWeekdaysBusinessHours() {
        const weekdayKeys = ["mon", "tue", "wed", "thu", "fri"];

        for (const dayKey of weekdayKeys) {
            updateAvailabilityRow(dayKey, {
                active: true,
                startTime: "09:00",
                endTime: "17:00",
                error: "",
            });
        }
    }

    function clearWeekendAvailability() {
        const weekendKeys = ["sat", "sun"];

        for (const dayKey of weekendKeys) {
            updateAvailabilityRow(dayKey, {
                active: false,
                error: "",
            });
        }
    }

    async function saveAllChangedAvailabilityRows() {
        if (isSavingAllAvailability || hasSavingAvailabilityRows) {
            return;
        }

        const dirtyRows = availabilityRows.filter((row) => !!row.dirty);
        if (!dirtyRows.length) {
            addSuccessToast("No unsaved availability changes.");
            return;
        }

        let invalidRowsCount = 0;
        const rowsToSave = [];

        for (const row of dirtyRows) {
            const validationError = validateAvailabilityRow(row);
            if (validationError) {
                updateAvailabilityRow(row.dayOfWeek, { error: validationError });
                invalidRowsCount += 1;
                continue;
            }

            updateAvailabilityRow(row.dayOfWeek, { error: "" });
            rowsToSave.push(row);
        }

        if (!rowsToSave.length) {
            addErrorToast("Fix validation errors before saving changed days.");
            return;
        }

        isSavingAllAvailability = true;

        try {
            const saveResults = await Promise.all(
                rowsToSave.map((row) => saveAvailabilityRow(row, { silent: true })),
            );

            const savedCount = saveResults.filter(Boolean).length;
            const failedCount = saveResults.length - savedCount;

            if (savedCount > 0) {
                addSuccessToast(`${savedCount} availability day${savedCount === 1 ? "" : "s"} saved.`);
            }

            if (failedCount > 0 || invalidRowsCount > 0) {
                addErrorToast("Some availability changes could not be saved.");
            }
        } finally {
            isSavingAllAvailability = false;
        }
    }

    function clearAppointmentFilters() {
        appointmentStatusFilter = "all";
        appointmentDateFilter = "all";
        appointmentServiceFilter = "all";
        appointmentSearch = "";
        appointmentSort = "newest";
    }

    export function reload() {
        return Promise.all([loadWebsites(), loadBookingData()]);
    }
    // NUVIO CUSTOM END: Booking MVP Phase 2 purpose-built backoffice module.
</script>

<PageWrapper>
    <section class="operations-head panel booking-head m-b-base">
        <div class="head-main">
            <div class="summary-title-wrap">
                <div class="title-row">
                    <h2 class="m-0">Booking</h2>
                    <RefreshButton class="btn-sm" tooltip={{ text: "Refresh", position: "right" }} on:refresh={reload} />
                </div>
                <p class="head-description txt-sm txt-hint m-b-0">
                    Manage appointment requests, services, and availability for this website.
                </p>
            </div>

            <div class="head-selector">
                <div class="selector-row">
                    <label class="txt-sm txt-hint selector-label m-b-0" for="booking-website-selector">Website</label>
                    <select
                        id="booking-website-selector"
                        class="input input-sm"
                        bind:value={selectedWebsiteId}
                        disabled={isLoadingWebsites || !websites.length}
                    >
                        {#if !websites.length}
                            <option value="">No websites available</option>
                        {:else}
                            {#each websites as website (website.id)}
                                <option value={website.id}>{resolveWebsiteLabel(website)}</option>
                            {/each}
                        {/if}
                    </select>
                </div>
            </div>
        </div>

        <div class="head-tools">
            <div class="tabs-header compact combined left operations-tabs booking-tabs booking-head-tabs">
                {#each bookingTabs as tab (tab.key)}
                    <button
                        type="button"
                        class="tab-item"
                        class:active={activeTab === tab.key}
                        on:click={() => (activeTab = tab.key)}
                    >
                        <i class={`${tab.icon} tab-icon`} aria-hidden="true" />
                        <span class="tab-label">{tab.label}</span>
                    </button>
                {/each}
            </div>

            <div class="summary-badges">
                <span class="summary-pill">
                    <i class="ri-time-line" />
                    Pending: {pendingAppointmentsCount}
                </span>
                <span class="summary-pill">
                    <i class="ri-checkbox-circle-line" />
                    Confirmed: {confirmedAppointmentsCount}
                </span>
                <span class="summary-pill">
                    <i class="ri-calendar-event-line" />
                    This week: {thisWeekAppointmentsCount}
                </span>
                <span class="summary-pill">
                    <i class="ri-briefcase-4-line" />
                    Active services: {activeServicesCount}
                </span>
            </div>
        </div>
    </section>

    <section class="panel booking-body m-b-base">
        {#if !hasBookingCollections}
            <div class="alert alert-warning m-b-0">
                <div class="icon">
                    <i class="ri-information-line" />
                </div>
                <div>
                    Booking collections were not found. This page expects BookingServices, BookingAvailability, and Appointments.
                </div>
            </div>
        {:else if !selectedWebsiteId}
            <div class="placeholder-section m-b-0">
                <h1>Select a website to manage Booking.</h1>
                <p class="txt-sm txt-hint m-b-0">
                    Once selected, appointments, services, and availability will load automatically.
                </p>
            </div>
        {:else if bookingLoadError}
            <div class="alert alert-danger m-b-0">
                <div class="icon">
                    <i class="ri-error-warning-line" />
                </div>
                <div>{bookingLoadError}</div>
            </div>
        {:else if $isCollectionsLoading || isLoadingBookingData}
            <div class="placeholder-section m-b-0">
                <span class="loader loader-lg" />
                <h1>Loading booking data...</h1>
            </div>
        {:else if activeTab === "appointments"}
            <div class="booking-split-layout">
                <div class="booking-main-column">
                    <div class="booking-section-head-row booking-section-head-row--appointments">
                        <div class="booking-section-head-copy">
                            <h4 class="m-0">Appointments inbox</h4>
                            <p class="txt-sm txt-hint m-b-0">
                                Monitor appointment requests, update status, and keep follow-up notes.
                            </p>
                        </div>
                        <div class="booking-section-head-meta">
                            <span class="summary-pill">{filteredAppointments.length} visible</span>
                            <button type="button" class="btn btn-outline btn-sm" on:click={openManualAppointmentPanel}>
                                <span class="txt">New appointment</span>
                            </button>
                        </div>
                    </div>

                    <div class="booking-controls booking-controls--appointments">
                        <div class="booking-control-cell booking-control-cell--search">
                            <label class="txt-sm txt-hint" for="booking-appointment-search">Search</label>
                            <input
                                id="booking-appointment-search"
                                class="input input-sm"
                                type="text"
                                placeholder="Search by name, email, phone, service, or notes..."
                                bind:value={appointmentSearch}
                            />
                        </div>

                        <div class="booking-control-cell booking-control-cell--select">
                            <label class="txt-sm txt-hint" for="booking-appointment-status-filter">Status</label>
                            <select id="booking-appointment-status-filter" class="input input-sm" bind:value={appointmentStatusFilter}>
                                {#each appointmentStatusOptions as option (option.key)}
                                    <option value={option.key}>{option.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="booking-control-cell booking-control-cell--select">
                            <label class="txt-sm txt-hint" for="booking-appointment-date-filter">Date</label>
                            <select id="booking-appointment-date-filter" class="input input-sm" bind:value={appointmentDateFilter}>
                                {#each appointmentDateOptions as option (option.key)}
                                    <option value={option.key}>{option.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="booking-control-cell booking-control-cell--select">
                            <label class="txt-sm txt-hint" for="booking-appointment-service-filter">Service</label>
                            <select id="booking-appointment-service-filter" class="input input-sm" bind:value={appointmentServiceFilter}>
                                <option value="all">All services</option>
                                {#each appointmentServiceOptions as serviceOption (serviceOption.id)}
                                    <option value={serviceOption.id}>{serviceOption.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="booking-control-cell booking-control-cell--select">
                            <label class="txt-sm txt-hint" for="booking-appointment-sort">Sort</label>
                            <select id="booking-appointment-sort" class="input input-sm" bind:value={appointmentSort}>
                                {#each appointmentSortOptions as option (option.key)}
                                    <option value={option.key}>{option.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="booking-control-cell booking-control-cell--actions">
                            <button type="button" class="btn btn-sm btn-outline" on:click={clearAppointmentFilters}>
                                <span class="txt">Reset filters</span>
                            </button>
                        </div>
                    </div>

                    {#if !normalizedAppointments.length}
                        <div class="booking-empty-state m-b-0">
                            <h5 class="m-0">No appointment requests yet.</h5>
                            <p class="txt-sm txt-hint m-b-0">
                                Appointment requests will appear here once visitors submit the booking form.
                            </p>
                            <div class="booking-empty-readiness">
                                <div class="booking-empty-readiness-row">
                                    <span class="txt-sm">Active services</span>
                                    <span class={`label label-sm ${activeServicesCount > 0 ? "label-success" : "label-warning"}`}>
                                        {activeServicesCount > 0 ? `${activeServicesCount} ready` : "Missing"}
                                    </span>
                                </div>
                                <div class="booking-empty-readiness-row">
                                    <span class="txt-sm">Active availability days</span>
                                    <span class={`label label-sm ${activeAvailabilityDaysCount > 0 ? "label-success" : "label-warning"}`}>
                                        {activeAvailabilityDaysCount > 0 ? `${activeAvailabilityDaysCount} active` : "Missing"}
                                    </span>
                                </div>
                                <div class="booking-empty-readiness-row">
                                    <span class="txt-sm">Public booking block</span>
                                    <span class="txt-xs txt-hint">Configure in CMS pages</span>
                                </div>
                            </div>
                        </div>
                    {:else if !filteredAppointments.length}
                        <div class="empty-state m-b-0">
                            No appointments match these filters.
                        </div>
                    {:else}
                        <div class="booking-appointments-list" role="list">
                            {#each filteredAppointments as appointment (appointment.id)}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-static-element-interactions -->
                                <article
                                    class="booking-appointment-item"
                                    class:selected={selectedAppointmentId === appointment.id}
                                    role="button"
                                    tabindex="0"
                                    on:click={() => (selectedAppointmentId = appointment.id)}
                                >
                                    <div class="booking-appointment-head">
                                        <span class={`label label-sm ${appointment.statusClassName}`}>{appointment.statusLabel}</span>
                                        <span class="txt-xs txt-hint">{formatDateTime(appointment.created)}</span>
                                    </div>

                                    <div class="booking-appointment-title">{appointment.name}</div>

                                    <div class="booking-appointment-meta txt-sm">
                                        <span>{appointment.serviceLabel}</span>
                                        <span aria-hidden="true">&middot;</span>
                                        <span>{formatAppointmentDateTime(appointment.date, appointment.time)}</span>
                                    </div>

                                    {#if appointment.email || appointment.phone}
                                        <div class="booking-appointment-contact txt-xs txt-hint">
                                            {appointment.email || "No email"}{appointment.phone ? ` - ${appointment.phone}` : ""}
                                        </div>
                                    {/if}
                                </article>
                            {/each}
                        </div>
                    {/if}
                </div>

                <aside class="booking-rail" aria-live="polite">
                    {#if selectedAppointment}
                        <section class="booking-rail-block">
                            <h5 class="m-0">Appointment summary</h5>
                            <p class="txt-sm txt-hint m-b-0">Overview of this appointment request and customer details.</p>
                            <div class="booking-summary-grid">
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Status</span>
                                    <span class={`label label-sm ${selectedAppointment.statusClassName}`}>{selectedAppointment.statusLabel}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Customer</span>
                                    <span class="txt-sm">{selectedAppointment.name}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Service</span>
                                    <span class="txt-sm">{selectedAppointment.serviceLabel}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Date &amp; time</span>
                                    <span class="txt-sm">{formatAppointmentDateTime(selectedAppointment.date, selectedAppointment.time)}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Created</span>
                                    <span class="txt-sm">{formatDateTime(selectedAppointment.created)}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Customer email</span>
                                    <span class="txt-sm">{selectedAppointment.email || "No email provided"}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Customer phone</span>
                                    <span class="txt-sm">{selectedAppointment.phone || "No phone provided"}</span>
                                </div>
                            </div>
                        </section>

                        <section class="booking-rail-block">
                            <h5 class="m-0">Actions</h5>
                            <div class="booking-actions-row">
                                {#if canSetSelectedAppointmentConfirmed}
                                    <button
                                        type="button"
                                        class="btn btn-sm"
                                        class:btn-loading={isUpdatingAppointmentStatus && updatingAppointmentId === selectedAppointment.id}
                                        disabled={isUpdatingAppointmentStatus}
                                        on:click={() => setSelectedAppointmentStatus("confirmed")}
                                    >
                                        <span class="txt">Confirm appointment</span>
                                    </button>
                                {/if}

                                {#if canSetSelectedAppointmentCancelled}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        class:btn-loading={isUpdatingAppointmentStatus && updatingAppointmentId === selectedAppointment.id}
                                        disabled={isUpdatingAppointmentStatus}
                                        on:click={() => setSelectedAppointmentStatus("cancelled")}
                                    >
                                        <span class="txt">Cancel appointment</span>
                                    </button>
                                {/if}

                                {#if canSetSelectedAppointmentPending}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        class:btn-loading={isUpdatingAppointmentStatus && updatingAppointmentId === selectedAppointment.id}
                                        disabled={isUpdatingAppointmentStatus}
                                        on:click={() => setSelectedAppointmentStatus("pending")}
                                    >
                                        <span class="txt">Mark pending</span>
                                    </button>
                                {/if}
                            </div>
                        </section>

                        <section class="booking-rail-block">
                            <h5 class="m-0">Notes</h5>
                            <div class="booking-form-stack">
                                <div>
                                    <p class="txt-xs txt-hint m-b-0">Customer notes</p>
                                    {#if selectedAppointment.customerNotes}
                                        <p class="txt-sm m-b-0 booking-readonly-note">{selectedAppointment.customerNotes}</p>
                                    {:else}
                                        <p class="txt-sm txt-hint m-b-0">No customer notes provided.</p>
                                    {/if}
                                </div>

                                <div>
                                    <p class="txt-xs txt-hint m-b-0">Internal notes</p>
                                    <p class="txt-sm txt-hint m-b-0 booking-internal-notes-helper">Keep private follow-up notes for this appointment.</p>
                                </div>
                            </div>
                            <textarea
                                class="input booking-notes-input"
                                rows="5"
                                placeholder="Add internal follow-up notes..."
                                bind:value={appointmentInternalNotesDraft}
                                disabled={isSavingAppointmentInternalNotes}
                            />
                            <div class="booking-actions-row">
                                <button
                                    type="button"
                                    class="btn btn-sm"
                                    class:btn-loading={isSavingAppointmentInternalNotes}
                                    disabled={isSavingAppointmentInternalNotes || appointmentInternalNotesDraft === (selectedAppointment.internalNotes || "")}
                                    on:click={saveSelectedAppointmentInternalNotes}
                                >
                                    <span class="txt">Save internal notes</span>
                                </button>
                            </div>
                        </section>
                    {:else}
                        <section class="booking-rail-block">
                            <h5 class="m-0">Appointment details</h5>
                            <p class="txt-sm txt-hint m-b-0">Select an appointment to view details and update status.</p>
                        </section>
                    {/if}
                </aside>
            </div>
        {:else if activeTab === "services"}
            <div class="booking-split-layout booking-split-layout--services">
                <div class="booking-main-column">
                    <div class="booking-section-head-row">
                        <div class="booking-section-head-copy">
                            <h4 class="m-0">Services</h4>
                            <p class="txt-sm txt-hint m-b-0">Define what visitors can book and how long each service takes.</p>
                        </div>
                        <div class="booking-section-head-actions">
                            <span class="summary-pill">{activeServicesCount} active</span>
                            <button type="button" class="btn btn-outline btn-sm" on:click={createNewService}>
                                <span class="txt">New service</span>
                            </button>
                        </div>
                    </div>

                    <div class="booking-services-controls">
                        <div class="booking-control-cell booking-control-cell--search">
                            <label class="txt-sm txt-hint" for="booking-services-search">Search</label>
                            <input
                                id="booking-services-search"
                                class="input input-sm"
                                type="text"
                                placeholder="Search by service name or duration..."
                                bind:value={serviceSearch}
                            />
                        </div>
                        <div class="booking-control-cell booking-control-cell--select">
                            <label class="txt-sm txt-hint" for="booking-services-status-filter">Status</label>
                            <select id="booking-services-status-filter" class="input input-sm" bind:value={serviceStatusFilter}>
                                <option value="all">All</option>
                                <option value="active">Active</option>
                                <option value="inactive">Inactive</option>
                            </select>
                        </div>
                    </div>

                    {#if !servicesRecords.length}
                        <div class="empty-state m-b-0">
                            No services yet. Add a service to start accepting bookings.
                        </div>
                    {:else if !filteredServices.length}
                        <div class="empty-state m-b-0">
                            No services match these filters.
                        </div>
                    {:else}
                        <div class="booking-services-list" role="list">
                            {#each filteredServices as service (service.id)}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-static-element-interactions -->
                                <article
                                    class="booking-service-item"
                                    class:selected={selectedServiceId === service.id}
                                    role="button"
                                    tabindex="0"
                                    on:click={() => selectService(service)}
                                >
                                    <div class="booking-service-main">
                                        <div class="booking-service-title-row">
                                            <div class="booking-service-title">{service.name || "Untitled service"}</div>
                                            <span class={`label label-sm ${service.active ? "label-success" : "label-warning"}`}>
                                                {service.active ? "Active" : "Inactive"}
                                            </span>
                                        </div>
                                        <div class="booking-service-meta txt-sm txt-hint">{service.durationMinutes || 0} minutes</div>
                                    </div>
                                    <div class="booking-service-actions">
                                        <button
                                            type="button"
                                            class="btn btn-outline btn-sm"
                                            on:click|stopPropagation={() => toggleServiceActive(service, !service.active)}
                                        >
                                            <span class="txt">{service.active ? "Deactivate" : "Activate"}</span>
                                        </button>
                                    </div>
                                </article>
                            {/each}
                        </div>
                    {/if}
                </div>

                <aside class="booking-rail">
                    <section class="booking-rail-block">
                        <h5 class="m-0">Service details</h5>
                        <p class="txt-sm txt-hint m-b-0">Create or update service name, duration, and active status.</p>
                        <div class="booking-actions-row">
                            <span class="summary-pill">{serviceForm.id ? "Editing selected service" : "Creating a new service"}</span>
                        </div>

                        <div class="booking-form-stack">
                            <div class="form-field">
                                <label class="txt-sm txt-hint" for="booking-service-name">Service name</label>
                                <input
                                    id="booking-service-name"
                                    class="input input-sm"
                                    type="text"
                                    placeholder="Example: Vehicle inspection"
                                    bind:value={serviceForm.name}
                                    disabled={isSavingService}
                                />
                            </div>

                            <div class="form-field">
                                <label class="txt-sm txt-hint" for="booking-service-duration">Duration (minutes)</label>
                                <input
                                    id="booking-service-duration"
                                    class="input input-sm"
                                    type="number"
                                    min="5"
                                    max="480"
                                    step="1"
                                    bind:value={serviceForm.durationMinutes}
                                    disabled={isSavingService}
                                />
                            </div>

                            <label class="booking-checkbox-row" for="booking-service-active">
                                <input
                                    id="booking-service-active"
                                    type="checkbox"
                                    bind:checked={serviceForm.active}
                                    disabled={isSavingService}
                                />
                                <span class="txt-sm">Active service</span>
                            </label>

                            {#if serviceFormError}
                                <p class="txt-xs txt-danger m-b-0">{serviceFormError}</p>
                            {/if}

                            <div class="booking-actions-row">
                                <button
                                    type="button"
                                    class="btn btn-sm"
                                    class:btn-loading={isSavingService}
                                    disabled={isSavingService}
                                    on:click={saveService}
                                >
                                    <span class="txt">{serviceForm.id ? "Update service" : "Create service"}</span>
                                </button>
                                {#if serviceForm.id}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        disabled={isSavingService}
                                        on:click={createNewService}
                                    >
                                        <span class="txt">New service</span>
                                    </button>
                                {/if}
                            </div>
                        </div>
                    </section>
                </aside>
            </div>
        {:else}
            <div class="booking-split-layout booking-split-layout--availability">
                <div class="booking-main-column booking-availability-main">
                    <div class="booking-section-head-row">
                        <div class="booking-section-head-copy">
                            <h4 class="m-0">Weekly availability</h4>
                            <p class="txt-sm txt-hint m-b-0">Set active days and time windows for appointment requests.</p>
                        </div>
                        <div class="booking-section-head-actions">
                            <span class="summary-pill">{activeAvailabilityDaysCount} active days</span>
                            <span class="summary-pill" class:warning={dirtyAvailabilityRowsCount > 0}>
                                {dirtyAvailabilityRowsCount} unsaved
                            </span>
                        </div>
                    </div>

                    <div class="booking-availability-helper txt-xs txt-hint">
                        Changes are local until you save. Inactive days do not require start/end times.
                    </div>

                    <div class="booking-availability-list">
                        {#each availabilityRows as row (row.dayOfWeek)}
                            <article class="booking-availability-row" class:is-inactive={!row.active}>
                                <div class="booking-availability-day">
                                    <div class="booking-availability-day-meta">
                                        <span class="txt-sm booking-availability-day-label">{row.label}</span>
                                        {#if row.dirty}
                                            <span class="label label-sm label-info booking-unsaved-pill">Unsaved</span>
                                        {/if}
                                    </div>
                                    <label class="booking-checkbox-row">
                                        <input
                                            type="checkbox"
                                            checked={row.active}
                                            on:change={(event) => updateAvailabilityRow(row.dayOfWeek, { active: !!event.currentTarget.checked, error: "" })}
                                        />
                                        <span class="txt-xs txt-hint">Active</span>
                                    </label>
                                </div>

                                <div class="booking-availability-time-range">
                                    <div class="form-field form-field--compact m-b-0">
                                        <label class="txt-xs txt-hint" for={`booking-start-${row.dayOfWeek}`}>Start</label>
                                        <input
                                            id={`booking-start-${row.dayOfWeek}`}
                                            class="input input-sm"
                                            type="time"
                                            value={row.startTime}
                                            disabled={!row.active || isDaySaving(row.dayOfWeek) || isSavingAllAvailability || hasSavingAvailabilityRows}
                                            on:input={(event) => updateAvailabilityRow(row.dayOfWeek, { startTime: event.currentTarget.value, error: "" })}
                                        />
                                    </div>
                                    <div class="form-field form-field--compact m-b-0">
                                        <label class="txt-xs txt-hint" for={`booking-end-${row.dayOfWeek}`}>End</label>
                                        <input
                                            id={`booking-end-${row.dayOfWeek}`}
                                            class="input input-sm"
                                            type="time"
                                            value={row.endTime}
                                            disabled={!row.active || isDaySaving(row.dayOfWeek) || isSavingAllAvailability || hasSavingAvailabilityRows}
                                            on:input={(event) => updateAvailabilityRow(row.dayOfWeek, { endTime: event.currentTarget.value, error: "" })}
                                        />
                                    </div>
                                </div>

                                {#if row.error}
                                    <p class="txt-xs txt-danger m-b-0 booking-availability-error">{row.error}</p>
                                {/if}
                            </article>
                        {/each}
                    </div>
                </div>

                <aside class="booking-rail">
                    <section class="booking-rail-block">
                        <h5 class="m-0">Schedule actions</h5>
                        <p class="txt-sm txt-hint m-b-0">Save schedule updates and apply common weekly presets.</p>
                        <button
                            type="button"
                            class="btn btn-sm"
                            class:btn-loading={isSavingAllAvailability}
                            disabled={isSavingAllAvailability || hasSavingAvailabilityRows || dirtyAvailabilityRowsCount === 0}
                            on:click={saveAllChangedAvailabilityRows}
                        >
                            <span class="txt">Save changes</span>
                        </button>
                        <div class="booking-availability-actions-list">
                            <button type="button" class="btn btn-outline btn-sm" disabled={isSavingAllAvailability || hasSavingAvailabilityRows} on:click={setWeekdaysBusinessHours}>
                                <span class="txt">Set weekdays 09:00-17:00</span>
                            </button>
                            <button type="button" class="btn btn-outline btn-sm" disabled={isSavingAllAvailability || hasSavingAvailabilityRows} on:click={applyMondayToWeekdays}>
                                <span class="txt">Apply Monday to weekdays</span>
                            </button>
                            <button type="button" class="btn btn-outline btn-sm" disabled={isSavingAllAvailability || hasSavingAvailabilityRows} on:click={clearWeekendAvailability}>
                                <span class="txt">Clear weekend</span>
                            </button>
                        </div>
                    </section>

                    <section class="booking-rail-block booking-health-panel">
                        <div class="booking-health-head">
                            <div class="booking-health-main">
                                <h5 class="m-0">Availability health</h5>
                                <p class="txt-sm txt-hint m-b-0">Check schedule readiness before receiving booking requests.</p>
                            </div>
                            <div class="booking-health-meta">
                                <span class={`label label-sm ${availabilityHealthState.badgeClass}`}>{availabilityHealthState.label}</span>
                                <span class="summary-pill">{availabilityHealthWarnings.length} warnings - {availabilityHealthSuggestions.length} suggestions</span>
                            </div>
                        </div>

                        <div class="booking-health-group m-t-8">
                            <div class="booking-health-group-title">Warnings</div>
                            {#if availabilityHealthWarnings.length}
                                {#each availabilityHealthWarnings as warning}
                                    <div class="booking-health-item warning">
                                        <span class="label label-sm booking-health-pill warning">Warning</span>
                                        <span>{warning}</span>
                                    </div>
                                {/each}
                            {:else}
                                <p class="txt-sm txt-hint m-b-0">Availability is in a healthy state.</p>
                            {/if}
                        </div>

                        <div class="booking-health-group m-t-8">
                            <div class="booking-health-group-title">Suggestions</div>
                            {#if availabilityHealthSuggestions.length}
                                {#each availabilityHealthSuggestions as suggestion}
                                    <div class="booking-health-item">
                                        <span class="label label-sm booking-health-pill">Info</span>
                                        <span>{suggestion}</span>
                                    </div>
                                {/each}
                            {:else}
                                <p class="txt-sm txt-hint m-b-0">No suggestions right now.</p>
                            {/if}
                        </div>
                    </section>

                    <section class="booking-rail-block booking-slot-preview-panel">
                        <div class="booking-health-head">
                            <div class="booking-health-main">
                                <h5 class="m-0">Slot preview</h5>
                                <p class="txt-sm txt-hint m-b-0">
                                    Preview the times visitors will be able to choose for a service and date.
                                </p>
                            </div>
                            <div class="booking-health-meta">
                                <span class="summary-pill">{slotPreviewSlots.length} slots</span>
                            </div>
                        </div>

                        <div class="booking-slot-preview-controls m-t-sm">
                            <div class="form-field m-b-0">
                                <label class="txt-sm txt-hint" for="booking-slot-preview-service">Service</label>
                                <select
                                    id="booking-slot-preview-service"
                                    class="input input-sm"
                                    value={slotPreviewServiceId}
                                    disabled={!slotPreviewServiceOptions.length || isLoadingSlotPreview}
                                    on:change={(event) => setSlotPreviewService(event.currentTarget.value)}
                                >
                                    {#if !slotPreviewServiceOptions.length}
                                        <option value="">No active services</option>
                                    {:else}
                                        {#each slotPreviewServiceOptions as service (service.id)}
                                            <option value={service.id}>{service.label}</option>
                                        {/each}
                                    {/if}
                                </select>
                            </div>

                            <div class="form-field m-b-0">
                                <label class="txt-sm txt-hint" for="booking-slot-preview-date">Date</label>
                                <input
                                    id="booking-slot-preview-date"
                                    class="input input-sm"
                                    type="date"
                                    value={slotPreviewDate}
                                    disabled={!slotPreviewServiceOptions.length || isLoadingSlotPreview}
                                    on:change={(event) => setSlotPreviewDate(event.currentTarget.value)}
                                />
                            </div>
                        </div>

                        {#if !slotPreviewServiceOptions.length}
                            <p class="txt-sm txt-hint m-t-sm m-b-0">Add an active service to preview booking slots.</p>
                        {:else if activeAvailabilityDaysCount <= 0}
                            <p class="txt-sm txt-hint m-t-sm m-b-0">Add active availability days to generate slots.</p>
                        {:else if !slotPreviewServiceId || !slotPreviewDate}
                            <p class="txt-sm txt-hint m-t-sm m-b-0">Select a service and date to preview slots.</p>
                        {:else if isLoadingSlotPreview}
                            <p class="txt-sm txt-hint m-t-sm m-b-0">Loading available times...</p>
                        {:else if slotPreviewError}
                            <p class="txt-sm txt-danger m-t-sm m-b-0">{slotPreviewError}</p>
                        {:else if !slotPreviewSlots.length}
                            <p class="txt-sm txt-hint m-t-sm m-b-0">No available times for this date.</p>
                        {:else}
                            <div class="booking-slot-preview-slots m-t-sm">
                                {#each slotPreviewSlots as slot}
                                    <span class="summary-pill booking-slot-preview-pill">{slot}</span>
                                {/each}
                            </div>
                        {/if}
                    </section>
                </aside>
            </div>
        {/if}
    </section>

    <OverlayPanel
        bind:active={isManualAppointmentPanelOpen}
        class="overlay-panel-lg booking-manual-appointment-panel"
        overlayClose={true}
        escClose={true}
        on:hide={closeManualAppointmentPanel}
    >
        <svelte:fragment slot="header">
            <h4>New appointment</h4>
        </svelte:fragment>

        <div class="booking-manual-form">
            <p class="txt-sm txt-hint m-b-0">
                Add appointments received by phone, WhatsApp, email, or in person. Manual appointments do not send confirmation emails by default.
            </p>

            {#if !manualAppointmentServiceOptions.length}
                <div class="alert alert-warning m-b-0">
                    <div class="icon">
                        <i class="ri-information-line" />
                    </div>
                    <div>Add and activate at least one service before creating manual appointments.</div>
                </div>
            {:else}
                <div class="booking-manual-grid">
                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-service">Service</label>
                        <select
                            id="booking-manual-service"
                            class="input input-sm"
                            value={manualAppointmentForm.serviceId}
                            disabled={isCreatingManualAppointment}
                            on:change={(event) => setManualAppointmentService(event.currentTarget.value)}
                        >
                            <option value="">Select service</option>
                            {#each manualAppointmentServiceOptions as service (service.id)}
                                <option value={service.id}>{service.label}</option>
                            {/each}
                        </select>
                    </div>

                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-date">Date</label>
                        <input
                            id="booking-manual-date"
                            class="input input-sm"
                            type="date"
                            value={manualAppointmentForm.date}
                            disabled={isCreatingManualAppointment}
                            on:change={(event) => setManualAppointmentDate(event.currentTarget.value)}
                        />
                    </div>

                    <div class="form-field m-b-0 booking-manual-slot-field">
                        <label class="txt-sm txt-hint">Time</label>
                        {#if !manualAppointmentForm.serviceId || !manualAppointmentForm.date}
                            <div class="txt-xs txt-hint">Select service and date to load available times.</div>
                        {:else if isLoadingManualAppointmentSlots}
                            <div class="txt-xs txt-hint">Loading available times...</div>
                        {:else if manualAppointmentSlotsError}
                            <div class="txt-xs txt-danger">{manualAppointmentSlotsError}</div>
                        {:else if !manualAppointmentAvailableSlots.length}
                            <div class="txt-xs txt-hint">No available times for this date.</div>
                        {:else}
                            <div class="booking-manual-slots">
                                {#each manualAppointmentAvailableSlots as slot}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm booking-manual-slot-btn"
                                        class:active={manualAppointmentForm.time === slot}
                                        disabled={isCreatingManualAppointment}
                                        on:click={() => selectManualAppointmentSlot(slot)}
                                    >
                                        <span class="txt">{slot}</span>
                                    </button>
                                {/each}
                            </div>
                        {/if}
                    </div>

                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-status">Status</label>
                        <select
                            id="booking-manual-status"
                            class="input input-sm"
                            bind:value={manualAppointmentForm.status}
                            disabled={isCreatingManualAppointment}
                        >
                            <option value="confirmed">Confirmed</option>
                            <option value="pending">Pending</option>
                        </select>
                    </div>
                </div>

                <div class="booking-manual-grid">
                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-name">Customer name</label>
                        <input
                            id="booking-manual-name"
                            class="input input-sm"
                            type="text"
                            placeholder="Customer name"
                            bind:value={manualAppointmentForm.name}
                            disabled={isCreatingManualAppointment}
                        />
                    </div>

                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-email">Customer email</label>
                        <input
                            id="booking-manual-email"
                            class="input input-sm"
                            type="email"
                            placeholder="customer@example.com"
                            bind:value={manualAppointmentForm.email}
                            disabled={isCreatingManualAppointment}
                        />
                    </div>

                    <div class="form-field m-b-0">
                        <label class="txt-sm txt-hint" for="booking-manual-phone">Customer phone</label>
                        <input
                            id="booking-manual-phone"
                            class="input input-sm"
                            type="text"
                            placeholder="Optional"
                            bind:value={manualAppointmentForm.phone}
                            disabled={isCreatingManualAppointment}
                        />
                    </div>
                </div>

                <div class="form-field m-b-0">
                    <label class="txt-sm txt-hint" for="booking-manual-notes">Customer notes</label>
                    <textarea
                        id="booking-manual-notes"
                        class="input booking-notes-input"
                        rows="4"
                        placeholder="Optional notes shared by the customer..."
                        bind:value={manualAppointmentForm.notes}
                        disabled={isCreatingManualAppointment}
                    />
                </div>

                <div class="form-field m-b-0">
                    <label class="txt-sm txt-hint" for="booking-manual-internal-notes">Internal notes</label>
                    <textarea
                        id="booking-manual-internal-notes"
                        class="input booking-notes-input"
                        rows="4"
                        placeholder="Private follow-up notes for the team..."
                        bind:value={manualAppointmentForm.internalNotes}
                        disabled={isCreatingManualAppointment}
                    />
                </div>
            {/if}

            {#if manualAppointmentFormError}
                <p class="txt-xs txt-danger m-b-0">{manualAppointmentFormError}</p>
            {/if}
        </div>

        <svelte:fragment slot="footer">
            <button type="button" class="btn btn-outline btn-sm" disabled={isCreatingManualAppointment} on:click={closeManualAppointmentPanel}>
                <span class="txt">Cancel</span>
            </button>
            <button
                type="button"
                class="btn btn-sm"
                class:btn-loading={isCreatingManualAppointment}
                disabled={isCreatingManualAppointment || isLoadingManualAppointmentSlots || !manualAppointmentServiceOptions.length}
                on:click={createManualAppointment}
            >
                <span class="txt">Create appointment</span>
            </button>
        </svelte:fragment>
    </OverlayPanel>
</PageWrapper>

<style>
    .booking-head.operations-head .head-description {
        max-width: 520px;
    }

    .booking-head.operations-head .head-main {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 20px;
        flex-wrap: wrap;
    }

    .booking-head.operations-head .head-selector {
        min-width: 280px;
        max-width: 420px;
        width: 100%;
    }

    .booking-head.operations-head .selector-row {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .booking-head.operations-head .selector-label {
        flex: 0 0 auto;
    }

    .booking-head.operations-head .selector-row .input {
        flex: 1;
        min-width: 0;
    }

    .booking-head.operations-head .head-tools {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        flex-wrap: wrap;
    }

    .booking-head.operations-head .head-tools > * {
        min-width: 0;
    }

    .booking-head-tabs {
        width: fit-content;
        margin-top: 0;
        flex: 0 0 auto;
    }

    .booking-head.operations-head .head-tools .booking-head-tabs {
        align-self: center;
    }

    .booking-head.operations-head .head-tools .summary-badges {
        margin-left: auto;
    }

    .booking-head.operations-head .summary-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        justify-content: flex-end;
    }

    .booking-body {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .booking-tabs {
        width: fit-content;
    }

    .booking-split-layout {
        display: grid;
        grid-template-columns: minmax(0, 1fr) 340px;
        gap: 14px;
        align-items: start;
    }

    .booking-main-column {
        min-width: 0;
    }

    .booking-rail {
        display: flex;
        flex-direction: column;
        gap: 10px;
        min-width: 0;
    }

    .booking-rail-block {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 12px;
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: var(--baseColor);
    }

    .booking-summary-grid {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .booking-summary-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 12px;
        padding-bottom: 8px;
        border-bottom: 1px dashed var(--baseAlt1);
    }

    .booking-summary-row:last-child {
        border-bottom: 0;
        padding-bottom: 0;
    }

    .booking-controls {
        display: grid;
        gap: 10px;
        margin-bottom: 12px;
    }

    .booking-controls--appointments {
        grid-template-columns: 1.6fr repeat(4, minmax(130px, 0.8fr)) auto;
    }

    .booking-control-cell {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .booking-control-cell--actions {
        display: flex;
        align-items: flex-end;
        justify-content: flex-end;
    }

    .booking-appointments-list {
        display: grid;
        gap: 8px;
    }

    .booking-appointment-item {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 10px 12px;
        display: flex;
        flex-direction: column;
        gap: 6px;
        transition: border-color 0.15s ease, background-color 0.15s ease;
        cursor: pointer;
    }

    .booking-appointment-item:hover {
        border-color: var(--baseAlt2);
    }

    .booking-appointment-item.selected {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 28%, transparent);
    }

    .booking-appointment-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .booking-appointment-title {
        font-weight: 600;
    }

    .booking-appointment-meta {
        display: inline-flex;
        align-items: center;
        gap: 6px;
    }

    .booking-actions-row {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .booking-notes-input {
        width: 100%;
        min-height: 108px;
        resize: vertical;
    }

    .booking-internal-notes-helper {
        margin-top: 4px;
    }

    .booking-readonly-note {
        white-space: pre-wrap;
        word-break: break-word;
    }

    .booking-empty-state {
        border: 1px dashed var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 14px;
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: color-mix(in srgb, var(--baseAlt1) 14%, transparent);
    }

    .booking-empty-readiness {
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin-top: 2px;
    }

    .booking-empty-readiness-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        padding-bottom: 8px;
        border-bottom: 1px dashed var(--baseAlt1);
    }

    .booking-empty-readiness-row:last-child {
        border-bottom: 0;
        padding-bottom: 0;
    }

    .booking-section-head-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 10px;
        margin-bottom: 12px;
    }

    .booking-section-head-copy {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .booking-section-head-meta,
    .booking-section-head-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        justify-content: flex-end;
    }

    .booking-services-list {
        display: grid;
        gap: 8px;
    }

    .booking-services-controls {
        display: grid;
        grid-template-columns: minmax(0, 1fr) 190px;
        gap: 10px;
        margin-bottom: 12px;
    }

    .booking-service-item {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 10px 12px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 12px;
        cursor: pointer;
        transition: border-color 0.15s ease, background-color 0.15s ease;
    }

    .booking-service-item:hover {
        border-color: var(--baseAlt2);
    }

    .booking-service-item.selected {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 28%, transparent);
    }

    .booking-service-title {
        font-weight: 600;
    }

    .booking-service-title-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
    }

    .booking-service-meta {
        margin-top: 4px;
    }

    .booking-service-actions {
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .booking-form-stack {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .booking-checkbox-row {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        margin: 0;
    }

    .booking-availability-layout {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .booking-split-layout--availability {
        grid-template-columns: minmax(0, 1fr) 360px;
    }

    .booking-availability-main {
        display: flex;
        flex-direction: column;
        gap: 10px;
        min-width: 0;
    }

    .booking-availability-list {
        display: grid;
        gap: 8px;
    }

    .booking-availability-actions-list {
        display: flex;
        flex-direction: column;
        flex-wrap: wrap;
        gap: 8px;
    }

    .booking-availability-helper {
        margin-top: -4px;
    }

    .booking-availability-row {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 8px 10px;
        display: grid;
        grid-template-columns: minmax(170px, 220px) minmax(190px, 280px);
        gap: 10px;
        align-items: center;
    }

    .booking-availability-row.is-inactive {
        background: color-mix(in srgb, var(--baseAlt1) 16%, transparent);
    }

    .booking-availability-day {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
    }

    .booking-availability-day-meta {
        display: inline-flex;
        align-items: center;
        gap: 8px;
    }

    .booking-availability-day-label {
        font-weight: 600;
    }

    .booking-unsaved-pill {
        font-size: 0.66rem;
        line-height: 1;
    }

    .booking-availability-time-range {
        display: grid;
        grid-template-columns: repeat(2, minmax(98px, 124px));
        gap: 6px;
    }

    .booking-availability-time-range .form-field--compact {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .booking-availability-error {
        grid-column: 1 / -1;
    }

    .booking-slot-preview-panel {
        border-style: dashed;
    }

    .booking-slot-preview-controls {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 10px;
    }

    .booking-slot-preview-slots {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .booking-slot-preview-pill {
        min-width: 58px;
        justify-content: center;
    }

    .booking-health-panel {
        border-style: dashed;
    }

    .booking-health-head {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .booking-health-main {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .booking-health-meta {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .booking-health-group-title {
        font-size: 0.7rem;
        line-height: 1.2;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        font-weight: 600;
        color: var(--txtHintColor);
        margin-bottom: 6px;
    }

    .booking-health-item {
        display: grid;
        grid-template-columns: auto minmax(0, 1fr);
        align-items: flex-start;
        gap: 8px;
        padding: 8px 0;
        border-top: 1px dashed var(--baseAlt1);
        font-size: 0.82rem;
    }

    .booking-health-item.warning {
        color: var(--txtPrimaryColor);
    }

    .booking-health-pill {
        align-self: start;
    }

    .booking-health-pill.warning {
        background: color-mix(in srgb, var(--dangerColor) 12%, transparent);
        color: var(--dangerColor);
    }

    .booking-manual-form {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .booking-manual-grid {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 10px;
    }

    .booking-manual-slot-field {
        grid-column: 1 / -1;
    }

    .booking-manual-slots {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .booking-manual-slot-btn.active {
        border-color: var(--txtPrimaryColor);
        background: color-mix(in srgb, var(--baseAlt1) 35%, transparent);
    }

    @media (max-width: 1280px) {
        .booking-controls--appointments {
            grid-template-columns: 1fr 1fr 1fr;
        }

        .booking-control-cell--search,
        .booking-control-cell--actions {
            grid-column: span 3;
        }

        .booking-control-cell--actions {
            justify-content: flex-start;
        }
    }

    @media (max-width: 1100px) {
        .booking-split-layout {
            grid-template-columns: 1fr;
        }

        .booking-rail {
            order: 2;
        }

        .booking-main-column {
            order: 1;
        }

        .booking-availability-row {
            grid-template-columns: 1fr;
        }

        .booking-availability-day {
            align-items: flex-start;
            flex-direction: column;
        }

        .booking-services-controls {
            grid-template-columns: 1fr;
        }

        .booking-slot-preview-controls {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 860px) {
        .booking-head.operations-head .head-tools {
            align-items: flex-start;
        }

        .booking-head.operations-head .head-tools .summary-badges {
            margin-left: 0;
        }

        .booking-head.operations-head .selector-row {
            flex-direction: column;
            align-items: flex-start;
        }

        .booking-head.operations-head .selector-row .input {
            width: 100%;
        }

        .booking-head.operations-head .summary-badges {
            justify-content: flex-start;
        }

        .booking-controls--appointments {
            grid-template-columns: 1fr;
        }

        .booking-control-cell--search,
        .booking-control-cell--actions {
            grid-column: auto;
        }

        .booking-control-cell--actions {
            align-items: flex-start;
        }

        .booking-section-head-row {
            flex-direction: column;
            align-items: flex-start;
        }

        .booking-section-head-meta,
        .booking-section-head-actions {
            justify-content: flex-start;
        }

        .booking-service-item {
            flex-direction: column;
            align-items: flex-start;
        }

        .booking-service-actions {
            width: 100%;
            justify-content: space-between;
        }

        .booking-manual-grid {
            grid-template-columns: 1fr;
        }
    }
</style>
