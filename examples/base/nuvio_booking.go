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
	nuvioBookingServicesCollectionID       = "pbc_1661203700"
	nuvioBookingAvailabilityCollectionID   = "pbc_1661203800"
	nuvioBookingExceptionsCollectionID     = "pbc_1778803400"
	nuvioAppointmentsCollectionID          = "pbc_1661203900"
	nuvioBookingBackofficeDashboardMaxScan = 5000
	nuvioBookingBackofficeInternalNotesMax = 4000
	nuvioBookingConfirmationModeRequest    = "request"
	nuvioBookingConfirmationModeAuto       = "autoConfirm"
	nuvioBookingBlockingModeService        = "service"
	nuvioBookingBlockingModeWebsite        = "website"
	nuvioBookingBlockingModeNone           = "none"
)

var (
	nuvioBookingDatePattern                         = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	nuvioBookingTimePattern                         = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	nuvioBookingIntegerValuePattern                 = regexp.MustCompile(`^-?\d+$`)
	nuvioServiceNameSnapshotAliases                 = []string{"serviceNameSnapshot", "service_name_snapshot"}
	nuvioServiceDurationSnapshotAliases             = []string{"serviceDurationMinutesSnapshot", "service_duration_minutes_snapshot"}
	nuvioServiceDescSnapshotAliases                 = []string{"serviceDescriptionSnapshot", "service_description_snapshot"}
	nuvioBookingBackofficeServicesCollectionAliases = []string{
		nuvioBookingServicesCollectionID,
		"BookingServices",
		"booking_services",
		"bookingservices",
	}
	nuvioBookingBackofficeAvailabilityCollectionAliases = []string{
		nuvioBookingAvailabilityCollectionID,
		"BookingAvailability",
		"booking_availability",
		"bookingavailability",
	}
	nuvioBookingBackofficeExceptionsCollectionAliases = []string{
		nuvioBookingExceptionsCollectionID,
		"BookingExceptions",
		"bookingexceptions",
	}
	nuvioBookingBackofficeAppointmentsCollectionAliases = []string{
		nuvioAppointmentsCollectionID,
		"Appointments",
		"appointments",
	}
	nuvioBookingBackofficeServiceBufferBeforeAliases         = []string{"bufferBefore", "buffer_before", "bufferBeforeMinutes", "buffer_before_minutes"}
	nuvioBookingBackofficeServiceBufferAfterAliases          = []string{"bufferAfter", "buffer_after", "bufferAfterMinutes", "buffer_after_minutes"}
	nuvioBookingBackofficeServicePriceAliases                = []string{"price", "amount", "servicePrice", "service_price"}
	nuvioBookingBackofficeServiceCalendarBlockingModeAliases = []string{"calendarBlockingMode", "calendar_blocking_mode"}
	nuvioBookingBackofficeServiceAutoConfirmAliases          = []string{"autoConfirm", "auto_confirm"}
	nuvioBookingBackofficeAvailabilityServiceAliases         = []string{"service", "serviceId", "service_id"}
	nuvioBookingBackofficeAvailabilityCapacityAliases        = []string{"capacity", "slots", "maxSlots"}
	nuvioBookingBackofficeExceptionsServiceAliases           = []string{"service", "serviceId", "service_id"}
	nuvioBookingBackofficeExceptionsReasonAliases            = []string{"reason", "note", "notes"}
	nuvioBookingBackofficeAppointmentNotesAliases            = []string{"notes", "message"}
	nuvioBookingBackofficeAppointmentMessageAliases          = []string{"message", "notes"}
	nuvioBookingBackofficeAppointmentInternalNotesAliases    = []string{"internalNotes", "internal_notes"}
	nuvioBookingBackofficeAppointmentArchivedAtAliases       = []string{"archivedAt", "archived_at"}
	nuvioBookingBackofficeAppointmentConfirmedAtAliases      = []string{"confirmedAt", "confirmed_at"}
	nuvioBookingBackofficeAppointmentCancelledAtAliases      = []string{"cancelledAt", "cancelled_at"}
	nuvioBookingBackofficeAppointmentRescheduledAtAliases    = []string{"rescheduledAt", "rescheduled_at"}
	nuvioBookingBackofficeAppointmentDurationAliases         = []string{"duration", "durationMinutes", "duration_minutes"}
	nuvioBookingBackofficeSensitiveAppointmentAliases        = []string{
		"manageToken",
		"manage_token",
		"publicManageToken",
		"public_manage_token",
		"providerPayload",
		"provider_payload",
		"icsPayload",
		"ics_payload",
	}
	nuvioBookingAllowedDayOfWeekValues = map[string]struct{}{
		"mon": {},
		"tue": {},
		"wed": {},
		"thu": {},
		"fri": {},
		"sat": {},
		"sun": {},
	}
	nuvioBookingAllowedExceptionTypeValues = map[string]string{
		"closed":      "closed",
		"customhours": "customHours",
	}
	nuvioBookingBackofficeServiceCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid":            {},
		"website":              {},
		"name":                 {},
		"description":          {},
		"duration":             {},
		"durationminutes":      {},
		"bufferbefore":         {},
		"bufferafter":          {},
		"price":                {},
		"active":               {},
		"enabled":              {},
		"status":               {},
		"displayorder":         {},
		"calendarblockingmode": {},
		"autoconfirm":          {},
	}
	nuvioBookingBackofficeServiceUpdateAllowedPayloadKeys = map[string]struct{}{
		"name":                 {},
		"description":          {},
		"duration":             {},
		"durationminutes":      {},
		"bufferbefore":         {},
		"bufferafter":          {},
		"price":                {},
		"active":               {},
		"enabled":              {},
		"status":               {},
		"displayorder":         {},
		"calendarblockingmode": {},
		"autoconfirm":          {},
	}
	nuvioBookingBackofficeAvailabilityCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid": {},
		"website":   {},
		"service":   {},
		"serviceid": {},
		"dayofweek": {},
		"starttime": {},
		"endtime":   {},
		"enabled":   {},
		"active":    {},
		"capacity":  {},
	}
	nuvioBookingBackofficeAvailabilityUpdateAllowedPayloadKeys = map[string]struct{}{
		"service":   {},
		"serviceid": {},
		"dayofweek": {},
		"starttime": {},
		"endtime":   {},
		"enabled":   {},
		"active":    {},
		"capacity":  {},
	}
	nuvioBookingBackofficeExceptionCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid": {},
		"website":   {},
		"service":   {},
		"serviceid": {},
		"date":      {},
		"starttime": {},
		"endtime":   {},
		"type":      {},
		"status":    {},
		"reason":    {},
		"note":      {},
		"enabled":   {},
		"active":    {},
	}
	nuvioBookingBackofficeExceptionUpdateAllowedPayloadKeys = map[string]struct{}{
		"service":   {},
		"serviceid": {},
		"date":      {},
		"starttime": {},
		"endtime":   {},
		"type":      {},
		"status":    {},
		"reason":    {},
		"note":      {},
		"enabled":   {},
		"active":    {},
	}
	nuvioBookingBackofficeSettingsRulesAllowedPayloadKeys = map[string]struct{}{
		"websiteid": {},
		"rules":     {},
	}
	nuvioBookingBackofficeRulesAllowedKeys = map[string]struct{}{
		"minnoticehours":    {},
		"bookingwindowdays": {},
		"bufferminutes":     {},
	}
	errNuvioBookingSlotUnavailable      = errors.New("nuvio booking slot unavailable")
	errNuvioBookingDateOutsideWindow    = errors.New("nuvio booking date outside window")
	errNuvioBookingTimeTooSoon          = errors.New("nuvio booking time too soon")
	errNuvioBookingAppointmentNotFound  = errors.New("nuvio booking appointment not found")
	errNuvioBookingServiceNotFound      = errors.New("nuvio booking service not found")
	errNuvioBookingAppointmentCancelled = errors.New("nuvio booking appointment cancelled")
)

type nuvioWebsiteBookingConfig struct {
	FeatureAvailable             bool
	Enabled                      bool
	ConfirmationMode             string
	EmailNotifications           nuvioEmailNotificationsConfig
	BusinessNotificationTemplate nuvioBookingBusinessNotificationTemplateConfig
	VisitorEmails                nuvioBookingVisitorEmailsConfig
	Rules                        nuvioBookingRulesConfig
}

type nuvioBookingBusinessNotificationTemplateConfig struct {
	Enabled                   bool
	Subject                   string
	IntroText                 string
	FooterText                string
	IncludeAppointmentDetails bool
}

type nuvioBookingVisitorNotificationTemplateConfig struct {
	Enabled    bool
	Subject    string
	IntroText  string
	FooterText string
}

