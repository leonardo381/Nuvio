package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioBookingServicesCollectionID     = "pbc_1661203700"
	nuvioBookingAvailabilityCollectionID = "pbc_1661203800"
	nuvioBookingExceptionsCollectionID   = "pbc_1778803400"
	nuvioAppointmentsCollectionID        = "pbc_1661203900"
	nuvioBookingConfirmationModeRequest  = "request"
	nuvioBookingConfirmationModeAuto     = "autoConfirm"
	nuvioBookingBlockingModeService      = "service"
	nuvioBookingBlockingModeWebsite      = "website"
	nuvioBookingBlockingModeNone         = "none"
)

var (
	nuvioBookingDatePattern             = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	nuvioBookingTimePattern             = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	nuvioBookingIntegerValuePattern     = regexp.MustCompile(`^-?\d+$`)
	nuvioServiceNameSnapshotAliases     = []string{"serviceNameSnapshot", "service_name_snapshot"}
	nuvioServiceDurationSnapshotAliases = []string{"serviceDurationMinutesSnapshot", "service_duration_minutes_snapshot"}
	nuvioServiceDescSnapshotAliases     = []string{"serviceDescriptionSnapshot", "service_description_snapshot"}
	errNuvioBookingSlotUnavailable      = errors.New("nuvio booking slot unavailable")
	errNuvioBookingDateOutsideWindow    = errors.New("nuvio booking date outside window")
	errNuvioBookingTimeTooSoon          = errors.New("nuvio booking time too soon")
	errNuvioBookingAppointmentNotFound  = errors.New("nuvio booking appointment not found")
	errNuvioBookingServiceNotFound      = errors.New("nuvio booking service not found")
	errNuvioBookingAppointmentCancelled = errors.New("nuvio booking appointment cancelled")
)

type nuvioWebsiteBookingConfig struct {
	FeatureAvailable   bool
	Enabled            bool
	ConfirmationMode   string
	EmailNotifications nuvioEmailNotificationsConfig
	Rules              nuvioBookingRulesConfig
}

type nuvioBookingRulesConfig struct {
	MinNoticeHours       int
	BookingWindowDays    int
	BufferMinutes        int
	CalendarBlockingMode string
}

type nuvioBookingDailyRange struct {
	StartMinutes int
	EndMinutes   int
}

type nuvioBookingInterval struct {
	StartMinutes int
	EndMinutes   int
}

type nuvioBookingSlotComputationOptions struct {
	ExcludeAppointmentID string
}

type nuvioBookingServiceSnapshot struct {
	Name            string
	DurationMinutes int
	Description     string
}

type nuvioBookingPublicService struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"durationMinutes"`
	Description     string `json:"description"`
	DisplayOrder    int    `json:"displayOrder"`
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

type nuvioBookingAdminCreateAppointmentPayload struct {
	WebsiteID             string `json:"websiteId"`
	ServiceID             string `json:"serviceId"`
	Date                  string `json:"date"`
	Time                  string `json:"time"`
	Name                  string `json:"name"`
	Email                 string `json:"email"`
	Phone                 string `json:"phone"`
	Notes                 string `json:"notes"`
	InternalNotes         string `json:"internalNotes"`
	Status                string `json:"status"`
	CreateContact         *bool  `json:"createContact"`
	SendConfirmationEmail *bool  `json:"sendConfirmationEmail"`
}

type nuvioBookingAdminRescheduleAppointmentPayload struct {
	ServiceID string `json:"serviceId"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	SendEmail *bool  `json:"sendEmail"`
}

type nuvioBookingAdminUpdateAppointmentStatusPayload struct {
	Status    string `json:"status"`
	SendEmail *bool  `json:"sendEmail"`
}

