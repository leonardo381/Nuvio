package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"strings"
	"unicode"
)

const nuvioTrustedMarkupMaxLength = 20000

type trustedMarkupProfile string

const (
	trustedSvgIllustration  trustedMarkupProfile = "trustedSvgIllustration"
	trustedHtmlIllustration trustedMarkupProfile = "trustedHtmlIllustration"
)

type trustedMarkupValidationReport struct {
	Profile           trustedMarkupProfile
	Valid             bool
	RemovedTags       []string
	RemovedAttributes []string
	Errors            []string
}

type trustedMarkupPolicy struct {
	AllowedTags       map[string]string
	AllowedAttributes map[string]string
	AllowDataAttrs    bool
	RequiredRootTag   string
	SingleRoot        bool
}

type sanitizedTrustedMarkupAttribute struct {
	Name  string
	Value string
}

var errNuvioTrustedMarkupUnsafe = errors.New("trusted markup is not safe")

var nuvioTrustedMarkupPolicies = map[trustedMarkupProfile]trustedMarkupPolicy{
	trustedSvgIllustration: {
		AllowedTags: trustedMarkupCanonicalMap(
			"svg", "g", "path", "rect", "circle", "ellipse", "line", "polyline", "polygon",
			"defs", "linearGradient", "radialGradient", "stop", "clipPath", "mask", "title", "desc",
		),
		AllowedAttributes: trustedMarkupCanonicalMap(
			"viewBox", "width", "height", "xmlns", "d", "x", "y", "x1", "y1", "x2", "y2",
			"cx", "cy", "r", "rx", "ry", "points", "fill", "stroke", "stroke-width", "stroke-linecap",
			"stroke-linejoin", "opacity", "fill-opacity", "stroke-opacity", "transform", "id", "class", "role",
			"focusable", "preserveAspectRatio", "gradientUnits", "offset", "stop-color", "stop-opacity", "clip-path", "mask",
		),
		RequiredRootTag: "svg",
		SingleRoot:      true,
	},
	trustedHtmlIllustration: {
		AllowedTags:       trustedMarkupCanonicalMap("div", "span", "ul", "ol", "li", "figure", "figcaption", "small", "strong", "em"),
		AllowedAttributes: trustedMarkupCanonicalMap("class", "id", "role", "title"),
		AllowDataAttrs:    true,
	},
}

