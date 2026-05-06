package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioBookingServicesCollectionID     = "pbc_1661203700"
	nuvioBookingAvailabilityCollectionID = "pbc_1661203800"
	nuvioAppointmentsCollectionID        = "pbc_1661203900"
)

var (
	nuvioBookingDatePattern        = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	nuvioBookingTimePattern        = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	errNuvioBookingSlotUnavailable = errors.New("nuvio booking slot unavailable")
)

type nuvioWebsiteBookingConfig struct {
	FeatureAvailable   bool
	Enabled            bool
	EmailNotifications nuvioEmailNotificationsConfig
}

type nuvioBookingPublicService struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"durationMinutes"`
}

type nuvioBookingCreateAppointmentPayload struct {
	WebsiteID string `json:"websiteId"`
	ServiceID string `json:"serviceId"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Notes     string `json:"notes"`
}

// NUVIO CUSTOM START: Booking MVP Phase 3 public booking endpoints.
func registerNuvioBookingRoutes(e *core.ServeEvent) {
	bookingGroup := e.Router.Group("/api/nuvio/booking")

	bookingGroup.GET("/services", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		_, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Booking feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("Booking is disabled for this website.", nil)
		}

		services, err := listNuvioActiveBookingServices(e.App, websiteID)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO booking services load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to load booking services right now.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"websiteId": websiteID,
			"services":  services,
		})
	})

	bookingGroup.GET("/slots", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		serviceID := strings.TrimSpace(e.Request.URL.Query().Get("serviceId"))
		dateValue := strings.TrimSpace(e.Request.URL.Query().Get("date"))

		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if serviceID == "" {
			return e.BadRequestError("Missing serviceId.", nil)
		}
		if !nuvioBookingDatePattern.MatchString(dateValue) {
			return e.BadRequestError("Date must use YYYY-MM-DD format.", nil)
		}

		_, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Booking feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("Booking is disabled for this website.", nil)
		}

		serviceRecord, err := e.App.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO booking service lookup failed",
				"websiteId",
				websiteID,
				"serviceId",
				serviceID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to load booking service right now.", nil)
		}

		if strings.TrimSpace(serviceRecord.GetString("website")) != websiteID {
			return e.NotFoundError("Service not found.", nil)
		}

		if !isNuvioBookingServiceActive(serviceRecord) {
			return e.BadRequestError("Service is not available.", nil)
		}

		slots, err := computeNuvioAvailableSlots(e.App, websiteID, serviceRecord, dateValue)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO booking slots generation failed",
				"websiteId",
				websiteID,
				"serviceId",
				serviceID,
				"date",
				dateValue,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Unable to compute slots for this date.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"date":      dateValue,
			"serviceId": serviceID,
			"slots":     slots,
		})
	})

	bookingGroup.POST("/appointments", func(e *core.RequestEvent) error {
		payload := nuvioBookingCreateAppointmentPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid booking appointment payload.", nil)
		}

		websiteID := strings.TrimSpace(payload.WebsiteID)
		if websiteID == "" {
			websiteID = strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		serviceID := strings.TrimSpace(payload.ServiceID)
		if serviceID == "" {
			return e.BadRequestError("Missing serviceId.", nil)
		}

		dateValue := strings.TrimSpace(payload.Date)
		if !nuvioBookingDatePattern.MatchString(dateValue) {
			return e.BadRequestError("Date must use YYYY-MM-DD format.", nil)
		}

		timeValue := strings.TrimSpace(payload.Time)
		if !nuvioBookingTimePattern.MatchString(timeValue) {
			return e.BadRequestError("Time must use HH:mm format.", nil)
		}

		name := strings.TrimSpace(payload.Name)
		if name == "" {
			return e.BadRequestError("Name is required.", nil)
		}

		email, ok := normalizeNuvioEmail(payload.Email)
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}

		phone := strings.TrimSpace(payload.Phone)
		notes := strings.TrimSpace(payload.Notes)

		website, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Booking feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("Booking is disabled for this website.", nil)
		}

		serviceRecord, err := e.App.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO booking service lookup failed",
				"websiteId",
				websiteID,
				"serviceId",
				serviceID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to load booking service right now.", nil)
		}

		if strings.TrimSpace(serviceRecord.GetString("website")) != websiteID {
			return e.NotFoundError("Service not found.", nil)
		}

		if !isNuvioBookingServiceActive(serviceRecord) {
			return e.BadRequestError("Service is not available.", nil)
		}

		serviceName := strings.TrimSpace(serviceRecord.GetString("name"))
		if serviceName == "" {
			serviceName = "Booking service"
		}

		var appointmentID string
		transactionErr := e.App.RunInTransaction(func(txApp core.App) error {
			txServiceRecord, err := txApp.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
			if err != nil {
				return err
			}

			if strings.TrimSpace(txServiceRecord.GetString("website")) != websiteID || !isNuvioBookingServiceActive(txServiceRecord) {
				return sql.ErrNoRows
			}

			slots, err := computeNuvioAvailableSlots(txApp, websiteID, txServiceRecord, dateValue)
			if err != nil {
				return err
			}
			if !containsNuvioBookingSlot(slots, timeValue) {
				return errNuvioBookingSlotUnavailable
			}

			appointmentsCollection, err := txApp.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
			if err != nil {
				return err
			}

			appointmentRecord := core.NewRecord(appointmentsCollection)
			appointmentRecord.Set("website", websiteID)
			appointmentRecord.Set("service", serviceID)
			appointmentRecord.Set("name", name)
			appointmentRecord.Set("email", email)
			appointmentRecord.Set("phone", phone)
			appointmentRecord.Set("date", dateValue)
			appointmentRecord.Set("time", timeValue)
			appointmentRecord.Set("notes", notes)
			appointmentRecord.Set("status", "pending")

			if err := txApp.Save(appointmentRecord); err != nil {
				return err
			}

			contactsCollection, err := txApp.FindCachedCollectionByNameOrId(nuvioContactsCollectionID)
			if err != nil {
				return err
			}

			contactRecord := core.NewRecord(contactsCollection)
			contactRecord.Set("website", websiteID)
			contactRecord.Set("channel", "booking")
			contactRecord.Set("name", name)
			contactRecord.Set("email", email)
			contactRecord.Set("phone", phone)
			contactRecord.Set("subject", fmt.Sprintf("Booking request - %s", serviceName))
			contactRecord.Set("message", buildNuvioBookingContactMessage(serviceName, dateValue, timeValue, notes))
			contactRecord.Set("status", "new")

			if err := txApp.Save(contactRecord); err != nil {
				return err
			}

			appointmentID = appointmentRecord.Id
			return nil
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingSlotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{
					"ok":    false,
					"error": "This time is no longer available.",
				})
			}

			if errors.Is(transactionErr, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking appointment create failed",
				"websiteId",
				websiteID,
				"serviceId",
				serviceID,
				"date",
				dateValue,
				"time",
				timeValue,
				"error",
				transactionErr.Error(),
			)
			return e.InternalServerError("Unable to create booking request right now.", nil)
		}

		responsePayload := map[string]any{
			"ok":            true,
			"appointmentId": appointmentID,
			"status":        "pending",
		}

		if emailErr := sendNuvioBookingEmails(
			e.Request.Context(),
			website,
			config,
			serviceName,
			nuvioBookingCreateAppointmentPayload{
				WebsiteID: websiteID,
				ServiceID: serviceID,
				Date:      dateValue,
				Time:      timeValue,
				Name:      name,
				Email:     email,
				Phone:     phone,
				Notes:     notes,
			},
		); emailErr != nil {
			e.App.Logger().Error(
				"NUVIO booking email send failed",
				"websiteId",
				websiteID,
				"appointmentId",
				appointmentID,
				"error",
				emailErr.Error(),
			)
			responsePayload["warning"] = "Booking request saved, but email notifications are temporarily unavailable."
		}

		return e.JSON(http.StatusOK, responsePayload)
	})
}