// NUVIO CUSTOM START: Booking MVP Phase 3 public booking endpoints.
func registerNuvioBookingRoutes(e *core.ServeEvent) {
	bookingGroup := e.Router.Group("/api/nuvio/booking")
	bookingAdminGroup := bookingGroup.Group("/admin").Bind(apis.RequireSuperuserAuth())

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

		slots, err := computeNuvioAvailableSlots(e.App, websiteID, serviceRecord, dateValue, config.Rules)
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

		serviceSnapshot := buildNuvioBookingServiceSnapshot(serviceRecord)
		serviceName := strings.TrimSpace(serviceSnapshot.Name)
		if serviceName == "" {
			serviceName = "Booking service"
		}

		confirmationMode := normalizeNuvioBookingConfirmationMode(config.ConfirmationMode)
		appointmentStatus := "pending"
		if confirmationMode == nuvioBookingConfirmationModeAuto {
			appointmentStatus = "confirmed"
		}

		var appointmentID string
		confirmedAt := ""
		transactionErr := e.App.RunInTransaction(func(txApp core.App) error {
			txServiceRecord, err := txApp.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
			if err != nil {
				return err
			}

			if strings.TrimSpace(txServiceRecord.GetString("website")) != websiteID || !isNuvioBookingServiceActive(txServiceRecord) {
				return sql.ErrNoRows
			}

			slots, err := computeNuvioAvailableSlots(txApp, websiteID, txServiceRecord, dateValue, config.Rules)
			if err != nil {
				return err
			}
			if !containsNuvioBookingSlot(slots, timeValue) {
				if timingErr := validateNuvioBookingSlotTiming(dateValue, timeValue, config.Rules); timingErr != nil {
					return timingErr
				}
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
			appointmentRecord.Set("status", appointmentStatus)
			setNuvioBookingAppointmentServiceSnapshot(
				appointmentRecord,
				appointmentsCollection,
				buildNuvioBookingServiceSnapshot(txServiceRecord),
			)
			if appointmentStatus == "confirmed" {
				nowISO := time.Now().UTC().Format(time.RFC3339)
				if appointmentsCollection.Fields.GetByName("confirmedAt") != nil {
					appointmentRecord.Set("confirmedAt", nowISO)
					confirmedAt = nowISO
				} else if appointmentsCollection.Fields.GetByName("confirmed_at") != nil {
					appointmentRecord.Set("confirmed_at", nowISO)
					confirmedAt = nowISO
				}
			}

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
			if errors.Is(transactionErr, errNuvioBookingDateOutsideWindow) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This date is outside the booking window.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingTimeTooSoon) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This time is too soon to book.",
				})
			}

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
			"status":        appointmentStatus,
		}
		if confirmedAt != "" {
			responsePayload["confirmedAt"] = confirmedAt
		}

		emailPayload := nuvioBookingCreateAppointmentPayload{
			WebsiteID: websiteID,
			ServiceID: serviceID,
			Date:      dateValue,
			Time:      timeValue,
			Name:      name,
			Email:     email,
			Phone:     phone,
			Notes:     notes,
		}

		if appointmentStatus == "confirmed" {
			sendErrors := []string{}

			attachments := []nuvioTransactionalEmailAttachment{}
			durationMinutes := serviceSnapshot.DurationMinutes
			if durationMinutes <= 0 {
				durationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
			}
			if durationMinutes <= 0 {
				e.App.Logger().Error(
					"NUVIO booking calendar duration parse failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"serviceId",
					serviceID,
				)
			} else {
				attachment, calendarErr := maybeBuildNuvioBookingICSAttachment(nuvioBookingCalendarInvitePayload{
					WebsiteName:        resolveWebsiteDisplayName(website),
					ServiceName:        serviceName,
					ServiceDescription: serviceSnapshot.Description,
					CustomerName:       name,
					CustomerEmail:      email,
					CustomerPhone:      phone,
					Date:               dateValue,
					Time:               timeValue,
					DurationMinutes:    durationMinutes,
					Notes:              notes,
					Location:           resolveNuvioBookingCalendarLocation(website),
					AppointmentID:      appointmentID,
				})
				if calendarErr != nil {
					e.App.Logger().Error(
						"NUVIO booking calendar attachment build failed",
						"websiteId",
						websiteID,
						"appointmentId",
						appointmentID,
						"error",
						calendarErr.Error(),
					)
				} else if attachment != nil {
					attachments = append(attachments, *attachment)
				}
			}

			if emailErr := sendNuvioBookingConfirmedVisitorEmail(
				e.Request.Context(),
				website,
				serviceName,
				emailPayload,
				attachments,
			); emailErr != nil {
				e.App.Logger().Error(
					"NUVIO booking visitor confirmation email send failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"error",
					emailErr.Error(),
				)
				sendErrors = append(sendErrors, "visitor confirmation failed")
			}

			if emailErr := sendNuvioBookingBusinessNotificationEmail(
				e.Request.Context(),
				website,
				config,
				serviceName,
				emailPayload,
			); emailErr != nil {
				e.App.Logger().Error(
					"NUVIO booking business notification send failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"error",
					emailErr.Error(),
				)
				sendErrors = append(sendErrors, "business notification failed")
			}

			if len(sendErrors) > 0 {
				responsePayload["warning"] = "Appointment confirmed, but email notifications are temporarily unavailable."
			}
		} else if emailErr := sendNuvioBookingEmails(
			e.Request.Context(),
			website,
			config,
			serviceName,
			emailPayload,
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

	bookingAdminGroup.POST("/appointments", func(e *core.RequestEvent) error {
		payload := nuvioBookingAdminCreateAppointmentPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid manual appointment payload.", nil)
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
		internalNotes := strings.TrimSpace(payload.InternalNotes)

		status := strings.ToLower(strings.TrimSpace(payload.Status))
		if status == "" {
			status = "confirmed"
		}
		if status != "pending" && status != "confirmed" {
			return e.BadRequestError("Status must be pending or confirmed.", nil)
		}

		createContact := true
		if payload.CreateContact != nil {
			createContact = *payload.CreateContact
		}

		sendConfirmationEmail := false
		if payload.SendConfirmationEmail != nil {
			sendConfirmationEmail = *payload.SendConfirmationEmail
		}

		website, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		serviceRecord, err := e.App.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO booking admin service lookup failed",
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
		serviceSnapshot := buildNuvioBookingServiceSnapshot(serviceRecord)
		if serviceSnapshot.Name != "" {
			serviceName = serviceSnapshot.Name
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

			slots, err := computeNuvioAvailableSlots(txApp, websiteID, txServiceRecord, dateValue, config.Rules)
			if err != nil {
				return err
			}
			if !containsNuvioBookingSlot(slots, timeValue) {
				if timingErr := validateNuvioBookingSlotTiming(dateValue, timeValue, config.Rules); timingErr != nil {
					return timingErr
				}
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
			appointmentRecord.Set("status", status)
			setNuvioBookingAppointmentServiceSnapshot(
				appointmentRecord,
				appointmentsCollection,
				buildNuvioBookingServiceSnapshot(txServiceRecord),
			)

			if status == "confirmed" {
				confirmedAt := time.Now().UTC().Format(time.RFC3339)
				if appointmentsCollection.Fields.GetByName("confirmedAt") != nil {
					appointmentRecord.Set("confirmedAt", confirmedAt)
				} else if appointmentsCollection.Fields.GetByName("confirmed_at") != nil {
					appointmentRecord.Set("confirmed_at", confirmedAt)
				}
			}

			if appointmentsCollection.Fields.GetByName("internalNotes") != nil {
				appointmentRecord.Set("internalNotes", internalNotes)
			} else if appointmentsCollection.Fields.GetByName("internal_notes") != nil {
				appointmentRecord.Set("internal_notes", internalNotes)
			}

			if err := txApp.Save(appointmentRecord); err != nil {
				return err
			}

			appointmentID = appointmentRecord.Id
			return nil
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingDateOutsideWindow) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This date is outside the booking window.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingTimeTooSoon) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This time is too soon to book.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingSlotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{
					"ok":    false,
					"error": "This time is no longer available. Please choose another time.",
				})
			}

			if errors.Is(transactionErr, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking admin appointment create failed",
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
			return e.InternalServerError("Unable to create appointment right now.", nil)
		}

		responsePayload := map[string]any{
			"ok":                    true,
			"appointmentId":         appointmentID,
			"status":                status,
			"contactCreated":        false,
			"confirmationEmailSent": false,
		}
		warnings := []string{}

		if createContact {
			if err := createNuvioBookingContactRecord(
				e.App,
				websiteID,
				name,
				email,
				phone,
				fmt.Sprintf("Manual booking - %s", serviceName),
				buildNuvioBookingContactMessage(serviceName, dateValue, timeValue, notes),
			); err != nil {
				e.App.Logger().Error(
					"NUVIO booking admin contact create failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"error",
					err.Error(),
				)
				warnings = append(warnings, "Appointment created, but contact sync is temporarily unavailable.")
			} else {
				responsePayload["contactCreated"] = true
			}
		}

		if sendConfirmationEmail {
			var emailErr error

			if status == "confirmed" {
				attachments := []nuvioTransactionalEmailAttachment{}
				durationMinutes := serviceSnapshot.DurationMinutes
				if durationMinutes <= 0 {
					durationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
				}
				if durationMinutes <= 0 {
					e.App.Logger().Error(
						"NUVIO booking calendar duration parse failed",
						"websiteId",
						websiteID,
						"appointmentId",
						appointmentID,
						"serviceId",
						serviceID,
					)
				} else {
					attachment, calendarErr := maybeBuildNuvioBookingICSAttachment(nuvioBookingCalendarInvitePayload{
						WebsiteName:        resolveWebsiteDisplayName(website),
						ServiceName:        serviceName,
						ServiceDescription: serviceSnapshot.Description,
						CustomerName:       name,
						CustomerEmail:      email,
						CustomerPhone:      phone,
						Date:               dateValue,
						Time:               timeValue,
						DurationMinutes:    durationMinutes,
						Notes:              notes,
						Location:           resolveNuvioBookingCalendarLocation(website),
						AppointmentID:      appointmentID,
					})
					if calendarErr != nil {
						e.App.Logger().Error(
							"NUVIO booking calendar attachment build failed",
							"websiteId",
							websiteID,
							"appointmentId",
							appointmentID,
							"error",
							calendarErr.Error(),
						)
					} else if attachment != nil {
						attachments = append(attachments, *attachment)
					}
				}

				emailErr = sendNuvioBookingConfirmedVisitorEmail(
					e.Request.Context(),
					website,
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
					attachments,
				)
			} else {
				visitorOnlyConfig := config
				visitorOnlyConfig.EmailNotifications = nuvioEmailNotificationsConfig{
					Enabled: false,
					To:      []string{},
					Cc:      []string{},
				}

				emailErr = sendNuvioBookingEmails(
					e.Request.Context(),
					website,
					visitorOnlyConfig,
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
				)
			}

			if emailErr != nil {
				e.App.Logger().Error(
					"NUVIO booking admin confirmation email failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"error",
					emailErr.Error(),
				)
				warnings = append(warnings, "Appointment created, but confirmation email could not be sent.")
			} else {
				responsePayload["confirmationEmailSent"] = true
			}
		}

		if len(warnings) > 0 {
			responsePayload["warning"] = strings.Join(warnings, " ")
		}

		return e.JSON(http.StatusOK, responsePayload)
	})

	bookingAdminGroup.POST("/appointments/{id}/reschedule", func(e *core.RequestEvent) error {
		appointmentID := strings.TrimSpace(e.Request.PathValue("id"))
		if appointmentID == "" {
			appointmentID = strings.TrimSpace(e.Request.URL.Query().Get("id"))
		}
		if appointmentID == "" {
			return e.BadRequestError("Missing appointment id.", nil)
		}

		payload := nuvioBookingAdminRescheduleAppointmentPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid reschedule payload.", nil)
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

		sendEmail := false
		if payload.SendEmail != nil {
			sendEmail = *payload.SendEmail
		}

		appointmentRecord, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Appointment not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking appointment lookup failed",
				"appointmentId",
				appointmentID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		currentStatus := strings.ToLower(strings.TrimSpace(appointmentRecord.GetString("status")))
		if currentStatus == "" {
			currentStatus = "pending"
		}
		if currentStatus != "pending" && currentStatus != "confirmed" && currentStatus != "cancelled" {
			currentStatus = "pending"
		}
		if currentStatus == "cancelled" {
			return e.BadRequestError("Cancelled appointments cannot be rescheduled.", nil)
		}

		websiteID := strings.TrimSpace(appointmentRecord.GetString("website"))
		if websiteID == "" {
			return e.BadRequestError("Appointment website is missing.", nil)
		}

		website, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		serviceRecord, err := e.App.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking reschedule service lookup failed",
				"appointmentId",
				appointmentID,
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

		newServiceSnapshot := buildNuvioBookingServiceSnapshot(serviceRecord)
		newServiceName := strings.TrimSpace(newServiceSnapshot.Name)
		if newServiceName == "" {
			newServiceName = "Booking service"
		}

		oldServiceID := strings.TrimSpace(appointmentRecord.GetString("service"))
		var oldServiceRecord *core.Record
		if oldServiceID != "" {
			if resolvedOldServiceRecord, oldServiceErr := e.App.FindRecordById(nuvioBookingServicesCollectionID, oldServiceID); oldServiceErr == nil {
				oldServiceRecord = resolvedOldServiceRecord
			}
		}
		oldServiceSnapshot := resolveNuvioBookingAppointmentServiceSnapshot(appointmentRecord, oldServiceRecord)
		oldServiceName := strings.TrimSpace(oldServiceSnapshot.Name)
		if oldServiceName == "" {
			oldServiceName = newServiceName
		}
		if oldServiceName == "" {
			oldServiceName = "Booking service"
		}

		oldDateValue := strings.TrimSpace(appointmentRecord.GetString("date"))
		oldTimeValue := strings.TrimSpace(appointmentRecord.GetString("time"))
		visitorName := strings.TrimSpace(appointmentRecord.GetString("name"))
		visitorPhone := strings.TrimSpace(appointmentRecord.GetString("phone"))
		visitorNotes := strings.TrimSpace(appointmentRecord.GetString("notes"))
		visitorEmail, hasVisitorEmail := normalizeNuvioEmail(appointmentRecord.GetString("email"))

		rescheduledAt := ""
		transactionErr := e.App.RunInTransaction(func(txApp core.App) error {
			txAppointmentRecord, err := txApp.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errNuvioBookingAppointmentNotFound
				}
				return err
			}

			txStatus := strings.ToLower(strings.TrimSpace(txAppointmentRecord.GetString("status")))
			if txStatus == "" {
				txStatus = "pending"
			}
			if txStatus != "pending" && txStatus != "confirmed" && txStatus != "cancelled" {
				txStatus = "pending"
			}
			if txStatus == "cancelled" {
				return errNuvioBookingAppointmentCancelled
			}

			txWebsiteID := strings.TrimSpace(txAppointmentRecord.GetString("website"))
			if txWebsiteID == "" {
				return errNuvioBookingAppointmentNotFound
			}

			txServiceRecord, err := txApp.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errNuvioBookingServiceNotFound
				}
				return err
			}

			if strings.TrimSpace(txServiceRecord.GetString("website")) != txWebsiteID || !isNuvioBookingServiceActive(txServiceRecord) {
				return errNuvioBookingServiceNotFound
			}

			slots, err := computeNuvioAvailableSlotsWithOptions(
				txApp,
				txWebsiteID,
				txServiceRecord,
				dateValue,
				config.Rules,
				nuvioBookingSlotComputationOptions{
					ExcludeAppointmentID: appointmentID,
				},
			)
			if err != nil {
				return err
			}
			if !containsNuvioBookingSlot(slots, timeValue) {
				if timingErr := validateNuvioBookingSlotTiming(dateValue, timeValue, config.Rules); timingErr != nil {
					return timingErr
				}
				return errNuvioBookingSlotUnavailable
			}

			txAppointmentRecord.Set("service", serviceID)
			txAppointmentRecord.Set("date", dateValue)
			txAppointmentRecord.Set("time", timeValue)
			appointmentsCollection, collectionErr := txApp.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
			if collectionErr == nil {
				setNuvioBookingAppointmentServiceSnapshot(
					txAppointmentRecord,
					appointmentsCollection,
					buildNuvioBookingServiceSnapshot(txServiceRecord),
				)
			}

			rescheduledAt = time.Now().UTC().Format(time.RFC3339)
			if collectionErr == nil {
				if appointmentsCollection.Fields.GetByName("rescheduledAt") != nil {
					txAppointmentRecord.Set("rescheduledAt", rescheduledAt)
				} else if appointmentsCollection.Fields.GetByName("rescheduled_at") != nil {
					txAppointmentRecord.Set("rescheduled_at", rescheduledAt)
				}
			}

			if err := txApp.Save(txAppointmentRecord); err != nil {
				return err
			}

			return nil
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingDateOutsideWindow) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This date is outside the booking window.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingTimeTooSoon) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"ok":    false,
					"error": "This time is too soon to book.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingSlotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{
					"ok":    false,
					"error": "This time is no longer available. Please choose another time.",
				})
			}

			if errors.Is(transactionErr, errNuvioBookingAppointmentCancelled) {
				return e.BadRequestError("Cancelled appointments cannot be rescheduled.", nil)
			}

			if errors.Is(transactionErr, errNuvioBookingAppointmentNotFound) {
				return e.NotFoundError("Appointment not found.", nil)
			}

			if errors.Is(transactionErr, errNuvioBookingServiceNotFound) {
				return e.NotFoundError("Service not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking appointment reschedule failed",
				"appointmentId",
				appointmentID,
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
			return e.InternalServerError("Unable to reschedule appointment right now.", nil)
		}

		responsePayload := map[string]any{
			"ok":            true,
			"appointmentId": appointmentID,
			"status":        currentStatus,
			"rescheduledAt": rescheduledAt,
			"emailSent":     false,
		}

		if sendEmail {
			if !hasVisitorEmail {
				responsePayload["warning"] = "Appointment rescheduled, but customer email is missing."
			} else {
				attachments := []nuvioTransactionalEmailAttachment{}
				durationMinutes := newServiceSnapshot.DurationMinutes
				if durationMinutes <= 0 {
					durationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
				}
				if durationMinutes <= 0 {
					e.App.Logger().Error(
						"NUVIO booking calendar duration parse failed",
						"websiteId",
						websiteID,
						"appointmentId",
						appointmentID,
						"serviceId",
						serviceID,
					)
				} else {
					attachment, calendarErr := maybeBuildNuvioBookingICSAttachment(nuvioBookingCalendarInvitePayload{
						WebsiteName:        resolveWebsiteDisplayName(website),
						ServiceName:        newServiceName,
						ServiceDescription: newServiceSnapshot.Description,
						CustomerName:       visitorName,
						CustomerEmail:      visitorEmail,
						CustomerPhone:      visitorPhone,
						Date:               dateValue,
						Time:               timeValue,
						DurationMinutes:    durationMinutes,
						Notes:              visitorNotes,
						Location:           resolveNuvioBookingCalendarLocation(website),
						AppointmentID:      appointmentID,
					})
					if calendarErr != nil {
						e.App.Logger().Error(
							"NUVIO booking calendar attachment build failed",
							"websiteId",
							websiteID,
							"appointmentId",
							appointmentID,
							"error",
							calendarErr.Error(),
						)
					} else if attachment != nil {
						attachments = append(attachments, *attachment)
					}
				}

				if emailErr := sendNuvioBookingRescheduleVisitorEmail(
					e.Request.Context(),
					website,
					visitorName,
					visitorEmail,
					oldServiceName,
					oldDateValue,
					oldTimeValue,
					newServiceName,
					dateValue,
					timeValue,
					attachments,
				); emailErr != nil {
					e.App.Logger().Error(
						"NUVIO booking reschedule email failed",
						"appointmentId",
						appointmentID,
						"websiteId",
						websiteID,
						"error",
						emailErr.Error(),
					)
					responsePayload["warning"] = "Appointment rescheduled, but confirmation email could not be sent."
				} else {
					responsePayload["emailSent"] = true
				}
			}
		}

		return e.JSON(http.StatusOK, responsePayload)
	})

	bookingAdminGroup.POST("/appointments/{id}/status", func(e *core.RequestEvent) error {
		appointmentID := strings.TrimSpace(e.Request.PathValue("id"))
		if appointmentID == "" {
			appointmentID = strings.TrimSpace(e.Request.URL.Query().Get("id"))
		}
		if appointmentID == "" {
			return e.BadRequestError("Missing appointment id.", nil)
		}

		payload := nuvioBookingAdminUpdateAppointmentStatusPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid appointment status payload.", nil)
		}

		nextStatus := strings.ToLower(strings.TrimSpace(payload.Status))
		if nextStatus != "pending" && nextStatus != "confirmed" && nextStatus != "cancelled" {
			return e.BadRequestError("Status must be pending, confirmed, or cancelled.", nil)
		}

		sendEmail := false
		if payload.SendEmail != nil {
			sendEmail = *payload.SendEmail
		}

		appointmentRecord, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Appointment not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking appointment lookup failed",
				"appointmentId",
				appointmentID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		websiteID := strings.TrimSpace(appointmentRecord.GetString("website"))
		if websiteID == "" {
			return e.BadRequestError("Appointment website is missing.", nil)
		}

		serviceID := strings.TrimSpace(appointmentRecord.GetString("service"))
		dateValue := strings.TrimSpace(appointmentRecord.GetString("date"))
		timeValue := strings.TrimSpace(appointmentRecord.GetString("time"))
		visitorName := strings.TrimSpace(appointmentRecord.GetString("name"))
		visitorPhone := strings.TrimSpace(appointmentRecord.GetString("phone"))
		visitorNotes := strings.TrimSpace(appointmentRecord.GetString("notes"))
		visitorEmail, hasVisitorEmail := normalizeNuvioEmail(appointmentRecord.GetString("email"))

		confirmedAt := ""
		cancelledAt := ""

		transactionErr := e.App.RunInTransaction(func(txApp core.App) error {
			txAppointmentRecord, err := txApp.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errNuvioBookingAppointmentNotFound
				}
				return err
			}

			txAppointmentRecord.Set("status", nextStatus)

			nowIso := time.Now().UTC().Format(time.RFC3339)
			appointmentsCollection, collectionErr := txApp.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
			if collectionErr == nil {
				if nextStatus == "confirmed" {
					confirmedAt = nowIso
					if appointmentsCollection.Fields.GetByName("confirmedAt") != nil {
						txAppointmentRecord.Set("confirmedAt", confirmedAt)
					} else if appointmentsCollection.Fields.GetByName("confirmed_at") != nil {
						txAppointmentRecord.Set("confirmed_at", confirmedAt)
					}
				} else if nextStatus == "cancelled" {
					cancelledAt = nowIso
					if appointmentsCollection.Fields.GetByName("cancelledAt") != nil {
						txAppointmentRecord.Set("cancelledAt", cancelledAt)
					} else if appointmentsCollection.Fields.GetByName("cancelled_at") != nil {
						txAppointmentRecord.Set("cancelled_at", cancelledAt)
					}
				}
			}

			if err := txApp.Save(txAppointmentRecord); err != nil {
				return err
			}

			return nil
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingAppointmentNotFound) {
				return e.NotFoundError("Appointment not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking appointment status update failed",
				"appointmentId",
				appointmentID,
				"status",
				nextStatus,
				"error",
				transactionErr.Error(),
			)
			return e.InternalServerError("Unable to update appointment status right now.", nil)
		}

		responsePayload := map[string]any{
			"ok":            true,
			"appointmentId": appointmentID,
			"status":        nextStatus,
			"emailSent":     false,
		}
		if confirmedAt != "" {
			responsePayload["confirmedAt"] = confirmedAt
		}
		if cancelledAt != "" {
			responsePayload["cancelledAt"] = cancelledAt
		}

		if nextStatus == "confirmed" && sendEmail {
			if !hasVisitorEmail {
				responsePayload["warning"] = "Appointment confirmed, but customer email is missing."
			} else {
				website, websiteErr := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID)
				if websiteErr != nil {
					e.App.Logger().Error(
						"NUVIO booking website lookup failed during status email",
						"appointmentId",
						appointmentID,
						"websiteId",
						websiteID,
						"error",
						websiteErr.Error(),
					)
				}

				var serviceRecord *core.Record
				if serviceID != "" {
					resolvedServiceRecord, serviceErr := e.App.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
					if serviceErr != nil {
						e.App.Logger().Error(
							"NUVIO booking service lookup failed during status email",
							"appointmentId",
							appointmentID,
							"websiteId",
							websiteID,
							"serviceId",
							serviceID,
							"error",
							serviceErr.Error(),
						)
					} else {
						serviceRecord = resolvedServiceRecord
					}
				}

				serviceSnapshot := resolveNuvioBookingAppointmentServiceSnapshot(appointmentRecord, serviceRecord)
				serviceName := strings.TrimSpace(serviceSnapshot.Name)
				if serviceName == "" {
					serviceName = "Booking service"
				}

				attachments := []nuvioTransactionalEmailAttachment{}
				durationMinutes := serviceSnapshot.DurationMinutes
				if durationMinutes <= 0 && serviceRecord != nil {
					durationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
				}

				if durationMinutes <= 0 {
					e.App.Logger().Error(
						"NUVIO booking calendar duration parse failed",
						"websiteId",
						websiteID,
						"appointmentId",
						appointmentID,
						"serviceId",
						serviceID,
					)
				} else {
					attachment, calendarErr := maybeBuildNuvioBookingICSAttachment(nuvioBookingCalendarInvitePayload{
						WebsiteName:        resolveWebsiteDisplayName(website),
						ServiceName:        serviceName,
						ServiceDescription: serviceSnapshot.Description,
						CustomerName:       visitorName,
						CustomerEmail:      visitorEmail,
						CustomerPhone:      visitorPhone,
						Date:               dateValue,
						Time:               timeValue,
						DurationMinutes:    durationMinutes,
						Notes:              visitorNotes,
						Location:           resolveNuvioBookingCalendarLocation(website),
						AppointmentID:      appointmentID,
					})
					if calendarErr != nil {
						e.App.Logger().Error(
							"NUVIO booking calendar attachment build failed",
							"websiteId",
							websiteID,
							"appointmentId",
							appointmentID,
							"error",
							calendarErr.Error(),
						)
					} else if attachment != nil {
						attachments = append(attachments, *attachment)
					}
				}

				if emailErr := sendNuvioBookingConfirmedVisitorEmail(
					e.Request.Context(),
					website,
					serviceName,
					nuvioBookingCreateAppointmentPayload{
						WebsiteID: websiteID,
						ServiceID: serviceID,
						Date:      dateValue,
						Time:      timeValue,
						Name:      visitorName,
						Email:     visitorEmail,
						Phone:     visitorPhone,
						Notes:     visitorNotes,
					},
					attachments,
				); emailErr != nil {
					e.App.Logger().Error(
						"NUVIO booking confirmation email failed during status update",
						"appointmentId",
						appointmentID,
						"websiteId",
						websiteID,
						"error",
						emailErr.Error(),
					)
					responsePayload["warning"] = "Appointment confirmed, but confirmation email could not be sent."
				} else {
					responsePayload["emailSent"] = true
				}
			}
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
		ConfirmationMode: nuvioBookingConfirmationModeRequest,
		EmailNotifications: nuvioEmailNotificationsConfig{
			Enabled: false,
			To:      []string{},
			Cc:      []string{},
		},
		Rules: nuvioBookingRulesConfig{
			MinNoticeHours:       0,
			BookingWindowDays:    0,
			BufferMinutes:        0,
			CalendarBlockingMode: nuvioBookingBlockingModeService,
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
		config.ConfirmationMode = normalizeNuvioBookingConfirmationMode(bookingSettings["confirmationMode"])

		legacyDestination := strings.TrimSpace(parseStringValue(bookingSettings["emailDestination"]))
		config.EmailNotifications = parseNuvioEmailNotificationsConfig(
			bookingSettings["emailNotifications"],
			legacyDestination,
		)
		config.Rules = parseNuvioBookingRulesConfig(bookingSettings["rules"])
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

func parseNuvioBookingRulesConfig(raw any) nuvioBookingRulesConfig {
	rules := nuvioBookingRulesConfig{
		MinNoticeHours:       0,
		BookingWindowDays:    0,
		BufferMinutes:        0,
		CalendarBlockingMode: nuvioBookingBlockingModeService,
	}

	settings, ok := toStringAnyMap(raw)
	if !ok {
		return rules
	}

	rules.MinNoticeHours = parseNuvioNonNegativeInt(settings["minNoticeHours"], 0)
	rules.BookingWindowDays = parseNuvioNonNegativeInt(settings["bookingWindowDays"], 0)
	rules.BufferMinutes = parseNuvioNonNegativeInt(settings["bufferMinutes"], 0)
	rules.CalendarBlockingMode = normalizeNuvioBookingCalendarBlockingMode(settings["calendarBlockingMode"])
	return rules
}

func normalizeNuvioBookingConfirmationMode(raw any) string {
	normalized := strings.TrimSpace(parseStringValue(raw))
	switch {
	case strings.EqualFold(normalized, nuvioBookingConfirmationModeAuto):
		return nuvioBookingConfirmationModeAuto
	case strings.EqualFold(normalized, nuvioBookingConfirmationModeRequest):
		return nuvioBookingConfirmationModeRequest
	default:
		return nuvioBookingConfirmationModeRequest
	}
}

func normalizeNuvioBookingCalendarBlockingMode(raw any) string {
	normalized := strings.ToLower(strings.TrimSpace(parseStringValue(raw)))
	switch normalized {
	case nuvioBookingBlockingModeService, nuvioBookingBlockingModeWebsite, nuvioBookingBlockingModeNone:
		return normalized
	default:
		return nuvioBookingBlockingModeService
	}
}

func parseNuvioNonNegativeInt(raw any, fallback int) int {
	if fallback < 0 {
		fallback = 0
	}

	switch typed := raw.(type) {
	case int:
		if typed < 0 {
			return 0
		}
		return typed
	case int64:
		if typed < 0 {
			return 0
		}
		return int(typed)
	case int32:
		if typed < 0 {
			return 0
		}
		return int(typed)
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return fallback
		}
		if typed != math.Trunc(typed) {
			return fallback
		}
		parsed := int(typed)
		if parsed < 0 {
			return 0
		}
		return parsed
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return fallback
		}
		if float64(typed) != math.Trunc(float64(typed)) {
			return fallback
		}
		parsed := int(typed)
		if parsed < 0 {
			return 0
		}
		return parsed
	case string:
		normalized := strings.TrimSpace(typed)
		if normalized == "" {
			return fallback
		}

		if !nuvioBookingIntegerValuePattern.MatchString(normalized) {
			return fallback
		}

		parsed, err := strconv.Atoi(normalized)
		if err != nil {
			return fallback
		}

		if parsed < 0 {
			return 0
		}
		return parsed
	default:
		normalized := strings.TrimSpace(parseStringValue(raw))
		if normalized == "" || !nuvioBookingIntegerValuePattern.MatchString(normalized) {
			return fallback
		}

		parsed, err := strconv.Atoi(normalized)
		if err != nil {
			return fallback
		}

		if parsed < 0 {
			return 0
		}
		return parsed
	}
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

	type sortableBookingService struct {
		dto        nuvioBookingPublicService
		createdKey string
	}

	services := make([]sortableBookingService, 0, len(records))
	for _, record := range records {
		duration, err := parseNuvioBookingServiceDuration(record)
		if err != nil {
			continue
		}

		services = append(services, sortableBookingService{
			dto: nuvioBookingPublicService{
				ID:              strings.TrimSpace(record.Id),
				Name:            strings.TrimSpace(record.GetString("name")),
				DurationMinutes: duration,
				Description:     strings.TrimSpace(record.GetString("description")),
				DisplayOrder:    parseNuvioNonNegativeInt(record.Get("displayOrder"), 0),
			},
			createdKey: strings.TrimSpace(record.GetString("created")),
		})
	}

	sort.SliceStable(services, func(i, j int) bool {
		first := services[i]
		second := services[j]

		if first.dto.DisplayOrder != second.dto.DisplayOrder {
			return first.dto.DisplayOrder < second.dto.DisplayOrder
		}

		firstName := strings.ToLower(strings.TrimSpace(first.dto.Name))
		secondName := strings.ToLower(strings.TrimSpace(second.dto.Name))
		if firstName != secondName {
			return firstName < secondName
		}

		firstCreated := strings.TrimSpace(first.createdKey)
		secondCreated := strings.TrimSpace(second.createdKey)
		if firstCreated != "" && secondCreated != "" && firstCreated != secondCreated {
			return firstCreated < secondCreated
		}

		return strings.TrimSpace(first.dto.ID) < strings.TrimSpace(second.dto.ID)
	})

	result := make([]nuvioBookingPublicService, 0, len(services))
	for _, service := range services {
		result = append(result, service.dto)
	}

	return result, nil
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