type nuvioBookingVisitorEmailsConfig struct {
	RequestTemplate      nuvioBookingVisitorNotificationTemplateConfig
	ConfirmationTemplate nuvioBookingVisitorNotificationTemplateConfig
	RescheduleTemplate   nuvioBookingVisitorNotificationTemplateConfig
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

type nuvioBookingBackofficeCreateAppointmentPayload struct {
	WebsiteID             string `json:"websiteId"`
	ServiceID             string `json:"serviceId"`
	Service               string `json:"service"`
	Date                  string `json:"date"`
	Time                  string `json:"time"`
	Name                  string `json:"name"`
	Email                 string `json:"email"`
	Phone                 string `json:"phone"`
	Notes                 string `json:"notes"`
	Message               string `json:"message"`
	InternalNotes         string `json:"internalNotes"`
	Status                string `json:"status"`
	CreateContact         *bool  `json:"createContact"`
	SendConfirmationEmail *bool  `json:"sendConfirmationEmail"`
}

type nuvioBookingBackofficeRescheduleAppointmentPayload struct {
	ServiceID string `json:"serviceId"`
	Service   string `json:"service"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	SendEmail *bool  `json:"sendEmail"`
}

type nuvioBookingBackofficeInternalNotesPayload struct {
	InternalNotes string `json:"internalNotes"`
}

type nuvioBookingBackofficeArchivePayload struct {
	Archived *bool `json:"archived"`
}

type nuvioBookingBackofficeDashboardWebsiteRulesDTO struct {
	MinNoticeHours       int    `json:"minNoticeHours"`
	BookingWindowDays    int    `json:"bookingWindowDays"`
	BufferMinutes        int    `json:"bufferMinutes"`
	CalendarBlockingMode string `json:"calendarBlockingMode"`
}

type nuvioBookingBackofficeDashboardWebsiteBookingDTO struct {
	FeatureAvailable           bool                                           `json:"featureAvailable"`
	Enabled                    bool                                           `json:"enabled"`
	ConfirmationMode           string                                         `json:"confirmationMode"`
	Rules                      nuvioBookingBackofficeDashboardWebsiteRulesDTO `json:"rules"`
	BusinessNotificationsReady bool                                           `json:"businessNotificationsReady"`
	UsingContactFormFallback   bool                                           `json:"usingContactFormFallback"`
}

type nuvioBookingBackofficeDashboardWebsiteDTO struct {
	ID          string                                           `json:"id"`
	DisplayName string                                           `json:"displayName"`
	Name        string                                           `json:"name,omitempty"`
	Title       string                                           `json:"title,omitempty"`
	Slug        string                                           `json:"slug,omitempty"`
	Booking     nuvioBookingBackofficeDashboardWebsiteBookingDTO `json:"booking"`
}

type nuvioBookingBackofficeDashboardServiceDTO struct {
	ID                   string  `json:"id"`
	Website              string  `json:"website"`
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	DurationMinutes      int     `json:"durationMinutes"`
	Duration             int     `json:"duration"`
	BufferBefore         int     `json:"bufferBefore"`
	BufferAfter          int     `json:"bufferAfter"`
	Price                float64 `json:"price"`
	Active               bool    `json:"active"`
	Enabled              bool    `json:"enabled"`
	Status               string  `json:"status"`
	DisplayOrder         int     `json:"displayOrder"`
	CalendarBlockingMode string  `json:"calendarBlockingMode"`
	AutoConfirm          bool    `json:"autoConfirm"`
	Created              string  `json:"created"`
	Updated              string  `json:"updated"`
}

type nuvioBookingBackofficeDashboardAvailabilityDTO struct {
	ID        string `json:"id"`
	Website   string `json:"website"`
	Service   string `json:"service"`
	DayOfWeek string `json:"dayOfWeek"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Active    bool   `json:"active"`
	Enabled   bool   `json:"enabled"`
	Status    string `json:"status"`
	Capacity  int    `json:"capacity"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type nuvioBookingBackofficeDashboardExceptionDTO struct {
	ID        string `json:"id"`
	Website   string `json:"website"`
	Service   string `json:"service"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	Note      string `json:"note"`
	Active    bool   `json:"active"`
	Enabled   bool   `json:"enabled"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type nuvioBookingBackofficeDashboardServiceSnapshotDTO struct {
	Name            string `json:"name"`
	DurationMinutes int    `json:"durationMinutes"`
	Description     string `json:"description"`
}

type nuvioBookingBackofficeDashboardAppointmentDTO struct {
	ID              string                                            `json:"id"`
	Website         string                                            `json:"website"`
	Service         string                                            `json:"service"`
	ServiceSnapshot nuvioBookingBackofficeDashboardServiceSnapshotDTO `json:"serviceSnapshot"`
	Status          string                                            `json:"status"`
	Name            string                                            `json:"name"`
	Email           string                                            `json:"email"`
	Phone           string                                            `json:"phone"`
	Date            string                                            `json:"date"`
	Time            string                                            `json:"time"`
	DurationMinutes int                                               `json:"durationMinutes"`
	Duration        int                                               `json:"duration"`
	Notes           string                                            `json:"notes"`
	Message         string                                            `json:"message"`
	InternalNotes   string                                            `json:"internalNotes"`
	ArchivedAt      string                                            `json:"archivedAt"`
	ConfirmedAt     string                                            `json:"confirmedAt"`
	CancelledAt     string                                            `json:"cancelledAt"`
	RescheduledAt   string                                            `json:"rescheduledAt"`
	Created         string                                            `json:"created"`
	Updated         string                                            `json:"updated"`
}

type nuvioBookingBackofficeDashboardDatasets struct {
	Services     []nuvioBookingBackofficeDashboardServiceDTO      `json:"services"`
	Availability []nuvioBookingBackofficeDashboardAvailabilityDTO `json:"availability"`
	Exceptions   []nuvioBookingBackofficeDashboardExceptionDTO    `json:"exceptions"`
	Appointments []nuvioBookingBackofficeDashboardAppointmentDTO  `json:"appointments"`
}

type nuvioBookingBackofficeDashboardAppointmentCapabilities struct {
	AllowedStatus         []string `json:"allowedStatus"`
	SupportsArchive       bool     `json:"supportsArchive"`
	SupportsInternalNotes bool     `json:"supportsInternalNotes"`
}

type nuvioBookingBackofficeDashboardServiceCapabilities struct {
	SupportsCreate     bool `json:"supportsCreate"`
	SupportsUpdate     bool `json:"supportsUpdate"`
	SupportsDeactivate bool `json:"supportsDeactivate"`
}

type nuvioBookingBackofficeDashboardAvailabilityCapabilities struct {
	SupportsCreate bool `json:"supportsCreate"`
	SupportsUpdate bool `json:"supportsUpdate"`
}

type nuvioBookingBackofficeDashboardExceptionsCapabilities struct {
	SupportsCreate bool     `json:"supportsCreate"`
	SupportsUpdate bool     `json:"supportsUpdate"`
	AllowedType    []string `json:"allowedType"`
}

type nuvioBookingBackofficeDashboardCapabilities struct {
	Appointments nuvioBookingBackofficeDashboardAppointmentCapabilities  `json:"appointments"`
	Services     nuvioBookingBackofficeDashboardServiceCapabilities      `json:"services"`
	Availability nuvioBookingBackofficeDashboardAvailabilityCapabilities `json:"availability"`
	Exceptions   nuvioBookingBackofficeDashboardExceptionsCapabilities   `json:"exceptions"`
}

type nuvioBookingBackofficeDashboardResponse struct {
	State        string                                      `json:"state"`
	WebsiteID    string                                      `json:"websiteId"`
	Website      nuvioBookingBackofficeDashboardWebsiteDTO   `json:"website"`
	Datasets     nuvioBookingBackofficeDashboardDatasets     `json:"datasets"`
	Capabilities nuvioBookingBackofficeDashboardCapabilities `json:"capabilities"`
}

// NUVIO CUSTOM START: Booking MVP Phase 3 public booking endpoints.
func registerNuvioBookingRoutes(e *core.ServeEvent) {
	bookingGroup := e.Router.Group("/api/nuvio/booking")
	bookingAdminGroup := bookingGroup.Group("/admin").Bind(apis.RequireAdminSuperuserAuth())
	bookingBackofficeGroup := bookingGroup.Group("/backoffice").Bind(apis.RequireSuperuserAuth())

	bookingBackofficeGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		websiteRecord, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Booking settings.", nil)
		}

		datasets, err := loadNuvioBookingBackofficeDashboardDatasets(e.App, websiteID)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice dashboard datasets load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Booking dashboard data.", nil)
		}

		response := nuvioBookingBackofficeDashboardResponse{
			State:     "ok",
			WebsiteID: websiteID,
			Website:   buildNuvioBookingBackofficeDashboardWebsiteDTO(websiteRecord, config),
			Datasets:  datasets,
			Capabilities: nuvioBookingBackofficeDashboardCapabilities{
				Appointments: nuvioBookingBackofficeDashboardAppointmentCapabilities{
					AllowedStatus:         []string{"pending", "confirmed", "cancelled"},
					SupportsArchive:       true,
					SupportsInternalNotes: true,
				},
				Services: nuvioBookingBackofficeDashboardServiceCapabilities{
					SupportsCreate:     true,
					SupportsUpdate:     true,
					SupportsDeactivate: true,
				},
				Availability: nuvioBookingBackofficeDashboardAvailabilityCapabilities{
					SupportsCreate: true,
					SupportsUpdate: true,
				},
				Exceptions: nuvioBookingBackofficeDashboardExceptionsCapabilities{
					SupportsCreate: true,
					SupportsUpdate: true,
					AllowedType:    []string{"closed", "customHours"},
				},
			},
		}

		return e.JSON(http.StatusOK, response)
	})

	bookingBackofficeGroup.POST("/appointments", func(e *core.RequestEvent) error {
		payload := nuvioBookingBackofficeCreateAppointmentPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid appointment payload.", nil)
		}

		websiteID := strings.TrimSpace(payload.WebsiteID)
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		serviceID := strings.TrimSpace(payload.ServiceID)
		if serviceID == "" {
			serviceID = strings.TrimSpace(payload.Service)
		}
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
		if notes == "" {
			notes = strings.TrimSpace(payload.Message)
		}
		internalNotes := strings.TrimSpace(payload.InternalNotes)
		if len([]rune(internalNotes)) > nuvioBookingBackofficeInternalNotesMax {
			return e.BadRequestError(fmt.Sprintf("Internal notes are too long. Maximum %d characters.", nuvioBookingBackofficeInternalNotesMax), nil)
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
				"NUVIO booking backoffice service lookup failed",
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

		status := strings.ToLower(strings.TrimSpace(payload.Status))
		if status == "" {
			if normalizeNuvioBookingConfirmationMode(config.ConfirmationMode) == nuvioBookingConfirmationModeAuto {
				status = "confirmed"
			} else {
				status = "pending"
			}
		}
		if status != "pending" && status != "confirmed" {
			return e.BadRequestError("Status must be pending or confirmed.", nil)
		}

		createContact := true
		if payload.CreateContact != nil {
			createContact = *payload.CreateContact
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

			internalNotesFieldName := resolveNuvioCollectionFieldNameByAliases(
				appointmentsCollection,
				nuvioBookingBackofficeAppointmentInternalNotesAliases,
			)
			if internalNotesFieldName != "" {
				appointmentRecord.Set(internalNotesFieldName, internalNotes)
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
					"data":    map[string]any{},
					"message": "This date is outside the booking window.",
					"status":  http.StatusBadRequest,
				})
			}

			if errors.Is(transactionErr, errNuvioBookingTimeTooSoon) {
				return e.JSON(http.StatusBadRequest, map[string]any{
					"data":    map[string]any{},
					"message": "This time is too soon to book.",
					"status":  http.StatusBadRequest,
				})
			}

			if errors.Is(transactionErr, errNuvioBookingSlotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{
					"data":    map[string]any{},
					"message": "This time is no longer available. Please choose another time.",
					"status":  http.StatusConflict,
				})
			}

			if errors.Is(transactionErr, sql.ErrNoRows) {
				return e.NotFoundError("Service not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO booking backoffice appointment create failed",
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
			"state": "ok",
		}

		createdRecord, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
		if err == nil {
			responsePayload["appointment"] = buildNuvioBookingBackofficeDashboardAppointmentDTO(
				createdRecord,
				serviceRecord,
			)
		}

		if createContact {
			serviceSnapshot := buildNuvioBookingServiceSnapshot(serviceRecord)
			serviceName := strings.TrimSpace(serviceSnapshot.Name)
			if serviceName == "" {
				serviceName = "Booking service"
			}
			if contactErr := createNuvioBookingContactRecord(
				e.App,
				websiteID,
				name,
				email,
				phone,
				fmt.Sprintf("Manual booking - %s", serviceName),
				buildNuvioBookingContactMessage(serviceName, dateValue, timeValue, notes),
			); contactErr != nil {
				e.App.Logger().Error(
					"NUVIO booking backoffice contact create failed",
					"websiteId",
					websiteID,
					"appointmentId",
					appointmentID,
					"error",
					contactErr.Error(),
				)
				responsePayload["warning"] = "Appointment created, but contact sync is temporarily unavailable."
			}
		}

		_ = website // keep parity with loaded config/website flow for future extensions.
		return e.JSON(http.StatusOK, responsePayload)
	})

	bookingBackofficeGroup.POST("/appointments/{id}/status", func(e *core.RequestEvent) error {
		appointmentsCollection, appointmentRecord, _, err := resolveNuvioBookingBackofficeAppointmentWriteTarget(e)
		if err != nil {
			return err
		}

		payload := map[string]any{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid appointment status payload.", nil)
		}

		if len(payload) == 0 {
			return e.BadRequestError("Status is required.", nil)
		}
		for key := range payload {
			if strings.TrimSpace(key) != "status" {
				return e.BadRequestError("Only status can be updated in this endpoint.", nil)
			}
		}

		nextStatus := strings.ToLower(strings.TrimSpace(parseStringValue(payload["status"])))
		if nextStatus != "pending" && nextStatus != "confirmed" && nextStatus != "cancelled" {
			return e.BadRequestError("Status must be pending, confirmed, or cancelled.", nil)
		}

		appointmentID := strings.TrimSpace(appointmentRecord.Id)
		transactionErr := e.App.RunInTransaction(func(txApp core.App) error {
			txAppointmentRecord, err := txApp.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errNuvioBookingAppointmentNotFound
				}
				return err
			}

			txAppointmentRecord.Set("status", nextStatus)
			nowISO := time.Now().UTC().Format(time.RFC3339)

			if appointmentsCollection.Fields.GetByName("confirmedAt") != nil && nextStatus == "confirmed" {
				txAppointmentRecord.Set("confirmedAt", nowISO)
			} else if appointmentsCollection.Fields.GetByName("confirmed_at") != nil && nextStatus == "confirmed" {
				txAppointmentRecord.Set("confirmed_at", nowISO)
			}
			if appointmentsCollection.Fields.GetByName("cancelledAt") != nil && nextStatus == "cancelled" {
				txAppointmentRecord.Set("cancelledAt", nowISO)
			} else if appointmentsCollection.Fields.GetByName("cancelled_at") != nil && nextStatus == "cancelled" {
				txAppointmentRecord.Set("cancelled_at", nowISO)
			}

			return txApp.Save(txAppointmentRecord)
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingAppointmentNotFound) {
				return e.NotFoundError("Appointment not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO booking backoffice appointment status update failed",
				"appointmentId",
				appointmentID,
				"status",
				nextStatus,
				"error",
				transactionErr.Error(),
			)
			return e.InternalServerError("Unable to update appointment status right now.", nil)
		}

		updatedAppointment, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
		if err != nil {
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":       "ok",
			"appointment": buildNuvioBookingBackofficeDashboardAppointmentDTO(updatedAppointment, nil),
		})
	})

	bookingBackofficeGroup.POST("/appointments/{id}/reschedule", func(e *core.RequestEvent) error {
		_, appointmentRecord, websiteID, err := resolveNuvioBookingBackofficeAppointmentWriteTarget(e)
		if err != nil {
			return err
		}

		payload := nuvioBookingBackofficeRescheduleAppointmentPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid reschedule payload.", nil)
		}

		dateValue := strings.TrimSpace(payload.Date)
		if !nuvioBookingDatePattern.MatchString(dateValue) {
			return e.BadRequestError("Date must use YYYY-MM-DD format.", nil)
		}
		timeValue := strings.TrimSpace(payload.Time)
		if !nuvioBookingTimePattern.MatchString(timeValue) {
			return e.BadRequestError("Time must use HH:mm format.", nil)
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

		serviceID := strings.TrimSpace(payload.ServiceID)
		if serviceID == "" {
			serviceID = strings.TrimSpace(payload.Service)
		}
		if serviceID == "" {
			serviceID = strings.TrimSpace(appointmentRecord.GetString("service"))
		}
		if serviceID == "" {
			return e.BadRequestError("Missing serviceId.", nil)
		}

		_, config, err := loadNuvioWebsiteBookingConfig(e.App, websiteID)
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
			return e.InternalServerError("Unable to load booking service right now.", nil)
		}
		if strings.TrimSpace(serviceRecord.GetString("website")) != websiteID {
			return e.NotFoundError("Service not found.", nil)
		}
		if !isNuvioBookingServiceActive(serviceRecord) {
			return e.BadRequestError("Service is not available.", nil)
		}

		appointmentID := strings.TrimSpace(appointmentRecord.Id)
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

			txServiceRecord, err := txApp.FindRecordById(nuvioBookingServicesCollectionID, serviceID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errNuvioBookingServiceNotFound
				}
				return err
			}
			if strings.TrimSpace(txServiceRecord.GetString("website")) != websiteID || !isNuvioBookingServiceActive(txServiceRecord) {
				return errNuvioBookingServiceNotFound
			}

			slots, err := computeNuvioAvailableSlotsWithOptions(
				txApp,
				websiteID,
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

				rescheduledAt := time.Now().UTC().Format(time.RFC3339)
				if appointmentsCollection.Fields.GetByName("rescheduledAt") != nil {
					txAppointmentRecord.Set("rescheduledAt", rescheduledAt)
				} else if appointmentsCollection.Fields.GetByName("rescheduled_at") != nil {
					txAppointmentRecord.Set("rescheduled_at", rescheduledAt)
				}
			}

			return txApp.Save(txAppointmentRecord)
		})

		if transactionErr != nil {
			if errors.Is(transactionErr, errNuvioBookingDateOutsideWindow) {
				return e.BadRequestError("This date is outside the booking window.", nil)
			}
			if errors.Is(transactionErr, errNuvioBookingTimeTooSoon) {
				return e.BadRequestError("This time is too soon to book.", nil)
			}
			if errors.Is(transactionErr, errNuvioBookingSlotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{
					"data":    map[string]any{},
					"message": "This time is no longer available. Please choose another time.",
					"status":  http.StatusConflict,
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
				"NUVIO booking backoffice appointment reschedule failed",
				"appointmentId",
				appointmentID,
				"websiteId",
				websiteID,
				"serviceId",
				serviceID,
				"error",
				transactionErr.Error(),
			)
			return e.InternalServerError("Unable to reschedule appointment right now.", nil)
		}

		updatedAppointment, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
		if err != nil {
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":       "ok",
			"appointment": buildNuvioBookingBackofficeDashboardAppointmentDTO(updatedAppointment, serviceRecord),
		})
	})

	bookingBackofficeGroup.PATCH("/appointments/{id}/internal-notes", func(e *core.RequestEvent) error {
		appointmentsCollection, appointmentRecord, _, err := resolveNuvioBookingBackofficeAppointmentWriteTarget(e)
		if err != nil {
			return err
		}

		payload := map[string]any{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid internal notes payload.", nil)
		}
		if len(payload) != 1 {
			return e.BadRequestError("Only internalNotes can be updated in this endpoint.", nil)
		}
		rawInternalNotes, hasInternalNotes := payload["internalNotes"]
		if !hasInternalNotes {
			return e.BadRequestError("Only internalNotes can be updated in this endpoint.", nil)
		}

		internalNotes := strings.TrimSpace(parseStringValue(rawInternalNotes))
		if len([]rune(internalNotes)) > nuvioBookingBackofficeInternalNotesMax {
			return e.BadRequestError(fmt.Sprintf("Internal notes are too long. Maximum %d characters.", nuvioBookingBackofficeInternalNotesMax), nil)
		}

		internalNotesFieldName := resolveNuvioCollectionFieldNameByAliases(
			appointmentsCollection,
			nuvioBookingBackofficeAppointmentInternalNotesAliases,
		)
		if internalNotesFieldName == "" {
			return e.BadRequestError("Internal notes are not supported for appointments.", nil)
		}

		appointmentRecord.Set(internalNotesFieldName, internalNotes)
		if err := e.App.Save(appointmentRecord); err != nil {
			return e.BadRequestError("Failed to update appointment.", nil)
		}

		updatedAppointment, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentRecord.Id)
		if err != nil {
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":       "ok",
			"appointment": buildNuvioBookingBackofficeDashboardAppointmentDTO(updatedAppointment, nil),
		})
	})

	bookingBackofficeGroup.PATCH("/appointments/{id}/archive", func(e *core.RequestEvent) error {
		appointmentsCollection, appointmentRecord, _, err := resolveNuvioBookingBackofficeAppointmentWriteTarget(e)
		if err != nil {
			return err
		}

		payload := nuvioBookingBackofficeArchivePayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid archive payload.", nil)
		}

		if payload.Archived == nil {
			return e.BadRequestError("Archived must be true or false.", nil)
		}

		archivedAtFieldName := resolveNuvioCollectionFieldNameByAliases(
			appointmentsCollection,
			nuvioBookingBackofficeAppointmentArchivedAtAliases,
		)
		if archivedAtFieldName == "" {
			return e.BadRequestError("Archive is not supported for appointments.", nil)
		}

		archivedAtValue := ""
		if *payload.Archived {
			archivedAtValue = time.Now().UTC().Format(time.RFC3339)
		}

		appointmentRecord.Set(archivedAtFieldName, archivedAtValue)
		if err := e.App.Save(appointmentRecord); err != nil {
			return e.BadRequestError("Failed to update appointment.", nil)
		}

		updatedAppointment, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentRecord.Id)
		if err != nil {
			return e.InternalServerError("Unable to load appointment right now.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":       "ok",
			"appointment": buildNuvioBookingBackofficeDashboardAppointmentDTO(updatedAppointment, nil),
		})
	})

	bookingBackofficeGroup.POST("/services", func(e *core.RequestEvent) error {
		servicesCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeServicesCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking services collection.", nil)
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeServiceCreateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
		if websiteID == "" {
			websiteID = strings.TrimSpace(parseStringValue(payload["website"]))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		record := core.NewRecord(servicesCollection)
		setNuvioBookingBackofficeRelationField(record, servicesCollection, []string{"website", "site"}, websiteID)
		if err := applyNuvioBookingBackofficeServicePayload(record, servicesCollection, payload, true); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice service create failed",
				"websiteId",
				websiteID,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to create service.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":   "ok",
			"service": buildNuvioBookingBackofficeDashboardServiceDTO(record),
		})
	})

	bookingBackofficeGroup.PATCH("/services/{id}", func(e *core.RequestEvent) error {
		servicesCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeServicesCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking services collection.", nil)
		}

		record, _, err := resolveNuvioBookingBackofficeRecordWriteTarget(e, servicesCollection, "Service not found.")
		if err != nil {
			return err
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one service field is required.", nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeServiceUpdateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if err := applyNuvioBookingBackofficeServicePayload(record, servicesCollection, payload, false); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice service update failed",
				"recordId",
				record.Id,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update service.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":   "ok",
			"service": buildNuvioBookingBackofficeDashboardServiceDTO(record),
		})
	})

	bookingBackofficeGroup.POST("/availability", func(e *core.RequestEvent) error {
		availabilityCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeAvailabilityCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking availability collection.", nil)
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeAvailabilityCreateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
		if websiteID == "" {
			websiteID = strings.TrimSpace(parseStringValue(payload["website"]))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		record := core.NewRecord(availabilityCollection)
		setNuvioBookingBackofficeRelationField(record, availabilityCollection, []string{"website", "site"}, websiteID)
		if err := applyNuvioBookingBackofficeAvailabilityPayload(e.App, record, availabilityCollection, payload, websiteID, true); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice availability create failed",
				"websiteId",
				websiteID,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to create availability window.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":        "ok",
			"availability": buildNuvioBookingBackofficeDashboardAvailabilityDTO(record),
		})
	})

	bookingBackofficeGroup.PATCH("/availability/{id}", func(e *core.RequestEvent) error {
		availabilityCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeAvailabilityCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking availability collection.", nil)
		}

		record, websiteID, err := resolveNuvioBookingBackofficeRecordWriteTarget(e, availabilityCollection, "Availability window not found.")
		if err != nil {
			return err
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one availability field is required.", nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeAvailabilityUpdateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if err := applyNuvioBookingBackofficeAvailabilityPayload(e.App, record, availabilityCollection, payload, websiteID, false); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice availability update failed",
				"recordId",
				record.Id,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update availability window.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":        "ok",
			"availability": buildNuvioBookingBackofficeDashboardAvailabilityDTO(record),
		})
	})

	bookingBackofficeGroup.POST("/exceptions", func(e *core.RequestEvent) error {
		exceptionsCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeExceptionsCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking exceptions collection.", nil)
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeExceptionCreateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
		if websiteID == "" {
			websiteID = strings.TrimSpace(parseStringValue(payload["website"]))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		record := core.NewRecord(exceptionsCollection)
		setNuvioBookingBackofficeRelationField(record, exceptionsCollection, []string{"website", "site"}, websiteID)
		if err := applyNuvioBookingBackofficeExceptionPayload(e.App, record, exceptionsCollection, payload, websiteID, true); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice exception create failed",
				"websiteId",
				websiteID,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to create booking exception.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":     "ok",
			"exception": buildNuvioBookingBackofficeDashboardExceptionDTO(record),
		})
	})

	bookingBackofficeGroup.PATCH("/exceptions/{id}", func(e *core.RequestEvent) error {
		exceptionsCollection, err := findNuvioBookingBackofficeCollectionByAliases(e.App, nuvioBookingBackofficeExceptionsCollectionAliases)
		if err != nil {
			return e.BadRequestError("Failed to resolve booking exceptions collection.", nil)
		}

		record, websiteID, err := resolveNuvioBookingBackofficeRecordWriteTarget(e, exceptionsCollection, "Booking exception not found.")
		if err != nil {
			return err
		}

		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one exception field is required.", nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeExceptionUpdateAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if err := applyNuvioBookingBackofficeExceptionPayload(e.App, record, exceptionsCollection, payload, websiteID, false); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(record); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice exception update failed",
				"recordId",
				record.Id,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update booking exception.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":     "ok",
			"exception": buildNuvioBookingBackofficeDashboardExceptionDTO(record),
		})
	})

	bookingBackofficeGroup.PATCH("/settings/rules", func(e *core.RequestEvent) error {
		payload, err := parseNuvioBookingBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if err := validateNuvioBookingBackofficePayloadKeys(payload, nuvioBookingBackofficeSettingsRulesAllowedPayloadKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		rawRules, ok := payload["rules"]
		if !ok {
			return e.BadRequestError("Missing rules object.", nil)
		}
		rulesPayload, ok := toStringAnyMap(rawRules)
		if !ok {
			return e.BadRequestError("Rules must be an object.", nil)
		}
		if len(rulesPayload) == 0 {
			return e.BadRequestError("At least one booking rule is required.", nil)
		}

		if err := validateNuvioBookingBackofficePayloadKeys(rulesPayload, nuvioBookingBackofficeRulesAllowedKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteRecord, err := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load website settings.", nil)
		}

		settings := parseNuvioSettingsObject(websiteRecord.Get("settings"))
		bookingSettings, _ := toStringAnyMap(settings["booking"])
		if bookingSettings == nil {
			bookingSettings = map[string]any{}
		}
		currentRules, _ := toStringAnyMap(bookingSettings["rules"])
		if currentRules == nil {
			currentRules = map[string]any{}
		}

		if rawMinNotice, hasMinNotice := rulesPayload["minNoticeHours"]; hasMinNotice {
			minNotice, ok := parseNuvioBookingBackofficeNonNegativeInt(rawMinNotice)
			if !ok {
				return e.BadRequestError("minNoticeHours must be a non-negative integer.", nil)
			}
			currentRules["minNoticeHours"] = minNotice
		}
		if rawBookingWindow, hasBookingWindow := rulesPayload["bookingWindowDays"]; hasBookingWindow {
			bookingWindow, ok := parseNuvioBookingBackofficeNonNegativeInt(rawBookingWindow)
			if !ok {
				return e.BadRequestError("bookingWindowDays must be a non-negative integer.", nil)
			}
			currentRules["bookingWindowDays"] = bookingWindow
		}
		if rawBuffer, hasBuffer := rulesPayload["bufferMinutes"]; hasBuffer {
			bufferMinutes, ok := parseNuvioBookingBackofficeNonNegativeInt(rawBuffer)
			if !ok {
				return e.BadRequestError("bufferMinutes must be a non-negative integer.", nil)
			}
			currentRules["bufferMinutes"] = bufferMinutes
		}

		bookingSettings["rules"] = currentRules
		settings["booking"] = bookingSettings
		websiteRecord.Set("settings", settings)

		if saveErr := e.App.Save(websiteRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO booking backoffice settings rules update failed",
				"websiteId",
				websiteID,
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update booking rules.", nil)
		}

		updatedWebsiteRecord, config, configErr := loadNuvioWebsiteBookingConfig(e.App, websiteID)
		if configErr != nil {
			return e.BadRequestError("Failed to load updated booking settings.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":     "ok",
			"websiteId": websiteID,
			"website":   buildNuvioBookingBackofficeDashboardWebsiteDTO(updatedWebsiteRecord, config),
		})
	})

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
		serviceDurationMinutes := serviceSnapshot.DurationMinutes
		if serviceDurationMinutes <= 0 {
			serviceDurationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
		}
		serviceDuration := formatNuvioBookingTemplateServiceDuration(serviceDurationMinutes)

		if appointmentStatus == "confirmed" {
			sendErrors := []string{}

			attachments := []nuvioTransactionalEmailAttachment{}
			durationMinutes := serviceDurationMinutes
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
				config,
				serviceName,
				serviceDuration,
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
				serviceDuration,
				appointmentStatus,
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
			serviceDuration,
			appointmentStatus,
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
			serviceDurationMinutes := serviceSnapshot.DurationMinutes
			if serviceDurationMinutes <= 0 {
				serviceDurationMinutes, _ = parseNuvioBookingServiceDuration(serviceRecord)
			}
			serviceDuration := formatNuvioBookingTemplateServiceDuration(serviceDurationMinutes)

			if status == "confirmed" {
				attachments := []nuvioTransactionalEmailAttachment{}
				durationMinutes := serviceDurationMinutes
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
					config,
					serviceName,
					serviceDuration,
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
					serviceDuration,
					status,
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
				serviceDuration := formatNuvioBookingTemplateServiceDuration(durationMinutes)
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
					config,
					visitorName,
					visitorEmail,
					visitorPhone,
					oldServiceName,
					oldDateValue,
					oldTimeValue,
					newServiceName,
					serviceDuration,
					dateValue,
					timeValue,
					currentStatus,
					visitorNotes,
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
				website := (*core.Record)(nil)
				bookingConfig := nuvioWebsiteBookingConfig{}
				resolvedWebsite, resolvedConfig, configErr := loadNuvioWebsiteBookingConfig(e.App, websiteID)
				if configErr != nil {
					e.App.Logger().Error(
						"NUVIO booking settings load failed during status email",
						"appointmentId",
						appointmentID,
						"websiteId",
						websiteID,
						"error",
						configErr.Error(),
					)
					if fallbackWebsite, websiteErr := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID); websiteErr == nil {
						website = fallbackWebsite
					} else {
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
				} else {
					website = resolvedWebsite
					bookingConfig = resolvedConfig
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
				serviceDuration := formatNuvioBookingTemplateServiceDuration(durationMinutes)

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
					bookingConfig,
					serviceName,
					serviceDuration,
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

func resolveNuvioBookingBackofficeAppointmentWriteTarget(
	e *core.RequestEvent,
) (*core.Collection, *core.Record, string, error) {
	appointmentID := strings.TrimSpace(e.Request.PathValue("id"))
	if appointmentID == "" {
		appointmentID = strings.TrimSpace(e.Request.URL.Query().Get("id"))
	}
	if appointmentID == "" {
		return nil, nil, "", e.BadRequestError("Missing appointment id.", nil)
	}

	appointmentsCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioAppointmentsCollectionID)
	if err != nil {
		return nil, nil, "", e.BadRequestError("Booking appointments are unavailable.", nil)
	}

	appointmentRecord, err := e.App.FindRecordById(nuvioAppointmentsCollectionID, appointmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, "", e.NotFoundError("Appointment not found.", nil)
		}
		e.App.Logger().Error(
			"NUVIO booking backoffice appointment lookup failed",
			"appointmentId",
			appointmentID,
			"error",
			err.Error(),
		)
		return nil, nil, "", e.InternalServerError("Unable to load appointment right now.", nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(appointmentRecord, "website", "site"))
	if websiteID == "" {
		websiteID = strings.TrimSpace(appointmentRecord.GetString("website"))
	}
	if websiteID == "" {
		return nil, nil, "", e.BadRequestError("Appointment website is missing.", nil)
	}

	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return nil, nil, "", err
	}

	return appointmentsCollection, appointmentRecord, websiteID, nil
}

func parseNuvioBookingBackofficePayloadMap(e *core.RequestEvent) (map[string]any, error) {
	payload := map[string]any{}
	if err := e.BindBody(&payload); err != nil {
		return nil, fmt.Errorf("Invalid request payload")
	}
	return payload, nil
}

func validateNuvioBookingBackofficePayloadKeys(
	payload map[string]any,
	allowed map[string]struct{},
) error {
	for key := range payload {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" {
			return fmt.Errorf("Invalid payload field")
		}
		if _, ok := allowed[normalizedKey]; !ok {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(key))
		}
	}
	return nil
}

func resolveNuvioBookingBackofficeRecordWriteTarget(
	e *core.RequestEvent,
	collection *core.Collection,
	notFoundMessage string,
) (*core.Record, string, error) {
	recordID := strings.TrimSpace(e.Request.PathValue("id"))
	if recordID == "" {
		return nil, "", e.BadRequestError("Missing record id.", nil)
	}
	if collection == nil {
		return nil, "", e.BadRequestError("Failed to resolve backoffice collection.", nil)
	}

	record, err := e.App.FindRecordById(collection.Id, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", e.NotFoundError(notFoundMessage, nil)
		}
		return nil, "", e.BadRequestError("Failed to load record.", nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		websiteID = strings.TrimSpace(record.GetString("website"))
	}
	if websiteID == "" {
		return nil, "", e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	if accessErr := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); accessErr != nil {
		return nil, "", accessErr
	}

	return record, websiteID, nil
}

func setNuvioBookingBackofficeRelationField(record *core.Record, collection *core.Collection, aliases []string, value string) {
	setNuvioBookingBackofficeStringField(record, collection, aliases, value)
}

func setNuvioBookingBackofficeStringField(record *core.Record, collection *core.Collection, aliases []string, value string) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return
	}

	record.Set(fieldName, strings.TrimSpace(value))
}

func setNuvioBookingBackofficeNumberField(record *core.Record, collection *core.Collection, aliases []string, value any) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return
	}

	record.Set(fieldName, value)
}

func setNuvioBookingBackofficeBooleanField(record *core.Record, collection *core.Collection, aliases []string, value bool) {
	if record == nil || collection == nil {
		return
	}

	updated := false
	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}
		if collection.Fields.GetByName(fieldName) == nil {
			continue
		}
		record.Set(fieldName, value)
		updated = true
	}

	if updated {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName != "" {
		record.Set(fieldName, value)
	}
}

func readNuvioBookingBackofficePayloadValue(payload map[string]any, keys ...string) (any, bool) {
	if len(payload) == 0 || len(keys) == 0 {
		return nil, false
	}

	for _, key := range keys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		if value, ok := payload[trimmedKey]; ok {
			return value, true
		}
	}

	for payloadKey, value := range payload {
		normalizedPayloadKey := strings.ToLower(strings.TrimSpace(payloadKey))
		if normalizedPayloadKey == "" {
			continue
		}
		for _, key := range keys {
			normalizedKey := strings.ToLower(strings.TrimSpace(key))
			if normalizedKey != "" && normalizedKey == normalizedPayloadKey {
				return value, true
			}
		}
	}

	return nil, false
}

func parseNuvioBookingBackofficeNonNegativeInt(raw any) (int, bool) {
	switch typed := raw.(type) {
	case int:
		return typed, typed >= 0
	case int64:
		return int(typed), typed >= 0
	case int32:
		return int(typed), typed >= 0
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || typed != math.Trunc(typed) || typed < 0 {
			return 0, false
		}
		return int(typed), true
	case float32:
		value := float64(typed)
		if math.IsNaN(value) || math.IsInf(value, 0) || value != math.Trunc(value) || value < 0 {
			return 0, false
		}
		return int(value), true
	case string:
		normalized := strings.TrimSpace(typed)
		if normalized == "" || !nuvioBookingIntegerValuePattern.MatchString(normalized) {
			return 0, false
		}
		value, err := strconv.Atoi(normalized)
		if err != nil || value < 0 {
			return 0, false
		}
		return value, true
	default:
		normalized := strings.TrimSpace(parseStringValue(raw))
		if normalized == "" || !nuvioBookingIntegerValuePattern.MatchString(normalized) {
			return 0, false
		}
		value, err := strconv.Atoi(normalized)
		if err != nil || value < 0 {
			return 0, false
		}
		return value, true
	}
}

func parseNuvioBookingBackofficeNonNegativeFloat(raw any) (float64, bool) {
	switch typed := raw.(type) {
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || typed < 0 {
			return 0, false
		}
		return typed, true
	case float32:
		value := float64(typed)
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return 0, false
		}
		return value, true
	case int:
		if typed < 0 {
			return 0, false
		}
		return float64(typed), true
	case int64:
		if typed < 0 {
			return 0, false
		}
		return float64(typed), true
	case int32:
		if typed < 0 {
			return 0, false
		}
		return float64(typed), true
	case string:
		normalized := strings.TrimSpace(typed)
		if normalized == "" {
			return 0, false
		}
		value, err := strconv.ParseFloat(normalized, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return 0, false
		}
		return value, true
	default:
		normalized := strings.TrimSpace(parseStringValue(raw))
		if normalized == "" {
			return 0, false
		}
		value, err := strconv.ParseFloat(normalized, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return 0, false
		}
		return value, true
	}
}

func parseNuvioBookingBackofficeDayOfWeek(raw any) (string, error) {
	value := strings.ToLower(strings.TrimSpace(parseStringValue(raw)))
	if value == "" {
		return "", fmt.Errorf("dayOfWeek is required")
	}
	if _, ok := nuvioBookingAllowedDayOfWeekValues[value]; !ok {
		return "", fmt.Errorf("Invalid dayOfWeek value")
	}
	return value, nil
}

func parseNuvioBookingBackofficeTimeString(raw any, required bool) (string, error) {
	value := strings.TrimSpace(parseStringValue(raw))
	if value == "" {
		if required {
			return "", fmt.Errorf("Time value is required")
		}
		return "", nil
	}
	if !nuvioBookingTimePattern.MatchString(value) {
		return "", fmt.Errorf("Time must use HH:mm format.")
	}
	return value, nil
}

func parseNuvioBookingBackofficeExceptionType(raw any) (string, error) {
	value := strings.ToLower(strings.TrimSpace(parseStringValue(raw)))
	if value == "" {
		return "", fmt.Errorf("Exception type is required")
	}
	normalized, ok := nuvioBookingAllowedExceptionTypeValues[value]
	if !ok {
		return "", fmt.Errorf("Exception type must be closed or customHours")
	}
	return normalized, nil
}

func parseNuvioBookingBackofficeServiceActiveFlag(payload map[string]any, defaultValue bool) (bool, bool, error) {
	if rawStatus, hasStatus := readNuvioBookingBackofficePayloadValue(payload, "status"); hasStatus {
		statusValue := strings.ToLower(strings.TrimSpace(parseStringValue(rawStatus)))
		switch statusValue {
		case "active":
			return true, true, nil
		case "inactive":
			return false, true, nil
		case "":
			return false, false, fmt.Errorf("Status is required")
		default:
			return false, false, fmt.Errorf("Status must be active or inactive")
		}
	}

	if rawActive, hasActive := readNuvioBookingBackofficePayloadValue(payload, "active"); hasActive {
		if activeValue, ok := parseBoolValue(rawActive); ok {
			return activeValue, true, nil
		}
		return false, false, fmt.Errorf("Active must be true or false")
	}

	if rawEnabled, hasEnabled := readNuvioBookingBackofficePayloadValue(payload, "enabled"); hasEnabled {
		if enabledValue, ok := parseBoolValue(rawEnabled); ok {
			return enabledValue, true, nil
		}
		return false, false, fmt.Errorf("Enabled must be true or false")
	}

	return defaultValue, false, nil
}

func parseNuvioBookingBackofficeBoolFlag(
	payload map[string]any,
	defaultValue bool,
	keys ...string,
) (bool, bool, error) {
	for _, key := range keys {
		if rawValue, hasValue := readNuvioBookingBackofficePayloadValue(payload, key); hasValue {
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return false, false, fmt.Errorf("%s must be true or false", strings.TrimSpace(key))
			}
			return value, true, nil
		}
	}
	return defaultValue, false, nil
}

func resolveNuvioBookingBackofficeRelationWebsiteID(record *core.Record) string {
	if record == nil {
		return ""
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID != "" {
		return websiteID
	}

	return strings.TrimSpace(record.GetString("website"))
}

func validateNuvioBookingBackofficeServiceBelongsToWebsite(app core.App, websiteID string, serviceID string) error {
	normalizedServiceID := strings.TrimSpace(serviceID)
	if normalizedServiceID == "" {
		return nil
	}

	servicesCollection, err := findNuvioBookingBackofficeCollectionByAliases(app, nuvioBookingBackofficeServicesCollectionAliases)
	if err != nil {
		return fmt.Errorf("Booking services are unavailable")
	}

	serviceRecord, err := app.FindRecordById(servicesCollection.Id, normalizedServiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("Service not found")
		}
		return fmt.Errorf("Failed to validate booking service")
	}

	serviceWebsiteID := resolveNuvioBookingBackofficeRelationWebsiteID(serviceRecord)
	if serviceWebsiteID == "" || serviceWebsiteID != strings.TrimSpace(websiteID) {
		return fmt.Errorf("Service must belong to the selected website")
	}

	return nil
}

func applyNuvioBookingBackofficeServicePayload(
	record *core.Record,
	collection *core.Collection,
	payload map[string]any,
	isCreate bool,
) error {
	if record == nil || collection == nil {
		return fmt.Errorf("Booking service is unavailable")
	}

	if rawName, hasName := readNuvioBookingBackofficePayloadValue(payload, "name"); hasName {
		name := strings.TrimSpace(parseStringValue(rawName))
		if name == "" {
			return fmt.Errorf("Service name is required")
		}
		setNuvioBookingBackofficeStringField(record, collection, []string{"name"}, name)
	} else if isCreate {
		return fmt.Errorf("Service name is required")
	}

	if rawDescription, hasDescription := readNuvioBookingBackofficePayloadValue(payload, "description"); hasDescription {
		setNuvioBookingBackofficeStringField(record, collection, []string{"description"}, strings.TrimSpace(parseStringValue(rawDescription)))
	}

	if rawDuration, hasDuration := readNuvioBookingBackofficePayloadValue(payload, "durationMinutes", "duration"); hasDuration {
		durationMinutes, ok := parseNuvioBookingBackofficeNonNegativeInt(rawDuration)
		if !ok || durationMinutes < 5 || durationMinutes > 480 {
			return fmt.Errorf("Duration must be an integer between 5 and 480 minutes")
		}
		setNuvioBookingBackofficeNumberField(record, collection, []string{"durationMinutes", "duration"}, durationMinutes)
	} else if isCreate {
		return fmt.Errorf("Duration is required")
	}

	if rawBufferBefore, hasBufferBefore := readNuvioBookingBackofficePayloadValue(payload, "bufferBefore"); hasBufferBefore {
		bufferBefore, ok := parseNuvioBookingBackofficeNonNegativeInt(rawBufferBefore)
		if !ok {
			return fmt.Errorf("bufferBefore must be a non-negative integer")
		}
		setNuvioBookingBackofficeNumberField(record, collection, nuvioBookingBackofficeServiceBufferBeforeAliases, bufferBefore)
	}

	if rawBufferAfter, hasBufferAfter := readNuvioBookingBackofficePayloadValue(payload, "bufferAfter"); hasBufferAfter {
		bufferAfter, ok := parseNuvioBookingBackofficeNonNegativeInt(rawBufferAfter)
		if !ok {
			return fmt.Errorf("bufferAfter must be a non-negative integer")
		}
		setNuvioBookingBackofficeNumberField(record, collection, nuvioBookingBackofficeServiceBufferAfterAliases, bufferAfter)
	}

	if rawPrice, hasPrice := readNuvioBookingBackofficePayloadValue(payload, "price"); hasPrice {
		priceValue, ok := parseNuvioBookingBackofficeNonNegativeFloat(rawPrice)
		if !ok {
			return fmt.Errorf("price must be a non-negative number")
		}
		setNuvioBookingBackofficeNumberField(record, collection, nuvioBookingBackofficeServicePriceAliases, priceValue)
	}

	if rawDisplayOrder, hasDisplayOrder := readNuvioBookingBackofficePayloadValue(payload, "displayOrder"); hasDisplayOrder {
		displayOrder, ok := parseNuvioBookingBackofficeNonNegativeInt(rawDisplayOrder)
		if !ok {
			return fmt.Errorf("displayOrder must be a non-negative integer")
		}
		setNuvioBookingBackofficeNumberField(record, collection, []string{"displayOrder", "priority"}, displayOrder)
	}

	if rawBlockingMode, hasBlockingMode := readNuvioBookingBackofficePayloadValue(payload, "calendarBlockingMode"); hasBlockingMode {
		normalizedMode := strings.ToLower(strings.TrimSpace(parseStringValue(rawBlockingMode)))
		switch normalizedMode {
		case nuvioBookingBlockingModeService, nuvioBookingBlockingModeWebsite, nuvioBookingBlockingModeNone:
			setNuvioBookingBackofficeStringField(record, collection, nuvioBookingBackofficeServiceCalendarBlockingModeAliases, normalizedMode)
		case "":
			return fmt.Errorf("calendarBlockingMode is required")
		default:
			return fmt.Errorf("calendarBlockingMode must be service, website, or none")
		}
	}

	if rawAutoConfirm, hasAutoConfirm := readNuvioBookingBackofficePayloadValue(payload, "autoConfirm"); hasAutoConfirm {
		autoConfirm, ok := parseBoolValue(rawAutoConfirm)
		if !ok {
			return fmt.Errorf("autoConfirm must be true or false")
		}
		setNuvioBookingBackofficeBooleanField(record, collection, nuvioBookingBackofficeServiceAutoConfirmAliases, autoConfirm)
	}

	activeValue, hasActive, activeErr := parseNuvioBookingBackofficeBoolFlag(payload, true, "active", "enabled")
	if activeErr != nil {
		return activeErr
	}
	if hasActive || isCreate {
		setNuvioBookingBackofficeBooleanField(record, collection, []string{"active", "enabled"}, activeValue)
	}

	return nil
}

func applyNuvioBookingBackofficeAvailabilityPayload(
	app core.App,
	record *core.Record,
	collection *core.Collection,
	payload map[string]any,
	websiteID string,
	isCreate bool,
) error {
	if record == nil || collection == nil {
		return fmt.Errorf("Booking availability is unavailable")
	}

	if rawService, hasService := readNuvioBookingBackofficePayloadValue(payload, "service", "serviceId"); hasService {
		serviceID := strings.TrimSpace(parseStringValue(rawService))
		if serviceID != "" {
			if err := validateNuvioBookingBackofficeServiceBelongsToWebsite(app, websiteID, serviceID); err != nil {
				return err
			}
		}
		setNuvioBookingBackofficeRelationField(record, collection, nuvioBookingBackofficeAvailabilityServiceAliases, serviceID)
	}

	resolvedDayOfWeek := strings.TrimSpace(record.GetString("dayOfWeek"))
	if rawDayOfWeek, hasDayOfWeek := readNuvioBookingBackofficePayloadValue(payload, "dayOfWeek"); hasDayOfWeek {
		dayOfWeek, err := parseNuvioBookingBackofficeDayOfWeek(rawDayOfWeek)
		if err != nil {
			return err
		}
		resolvedDayOfWeek = dayOfWeek
		setNuvioBookingBackofficeStringField(record, collection, []string{"dayOfWeek"}, dayOfWeek)
	} else if isCreate {
		return fmt.Errorf("dayOfWeek is required")
	}

	resolvedStartTime := strings.TrimSpace(record.GetString("startTime"))
	if rawStartTime, hasStartTime := readNuvioBookingBackofficePayloadValue(payload, "startTime"); hasStartTime {
		startTime, err := parseNuvioBookingBackofficeTimeString(rawStartTime, true)
		if err != nil {
			return err
		}
		resolvedStartTime = startTime
		setNuvioBookingBackofficeStringField(record, collection, []string{"startTime"}, startTime)
	} else if isCreate {
		return fmt.Errorf("startTime is required")
	}

	resolvedEndTime := strings.TrimSpace(record.GetString("endTime"))
	if rawEndTime, hasEndTime := readNuvioBookingBackofficePayloadValue(payload, "endTime"); hasEndTime {
		endTime, err := parseNuvioBookingBackofficeTimeString(rawEndTime, true)
		if err != nil {
			return err
		}
		resolvedEndTime = endTime
		setNuvioBookingBackofficeStringField(record, collection, []string{"endTime"}, endTime)
	} else if isCreate {
		return fmt.Errorf("endTime is required")
	}

	if strings.TrimSpace(resolvedDayOfWeek) == "" {
		return fmt.Errorf("dayOfWeek is required")
	}
	if strings.TrimSpace(resolvedStartTime) == "" || strings.TrimSpace(resolvedEndTime) == "" {
		return fmt.Errorf("Availability startTime and endTime are required")
	}

	startMinutes, endMinutes, err := parseNuvioBookingBackofficeTimeRange(resolvedStartTime, resolvedEndTime)
	if err != nil {
		return err
	}
	if startMinutes >= endMinutes {
		return fmt.Errorf("startTime must be before endTime")
	}

	if rawCapacity, hasCapacity := readNuvioBookingBackofficePayloadValue(payload, "capacity"); hasCapacity {
		capacityValue, ok := parseNuvioBookingBackofficeNonNegativeInt(rawCapacity)
		if !ok {
			return fmt.Errorf("capacity must be a non-negative integer")
		}
		setNuvioBookingBackofficeNumberField(record, collection, nuvioBookingBackofficeAvailabilityCapacityAliases, capacityValue)
	}

	activeValue, hasActive, activeErr := parseNuvioBookingBackofficeBoolFlag(payload, true, "active", "enabled")
	if activeErr != nil {
		return activeErr
	}
	if hasActive || isCreate {
		setNuvioBookingBackofficeBooleanField(record, collection, []string{"active", "enabled"}, activeValue)
	}

	return nil
}

func applyNuvioBookingBackofficeExceptionPayload(
	app core.App,
	record *core.Record,
	collection *core.Collection,
	payload map[string]any,
	websiteID string,
	isCreate bool,
) error {
	if record == nil || collection == nil {
		return fmt.Errorf("Booking exception is unavailable")
	}

	if rawService, hasService := readNuvioBookingBackofficePayloadValue(payload, "service", "serviceId"); hasService {
		serviceID := strings.TrimSpace(parseStringValue(rawService))
		if serviceID != "" {
			if err := validateNuvioBookingBackofficeServiceBelongsToWebsite(app, websiteID, serviceID); err != nil {
				return err
			}
		}
		setNuvioBookingBackofficeRelationField(record, collection, nuvioBookingBackofficeExceptionsServiceAliases, serviceID)
	}

	resolvedDate := strings.TrimSpace(record.GetString("date"))
	if rawDate, hasDate := readNuvioBookingBackofficePayloadValue(payload, "date"); hasDate {
		dateValue := strings.TrimSpace(parseStringValue(rawDate))
		if !nuvioBookingDatePattern.MatchString(dateValue) {
			return fmt.Errorf("Date must use YYYY-MM-DD format.")
		}
		resolvedDate = dateValue
		setNuvioBookingBackofficeStringField(record, collection, []string{"date"}, dateValue)
	} else if isCreate {
		return fmt.Errorf("Date is required")
	}

	resolvedType := strings.TrimSpace(record.GetString("type"))
	if rawType, hasType := readNuvioBookingBackofficePayloadValue(payload, "type"); hasType {
		typeValue, err := parseNuvioBookingBackofficeExceptionType(rawType)
		if err != nil {
			return err
		}
		resolvedType = typeValue
		setNuvioBookingBackofficeStringField(record, collection, []string{"type"}, typeValue)
	} else if rawStatusAsType, hasStatusAsType := readNuvioBookingBackofficePayloadValue(payload, "status"); hasStatusAsType {
		typeValue, err := parseNuvioBookingBackofficeExceptionType(rawStatusAsType)
		if err != nil {
			return err
		}
		resolvedType = typeValue
		setNuvioBookingBackofficeStringField(record, collection, []string{"type"}, typeValue)
	} else if isCreate {
		resolvedType = "closed"
		setNuvioBookingBackofficeStringField(record, collection, []string{"type"}, resolvedType)
	}

	resolvedStartTime := strings.TrimSpace(record.GetString("startTime"))
	resolvedEndTime := strings.TrimSpace(record.GetString("endTime"))

	if rawStartTime, hasStartTime := readNuvioBookingBackofficePayloadValue(payload, "startTime"); hasStartTime {
		startTime, err := parseNuvioBookingBackofficeTimeString(rawStartTime, false)
		if err != nil {
			return err
		}
		resolvedStartTime = startTime
		setNuvioBookingBackofficeStringField(record, collection, []string{"startTime"}, startTime)
	}

	if rawEndTime, hasEndTime := readNuvioBookingBackofficePayloadValue(payload, "endTime"); hasEndTime {
		endTime, err := parseNuvioBookingBackofficeTimeString(rawEndTime, false)
		if err != nil {
			return err
		}
		resolvedEndTime = endTime
		setNuvioBookingBackofficeStringField(record, collection, []string{"endTime"}, endTime)
	}

	switch strings.ToLower(strings.TrimSpace(resolvedType)) {
	case "customhours":
		resolvedType = "customHours"
		if strings.TrimSpace(resolvedStartTime) == "" || strings.TrimSpace(resolvedEndTime) == "" {
			return fmt.Errorf("Custom hours exceptions require startTime and endTime")
		}
		startMinutes, endMinutes, err := parseNuvioBookingBackofficeTimeRange(resolvedStartTime, resolvedEndTime)
		if err != nil {
			return err
		}
		if startMinutes >= endMinutes {
			return fmt.Errorf("startTime must be before endTime")
		}
	case "closed", "":
		resolvedType = "closed"
		resolvedStartTime = ""
		resolvedEndTime = ""
		setNuvioBookingBackofficeStringField(record, collection, []string{"startTime"}, "")
		setNuvioBookingBackofficeStringField(record, collection, []string{"endTime"}, "")
	default:
		return fmt.Errorf("Exception type must be closed or customHours")
	}

	if strings.TrimSpace(resolvedDate) == "" {
		return fmt.Errorf("Date is required")
	}

	setNuvioBookingBackofficeStringField(record, collection, []string{"type"}, resolvedType)

	if rawReason, hasReason := readNuvioBookingBackofficePayloadValue(payload, "reason"); hasReason {
		reason := strings.TrimSpace(parseStringValue(rawReason))
		setNuvioBookingBackofficeStringField(record, collection, nuvioBookingBackofficeExceptionsReasonAliases, reason)
	}

	if rawNote, hasNote := readNuvioBookingBackofficePayloadValue(payload, "note"); hasNote {
		note := strings.TrimSpace(parseStringValue(rawNote))
		setNuvioBookingBackofficeStringField(record, collection, []string{"note", "notes"}, note)
	}

	activeValue, hasActive, activeErr := parseNuvioBookingBackofficeServiceActiveFlag(payload, true)
	if activeErr != nil {
		return activeErr
	}
	if hasActive || isCreate {
		setNuvioBookingBackofficeBooleanField(record, collection, []string{"active", "enabled"}, activeValue)
	}

	return nil
}

func parseNuvioBookingBackofficeTimeRange(startTime string, endTime string) (int, int, error) {
	startValue := strings.TrimSpace(startTime)
	endValue := strings.TrimSpace(endTime)
	if !nuvioBookingTimePattern.MatchString(startValue) || !nuvioBookingTimePattern.MatchString(endValue) {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}

	startParts := strings.Split(startValue, ":")
	endParts := strings.Split(endValue, ":")
	if len(startParts) != 2 || len(endParts) != 2 {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}

	startHour, err := strconv.Atoi(startParts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}
	startMinute, err := strconv.Atoi(startParts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}
	endHour, err := strconv.Atoi(endParts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}
	endMinute, err := strconv.Atoi(endParts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Time must use HH:mm format.")
	}

	return (startHour * 60) + startMinute, (endHour * 60) + endMinute, nil
}

func loadNuvioBookingBackofficeDashboardDatasets(
	app core.App,
	websiteID string,
) (nuvioBookingBackofficeDashboardDatasets, error) {
	datasets := nuvioBookingBackofficeDashboardDatasets{
		Services:     []nuvioBookingBackofficeDashboardServiceDTO{},
		Availability: []nuvioBookingBackofficeDashboardAvailabilityDTO{},
		Exceptions:   []nuvioBookingBackofficeDashboardExceptionDTO{},
		Appointments: []nuvioBookingBackofficeDashboardAppointmentDTO{},
	}

	servicesCollection, err := findNuvioBookingBackofficeCollectionByAliases(app, nuvioBookingBackofficeServicesCollectionAliases)
	if err != nil {
		return datasets, err
	}
	availabilityCollection, err := findNuvioBookingBackofficeCollectionByAliases(app, nuvioBookingBackofficeAvailabilityCollectionAliases)
	if err != nil {
		return datasets, err
	}
	exceptionsCollection, err := findNuvioBookingBackofficeCollectionByAliases(app, nuvioBookingBackofficeExceptionsCollectionAliases)
	if err != nil {
		return datasets, err
	}
	appointmentsCollection, err := findNuvioBookingBackofficeCollectionByAliases(app, nuvioBookingBackofficeAppointmentsCollectionAliases)
	if err != nil {
		return datasets, err
	}

	servicesRecords, err := findNuvioBookingBackofficeRecordsByWebsite(
		app,
		servicesCollection,
		websiteID,
		"+name,+created",
	)
	if err != nil {
		return datasets, err
	}
	availabilityRecords, err := findNuvioBookingBackofficeRecordsByWebsite(
		app,
		availabilityCollection,
		websiteID,
		"+dayOfWeek,+startTime,+endTime,+created",
	)
	if err != nil {
		return datasets, err
	}
	exceptionsRecords, err := findNuvioBookingBackofficeRecordsByWebsite(
		app,
		exceptionsCollection,
		websiteID,
		"-date,-updated,-created",
	)
	if err != nil {
		return datasets, err
	}
	appointmentsRecords, err := findNuvioBookingBackofficeRecordsByWebsite(
		app,
		appointmentsCollection,
		websiteID,
		"-created",
	)
	if err != nil {
		return datasets, err
	}

	servicesByID := map[string]*core.Record{}
	for _, record := range servicesRecords {
		dto := buildNuvioBookingBackofficeDashboardServiceDTO(record)
		datasets.Services = append(datasets.Services, dto)
		if dto.ID != "" {
			servicesByID[dto.ID] = record
		}
	}

	for _, record := range availabilityRecords {
		datasets.Availability = append(datasets.Availability, buildNuvioBookingBackofficeDashboardAvailabilityDTO(record))
	}
	for _, record := range exceptionsRecords {
		datasets.Exceptions = append(datasets.Exceptions, buildNuvioBookingBackofficeDashboardExceptionDTO(record))
	}
	for _, record := range appointmentsRecords {
		datasets.Appointments = append(
			datasets.Appointments,
			buildNuvioBookingBackofficeDashboardAppointmentDTO(
				record,
				servicesByID[strings.TrimSpace(record.GetString("service"))],
			),
		)
	}

	return datasets, nil
}

func findNuvioBookingBackofficeCollectionByAliases(app core.App, aliases []string) (*core.Collection, error) {
	if len(aliases) == 0 {
		return nil, fmt.Errorf("missing collection aliases")
	}

	var firstErr error
	for _, alias := range aliases {
		identifier := strings.TrimSpace(alias)
		if identifier == "" {
			continue
		}

		collection, err := app.FindCachedCollectionByNameOrId(identifier)
		if err == nil && collection != nil {
			return collection, nil
		}

		if firstErr == nil && err != nil && !errors.Is(err, sql.ErrNoRows) {
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return nil, sql.ErrNoRows
}

func findNuvioBookingBackofficeRecordsByWebsite(
	app core.App,
	collection *core.Collection,
	websiteID string,
	sortExpr string,
) ([]*core.Record, error) {
	if collection == nil {
		return nil, fmt.Errorf("missing collection")
	}

	filter := "website={:websiteId}"
	params := dbx.Params{
		"websiteId": websiteID,
	}

	records, err := app.FindRecordsByFilter(
		collection,
		filter,
		sortExpr,
		nuvioBookingBackofficeDashboardMaxScan,
		0,
		params,
	)
	if err == nil {
		return records, nil
	}

	if strings.TrimSpace(sortExpr) != "" && strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
		return app.FindRecordsByFilter(
			collection,
			filter,
			"",
			nuvioBookingBackofficeDashboardMaxScan,
			0,
			params,
		)
	}

	return nil, err
}

func buildNuvioBookingBackofficeDashboardWebsiteDTO(
	website *core.Record,
	config nuvioWebsiteBookingConfig,
) nuvioBookingBackofficeDashboardWebsiteDTO {
	settings := parseNuvioSettingsObject(nil)
	if website != nil {
		settings = parseNuvioSettingsObject(website.Get("settings"))
	}

	bookingRules := config.Rules
	dto := nuvioBookingBackofficeDashboardWebsiteDTO{
		ID:          "",
		DisplayName: "",
		Name:        "",
		Title:       "",
		Slug:        "",
		Booking: nuvioBookingBackofficeDashboardWebsiteBookingDTO{
			FeatureAvailable: config.FeatureAvailable,
			Enabled:          config.Enabled,
			ConfirmationMode: config.ConfirmationMode,
			Rules: nuvioBookingBackofficeDashboardWebsiteRulesDTO{
				MinNoticeHours:       bookingRules.MinNoticeHours,
				BookingWindowDays:    bookingRules.BookingWindowDays,
				BufferMinutes:        bookingRules.BufferMinutes,
				CalendarBlockingMode: bookingRules.CalendarBlockingMode,
			},
			BusinessNotificationsReady: config.EmailNotifications.Enabled && len(config.EmailNotifications.To) > 0,
			UsingContactFormFallback:   resolveNuvioBookingBackofficeUsingContactFormFallback(settings),
		},
	}

	if website == nil {
		return dto
	}

	dto.ID = strings.TrimSpace(website.Id)
	dto.Name = strings.TrimSpace(website.GetString("name"))
	dto.Title = strings.TrimSpace(website.GetString("title"))
	dto.Slug = strings.TrimSpace(website.GetString("slug"))
	dto.DisplayName = resolveNuvioBookingBackofficeWebsiteDisplayName(website)

	return dto
}

func resolveNuvioBookingBackofficeUsingContactFormFallback(settings map[string]any) bool {
	bookingSettings, _ := toStringAnyMap(settings["booking"])
	contactSettings, _ := toStringAnyMap(settings["contactForm"])

	bookingNotifications := parseNuvioEmailNotificationsConfig(
		bookingSettings["emailNotifications"],
		strings.TrimSpace(parseStringValue(bookingSettings["emailDestination"])),
	)
	if len(bookingNotifications.To) > 0 {
		return false
	}

	contactNotifications := parseNuvioEmailNotificationsConfig(
		contactSettings["emailNotifications"],
		strings.TrimSpace(parseStringValue(contactSettings["emailDestination"])),
	)

	return len(contactNotifications.To) > 0
}

func resolveNuvioBookingBackofficeWebsiteDisplayName(website *core.Record) string {
	if website == nil {
		return ""
	}

	for _, fieldName := range []string{"title", "name", "slug"} {
		value := strings.TrimSpace(website.GetString(fieldName))
		if value != "" {
			return value
		}
	}

	return strings.TrimSpace(website.Id)
}

func buildNuvioBookingBackofficeDashboardServiceDTO(record *core.Record) nuvioBookingBackofficeDashboardServiceDTO {
	if record == nil {
		return nuvioBookingBackofficeDashboardServiceDTO{}
	}

	durationMinutes := parseNuvioNonNegativeInt(record.Get("durationMinutes"), 0)
	if durationMinutes <= 0 {
		if parsedDuration, err := parseNuvioBookingServiceDuration(record); err == nil && parsedDuration > 0 {
			durationMinutes = parsedDuration
		}
	}

	active := parseNuvioBookingBackofficeBoolByAliases(record, []string{"active", "enabled"}, true)
	status := "inactive"
	if active {
		status = "active"
	}

	return nuvioBookingBackofficeDashboardServiceDTO{
		ID:                   strings.TrimSpace(record.Id),
		Website:              strings.TrimSpace(record.GetString("website")),
		Name:                 strings.TrimSpace(record.GetString("name")),
		Description:          strings.TrimSpace(record.GetString("description")),
		DurationMinutes:      durationMinutes,
		Duration:             durationMinutes,
		BufferBefore:         parseNuvioBookingBackofficeIntByAliases(record, nuvioBookingBackofficeServiceBufferBeforeAliases, 0),
		BufferAfter:          parseNuvioBookingBackofficeIntByAliases(record, nuvioBookingBackofficeServiceBufferAfterAliases, 0),
		Price:                parseNuvioBookingBackofficeFloatByAliases(record, nuvioBookingBackofficeServicePriceAliases, 0),
		Active:               active,
		Enabled:              active,
		Status:               status,
		DisplayOrder:         parseNuvioNonNegativeInt(record.Get("displayOrder"), 0),
		CalendarBlockingMode: normalizeNuvioBookingCalendarBlockingMode(parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeServiceCalendarBlockingModeAliases)),
		AutoConfirm:          parseNuvioBookingBackofficeBoolByAliases(record, nuvioBookingBackofficeServiceAutoConfirmAliases, false),
		Created:              strings.TrimSpace(record.GetString("created")),
		Updated:              strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioBookingBackofficeDashboardAvailabilityDTO(record *core.Record) nuvioBookingBackofficeDashboardAvailabilityDTO {
	if record == nil {
		return nuvioBookingBackofficeDashboardAvailabilityDTO{}
	}

	active := parseNuvioBookingBackofficeBoolByAliases(record, []string{"active", "enabled"}, true)
	status := "inactive"
	if active {
		status = "active"
	}

	return nuvioBookingBackofficeDashboardAvailabilityDTO{
		ID:        strings.TrimSpace(record.Id),
		Website:   strings.TrimSpace(record.GetString("website")),
		Service:   parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAvailabilityServiceAliases),
		DayOfWeek: strings.TrimSpace(record.GetString("dayOfWeek")),
		StartTime: strings.TrimSpace(record.GetString("startTime")),
		EndTime:   strings.TrimSpace(record.GetString("endTime")),
		Active:    active,
		Enabled:   active,
		Status:    status,
		Capacity:  parseNuvioBookingBackofficeIntByAliases(record, nuvioBookingBackofficeAvailabilityCapacityAliases, 0),
		Created:   strings.TrimSpace(record.GetString("created")),
		Updated:   strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioBookingBackofficeDashboardExceptionDTO(record *core.Record) nuvioBookingBackofficeDashboardExceptionDTO {
	if record == nil {
		return nuvioBookingBackofficeDashboardExceptionDTO{}
	}

	exceptionType := strings.TrimSpace(record.GetString("type"))
	if strings.EqualFold(exceptionType, "customhours") {
		exceptionType = "customHours"
	}
	if exceptionType == "" {
		exceptionType = "closed"
	}

	active := parseNuvioBookingBackofficeBoolByAliases(record, []string{"active", "enabled"}, true)
	status := "inactive"
	if active {
		status = "active"
	}

	reason := parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeExceptionsReasonAliases)
	note := strings.TrimSpace(record.GetString("note"))
	if note == "" {
		note = reason
	}

	return nuvioBookingBackofficeDashboardExceptionDTO{
		ID:        strings.TrimSpace(record.Id),
		Website:   strings.TrimSpace(record.GetString("website")),
		Service:   parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeExceptionsServiceAliases),
		Date:      strings.TrimSpace(record.GetString("date")),
		StartTime: strings.TrimSpace(record.GetString("startTime")),
		EndTime:   strings.TrimSpace(record.GetString("endTime")),
		Type:      exceptionType,
		Status:    status,
		Reason:    reason,
		Note:      note,
		Active:    active,
		Enabled:   active,
		Created:   strings.TrimSpace(record.GetString("created")),
		Updated:   strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioBookingBackofficeDashboardAppointmentDTO(
	record *core.Record,
	serviceRecord *core.Record,
) nuvioBookingBackofficeDashboardAppointmentDTO {
	if record == nil {
		return nuvioBookingBackofficeDashboardAppointmentDTO{}
	}

	serviceID := strings.TrimSpace(record.GetString("service"))
	snapshot := resolveNuvioBookingAppointmentServiceSnapshot(record, serviceRecord)
	durationMinutes := readNuvioBookingRecordPositiveIntByAliases(record, nuvioBookingBackofficeAppointmentDurationAliases)
	if durationMinutes <= 0 {
		durationMinutes = snapshot.DurationMinutes
	}
	if durationMinutes <= 0 && serviceRecord != nil {
		if resolvedDuration, err := parseNuvioBookingServiceDuration(serviceRecord); err == nil && resolvedDuration > 0 {
			durationMinutes = resolvedDuration
		}
	}

	status := strings.ToLower(strings.TrimSpace(record.GetString("status")))
	if status != "confirmed" && status != "cancelled" {
		status = "pending"
	}

	snapshotDTO := nuvioBookingBackofficeDashboardServiceSnapshotDTO{
		Name:            strings.TrimSpace(snapshot.Name),
		DurationMinutes: snapshot.DurationMinutes,
		Description:     strings.TrimSpace(snapshot.Description),
	}
	if snapshotDTO.Name == "" {
		snapshotDTO.Name = "Booking service"
	}

	return nuvioBookingBackofficeDashboardAppointmentDTO{
		ID:              strings.TrimSpace(record.Id),
		Website:         strings.TrimSpace(record.GetString("website")),
		Service:         serviceID,
		ServiceSnapshot: snapshotDTO,
		Status:          status,
		Name:            strings.TrimSpace(record.GetString("name")),
		Email:           strings.TrimSpace(record.GetString("email")),
		Phone:           strings.TrimSpace(record.GetString("phone")),
		Date:            strings.TrimSpace(record.GetString("date")),
		Time:            strings.TrimSpace(record.GetString("time")),
		DurationMinutes: durationMinutes,
		Duration:        durationMinutes,
		Notes:           parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentNotesAliases),
		Message:         parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentMessageAliases),
		InternalNotes:   parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentInternalNotesAliases),
		ArchivedAt:      parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentArchivedAtAliases),
		ConfirmedAt:     parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentConfirmedAtAliases),
		CancelledAt:     parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentCancelledAtAliases),
		RescheduledAt:   parseNuvioBookingBackofficeStringByAliases(record, nuvioBookingBackofficeAppointmentRescheduledAtAliases),
		Created:         strings.TrimSpace(record.GetString("created")),
		Updated:         strings.TrimSpace(record.GetString("updated")),
	}
}

func parseNuvioBookingBackofficeStringByAliases(record *core.Record, aliases []string) string {
	if record == nil {
		return ""
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		value := strings.TrimSpace(record.GetString(fieldName))
		if value != "" {
			return value
		}
	}

	return ""
}

func parseNuvioBookingBackofficeIntByAliases(record *core.Record, aliases []string, fallback int) int {
	if record == nil {
		return fallback
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		value := parseNuvioNonNegativeInt(record.Get(fieldName), fallback)
		if value > 0 {
			return value
		}
	}

	return fallback
}

func parseNuvioBookingBackofficeFloatByAliases(record *core.Record, aliases []string, fallback float64) float64 {
	if record == nil {
		return fallback
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		switch raw := record.Get(fieldName).(type) {
		case float64:
			return raw
		case float32:
			return float64(raw)
		case int:
			return float64(raw)
		case int64:
			return float64(raw)
		case int32:
			return float64(raw)
		case string:
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" {
				continue
			}
			if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				return parsed
			}
		}
	}

	return fallback
}

func parseNuvioBookingBackofficeBoolByAliases(record *core.Record, aliases []string, fallback bool) bool {
	if record == nil {
		return fallback
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		if value, ok := parseBoolValue(record.Get(fieldName)); ok {
			return value
		}
	}

	return fallback
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
		BusinessNotificationTemplate: nuvioBookingBusinessNotificationTemplateConfig{
			Enabled:                   false,
			Subject:                   "",
			IntroText:                 "",
			FooterText:                "",
			IncludeAppointmentDetails: true,
		},
		VisitorEmails: nuvioBookingVisitorEmailsConfig{
			RequestTemplate: nuvioBookingVisitorNotificationTemplateConfig{
				Enabled:    false,
				Subject:    "",
				IntroText:  "",
				FooterText: "",
			},
			ConfirmationTemplate: nuvioBookingVisitorNotificationTemplateConfig{
				Enabled:    false,
				Subject:    "",
				IntroText:  "",
				FooterText: "",
			},
			RescheduleTemplate: nuvioBookingVisitorNotificationTemplateConfig{
				Enabled:    false,
				Subject:    "",
				IntroText:  "",
				FooterText: "",
			},
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

		if emailNotificationsSettings, ok := toStringAnyMap(bookingSettings["emailNotifications"]); ok {
			config.BusinessNotificationTemplate = parseNuvioBookingBusinessNotificationTemplateConfig(
				emailNotificationsSettings["businessTemplate"],
			)
		}

		if visitorEmailsSettings, ok := toStringAnyMap(bookingSettings["visitorEmails"]); ok {
			config.VisitorEmails.RequestTemplate = parseNuvioBookingVisitorNotificationTemplateConfig(
				visitorEmailsSettings["requestTemplate"],
			)
			config.VisitorEmails.ConfirmationTemplate = parseNuvioBookingVisitorNotificationTemplateConfig(
				visitorEmailsSettings["confirmationTemplate"],
			)
			config.VisitorEmails.RescheduleTemplate = parseNuvioBookingVisitorNotificationTemplateConfig(
				visitorEmailsSettings["rescheduleTemplate"],
			)
		}

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

func parseNuvioBookingBusinessNotificationTemplateConfig(
	raw any,
) nuvioBookingBusinessNotificationTemplateConfig {
	config := nuvioBookingBusinessNotificationTemplateConfig{
		Enabled:                   false,
		Subject:                   "",
		IntroText:                 "",
		FooterText:                "",
		IncludeAppointmentDetails: true,
	}

	settings, ok := toStringAnyMap(raw)
	if !ok {
		return config
	}

	if value, ok := parseBoolValue(settings["enabled"]); ok {
		config.Enabled = value
	}

	config.Subject = parseNuvioBookingTemplateStringValue(settings["subject"])
	config.IntroText = parseNuvioBookingTemplateStringValue(settings["introText"])
	config.FooterText = parseNuvioBookingTemplateStringValue(settings["footerText"])

	if value, ok := parseBoolValue(settings["includeAppointmentDetails"]); ok {
		config.IncludeAppointmentDetails = value
	}

	return config
}

func parseNuvioBookingVisitorNotificationTemplateConfig(
	raw any,
) nuvioBookingVisitorNotificationTemplateConfig {
	config := nuvioBookingVisitorNotificationTemplateConfig{
		Enabled:    false,
		Subject:    "",
		IntroText:  "",
		FooterText: "",
	}

	settings, ok := toStringAnyMap(raw)
	if !ok {
		return config
	}

	if value, ok := parseBoolValue(settings["enabled"]); ok {
		config.Enabled = value
	}

	config.Subject = parseNuvioBookingTemplateStringValue(settings["subject"])
	config.IntroText = parseNuvioBookingTemplateStringValue(settings["introText"])
	config.FooterText = parseNuvioBookingTemplateStringValue(settings["footerText"])

	return config
}

func parseNuvioBookingTemplateStringValue(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}

	return value
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

	filter := "website={:website} && dayOfWeek={:dayOfWeek} && active=true"
	sortExpr := "+startTime,+endTime,+created"
	params := dbx.Params{
		"website":   websiteID,
		"dayOfWeek": dayOfWeek,
	}

	records, err := app.FindRecordsByFilter(
		availabilityCollection,
		filter,
		sortExpr,
		5000,
		0,
		params,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
			return app.FindRecordsByFilter(
				availabilityCollection,
				filter,
				"",
				5000,
				0,
				params,
			)
		}
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

	filter := "website={:website} && date={:date} && active=true"
	sortExpr := "-updated,-created"
	params := dbx.Params{
		"website": websiteID,
		"date":    dateValue,
	}

	records, err := app.FindRecordsByFilter(
		exceptionsCollection,
		filter,
		sortExpr,
		10,
		0,
		params,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
			records, err = app.FindRecordsByFilter(
				exceptionsCollection,
				filter,
				"",
				10,
				0,
				params,
			)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
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

	sortExpr := "-created"
	records, err := app.FindRecordsByFilter(
		appointmentsCollection,
		filter,
		sortExpr,
		5000,
		0,
		params,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
			records, err = app.FindRecordsByFilter(
				appointmentsCollection,
				filter,
				"",
				5000,
				0,
				params,
			)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
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

type nuvioBookingTemplateVariablesContext struct {
	WebsiteName       string
	SubmittedAt       time.Time
	CustomerName      string
	CustomerEmail     string
	CustomerPhone     string
	ServiceName       string
	ServiceDuration   string
	AppointmentDate   string
	AppointmentTime   string
	AppointmentStatus string
	Notes             string
}

func formatNuvioBookingTemplateServiceDuration(rawMinutes int) string {
	if rawMinutes <= 0 {
		return ""
	}

	return fmt.Sprintf("%d minutes", rawMinutes)
}

func buildNuvioBookingTemplateVariables(
	context nuvioBookingTemplateVariablesContext,
) map[string]string {
	websiteName := sanitizeNuvioTemplateSingleLineValue(context.WebsiteName, "Website")
	submittedAt := context.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now().UTC()
	}

	return map[string]string{
		"websiteName":       websiteName,
		"submittedAt":       submittedAt.Format(time.RFC3339),
		"customerName":      sanitizeNuvioTemplateSingleLineValue(context.CustomerName, "Not provided"),
		"customerEmail":     sanitizeNuvioTemplateSingleLineValue(context.CustomerEmail, "Not provided"),
		"customerPhone":     sanitizeNuvioTemplateSingleLineValue(context.CustomerPhone, "Not provided"),
		"serviceName":       sanitizeNuvioTemplateSingleLineValue(context.ServiceName, "Not provided"),
		"serviceDuration":   sanitizeNuvioTemplateSingleLineValue(context.ServiceDuration, "Not provided"),
		"appointmentDate":   sanitizeNuvioTemplateSingleLineValue(context.AppointmentDate, "Not provided"),
		"appointmentTime":   sanitizeNuvioTemplateSingleLineValue(context.AppointmentTime, "Not provided"),
		"appointmentStatus": sanitizeNuvioTemplateSingleLineValue(context.AppointmentStatus, "Not provided"),
		"notes":             sanitizeNuvioTemplateMultilineValue(context.Notes, "Not provided"),
	}
}

func replaceNuvioBookingTemplateVariables(raw string, values map[string]string) string {
	replacer := strings.NewReplacer(
		"{{websiteName}}", values["websiteName"],
		"{{submittedAt}}", values["submittedAt"],
		"{{customerName}}", values["customerName"],
		"{{customerEmail}}", values["customerEmail"],
		"{{customerPhone}}", values["customerPhone"],
		"{{serviceName}}", values["serviceName"],
		"{{serviceDuration}}", values["serviceDuration"],
		"{{appointmentDate}}", values["appointmentDate"],
		"{{appointmentTime}}", values["appointmentTime"],
		"{{appointmentStatus}}", values["appointmentStatus"],
		"{{notes}}", values["notes"],
	)

	return replacer.Replace(raw)
}

func buildNuvioBookingBusinessTemplateBodyDetails(values map[string]string) string {
	lines := []string{
		fmt.Sprintf("Website: %s", values["websiteName"]),
		fmt.Sprintf("Submitted at: %s", values["submittedAt"]),
		fmt.Sprintf("Service: %s", values["serviceName"]),
		fmt.Sprintf("Service duration: %s", values["serviceDuration"]),
		fmt.Sprintf("Date: %s", values["appointmentDate"]),
		fmt.Sprintf("Time: %s", values["appointmentTime"]),
		fmt.Sprintf("Status: %s", values["appointmentStatus"]),
		fmt.Sprintf("Name: %s", values["customerName"]),
		fmt.Sprintf("Email: %s", values["customerEmail"]),
		fmt.Sprintf("Phone: %s", values["customerPhone"]),
		"",
		"Notes:",
		values["notes"],
	}

	return strings.Join(lines, "\n")
}

func buildNuvioBookingVisitorTemplateBodyDetails(values map[string]string) string {
	lines := []string{
		fmt.Sprintf("Website: %s", values["websiteName"]),
		fmt.Sprintf("Service: %s", values["serviceName"]),
		fmt.Sprintf("Service duration: %s", values["serviceDuration"]),
		fmt.Sprintf("Date: %s", values["appointmentDate"]),
		fmt.Sprintf("Time: %s", values["appointmentTime"]),
		fmt.Sprintf("Status: %s", values["appointmentStatus"]),
		fmt.Sprintf("Customer: %s", values["customerName"]),
		fmt.Sprintf("Customer email: %s", values["customerEmail"]),
		fmt.Sprintf("Customer phone: %s", values["customerPhone"]),
		"",
		"Notes:",
		values["notes"],
	}

	return strings.Join(lines, "\n")
}

func buildNuvioBookingVisitorRescheduleTemplateBodyDetails(
	values map[string]string,
	oldServiceName string,
	oldDate string,
	oldTime string,
) string {
	lines := []string{
		fmt.Sprintf("Website: %s", values["websiteName"]),
		fmt.Sprintf("Customer: %s", values["customerName"]),
		fmt.Sprintf("Customer email: %s", values["customerEmail"]),
		fmt.Sprintf("Customer phone: %s", values["customerPhone"]),
		fmt.Sprintf("Status: %s", values["appointmentStatus"]),
		"",
		"Previous appointment:",
		fmt.Sprintf("Service: %s", sanitizeNuvioTemplateSingleLineValue(oldServiceName, "Not provided")),
		fmt.Sprintf("Date: %s", sanitizeNuvioTemplateSingleLineValue(oldDate, "Not provided")),
		fmt.Sprintf("Time: %s", sanitizeNuvioTemplateSingleLineValue(oldTime, "Not provided")),
		"",
		"Updated appointment:",
		fmt.Sprintf("Service: %s", values["serviceName"]),
		fmt.Sprintf("Service duration: %s", values["serviceDuration"]),
		fmt.Sprintf("Date: %s", values["appointmentDate"]),
		fmt.Sprintf("Time: %s", values["appointmentTime"]),
		"",
		"Notes:",
		values["notes"],
	}

	return strings.Join(lines, "\n")
}

func buildNuvioBookingBusinessTemplateEmail(
	config nuvioBookingBusinessNotificationTemplateConfig,
	defaultSubject string,
	values map[string]string,
) (string, string, bool) {
	if !config.Enabled {
		return "", "", false
	}

	subject := sanitizeNuvioTemplateSubject(
		replaceNuvioBookingTemplateVariables(config.Subject, values),
	)
	if subject == "" {
		subject = sanitizeNuvioTemplateSubject(defaultSubject)
	}
	if subject == "" {
		return "", "", false
	}

	sections := []string{}
	introText := sanitizeNuvioTemplateText(
		replaceNuvioBookingTemplateVariables(config.IntroText, values),
	)
	if introText != "" {
		sections = append(sections, introText)
	}

	if config.IncludeAppointmentDetails {
		detailsText := sanitizeNuvioTemplateText(buildNuvioBookingBusinessTemplateBodyDetails(values))
		if detailsText != "" {
			sections = append(sections, detailsText)
		}
	}

	footerText := sanitizeNuvioTemplateText(
		replaceNuvioBookingTemplateVariables(config.FooterText, values),
	)
	if footerText != "" {
		sections = append(sections, footerText)
	}

	body := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if body == "" {
		return "", "", false
	}

	return subject, body, true
}

func buildNuvioBookingVisitorTemplateEmail(
	config nuvioBookingVisitorNotificationTemplateConfig,
	defaultSubject string,
	values map[string]string,
	detailsBuilder func() string,
) (string, string, bool) {
	if !config.Enabled {
		return "", "", false
	}

	subject := sanitizeNuvioTemplateSubject(
		replaceNuvioBookingTemplateVariables(config.Subject, values),
	)
	if subject == "" {
		subject = sanitizeNuvioTemplateSubject(defaultSubject)
	}
	if subject == "" {
		return "", "", false
	}

	sections := []string{}
	introText := sanitizeNuvioTemplateText(
		replaceNuvioBookingTemplateVariables(config.IntroText, values),
	)
	if introText != "" {
		sections = append(sections, introText)
	}

	detailsText := sanitizeNuvioTemplateText(detailsBuilder())
	if detailsText == "" {
		return "", "", false
	}
	sections = append(sections, detailsText)

	footerText := sanitizeNuvioTemplateText(
		replaceNuvioBookingTemplateVariables(config.FooterText, values),
	)
	if footerText != "" {
		sections = append(sections, footerText)
	}

	body := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if body == "" {
		return "", "", false
	}

	return subject, body, true
}

func buildNuvioDefaultBookingBusinessNotificationEmail(
	websiteName string,
	submittedAt time.Time,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
) (string, string) {
	visitorEmail, ok := normalizeNuvioEmail(payload.Email)
	if !ok {
		visitorEmail = strings.TrimSpace(payload.Email)
	}

	lines := []string{
		fmt.Sprintf("Website: %s", websiteName),
		fmt.Sprintf("Submitted at: %s", submittedAt.Format(time.RFC3339)),
		fmt.Sprintf("Service: %s", strings.TrimSpace(serviceName)),
		fmt.Sprintf("Date: %s", strings.TrimSpace(payload.Date)),
		fmt.Sprintf("Time: %s", strings.TrimSpace(payload.Time)),
		fmt.Sprintf("Name: %s", strings.TrimSpace(payload.Name)),
		fmt.Sprintf("Email: %s", visitorEmail),
	}

	if phone := strings.TrimSpace(payload.Phone); phone != "" {
		lines = append(lines, fmt.Sprintf("Phone: %s", phone))
	}

	if notes := strings.TrimSpace(payload.Notes); notes != "" {
		lines = append(lines, "", "Notes:", notes)
	}

	subject := fmt.Sprintf("New booking request - %s", websiteName)
	return subject, strings.Join(lines, "\n")
}

func buildNuvioDefaultBookingRequestVisitorEmail(
	websiteName string,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
) (string, string) {
	lines := []string{
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
		lines = append(lines, "", "Notes received:", notes)
	}

	return "Booking request received", strings.Join(lines, "\n")
}

func buildNuvioDefaultBookingConfirmedVisitorEmail(
	websiteName string,
	serviceName string,
	payload nuvioBookingCreateAppointmentPayload,
) (string, string) {
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

	return "Appointment confirmed", strings.Join(lines, "\n")
}

func buildNuvioDefaultBookingRescheduleVisitorEmail(
	websiteName string,
	visitorName string,
	oldServiceName string,
	oldDate string,
	oldTime string,
	newServiceName string,
	newDate string,
	newTime string,
) (string, string) {
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

	return "Appointment rescheduled", strings.Join(lines, "\n")
}

func sendNuvioBookingConfirmedVisitorEmail(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteBookingConfig,
	serviceName string,
	serviceDuration string,
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
	submittedAt := time.Now().UTC()
	subject, textBody := buildNuvioDefaultBookingConfirmedVisitorEmail(
		websiteName,
		serviceName,
		payload,
	)

	templateValues := buildNuvioBookingTemplateVariables(nuvioBookingTemplateVariablesContext{
		WebsiteName:       websiteName,
		SubmittedAt:       submittedAt,
		CustomerName:      payload.Name,
		CustomerEmail:     visitorEmail,
		CustomerPhone:     payload.Phone,
		ServiceName:       serviceName,
		ServiceDuration:   serviceDuration,
		AppointmentDate:   payload.Date,
		AppointmentTime:   payload.Time,
		AppointmentStatus: "confirmed",
		Notes:             payload.Notes,
	})
	if templateSubject, templateBody, ok := buildNuvioBookingVisitorTemplateEmail(
		config.VisitorEmails.ConfirmationTemplate,
		subject,
		templateValues,
		func() string {
			return buildNuvioBookingVisitorTemplateBodyDetails(templateValues)
		},
	); ok {
		subject = templateSubject
		textBody = templateBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:          []string{visitorEmail},
		Subject:     subject,
		Text:        textBody,
		Attachments: attachments,
	})
}

func sendNuvioBookingBusinessNotificationEmail(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteBookingConfig,
	serviceName string,
	serviceDuration string,
	appointmentStatus string,
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
	submittedAt := time.Now().UTC()
	subject, textBody := buildNuvioDefaultBookingBusinessNotificationEmail(
		websiteName,
		submittedAt,
		serviceName,
		payload,
	)

	templateValues := buildNuvioBookingTemplateVariables(nuvioBookingTemplateVariablesContext{
		WebsiteName:       websiteName,
		SubmittedAt:       submittedAt,
		CustomerName:      payload.Name,
		CustomerEmail:     visitorEmail,
		CustomerPhone:     payload.Phone,
		ServiceName:       serviceName,
		ServiceDuration:   serviceDuration,
		AppointmentDate:   payload.Date,
		AppointmentTime:   payload.Time,
		AppointmentStatus: appointmentStatus,
		Notes:             payload.Notes,
	})
	if templateSubject, templateBody, ok := buildNuvioBookingBusinessTemplateEmail(
		config.BusinessNotificationTemplate,
		subject,
		templateValues,
	); ok {
		subject = templateSubject
		textBody = templateBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      config.EmailNotifications.To,
		Cc:      config.EmailNotifications.Cc,
		ReplyTo: []string{visitorEmail},
		Subject: subject,
		Text:    textBody,
	})
}

func sendNuvioBookingEmails(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteBookingConfig,
	serviceName string,
	serviceDuration string,
	appointmentStatus string,
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
	submittedAt := time.Now().UTC()
	visitorSubject, visitorTextBody := buildNuvioDefaultBookingRequestVisitorEmail(
		websiteName,
		serviceName,
		payload,
	)
	visitorTemplateValues := buildNuvioBookingTemplateVariables(nuvioBookingTemplateVariablesContext{
		WebsiteName:       websiteName,
		SubmittedAt:       submittedAt,
		CustomerName:      payload.Name,
		CustomerEmail:     visitorEmail,
		CustomerPhone:     payload.Phone,
		ServiceName:       serviceName,
		ServiceDuration:   serviceDuration,
		AppointmentDate:   payload.Date,
		AppointmentTime:   payload.Time,
		AppointmentStatus: appointmentStatus,
		Notes:             payload.Notes,
	})
	if templateSubject, templateBody, ok := buildNuvioBookingVisitorTemplateEmail(
		config.VisitorEmails.RequestTemplate,
		visitorSubject,
		visitorTemplateValues,
		func() string {
			return buildNuvioBookingVisitorTemplateBodyDetails(visitorTemplateValues)
		},
	); ok {
		visitorSubject = templateSubject
		visitorTextBody = templateBody
	}

	sendErrors := []string{}

	if err := sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      []string{visitorEmail},
		Subject: visitorSubject,
		Text:    visitorTextBody,
	}); err != nil {
		sendErrors = append(sendErrors, fmt.Sprintf("visitor confirmation failed: %s", err.Error()))
	}

	if config.EmailNotifications.Enabled && len(config.EmailNotifications.To) > 0 {
		businessSubject, businessTextBody := buildNuvioDefaultBookingBusinessNotificationEmail(
			websiteName,
			submittedAt,
			serviceName,
			payload,
		)
		businessTemplateValues := buildNuvioBookingTemplateVariables(nuvioBookingTemplateVariablesContext{
			WebsiteName:       websiteName,
			SubmittedAt:       submittedAt,
			CustomerName:      payload.Name,
			CustomerEmail:     visitorEmail,
			CustomerPhone:     payload.Phone,
			ServiceName:       serviceName,
			ServiceDuration:   serviceDuration,
			AppointmentDate:   payload.Date,
			AppointmentTime:   payload.Time,
			AppointmentStatus: appointmentStatus,
			Notes:             payload.Notes,
		})
		if templateSubject, templateBody, ok := buildNuvioBookingBusinessTemplateEmail(
			config.BusinessNotificationTemplate,
			businessSubject,
			businessTemplateValues,
		); ok {
			businessSubject = templateSubject
			businessTextBody = templateBody
		}

		if err := sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
			To:      config.EmailNotifications.To,
			Cc:      config.EmailNotifications.Cc,
			ReplyTo: []string{visitorEmail},
			Subject: businessSubject,
			Text:    businessTextBody,
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
	config nuvioWebsiteBookingConfig,
	visitorName string,
	visitorEmail string,
	visitorPhone string,
	oldServiceName string,
	oldDate string,
	oldTime string,
	newServiceName string,
	newServiceDuration string,
	newDate string,
	newTime string,
	appointmentStatus string,
	notes string,
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
	submittedAt := time.Now().UTC()
	subject, textBody := buildNuvioDefaultBookingRescheduleVisitorEmail(
		websiteName,
		visitorName,
		oldServiceName,
		oldDate,
		oldTime,
		newServiceName,
		newDate,
		newTime,
	)
	templateValues := buildNuvioBookingTemplateVariables(nuvioBookingTemplateVariablesContext{
		WebsiteName:       websiteName,
		SubmittedAt:       submittedAt,
		CustomerName:      visitorName,
		CustomerEmail:     normalizedEmail,
		CustomerPhone:     visitorPhone,
		ServiceName:       newServiceName,
		ServiceDuration:   newServiceDuration,
		AppointmentDate:   newDate,
		AppointmentTime:   newTime,
		AppointmentStatus: appointmentStatus,
		Notes:             notes,
	})
	if templateSubject, templateBody, ok := buildNuvioBookingVisitorTemplateEmail(
		config.VisitorEmails.RescheduleTemplate,
		subject,
		templateValues,
		func() string {
			return buildNuvioBookingVisitorRescheduleTemplateBodyDetails(
				templateValues,
				oldServiceName,
				oldDate,
				oldTime,
			)
		},
	); ok {
		subject = templateSubject
		textBody = templateBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:          []string{normalizedEmail},
		Subject:     subject,
		Text:        textBody,
		Attachments: attachments,
	})
}

// NUVIO CUSTOM END: Booking MVP Phase 3 public booking endpoints.