func loadNuvioWebsiteBookingConfig(
	app core.App,
	websiteID string,
) (*core.Record, nuvioWebsiteBookingConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteBookingConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))
	config := nuvioWebsiteBookingConfig{
		FeatureAvailable: true,
		Enabled:          true,
		EmailNotifications: nuvioEmailNotificationsConfig{
			Enabled: false,
			To:      []string{},
			Cc:      []string{},
		},
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["booking"]); ok {
			config.FeatureAvailable = value
		}
	}

	if bookingSettings, ok := toStringAnyMap(settings["booking"]); ok {
		if value, ok := parseBoolValue(bookingSettings["enabled"]); ok {
			config.Enabled = value
		}

		legacyDestination := strings.TrimSpace(parseStringValue(bookingSettings["emailDestination"]))
		config.EmailNotifications = parseNuvioEmailNotificationsConfig(
			bookingSettings["emailNotifications"],
			legacyDestination,
		)
	}

	// Fallback to Contact Form notification recipients when Booking-specific recipients are not configured yet.
	if len(config.EmailNotifications.To) == 0 {
		if contactFormSettings, ok := toStringAnyMap(settings["contactForm"]); ok {
			legacyDestination := strings.TrimSpace(parseStringValue(contactFormSettings["emailDestination"]))
			fallbackNotifications := parseNuvioEmailNotificationsConfig(
				contactFormSettings["emailNotifications"],
				legacyDestination,
			)

			if len(fallbackNotifications.To) > 0 {
				config.EmailNotifications.To = append([]string{}, fallbackNotifications.To...)
			}
			if len(fallbackNotifications.Cc) > 0 {
				config.EmailNotifications.Cc = append([]string{}, fallbackNotifications.Cc...)
			}
			if fallbackNotifications.Enabled {
				config.EmailNotifications.Enabled = true
			}
		}
	}

	return website, config, nil
}