func resolveNuvioCollectionFieldNameByAliases(
	collection *core.Collection,
	aliases []string,
) string {
	if collection == nil {
		return ""
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}
		if collection.Fields.GetByName(fieldName) != nil {
			return fieldName
		}
	}

	return ""
}

func readNuvioBookingRecordStringByAliases(record *core.Record, aliases []string) string {
	if record == nil {
		return ""
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}
		if value := strings.TrimSpace(record.GetString(fieldName)); value != "" {
			return value
		}
	}

	return ""
}

func readNuvioBookingRecordPositiveIntByAliases(record *core.Record, aliases []string) int {
	if record == nil {
		return 0
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		value := parseNuvioNonNegativeInt(record.Get(fieldName), 0)
		if value > 0 {
			return value
		}
	}

	return 0
}

func buildNuvioBookingServiceSnapshot(serviceRecord *core.Record) nuvioBookingServiceSnapshot {
	snapshot := nuvioBookingServiceSnapshot{}

	if serviceRecord != nil {
		snapshot.Name = strings.TrimSpace(serviceRecord.GetString("name"))
		snapshot.Description = strings.TrimSpace(serviceRecord.GetString("description"))
		if durationMinutes, err := parseNuvioBookingServiceDuration(serviceRecord); err == nil && durationMinutes > 0 {
			snapshot.DurationMinutes = durationMinutes
		} else {
			parsedDuration := parseNuvioNonNegativeInt(serviceRecord.Get("durationMinutes"), 0)
			if parsedDuration > 0 {
				snapshot.DurationMinutes = parsedDuration
			}
		}
	}

	if snapshot.Name == "" {
		snapshot.Name = "Booking service"
	}

	return snapshot
}