func sanitizeTrustedMarkup(input string, profile trustedMarkupProfile) (string, trustedMarkupValidationReport, error) {
	report := trustedMarkupValidationReport{Profile: profile}
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		report.Valid = true
		return "", report, nil
	}
	if len(trimmed) > nuvioTrustedMarkupMaxLength {
		return "", report, addTrustedMarkupError(&report, "markup exceeds trusted markup size limit")
	}

	policy, ok := nuvioTrustedMarkupPolicies[profile]
	if !ok {
		return "", report, addTrustedMarkupError(&report, "unknown trusted markup profile")
	}

	decoder := xml.NewDecoder(strings.NewReader(trimmed))
	decoder.Strict = true

	var builder strings.Builder
	var stack []string
	topLevelElements := 0
	sawElement := false

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", report, addTrustedMarkupError(&report, "markup is malformed")
		}

		switch typed := token.(type) {
		case xml.StartElement:
			tagName := normalizeTrustedMarkupName(typed.Name)
			canonicalTag, ok := policy.AllowedTags[tagName]
			if !ok {
				report.RemovedTags = append(report.RemovedTags, tagName)
				return "", report, addTrustedMarkupError(&report, fmt.Sprintf("tag <%s> is not allowed", tagName))
			}
			if len(stack) == 0 {
				topLevelElements++
				if policy.SingleRoot && topLevelElements > 1 {
					return "", report, addTrustedMarkupError(&report, "profile allows only one root element")
				}
				if policy.RequiredRootTag != "" && tagName != policy.RequiredRootTag {
					return "", report, addTrustedMarkupError(&report, "profile requires <"+policy.RequiredRootTag+"> as the root element")
				}
			}

			attrs, err := sanitizeTrustedMarkupAttributes(tagName, typed.Attr, policy, &report)
			if err != nil {
				return "", report, err
			}

			builder.WriteString("<")
			builder.WriteString(canonicalTag)
			for _, attr := range attrs {
				builder.WriteString(" ")
				builder.WriteString(attr.Name)
				builder.WriteString("=\"")
				builder.WriteString(html.EscapeString(attr.Value))
				builder.WriteString("\"")
			}
			builder.WriteString(">")
			stack = append(stack, canonicalTag)
			sawElement = true
		case xml.EndElement:
			tagName := normalizeTrustedMarkupName(typed.Name)
			canonicalTag, ok := policy.AllowedTags[tagName]
			if !ok {
				report.RemovedTags = append(report.RemovedTags, tagName)
				return "", report, addTrustedMarkupError(&report, fmt.Sprintf("closing tag </%s> is not allowed", tagName))
			}
			if len(stack) == 0 || stack[len(stack)-1] != canonicalTag {
				return "", report, addTrustedMarkupError(&report, "markup has mismatched closing tags")
			}
			stack = stack[:len(stack)-1]
			builder.WriteString("</")
			builder.WriteString(canonicalTag)
			builder.WriteString(">")
		case xml.CharData:
			text := string(typed)
			if strings.TrimSpace(text) == "" {
				if len(stack) > 0 {
					builder.WriteString(html.EscapeString(text))
				}
				continue
			}
			if len(stack) == 0 {
				return "", report, addTrustedMarkupError(&report, "text outside trusted markup elements is not allowed")
			}
			if profile == trustedSvgIllustration {
				parentTag := strings.ToLower(stack[len(stack)-1])
				if parentTag != "title" && parentTag != "desc" {
					return "", report, addTrustedMarkupError(&report, "visible SVG text is not allowed in trusted illustration markup")
				}
			}
			builder.WriteString(html.EscapeString(text))
		case xml.Comment:
			return "", report, addTrustedMarkupError(&report, "comments are not allowed in trusted markup")
		case xml.Directive, xml.ProcInst:
			return "", report, addTrustedMarkupError(&report, "directives are not allowed in trusted markup")
		}
	}

	if len(stack) > 0 {
		return "", report, addTrustedMarkupError(&report, "markup has unclosed tags")
	}
	if !sawElement {
		return "", report, addTrustedMarkupError(&report, "trusted markup must contain at least one element")
	}

	report.Valid = true
	return strings.TrimSpace(builder.String()), report, nil
}

func validateTrustedMarkup(input string, profile trustedMarkupProfile) error {
	_, _, err := sanitizeTrustedMarkup(input, profile)
	return err
}
func sanitizeTrustedMarkupAttributes(tagName string, attrs []xml.Attr, policy trustedMarkupPolicy, report *trustedMarkupValidationReport) ([]sanitizedTrustedMarkupAttribute, error) {
	result := make([]sanitizedTrustedMarkupAttribute, 0, len(attrs))
	seen := map[string]struct{}{}

	for _, attr := range attrs {
		attrName := normalizeTrustedMarkupAttributeName(attr.Name)
		if attrName == "" {
			return nil, addTrustedMarkupError(report, "empty attribute names are not allowed")
		}
		if isForbiddenTrustedMarkupAttribute(attrName) {
			report.RemovedAttributes = append(report.RemovedAttributes, attrName)
			return nil, addTrustedMarkupError(report, fmt.Sprintf("attribute %q is not allowed", attrName))
		}

		canonicalName, ok := policy.AllowedAttributes[attrName]
		if !ok {
			switch {
			case strings.HasPrefix(attrName, "aria-") && isValidTrustedMarkupDashName(attrName):
				canonicalName = attrName
			case policy.AllowDataAttrs && strings.HasPrefix(attrName, "data-") && isValidTrustedMarkupDashName(attrName):
				canonicalName = attrName
			default:
				report.RemovedAttributes = append(report.RemovedAttributes, attrName)
				return nil, addTrustedMarkupError(report, fmt.Sprintf("attribute %q is not allowed", attrName))
			}
		}

		if _, exists := seen[canonicalName]; exists {
			return nil, addTrustedMarkupError(report, fmt.Sprintf("duplicate attribute %q is not allowed", canonicalName))
		}
		seen[canonicalName] = struct{}{}

		if err := validateTrustedMarkupAttributeValue(tagName, canonicalName, attr.Value, report); err != nil {
			return nil, err
		}

		result = append(result, sanitizedTrustedMarkupAttribute{Name: canonicalName, Value: strings.TrimSpace(attr.Value)})
	}

	return result, nil
}