func listNuvioActiveBookingServices(app core.App, websiteID string) ([]nuvioBookingPublicService, error) {
	servicesCollection, err := app.FindCachedCollectionByNameOrId(nuvioBookingServicesCollectionID)
	if err != nil {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		servicesCollection,
		"website={:website} && active=true",
		"+name",
		5000,
		0,
		dbx.Params{
			"website": websiteID,
		},
	)
	if err != nil {
		return nil, err
	}

	services := make([]nuvioBookingPublicService, 0, len(records))
	for _, record := range records {
		duration, err := parseNuvioBookingServiceDuration(record)
		if err != nil {
			continue
		}

		services = append(services, nuvioBookingPublicService{
			ID:              strings.TrimSpace(record.Id),
			Name:            strings.TrimSpace(record.GetString("name")),
			DurationMinutes: duration,
		})
	}

	sort.SliceStable(services, func(i, j int) bool {
		return strings.ToLower(strings.TrimSpace(services[i].Name)) < strings.ToLower(strings.TrimSpace(services[j].Name))
	})

	return services, nil
}

func isNuvioBookingServiceActive(serviceRecord *core.Record) bool {
	if serviceRecord == nil {
		return false
	}

	if value, ok := parseBoolValue(serviceRecord.Get("active")); ok {
		return value
	}

	return false
}

func parseNuvioBookingServiceDuration(serviceRecord *core.Record) (int, error) {
	if serviceRecord == nil {
		return 0, fmt.Errorf("missing service record")
	}

	duration := int(serviceRecord.GetFloat("durationMinutes"))
	if duration <= 0 {
		if parsed, err := strconv.Atoi(strings.TrimSpace(serviceRecord.GetString("durationMinutes"))); err == nil {
			duration = parsed
		}
	}

	if duration < 5 || duration > 480 {
		return 0, fmt.Errorf("invalid service duration")
	}

	return duration, nil
}