func resolveNuvioBookingAppointmentServiceSnapshot(
	appointmentRecord *core.Record,
	serviceRecord *core.Record,
) nuvioBookingServiceSnapshot {
	fallbackSnapshot := buildNuvioBookingServiceSnapshot(serviceRecord)
	snapshot := nuvioBookingServiceSnapshot{
		Name:            readNuvioBookingRecordStringByAliases(appointmentRecord, nuvioServiceNameSnapshotAliases),
		DurationMinutes: readNuvioBookingRecordPositiveIntByAliases(appointmentRecord, nuvioServiceDurationSnapshotAliases),
		Description:     readNuvioBookingRecordStringByAliases(appointmentRecord, nuvioServiceDescSnapshotAliases),
	}

	if snapshot.Name == "" {
		snapshot.Name = fallbackSnapshot.Name
	}
	if snapshot.DurationMinutes <= 0 {
		snapshot.DurationMinutes = fallbackSnapshot.DurationMinutes
	}
	if snapshot.Description == "" {
		snapshot.Description = fallbackSnapshot.Description
	}

	return snapshot
}

func setNuvioBookingAppointmentServiceSnapshot(
	appointmentRecord *core.Record,
	collection *core.Collection,
	snapshot nuvioBookingServiceSnapshot,
) {
	if appointmentRecord == nil || collection == nil {
		return
	}

	if snapshotFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioServiceNameSnapshotAliases); snapshotFieldName != "" {
		name := strings.TrimSpace(snapshot.Name)
		if name == "" {
			name = "Booking service"
		}
		appointmentRecord.Set(snapshotFieldName, name)
	}

	if snapshotFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioServiceDurationSnapshotAliases); snapshotFieldName != "" && snapshot.DurationMinutes > 0 {
		appointmentRecord.Set(snapshotFieldName, snapshot.DurationMinutes)
	}

	if snapshotFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioServiceDescSnapshotAliases); snapshotFieldName != "" {
		appointmentRecord.Set(snapshotFieldName, strings.TrimSpace(snapshot.Description))
	}
}

