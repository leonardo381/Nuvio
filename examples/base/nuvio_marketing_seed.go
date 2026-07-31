package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

const (
	nuvioMarketingSeedDefaultFixturePath = "fixtures/nuvio_marketing_cms_fixture.json"
	nuvioMarketingSeedExpectedOwnerKey   = "nuvio-marketing-cms-fixture"
)

//go:embed fixtures/nuvio_marketing_cms_fixture.json
var nuvioMarketingSeedFixtures embed.FS

type nuvioMarketingSeedFixture struct {
	Version    int                           `json:"version"`
	OwnerKey   string                        `json:"ownerKey"`
	Source     map[string]any                `json:"source"`
	Website    nuvioMarketingSeedWebsite     `json:"website"`
	Components []nuvioMarketingSeedComponent `json:"components"`
	Pages      []nuvioMarketingSeedPage      `json:"pages"`
}

type nuvioMarketingSeedWebsite struct {
	Slug     string         `json:"slug"`
	Name     string         `json:"name"`
	Title    string         `json:"title"`
	Domain   string         `json:"domain"`
	Settings map[string]any `json:"settings"`
}

type nuvioMarketingSeedComponent struct {
	Key    string         `json:"key"`
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
}

type nuvioMarketingSeedPage struct {
	Slug                  string                    `json:"slug"`
	Title                 string                    `json:"title"`
	SeoTitle              string                    `json:"seoTitle"`
	SeoDescription        string                    `json:"seoDescription"`
	Published             bool                      `json:"published"`
	SeoNoindex            bool                      `json:"seoNoindex"`
	SeoExcludeFromSitemap bool                      `json:"seoExcludeFromSitemap"`
	Blocks                []nuvioMarketingSeedBlock `json:"blocks"`
}

type nuvioMarketingSeedBlock struct {
	Slot         string         `json:"slot"`
	ComponentKey string         `json:"componentKey"`
	Title        string         `json:"title"`
	Order        int            `json:"order"`
	Props        map[string]any `json:"props"`
}

type nuvioMarketingSeedOptions struct {
	FixturePath string
	Apply       bool
}

type nuvioMarketingSeedStats struct {
	Created  map[string]int
	Updated  map[string]int
	Skipped  map[string]int
	Warnings []string
}

func newNuvioMarketingSeedCommand(app core.App) *cobra.Command {
	opts := nuvioMarketingSeedOptions{}

	command := &cobra.Command{
		Use:          "seed-nuvio-marketing",
		Short:        "Creates or updates controlled Nuvio marketing CMS records",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if !app.IsBootstrapped() {
				if err := app.Bootstrap(); err != nil {
					return fmt.Errorf("failed to bootstrap PocketBase before seeding Nuvio marketing records: %w", err)
				}
			}
			if err := app.RunAppMigrations(); err != nil {
				return fmt.Errorf("failed to apply app migrations before seeding Nuvio marketing records: %w", err)
			}

			fixture, err := loadNuvioMarketingSeedFixture(opts.FixturePath)
			if err != nil {
				return err
			}
			if err := validateNuvioMarketingSeedFixture(fixture); err != nil {
				return err
			}

			stats, err := applyNuvioMarketingSeedFixture(app, fixture, opts.Apply)
			if err != nil {
				return err
			}

			printNuvioMarketingSeedSummary(fixture, stats, opts.Apply)
			return nil
		},
	}

	command.Flags().StringVar(&opts.FixturePath, "fixture", "", "Optional fixture JSON path. Defaults to the embedded Nuvio marketing fixture.")
	command.Flags().BoolVar(&opts.Apply, "apply", false, "Write changes. Without this flag the command runs in dry-run mode.")

	return command
}

func loadNuvioMarketingSeedFixture(path string) (nuvioMarketingSeedFixture, error) {
	var raw []byte
	var err error
	if strings.TrimSpace(path) != "" {
		raw, err = os.ReadFile(strings.TrimSpace(path))
	} else {
		raw, err = nuvioMarketingSeedFixtures.ReadFile(nuvioMarketingSeedDefaultFixturePath)
	}
	if err != nil {
		return nuvioMarketingSeedFixture{}, fmt.Errorf("failed to read Nuvio marketing fixture: %w", err)
	}

	var fixture nuvioMarketingSeedFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		return nuvioMarketingSeedFixture{}, fmt.Errorf("failed to parse Nuvio marketing fixture: %w", err)
	}
	return fixture, nil
}