func computeNuvioAvailableSlots(
	app core.App,
	websiteID string,
	serviceRecord *core.Record,
	dateValue string,
) ([]string, error) {
	if !nuvioBookingDatePattern.MatchString(strings.TrimSpace(dateValue)) {
		return nil, fmt.Errorf("invalid date format")
	}

	serviceID := strings.TrimSpace(serviceRecord.Id)
	if serviceID == "" {
		return nil, fmt.Errorf("missing service id")
	}

	dayOfWeek, err := dateToNuvioBookingDayOfWeek(dateValue)
	if err != nil {
		return nil, err
	}

	availabilityRecord, err := findNuvioBookingAvailabilityRecord(app, websiteID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	if availabilityRecord == nil {
		return []string{}, nil
	}

	startMinutes, err := parseNuvioBookingHHMM(strings.TrimSpace(availabilityRecord.GetString("startTime")))
	if err != nil {
		return nil, err
	}

	endMinutes, err := parseNuvioBookingHHMM(strings.TrimSpace(availabilityRecord.GetString("endTime")))
	if err != nil {
		return nil, err
	}

	if endMinutes <= startMinutes {
		return nil, fmt.Errorf("invalid availability range")
	}

	durationMinutes, err := parseNuvioBookingServiceDuration(serviceRecord)
	if err != nil {
		return nil, err
	}

	candidateSlots := generateNuvioBookingSlots(startMinutes, endMinutes, durationMinutes)
	if len(candidateSlots) == 0 {
		return []string{}, nil
	}

	blockedSlots, err := loadNuvioBlockedAppointmentSlots(app, websiteID, serviceID, dateValue)
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(candidateSlots))
	for _, slot := range candidateSlots {
		if _, isBlocked := blockedSlots[slot]; isBlocked {
			continue
		}
		filtered = append(filtered, slot)
	}

	return filtered, nil
}

func findNuvioBookingAvailabilityRecord(
	app core.App,
	websiteID string,
	dayOfWeek string,
) (*core.Record, error) {
	availabilityCollection, err := app.FindCachedCollectionByNameOrId(nuvioBookingAvailabilityCollectionID)
	if err != nil {
		return nil, err
	}

	record, err := app.FindFirstRecordByFilter(
		availabilityCollection,
		"website={:website} && dayOfWeek={:dayOfWeek} && active=true",
		dbx.Params{
			"website":   websiteID,
			"dayOfWeek": dayOfWeek,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func loadNuvioBlockedAppointmentSlots(
	app core.App,
	websiteID string,
	serviceID string,
	dateValue string,
) (map[string]struct{}, error) {
	appointmentsCollection, err := app.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
	if err != nil {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		appointmentsCollection,
		"website={:website} && service={:service} && date={:date}",
		"-created",
		5000,
		0,
		dbx.Params{
			"website": websiteID,
			"service": serviceID,
			"date":    dateValue,
		},
	)
	if err != nil {
		return nil, err
	}

	blocked := map[string]struct{}{}
	for _, record := range records {
		status := strings.ToLower(strings.TrimSpace(record.GetString("status")))
		if status != "pending" && status != "confirmed" {
			continue
		}

		timeValue := strings.TrimSpace(record.GetString("time"))
		if !nuvioBookingTimePattern.MatchString(timeValue) {
			continue
		}

		blocked[timeValue] = struct{}{}
	}

	return blocked, nil
}

func dateToNuvioBookingDayOfWeek(dateValue string) (string, error) {
	location := time.UTC
	if lisbon, err := time.LoadLocation("Europe/Lisbon"); err == nil && lisbon != nil {
		location = lisbon
	}

	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateValue), location)
	if err != nil {
		return "", fmt.Errorf("invalid date value")
	}

	switch date.Weekday() {
	case time.Monday:
		return "mon", nil
	case time.Tuesday:
		return "tue", nil
	case time.Wednesday:
		return "wed", nil
	case time.Thursday:
		return "thu", nil
	case time.Friday:
		return "fri", nil
	case time.Saturday:
		return "sat", nil
	case time.Sunday:
		return "sun", nil
	default:
		return "", fmt.Errorf("invalid day of week")
	}
}

func parseNuvioBookingHHMM(value string) (int, error) {
	normalized := strings.TrimSpace(value)
	if !nuvioBookingTimePattern.MatchString(normalized) {
		return 0, fmt.Errorf("invalid time format")
	}

	parts := strings.Split(normalized, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time value")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid time value")
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid time value")
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid time value")
	}

	return hour*60 + minute, nil
}