func computeNuvioAvailableSlots(
	app core.App,
	websiteID string,
	serviceRecord *core.Record,
	dateValue string,
	rules nuvioBookingRulesConfig,
) ([]string, error) {
	return computeNuvioAvailableSlotsWithOptions(
		app,
		websiteID,
		serviceRecord,
		dateValue,
		rules,
		nuvioBookingSlotComputationOptions{},
	)
}

func computeNuvioAvailableSlotsWithOptions(
	app core.App,
	websiteID string,
	serviceRecord *core.Record,
	dateValue string,
	rules nuvioBookingRulesConfig,
	options nuvioBookingSlotComputationOptions,
) ([]string, error) {
	normalizedDate := strings.TrimSpace(dateValue)
	if !nuvioBookingDatePattern.MatchString(normalizedDate) {
		return nil, fmt.Errorf("invalid date format")
	}

	serviceID := strings.TrimSpace(serviceRecord.Id)
	if serviceID == "" {
		return nil, fmt.Errorf("missing service id")
	}

	durationMinutes, err := parseNuvioBookingServiceDuration(serviceRecord)
	if err != nil {
		return nil, err
	}

	location := getNuvioBookingLocation()
	bookingDate, err := parseNuvioBookingDateInLocation(normalizedDate, location)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(location)
	if isNuvioBookingDateOutsideWindow(bookingDate, now, rules.BookingWindowDays) {
		return []string{}, nil
	}

	dailyRanges, err := resolveNuvioBookingDailyRanges(app, websiteID, normalizedDate)
	if err != nil {
		return nil, err
	}
	if len(dailyRanges) == 0 {
		return []string{}, nil
	}

	candidateSlotMinutesSet := map[int]struct{}{}
	for _, dailyRange := range dailyRanges {
		windowSlots := generateNuvioBookingSlots(dailyRange.StartMinutes, dailyRange.EndMinutes, durationMinutes)
		for _, windowSlot := range windowSlots {
			windowSlotMinutes, parseErr := parseNuvioBookingHHMM(windowSlot)
			if parseErr != nil {
				continue
			}
			candidateSlotMinutesSet[windowSlotMinutes] = struct{}{}
		}
	}

	candidateSlotMinutes := make([]int, 0, len(candidateSlotMinutesSet))
	for minutes := range candidateSlotMinutesSet {
		candidateSlotMinutes = append(candidateSlotMinutes, minutes)
	}
	sort.Ints(candidateSlotMinutes)

	candidateSlots := make([]string, 0, len(candidateSlotMinutes))
	for _, minutes := range candidateSlotMinutes {
		candidateSlots = append(candidateSlots, formatNuvioBookingHHMM(minutes))
	}
	if len(candidateSlots) == 0 {
		return []string{}, nil
	}

	blockedIntervals, err := loadNuvioBlockedAppointmentIntervals(
		app,
		websiteID,
		serviceID,
		normalizedDate,
		durationMinutes,
		rules.BufferMinutes,
		rules.CalendarBlockingMode,
		options.ExcludeAppointmentID,
	)
	if err != nil {
		return nil, err
	}

	minAllowed := now
	if rules.MinNoticeHours > 0 {
		minAllowed = now.Add(time.Duration(rules.MinNoticeHours) * time.Hour)
	}

	filtered := make([]string, 0, len(candidateSlots))
	for _, slot := range candidateSlots {
		slotMinutes, err := parseNuvioBookingHHMM(slot)
		if err != nil {
			continue
		}

		slotStart := time.Date(
			bookingDate.Year(),
			bookingDate.Month(),
			bookingDate.Day(),
			slotMinutes/60,
			slotMinutes%60,
			0,
			0,
			location,
		)

		if slotStart.Before(now) {
			continue
		}

		if rules.MinNoticeHours > 0 && slotStart.Before(minAllowed) {
			continue
		}

		if shouldNuvioBookingSlotBeBlocked(slotMinutes, durationMinutes, blockedIntervals) {
			continue
		}

		filtered = append(filtered, slot)
	}

	return filtered, nil
}

