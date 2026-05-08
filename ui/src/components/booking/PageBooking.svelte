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
    const availabilityTabs = [
        { key: "weekly", label: "Weekly schedule", icon: "ri-calendar-week-line" },
        { key: "rulesExceptions", label: "Rules & exceptions", icon: "ri-settings-5-line" },
    ];
    const exceptionTypeFilterOptions = [
        { key: "all", label: "All" },
        { key: "closed", label: "Closed" },
        { key: "customHours", label: "Custom hours" },
    ];
    const exceptionActiveFilterOptions = [
        { key: "all", label: "All" },
        { key: "active", label: "Active" },
        { key: "inactive", label: "Inactive" },
    ];
    const servicePriorityOptions = [
        { key: "high", label: "High" },
        { key: "normal", label: "Normal" },
        { key: "low", label: "Low" },
    ];

    const appointmentStatusFieldAliases = ["status"];
    const appointmentCustomerNotesFieldAliases = ["notes", "note"];
    const appointmentInternalNotesFieldAliases = ["internalNotes", "internal_notes"];
    const appointmentConfirmedAtFieldAliases = ["confirmedAt", "confirmed_at"];
    const appointmentCancelledAtFieldAliases = ["cancelledAt", "cancelled_at"];
    const appointmentRescheduledAtFieldAliases = ["rescheduledAt", "rescheduled_at"];
    const appointmentServiceNameSnapshotFieldAliases = ["serviceNameSnapshot", "service_name_snapshot"];
    const appointmentServiceDurationSnapshotFieldAliases = ["serviceDurationMinutesSnapshot", "service_duration_minutes_snapshot"];
    const appointmentServiceDescriptionSnapshotFieldAliases = ["serviceDescriptionSnapshot", "service_description_snapshot"];
    const bookingDatePattern = /^\d{4}-\d{2}-\d{2}$/;
    const bookingTimePattern = /^([01]\d|2[0-3]):[0-5]\d$/;
    const bookingEmailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const bookingConfirmationModes = new Set(["request", "autoConfirm"]);
    const bookingCalendarBlockingModes = new Set(["service", "website", "none"]);

    let activeTab = "appointments";
    let activeAvailabilityTab = "weekly";

    let websites = [];
    let selectedWebsiteId = "";

    let servicesRecords = [];
    let availabilityRecords = [];
    let exceptionsRecords = [];
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
    let sendConfirmationEmailOnConfirm = false;
    let sendConfirmationEmailTargetAppointmentId = "";
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
    let isReschedulePanelOpen = false;
    let isSavingRescheduleAppointment = false;
    let isLoadingRescheduleSlots = false;
    let rescheduleFormError = "";
    let rescheduleSlotsError = "";
    let rescheduleAvailableSlots = [];
    let lastRescheduleSlotsQueryKey = "";
    let rescheduleSourceAppointment = null;
    let rescheduleSourceWebsiteId = "";
    let rescheduleForm = createDefaultRescheduleForm();

    let selectedServiceId = "";
    let isCreatingService = false;
    let isSavingService = false;
    let serviceSearch = "";
    let serviceStatusFilter = "all";
    let serviceForm = {
        id: "",
        name: "",
        durationMinutes: "30",
        description: "",
        priority: "normal",
        active: true,
    };
    let serviceFormError = "";

    let availabilityRows = createDefaultAvailabilityRows();
    let availabilityWindowCounter = 0;
    let isSavingAvailability = {};
    let isSavingAllAvailability = false;
    let slotPreviewServiceId = "";
    let slotPreviewDate = "";
    let slotPreviewSlots = [];
    let slotPreviewError = "";
    let isLoadingSlotPreview = false;
    let lastSlotPreviewQueryKey = "";
    let selectedExceptionId = "";
    let isCreatingException = false;
    let exceptionSearch = "";
    let exceptionTypeFilter = "all";
    let exceptionActiveFilter = "all";
    let exceptionForm = createDefaultExceptionForm();
    let exceptionFormError = "";
    let isSavingException = false;

    let bookingRulesDraft = createDefaultBookingRulesDraft();
    let bookingRulesDraftWebsiteId = "";
    let bookingRulesFormError = "";
    let isSavingBookingRules = false;

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: bookingServicesCollection = resolveCollectionByAliases(["bookingservices"]);
    $: bookingAvailabilityCollection = resolveCollectionByAliases(["bookingavailability"]);
    $: bookingExceptionsCollection = resolveCollectionByAliases(["bookingexceptions"]);
    $: appointmentsCollection = resolveCollectionByAliases(["appointments"]);
    $: websiteSettingsFieldName = resolveCollectionFieldNameByAliases(websitesCollection, ["settings"]) || "settings";

    $: hasBookingCollections = !!bookingServicesCollection?.id
        && !!bookingAvailabilityCollection?.id
        && !!appointmentsCollection?.id;

    $: appointmentStatusFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentStatusFieldAliases) || "status";
    $: appointmentCustomerNotesFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentCustomerNotesFieldAliases) || "notes";
    $: appointmentInternalNotesFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentInternalNotesFieldAliases) || "internalNotes";
    $: appointmentConfirmedAtFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentConfirmedAtFieldAliases) || "";
    $: appointmentCancelledAtFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentCancelledAtFieldAliases) || "";
    $: appointmentRescheduledAtFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentRescheduledAtFieldAliases) || "";
    $: appointmentServiceNameSnapshotFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentServiceNameSnapshotFieldAliases) || "";
    $: appointmentServiceDurationSnapshotFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentServiceDurationSnapshotFieldAliases) || "";
    $: appointmentServiceDescriptionSnapshotFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentServiceDescriptionSnapshotFieldAliases) || "";

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
        bookingExceptionsCollection?.id || "",
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
    $: selectedWebsiteBookingSettings = resolveWebsiteBookingSettings(selectedWebsite?.[websiteSettingsFieldName]);

    $: normalizedAppointments = appointmentRecords.map((record) => normalizeAppointment(record));

    $: pendingAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "pending").length;
    $: confirmedAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "confirmed").length;
    $: thisWeekAppointmentsCount = normalizedAppointments.filter((appointment) => isInCurrentWeek(appointment.date)).length;
    $: activeServicesCount = servicesRecords.filter((service) => !!service?.active).length;
    $: activeAvailabilityDaysCount = countActiveAvailabilityDays(availabilityRows);
    $: normalizedExceptions = normalizeExceptionsRecords(exceptionsRecords);
    $: activeExceptionsCount = normalizedExceptions.filter((exception) => !!exception.active).length;
    $: bookingRulesConfiguredCount = [
        Number(selectedWebsiteBookingSettings?.rules?.minNoticeHours || 0),
        Number(selectedWebsiteBookingSettings?.rules?.bookingWindowDays || 0),
        Number(selectedWebsiteBookingSettings?.rules?.bufferMinutes || 0),
    ].filter((value) => Number.isFinite(value) && value > 0).length;
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
    $: availabilityValidationIssuesCount = countAvailabilityValidationIssues(availabilityRows);
    $: availabilityHealthWarnings = buildAvailabilityHealthWarnings({
        bookingWarnings: bookingReadinessWarnings,
        availabilityValidationIssuesCount,
        dirtyAvailabilityRowsCount,
        activeExceptionsCount,
        bookingRulesConfiguredCount,
    });
    $: availabilityHealthSuggestions = buildAvailabilityHealthSuggestions({
        bookingSuggestions: bookingReadinessSuggestions,
        dirtyAvailabilityRowsCount,
        availabilityValidationIssuesCount,
        activeExceptionsCount,
        bookingRulesConfiguredCount,
    });
    $: availabilityHealthState = resolveBookingReadinessState(availabilityHealthWarnings.length);
    $: normalizedServices = sortServicesForBackoffice(
        servicesRecords.map((service) => normalizeServiceRecord(service)),
    );
    $: normalizedServiceSearch = normalizeLower(serviceSearch);
    $: filteredServices = normalizedServices.filter((service) => {
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
                normalizeString(service?.description),
                normalizeString(service?.priorityLabel),
                normalizeString(service?.priorityKey),
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
    $: dirtyAvailabilityRowsCount = countDirtyAvailabilityDays(availabilityRows);
    $: dirtyAvailabilityWindowsCount = countDirtyAvailabilityWindows(availabilityRows);
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
    $: rescheduleSlotsQueryKey = isReschedulePanelOpen
        ? `${selectedWebsiteId}:${rescheduleForm.appointmentId}:${rescheduleForm.serviceId}:${rescheduleForm.date}`
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
    $: if (
        isReschedulePanelOpen
        && manualAppointmentServiceOptions.length
        && !manualAppointmentServiceOptions.some((service) => service.id === rescheduleForm.serviceId)
    ) {
        rescheduleForm = {
            ...rescheduleForm,
            serviceId: manualAppointmentServiceOptions[0].id,
            time: "",
        };
        rescheduleSlotsError = "";
        rescheduleFormError = "";
    }
    $: if (rescheduleSlotsQueryKey !== lastRescheduleSlotsQueryKey) {
        lastRescheduleSlotsQueryKey = rescheduleSlotsQueryKey;
        loadRescheduleAppointmentSlots();
    }
    $: if (isReschedulePanelOpen && rescheduleSourceWebsiteId && selectedWebsiteId !== rescheduleSourceWebsiteId) {
        closeReschedulePanel();
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
    $: if (selectedWebsiteId !== bookingRulesDraftWebsiteId) {
        bookingRulesDraftWebsiteId = selectedWebsiteId;
        bookingRulesDraft = createBookingRulesDraftFromSettings(selectedWebsiteBookingSettings?.rules);
        bookingRulesFormError = "";
    }
    $: bookingRulesDirty = (() => {
        const source = selectedWebsiteBookingSettings?.rules || {};
        const current = createBookingRulesDraftFromSettings(source);
        return current.minNoticeHours !== normalizeString(bookingRulesDraft.minNoticeHours)
            || current.bookingWindowDays !== normalizeString(bookingRulesDraft.bookingWindowDays)
            || current.bufferMinutes !== normalizeString(bookingRulesDraft.bufferMinutes);
    })();
    $: if (!isCreatingException && normalizedExceptions.length) {
        const hasSelectedException = selectedExceptionId
            && normalizedExceptions.some((exception) => exception.id === selectedExceptionId);

        if (!hasSelectedException) {
            selectedExceptionId = normalizedExceptions[0].id;
        }
    } else if (!normalizedExceptions.length && selectedExceptionId) {
        selectedExceptionId = "";
    }
    $: selectedException = normalizedExceptions.find((exception) => exception.id === selectedExceptionId) || null;
    $: filteredExceptions = sortExceptions(
        normalizedExceptions.filter((exception) => {
            if (exceptionTypeFilter !== "all" && exception.type !== exceptionTypeFilter) {
                return false;
            }

            if (exceptionActiveFilter === "active" && !exception.active) {
                return false;
            }

            if (exceptionActiveFilter === "inactive" && exception.active) {
                return false;
            }

            const normalizedQuery = normalizeLower(exceptionSearch);
            if (normalizedQuery) {
                const searchable = [exception.date, exception.note, exception.typeLabel]
                    .filter(Boolean)
                    .join(" ")
                    .toLowerCase();
                if (!searchable.includes(normalizedQuery)) {
                    return false;
                }
            }

            return true;
        }),
    );
    $: if (selectedException?.id) {
        if (exceptionForm.id !== selectedException.id) {
            exceptionForm = createExceptionFormFromRecord(selectedException);
            exceptionFormError = "";
        }
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
    $: selectedAppointmentDurationMinutes = resolveAppointmentDurationMinutes(selectedAppointment);
    $: selectedAppointmentGoogleCalendar = buildGoogleCalendarEventForAppointment(
        selectedAppointment,
        selectedWebsite,
        selectedAppointmentDurationMinutes,
    );

    $: canSetSelectedAppointmentPending = !!selectedAppointment
        && selectedAppointmentStatusKey !== "pending";
    $: canSetSelectedAppointmentConfirmed = !!selectedAppointment
        && selectedAppointmentStatusKey === "pending";
    $: canSetSelectedAppointmentCancelled = !!selectedAppointment
        && (selectedAppointmentStatusKey === "pending" || selectedAppointmentStatusKey === "confirmed");
    $: canRescheduleSelectedAppointment = !!selectedAppointment
        && (selectedAppointmentStatusKey === "pending" || selectedAppointmentStatusKey === "confirmed");
    $: if (selectedAppointment?.id) {
        if (sendConfirmationEmailTargetAppointmentId !== selectedAppointment.id) {
            sendConfirmationEmailTargetAppointmentId = selectedAppointment.id;
            sendConfirmationEmailOnConfirm = false;
        }
    } else if (sendConfirmationEmailTargetAppointmentId || sendConfirmationEmailOnConfirm) {
        sendConfirmationEmailTargetAppointmentId = "";
        sendConfirmationEmailOnConfirm = false;
    }

    $: if (selectedAppointment?.id) {
        if (appointmentInternalNotesDraftId !== selectedAppointment.id) {
            appointmentInternalNotesDraftId = selectedAppointment.id;
            appointmentInternalNotesDraft = selectedAppointment.internalNotes || "";
        }
    } else if (appointmentInternalNotesDraftId || appointmentInternalNotesDraft) {
        appointmentInternalNotesDraftId = "";
        appointmentInternalNotesDraft = "";
    }

    $: selectedService = normalizedServices.find((service) => service.id === selectedServiceId) || null;

    $: if (normalizedServices.length && !selectedServiceId && !serviceForm.id && !isCreatingService) {
        selectedServiceId = normalizedServices[0].id;
    }

    $: if (selectedService) {
        if (serviceForm.id !== selectedService.id) {
            setServiceFormFromRecord(selectedService);
        }
    }

    $: serviceFormDurationValue = Number.parseInt(`${serviceForm.durationMinutes || ""}`.trim(), 10);
    $: serviceFormDurationValid = Number.isFinite(serviceFormDurationValue)
        && serviceFormDurationValue >= 5
        && serviceFormDurationValue <= 480;
    $: serviceFormPriorityKey = normalizeServicePriority(serviceForm.priority);
    $: serviceHealthWarnings = buildServiceHealthWarnings({
        name: serviceForm.name,
        durationValid: serviceFormDurationValid,
    });
    $: serviceHealthSuggestions = buildServiceHealthSuggestions({
        active: !!serviceForm.active,
        description: serviceForm.description,
        priority: serviceFormPriorityKey,
        servicesCount: normalizedServices.length,
    });
    $: serviceHealthState = resolveBookingReadinessState(serviceHealthWarnings.length);

    function normalizeString(value) {
        return `${value || ""}`.trim();
    }

    function normalizeLower(value) {
        return normalizeString(value).toLowerCase();
    }

    function normalizeServiceRecord(record) {
        const displayOrder = readNonNegativeInteger(record?.displayOrder, 50);
        const priorityKey = mapDisplayOrderToPriority(displayOrder);
        return {
            ...record,
            id: normalizeString(record?.id),
            name: normalizeString(record?.name),
            description: normalizeString(record?.description),
            durationMinutes: readNonNegativeInteger(record?.durationMinutes, 0),
            displayOrder,
            priorityKey,
            priorityLabel: formatServicePriorityLabel(priorityKey),
            active: !!record?.active,
            created: normalizeString(record?.created),
        };
    }

    function sortServicesForBackoffice(list = []) {
        return [...list].sort((a, b) => {
            const firstOrder = readNonNegativeInteger(a?.displayOrder, 50);
            const secondOrder = readNonNegativeInteger(b?.displayOrder, 50);
            if (firstOrder !== secondOrder) {
                return firstOrder - secondOrder;
            }

            const firstName = normalizeLower(a?.name);
            const secondName = normalizeLower(b?.name);
            if (firstName !== secondName) {
                return firstName.localeCompare(secondName);
            }

            const firstCreated = normalizeString(a?.created);
            const secondCreated = normalizeString(b?.created);
            if (firstCreated && secondCreated && firstCreated !== secondCreated) {
                return firstCreated.localeCompare(secondCreated);
            }

            return normalizeString(a?.id).localeCompare(normalizeString(b?.id));
        });
    }

    function normalizeServicePriority(value) {
        const normalized = normalizeLower(value);
        if (normalized === "high" || normalized === "normal" || normalized === "low") {
            return normalized;
        }
        return "normal";
    }

    function mapPriorityToDisplayOrder(priority) {
        const normalized = normalizeServicePriority(priority);
        if (normalized === "high") {
            return 0;
        }
        if (normalized === "low") {
            return 100;
        }
        return 50;
    }

    function mapDisplayOrderToPriority(displayOrderValue) {
        const normalized = readNonNegativeInteger(displayOrderValue, 50);
        if (normalized <= 24) {
            return "high";
        }
        if (normalized >= 75) {
            return "low";
        }
        return "normal";
    }

    function formatServicePriorityLabel(priority) {
        const normalized = normalizeServicePriority(priority);
        if (normalized === "high") {
            return "High priority";
        }
        if (normalized === "low") {
            return "Low priority";
        }
        return "Normal priority";
    }

    function buildServiceHealthWarnings({
        name = "",
        durationValid = true,
    } = {}) {
        const warnings = [];

        if (!normalizeString(name)) {
            warnings.push("Service name is required.");
        }

        if (!durationValid) {
            warnings.push("Duration must be an integer between 5 and 480 minutes.");
        }

        return warnings;
    }

    function buildServiceHealthSuggestions({
        active = true,
        description = "",
        priority = "normal",
        servicesCount = 0,
    } = {}) {
        const suggestions = [];

        if (!active) {
            suggestions.push("Inactive services are hidden from new booking requests.");
        }

        if (!normalizeString(description)) {
            suggestions.push("Add a short description to help visitors choose this service.");
        }

        if (servicesCount > 1) {
            suggestions.push("Use priority to control which services appear first.");
            if (normalizeServicePriority(priority) !== "high") {
                suggestions.push("High priority services appear before normal and low priority services.");
            }
        }

        return suggestions;
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

    function readNonNegativeInteger(value, fallback = 0) {
        if (typeof value === "number" && Number.isFinite(value)) {
            return Math.max(0, Math.trunc(value));
        }

        if (typeof value === "string") {
            const normalized = value.trim();
            if (!normalized) {
                return fallback;
            }

            if (!/^-?\d+$/.test(normalized)) {
                return fallback;
            }

            const parsed = Number.parseInt(normalized, 10);
            if (!Number.isFinite(parsed)) {
                return fallback;
            }

            return Math.max(0, parsed);
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

    function normalizeBookingConfirmationMode(value) {
        if (typeof value !== "string") {
            return "request";
        }

        const normalized = value.trim();
        if (bookingConfirmationModes.has(normalized)) {
            return normalized;
        }

        if (normalized.toLowerCase() === "autoconfirm") {
            return "autoConfirm";
        }

        return "request";
    }

    function normalizeBookingCalendarBlockingMode(value) {
        if (typeof value !== "string") {
            return "service";
        }

        const normalized = value.trim();
        if (bookingCalendarBlockingModes.has(normalized)) {
            return normalized;
        }

        return "service";
    }

    function parseBookingRules(rawRules) {
        const source = readObject(rawRules);
        return {
            minNoticeHours: readNonNegativeInteger(source.minNoticeHours, 0),
            bookingWindowDays: readNonNegativeInteger(source.bookingWindowDays, 0),
            bufferMinutes: readNonNegativeInteger(source.bufferMinutes, 0),
            calendarBlockingMode: normalizeBookingCalendarBlockingMode(source.calendarBlockingMode),
        };
    }

    function resolveWebsiteBookingSettings(rawSettings) {
        const settings = parseSettingsObject(rawSettings);
        const featureFlags = readObject(settings.featureFlags);
        const booking = readObject(settings.booking);
        const contactForm = readObject(settings.contactForm);

        const bookingFeatureAvailable = readBoolean(featureFlags.booking, true);
        const bookingEnabled = readBoolean(booking.enabled, true);
        const confirmationMode = normalizeBookingConfirmationMode(booking.confirmationMode);
        const rules = parseBookingRules(booking.rules);

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
            confirmationMode,
            rules,
            bookingNotifications,
            contactNotifications,
            effectiveNotifications,
            usingContactFormFallback,
            businessNotificationsReady,
        };
    }

    function createDefaultExceptionForm() {
        return {
            id: "",
            date: "",
            type: "closed",
            startTime: "09:00",
            endTime: "17:00",
            note: "",
            active: true,
        };
    }

    function normalizeExceptionType(value) {
        const normalized = normalizeLower(value);
        if (normalized === "customhours") {
            return "customHours";
        }
        return "closed";
    }

    function normalizeExceptionRecord(record) {
        const type = normalizeExceptionType(record?.type);
        const startTime = normalizeString(record?.startTime);
        const endTime = normalizeString(record?.endTime);
        const note = normalizeString(record?.note);
        const active = !!record?.active;

        return {
            ...record,
            id: normalizeString(record?.id),
            date: normalizeString(record?.date),
            type,
            typeLabel: type === "customHours" ? "Custom hours" : "Closed",
            startTime,
            endTime,
            note,
            active,
            timeRangeLabel: type === "customHours" && startTime && endTime
                ? `${startTime} - ${endTime}`
                : "Closed all day",
        };
    }

    function normalizeExceptionsRecords(records = []) {
        if (!Array.isArray(records)) {
            return [];
        }

        return records
            .map((record) => normalizeExceptionRecord(record))
            .filter((record) => !!record.id);
    }

    function sortExceptions(list = []) {
        return [...list].sort((a, b) => {
            const firstDate = normalizeString(a?.date);
            const secondDate = normalizeString(b?.date);

            if (firstDate !== secondDate) {
                return secondDate.localeCompare(firstDate);
            }

            return toTimestamp(b?.updated || b?.created) - toTimestamp(a?.updated || a?.created);
        });
    }

    function createExceptionFormFromRecord(record) {
        const normalized = normalizeExceptionRecord(record || {});
        return {
            id: normalized.id,
            date: normalized.date,
            type: normalized.type,
            startTime: normalized.startTime || "09:00",
            endTime: normalized.endTime || "17:00",
            note: normalized.note,
            active: normalized.active,
        };
    }

    function createDefaultBookingRulesDraft() {
        return {
            minNoticeHours: "0",
            bookingWindowDays: "0",
            bufferMinutes: "0",
        };
    }

    function createBookingRulesDraftFromSettings(rules = {}) {
        const source = isPlainObject(rules) ? rules : {};
        return {
            minNoticeHours: `${readNonNegativeInteger(source.minNoticeHours, 0)}`,
            bookingWindowDays: `${readNonNegativeInteger(source.bookingWindowDays, 0)}`,
            bufferMinutes: `${readNonNegativeInteger(source.bufferMinutes, 0)}`,
        };
    }

    function parseRulesDraftField(value, label) {
        const normalized = normalizeString(value);
        if (!normalized) {
            return 0;
        }

        if (!/^-?\d+$/.test(normalized)) {
            throw new Error(`${label} must be a whole number.`);
        }

        const parsed = Number.parseInt(normalized, 10);
        if (!Number.isFinite(parsed)) {
            throw new Error(`${label} must be a whole number.`);
        }

        return Math.max(0, parsed);
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

    function createDefaultRescheduleForm() {
        return {
            appointmentId: "",
            serviceId: "",
            date: "",
            time: "",
            sendEmail: false,
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

    function resetRescheduleForm() {
        rescheduleForm = createDefaultRescheduleForm();
        rescheduleFormError = "";
        rescheduleSlotsError = "";
        rescheduleAvailableSlots = [];
        lastRescheduleSlotsQueryKey = "";
    }

    function openReschedulePanel() {
        if (!selectedAppointment?.id) {
            return;
        }

        if (!canRescheduleSelectedAppointment) {
            addErrorToast("Only pending or confirmed appointments can be rescheduled.");
            return;
        }

        const defaultServiceId = normalizeString(selectedAppointment.serviceId);
        const fallbackServiceId = manualAppointmentServiceOptions[0]?.id || "";
        const hasDefaultService = !!defaultServiceId
            && manualAppointmentServiceOptions.some((service) => service.id === defaultServiceId);
        const initialServiceId = hasDefaultService ? defaultServiceId : fallbackServiceId;
        rescheduleSourceAppointment = {
            id: selectedAppointment.id,
            serviceLabel: selectedAppointment.serviceLabel,
            date: selectedAppointment.date,
            time: selectedAppointment.time,
        };
        rescheduleSourceWebsiteId = selectedWebsiteId;
        resetRescheduleForm();
        rescheduleForm = {
            appointmentId: selectedAppointment.id,
            serviceId: initialServiceId,
            date: normalizeString(selectedAppointment.date),
            time: "",
            sendEmail: false,
        };
        isReschedulePanelOpen = true;
    }

    function closeReschedulePanel() {
        isReschedulePanelOpen = false;
        isSavingRescheduleAppointment = false;
        isLoadingRescheduleSlots = false;
        rescheduleSourceAppointment = null;
        rescheduleSourceWebsiteId = "";
        resetRescheduleForm();
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

    function setRescheduleService(serviceId) {
        rescheduleForm = {
            ...rescheduleForm,
            serviceId: normalizeString(serviceId),
            time: "",
        };
        rescheduleFormError = "";
        rescheduleSlotsError = "";
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

    function setRescheduleDate(dateValue) {
        rescheduleForm = {
            ...rescheduleForm,
            date: normalizeString(dateValue),
            time: "",
        };
        rescheduleFormError = "";
        rescheduleSlotsError = "";
    }

    function selectManualAppointmentSlot(slot) {
        manualAppointmentForm = {
            ...manualAppointmentForm,
            time: normalizeString(slot),
        };
        manualAppointmentFormError = "";
    }

    function selectRescheduleSlot(slot) {
        rescheduleForm = {
            ...rescheduleForm,
            time: normalizeString(slot),
        };
        rescheduleFormError = "";
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

    function readRecordStringByAliases(record, resolvedFieldName, aliases = []) {
        const candidates = [resolvedFieldName, ...aliases]
            .map((candidate) => normalizeString(candidate))
            .filter(Boolean);

        for (const fieldName of candidates) {
            const value = normalizeString(record?.[fieldName]);
            if (value) {
                return value;
            }
        }

        return "";
    }

    function readRecordPositiveIntegerByAliases(record, resolvedFieldName, aliases = []) {
        const candidates = [resolvedFieldName, ...aliases]
            .map((candidate) => normalizeString(candidate))
            .filter(Boolean);

        for (const fieldName of candidates) {
            const value = readNonNegativeInteger(record?.[fieldName], -1);
            if (value > 0) {
                return value;
            }
        }

        return 0;
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
            exceptionsRecords = [];
            appointmentRecords = [];
            availabilityWindowCounter = 0;
            availabilityRows = createDefaultAvailabilityRows();
            isSavingAvailability = {};
            selectedExceptionId = "";
            isCreatingException = false;
            exceptionForm = createDefaultExceptionForm();
            exceptionFormError = "";
            return;
        }

        isLoadingBookingData = true;
        bookingLoadError = "";

        try {
            const filter = `website="${selectedWebsiteId}"`;
            const [services, availability, exceptions, appointments] = await Promise.all([
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
                bookingExceptionsCollection?.id
                    ? ApiClient.collection(bookingExceptionsCollection.id).getFullList({
                        filter,
                        sort: "-date,-updated,-created",
                        requestKey: `nuvio_booking_exceptions_${selectedWebsiteId}`,
                    })
                    : Promise.resolve([]),
                ApiClient.collection(appointmentsCollection.id).getFullList({
                    filter,
                    sort: "-created",
                    expand: "service",
                    requestKey: `nuvio_booking_appointments_${selectedWebsiteId}`,
                }),
            ]);

            servicesRecords = services;
            availabilityRecords = availability;
            exceptionsRecords = Array.isArray(exceptions) ? exceptions : [];
            appointmentRecords = appointments;
            availabilityWindowCounter = 0;
            availabilityRows = createAvailabilityRowsFromRecords(availabilityRecords);
            selectedExceptionId = "";
            exceptionForm = createDefaultExceptionForm();
            exceptionFormError = "";
            isCreatingException = false;
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
            exceptionsRecords = [];
            appointmentRecords = [];
            availabilityWindowCounter = 0;
            availabilityRows = createDefaultAvailabilityRows();
            isSavingAvailability = {};
            isCreatingException = false;
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

    async function loadRescheduleAppointmentSlots() {
        if (!isReschedulePanelOpen) {
            rescheduleAvailableSlots = [];
            rescheduleSlotsError = "";
            return;
        }

        const appointmentId = normalizeString(rescheduleForm.appointmentId);
        const serviceId = normalizeString(rescheduleForm.serviceId);
        const dateValue = normalizeString(rescheduleForm.date);
        const hasValidService = manualAppointmentServiceOptions.some((service) => service.id === serviceId);
        const requestScopeKey = `${selectedWebsiteId}:${appointmentId}:${serviceId}:${dateValue}`;

        if (!selectedWebsiteId || !appointmentId || !serviceId || !hasValidService || !bookingDatePattern.test(dateValue)) {
            rescheduleAvailableSlots = [];
            rescheduleSlotsError = "";
            isLoadingRescheduleSlots = false;
            return;
        }

        isLoadingRescheduleSlots = true;
        rescheduleSlotsError = "";

        try {
            const query = new URLSearchParams({
                websiteId: selectedWebsiteId,
                serviceId,
                date: dateValue,
            });

            const response = await ApiClient.send(`/api/nuvio/booking/slots?${query.toString()}`, {
                method: "GET",
                requestKey: `nuvio_booking_reschedule_slots_${selectedWebsiteId}_${appointmentId}_${serviceId}_${dateValue}`,
            });

            const slots = Array.isArray(response?.slots)
                ? response.slots.map((slot) => normalizeString(slot)).filter((slot) => bookingTimePattern.test(slot))
                : [];

            if (!isReschedulePanelOpen || rescheduleSlotsQueryKey !== requestScopeKey) {
                return;
            }

            rescheduleAvailableSlots = slots;

            if (!slots.includes(normalizeString(rescheduleForm.time))) {
                rescheduleForm = {
                    ...rescheduleForm,
                    time: "",
                };
            }
        } catch (err) {
            ApiClient.error(err, false);

            if (!isReschedulePanelOpen || rescheduleSlotsQueryKey !== requestScopeKey) {
                return;
            }

            rescheduleAvailableSlots = [];
            rescheduleSlotsError = "Unable to load available times right now.";
            rescheduleForm = {
                ...rescheduleForm,
                time: "",
            };
        } finally {
            if (isReschedulePanelOpen && rescheduleSlotsQueryKey === requestScopeKey) {
                isLoadingRescheduleSlots = false;
            }
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

    function validateRescheduleForm() {
        const appointmentId = normalizeString(rescheduleForm.appointmentId);
        const serviceId = normalizeString(rescheduleForm.serviceId);
        const dateValue = normalizeString(rescheduleForm.date);
        const timeValue = normalizeString(rescheduleForm.time);
        const hasValidService = manualAppointmentServiceOptions.some((service) => service.id === serviceId);

        if (!selectedWebsiteId) {
            return "Select a website before rescheduling.";
        }

        if (!appointmentId) {
            return "Select an appointment to reschedule.";
        }

        if (!serviceId || !hasValidService) {
            return "Service is required.";
        }

        if (!bookingDatePattern.test(dateValue)) {
            return "Date is required.";
        }

        if (!timeValue || !bookingTimePattern.test(timeValue)) {
            return "Select an available time.";
        }

        if (
            Array.isArray(rescheduleAvailableSlots)
            && rescheduleAvailableSlots.length
            && !rescheduleAvailableSlots.includes(timeValue)
        ) {
            return "Select an available time.";
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

    async function saveRescheduledAppointment() {
        if (isSavingRescheduleAppointment || isLoadingRescheduleSlots) {
            return;
        }

        const validationError = validateRescheduleForm();
        if (validationError) {
            rescheduleFormError = validationError;
            return;
        }

        const appointmentId = normalizeString(rescheduleForm.appointmentId);

        isSavingRescheduleAppointment = true;
        rescheduleFormError = "";

        const payload = {
            serviceId: normalizeString(rescheduleForm.serviceId),
            date: normalizeString(rescheduleForm.date),
            time: normalizeString(rescheduleForm.time),
            sendEmail: !!rescheduleForm.sendEmail,
        };

        try {
            const response = await ApiClient.send(`/api/nuvio/booking/admin/appointments/${encodeURIComponent(appointmentId)}/reschedule`, {
                method: "POST",
                body: payload,
                requestKey: `nuvio_booking_reschedule_${selectedWebsiteId}_${appointmentId}`,
            });

            await loadBookingData();
            selectedAppointmentId = appointmentId;
            closeReschedulePanel();
            addSuccessToast("Appointment rescheduled.");

            if (normalizeString(response?.warning)) {
                addErrorToast(normalizeString(response.warning));
            }
        } catch (err) {
            ApiClient.error(err, false);

            const statusCode = (err?.status << 0) || 0;
            const conflictMessage = "This time is no longer available. Please choose another time.";
            if (statusCode === 409) {
                rescheduleFormError = conflictMessage;
                addErrorToast(conflictMessage);
                await loadRescheduleAppointmentSlots();
            } else {
                addErrorToast("Unable to reschedule appointment right now.");
            }
        } finally {
            isSavingRescheduleAppointment = false;
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

    function resolveAppointmentDurationMinutes(appointment) {
        if (!appointment) {
            return 0;
        }

        const serviceId = normalizeString(appointment.serviceId);
        const serviceRecord = servicesRecords.find((service) => normalizeString(service?.id) === serviceId);
        const candidates = [
            appointment?.serviceDurationMinutes,
            serviceRecord?.durationMinutes,
            appointment?.expand?.service?.durationMinutes,
            Array.isArray(appointment?.expand?.service) ? appointment.expand.service[0]?.durationMinutes : "",
        ];

        for (const candidate of candidates) {
            const parsed = Number.parseInt(`${candidate || ""}`.trim(), 10);
            if (Number.isFinite(parsed) && parsed > 0) {
                return parsed;
            }
        }

        return 0;
    }

    function parseGoogleCalendarDateTime(dateValue, timeValue) {
        const dateMatch = normalizeString(dateValue).match(/^(\d{4})-(\d{2})-(\d{2})$/);
        const timeMatch = normalizeString(timeValue).match(/^([01]\d|2[0-3]):([0-5]\d)$/);
        if (!dateMatch || !timeMatch) {
            return null;
        }

        const year = Number.parseInt(dateMatch[1], 10);
        const month = Number.parseInt(dateMatch[2], 10);
        const day = Number.parseInt(dateMatch[3], 10);
        const hour = Number.parseInt(timeMatch[1], 10);
        const minute = Number.parseInt(timeMatch[2], 10);
        if (!Number.isFinite(year) || !Number.isFinite(month) || !Number.isFinite(day) || !Number.isFinite(hour) || !Number.isFinite(minute)) {
            return null;
        }

        return new Date(Date.UTC(year, month - 1, day, hour, minute, 0, 0));
    }

    function formatGoogleCalendarDateTimeValue(value) {
        if (!(value instanceof Date) || Number.isNaN(value.getTime())) {
            return "";
        }

        const year = `${value.getUTCFullYear()}`;
        const month = `${value.getUTCMonth() + 1}`.padStart(2, "0");
        const day = `${value.getUTCDate()}`.padStart(2, "0");
        const hours = `${value.getUTCHours()}`.padStart(2, "0");
        const minutes = `${value.getUTCMinutes()}`.padStart(2, "0");
        const seconds = `${value.getUTCSeconds()}`.padStart(2, "0");
        return `${year}${month}${day}T${hours}${minutes}${seconds}`;
    }

    function resolveGoogleCalendarLocation(website) {
        const source = isPlainObject(website) ? website : {};

        const directCandidates = [
            source.businessAddress,
            source.address,
            source.addressLine1,
            source.localBusinessAddress,
            source.local_business_address,
            source.localBusinessStreetAddress,
        ]
            .map((value) => normalizeString(value))
            .filter(Boolean);

        if (directCandidates.length) {
            return directCandidates[0];
        }

        const localParts = [
            source.localBusinessStreet,
            source.local_business_street,
            source.localBusinessCity,
            source.local_business_city,
            source.localBusinessRegion,
            source.local_business_region,
            source.localBusinessPostalCode,
            source.local_business_postal_code,
            source.localBusinessCountry,
            source.local_business_country,
        ]
            .map((value) => normalizeString(value))
            .filter(Boolean);

        return localParts.join(", ");
    }

    function buildGoogleCalendarEventForAppointment(appointment, website, durationMinutes) {
        const base = {
            available: false,
            href: "",
            helper: "",
        };

        if (!appointment) {
            return base;
        }

        const duration = Number.parseInt(`${durationMinutes || 0}`, 10);
        if (!Number.isFinite(duration) || duration <= 0) {
            return {
                ...base,
                helper: "Service duration is required to add this appointment to Google Calendar.",
            };
        }

        const startDate = parseGoogleCalendarDateTime(appointment.date, appointment.time);
        if (!startDate) {
            return {
                ...base,
                helper: "Date and time are required to add this appointment to Google Calendar.",
            };
        }

        const endDate = new Date(startDate.getTime() + duration * 60 * 1000);
        const startValue = formatGoogleCalendarDateTimeValue(startDate);
        const endValue = formatGoogleCalendarDateTimeValue(endDate);
        if (!startValue || !endValue) {
            return base;
        }

        const customerName = normalizeString(appointment.name);
        const serviceLabel = normalizeString(appointment.serviceLabel);
        const hasServiceLabel = serviceLabel && normalizeLower(serviceLabel) !== "service not found";

        let title = "Appointment";
        if (hasServiceLabel && customerName) {
            title = `${serviceLabel} - ${customerName}`;
        } else if (customerName) {
            title = `Appointment - ${customerName}`;
        }

        const detailsLines = [];
        if (customerName) {
            detailsLines.push(`Customer: ${customerName}`);
        }
        if (normalizeString(appointment.email)) {
            detailsLines.push(`Email: ${normalizeString(appointment.email)}`);
        }
        if (normalizeString(appointment.phone)) {
            detailsLines.push(`Phone: ${normalizeString(appointment.phone)}`);
        }
        if (normalizeString(appointment.customerNotes)) {
            detailsLines.push("", "Customer notes:", normalizeString(appointment.customerNotes));
        }
        detailsLines.push("", "Source: Nuvio Booking");

        const params = new URLSearchParams();
        params.set("action", "TEMPLATE");
        params.set("text", title);
        params.set("dates", `${startValue}/${endValue}`);
        params.set("details", detailsLines.join("\n"));
        params.set("ctz", "Europe/Lisbon");

        const location = resolveGoogleCalendarLocation(website);
        if (location) {
            params.set("location", location);
        }

        return {
            available: true,
            href: `https://calendar.google.com/calendar/render?${params.toString()}`,
            helper: "Opens Google Calendar with this appointment pre-filled.",
        };
    }

    function openGoogleCalendarForSelectedAppointment() {
        const targetUrl = normalizeString(selectedAppointmentGoogleCalendar?.href);
        if (!targetUrl || !selectedAppointmentGoogleCalendar?.available) {
            return;
        }

        const openedWindow = window.open(targetUrl, "_blank", "noopener,noreferrer");
        if (openedWindow) {
            openedWindow.opener = null;
        }
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
        const confirmationMode = normalizeBookingConfirmationMode(settings.confirmationMode);
        const blockingMode = normalizeBookingCalendarBlockingMode(settings?.rules?.calendarBlockingMode);
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

        if (featureAvailable && bookingEnabled && confirmationMode === "autoConfirm" && blockingMode === "none") {
            warnings.push("Auto-confirm is enabled while appointment blocking is disabled. Overlapping confirmed appointments may be created.");
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
        activeExceptionsCount = 0,
        bookingRulesConfiguredCount = 0,
    } = {}) {
        const warnings = Array.isArray(bookingWarnings) ? [...bookingWarnings] : [];

        if (availabilityValidationIssuesCount > 0) {
            warnings.push(
                `${availabilityValidationIssuesCount} active window${availabilityValidationIssuesCount === 1 ? "" : "s"} need valid, non-overlapping time ranges.`,
            );
        }

        if (dirtyAvailabilityRowsCount > 0) {
            warnings.push("There are unsaved availability changes.");
        }

        if (activeExceptionsCount <= 0) {
            warnings.push("No active exceptions are configured for special dates.");
        }

        if (bookingRulesConfiguredCount <= 0) {
            warnings.push("Booking rules are all set to default values.");
        }

        return [...new Set(warnings)];
    }

    function buildAvailabilityHealthSuggestions({
        bookingSuggestions = [],
        dirtyAvailabilityRowsCount = 0,
        availabilityValidationIssuesCount = 0,
        activeExceptionsCount = 0,
        bookingRulesConfiguredCount = 0,
    } = {}) {
        const suggestions = Array.isArray(bookingSuggestions) ? [...bookingSuggestions] : [];

        if (dirtyAvailabilityRowsCount > 0) {
            suggestions.unshift("Save changes after updating schedule windows or using presets.");
        }

        if (availabilityValidationIssuesCount > 0) {
            suggestions.unshift("Fix active windows where times are invalid or overlap.");
        }

        if (activeExceptionsCount <= 0) {
            suggestions.push("Add exceptions for holidays, closed dates, or custom special-day hours.");
        }

        if (bookingRulesConfiguredCount <= 0) {
            suggestions.push("Review booking rules to set minimum notice, booking window, and slot buffer.");
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
        const serviceRecord = servicesRecords.find((service) => normalizeString(service?.id) === serviceId);
        const serviceNameSnapshot = readRecordStringByAliases(
            record,
            appointmentServiceNameSnapshotFieldName,
            appointmentServiceNameSnapshotFieldAliases,
        );
        const serviceDurationSnapshot = readRecordPositiveIntegerByAliases(
            record,
            appointmentServiceDurationSnapshotFieldName,
            appointmentServiceDurationSnapshotFieldAliases,
        );
        const serviceDescriptionSnapshot = readRecordStringByAliases(
            record,
            appointmentServiceDescriptionSnapshotFieldName,
            appointmentServiceDescriptionSnapshotFieldAliases,
        );

        const serviceLabel = serviceNameSnapshot
            || normalizeString(expandedService?.name)
            || serviceLabelById.get(serviceId)
            || "Service not found";
        const serviceDurationMinutes = [
            serviceDurationSnapshot,
            readNonNegativeInteger(expandedService?.durationMinutes, 0),
            readNonNegativeInteger(serviceRecord?.durationMinutes, 0),
        ].find((value) => Number.isFinite(value) && value > 0) || 0;
        const serviceDescription = serviceDescriptionSnapshot
            || normalizeString(expandedService?.description)
            || normalizeString(serviceRecord?.description);

        const statusKey = normalizeStatus(record?.[appointmentStatusFieldName]);
        const statusMeta = getStatusMeta(statusKey);

        return {
            ...record,
            id,
            serviceId,
            serviceLabel,
            serviceDurationMinutes,
            serviceDescription,
            name: normalizeString(record?.name) || "Unnamed customer",
            email: normalizeString(record?.email),
            phone: normalizeString(record?.phone),
            date: normalizeString(record?.date),
            time: normalizeString(record?.time),
            customerNotes: normalizeString(record?.[appointmentCustomerNotesFieldName]),
            internalNotes: normalizeString(record?.[appointmentInternalNotesFieldName]),
            confirmedAt: appointmentConfirmedAtFieldName
                ? normalizeString(record?.[appointmentConfirmedAtFieldName])
                : normalizeString(record?.confirmedAt || record?.confirmed_at),
            cancelledAt: appointmentCancelledAtFieldName
                ? normalizeString(record?.[appointmentCancelledAtFieldName])
                : normalizeString(record?.cancelledAt || record?.cancelled_at),
            rescheduledAt: appointmentRescheduledAtFieldName
                ? normalizeString(record?.[appointmentRescheduledAtFieldName])
                : normalizeString(record?.rescheduledAt || record?.rescheduled_at),
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

    async function setSelectedAppointmentStatus(nextStatus, options = {}) {
        if (!selectedAppointment?.id || !nextStatus || isUpdatingAppointmentStatus || !appointmentsCollection?.id) {
            return;
        }

        const statusValue = normalizeStatus(nextStatus);
        const sendEmail = statusValue === "confirmed" ? !!options?.sendEmail : false;

        isUpdatingAppointmentStatus = true;
        updatingAppointmentId = selectedAppointment.id;

        try {
            const response = await ApiClient.send(`/api/nuvio/booking/admin/appointments/${encodeURIComponent(selectedAppointment.id)}/status`, {
                method: "POST",
                body: {
                    status: statusValue,
                    sendEmail,
                },
                requestKey: `nuvio_booking_status_${selectedWebsiteId}_${selectedAppointment.id}`,
            });

            const resolvedStatus = normalizeStatus(response?.status || statusValue);
            const patchPayload = {
                [appointmentStatusFieldName]: resolvedStatus,
            };

            const confirmedAtValue = normalizeString(response?.confirmedAt);
            if (confirmedAtValue) {
                patchPayload[appointmentConfirmedAtFieldName || "confirmedAt"] = confirmedAtValue;
            }

            const cancelledAtValue = normalizeString(response?.cancelledAt);
            if (cancelledAtValue) {
                patchPayload[appointmentCancelledAtFieldName || "cancelledAt"] = cancelledAtValue;
            }

            patchAppointmentRecord(selectedAppointment.id, patchPayload);

            addSuccessToast(`Appointment marked as ${resolvedStatus}.`);
            if (normalizeString(response?.warning)) {
                addErrorToast(normalizeString(response.warning));
            }
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
            description: "",
            priority: "normal",
            active: true,
        };
        serviceFormError = "";
    }

    function setServiceFormFromRecord(service) {
        serviceForm = {
            id: normalizeString(service?.id),
            name: normalizeString(service?.name),
            durationMinutes: `${service?.durationMinutes || "30"}`,
            description: normalizeString(service?.description),
            priority: mapDisplayOrderToPriority(service?.displayOrder),
            active: !!service?.active,
        };
        serviceFormError = "";
    }

    function createNewService() {
        isCreatingService = true;
        selectedServiceId = "";
        resetServiceForm();
    }

    function selectService(service) {
        isCreatingService = false;
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
        const normalizedDescription = normalizeString(serviceForm.description);
        const normalizedPriority = normalizeServicePriority(serviceForm.priority);

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
            description: normalizedDescription,
            displayOrder: mapPriorityToDisplayOrder(normalizedPriority),
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
                isCreatingService = false;
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

    function nextAvailabilityWindowKey(dayKey = "day") {
        availabilityWindowCounter += 1;
        return `${dayKey}_window_${availabilityWindowCounter}`;
    }

    function normalizeAvailabilityTimeValue(value, fallback = "09:00") {
        const normalized = normalizeString(value);
        return bookingTimePattern.test(normalized) ? normalized : fallback;
    }

    function createAvailabilityWindowDraft(dayKey, source = {}, options = {}) {
        const recordId = normalizeString(source?.recordId || source?.id);
        const active = source?.active !== false;
        const startTime = normalizeString(source?.startTime) || "09:00";
        const endTime = normalizeString(source?.endTime) || "17:00";
        const initialActive = typeof source?.initialActive === "boolean" ? source.initialActive : active;
        const initialStartTime = normalizeString(source?.initialStartTime) || startTime;
        const initialEndTime = normalizeString(source?.initialEndTime) || endTime;
        const isNew = !!options?.isNew || !recordId;

        const draft = {
            key: recordId ? `availability_${recordId}` : nextAvailabilityWindowKey(dayKey),
            recordId,
            dayOfWeek: dayKey,
            active: !!active,
            startTime,
            endTime,
            initialActive: !!initialActive,
            initialStartTime,
            initialEndTime,
            isNew,
            dirty: false,
            error: "",
        };

        draft.dirty = isAvailabilityWindowDirty(draft);
        return draft;
    }

    function sortAvailabilityWindows(windows = []) {
        return [...windows].sort((firstWindow, secondWindow) => {
            const firstActive = firstWindow?.active ? 1 : 0;
            const secondActive = secondWindow?.active ? 1 : 0;
            if (firstActive !== secondActive) {
                return secondActive - firstActive;
            }

            const firstStart = parseTimeToMinutes(firstWindow?.startTime);
            const secondStart = parseTimeToMinutes(secondWindow?.startTime);
            if (firstStart !== secondStart) {
                if (firstStart < 0) {
                    return 1;
                }
                if (secondStart < 0) {
                    return -1;
                }
                return firstStart - secondStart;
            }

            const firstEnd = parseTimeToMinutes(firstWindow?.endTime);
            const secondEnd = parseTimeToMinutes(secondWindow?.endTime);
            if (firstEnd !== secondEnd) {
                if (firstEnd < 0) {
                    return 1;
                }
                if (secondEnd < 0) {
                    return -1;
                }
                return firstEnd - secondEnd;
            }

            return normalizeString(firstWindow?.key).localeCompare(normalizeString(secondWindow?.key));
        });
    }

    function createDefaultAvailabilityRows() {
        return availabilityDays.map((day) => ({
            dayOfWeek: day.key,
            label: day.label,
            windows: [],
        }));
    }

    function createAvailabilityRowsFromRecords(records = []) {
        const rowsByDay = new Map(
            availabilityDays.map((day) => [
                day.key,
                {
                    dayOfWeek: day.key,
                    label: day.label,
                    windows: [],
                },
            ]),
        );

        for (const record of records || []) {
            const dayKey = normalizeLower(record?.dayOfWeek);
            if (!rowsByDay.has(dayKey)) {
                continue;
            }

            const row = rowsByDay.get(dayKey);
            row.windows.push(
                createAvailabilityWindowDraft(dayKey, {
                    id: normalizeString(record?.id),
                    active: !!record?.active,
                    startTime: normalizeString(record?.startTime) || "09:00",
                    endTime: normalizeString(record?.endTime) || "17:00",
                    initialActive: !!record?.active,
                    initialStartTime: normalizeString(record?.startTime) || "09:00",
                    initialEndTime: normalizeString(record?.endTime) || "17:00",
                }),
            );
        }

        const rows = availabilityDays.map((day) => {
            const row = rowsByDay.get(day.key) || {
                dayOfWeek: day.key,
                label: day.label,
                windows: [],
            };

            return {
                ...row,
                windows: sortAvailabilityWindows(row.windows || []),
            };
        });

        return applyAvailabilityValidation(rows);
    }

    function isAvailabilityWindowDirty(window) {
        if (!window) {
            return false;
        }

        if (window.isNew) {
            return true;
        }

        return !!window.active !== !!window.initialActive
            || normalizeString(window.startTime) !== normalizeString(window.initialStartTime)
            || normalizeString(window.endTime) !== normalizeString(window.initialEndTime);
    }

    function isAvailabilityDayDirty(dayRow) {
        return (dayRow?.windows || []).some((window) => !!window?.dirty);
    }

    function countActiveWindowsForDay(dayRow) {
        return (dayRow?.windows || []).filter((window) => !!window?.active).length;
    }

    function countActiveAvailabilityDays(rows = availabilityRows) {
        return (rows || []).filter((dayRow) => countActiveWindowsForDay(dayRow) > 0).length;
    }

    function countDirtyAvailabilityWindows(rows = availabilityRows) {
        return (rows || []).reduce((count, dayRow) =>
            count + (dayRow?.windows || []).filter((window) => !!window?.dirty).length, 0);
    }

    function countDirtyAvailabilityDays(rows = availabilityRows) {
        return (rows || []).filter((dayRow) => isAvailabilityDayDirty(dayRow)).length;
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

    function validateAvailabilityWindow(window) {
        if (!window?.active) {
            return "";
        }

        const normalizedStart = normalizeString(window?.startTime);
        const normalizedEnd = normalizeString(window?.endTime);
        if (!bookingTimePattern.test(normalizedStart) || !bookingTimePattern.test(normalizedEnd)) {
            return "Start and end time must use HH:mm format.";
        }

        const startMinutes = parseTimeToMinutes(normalizedStart);
        const endMinutes = parseTimeToMinutes(normalizedEnd);
        if (startMinutes < 0 || endMinutes < 0 || endMinutes <= startMinutes) {
            return "End time must be after start time.";
        }

        return "";
    }

    function buildAvailabilityValidationResult(rows = availabilityRows) {
        const errorByWindowKey = new Map();
        let invalidActiveWindowsCount = 0;

        for (const dayRow of rows || []) {
            const activeWindows = [];

            for (const window of dayRow?.windows || []) {
                const baseError = validateAvailabilityWindow(window);
                if (baseError) {
                    errorByWindowKey.set(window.key, baseError);
                    if (window?.active) {
                        invalidActiveWindowsCount += 1;
                    }
                    continue;
                }

                if (window?.active) {
                    activeWindows.push({
                        key: window.key,
                        start: parseTimeToMinutes(window.startTime),
                        end: parseTimeToMinutes(window.endTime),
                    });
                }
            }

            for (let index = 0; index < activeWindows.length; index += 1) {
                const currentWindow = activeWindows[index];
                if (currentWindow.start < 0 || currentWindow.end < 0 || currentWindow.end <= currentWindow.start) {
                    continue;
                }

                for (let compareIndex = index + 1; compareIndex < activeWindows.length; compareIndex += 1) {
                    const otherWindow = activeWindows[compareIndex];
                    if (otherWindow.start < 0 || otherWindow.end < 0 || otherWindow.end <= otherWindow.start) {
                        continue;
                    }

                    if (currentWindow.start < otherWindow.end && otherWindow.start < currentWindow.end) {
                        errorByWindowKey.set(currentWindow.key, "Active windows cannot overlap.");
                        errorByWindowKey.set(otherWindow.key, "Active windows cannot overlap.");
                    }
                }
            }
        }

        const nextRows = (rows || []).map((dayRow) => ({
            ...dayRow,
            windows: sortAvailabilityWindows(
                (dayRow?.windows || []).map((window) => {
                    const nextError = errorByWindowKey.get(window.key) || "";
                    const normalizedWindow = {
                        ...window,
                        error: nextError,
                    };
                    normalizedWindow.dirty = isAvailabilityWindowDirty(normalizedWindow);
                    if (!nextError && !!normalizedWindow.active) {
                        return normalizedWindow;
                    }
                    return normalizedWindow;
                }),
            ),
        }));

        invalidActiveWindowsCount = nextRows.reduce((count, dayRow) =>
            count + (dayRow?.windows || []).filter((window) => !!window?.active && !!window?.error).length, 0);

        return {
            rows: nextRows,
            invalidActiveWindowsCount,
        };
    }

    function applyAvailabilityValidation(rows = availabilityRows) {
        return buildAvailabilityValidationResult(rows).rows;
    }

    function countAvailabilityValidationIssues(rows = availabilityRows) {
        return buildAvailabilityValidationResult(rows).invalidActiveWindowsCount;
    }

    function findAvailabilityWindow(dayKey, windowKey) {
        const dayRow = availabilityRows.find((row) => row.dayOfWeek === dayKey);
        if (!dayRow) {
            return null;
        }

        const window = dayRow.windows.find((entry) => entry.key === windowKey);
        if (!window) {
            return null;
        }

        return { dayRow, window };
    }

    function updateAvailabilityWindow(dayKey, windowKey, patchData = {}) {
        const nextRows = availabilityRows.map((dayRow) => {
            if (dayRow.dayOfWeek !== dayKey) {
                return dayRow;
            }

            const nextWindows = dayRow.windows.map((window) => {
                if (window.key !== windowKey) {
                    return window;
                }

                const nextWindow = {
                    ...window,
                    ...patchData,
                    startTime: normalizeString(
                        typeof patchData.startTime === "string" ? patchData.startTime : window.startTime,
                    ),
                    endTime: normalizeString(
                        typeof patchData.endTime === "string" ? patchData.endTime : window.endTime,
                    ),
                };
                nextWindow.dirty = isAvailabilityWindowDirty(nextWindow);
                return nextWindow;
            });

            return {
                ...dayRow,
                windows: sortAvailabilityWindows(nextWindows),
            };
        });

        availabilityRows = applyAvailabilityValidation(nextRows);
    }

    function addAvailabilityWindow(dayKey) {
        const nextRows = availabilityRows.map((dayRow) => {
            if (dayRow.dayOfWeek !== dayKey) {
                return dayRow;
            }

            const newWindow = createAvailabilityWindowDraft(dayKey, {
                startTime: "09:00",
                endTime: "17:00",
                active: true,
                initialActive: false,
                initialStartTime: "",
                initialEndTime: "",
            }, { isNew: true });

            return {
                ...dayRow,
                windows: sortAvailabilityWindows([...dayRow.windows, newWindow]),
            };
        });

        availabilityRows = applyAvailabilityValidation(nextRows);
    }

    function removeOrDeactivateAvailabilityWindow(dayKey, windowKey) {
        const target = findAvailabilityWindow(dayKey, windowKey);
        if (!target) {
            return;
        }

        if (target.window.isNew || !target.window.recordId) {
            const nextRows = availabilityRows.map((dayRow) => {
                if (dayRow.dayOfWeek !== dayKey) {
                    return dayRow;
                }

                return {
                    ...dayRow,
                    windows: sortAvailabilityWindows(dayRow.windows.filter((window) => window.key !== windowKey)),
                };
            });
            availabilityRows = applyAvailabilityValidation(nextRows);
            return;
        }

        updateAvailabilityWindow(dayKey, windowKey, {
            active: false,
            error: "",
        });
    }

    function toggleAvailabilityWindowActive(dayKey, windowKey, nextActive) {
        updateAvailabilityWindow(dayKey, windowKey, {
            active: !!nextActive,
            error: "",
        });
    }

    async function saveAvailabilityWindow(dayKey, windowKey, options = {}) {
        if (!selectedWebsiteId || !bookingAvailabilityCollection?.id) {
            return false;
        }

        const target = findAvailabilityWindow(dayKey, windowKey);
        if (!target?.window) {
            return false;
        }

        const row = target.dayRow;
        const window = target.window;

        if (window.isNew && !window.active) {
            const nextRows = availabilityRows.map((dayRow) =>
                dayRow.dayOfWeek === dayKey
                    ? {
                        ...dayRow,
                        windows: sortAvailabilityWindows(dayRow.windows.filter((entry) => entry.key !== windowKey)),
                    }
                    : dayRow);
            availabilityRows = applyAvailabilityValidation(nextRows);
            return true;
        }

        isSavingAvailability = { ...isSavingAvailability, [windowKey]: true };

        const normalizedStartTime = normalizeString(window.startTime) || "09:00";
        const normalizedEndTime = normalizeString(window.endTime) || "17:00";
        const normalizedActive = !!window.active;

        const payload = {
            website: selectedWebsiteId,
            dayOfWeek: dayKey,
            startTime: normalizedStartTime,
            endTime: normalizedEndTime,
            active: normalizedActive,
        };

        try {
            if (window.recordId) {
                const updated = await ApiClient.collection(bookingAvailabilityCollection.id).update(window.recordId, payload);
                availabilityRecords = availabilityRecords.map((record) =>
                    normalizeString(record?.id) === normalizeString(updated?.id)
                        ? { ...record, ...updated }
                        : record,
                );

                const nextRows = availabilityRows.map((dayRow) => {
                    if (dayRow.dayOfWeek !== dayKey) {
                        return dayRow;
                    }

                    return {
                        ...dayRow,
                        windows: dayRow.windows.map((entry) => {
                            if (entry.key !== windowKey) {
                                return entry;
                            }

                            const nextWindow = {
                                ...entry,
                                recordId: normalizeString(updated?.id) || entry.recordId,
                                active: normalizedActive,
                                startTime: normalizedStartTime,
                                endTime: normalizedEndTime,
                                initialActive: normalizedActive,
                                initialStartTime: normalizedStartTime,
                                initialEndTime: normalizedEndTime,
                                isNew: false,
                                error: "",
                            };
                            nextWindow.dirty = isAvailabilityWindowDirty(nextWindow);
                            return nextWindow;
                        }),
                    };
                });
                availabilityRows = applyAvailabilityValidation(nextRows);

                if (!options?.silent) {
                    addSuccessToast(`${row.label} window updated.`);
                }
            } else {
                const created = await ApiClient.collection(bookingAvailabilityCollection.id).create(payload);
                availabilityRecords = [created, ...availabilityRecords];

                const nextRows = availabilityRows.map((dayRow) => {
                    if (dayRow.dayOfWeek !== dayKey) {
                        return dayRow;
                    }

                    return {
                        ...dayRow,
                        windows: dayRow.windows.map((entry) => {
                            if (entry.key !== windowKey) {
                                return entry;
                            }

                            const nextWindow = {
                                ...entry,
                                recordId: normalizeString(created?.id),
                                active: normalizedActive,
                                startTime: normalizedStartTime,
                                endTime: normalizedEndTime,
                                initialActive: normalizedActive,
                                initialStartTime: normalizedStartTime,
                                initialEndTime: normalizedEndTime,
                                isNew: false,
                                error: "",
                            };
                            nextWindow.dirty = isAvailabilityWindowDirty(nextWindow);
                            return nextWindow;
                        }),
                    };
                });
                availabilityRows = applyAvailabilityValidation(nextRows);

                if (!options?.silent) {
                    addSuccessToast(`${row.label} window saved.`);
                }
            }

            if (!options?.skipSlotPreviewRefresh) {
                await loadSlotPreview();
            }
            return true;
        } catch (err) {
            ApiClient.error(err, false);
            updateAvailabilityWindow(dayKey, windowKey, { error: "Unable to save this window right now." });
            if (!options?.silent) {
                addErrorToast("Unable to save availability right now.");
            }
            return false;
        } finally {
            const nextSavingState = { ...isSavingAvailability };
            delete nextSavingState[windowKey];
            isSavingAvailability = nextSavingState;
        }
    }

    function isWindowSaving(windowKey) {
        return !!isSavingAvailability?.[windowKey];
    }

    function replaceDayWindowsFromTemplates(dayKey, templates = []) {
        const normalizedTemplates = (templates || []).map((template) => ({
            startTime: normalizeAvailabilityTimeValue(template?.startTime, "09:00"),
            endTime: normalizeAvailabilityTimeValue(template?.endTime, "17:00"),
            active: template?.active !== false,
        }));

        const nextRows = availabilityRows.map((dayRow) => {
            if (dayRow.dayOfWeek !== dayKey) {
                return dayRow;
            }

            const persistedWindows = dayRow.windows.filter((window) => !!window.recordId);
            const nextWindows = [];
            let persistedIndex = 0;

            for (const template of normalizedTemplates) {
                const persistedWindow = persistedWindows[persistedIndex];
                if (persistedWindow) {
                    persistedIndex += 1;
                    const nextWindow = {
                        ...persistedWindow,
                        active: !!template.active,
                        startTime: template.startTime,
                        endTime: template.endTime,
                        error: "",
                    };
                    nextWindow.dirty = isAvailabilityWindowDirty(nextWindow);
                    nextWindows.push(nextWindow);
                } else {
                    const newWindow = createAvailabilityWindowDraft(dayKey, {
                        active: !!template.active,
                        startTime: template.startTime,
                        endTime: template.endTime,
                        initialActive: false,
                        initialStartTime: "",
                        initialEndTime: "",
                    }, { isNew: true });
                    nextWindows.push(newWindow);
                }
            }

            for (; persistedIndex < persistedWindows.length; persistedIndex += 1) {
                const persistedWindow = persistedWindows[persistedIndex];
                const nextWindow = {
                    ...persistedWindow,
                    active: false,
                    error: "",
                };
                nextWindow.dirty = isAvailabilityWindowDirty(nextWindow);
                nextWindows.push(nextWindow);
            }

            return {
                ...dayRow,
                windows: sortAvailabilityWindows(nextWindows),
            };
        });

        availabilityRows = applyAvailabilityValidation(nextRows);
    }

    function applyMondayToWeekdays() {
        const mondayRow = availabilityRows.find((row) => row.dayOfWeek === "mon");
        if (!mondayRow) {
            return;
        }

        const mondayTemplates = (mondayRow.windows || [])
            .filter((window) => !!window?.active)
            .map((window) => ({
                startTime: normalizeAvailabilityTimeValue(window.startTime, "09:00"),
                endTime: normalizeAvailabilityTimeValue(window.endTime, "17:00"),
                active: true,
            }));

        const weekdayKeys = ["tue", "wed", "thu", "fri"];
        for (const dayKey of weekdayKeys) {
            replaceDayWindowsFromTemplates(dayKey, mondayTemplates);
        }
    }

    function setWeekdaysBusinessHours() {
        const weekdayKeys = ["mon", "tue", "wed", "thu", "fri"];
        for (const dayKey of weekdayKeys) {
            replaceDayWindowsFromTemplates(dayKey, [
                {
                    startTime: "09:00",
                    endTime: "17:00",
                    active: true,
                },
            ]);
        }
    }

    function clearWeekendAvailability() {
        const weekendKeys = ["sat", "sun"];
        for (const dayKey of weekendKeys) {
            replaceDayWindowsFromTemplates(dayKey, []);
        }
    }

    async function saveAllChangedAvailabilityRows() {
        if (isSavingAllAvailability || hasSavingAvailabilityRows) {
            return;
        }

        const validationResult = buildAvailabilityValidationResult(availabilityRows);
        availabilityRows = validationResult.rows;

        const dirtyWindows = [];
        let invalidDirtyWindowsCount = 0;
        for (const dayRow of availabilityRows) {
            for (const window of dayRow.windows || []) {
                if (!window?.dirty) {
                    continue;
                }

                if (window.error) {
                    invalidDirtyWindowsCount += 1;
                    continue;
                }

                dirtyWindows.push({
                    dayOfWeek: dayRow.dayOfWeek,
                    key: window.key,
                });
            }
        }

        if (!dirtyWindows.length && !invalidDirtyWindowsCount) {
            addSuccessToast("No unsaved availability changes.");
            return;
        }

        if (!dirtyWindows.length && invalidDirtyWindowsCount > 0) {
            addErrorToast("Fix invalid windows before saving schedule changes.");
            return;
        }

        isSavingAllAvailability = true;

        try {
            let savedCount = 0;
            let failedCount = 0;
            for (const entry of dirtyWindows) {
                const saved = await saveAvailabilityWindow(entry.dayOfWeek, entry.key, {
                    silent: true,
                    skipSlotPreviewRefresh: true,
                });
                if (saved) {
                    savedCount += 1;
                } else {
                    failedCount += 1;
                }
            }

            if (savedCount > 0) {
                addSuccessToast(`${savedCount} schedule window${savedCount === 1 ? "" : "s"} saved.`);
            }

            if (failedCount > 0 || invalidDirtyWindowsCount > 0) {
                addErrorToast("Some schedule changes could not be saved.");
            }

            if (savedCount > 0) {
                await loadSlotPreview();
            }
        } finally {
            isSavingAllAvailability = false;
        }
    }

    function createNewException() {
        selectedExceptionId = "";
        isCreatingException = true;
        exceptionForm = createDefaultExceptionForm();
        exceptionFormError = "";
    }

    function selectException(exception) {
        isCreatingException = false;
        selectedExceptionId = normalizeString(exception?.id);
    }

    function setExceptionType(nextType) {
        const normalizedType = normalizeExceptionType(nextType);
        exceptionForm = {
            ...exceptionForm,
            type: normalizedType,
            ...(normalizedType === "closed"
                ? {}
                : {
                    startTime: normalizeString(exceptionForm.startTime) || "09:00",
                    endTime: normalizeString(exceptionForm.endTime) || "17:00",
                }),
        };
        exceptionFormError = "";
    }

    function validateExceptionForm(form = exceptionForm) {
        const dateValue = normalizeString(form?.date);
        const typeValue = normalizeExceptionType(form?.type);
        const startTimeValue = normalizeString(form?.startTime);
        const endTimeValue = normalizeString(form?.endTime);

        if (!bookingDatePattern.test(dateValue)) {
            return "Date must use YYYY-MM-DD format.";
        }

        if (!["closed", "customHours"].includes(typeValue)) {
            return "Type must be Closed or Custom hours.";
        }

        if (typeValue === "customHours") {
            if (!bookingTimePattern.test(startTimeValue) || !bookingTimePattern.test(endTimeValue)) {
                return "Start and end time must use HH:mm format.";
            }

            const startMinutes = parseTimeToMinutes(startTimeValue);
            const endMinutes = parseTimeToMinutes(endTimeValue);
            if (startMinutes < 0 || endMinutes < 0 || endMinutes <= startMinutes) {
                return "End time must be after start time.";
            }
        }

        return "";
    }

    async function saveException() {
        if (!selectedWebsiteId || !bookingExceptionsCollection?.id || isSavingException) {
            return;
        }

        const validationError = validateExceptionForm(exceptionForm);
        if (validationError) {
            exceptionFormError = validationError;
            return;
        }

        exceptionFormError = "";
        isSavingException = true;

        const typeValue = normalizeExceptionType(exceptionForm.type);
        const payload = {
            website: selectedWebsiteId,
            date: normalizeString(exceptionForm.date),
            type: typeValue,
            startTime: typeValue === "customHours" ? normalizeString(exceptionForm.startTime) : "",
            endTime: typeValue === "customHours" ? normalizeString(exceptionForm.endTime) : "",
            note: normalizeString(exceptionForm.note),
            active: !!exceptionForm.active,
        };

        try {
            if (normalizeString(exceptionForm.id)) {
                const updated = await ApiClient.collection(bookingExceptionsCollection.id).update(exceptionForm.id, payload);
                exceptionsRecords = exceptionsRecords.map((record) =>
                    normalizeString(record?.id) === normalizeString(updated?.id)
                        ? { ...record, ...updated }
                        : record,
                );
                isCreatingException = false;
                selectedExceptionId = normalizeString(updated?.id);
                addSuccessToast("Exception updated.");
            } else {
                const created = await ApiClient.collection(bookingExceptionsCollection.id).create(payload);
                exceptionsRecords = [created, ...exceptionsRecords];
                isCreatingException = false;
                selectedExceptionId = normalizeString(created?.id);
                addSuccessToast("Exception created.");
            }

            await loadSlotPreview();
        } catch (err) {
            ApiClient.error(err, false);
            exceptionFormError = "Unable to save exception right now.";
            addErrorToast("Unable to save exception right now.");
        } finally {
            isSavingException = false;
        }
    }

    async function saveBookingRules() {
        if (!selectedWebsiteId || !websitesCollection?.id || !selectedWebsite) {
            return;
        }

        bookingRulesFormError = "";

        let minNoticeHours = 0;
        let bookingWindowDays = 0;
        let bufferMinutes = 0;

        try {
            minNoticeHours = parseRulesDraftField(bookingRulesDraft.minNoticeHours, "Minimum notice");
            bookingWindowDays = parseRulesDraftField(bookingRulesDraft.bookingWindowDays, "Booking window");
            bufferMinutes = parseRulesDraftField(bookingRulesDraft.bufferMinutes, "Buffer between appointments");
        } catch (err) {
            bookingRulesFormError = normalizeString(err?.message) || "Rules values are invalid.";
            return;
        }

        isSavingBookingRules = true;

        try {
            const currentSettings = parseSettingsObject(selectedWebsite?.[websiteSettingsFieldName]);
            const bookingSettings = readObject(currentSettings.booking);
            const currentRules = readObject(bookingSettings.rules);

            const nextSettings = {
                ...currentSettings,
                booking: {
                    ...bookingSettings,
                    rules: {
                        ...currentRules,
                        minNoticeHours,
                        bookingWindowDays,
                        bufferMinutes,
                    },
                },
            };

            const updatedWebsite = await ApiClient.collection(websitesCollection.id).update(selectedWebsiteId, {
                [websiteSettingsFieldName]: nextSettings,
            });

            websites = websites.map((website) =>
                normalizeString(website?.id) === normalizeString(updatedWebsite?.id)
                    ? { ...website, ...updatedWebsite }
                    : website,
            );

            bookingRulesDraft = createBookingRulesDraftFromSettings({
                minNoticeHours,
                bookingWindowDays,
                bufferMinutes,
            });
            addSuccessToast("Booking rules saved.");
            await loadSlotPreview();
        } catch (err) {
            ApiClient.error(err, false);
            bookingRulesFormError = "Unable to save booking rules right now.";
            addErrorToast("Unable to save booking rules right now.");
        } finally {
            isSavingBookingRules = false;
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
                                {#if selectedAppointment.serviceDurationMinutes > 0}
                                    <div class="booking-summary-row">
                                        <span class="txt-xs txt-hint">Duration</span>
                                        <span class="txt-sm">{selectedAppointment.serviceDurationMinutes} min</span>
                                    </div>
                                {/if}
                                {#if selectedAppointment.serviceDescription}
                                    <div class="booking-summary-row">
                                        <span class="txt-xs txt-hint">Description</span>
                                        <span class="txt-sm">{selectedAppointment.serviceDescription}</span>
                                    </div>
                                {/if}
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Date &amp; time</span>
                                    <span class="txt-sm">{formatAppointmentDateTime(selectedAppointment.date, selectedAppointment.time)}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Created</span>
                                    <span class="txt-sm">{formatDateTime(selectedAppointment.created)}</span>
                                </div>
                                {#if selectedAppointment.confirmedAt}
                                    <div class="booking-summary-row">
                                        <span class="txt-xs txt-hint">Confirmed at</span>
                                        <span class="txt-sm">{formatDateTime(selectedAppointment.confirmedAt)}</span>
                                    </div>
                                {/if}
                                {#if selectedAppointment.cancelledAt}
                                    <div class="booking-summary-row">
                                        <span class="txt-xs txt-hint">Cancelled at</span>
                                        <span class="txt-sm">{formatDateTime(selectedAppointment.cancelledAt)}</span>
                                    </div>
                                {/if}
                                {#if selectedAppointment.rescheduledAt}
                                    <div class="booking-summary-row">
                                        <span class="txt-xs txt-hint">Rescheduled at</span>
                                        <span class="txt-sm">{formatDateTime(selectedAppointment.rescheduledAt)}</span>
                                    </div>
                                {/if}
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
                                {#if canRescheduleSelectedAppointment}
                                    <button
                                        type="button"
                                        class="btn btn-outline btn-sm"
                                        disabled={isUpdatingAppointmentStatus || isSavingRescheduleAppointment}
                                        on:click={openReschedulePanel}
                                    >
                                        <span class="txt">Reschedule</span>
                                    </button>
                                {/if}

                                {#if canSetSelectedAppointmentConfirmed}
                                    <label class="booking-checkbox-row booking-actions-email-option">
                                        <input
                                            type="checkbox"
                                            checked={sendConfirmationEmailOnConfirm}
                                            disabled={isUpdatingAppointmentStatus}
                                            on:change={(event) => {
                                                sendConfirmationEmailOnConfirm = !!event.currentTarget.checked;
                                            }}
                                        />
                                        <span class="txt-xs txt-hint">Send confirmation email</span>
                                    </label>
                                    <button
                                        type="button"
                                        class="btn btn-sm"
                                        class:btn-loading={isUpdatingAppointmentStatus && updatingAppointmentId === selectedAppointment.id}
                                        disabled={isUpdatingAppointmentStatus}
                                        on:click={() => setSelectedAppointmentStatus("confirmed", { sendEmail: sendConfirmationEmailOnConfirm })}
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

                            <div class="booking-actions-row booking-actions-row--utility">
                                <button
                                    type="button"
                                    class="btn btn-outline btn-sm"
                                    disabled={!selectedAppointmentGoogleCalendar.available}
                                    on:click={openGoogleCalendarForSelectedAppointment}
                                >
                                    <span class="txt">Add to Google Calendar</span>
                                </button>
                            </div>
                            <p class="txt-xs txt-hint m-b-0">
                                {selectedAppointmentGoogleCalendar.helper || "Opens Google Calendar with this appointment pre-filled."}
                            </p>
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
                                placeholder="Search by name, description, duration, or priority..."
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
                                        <div class="booking-service-meta txt-sm txt-hint">
                                            {service.durationMinutes || 0} minutes
                                            <span class="booking-service-meta-separator">·</span>
                                            <span>{service.priorityLabel}</span>
                                        </div>
                                        {#if service.description}
                                            <p class="booking-service-description txt-xs txt-hint m-b-0">{service.description}</p>
                                        {/if}
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
                        <p class="txt-sm txt-hint m-b-0">Create or update service name, duration, description, priority, and active status.</p>
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

                            <div class="form-field">
                                <label class="txt-sm txt-hint" for="booking-service-priority">Priority</label>
                                <select
                                    id="booking-service-priority"
                                    class="input input-sm"
                                    bind:value={serviceForm.priority}
                                    disabled={isSavingService}
                                >
                                    {#each servicePriorityOptions as option (option.key)}
                                        <option value={option.key}>{option.label}</option>
                                    {/each}
                                </select>
                                <p class="txt-xs txt-hint m-b-0">Higher priority services appear first when visitors choose a service.</p>
                            </div>

                            <div class="form-field">
                                <label class="txt-sm txt-hint" for="booking-service-description">Description</label>
                                <textarea
                                    id="booking-service-description"
                                    class="input booking-service-description-input"
                                    rows="4"
                                    placeholder="Short visitor-facing service description..."
                                    bind:value={serviceForm.description}
                                    disabled={isSavingService}
                                />
                                <p class="txt-xs txt-hint m-b-0">Short description shown to visitors when choosing a service.</p>
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

                    <section class="booking-rail-block">
                        <h5 class="m-0">Public preview</h5>
                        <p class="txt-sm txt-hint m-b-0">Approximate visitor-facing service presentation.</p>
                        <div class="booking-service-preview-card">
                            <div class="booking-service-title-row">
                                <div class="booking-service-title">{normalizeString(serviceForm.name) || "Service name"}</div>
                                <span class={`label label-sm ${serviceForm.active ? "label-success" : "label-warning"}`}>
                                    {serviceForm.active ? "Active" : "Inactive"}
                                </span>
                            </div>
                            <div class="booking-service-meta txt-sm txt-hint">
                                {serviceFormDurationValid ? serviceFormDurationValue : 0} minutes
                            </div>
                            {#if normalizeString(serviceForm.description)}
                                <p class="booking-service-description txt-xs txt-hint m-b-0">{normalizeString(serviceForm.description)}</p>
                            {:else}
                                <p class="txt-xs txt-hint m-b-0">No service description yet.</p>
                            {/if}
                        </div>
                    </section>

                    <section class="booking-rail-block booking-health-panel">
                        <div class="booking-health-head">
                            <div class="booking-health-main">
                                <h5 class="m-0">Service health</h5>
                                <p class="txt-sm txt-hint m-b-0">Check whether this service is ready for visitors.</p>
                            </div>
                            <div class="booking-health-meta">
                                <span class={`label label-sm ${serviceHealthState.badgeClass}`}>{serviceHealthState.label}</span>
                                <span class="summary-pill">{serviceHealthWarnings.length} warnings - {serviceHealthSuggestions.length} suggestions</span>
                            </div>
                        </div>

                        <div class="booking-health-group m-t-8">
                            <div class="booking-health-group-title">Warnings</div>
                            {#if serviceHealthWarnings.length}
                                {#each serviceHealthWarnings as warning}
                                    <div class="booking-health-item warning">
                                        <span class="label label-sm booking-health-pill warning">Warning</span>
                                        <span>{warning}</span>
                                    </div>
                                {/each}
                            {:else}
                                <p class="txt-sm txt-hint m-b-0">This service has no blocking issues.</p>
                            {/if}
                        </div>

                        <div class="booking-health-group m-t-8">
                            <div class="booking-health-group-title">Suggestions</div>
                            {#if serviceHealthSuggestions.length}
                                {#each serviceHealthSuggestions as suggestion}
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
                </aside>
            </div>
        {:else}
            <div class="booking-availability-layout">
                <div class="tabs-header compact combined left operations-tabs operations-tabs--nested booking-availability-tabs">
                    {#each availabilityTabs as tab (tab.key)}
                        <button
                            type="button"
                            class="tab-item"
                            class:active={activeAvailabilityTab === tab.key}
                            on:click={() => (activeAvailabilityTab = tab.key)}
                        >
                            <span class="tab-label">
                                <i class={tab.icon} aria-hidden="true" />
                                {tab.label}
                            </span>
                        </button>
                    {/each}
                </div>

                {#if activeAvailabilityTab === "weekly"}
                    <div class="booking-split-layout booking-split-layout--availability">
                        <div class="booking-main-column booking-availability-main">
                            <div class="booking-section-head-row booking-section-head-row--compact">
                                <div class="booking-section-head-copy booking-section-head-copy--inline">
                                    <h4 class="m-0">Weekly schedule</h4>
                                    <p class="txt-sm txt-hint m-b-0">Set active days and time windows for appointment requests.</p>
                                </div>
                                <div class="booking-section-head-actions">
                                    <span class="summary-pill">{activeAvailabilityDaysCount} active days</span>
                                    <span class="summary-pill" class:warning={dirtyAvailabilityRowsCount > 0}>
                                        {dirtyAvailabilityRowsCount} unsaved day{dirtyAvailabilityRowsCount === 1 ? "" : "s"} ({dirtyAvailabilityWindowsCount} window{dirtyAvailabilityWindowsCount === 1 ? "" : "s"})
                                    </span>
                                </div>
                            </div>

                            <div class="booking-availability-list">
                                {#each availabilityRows as row (row.dayOfWeek)}
                                    <article class="booking-availability-day">
                                        <div class="booking-availability-day-head">
                                            <div class="booking-availability-day-title">
                                                <span class="txt-sm booking-availability-day-label">{row.label}</span>
                                            </div>
                                            <div class="booking-availability-day-meta">
                                                <span class="summary-pill">{countActiveWindowsForDay(row)} active</span>
                                                {#if isAvailabilityDayDirty(row)}
                                                    <span class="label label-sm label-info booking-unsaved-pill">Unsaved</span>
                                                {/if}
                                                <button
                                                    type="button"
                                                    class="btn btn-outline btn-sm booking-day-add-window-btn"
                                                    disabled={isSavingAllAvailability || hasSavingAvailabilityRows}
                                                    on:click={() => addAvailabilityWindow(row.dayOfWeek)}
                                                >
                                                    <span class="txt">Add time window</span>
                                                </button>
                                            </div>
                                        </div>

                                        {#if row.windows.length}
                                            <div class="booking-availability-window-list">
                                                {#each row.windows as window (window.key)}
                                                    <div class="booking-availability-window" class:is-inactive={!window.active}>
                                                        <label class="booking-checkbox-row booking-checkbox-row--compact">
                                                            <input
                                                                type="checkbox"
                                                                checked={window.active}
                                                                disabled={isSavingAllAvailability || hasSavingAvailabilityRows || isWindowSaving(window.key)}
                                                                on:change={(event) => toggleAvailabilityWindowActive(row.dayOfWeek, window.key, !!event.currentTarget.checked)}
                                                            />
                                                            <span class="txt-xs" class:txt-hint={!window.active}>{window.active ? "Active" : "Inactive"}</span>
                                                        </label>

                                                        <div class="booking-availability-window-time">
                                                            <input
                                                                id={`booking-start-${row.dayOfWeek}-${window.key}`}
                                                                class="input input-sm booking-time-input"
                                                                type="time"
                                                                value={window.startTime}
                                                                disabled={!window.active || isSavingAllAvailability || hasSavingAvailabilityRows || isWindowSaving(window.key)}
                                                                on:input={(event) => updateAvailabilityWindow(row.dayOfWeek, window.key, { startTime: event.currentTarget.value, error: "" })}
                                                            />
                                                            <span class="txt-xs txt-hint">→</span>
                                                            <input
                                                                id={`booking-end-${row.dayOfWeek}-${window.key}`}
                                                                class="input input-sm booking-time-input"
                                                                type="time"
                                                                value={window.endTime}
                                                                disabled={!window.active || isSavingAllAvailability || hasSavingAvailabilityRows || isWindowSaving(window.key)}
                                                                on:input={(event) => updateAvailabilityWindow(row.dayOfWeek, window.key, { endTime: event.currentTarget.value, error: "" })}
                                                            />
                                                        </div>

                                                        <div class="booking-availability-window-state">
                                                            {#if window.error}
                                                                <span class="label label-sm label-danger booking-state-pill">Invalid</span>
                                                            {:else if window.dirty}
                                                                <span class="label label-sm label-info booking-unsaved-pill">Unsaved</span>
                                                            {:else}
                                                                <span class="txt-xs txt-hint booking-state-saved">Saved</span>
                                                            {/if}
                                                        </div>

                                                        <div class="booking-availability-window-actions">
                                                            {#if window.isNew}
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-outline btn-sm"
                                                                    disabled={isSavingAllAvailability || hasSavingAvailabilityRows || isWindowSaving(window.key)}
                                                                    on:click={() => removeOrDeactivateAvailabilityWindow(row.dayOfWeek, window.key)}
                                                                >
                                                                    <span class="txt">Remove</span>
                                                                </button>
                                                            {:else}
                                                                <button
                                                                    type="button"
                                                                    class="btn btn-outline btn-sm"
                                                                    disabled={isSavingAllAvailability || hasSavingAvailabilityRows || isWindowSaving(window.key)}
                                                                    on:click={() => toggleAvailabilityWindowActive(row.dayOfWeek, window.key, !window.active)}
                                                                >
                                                                    <span class="txt">{window.active ? "Disable" : "Restore"}</span>
                                                                </button>
                                                            {/if}
                                                        </div>

                                                        {#if window.error}
                                                            <p class="txt-xs txt-danger m-b-0 booking-availability-error">{window.error}</p>
                                                        {/if}
                                                    </div>
                                                {/each}
                                            </div>
                                        {:else}
                                            <p class="txt-sm txt-hint m-b-0">No windows configured.</p>
                                        {/if}

                                    </article>
                                {/each}
                            </div>
                        </div>

                        <aside class="booking-rail">
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
                                            Preview uses services, weekly hours, exceptions, booking rules, and existing appointments.
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
                                {:else if activeAvailabilityDaysCount <= 0 && activeExceptionsCount <= 0}
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

                            <section class="booking-rail-block booking-quick-actions-panel">
                                <h5 class="m-0">Quick actions</h5>
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
                        </aside>
                    </div>
                {:else}
                    <div class="booking-split-layout booking-split-layout--availability">
                        <div class="booking-main-column booking-availability-main">
                            <section class="booking-rules-exceptions-panel">
                                <div class="booking-section-head-row booking-section-head-row--compact">
                                    <div class="booking-section-head-copy">
                                        <h4 class="m-0">Date exceptions</h4>
                                        <p class="txt-sm txt-hint m-b-0">Override weekly schedule for closed dates or custom special-day hours.</p>
                                    </div>
                                    <div class="booking-section-head-actions">
                                        <span class="summary-pill">{activeExceptionsCount} active exceptions</span>
                                    </div>
                                </div>

                                <div class="booking-rules-exceptions-grid">
                                    <div class="booking-rules-exceptions-list">
                                        <div class="booking-services-controls booking-exceptions-controls">
                                            <div class="booking-control-cell booking-control-cell--search">
                                                <label class="txt-sm txt-hint" for="booking-exception-search">Search</label>
                                                <input
                                                    id="booking-exception-search"
                                                    class="input input-sm"
                                                    type="search"
                                                    placeholder="Search by date or note"
                                                    bind:value={exceptionSearch}
                                                />
                                            </div>
                                            <div class="booking-control-cell booking-control-cell--select">
                                                <label class="txt-sm txt-hint" for="booking-exception-type-filter">Type</label>
                                                <select id="booking-exception-type-filter" class="input input-sm" bind:value={exceptionTypeFilter}>
                                                    {#each exceptionTypeFilterOptions as option (option.key)}
                                                        <option value={option.key}>{option.label}</option>
                                                    {/each}
                                                </select>
                                            </div>
                                            <div class="booking-control-cell booking-control-cell--select">
                                                <label class="txt-sm txt-hint" for="booking-exception-active-filter">Active</label>
                                                <select id="booking-exception-active-filter" class="input input-sm" bind:value={exceptionActiveFilter}>
                                                    {#each exceptionActiveFilterOptions as option (option.key)}
                                                        <option value={option.key}>{option.label}</option>
                                                    {/each}
                                                </select>
                                            </div>
                                        </div>

                                        {#if !normalizedExceptions.length}
                                            <div class="booking-empty-state m-b-0">
                                                <h4 class="m-0">No exceptions yet</h4>
                                                <p class="txt-sm txt-hint m-b-0">Add closed dates or custom hours for special days.</p>
                                            </div>
                                        {:else if !filteredExceptions.length}
                                            <div class="booking-empty-state m-b-0">
                                                <h4 class="m-0">No exceptions match these filters</h4>
                                                <p class="txt-sm txt-hint m-b-0">Try a different search or filter selection.</p>
                                            </div>
                                        {:else}
                                            <div class="booking-services-list booking-exceptions-list" role="list">
                                                {#each filteredExceptions as exception (exception.id)}
                                                    <article
                                                        role="listitem"
                                                        class="booking-service-item booking-exception-item"
                                                        class:selected={exception.id === selectedExceptionId}
                                                        on:click={() => selectException(exception)}
                                                    >
                                                        <div class="booking-service-main">
                                                            <div class="booking-service-title-row">
                                                                <div class="booking-service-title">{formatDate(exception.date)}</div>
                                                                <span class={`label label-sm ${exception.active ? "label-success" : "label-warning"}`}>
                                                                    {exception.active ? "Active" : "Inactive"}
                                                                </span>
                                                            </div>
                                                            <div class="booking-service-meta txt-sm txt-hint">
                                                                {exception.typeLabel} {#if exception.type === "customHours"}- {exception.timeRangeLabel}{/if}
                                                            </div>
                                                            {#if exception.note}
                                                                <div class="txt-xs txt-hint m-t-4">{exception.note}</div>
                                                            {/if}
                                                        </div>
                                                    </article>
                                                {/each}
                                            </div>
                                        {/if}
                                    </div>

                                    <section class="booking-rail-block booking-exception-details-panel">
                                        <h5 class="m-0">Exception details</h5>
                                        <p class="txt-sm txt-hint m-b-0">Create, update, or disable special-day overrides.</p>
                                        <div class="booking-actions-row booking-exception-actions-row">
                                            <button type="button" class="btn btn-outline btn-sm" disabled={isSavingException} on:click={createNewException}>
                                                <span class="txt">{exceptionForm.id ? "New exception" : "Reset form"}</span>
                                            </button>
                                            <button
                                                type="button"
                                                class="btn btn-sm"
                                                class:btn-loading={isSavingException}
                                                disabled={isSavingException}
                                                on:click={saveException}
                                            >
                                                <span class="txt">{exceptionForm.id ? "Update exception" : "Create exception"}</span>
                                            </button>
                                        </div>

                                        <div class="booking-form-stack">
                                            <div class="form-field m-b-0">
                                                <label class="txt-sm txt-hint" for="booking-exception-date">Date</label>
                                                <input
                                                    id="booking-exception-date"
                                                    class="input input-sm"
                                                    type="date"
                                                    bind:value={exceptionForm.date}
                                                    disabled={isSavingException}
                                                />
                                            </div>

                                            <div class="form-field m-b-0">
                                                <label class="txt-sm txt-hint" for="booking-exception-type">Type</label>
                                                <select
                                                    id="booking-exception-type"
                                                    class="input input-sm"
                                                    value={exceptionForm.type}
                                                    disabled={isSavingException}
                                                    on:change={(event) => setExceptionType(event.currentTarget.value)}
                                                >
                                                    <option value="closed">Closed</option>
                                                    <option value="customHours">Custom hours</option>
                                                </select>
                                            </div>

                                            {#if exceptionForm.type === "customHours"}
                                                <div class="booking-manual-grid">
                                                    <div class="form-field m-b-0">
                                                        <label class="txt-sm txt-hint" for="booking-exception-start-time">Start time</label>
                                                        <input
                                                            id="booking-exception-start-time"
                                                            class="input input-sm"
                                                            type="time"
                                                            bind:value={exceptionForm.startTime}
                                                            disabled={isSavingException}
                                                        />
                                                    </div>
                                                    <div class="form-field m-b-0">
                                                        <label class="txt-sm txt-hint" for="booking-exception-end-time">End time</label>
                                                        <input
                                                            id="booking-exception-end-time"
                                                            class="input input-sm"
                                                            type="time"
                                                            bind:value={exceptionForm.endTime}
                                                            disabled={isSavingException}
                                                        />
                                                    </div>
                                                </div>
                                            {/if}

                                            <div class="form-field m-b-0">
                                                <label class="txt-sm txt-hint" for="booking-exception-note">Note</label>
                                                <textarea
                                                    id="booking-exception-note"
                                                    class="input booking-notes-input"
                                                    rows="3"
                                                    placeholder="Optional note for this date..."
                                                    bind:value={exceptionForm.note}
                                                    disabled={isSavingException}
                                                />
                                            </div>

                                            <label class="booking-checkbox-row" for="booking-exception-active">
                                                <input
                                                    id="booking-exception-active"
                                                    type="checkbox"
                                                    bind:checked={exceptionForm.active}
                                                    disabled={isSavingException}
                                                />
                                                <span class="txt-sm">Active exception</span>
                                            </label>

                                            {#if exceptionFormError}
                                                <p class="txt-xs txt-danger m-b-0">{exceptionFormError}</p>
                                            {/if}
                                        </div>
                                    </section>
                                </div>
                            </section>

                            <section class="booking-rail-block booking-rules-panel">
                                <div class="booking-section-head-row booking-section-head-row--compact">
                                    <div class="booking-section-head-copy">
                                        <h4 class="m-0">Booking rules</h4>
                                        <p class="txt-sm txt-hint m-b-0">Control notice time, booking window, and buffer between appointments.</p>
                                    </div>
                                    <div class="booking-section-head-actions">
                                        <span class="summary-pill">{bookingRulesConfiguredCount} configured rules</span>
                                    </div>
                                </div>

                                <div class="booking-form-stack">
                                    <div class="booking-rules-fields">
                                        <div class="form-field m-b-0">
                                            <label class="txt-sm txt-hint" for="booking-rules-min-notice">Minimum notice (hours)</label>
                                            <input
                                                id="booking-rules-min-notice"
                                                class="input input-sm"
                                                type="number"
                                                min="0"
                                                step="1"
                                                bind:value={bookingRulesDraft.minNoticeHours}
                                                disabled={isSavingBookingRules}
                                            />
                                            <p class="txt-xs txt-hint m-b-0">How many hours in advance visitors must book.</p>
                                        </div>

                                        <div class="form-field m-b-0">
                                            <label class="txt-sm txt-hint" for="booking-rules-window">Booking window (days)</label>
                                            <input
                                                id="booking-rules-window"
                                                class="input input-sm"
                                                type="number"
                                                min="0"
                                                step="1"
                                                bind:value={bookingRulesDraft.bookingWindowDays}
                                                disabled={isSavingBookingRules}
                                            />
                                            <p class="txt-xs txt-hint m-b-0">How many days ahead visitors can book. Use 0 for no limit.</p>
                                        </div>

                                        <div class="form-field m-b-0">
                                            <label class="txt-sm txt-hint" for="booking-rules-buffer">Buffer between appointments (minutes)</label>
                                            <input
                                                id="booking-rules-buffer"
                                                class="input input-sm"
                                                type="number"
                                                min="0"
                                                step="1"
                                                bind:value={bookingRulesDraft.bufferMinutes}
                                                disabled={isSavingBookingRules}
                                            />
                                            <p class="txt-xs txt-hint m-b-0">Extra blocked time around appointments.</p>
                                        </div>
                                    </div>

                                    {#if bookingRulesFormError}
                                        <p class="txt-xs txt-danger m-b-0">{bookingRulesFormError}</p>
                                    {/if}

                                    <div class="booking-actions-row booking-rules-footer">
                                        <span class="summary-pill" class:warning={bookingRulesDirty}>
                                            {bookingRulesDirty ? "Unsaved changes" : "All changes saved"}
                                        </span>
                                        <button
                                            type="button"
                                            class="btn btn-sm"
                                            class:btn-loading={isSavingBookingRules}
                                            disabled={isSavingBookingRules || !bookingRulesDirty}
                                            on:click={saveBookingRules}
                                        >
                                            <span class="txt">Save rules</span>
                                        </button>
                                    </div>
                                </div>
                            </section>
                        </div>

                        <aside class="booking-rail">
                            <section class="booking-rail-block booking-slot-preview-panel">
                                <div class="booking-health-head">
                                    <div class="booking-health-main">
                                        <h5 class="m-0">Slot preview</h5>
                                        <p class="txt-sm txt-hint m-b-0">
                                            Preview uses services, weekly hours, exceptions, booking rules, and existing appointments.
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
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-hint m-b-0">Add an active service to preview booking slots.</p>
                                    </div>
                                {:else if activeAvailabilityDaysCount <= 0 && activeExceptionsCount <= 0}
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-hint m-b-0">Add active availability days to generate slots.</p>
                                    </div>
                                {:else if !slotPreviewServiceId || !slotPreviewDate}
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-hint m-b-0">Select a service and date to preview slots.</p>
                                    </div>
                                {:else if isLoadingSlotPreview}
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-hint m-b-0">Loading available times...</p>
                                    </div>
                                {:else if slotPreviewError}
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-danger m-b-0">{slotPreviewError}</p>
                                    </div>
                                {:else if !slotPreviewSlots.length}
                                    <div class="booking-slot-preview-status m-t-sm">
                                        <p class="txt-sm txt-hint m-b-0">No available times for this date.</p>
                                    </div>
                                {:else}
                                    <div class="booking-slot-preview-slots-wrap m-t-sm">
                                        <div class="booking-slot-preview-slots">
                                            {#each slotPreviewSlots as slot}
                                                <span class="summary-pill booking-slot-preview-pill">{slot}</span>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}
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

                            <section class="booking-rail-block booking-quick-actions-panel">
                                <h5 class="m-0">Quick actions</h5>
                                <p class="txt-sm txt-hint m-b-0">Move between key booking operations.</p>
                                <div class="booking-availability-actions-list booking-quick-actions-list">
                                    <button type="button" class="btn btn-outline btn-sm" on:click={() => (activeAvailabilityTab = "weekly")}>
                                        <span class="txt">View weekly schedule</span>
                                    </button>
                                    <button type="button" class="btn btn-outline btn-sm" on:click={() => (activeTab = "services")}>
                                        <span class="txt">Manage services</span>
                                    </button>
                                    <button type="button" class="btn btn-outline btn-sm" on:click={() => (activeTab = "appointments")}>
                                        <span class="txt">View appointments</span>
                                    </button>
                                </div>
                            </section>
                        </aside>
                    </div>
                {/if}
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

    <OverlayPanel
        bind:active={isReschedulePanelOpen}
        class="overlay-panel-lg booking-reschedule-panel"
        overlayClose={true}
        escClose={true}
        on:hide={closeReschedulePanel}
    >
        <svelte:fragment slot="header">
            <h4>Reschedule appointment</h4>
        </svelte:fragment>

        <div class="booking-manual-form">
            {#if !rescheduleSourceAppointment}
                <p class="txt-sm txt-hint m-b-0">Select an appointment to start rescheduling.</p>
            {:else}
                <div class="booking-reschedule-current">
                    <p class="txt-xs txt-hint m-b-0">Current appointment</p>
                    <p class="txt-sm m-b-0">
                        {rescheduleSourceAppointment.serviceLabel} · {formatAppointmentDateTime(rescheduleSourceAppointment.date, rescheduleSourceAppointment.time)}
                    </p>
                </div>

                {#if !manualAppointmentServiceOptions.length}
                    <div class="alert alert-warning m-b-0">
                        <div class="icon">
                            <i class="ri-information-line" />
                        </div>
                        <div>Add and activate at least one service before rescheduling appointments.</div>
                    </div>
                {:else}
                    <div class="booking-manual-grid">
                        <div class="form-field m-b-0">
                            <label class="txt-sm txt-hint" for="booking-reschedule-service">Service</label>
                            <select
                                id="booking-reschedule-service"
                                class="input input-sm"
                                value={rescheduleForm.serviceId}
                                disabled={isSavingRescheduleAppointment}
                                on:change={(event) => setRescheduleService(event.currentTarget.value)}
                            >
                                <option value="">Select service</option>
                                {#each manualAppointmentServiceOptions as service (service.id)}
                                    <option value={service.id}>{service.label}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="form-field m-b-0">
                            <label class="txt-sm txt-hint" for="booking-reschedule-date">Date</label>
                            <input
                                id="booking-reschedule-date"
                                class="input input-sm"
                                type="date"
                                value={rescheduleForm.date}
                                disabled={isSavingRescheduleAppointment}
                                on:change={(event) => setRescheduleDate(event.currentTarget.value)}
                            />
                        </div>

                        <div class="form-field m-b-0 booking-manual-slot-field">
                            <label class="txt-sm txt-hint">Available slot</label>
                            {#if !rescheduleForm.serviceId || !rescheduleForm.date}
                                <div class="txt-xs txt-hint">Select service and date to load available times.</div>
                            {:else if isLoadingRescheduleSlots}
                                <div class="txt-xs txt-hint">Loading available times...</div>
                            {:else if rescheduleSlotsError}
                                <div class="txt-xs txt-danger">{rescheduleSlotsError}</div>
                            {:else if !rescheduleAvailableSlots.length}
                                <div class="txt-xs txt-hint">No available times for this date.</div>
                            {:else}
                                <div class="booking-manual-slots">
                                    {#each rescheduleAvailableSlots as slot}
                                        <button
                                            type="button"
                                            class="btn btn-outline btn-sm booking-manual-slot-btn"
                                            class:active={rescheduleForm.time === slot}
                                            disabled={isSavingRescheduleAppointment}
                                            on:click={() => selectRescheduleSlot(slot)}
                                        >
                                            <span class="txt">{slot}</span>
                                        </button>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    </div>

                    <label class="booking-checkbox-row booking-reschedule-email-row" for="booking-reschedule-send-email">
                        <input
                            id="booking-reschedule-send-email"
                            type="checkbox"
                            checked={rescheduleForm.sendEmail}
                            disabled={isSavingRescheduleAppointment}
                            on:change={(event) => {
                                rescheduleForm = {
                                    ...rescheduleForm,
                                    sendEmail: !!event.currentTarget.checked,
                                };
                            }}
                        />
                        <span class="txt-sm">Send reschedule email to customer</span>
                    </label>

                    {#if rescheduleForm.serviceId && rescheduleForm.date && rescheduleForm.time}
                        <p class="txt-sm txt-hint m-b-0">
                            New selected time: {serviceLabelById.get(rescheduleForm.serviceId) || "Service"} · {formatAppointmentDateTime(rescheduleForm.date, rescheduleForm.time)}
                        </p>
                    {/if}
                {/if}
            {/if}

            {#if rescheduleFormError}
                <p class="txt-xs txt-danger m-b-0">{rescheduleFormError}</p>
            {/if}
        </div>

        <svelte:fragment slot="footer">
            <button type="button" class="btn btn-outline btn-sm" disabled={isSavingRescheduleAppointment} on:click={closeReschedulePanel}>
                <span class="txt">Cancel</span>
            </button>
            <button
                type="button"
                class="btn btn-sm"
                class:btn-loading={isSavingRescheduleAppointment}
                disabled={isSavingRescheduleAppointment || isLoadingRescheduleSlots || !rescheduleSourceAppointment || !manualAppointmentServiceOptions.length}
                on:click={saveRescheduledAppointment}
            >
                <span class="txt">Save reschedule</span>
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

    .booking-actions-row--utility {
        margin-top: 2px;
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

    .booking-service-main {
        min-width: 0;
        flex: 1;
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

    .booking-service-meta-separator {
        margin: 0 4px;
    }

    .booking-service-description {
        margin-top: 6px;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        word-break: break-word;
    }

    .booking-service-actions {
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .booking-service-description-input {
        resize: vertical;
        min-height: 88px;
    }

    .booking-service-preview-card {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        background: color-mix(in srgb, var(--baseAlt1) 10%, transparent);
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

    .booking-availability-tabs {
        align-self: flex-start;
        width: auto;
        max-width: 100%;
    }

    .booking-availability-tabs .tab-item {
        flex: 0 0 auto;
    }

    .booking-availability-tabs .tab-label {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        white-space: nowrap;
    }

    .booking-availability-tabs .tab-label i {
        font-size: 0.92rem;
        line-height: 1;
    }

    .booking-split-layout--availability {
        grid-template-columns: minmax(0, 1fr) 320px;
    }

    .booking-availability-main {
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-width: 0;
    }

    .booking-rules-exceptions-panel {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .booking-rules-exceptions-grid {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(260px, 320px);
        gap: 12px;
        align-items: start;
    }

    .booking-rules-exceptions-list {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .booking-exceptions-controls {
        margin-bottom: 2px;
    }

    .booking-exceptions-list {
        max-height: 520px;
        overflow: auto;
    }

    .booking-exception-details-panel {
        gap: 10px;
        height: 100%;
    }

    .booking-exception-actions-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        flex-wrap: wrap;
        margin-top: -2px;
        padding-bottom: 8px;
        border-bottom: 1px dashed var(--baseAlt1);
    }

    .booking-rules-panel {
        gap: 12px;
    }

    .booking-rules-fields {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 10px;
    }

    .booking-rules-footer {
        justify-content: space-between;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        padding-top: 8px;
        border-top: 1px dashed var(--baseAlt1);
    }

    .booking-section-head-row--compact {
        margin-bottom: 8px;
    }

    .booking-section-head-copy--inline {
        flex-direction: row;
        align-items: baseline;
        gap: 8px;
        flex-wrap: wrap;
    }

    .booking-availability-list {
        display: grid;
        gap: 10px;
    }

    .booking-availability-actions-list {
        display: flex;
        flex-direction: column;
        flex-wrap: wrap;
        gap: 6px;
    }

    .booking-availability-day {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 11px 12px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .booking-availability-day-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .booking-availability-day-meta {
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .booking-day-add-window-btn {
        margin-left: 2px;
    }

    .booking-availability-window-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .booking-availability-window {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 8px 9px;
        display: grid;
        grid-template-columns: auto minmax(0, 1fr) auto auto;
        gap: 8px;
        align-items: center;
    }

    .booking-availability-window.is-inactive {
        background: color-mix(in srgb, var(--baseAlt1) 14%, transparent);
    }

    .booking-availability-window-time {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        min-width: 0;
    }

    .booking-availability-window-state,
    .booking-availability-window-actions {
        display: flex;
        align-items: center;
    }

    .booking-availability-window-state,
    .booking-availability-window-actions {
        justify-content: flex-end;
    }

    .booking-checkbox-row--compact {
        gap: 10px;
    }

    .booking-time-input {
        width: 100%;
        min-width: 0;
        max-width: 108px;
    }

    .booking-availability-day-label {
        font-weight: 600;
    }

    .booking-unsaved-pill {
        font-size: 0.66rem;
        line-height: 1;
    }

    .booking-state-pill {
        justify-content: center;
    }

    .booking-state-saved {
        white-space: nowrap;
    }

    .booking-availability-error {
        grid-column: 1 / -1;
        margin-top: -2px;
        padding-left: 2px;
    }

    .booking-slot-preview-panel {
        background: var(--baseColor);
        order: 1;
        border-color: color-mix(in srgb, var(--baseAlt2) 68%, var(--baseAlt1));
    }

    .booking-slot-preview-panel,
    .booking-health-panel,
    .booking-quick-actions-panel {
        box-shadow: inset 0 1px 0 color-mix(in srgb, var(--baseAlt2) 22%, transparent);
    }

    .booking-slot-preview-controls {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 10px;
    }

    .booking-slot-preview-status,
    .booking-slot-preview-slots-wrap {
        border: 1px dashed var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 9px 10px;
        background: color-mix(in srgb, var(--baseAlt1) 11%, transparent);
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
        background: var(--baseColor);
        order: 2;
        gap: 12px;
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
        padding-bottom: 4px;
        border-bottom: 1px dashed var(--baseAlt1);
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

    .booking-quick-actions-panel {
        order: 3;
        gap: 12px;
    }

    .booking-quick-actions-list {
        gap: 8px;
    }

    .booking-quick-actions-list .btn {
        justify-content: flex-start;
    }

    .booking-manual-form {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .booking-reschedule-current {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        background: color-mix(in srgb, var(--baseAlt1) 12%, transparent);
    }

    .booking-reschedule-email-row {
        margin-top: -2px;
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

        .booking-rules-fields {
            grid-template-columns: repeat(2, minmax(0, 1fr));
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

        .booking-availability-window {
            grid-template-columns: 1fr;
            align-items: flex-start;
            gap: 6px;
        }

        .booking-availability-window-time {
            width: 100%;
            flex-wrap: wrap;
        }

        .booking-availability-window-state,
        .booking-availability-window-actions {
            justify-content: flex-start;
        }

        .booking-time-input {
            max-width: none;
        }

        .booking-services-controls {
            grid-template-columns: 1fr;
        }

        .booking-rules-exceptions-grid {
            grid-template-columns: 1fr;
        }

        .booking-rules-fields {
            grid-template-columns: 1fr;
        }

        .booking-exceptions-list {
            max-height: none;
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