func validateTrustedMarkupAttributeValue(tagName string, attrName string, value string, report *trustedMarkupValidationReport) error {
	trimmed := strings.TrimSpace(value)
	if attrName == "xmlns" {
		if tagName == "svg" && trimmed == "http://www.w3.org/2000/svg" {
			return nil
		}
		return addTrustedMarkupError(report, "only the default SVG namespace is allowed")
	}
	if attrName == "id" && trimmed != "" && !isValidTrustedMarkupIdentifier(trimmed) {
		return addTrustedMarkupError(report, "id values must be internal identifiers")
	}
	if strings.Contains(strings.ToLower(trimmed), "url(") && !hasOnlyTrustedMarkupInternalURLReferences(trimmed) {
		return addTrustedMarkupError(report, "url() references must be internal fragment references")
	}
	if containsTrustedMarkupDangerousProtocol(trimmed) {
		return addTrustedMarkupError(report, "external or dangerous URL protocol is not allowed")
	}
	if (attrName == "clip-path" || attrName == "mask") && !isTrustedMarkupInternalURLReference(trimmed) {
		return addTrustedMarkupError(report, attrName+" must use an internal url(#id) reference")
	}
	return nil
}

func addTrustedMarkupError(report *trustedMarkupValidationReport, message string) error {
	if report != nil {
		report.Valid = false
		report.Errors = append(report.Errors, message)
	}
	return fmt.Errorf("%w: %s", errNuvioTrustedMarkupUnsafe, message)
}

func trustedMarkupCanonicalMap(values ...string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		result[strings.ToLower(strings.TrimSpace(value))] = strings.TrimSpace(value)
	}
	return result
}

func normalizeTrustedMarkupName(name xml.Name) string {
	return strings.ToLower(strings.TrimSpace(name.Local))
}

func normalizeTrustedMarkupAttributeName(name xml.Name) string {
	local := strings.ToLower(strings.TrimSpace(name.Local))
	space := strings.ToLower(strings.TrimSpace(name.Space))
	if space == "" {
		return local
	}
	if space == "xmlns" {
		return "xmlns:" + local
	}
	return space + ":" + local
}

func isForbiddenTrustedMarkupAttribute(attrName string) bool {
	if strings.HasPrefix(attrName, "on") {
		return true
	}
	for _, forbidden := range []string{
		"href",
		"xlink:href",
		"http://www.w3.org/1999/xlink:href",
		"style",
		"src",
		"action",
		"formaction",
		"target",
		"download",
		"autofocus",
		"contenteditable",
	} {
		if attrName == forbidden {
			return true
		}
	}
	return false
}

func containsTrustedMarkupDangerousProtocol(value string) bool {
	compact := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, value)

	for _, protocol := range []string{"javascript:", "data:", "vbscript:", "http:", "https:", "file:", "ftp:", "blob:", "mailto:", "tel:"} {
		if strings.Contains(compact, protocol) {
			return true
		}
	}
	return false
}

func hasOnlyTrustedMarkupInternalURLReferences(value string) bool {
	remaining := value
	for {
		index := strings.Index(strings.ToLower(remaining), "url(")
		if index < 0 {
			return true
		}
		start := index + len("url(")
		end := strings.Index(remaining[start:], ")")
		if end < 0 {
			return false
		}
		reference := remaining[start : start+end]
		if !isTrustedMarkupInternalURLReference(reference) {
			return false
		}
		remaining = remaining[start+end+1:]
	}
}

func isTrustedMarkupInternalURLReference(value string) bool {
	trimmed := strings.Trim(strings.TrimSpace(value), `"'`)
	if strings.HasPrefix(strings.ToLower(trimmed), "url(") && strings.HasSuffix(trimmed, ")") {
		trimmed = strings.Trim(strings.TrimSpace(trimmed[4:len(trimmed)-1]), `"'`)
	}
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}
	return isValidTrustedMarkupIdentifier(strings.TrimPrefix(trimmed, "#"))
}

func isValidTrustedMarkupIdentifier(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for index, r := range trimmed {
		valid := r == '_' || r == '-' || r == ':' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r)
		if !valid {
			return false
		}
		if index == 0 && unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isValidTrustedMarkupDashName(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}