func findNuvioBookingAvailabilityRecords(
	app core.App,
	websiteID string,
	dayOfWeek string,
) ([]*core.Record, error) {
	availabilityCollection, err := app.FindCachedCollectionByNameOrId(nuvioBookingAvailabilityCollectionID)
	if err != nil {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		availabilityCollection,
		"website={:website} && dayOfWeek={:dayOfWeek} && active=true",
		"+startTime,+endTime,+created",
		5000,
		0,
		dbx.Params{
			"website":   websiteID,
			"dayOfWeek": dayOfWeek,
		},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func resolveNuvioBookingDailyRanges(
	app core.App,
	websiteID string,
	dateValue string,
) ([]nuvioBookingDailyRange, error) {
	exceptionRecord, err := findNuvioActiveBookingExceptionRecord(app, websiteID, dateValue)
	if err != nil {
		return nil, err
	}

	if exceptionRecord != nil {
		exceptionType := strings.ToLower(strings.TrimSpace(exceptionRecord.GetString("type")))
		if exceptionType == "closed" {
			return []nuvioBookingDailyRange{}, nil
		}

		if exceptionType == "customhours" {
			startRaw := strings.TrimSpace(exceptionRecord.GetString("startTime"))
			endRaw := strings.TrimSpace(exceptionRecord.GetString("endTime"))

			startMinutes, startErr := parseNuvioBookingHHMM(startRaw)
			endMinutes, endErr := parseNuvioBookingHHMM(endRaw)
			if startErr != nil || endErr != nil || endMinutes <= startMinutes {
				app.Logger().Error(
					"NUVIO booking customHours exception is invalid",
					"websiteId",
					websiteID,
					"date",
					dateValue,
				)
				return []nuvioBookingDailyRange{}, nil
			}

			return []nuvioBookingDailyRange{{
				StartMinutes: startMinutes,
				EndMinutes:   endMinutes,
			}}, nil
		}

		return []nuvioBookingDailyRange{}, nil
	}

	dayOfWeek, err := dateToNuvioBookingDayOfWeek(dateValue)
	if err != nil {
		return nil, err
	}

	availabilityRecords, err := findNuvioBookingAvailabilityRecords(app, websiteID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	if len(availabilityRecords) == 0 {
		return []nuvioBookingDailyRange{}, nil
	}

	ranges := make([]nuvioBookingDailyRange, 0, len(availabilityRecords))
	for _, availabilityRecord := range availabilityRecords {
		startRaw := strings.TrimSpace(availabilityRecord.GetString("startTime"))
		endRaw := strings.TrimSpace(availabilityRecord.GetString("endTime"))

		startMinutes, startErr := parseNuvioBookingHHMM(startRaw)
		endMinutes, endErr := parseNuvioBookingHHMM(endRaw)
		if startErr != nil || endErr != nil || endMinutes <= startMinutes {
			app.Logger().Error(
				"NUVIO booking weekly availability row is invalid",
				"websiteId",
				websiteID,
				"dayOfWeek",
				dayOfWeek,
				"availabilityId",
				strings.TrimSpace(availabilityRecord.Id),
			)
			continue
		}

		ranges = append(ranges, nuvioBookingDailyRange{
			StartMinutes: startMinutes,
			EndMinutes:   endMinutes,
		})
	}

	if len(ranges) == 0 {
		return []nuvioBookingDailyRange{}, nil
	}

	sort.SliceStable(ranges, func(i, j int) bool {
		if ranges[i].StartMinutes != ranges[j].StartMinutes {
			return ranges[i].StartMinutes < ranges[j].StartMinutes
		}
		return ranges[i].EndMinutes < ranges[j].EndMinutes
	})

	return ranges, nil
}

func findNuvioActiveBookingExceptionRecord(
	app core.App,
	websiteID string,
	dateValue string,
) (*core.Record, error) {
	exceptionsCollection, err := app.FindCachedCollectionByNameOrId(nuvioBookingExceptionsCollectionID)
	if err != nil {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		exceptionsCollection,
		"website={:website} && date={:date} && active=true",
		"-updated,-created",
		10,
		0,
		dbx.Params{
			"website": websiteID,
			"date":    dateValue,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

func loadNuvioBlockedAppointmentIntervals(
	app core.App,
	websiteID string,
	serviceID string,
	dateValue string,
	serviceDurationMinutes int,
	bufferMinutes int,
	calendarBlockingMode string,
	excludeAppointmentID string,
) ([]nuvioBookingInterval, error) {
	if serviceDurationMinutes <= 0 {
		return []nuvioBookingInterval{}, nil
	}
	if bufferMinutes < 0 {
		bufferMinutes = 0
	}

	blockingMode := normalizeNuvioBookingCalendarBlockingMode(calendarBlockingMode)
	if blockingMode == nuvioBookingBlockingModeNone {
		return []nuvioBookingInterval{}, nil
	}

	appointmentsCollection, err := app.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
	if err != nil {
		return nil, err
	}

	filter := "website={:website} && date={:date}"
	params := dbx.Params{
		"website": websiteID,
		"date":    dateValue,
	}
	if blockingMode == nuvioBookingBlockingModeService {
		filter = "website={:website} && service={:service} && date={:date}"
		params["service"] = serviceID
	}

	records, err := app.FindRecordsByFilter(
		appointmentsCollection,
		filter,
		"-created",
		5000,
		0,
		params,
	)
	if err != nil {
		return nil, err
	}

	serviceDurationByID := map[string]int{}
	blocked := make([]nuvioBookingInterval, 0, len(records))
	for _, record := range records {
		if strings.TrimSpace(record.Id) != "" && strings.TrimSpace(record.Id) == strings.TrimSpace(excludeAppointmentID) {
			continue
		}

		status := strings.ToLower(strings.TrimSpace(record.GetString("status")))
		if status != "pending" && status != "confirmed" {
			continue
		}

		timeValue := strings.TrimSpace(record.GetString("time"))
		if !nuvioBookingTimePattern.MatchString(timeValue) {
			continue
		}

		appointmentStart, err := parseNuvioBookingHHMM(timeValue)
		if err != nil {
			continue
		}

		blockedDurationMinutes := serviceDurationMinutes
		if snapshotDuration := readNuvioBookingRecordPositiveIntByAliases(record, nuvioServiceDurationSnapshotAliases); snapshotDuration > 0 {
			blockedDurationMinutes = snapshotDuration
		} else {
			appointmentServiceID := strings.TrimSpace(record.GetString("service"))
			if appointmentServiceID != "" {
				if cachedDuration, ok := serviceDurationByID[appointmentServiceID]; ok {
					if cachedDuration > 0 {
						blockedDurationMinutes = cachedDuration
					}
				} else {
					resolvedDuration := 0
					if existingServiceRecord, findErr := app.FindRecordById(nuvioBookingServicesCollectionID, appointmentServiceID); findErr == nil {
						if parsedDuration, parseErr := parseNuvioBookingServiceDuration(existingServiceRecord); parseErr == nil && parsedDuration > 0 {
							resolvedDuration = parsedDuration
						} else if fallbackDuration := parseNuvioNonNegativeInt(existingServiceRecord.Get("durationMinutes"), 0); fallbackDuration > 0 {
							resolvedDuration = fallbackDuration
						}
					}

					serviceDurationByID[appointmentServiceID] = resolvedDuration
					if resolvedDuration > 0 {
						blockedDurationMinutes = resolvedDuration
					}
				}
			}
		}

		if blockedDurationMinutes <= 0 {
			continue
		}

		blockedStart := appointmentStart - bufferMinutes
		if blockedStart < 0 {
			blockedStart = 0
		}

		blockedEnd := appointmentStart + blockedDurationMinutes + bufferMinutes
		if blockedEnd > 24*60 {
			blockedEnd = 24 * 60
		}
		if blockedEnd <= blockedStart {
			continue
		}

		blocked = append(blocked, nuvioBookingInterval{
			StartMinutes: blockedStart,
			EndMinutes:   blockedEnd,
		})
	}

	sort.SliceStable(blocked, func(i, j int) bool {
		return blocked[i].StartMinutes < blocked[j].StartMinutes
	})

	return blocked, nil
}

func shouldNuvioBookingSlotBeBlocked(
	slotStartMinutes int,
	durationMinutes int,
	blockedIntervals []nuvioBookingInterval,
) bool {
	slotEndMinutes := slotStartMinutes + durationMinutes
	for _, blockedInterval := range blockedIntervals {
		if nuvioBookingIntervalsOverlap(slotStartMinutes, slotEndMinutes, blockedInterval.StartMinutes, blockedInterval.EndMinutes) {
			return true
		}
	}
	return false
}

func nuvioBookingIntervalsOverlap(
	startA int,
	endA int,
	startB int,
	endB int,
) bool {
	return startA < endB && startB < endA
}

func getNuvioBookingLocation() *time.Location {
	location := time.UTC
	if lisbon, err := time.LoadLocation("Europe/Lisbon"); err == nil && lisbon != nil {
		location = lisbon
	}
	return location
}

func parseNuvioBookingDateInLocation(dateValue string, location *time.Location) (time.Time, error) {
	if location == nil {
		location = getNuvioBookingLocation()
	}

	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateValue), location)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date value")
	}

	return date, nil
}

func isNuvioBookingDateOutsideWindow(
	bookingDate time.Time,
	now time.Time,
	bookingWindowDays int,
) bool {
	if bookingWindowDays <= 0 {
		return false
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	requestedDate := time.Date(bookingDate.Year(), bookingDate.Month(), bookingDate.Day(), 0, 0, 0, 0, now.Location())
	lastAllowedDate := today.AddDate(0, 0, bookingWindowDays)

	return requestedDate.After(lastAllowedDate)
}

func validateNuvioBookingSlotTiming(
	dateValue string,
	timeValue string,
	rules nuvioBookingRulesConfig,
) error {
	location := getNuvioBookingLocation()
	bookingDate, err := parseNuvioBookingDateInLocation(dateValue, location)
	if err != nil {
		return fmt.Errorf("invalid date value")
	}

	now := time.Now().In(location)
	if isNuvioBookingDateOutsideWindow(bookingDate, now, rules.BookingWindowDays) {
		return errNuvioBookingDateOutsideWindow
	}

	slotMinutes, err := parseNuvioBookingHHMM(timeValue)
	if err != nil {
		return fmt.Errorf("invalid time value")
	}

	slotStart := time.Date(
		bookingDate.Year(),
		bookingDate.Month(),
		bookingDate.Day(),
		slotMinutes/60,
		slotMinutes%60,
		0,
		0,
		location,
	)

	if slotStart.Before(now) {
		return errNuvioBookingTimeTooSoon
	}

	if rules.MinNoticeHours > 0 {
		minAllowed := now.Add(time.Duration(rules.MinNoticeHours) * time.Hour)
		if slotStart.Before(minAllowed) {
			return errNuvioBookingTimeTooSoon
		}
	}

	return nil
}

func dateToNuvioBookingDayOfWeek(dateValue string) (string, error) {
	location := getNuvioBookingLocation()
	date, err := parseNuvioBookingDateInLocation(dateValue, location)
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

func createNuvioBookingContactRecord(
	app core.App,
	websiteID string,
	name string,
	email string,
	phone string,
	subject string,
	message string,
) error {
	contactsCollection, err := app.FindCachedCollectionByNameOrId(nuvioContactsCollectionID)
	if err != nil {
		return err
	}

	contactRecord := core.NewRecord(contactsCollection)
	contactRecord.Set("website", strings.TrimSpace(websiteID))
	contactRecord.Set("channel", "booking")
	contactRecord.Set("name", strings.TrimSpace(name))
	contactRecord.Set("email", strings.TrimSpace(email))
	contactRecord.Set("phone", strings.TrimSpace(phone))
	contactRecord.Set("subject", strings.TrimSpace(subject))
	contactRecord.Set("message", strings.TrimSpace(message))
	contactRecord.Set("status", "new")

	return app.Save(contactRecord)
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

func maybeBuildNuvioBookingICSAttachment(
	payload nuvioBookingCalendarInvitePayload,
) (*nuvioTransactionalEmailAttachment, error) {
	content, err := buildNuvioBookingICSContent(payload)
	if err != nil {
		return nil, err
	}

	filename := buildNuvioBookingICSFilename(payload.Date, payload.Time)
	return &nuvioTransactionalEmailAttachment{
		Filename:    filename,
		Content:     content,
		ContentType: "text/calendar; charset=utf-8; method=REQUEST",
	}, nil
}

func resolveNuvioBookingCalendarLocation(website *core.Record) string {
	if website == nil {
		return ""
	}

	candidates := []string{
		website.GetString("businessAddress"),
		website.GetString("address"),
		website.GetString("local_business_address"),
		website.GetString("localBusinessAddress"),
	}

	for _, candidate := range candidates {
		if normalized := strings.TrimSpace(candidate); normalized != "" {
			return normalized
		}
	}

	return ""
}

func sendNuvioBookingConfirmedVisitorEmail(
	ctx context.Context,
	website *core.Record,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
	attachments []nuvioTransactionalEmailAttachment,
) error {
	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	visitorEmail, ok := normalizeNuvioEmail(payload.Email)
	if !ok {
		return fmt.Errorf("invalid visitor email")
	}

	websiteName := resolveWebsiteDisplayName(website)
	if websiteName == "" {
		websiteName = "Website"
	}

	lines := []string{
		fmt.Sprintf("Hi %s,", strings.TrimSpace(payload.Name)),
		"",
		"Your appointment is confirmed.",
		fmt.Sprintf("Service: %s", strings.TrimSpace(serviceName)),
		fmt.Sprintf("Date: %s", strings.TrimSpace(payload.Date)),
		fmt.Sprintf("Time: %s", strings.TrimSpace(payload.Time)),
		fmt.Sprintf("Website: %s", websiteName),
	}

	if notes := strings.TrimSpace(payload.Notes); notes != "" {
		lines = append(lines, "", "Notes received:", notes)
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:          []string{visitorEmail},
		Subject:     "Appointment confirmed",
		Text:        strings.Join(lines, "\n"),
		Attachments: attachments,
	})
}

func sendNuvioBookingBusinessNotificationEmail(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteBookingConfig,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
) error {
	if !config.EmailNotifications.Enabled || len(config.EmailNotifications.To) == 0 {
		return nil
	}

	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	visitorEmail, ok := normalizeNuvioEmail(payload.Email)
	if !ok {
		return fmt.Errorf("invalid visitor email")
	}

	websiteName := resolveWebsiteDisplayName(website)
	if websiteName == "" {
		websiteName = "Website"
	}

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

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      config.EmailNotifications.To,
		Cc:      config.EmailNotifications.Cc,
		ReplyTo: []string{visitorEmail},
		Subject: fmt.Sprintf("New booking request - %s", websiteName),
		Text:    strings.Join(businessLines, "\n"),
	})
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

func sendNuvioBookingRescheduleVisitorEmail(
	ctx context.Context,
	website *core.Record,
	visitorName string,
	visitorEmail string,
	oldServiceName string,
	oldDate string,
	oldTime string,
	newServiceName string,
	newDate string,
	newTime string,
	attachments []nuvioTransactionalEmailAttachment,
) error {
	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	normalizedEmail, ok := normalizeNuvioEmail(visitorEmail)
	if !ok {
		return fmt.Errorf("invalid visitor email")
	}

	websiteName := resolveWebsiteDisplayName(website)
	if websiteName == "" {
		websiteName = "Website"
	}

	name := strings.TrimSpace(visitorName)
	if name == "" {
		name = "there"
	}

	lines := []string{
		fmt.Sprintf("Hi %s,", name),
		"",
		"Your appointment was rescheduled.",
		"",
		fmt.Sprintf("Previous service: %s", strings.TrimSpace(oldServiceName)),
		fmt.Sprintf("Previous date: %s", strings.TrimSpace(oldDate)),
		fmt.Sprintf("Previous time: %s", strings.TrimSpace(oldTime)),
		"",
		fmt.Sprintf("New service: %s", strings.TrimSpace(newServiceName)),
		fmt.Sprintf("New date: %s", strings.TrimSpace(newDate)),
		fmt.Sprintf("New time: %s", strings.TrimSpace(newTime)),
		fmt.Sprintf("Website: %s", websiteName),
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:          []string{normalizedEmail},
		Subject:     "Appointment rescheduled",
		Text:        strings.Join(lines, "\n"),
		Attachments: attachments,
	})
}

// NUVIO CUSTOM END: Booking MVP Phase 3 public booking endpoints.
