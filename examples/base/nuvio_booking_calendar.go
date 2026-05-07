package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	nuvioBookingCalendarProdID = "-//Nuvio//Booking//EN"
	nuvioBookingCalendarMethod = "REQUEST"
)

type nuvioBookingCalendarInvitePayload struct {
	WebsiteName     string
	ServiceName     string
	CustomerName    string
	CustomerEmail   string
	CustomerPhone   string
	Date            string
	Time            string
	DurationMinutes int
	Notes           string
	Location        string
	AppointmentID   string
}

func buildNuvioBookingICSContent(payload nuvioBookingCalendarInvitePayload) ([]byte, error) {
	return buildNuvioBookingICSContentAt(payload, time.Now().UTC())
}

func buildNuvioBookingICSContentAt(
	payload nuvioBookingCalendarInvitePayload,
	now time.Time,
) ([]byte, error) {
	location := getNuvioBookingLocation()
	bookingDate, err := parseNuvioBookingDateInLocation(payload.Date, location)
	if err != nil {
		return nil, fmt.Errorf("invalid booking date")
	}

	slotMinutes, err := parseNuvioBookingHHMM(payload.Time)
	if err != nil {
		return nil, fmt.Errorf("invalid booking time")
	}

	durationMinutes := payload.DurationMinutes
	if durationMinutes <= 0 {
		return nil, fmt.Errorf("invalid booking duration")
	}

	startAt := time.Date(
		bookingDate.Year(),
		bookingDate.Month(),
		bookingDate.Day(),
		slotMinutes/60,
		slotMinutes%60,
		0,
		0,
		location,
	)
	endAt := startAt.Add(time.Duration(durationMinutes) * time.Minute)
	if !endAt.After(startAt) {
		return nil, fmt.Errorf("invalid booking end time")
	}

	stamp := now
	if stamp.IsZero() {
		stamp = time.Now().UTC()
	}

	uid := buildNuvioBookingCalendarUID(payload)
	summary := buildNuvioBookingCalendarSummary(payload)
	description := buildNuvioBookingCalendarDescription(payload)

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		fmt.Sprintf("PRODID:%s", nuvioBookingCalendarProdID),
		fmt.Sprintf("METHOD:%s", nuvioBookingCalendarMethod),
		"BEGIN:VEVENT",
		fmt.Sprintf("UID:%s", escapeNuvioICSText(uid)),
		fmt.Sprintf("DTSTAMP:%s", stamp.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("DTSTART:%s", startAt.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("DTEND:%s", endAt.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("SUMMARY:%s", escapeNuvioICSText(summary)),
		fmt.Sprintf("DESCRIPTION:%s", escapeNuvioICSText(description)),
	}

	if locationValue := strings.TrimSpace(payload.Location); locationValue != "" {
		lines = append(lines, fmt.Sprintf("LOCATION:%s", escapeNuvioICSText(locationValue)))
	}

	lines = append(lines, "END:VEVENT", "END:VCALENDAR")
	return []byte(strings.Join(lines, "\r\n") + "\r\n"), nil
}

func buildNuvioBookingICSFilename(dateValue string, timeValue string) string {
	datePart := sanitizeNuvioCalendarFilenamePart(dateValue, "date")
	timePart := sanitizeNuvioCalendarFilenamePart(strings.ReplaceAll(timeValue, ":", "-"), "time")
	return fmt.Sprintf("appointment-%s-%s.ics", datePart, timePart)
}

func buildNuvioBookingCalendarUID(payload nuvioBookingCalendarInvitePayload) string {
	if appointmentID := sanitizeNuvioCalendarUIDPart(payload.AppointmentID); appointmentID != "" {
		return fmt.Sprintf("nuvio-booking-%s@nuvio.local", appointmentID)
	}

	fallback := strings.Join(
		[]string{
			strings.TrimSpace(payload.Date),
			strings.TrimSpace(payload.Time),
			strings.TrimSpace(payload.ServiceName),
			strings.TrimSpace(payload.CustomerName),
			strings.TrimSpace(payload.CustomerEmail),
			strings.TrimSpace(payload.CustomerPhone),
			strings.TrimSpace(payload.WebsiteName),
		},
		"|",
	)
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(fallback))))
	return fmt.Sprintf("nuvio-booking-%s@nuvio.local", hex.EncodeToString(sum[:12]))
}

func buildNuvioBookingCalendarSummary(payload nuvioBookingCalendarInvitePayload) string {
	serviceName := strings.TrimSpace(payload.ServiceName)
	customerName := strings.TrimSpace(payload.CustomerName)

	if serviceName != "" && customerName != "" {
		return fmt.Sprintf("%s - %s", serviceName, customerName)
	}
	if serviceName != "" {
		return serviceName
	}
	if customerName != "" {
		return fmt.Sprintf("Appointment - %s", customerName)
	}
	return "Appointment"
}

func buildNuvioBookingCalendarDescription(payload nuvioBookingCalendarInvitePayload) string {
	lines := []string{}

	if websiteName := strings.TrimSpace(payload.WebsiteName); websiteName != "" {
		lines = append(lines, fmt.Sprintf("Website: %s", websiteName))
	}
	if customerName := strings.TrimSpace(payload.CustomerName); customerName != "" {
		lines = append(lines, fmt.Sprintf("Customer: %s", customerName))
	}
	if customerEmail := strings.TrimSpace(payload.CustomerEmail); customerEmail != "" {
		lines = append(lines, fmt.Sprintf("Email: %s", customerEmail))
	}
	if customerPhone := strings.TrimSpace(payload.CustomerPhone); customerPhone != "" {
		lines = append(lines, fmt.Sprintf("Phone: %s", customerPhone))
	}
	if notes := strings.TrimSpace(payload.Notes); notes != "" {
		lines = append(lines, fmt.Sprintf("Notes: %s", notes))
	}

	lines = append(lines, "Source: Nuvio Booking")
	return strings.Join(lines, "\n")
}

func escapeNuvioICSText(value string) string {
	normalized := strings.TrimSpace(value)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.ReplaceAll(normalized, "\\", "\\\\")
	normalized = strings.ReplaceAll(normalized, ";", "\\;")
	normalized = strings.ReplaceAll(normalized, ",", "\\,")
	normalized = strings.ReplaceAll(normalized, "\n", "\\n")
	return normalized
}

func sanitizeNuvioCalendarUIDPart(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))
	for _, char := range trimmed {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' {
			builder.WriteRune(char)
		} else {
			builder.WriteByte('-')
		}
	}

	sanitized := strings.Trim(builder.String(), "-")
	return strings.TrimSpace(sanitized)
}

func sanitizeNuvioCalendarFilenamePart(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))
	lastDash := false

	for _, char := range trimmed {
		isAllowed := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.'

		if isAllowed {
			builder.WriteRune(char)
			lastDash = false
			continue
		}

		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}

	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		return fallback
	}

	return sanitized
}