func validateNuvioMarketingSeedFixture(fixture nuvioMarketingSeedFixture) error {
	if fixture.Version != 1 {
		return fmt.Errorf("unsupported Nuvio marketing fixture version: %d", fixture.Version)
	}
	if strings.TrimSpace(fixture.OwnerKey) != nuvioMarketingSeedExpectedOwnerKey {
		return fmt.Errorf("unexpected Nuvio marketing fixture owner key: %q", fixture.OwnerKey)
	}
	if strings.TrimSpace(fixture.Website.Slug) != "nuvio" {
		return fmt.Errorf("fixture website slug must be nuvio, got %q", fixture.Website.Slug)
	}

	expectedPages := map[string]int{"home": 9, "services": 6, "pricing": 4, "contact": 1}
	seenPages := map[string]bool{}
	componentKeys := map[string]bool{}
	for _, component := range fixture.Components {
		key := strings.TrimSpace(component.Key)
		if key == "" {
			return errors.New("fixture contains component without key")
		}
		componentKeys[key] = true
		if !strings.HasPrefix(key, "nuvio-") {
			return fmt.Errorf("fixture component key must be Nuvio-scoped: %s", key)
		}
		if !isNuvioMarketingSeedOwnedMap(component.Schema, fixture.OwnerKey) {
			return fmt.Errorf("fixture component schema is missing ownership marker: %s", key)
		}
	}

	for _, page := range fixture.Pages {
		pageSlug := strings.TrimSpace(page.Slug)
		seenPages[pageSlug] = true
		expectedBlockCount, ok := expectedPages[pageSlug]
		if !ok {
			return fmt.Errorf("fixture contains unexpected page: %s", pageSlug)
		}
		if len(page.Blocks) != expectedBlockCount {
			return fmt.Errorf("fixture page %s has %d blocks, expected %d", pageSlug, len(page.Blocks), expectedBlockCount)
		}
		seenSlots := map[string]bool{}
		for _, block := range page.Blocks {
			if strings.TrimSpace(block.Slot) == "" {
				return fmt.Errorf("fixture page %s contains block without slot", pageSlug)
			}
			if seenSlots[block.Slot] {
				return fmt.Errorf("fixture page %s contains duplicate block slot: %s", pageSlug, block.Slot)
			}
			seenSlots[block.Slot] = true
			if !componentKeys[strings.TrimSpace(block.ComponentKey)] {
				return fmt.Errorf("fixture block references missing component key: %s", block.ComponentKey)
			}
			if containsNuvioMarketingSeedRawHTML(block.Props) {
				return fmt.Errorf("fixture block %s contains raw HTML-like content", block.Slot)
			}
		}
	}
	for pageSlug := range expectedPages {
		if !seenPages[pageSlug] {
			return fmt.Errorf("fixture missing page: %s", pageSlug)
		}
	}

	if err := validateNuvioMarketingSeedPricing(fixture); err != nil {
		return err
	}
	if err := validateNuvioMarketingSeedContact(fixture); err != nil {
		return err
	}

	return nil
}