func formatNuvioBookingHHMM(totalMinutes int) string {
	if totalMinutes < 0 {
		totalMinutes = 0
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func generateNuvioBookingSlots(startMinutes int, endMinutes int, durationMinutes int) []string {
	if durationMinutes <= 0 || startMinutes < 0 || endMinutes <= startMinutes {
		return []string{}
	}

	slots := make([]string, 0, 24)
	for current := startMinutes; current+durationMinutes <= endMinutes; current += durationMinutes {
		slots = append(slots, formatNuvioBookingHHMM(current))
	}

	return slots
}

func containsNuvioBookingSlot(slots []string, requestedTime string) bool {
	normalizedTime := strings.TrimSpace(requestedTime)
	for _, slot := range slots {
		if strings.TrimSpace(slot) == normalizedTime {
			return true
		}
	}
	return false
}

func buildNuvioBookingContactMessage(
	serviceName string,
	dateValue string,
	timeValue string,
	notes string,
) string {
	lines := []string{
		fmt.Sprintf("Service: %s", strings.TrimSpace(serviceName)),
		fmt.Sprintf("Date: %s", strings.TrimSpace(dateValue)),
		fmt.Sprintf("Time: %s", strings.TrimSpace(timeValue)),
	}

	if trimmedNotes := strings.TrimSpace(notes); trimmedNotes != "" {
		lines = append(lines, "", "Notes:", trimmedNotes)
	}

	return strings.Join(lines, "\n")
}

func sendNuvioBookingEmails(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteBookingConfig,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
) error {
	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	websiteName := resolveWebsiteDisplayName(website)
	if websiteName == "" {
		websiteName = "Website"
	}

	visitorEmail, ok := normalizeNuvioEmail(payload.Email)
	if !ok {
		return fmt.Errorf("invalid visitor email")
	}

	visitorLines := []string{
		fmt.Sprintf("Hi %s,", strings.TrimSpace(payload.Name)),
		"",
		"We received your booking request.",
		fmt.Sprintf("Service: %s", strings.TrimSpace(serviceName)),
		fmt.Sprintf("Date: %s", strings.TrimSpace(payload.Date)),
		fmt.Sprintf("Time: %s", strings.TrimSpace(payload.Time)),
		fmt.Sprintf("Website: %s", websiteName),
		"",
		"We will contact you shortly to confirm the appointment.",
	}

	if notes := strings.TrimSpace(payload.Notes); notes != "" {
		visitorLines = append(visitorLines, "", "Notes received:", notes)
	}

	sendErrors := []string{}

	if err := sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      []string{visitorEmail},
		Subject: "Booking request received",
		Text:    strings.Join(visitorLines, "\n"),
	}); err != nil {
		sendErrors = append(sendErrors, fmt.Sprintf("visitor confirmation failed: %s", err.Error()))
	}

	if config.EmailNotifications.Enabled && len(config.EmailNotifications.To) > 0 {
		businessLines := []string{
			fmt.Sprintf("Website: %s", websiteName),
			fmt.Sprintf("Submitted at: %s", time.Now().UTC().Format(time.RFC3339)),
			fmt.Sprintf("Service: %s", strings.TrimSpace(serviceName)),
			fmt.Sprintf("Date: %s", strings.TrimSpace(payload.Date)),
			fmt.Sprintf("Time: %s", strings.TrimSpace(payload.Time)),
			fmt.Sprintf("Name: %s", strings.TrimSpace(payload.Name)),
			fmt.Sprintf("Email: %s", visitorEmail),
		}

		if phone := strings.TrimSpace(payload.Phone); phone != "" {
			businessLines = append(businessLines, fmt.Sprintf("Phone: %s", phone))
		}

		if notes := strings.TrimSpace(payload.Notes); notes != "" {
			businessLines = append(businessLines, "", "Notes:", notes)
		}

		if err := sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
			To:      config.EmailNotifications.To,
			Cc:      config.EmailNotifications.Cc,
			ReplyTo: []string{visitorEmail},
			Subject: fmt.Sprintf("New booking request - %s", websiteName),
			Text:    strings.Join(businessLines, "\n"),
		}); err != nil {
			sendErrors = append(sendErrors, fmt.Sprintf("business notification failed: %s", err.Error()))
		}
	}

	if len(sendErrors) > 0 {
		return errors.New(strings.Join(sendErrors, "; "))
	}

	return nil
}

// NUVIO CUSTOM END: Booking MVP Phase 3 public booking endpoints.
