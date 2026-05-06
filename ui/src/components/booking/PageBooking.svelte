<script>
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
    const appointmentNotesFieldAliases = ["notes", "note", "internalNotes", "internal_notes"];

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
    let appointmentNotesDraft = "";
    let appointmentNotesDraftId = "";
    let isSavingAppointmentNotes = false;

    let selectedServiceId = "";
    let isSavingService = false;
    let serviceForm = {
        id: "",
        name: "",
        durationMinutes: "30",
        active: true,
    };
    let serviceFormError = "";

    let availabilityRows = createDefaultAvailabilityRows();
    let isSavingAvailability = {};

    loadCollections();

    $: websitesCollection = resolveCollectionByAliases(["websites"]);
    $: bookingServicesCollection = resolveCollectionByAliases(["bookingservices"]);
    $: bookingAvailabilityCollection = resolveCollectionByAliases(["bookingavailability"]);
    $: appointmentsCollection = resolveCollectionByAliases(["appointments"]);

    $: hasBookingCollections = !!bookingServicesCollection?.id
        && !!bookingAvailabilityCollection?.id
        && !!appointmentsCollection?.id;

    $: appointmentStatusFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentStatusFieldAliases) || "status";
    $: appointmentNotesFieldName = resolveCollectionFieldNameByAliases(appointmentsCollection, appointmentNotesFieldAliases) || "notes";

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

    $: normalizedAppointments = appointmentRecords.map((record) => normalizeAppointment(record));

    $: pendingAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "pending").length;
    $: confirmedAppointmentsCount = normalizedAppointments.filter((appointment) => appointment.statusKey === "confirmed").length;
    $: thisWeekAppointmentsCount = normalizedAppointments.filter((appointment) => isInCurrentWeek(appointment.date)).length;
    $: activeServicesCount = servicesRecords.filter((service) => !!service?.active).length;

    $: appointmentServiceOptions = servicesRecords
        .map((service) => ({
            id: normalizeString(service?.id),
            label: normalizeString(service?.name) || "Untitled service",
        }))
        .filter((service) => service.id)
        .sort((a, b) => a.label.localeCompare(b.label));

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
                    appointment.notes,
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
        if (appointmentNotesDraftId !== selectedAppointment.id) {
            appointmentNotesDraftId = selectedAppointment.id;
            appointmentNotesDraft = selectedAppointment.notes || "";
        }
    } else if (appointmentNotesDraftId || appointmentNotesDraft) {
        appointmentNotesDraftId = "";
        appointmentNotesDraft = "";
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
            notes: normalizeString(record?.[appointmentNotesFieldName]),
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

    async function saveSelectedAppointmentNotes() {
        if (!selectedAppointment?.id || !appointmentsCollection?.id || isSavingAppointmentNotes) {
            return;
        }

        isSavingAppointmentNotes = true;

        try {
            await ApiClient.collection(appointmentsCollection.id).update(selectedAppointment.id, {
                [appointmentNotesFieldName]: appointmentNotesDraft,
            });

            patchAppointmentRecord(selectedAppointment.id, {
                [appointmentNotesFieldName]: appointmentNotesDraft,
            });

            addSuccessToast("Appointment notes saved.");
        } catch (err) {
            ApiClient.error(err, false);
            addErrorToast("Unable to save appointment notes right now.");
        } finally {
            isSavingAppointmentNotes = false;
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
                error: "",
            };
        });
    }

    function updateAvailabilityRow(dayKey, patchData = {}) {
        availabilityRows = availabilityRows.map((row) =>
            row.dayOfWeek === dayKey
                ? {
                    ...row,
                    ...patchData,
                }
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

    async function saveAvailabilityRow(row) {
        if (!row || !selectedWebsiteId || !bookingAvailabilityCollection?.id) {
            return;
        }

        const validationError = validateAvailabilityRow(row);
        if (validationError) {
            updateAvailabilityRow(row.dayOfWeek, { error: validationError });
            return;
        }

        updateAvailabilityRow(row.dayOfWeek, { error: "" });
        isSavingAvailability = { ...isSavingAvailability, [row.dayOfWeek]: true };

        const payload = {
            website: selectedWebsiteId,
            dayOfWeek: row.dayOfWeek,
            startTime: normalizeString(row.startTime) || "09:00",
            endTime: normalizeString(row.endTime) || "17:00",
            active: !!row.active,
        };

        try {
            if (row.recordId) {
                const updated = await ApiClient.collection(bookingAvailabilityCollection.id).update(row.recordId, payload);
                availabilityRecords = availabilityRecords.map((record) =>
                    normalizeString(record?.id) === normalizeString(updated?.id)
                        ? { ...record, ...updated }
                        : record,
                );
                updateAvailabilityRow(row.dayOfWeek, { recordId: updated.id, error: "" });
                addSuccessToast(`${row.label} availability updated.`);
            } else {
                const created = await ApiClient.collection(bookingAvailabilityCollection.id).create(payload);
                availabilityRecords = [created, ...availabilityRecords];
                updateAvailabilityRow(row.dayOfWeek, { recordId: created.id, error: "" });
                addSuccessToast(`${row.label} availability saved.`);
            }
        } catch (err) {
            ApiClient.error(err, false);
            updateAvailabilityRow(row.dayOfWeek, { error: "Unable to save this day right now." });
            addErrorToast("Unable to save availability right now.");
        } finally {
            isSavingAvailability = { ...isSavingAvailability, [row.dayOfWeek]: false };
        }
    }

    function isDaySaving(dayKey) {
        return !!isSavingAvailability?.[dayKey];
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
        <div class="tabs-header compact combined left operations-tabs booking-tabs">
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
                        <div class="empty-state m-b-0">
                            No appointments yet. Appointment requests will appear here.
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
                            <p class="txt-sm txt-hint m-b-0">Overview of this appointment request.</p>
                            <div class="booking-summary-grid">
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Status</span>
                                    <span class={`label label-sm ${selectedAppointment.statusClassName}`}>{selectedAppointment.statusLabel}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Created</span>
                                    <span class="txt-sm">{formatDateTime(selectedAppointment.created)}</span>
                                </div>
                            </div>
                        </section>

                        <section class="booking-rail-block">
                            <h5 class="m-0">Contact</h5>
                            <div class="booking-summary-grid">
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Name</span>
                                    <span class="txt-sm">{selectedAppointment.name}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Email</span>
                                    <span class="txt-sm">{selectedAppointment.email || "No email provided"}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Phone</span>
                                    <span class="txt-sm">{selectedAppointment.phone || "No phone provided"}</span>
                                </div>
                            </div>
                        </section>

                        <section class="booking-rail-block">
                            <h5 class="m-0">Service &amp; time</h5>
                            <div class="booking-summary-grid">
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Service</span>
                                    <span class="txt-sm">{selectedAppointment.serviceLabel}</span>
                                </div>
                                <div class="booking-summary-row">
                                    <span class="txt-xs txt-hint">Date &amp; time</span>
                                    <span class="txt-sm">{formatAppointmentDateTime(selectedAppointment.date, selectedAppointment.time)}</span>
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
                            <p class="txt-sm txt-hint m-b-0">Keep internal context for follow-up.</p>
                            <textarea
                                class="input booking-notes-input"
                                rows="5"
                                placeholder="Add follow-up notes for this appointment..."
                                bind:value={appointmentNotesDraft}
                                disabled={isSavingAppointmentNotes}
                            />
                            <div class="booking-actions-row">
                                <button
                                    type="button"
                                    class="btn btn-sm"
                                    class:btn-loading={isSavingAppointmentNotes}
                                    disabled={isSavingAppointmentNotes || appointmentNotesDraft === (selectedAppointment.notes || "")}
                                    on:click={saveSelectedAppointmentNotes}
                                >
                                    <span class="txt">Save notes</span>
                                </button>
                            </div>
                        </section>
                    {:else}
                        <div class="empty-state m-b-0">Select an appointment to view details.</div>
                    {/if}
                </aside>
            </div>
        {:else if activeTab === "services"}
            <div class="booking-split-layout booking-split-layout--services">
                <div class="booking-main-column">
                    <div class="booking-section-head-row">
                        <h4 class="m-0">Services</h4>
                        <button type="button" class="btn btn-outline btn-sm" on:click={createNewService}>
                            <span class="txt">New service</span>
                        </button>
                    </div>

                    {#if !servicesRecords.length}
                        <div class="empty-state m-b-0">
                            No services yet. Add a service to start accepting bookings.
                        </div>
                    {:else}
                        <div class="booking-services-list" role="list">
                            {#each servicesRecords as service (service.id)}
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
                                        <div class="booking-service-title">{service.name || "Untitled service"}</div>
                                        <div class="txt-sm txt-hint">{service.durationMinutes || 0} minutes</div>
                                    </div>
                                    <div class="booking-service-actions">
                                        <span class={`label label-sm ${service.active ? "label-success" : "label-warning"}`}>
                                            {service.active ? "Active" : "Inactive"}
                                        </span>
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
                        <h5 class="m-0">{serviceForm.id ? "Edit service" : "Create service"}</h5>
                        <p class="txt-sm txt-hint m-b-0">Define service name, duration, and active status.</p>

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
            <div class="booking-availability-layout">
                <div class="booking-section-head-row">
                    <h4 class="m-0">Weekly availability</h4>
                    <p class="txt-sm txt-hint m-b-0">Set active days and time windows for appointment requests.</p>
                </div>

                <div class="booking-availability-list">
                    {#each availabilityRows as row (row.dayOfWeek)}
                        <article class="booking-availability-row" class:is-inactive={!row.active}>
                            <div class="booking-availability-day">
                                <span class="txt-sm booking-availability-day-label">{row.label}</span>
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
                                <div class="form-field m-b-0">
                                    <label class="txt-xs txt-hint" for={`booking-start-${row.dayOfWeek}`}>Start</label>
                                    <input
                                        id={`booking-start-${row.dayOfWeek}`}
                                        class="input input-sm"
                                        type="time"
                                        value={row.startTime}
                                        disabled={!row.active || isDaySaving(row.dayOfWeek)}
                                        on:input={(event) => updateAvailabilityRow(row.dayOfWeek, { startTime: event.currentTarget.value, error: "" })}
                                    />
                                </div>
                                <div class="form-field m-b-0">
                                    <label class="txt-xs txt-hint" for={`booking-end-${row.dayOfWeek}`}>End</label>
                                    <input
                                        id={`booking-end-${row.dayOfWeek}`}
                                        class="input input-sm"
                                        type="time"
                                        value={row.endTime}
                                        disabled={!row.active || isDaySaving(row.dayOfWeek)}
                                        on:input={(event) => updateAvailabilityRow(row.dayOfWeek, { endTime: event.currentTarget.value, error: "" })}
                                    />
                                </div>
                            </div>

                            <div class="booking-availability-actions">
                                <button
                                    type="button"
                                    class="btn btn-sm btn-outline"
                                    class:btn-loading={isDaySaving(row.dayOfWeek)}
                                    disabled={isDaySaving(row.dayOfWeek)}
                                    on:click={() => saveAvailabilityRow(row)}
                                >
                                    <span class="txt">Save</span>
                                </button>
                            </div>

                            {#if row.error}
                                <p class="txt-xs txt-danger m-b-0 booking-availability-error">{row.error}</p>
                            {/if}
                        </article>
                    {/each}
                </div>
            </div>
        {/if}
    </section>
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
        justify-content: flex-end;
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

    .booking-section-head-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 10px;
        margin-bottom: 12px;
    }

    .booking-services-list {
        display: grid;
        gap: 8px;
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

    .booking-availability-list {
        display: grid;
        gap: 8px;
    }

    .booking-availability-row {
        border: 1px solid var(--baseAlt1);
        border-radius: var(--baseRadius);
        padding: 10px 12px;
        display: grid;
        grid-template-columns: minmax(160px, 220px) minmax(220px, 1fr) auto;
        gap: 12px;
        align-items: end;
    }

    .booking-availability-row.is-inactive {
        background: color-mix(in srgb, var(--baseAlt1) 16%, transparent);
    }

    .booking-availability-day {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .booking-availability-day-label {
        font-weight: 600;
    }

    .booking-availability-time-range {
        display: grid;
        grid-template-columns: repeat(2, minmax(120px, 1fr));
        gap: 8px;
    }

    .booking-availability-actions {
        display: flex;
        justify-content: flex-end;
    }

    .booking-availability-error {
        grid-column: 1 / -1;
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

        .booking-availability-actions {
            justify-content: flex-start;
        }
    }

    @media (max-width: 860px) {
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

        .booking-service-item {
            flex-direction: column;
            align-items: flex-start;
        }

        .booking-service-actions {
            width: 100%;
            justify-content: space-between;
        }
    }
</style>