func applyNuvioMarketingSeedFixture(app core.App, fixture nuvioMarketingSeedFixture, apply bool) (nuvioMarketingSeedStats, error) {
	stats := nuvioMarketingSeedStats{
		Created: map[string]int{},
		Updated: map[string]int{},
		Skipped: map[string]int{},
	}

	websitesCollection, err := app.FindCachedCollectionByNameOrId(nuvioWebsitesCollectionID)
	if err != nil {
		return stats, fmt.Errorf("failed to resolve Websites collection: %w", err)
	}
	pagesCollection, err := app.FindCachedCollectionByNameOrId(nuvioPagesCollectionID)
	if err != nil {
		return stats, fmt.Errorf("failed to resolve Pages collection: %w", err)
	}
	blocksCollection, err := app.FindCachedCollectionByNameOrId(nuvioBlocksCollectionID)
	if err != nil {
		return stats, fmt.Errorf("failed to resolve Blocks collection: %w", err)
	}
	componentsCollection, err := app.FindCachedCollectionByNameOrId(nuvioComponentsCollectionID)
	if err != nil {
		return stats, fmt.Errorf("failed to resolve Components collection: %w", err)
	}

	websiteRecord, err := upsertNuvioMarketingSeedWebsite(app, websitesCollection, fixture, apply, &stats)
	if err != nil {
		return stats, err
	}

	componentRecords := map[string]*core.Record{}
	for _, component := range fixture.Components {
		record, err := upsertNuvioMarketingSeedComponent(app, componentsCollection, fixture, component, apply, &stats)
		if err != nil {
			return stats, err
		}
		componentRecords[strings.TrimSpace(component.Key)] = record
	}

	websiteID := strings.TrimSpace(websiteRecord.Id)
	if websiteID == "" && !apply {
		websiteID = "dry-run-nuvio-website"
	}

	for _, page := range fixture.Pages {
		pageRecord, err := upsertNuvioMarketingSeedPage(app, pagesCollection, fixture, page, websiteID, apply, &stats)
		if err != nil {
			return stats, err
		}

		pageID := strings.TrimSpace(pageRecord.Id)
		if pageID == "" && !apply {
			pageID = "dry-run-page-" + page.Slug
		}

		for _, block := range page.Blocks {
			componentRecord := componentRecords[strings.TrimSpace(block.ComponentKey)]
			if err := upsertNuvioMarketingSeedBlock(app, blocksCollection, fixture, block, pageID, componentRecord, apply, &stats); err != nil {
				return stats, err
			}
		}
	}

	if !hasNuvioMarketingSeedField(blocksCollection, "displayOrder") && !hasNuvioMarketingSeedField(blocksCollection, "order") {
		stats.Warnings = append(stats.Warnings, "Blocks collection has no displayOrder/order field; public order relies on deterministic creation order.")
	}

	return stats, nil
}

func upsertNuvioMarketingSeedWebsite(app core.App, collection *core.Collection, fixture nuvioMarketingSeedFixture, apply bool, stats *nuvioMarketingSeedStats) (*core.Record, error) {
	slug := strings.TrimSpace(fixture.Website.Slug)
	record, err := app.FindFirstRecordByFilter(collection, "slug={:slug}", dbx.Params{"slug": slug})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to lookup website %q: %w", slug, err)
	}

	created := false
	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		created = true
	} else if !isNuvioMarketingSeedOwnedMap(nuvioMarketingSeedMapValue(record.Get("settings")), fixture.OwnerKey) {
		return nil, fmt.Errorf("website %q already exists and is not marked as owned by %s; refusing to update", slug, fixture.OwnerKey)
	}

	if created {
		countNuvioMarketingSeedStat(stats.Created, "Websites")
	} else {
		countNuvioMarketingSeedStat(stats.Updated, "Websites")
	}
	if !apply {
		return record, nil
	}

	setNuvioMarketingSeedField(record, collection, "slug", fixture.Website.Slug)
	setNuvioMarketingSeedField(record, collection, "name", fixture.Website.Name)
	setNuvioMarketingSeedField(record, collection, "title", fixture.Website.Title)
	setNuvioMarketingSeedField(record, collection, "domain", fixture.Website.Domain)
	setNuvioMarketingSeedField(record, collection, "enabled", true)
	setNuvioMarketingSeedField(record, collection, "active", true)
	setNuvioMarketingSeedField(record, collection, "published", true)
	setNuvioMarketingSeedField(record, collection, "visible", true)
	setNuvioMarketingSeedField(record, collection, "private", false)
	setNuvioMarketingSeedField(record, collection, "status", "published")
	setNuvioMarketingSeedField(record, collection, "settings", fixture.Website.Settings)

	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to save Nuvio marketing website: %w", err)
	}
	return record, nil
}

func upsertNuvioMarketingSeedComponent(app core.App, collection *core.Collection, fixture nuvioMarketingSeedFixture, component nuvioMarketingSeedComponent, apply bool, stats *nuvioMarketingSeedStats) (*core.Record, error) {
	keyField := firstNuvioMarketingSeedField(collection, "key", "component_key", "componentKey")
	if keyField == "" {
		return nil, errors.New("Components collection has no key/component_key field")
	}
	key := strings.TrimSpace(component.Key)
	record, err := app.FindFirstRecordByFilter(collection, keyField+"={:key}", dbx.Params{"key": key})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to lookup component %q: %w", key, err)
	}

	created := false
	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		created = true
	} else if !isNuvioMarketingSeedOwnedMap(nuvioMarketingSeedMapValue(record.Get("schema")), fixture.OwnerKey) {
		return nil, fmt.Errorf("component %q already exists and is not marked as owned by %s; refusing to update", key, fixture.OwnerKey)
	}

	if created {
		countNuvioMarketingSeedStat(stats.Created, "Components")
	} else {
		countNuvioMarketingSeedStat(stats.Updated, "Components")
	}
	if !apply {
		return record, nil
	}

	setNuvioMarketingSeedField(record, collection, keyField, component.Key)
	setNuvioMarketingSeedField(record, collection, "name", component.Name)
	setNuvioMarketingSeedField(record, collection, "schema", component.Schema)
	if hasNuvioMarketingSeedField(collection, "gallery") && strings.TrimSpace(record.GetString("gallery")) == "" {
		setNuvioMarketingSeedField(record, collection, "gallery", "tailwindUI")
	}

	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to save component %q: %w", key, err)
	}
	return record, nil
}

func upsertNuvioMarketingSeedPage(app core.App, collection *core.Collection, fixture nuvioMarketingSeedFixture, page nuvioMarketingSeedPage, websiteID string, apply bool, stats *nuvioMarketingSeedStats) (*core.Record, error) {
	websiteField := resolveNuvioPublicPagesWebsiteFieldName(collection)
	if websiteField == "" {
		return nil, errors.New("Pages collection has no website/site relation field")
	}
	pageSlug := strings.TrimSpace(page.Slug)
	record, err := app.FindFirstRecordByFilter(collection, websiteField+"={:website} && slug={:slug}", dbx.Params{"website": websiteID, "slug": pageSlug})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to lookup page %q: %w", pageSlug, err)
	}

	created := false
	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		created = true
	}

	if created {
		countNuvioMarketingSeedStat(stats.Created, "Pages")
	} else {
		countNuvioMarketingSeedStat(stats.Updated, "Pages")
	}
	if !apply {
		return record, nil
	}

	setNuvioMarketingSeedField(record, collection, websiteField, websiteID)
	setNuvioMarketingSeedField(record, collection, "slug", page.Slug)
	setNuvioMarketingSeedField(record, collection, "name", page.Title)
	setNuvioMarketingSeedField(record, collection, "title", page.Title)
	setNuvioMarketingSeedField(record, collection, "seo_title", page.SeoTitle)
	setNuvioMarketingSeedField(record, collection, "seoTitle", page.SeoTitle)
	setNuvioMarketingSeedField(record, collection, "seo_description", page.SeoDescription)
	setNuvioMarketingSeedField(record, collection, "seoDescription", page.SeoDescription)
	setNuvioMarketingSeedField(record, collection, "published", page.Published)
	setNuvioMarketingSeedField(record, collection, "enabled", true)
	setNuvioMarketingSeedField(record, collection, "active", true)
	setNuvioMarketingSeedField(record, collection, "visible", true)
	setNuvioMarketingSeedField(record, collection, "private", false)
	setNuvioMarketingSeedField(record, collection, "status", "published")
	setNuvioMarketingSeedField(record, collection, "seo_noindex", page.SeoNoindex)
	setNuvioMarketingSeedField(record, collection, "seo_exclude_from_sitemap", page.SeoExcludeFromSitemap)
	setNuvioMarketingSeedField(record, collection, "seo_translations", map[string]any{})

	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to save page %q: %w", pageSlug, err)
	}
	return record, nil
}

func upsertNuvioMarketingSeedBlock(app core.App, collection *core.Collection, fixture nuvioMarketingSeedFixture, block nuvioMarketingSeedBlock, pageID string, componentRecord *core.Record, apply bool, stats *nuvioMarketingSeedStats) error {
	if !hasNuvioMarketingSeedField(collection, "page") {
		return errors.New("Blocks collection has no page relation field")
	}
	if !hasNuvioMarketingSeedField(collection, "slot") {
		return errors.New("Blocks collection has no slot field")
	}

	record, err := app.FindFirstRecordByFilter(collection, "page={:page} && slot={:slot}", dbx.Params{"page": pageID, "slot": block.Slot})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to lookup block %q: %w", block.Slot, err)
	}

	created := false
	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		created = true
	} else {
		existingComponentKey := strings.TrimSpace(record.GetString("component_key"))
		if existingComponentKey != "" && existingComponentKey != strings.TrimSpace(block.ComponentKey) {
			return fmt.Errorf("block slot %q already exists with component %q; refusing to overwrite with %q", block.Slot, existingComponentKey, block.ComponentKey)
		}
	}

	if created {
		countNuvioMarketingSeedStat(stats.Created, "Blocks")
	} else {
		countNuvioMarketingSeedStat(stats.Updated, "Blocks")
	}
	if !apply {
		return nil
	}

	setNuvioMarketingSeedField(record, collection, "page", pageID)
	setNuvioMarketingSeedField(record, collection, "slot", block.Slot)
	setNuvioMarketingSeedField(record, collection, "title", block.Title)
	setNuvioMarketingSeedField(record, collection, "component_key", block.ComponentKey)
	setNuvioMarketingSeedField(record, collection, "componentKey", block.ComponentKey)
	if componentRecord != nil && strings.TrimSpace(componentRecord.Id) != "" {
		setNuvioMarketingSeedField(record, collection, "component", componentRecord.Id)
	}
	setNuvioMarketingSeedField(record, collection, "enabled", true)
	setNuvioMarketingSeedField(record, collection, "visible", true)
	setNuvioMarketingSeedField(record, collection, "private", false)
	setNuvioMarketingSeedField(record, collection, "status", "active")
	setNuvioMarketingSeedField(record, collection, "displayOrder", block.Order)
	setNuvioMarketingSeedField(record, collection, "order", block.Order)
	setNuvioMarketingSeedField(record, collection, "props", block.Props)
	setNuvioMarketingSeedField(record, collection, "translations", map[string]any{})

	if err := app.Save(record); err != nil {
		return fmt.Errorf("failed to save block %q: %w", block.Slot, err)
	}
	return nil
}

func firstNuvioMarketingSeedField(collection *core.Collection, fields ...string) string {
	for _, fieldName := range fields {
		if hasNuvioMarketingSeedField(collection, fieldName) {
			return fieldName
		}
	}
	return ""
}

func hasNuvioMarketingSeedField(collection *core.Collection, fieldName string) bool {
	return collection != nil && collection.Fields.GetByName(fieldName) != nil
}

func setNuvioMarketingSeedField(record *core.Record, collection *core.Collection, fieldName string, value any) {
	if hasNuvioMarketingSeedField(collection, fieldName) {
		record.Set(fieldName, value)
	}
}

func nuvioMarketingSeedMapValue(value any) map[string]any {
	if result := toStringAnyMapValue(normalizeNuvioPublicJSONValue(value)); result != nil {
		return result
	}
	return map[string]any{}
}

func isNuvioMarketingSeedOwnedMap(value map[string]any, ownerKey string) bool {
	marker := nuvioMarketingSeedMapValue(value["nuvioMarketingFixture"])
	if marker == nil {
		return false
	}
	return strings.TrimSpace(fmt.Sprint(marker["ownerKey"])) == strings.TrimSpace(ownerKey) && marker["managed"] == true
}

func containsNuvioMarketingSeedRawHTML(value any) bool {
	switch typed := value.(type) {
	case string:
		lower := strings.ToLower(typed)
		return strings.Contains(lower, "<script") || strings.Contains(lower, "<iframe") || strings.Contains(lower, "<style") || strings.Contains(lower, "</")
	case []any:
		for _, item := range typed {
			if containsNuvioMarketingSeedRawHTML(item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range typed {
			if containsNuvioMarketingSeedRawHTML(item) {
				return true
			}
		}
	}
	return false
}

func validateNuvioMarketingSeedPricing(fixture nuvioMarketingSeedFixture) error {
	pricingPage := findNuvioMarketingSeedPage(fixture, "pricing")
	if pricingPage == nil {
		return errors.New("fixture missing pricing page")
	}
	plansBlock := findNuvioMarketingSeedBlock(*pricingPage, "nuvio-pricing-plans")
	if plansBlock == nil {
		return errors.New("fixture missing pricing plans block")
	}
	plansSection := nuvioMarketingSeedMapValue(plansBlock.Props["plansSection"])
	plansValue, ok := plansSection["plans"].([]any)
	if !ok || len(plansValue) != 3 {
		return errors.New("fixture pricing plansSection must contain three plans")
	}

	locked := map[string]map[string]string{
		"presenca":    {"name": "Presen\u00e7a", "founderSetup": "\u20ac590", "standardSetup": "\u20ac990", "monthly": "\u20ac69/month"},
		"crescimento": {"name": "Crescimento", "founderSetup": "\u20ac990", "standardSetup": "\u20ac1,490", "monthly": "\u20ac99/month"},
		"parceiro":    {"name": "Parceiro", "founderSetup": "\u20ac1,390", "standardSetup": "\u20ac1,990", "monthly": "\u20ac149/month"},
	}
	for _, rawPlan := range plansValue {
		plan := nuvioMarketingSeedMapValue(rawPlan)
		id := strings.TrimSpace(fmt.Sprint(plan["id"]))
		lockedPlan, ok := locked[id]
		if !ok {
			return fmt.Errorf("fixture pricing contains unknown plan id: %s", id)
		}
		for field, expected := range lockedPlan {
			if strings.TrimSpace(fmt.Sprint(plan[field])) != expected {
				return fmt.Errorf("fixture pricing drift for %s.%s", id, field)
			}
		}
	}

	if _, ok := plansBlock.Props["comparison"]; !ok {
		return errors.New("fixture pricing plans block missing comparison props")
	}
	if _, ok := plansBlock.Props["foundation"]; !ok {
		return errors.New("fixture pricing plans block missing foundation props")
	}
	return nil
}

func validateNuvioMarketingSeedContact(fixture nuvioMarketingSeedFixture) error {
	contactPage := findNuvioMarketingSeedPage(fixture, "contact")
	if contactPage == nil {
		return errors.New("fixture missing contact page")
	}
	requestBlock := findNuvioMarketingSeedBlock(*contactPage, "nuvio-contact-request")
	if requestBlock == nil {
		return errors.New("fixture missing contact request block")
	}
	forbiddenKeys := map[string]bool{
		"action": true, "endpoint": true, "honeypot": true, "honeypotField": true,
		"website_url": true, "fieldNames": true, "acceptedFields": true, "validation": true,
	}
	var walk func(any) error
	walk = func(value any) error {
		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				if err := walk(item); err != nil {
					return err
				}
			}
		case map[string]any:
			for key, item := range typed {
				if forbiddenKeys[key] {
					return fmt.Errorf("fixture contact props include forbidden mechanic key: %s", key)
				}
				if err := walk(item); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(requestBlock.Props)
}

func findNuvioMarketingSeedPage(fixture nuvioMarketingSeedFixture, slug string) *nuvioMarketingSeedPage {
	for index := range fixture.Pages {
		if strings.TrimSpace(fixture.Pages[index].Slug) == slug {
			return &fixture.Pages[index]
		}
	}
	return nil
}

func findNuvioMarketingSeedBlock(page nuvioMarketingSeedPage, componentKey string) *nuvioMarketingSeedBlock {
	for index := range page.Blocks {
		if strings.TrimSpace(page.Blocks[index].ComponentKey) == componentKey {
			return &page.Blocks[index]
		}
	}
	return nil
}

func countNuvioMarketingSeedStat(target map[string]int, collection string) {
	target[collection] = target[collection] + 1
}

func printNuvioMarketingSeedSummary(fixture nuvioMarketingSeedFixture, stats nuvioMarketingSeedStats, apply bool) {
	mode := "DRY-RUN"
	if apply {
		mode = "APPLY"
	}
	fmt.Printf("Nuvio marketing seed mode: %s\n", mode)
	fmt.Printf("Website slug: %s\n", fixture.Website.Slug)
	printNuvioMarketingSeedStatMap("created", stats.Created)
	printNuvioMarketingSeedStatMap("updated", stats.Updated)
	printNuvioMarketingSeedStatMap("skipped", stats.Skipped)
	for _, warning := range stats.Warnings {
		fmt.Printf("WARNING: %s\n", warning)
	}
	if !apply {
		fmt.Println("No records were written. Re-run with --apply to create/update fixture-owned records.")
	}
}

func printNuvioMarketingSeedStatMap(label string, values map[string]int) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		fmt.Printf("%s: none\n", label)
		return
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, values[key]))
	}
	fmt.Printf("%s: %s\n", label, strings.Join(parts, ", "))
}
