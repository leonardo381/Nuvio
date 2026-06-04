package cmsqasnapshot

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	createConfirmToken   = "CREATE_CMS_QA_SNAPSHOT"
	restoreConfirmToken  = "RESTORE_CMS_QA_SNAPSHOT"
	devResetOptInEnv     = "NUVIO_ALLOW_DEV_RESET"
	cmsSnapshotNote      = "Local dev CMS QA snapshot. Do not use in production."
	cmsRestoreNote       = "Local dev CMS QA restore. Do not use in production."
	cmsRestoreBackupNote = "Local dev CMS QA pre-restore safety backup. Do not use in production."
	assetModeAll         = "all"
	assetModeWebsite     = "website"
)

var snapshotNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var assetRefPattern = regexp.MustCompile(`(?is)"collection"\s*:\s*"Assets".{0,300}?"recordId"\s*:\s*"([a-z0-9]+)"`)

type collectionSpec struct {
	Key          string
	Label        string
	CollectionID string
	Table        string
	FileFields   []string
}

var specsByKey = map[string]collectionSpec{
	"websites": {
		Key:          "websites",
		Label:        "Websites",
		CollectionID: "pbc_2619338178",
		Table:        "Websites",
		FileFields:   []string{"logo", "seoImage"},
	},
	"pages": {
		Key:          "pages",
		Label:        "Pages",
		CollectionID: "pbc_3945946014",
		Table:        "pages",
		FileFields:   []string{"seo_social_image"},
	},
	"blocks": {
		Key:          "blocks",
		Label:        "Blocks",
		CollectionID: "pbc_4194232374",
		Table:        "blocks",
		FileFields:   []string{"image"},
	},
	"components": {
		Key:          "components",
		Label:        "Components",
		CollectionID: "pbc_184785686",
		Table:        "Components",
	},
	"assets": {
		Key:          "assets",
		Label:        "Assets",
		CollectionID: "pbc_1321337024",
		Table:        "assets",
		FileFields:   []string{"file"},
	},
}

var snapshotOrder = []string{"websites", "components", "assets", "pages", "blocks"}
var restoreDeleteOrder = []string{"blocks", "pages", "assets", "websites", "components"}
var restoreInsertOrder = []string{"websites", "components", "assets", "pages", "blocks"}

var optionalCMSCollectionCandidates = []manifestSkippedCollection{
	{Name: "Templates", Reason: "No standalone Templates collection is present in the current workspace."},
	{Name: "Navigation", Reason: "No standalone navigation collection is present in the current workspace."},
	{Name: "Menus", Reason: "No standalone menus collection is present in the current workspace."},
	{Name: "Redirects", Reason: "No standalone redirects collection is present in the current workspace."},
	{Name: "CMS settings", Reason: "CMS settings are stored on Websites/settings and included with the Websites record."},
	{Name: "SEO settings", Reason: "Website/page SEO fields are stored on Websites and Pages records and included there."},
	{Name: "Translations", Reason: "Block/Page translations are stored on Blocks.translations and Pages.seo_translations and included there."},
}

var operationalCollectionCandidates = []string{
	"Contacts",
	"Whatsapp",
	"Appointments",
	"Subscribers",
	"SubscriberGroups",
	"Campaigns",
	"BookingServices",
	"BookingAvailability",
	"BookingExceptions",
	"Reviews",
}

type createOptions struct {
	Name           string
	WebsiteID      string
	WebsiteSlug    string
	AssetsMode     string
	Confirm        string
	Overwrite      bool
	BackendStopped bool
}

type restoreOptions struct {
	Name           string
	WebsiteID      string
	Confirm        string
	BackendStopped bool
}

type snapshotManifest struct {
	SnapshotName                   string                      `json:"snapshotName"`
	CreatedAt                      string                      `json:"createdAt"`
	WebsiteID                      string                      `json:"websiteId"`
	WebsiteSlug                    string                      `json:"websiteSlug"`
	AssetsMode                     string                      `json:"assetsMode"`
	SourcePath                     string                      `json:"sourcePath"`
	DestinationPath                string                      `json:"destinationPath"`
	Collections                    []manifestCollection        `json:"collections"`
	NativeFileFields               []string                    `json:"nativeFileFields"`
	StorageByCollection            []manifestStorageCollection `json:"storageByCollection"`
	SkippedCollections             []manifestSkippedCollection `json:"skippedCollections"`
	OperationalCollectionsExcluded []string                    `json:"operationalCollectionsExcluded"`
	StorageFiles                   int64                       `json:"storageFiles"`
	StorageBytes                   int64                       `json:"storageBytes"`
	Note                           string                      `json:"note"`
}

type manifestCollection struct {
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	CollectionID string   `json:"collectionId"`
	Table        string   `json:"table"`
	Records      int      `json:"records"`
	FileFields   []string `json:"fileFields,omitempty"`
}

type manifestStorageCollection struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	CollectionID string `json:"collectionId"`
	Files        int64  `json:"files"`
	Bytes        int64  `json:"bytes"`
}

type manifestSkippedCollection struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type restoreLog struct {
	RestoredSnapshotName string `json:"restoredSnapshotName"`
	RestoredAt           string `json:"restoredAt"`
	WebsiteID            string `json:"websiteId"`
	WebsiteSlug          string `json:"websiteSlug"`
	SourceSnapshotPath   string `json:"sourceSnapshotPath"`
	DestinationPath      string `json:"destinationPath"`
	SafetyBackupPath     string `json:"safetyBackupPath"`
	RestoredRecords      int    `json:"restoredRecords"`
	StorageFiles         int64  `json:"storageFiles"`
	StorageBytes         int64  `json:"storageBytes"`
	Note                 string `json:"note"`
}

type snapshotCollection struct {
	Key          string                   `json:"key"`
	Label        string                   `json:"label"`
	CollectionID string                   `json:"collectionId"`
	Table        string                   `json:"table"`
	Columns      []string                 `json:"columns"`
	FileFields   []string                 `json:"fileFields,omitempty"`
	Records      []map[string]interface{} `json:"records"`
}

type copyStats struct {
	Files int64
	Bytes int64
}

type storageRef struct {
	CollectionKey string
	CollectionID  string
	RecordID      string
}

type websiteInfo struct {
	ID   string
	Slug string
}

type snapshotData struct {
	Website                        websiteInfo
	AssetsMode                     string
	Collections                    map[string]snapshotCollection
	SkippedCollections             []manifestSkippedCollection
	OperationalCollectionsExcluded []string
}

func RunCreate(args []string) error {
	opts, err := parseCreateOptions(args)
	if err != nil {
		return err
	}
	if err := requireDevOptIn(); err != nil {
		return err
	}
	if !snapshotNamePattern.MatchString(opts.Name) {
		return errors.New("invalid snapshot name. Use only letters, numbers, dash, and underscore")
	}
	if opts.WebsiteID == "" && opts.WebsiteSlug == "" {
		return errors.New("provide --websiteId or --websiteSlug")
	}
	if opts.WebsiteID != "" && opts.WebsiteSlug != "" {
		return errors.New("provide only one of --websiteId or --websiteSlug")
	}

	writeMode := strings.TrimSpace(opts.Confirm) == createConfirmToken
	if strings.TrimSpace(opts.Confirm) != "" && !writeMode {
		return fmt.Errorf("invalid --confirm token; expected %s", createConfirmToken)
	}
	if writeMode && !opts.BackendStopped {
		return errors.New("Stop the Nuvio/PocketBase backend before creating a CMS snapshot.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve current workspace: %w", err)
	}
	if err := ensureWorkspace(cwd); err != nil {
		return err
	}

	db, cleanup, err := openWorkspaceDBSnapshot(cwd)
	if err != nil {
		return err
	}
	defer cleanup()

	data, err := buildSnapshotData(db, cwd, opts.WebsiteID, opts.WebsiteSlug, opts.AssetsMode)
	if err != nil {
		return err
	}

	snapshotRoot := filepath.Join(cwd, "dev_qa_snapshots", "cms", opts.Name)
	recordsRoot := filepath.Join(snapshotRoot, "records")
	storageRoot := filepath.Join(snapshotRoot, "storage")
	manifestPath := filepath.Join(snapshotRoot, "manifest.json")

	if err := ensurePathInsideWorkspace(cwd, snapshotRoot); err != nil {
		return err
	}
	if err := ensurePathInsideWorkspace(cwd, manifestPath); err != nil {
		return err
	}

	storageByCollection, err := scanSnapshotStorageByCollection(cwd, data.Collections)
	if err != nil {
		return err
	}
	storageStats := sumStorageStats(storageByCollection)

	if !writeMode {
		printCreateSummary("DRY-RUN", opts.Name, data, filepath.Join(cwd, "pb_data"), snapshotRoot, storageStats, storageByCollection)
		return nil
	}

	if err := ensureSnapshotDestination(snapshotRoot, opts.Overwrite); err != nil {
		return err
	}
	if err := os.MkdirAll(recordsRoot, 0o755); err != nil {
		return fmt.Errorf("failed to prepare records directory: %w", err)
	}
	if err := os.MkdirAll(storageRoot, 0o755); err != nil {
		return fmt.Errorf("failed to prepare storage directory: %w", err)
	}

	for _, key := range snapshotOrder {
		if err := writeJSONFile(filepath.Join(recordsRoot, key+".json"), data.Collections[key]); err != nil {
			return fmt.Errorf("failed to write %s snapshot records: %w", key, err)
		}
	}

	storageStats, err = copySnapshotStorage(cwd, storageRoot, data.Collections)
	if err != nil {
		return err
	}

	manifest := snapshotManifest{
		SnapshotName:                   opts.Name,
		CreatedAt:                      time.Now().UTC().Format(time.RFC3339),
		WebsiteID:                      data.Website.ID,
		WebsiteSlug:                    data.Website.Slug,
		AssetsMode:                     data.AssetsMode,
		SourcePath:                     filepath.Join(cwd, "pb_data"),
		DestinationPath:                snapshotRoot,
		Collections:                    buildManifestCollections(data.Collections),
		NativeFileFields:               buildNativeFileFieldList(data.Collections),
		StorageByCollection:            buildManifestStorageCollections(data.Collections, storageByCollection),
		SkippedCollections:             data.SkippedCollections,
		OperationalCollectionsExcluded: data.OperationalCollectionsExcluded,
		StorageFiles:                   storageStats.Files,
		StorageBytes:                   storageStats.Bytes,
		Note:                           cmsSnapshotNote,
	}
	if err := writeJSONFile(manifestPath, manifest); err != nil {
		return fmt.Errorf("failed to write snapshot manifest: %w", err)
	}

	printCreateSummary("WRITE", opts.Name, data, filepath.Join(cwd, "pb_data"), snapshotRoot, storageStats, storageByCollection)
	return nil
}

func RunRestore(args []string) error {
	opts, err := parseRestoreOptions(args)
	if err != nil {
		return err
	}
	if err := requireDevOptIn(); err != nil {
		return err
	}
	if !snapshotNamePattern.MatchString(opts.Name) {
		return errors.New("invalid snapshot name. Use only letters, numbers, dash, and underscore")
	}

	writeMode := strings.TrimSpace(opts.Confirm) == restoreConfirmToken
	if strings.TrimSpace(opts.Confirm) != "" && !writeMode {
		return fmt.Errorf("invalid --confirm token; expected %s", restoreConfirmToken)
	}
	if writeMode && !opts.BackendStopped {
		return errors.New("Stop the Nuvio/PocketBase backend before restoring a CMS snapshot.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve current workspace: %w", err)
	}
	if err := ensureWorkspace(cwd); err != nil {
		return err
	}
	if writeMode {
		if err := ensureLocalDevWorkspacePath(cwd); err != nil {
			return err
		}
	}

	snapshotRoot := filepath.Join(cwd, "dev_qa_snapshots", "cms", opts.Name)
	if err := ensurePathInsideWorkspace(cwd, snapshotRoot); err != nil {
		return err
	}

	manifest, collections, err := loadSnapshot(snapshotRoot)
	if err != nil {
		return err
	}
	if opts.WebsiteID != "" && opts.WebsiteID != manifest.WebsiteID {
		return fmt.Errorf("snapshot websiteId %q does not match requested --websiteId %q", manifest.WebsiteID, opts.WebsiteID)
	}
	if strings.TrimSpace(manifest.WebsiteID) == "" {
		return errors.New("snapshot manifest is missing websiteId")
	}
	assetsMode, err := normalizeAssetsModeForRestore(manifest.AssetsMode)
	if err != nil {
		return err
	}

	if err := validateSnapshotStorage(snapshotRoot, collections); err != nil {
		return err
	}
	storageByCollection, err := scanSnapshotStorageRootByCollection(filepath.Join(snapshotRoot, "storage"), collections)
	if err != nil {
		return err
	}
	storageStats := sumStorageStats(storageByCollection)

	pbDataPath := filepath.Join(cwd, "pb_data")
	utcNow := time.Now().UTC()
	timestamp := utcNow.Format("20060102T150405Z")
	safetyBackupName := "cms_pre_restore_backup_" + timestamp
	safetyBackupRoot := filepath.Join(cwd, "dev_qa_snapshots", "cms", safetyBackupName)
	safetyBackupPBData := filepath.Join(safetyBackupRoot, "pb_data")
	restoreLogsRoot := filepath.Join(cwd, "dev_qa_snapshots", "cms", "restore_logs")
	restoreLogPath := filepath.Join(restoreLogsRoot, "restore_"+timestamp+".json")
	tempStorageRoot := filepath.Join(cwd, "dev_qa_snapshots", "cms", "restore_tmp_"+timestamp)

	if !writeMode {
		printRestoreDryRunSummary(opts.Name, manifest, collections, pbDataPath, safetyBackupPBData, storageStats, storageByCollection, assetsMode)
		return nil
	}

	for _, path := range []string{safetyBackupRoot, safetyBackupPBData, restoreLogsRoot, restoreLogPath, tempStorageRoot} {
		if err := ensurePathInsideWorkspace(cwd, path); err != nil {
			return err
		}
	}
	if err := ensureExistingPBDataIsSafeDirectory(pbDataPath); err != nil {
		return err
	}
	if err := ensureSafetyBackupDestination(safetyBackupRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(safetyBackupPBData, 0o755); err != nil {
		return fmt.Errorf("failed to prepare safety backup directory: %w", err)
	}
	backupStats, err := copyTree(pbDataPath, safetyBackupPBData)
	if err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}
	backupManifest := snapshotManifest{
		SnapshotName:    safetyBackupName,
		CreatedAt:       utcNow.Format(time.RFC3339),
		WebsiteID:       manifest.WebsiteID,
		WebsiteSlug:     manifest.WebsiteSlug,
		SourcePath:      pbDataPath,
		DestinationPath: safetyBackupPBData,
		StorageFiles:    backupStats.Files,
		StorageBytes:    backupStats.Bytes,
		Note:            cmsRestoreBackupNote,
	}
	if err := writeJSONFile(filepath.Join(safetyBackupRoot, "manifest.json"), backupManifest); err != nil {
		return fmt.Errorf("failed to write safety backup manifest: %w", err)
	}

	if err := os.MkdirAll(tempStorageRoot, 0o755); err != nil {
		return fmt.Errorf("failed to prepare temporary restore storage: %w", err)
	}
	defer os.RemoveAll(tempStorageRoot)
	preparedStorageRoot := filepath.Join(tempStorageRoot, "storage")
	if err := copySnapshotStorageRoot(filepath.Join(snapshotRoot, "storage"), preparedStorageRoot); err != nil {
		return fmt.Errorf("failed to prepare restore storage copy. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	db, err := openSQLite(filepath.Join(pbDataPath, "data.db"))
	if err != nil {
		return fmt.Errorf("failed to open pb_data/data.db. Stop the backend before restoring. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}
	defer db.Close()

	currentRefs, err := collectCurrentStorageRefs(db, manifest.WebsiteID, assetsMode, collections)
	if err != nil {
		return fmt.Errorf("failed to collect current CMS storage refs. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	if err := restoreDatabase(db, manifest.WebsiteID, assetsMode, collections); err != nil {
		return fmt.Errorf("failed to restore CMS records. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	if err := restoreStorage(cwd, preparedStorageRoot, currentRefs, collections); err != nil {
		return fmt.Errorf("failed to restore CMS storage. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	recordCount := countSnapshotRecords(collections)
	restoreLog := restoreLog{
		RestoredSnapshotName: opts.Name,
		RestoredAt:           time.Now().UTC().Format(time.RFC3339),
		WebsiteID:            manifest.WebsiteID,
		WebsiteSlug:          manifest.WebsiteSlug,
		SourceSnapshotPath:   snapshotRoot,
		DestinationPath:      pbDataPath,
		SafetyBackupPath:     safetyBackupPBData,
		RestoredRecords:      recordCount,
		StorageFiles:         storageStats.Files,
		StorageBytes:         storageStats.Bytes,
		Note:                 cmsRestoreNote,
	}
	if err := os.MkdirAll(restoreLogsRoot, 0o755); err != nil {
		return fmt.Errorf("restore succeeded but failed to prepare restore logs directory: %w", err)
	}
	if err := writeJSONFile(restoreLogPath, restoreLog); err != nil {
		return fmt.Errorf("restore succeeded but failed to write restore log: %w", err)
	}

	printRestoreWriteSummary(opts.Name, manifest, pbDataPath, safetyBackupPBData, restoreLogPath, recordCount, storageStats, storageByCollection, assetsMode)
	return nil
}

func parseCreateOptions(args []string) (createOptions, error) {
	var opts createOptions
	fs := flag.NewFlagSet("create_cms_qa_snapshot", flag.ContinueOnError)
	fs.StringVar(&opts.Name, "name", "", "Snapshot name")
	fs.StringVar(&opts.Name, "snapshotName", "", "Snapshot name alias")
	fs.StringVar(&opts.WebsiteID, "websiteId", "", "Target website id")
	fs.StringVar(&opts.WebsiteSlug, "websiteSlug", "", "Target website slug")
	fs.StringVar(&opts.AssetsMode, "assetsMode", assetModeAll, "Assets coverage mode: all or website")
	fs.StringVar(&opts.Confirm, "confirm", "", "Confirmation token for write mode")
	fs.BoolVar(&opts.Overwrite, "overwrite", false, "Overwrite existing snapshot if present")
	fs.BoolVar(&opts.BackendStopped, "backendStopped", false, "Confirm backend is stopped before copying")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	opts.Name = strings.TrimSpace(opts.Name)
	opts.WebsiteID = strings.TrimSpace(opts.WebsiteID)
	opts.WebsiteSlug = strings.TrimSpace(opts.WebsiteSlug)
	assetsMode, err := normalizeAssetsMode(opts.AssetsMode, assetModeAll)
	if err != nil {
		return opts, err
	}
	opts.AssetsMode = assetsMode
	if opts.Name == "" {
		return opts, errors.New("--name or --snapshotName is required")
	}
	return opts, nil
}

func parseRestoreOptions(args []string) (restoreOptions, error) {
	var opts restoreOptions
	fs := flag.NewFlagSet("restore_cms_qa_snapshot", flag.ContinueOnError)
	fs.StringVar(&opts.Name, "name", "", "Snapshot name")
	fs.StringVar(&opts.Name, "snapshotName", "", "Snapshot name alias")
	fs.StringVar(&opts.WebsiteID, "websiteId", "", "Optional guard: expected snapshot website id")
	fs.StringVar(&opts.Confirm, "confirm", "", "Confirmation token for write mode")
	fs.BoolVar(&opts.BackendStopped, "backendStopped", false, "Confirm backend is stopped before restoring")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	opts.Name = strings.TrimSpace(opts.Name)
	opts.WebsiteID = strings.TrimSpace(opts.WebsiteID)
	if opts.Name == "" {
		return opts, errors.New("--name or --snapshotName is required")
	}
	return opts, nil
}

func normalizeAssetsMode(value string, defaultMode string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(value))
	if mode == "" {
		mode = defaultMode
	}
	switch mode {
	case assetModeAll, assetModeWebsite:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid --assetsMode %q; expected %q or %q", value, assetModeAll, assetModeWebsite)
	}
}

func normalizeAssetsModeForRestore(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		// Older snapshots did not declare full-library Assets coverage, so restore them
		// using the original website-scoped behavior instead of deleting every asset.
		return assetModeWebsite, nil
	}
	return normalizeAssetsMode(value, assetModeWebsite)
}

func requireDevOptIn() error {
	if strings.TrimSpace(os.Getenv(devResetOptInEnv)) != "1" {
		return fmt.Errorf("%s must be set to 1", devResetOptInEnv)
	}
	return nil
}

func ensureWorkspace(cwd string) error {
	required := []string{
		filepath.Join(cwd, "AGENTS.md"),
		filepath.Join(cwd, "go.mod"),
		filepath.Join(cwd, "pb_data"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			return errors.New("this tool must be run from the Nuvio workspace root")
		}
	}
	return nil
}

func ensureLocalDevWorkspacePath(cwd string) error {
	lowerPath := strings.ToLower(filepath.Clean(cwd))
	if strings.HasPrefix(lowerPath, `\\`) {
		return errors.New("write mode is allowed only for local/dev workspace paths")
	}
	localMarkers := []string{`\users\`, `/users/`, `\home\`, `/home/`, `\tmp\`, `/tmp/`, `\documents\`, `/documents/`}
	for _, marker := range localMarkers {
		if strings.Contains(lowerPath, marker) {
			return nil
		}
	}
	return errors.New("write mode is allowed only for local/dev workspace paths")
}

func openWorkspaceDBSnapshot(cwd string) (*sql.DB, func(), error) {
	source := filepath.Join(cwd, "pb_data", "data.db")
	if _, err := os.Stat(source); err != nil {
		return nil, func() {}, fmt.Errorf("source pb_data/data.db not found: %w", err)
	}
	tmp, err := os.CreateTemp("", "nuvio_cms_snapshot_*.db")
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create temporary db copy: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		return nil, func() {}, err
	}
	if err := copyFile(source, tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, func() {}, fmt.Errorf("failed to copy pb_data/data.db for snapshot read: %w", err)
	}
	db, err := openSQLite(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return nil, func() {}, fmt.Errorf("failed to open temporary db copy: %w", err)
	}
	cleanup := func() {
		db.Close()
		os.Remove(tmpPath)
	}
	return db, cleanup, nil
}

func openSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func buildSnapshotData(db *sql.DB, cwd string, websiteID string, websiteSlug string, assetsMode string) (snapshotData, error) {
	assetsMode, err := normalizeAssetsMode(assetsMode, assetModeAll)
	if err != nil {
		return snapshotData{}, err
	}

	website, err := resolveWebsite(db, websiteID, websiteSlug)
	if err != nil {
		return snapshotData{}, err
	}

	collections := map[string]snapshotCollection{}
	websiteRecords, err := queryRecords(db, specsByKey["websites"], "id = ?", []interface{}{website.ID})
	if err != nil {
		return snapshotData{}, err
	}
	collections["websites"] = websiteRecords

	components, err := queryRecords(db, specsByKey["components"], "", nil)
	if err != nil {
		return snapshotData{}, err
	}
	collections["components"] = components

	pages, err := queryRecords(db, specsByKey["pages"], "website = ?", []interface{}{website.ID})
	if err != nil {
		return snapshotData{}, err
	}
	collections["pages"] = pages
	pageIDs := recordIDs(pages.Records)

	blocks, err := queryRecordsByFieldIDs(db, specsByKey["blocks"], "page", pageIDs)
	if err != nil {
		return snapshotData{}, err
	}
	collections["blocks"] = blocks

	if assetsMode == assetModeAll {
		assets, err := queryRecords(db, specsByKey["assets"], "", nil)
		if err != nil {
			return snapshotData{}, err
		}
		collections["assets"] = assets
	} else {
		assetIDs := map[string]bool{}
		assetIDsFromWebsite, err := queryIDs(db, specsByKey["assets"], "website = ?", []interface{}{website.ID})
		if err != nil {
			return snapshotData{}, err
		}
		for _, id := range assetIDsFromWebsite {
			assetIDs[id] = true
		}
		collectAssetIDsFromCollections(assetIDs, websiteRecords, pages, blocks)
		assets, err := queryRecordsByIDs(db, specsByKey["assets"], sortedKeys(assetIDs))
		if err != nil {
			return snapshotData{}, err
		}
		collections["assets"] = assets
	}

	for _, key := range snapshotOrder {
		if _, ok := collections[key]; !ok {
			collections[key] = emptySnapshotCollection(specsByKey[key])
		}
	}

	if err := validateSourceStorage(cwd, collections); err != nil {
		return snapshotData{}, err
	}

	return snapshotData{
		Website:                        website,
		AssetsMode:                     assetsMode,
		Collections:                    collections,
		SkippedCollections:             buildSkippedCMSCollections(db),
		OperationalCollectionsExcluded: buildOperationalExclusions(db),
	}, nil
}

func resolveWebsite(db *sql.DB, websiteID string, websiteSlug string) (websiteInfo, error) {
	if websiteID != "" {
		row := db.QueryRow("SELECT id, slug FROM "+quoteIdent(specsByKey["websites"].Table)+" WHERE id = ?", websiteID)
		var info websiteInfo
		if err := row.Scan(&info.ID, &info.Slug); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return websiteInfo{}, fmt.Errorf("website not found by id %q", websiteID)
			}
			return websiteInfo{}, err
		}
		return info, nil
	}
	row := db.QueryRow("SELECT id, slug FROM "+quoteIdent(specsByKey["websites"].Table)+" WHERE lower(slug) = lower(?)", websiteSlug)
	var info websiteInfo
	if err := row.Scan(&info.ID, &info.Slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return websiteInfo{}, fmt.Errorf("website not found by slug %q", websiteSlug)
		}
		return websiteInfo{}, err
	}
	return info, nil
}

func queryRecords(db *sql.DB, spec collectionSpec, where string, args []interface{}) (snapshotCollection, error) {
	columns, err := tableColumns(db, spec.Table)
	if err != nil {
		return snapshotCollection{}, err
	}
	fileFields, err := collectionFileFields(db, spec)
	if err != nil {
		return snapshotCollection{}, err
	}
	query := "SELECT " + quoteIdentList(columns) + " FROM " + quoteIdent(spec.Table)
	if strings.TrimSpace(where) != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY id"
	rows, err := db.Query(query, args...)
	if err != nil {
		return snapshotCollection{}, fmt.Errorf("failed to query %s: %w", spec.Label, err)
	}
	defer rows.Close()
	records, err := scanRecordRows(rows, columns)
	if err != nil {
		return snapshotCollection{}, err
	}
	return snapshotCollection{Key: spec.Key, Label: spec.Label, CollectionID: spec.CollectionID, Table: spec.Table, Columns: columns, FileFields: fileFields, Records: records}, nil
}

func queryRecordsByIDs(db *sql.DB, spec collectionSpec, ids []string) (snapshotCollection, error) {
	if len(ids) == 0 {
		return emptySnapshotCollectionWithColumns(db, spec)
	}
	where, args := inClause("id", ids)
	return queryRecords(db, spec, where, args)
}

func queryRecordsByFieldIDs(db *sql.DB, spec collectionSpec, field string, ids []string) (snapshotCollection, error) {
	if len(ids) == 0 {
		return emptySnapshotCollectionWithColumns(db, spec)
	}
	where, args := inClause(field, ids)
	return queryRecords(db, spec, where, args)
}

func emptySnapshotCollection(spec collectionSpec) snapshotCollection {
	return snapshotCollection{Key: spec.Key, Label: spec.Label, CollectionID: spec.CollectionID, Table: spec.Table, Columns: []string{}, FileFields: spec.FileFields, Records: []map[string]interface{}{}}
}

func emptySnapshotCollectionWithColumns(db *sql.DB, spec collectionSpec) (snapshotCollection, error) {
	columns, err := tableColumns(db, spec.Table)
	if err != nil {
		return snapshotCollection{}, err
	}
	fileFields, err := collectionFileFields(db, spec)
	if err != nil {
		return snapshotCollection{}, err
	}
	return snapshotCollection{Key: spec.Key, Label: spec.Label, CollectionID: spec.CollectionID, Table: spec.Table, Columns: columns, FileFields: fileFields, Records: []map[string]interface{}{}}, nil
}

func collectionFileFields(db *sql.DB, spec collectionSpec) ([]string, error) {
	row := db.QueryRow(
		"SELECT fields FROM _collections WHERE id = ? OR lower(name) = lower(?) OR lower(name) = lower(?) LIMIT 1",
		spec.CollectionID,
		spec.Table,
		spec.Label,
	)
	var rawFields string
	if err := row.Scan(&rawFields); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return append([]string{}, spec.FileFields...), nil
		}
		return nil, fmt.Errorf("failed to inspect %s collection fields: %w", spec.Label, err)
	}
	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(rawFields), &fields); err != nil {
		return nil, fmt.Errorf("failed to parse %s collection fields: %w", spec.Label, err)
	}
	fileFields := []string{}
	for _, field := range fields {
		if !strings.EqualFold(strings.TrimSpace(toString(field["type"])), "file") {
			continue
		}
		name := strings.TrimSpace(toString(field["name"]))
		if name != "" {
			fileFields = append(fileFields, name)
		}
	}
	return uniqueStringsPreserveOrder(append(fileFields, spec.FileFields...)), nil
}

func tableColumns(db *sql.DB, table string) ([]string, error) {
	rows, err := db.Query("PRAGMA table_info(" + quoteIdent(table) + ")")
	if err != nil {
		return nil, fmt.Errorf("failed to inspect table %s: %w", table, err)
	}
	defer rows.Close()
	columns := []string{}
	for rows.Next() {
		var cid int
		var name, typeName string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typeName, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("table %s not found or has no columns", table)
	}
	return columns, nil
}

func scanRecordRows(rows *sql.Rows, columns []string) ([]map[string]interface{}, error) {
	records := []map[string]interface{}{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		dest := make([]interface{}, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		record := map[string]interface{}{}
		for i, column := range columns {
			record[column] = normalizeDBValue(values[i])
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func normalizeDBValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	default:
		return typed
	}
}

func queryIDs(db *sql.DB, spec collectionSpec, where string, args []interface{}) ([]string, error) {
	query := "SELECT id FROM " + quoteIdent(spec.Table)
	if strings.TrimSpace(where) != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY id"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}

func collectAssetIDsFromCollections(ids map[string]bool, collections ...snapshotCollection) {
	for _, collection := range collections {
		for _, record := range collection.Records {
			for _, value := range record {
				collectAssetIDsFromValue(ids, value)
			}
		}
	}
}

func collectAssetIDsFromValue(ids map[string]bool, value interface{}) {
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return
		}
		for _, match := range assetRefPattern.FindAllStringSubmatch(text, -1) {
			if len(match) > 1 && isPocketBaseID(match[1]) {
				ids[match[1]] = true
			}
		}
		if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
			var decoded interface{}
			if err := json.Unmarshal([]byte(text), &decoded); err == nil {
				scanJSONForAssetRefs(ids, decoded)
			}
		}
	case map[string]interface{}:
		scanJSONForAssetRefs(ids, typed)
	case []interface{}:
		scanJSONForAssetRefs(ids, typed)
	}
}

func scanJSONForAssetRefs(ids map[string]bool, value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		collection := strings.TrimSpace(toString(typed["collection"]))
		if strings.EqualFold(collection, "Assets") {
			recordID := strings.TrimSpace(toString(typed["recordId"]))
			if isPocketBaseID(recordID) {
				ids[recordID] = true
			}
		}
		for _, nested := range typed {
			scanJSONForAssetRefs(ids, nested)
		}
	case []interface{}:
		for _, nested := range typed {
			scanJSONForAssetRefs(ids, nested)
		}
	}
}

func isPocketBaseID(value string) bool {
	if len(value) < 8 || len(value) > 32 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func validateSourceStorage(cwd string, collections map[string]snapshotCollection) error {
	return validateStorage(filepath.Join(cwd, "pb_data", "storage"), collections)
}

func validateSnapshotStorage(snapshotRoot string, collections map[string]snapshotCollection) error {
	return validateStorage(filepath.Join(snapshotRoot, "storage"), collections)
}

func validateStorage(storageRoot string, collections map[string]snapshotCollection) error {
	for key, collection := range collections {
		spec := specsByKey[key]
		fileFields := collection.FileFields
		if len(fileFields) == 0 {
			fileFields = spec.FileFields
		}
		if len(fileFields) == 0 {
			continue
		}
		for _, record := range collection.Records {
			recordID := strings.TrimSpace(toString(record["id"]))
			if recordID == "" {
				continue
			}
			for _, fileField := range fileFields {
				for _, fileName := range extractFileNames(record[fileField]) {
					if err := validateFileName(fileName); err != nil {
						return fmt.Errorf("invalid file name in %s/%s.%s: %w", collection.Label, recordID, fileField, err)
					}
					path := filepath.Join(storageRoot, spec.CollectionID, recordID, fileName)
					info, err := os.Stat(path)
					if err != nil {
						if os.IsNotExist(err) {
							return fmt.Errorf("missing storage file for %s/%s.%s: %s", collection.Label, recordID, fileField, path)
						}
						return err
					}
					if info.IsDir() {
						return fmt.Errorf("storage file path is a directory for %s/%s.%s: %s", collection.Label, recordID, fileField, path)
					}
				}
			}
		}
	}
	return nil
}

func extractFileNames(value interface{}) []string {
	result := []string{}
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return result
		}
		if strings.HasPrefix(text, "[") {
			var decoded []interface{}
			if err := json.Unmarshal([]byte(text), &decoded); err == nil {
				for _, item := range decoded {
					result = appendFileName(result, toString(item))
				}
				return result
			}
		}
		result = appendFileName(result, text)
	case []interface{}:
		for _, item := range typed {
			result = appendFileName(result, toString(item))
		}
	}
	return result
}

func appendFileName(result []string, raw string) []string {
	name := strings.TrimSpace(raw)
	if name != "" {
		result = append(result, name)
	}
	return result
}

func validateFileName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("empty filename")
	}
	if filepath.Base(name) != name || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("unsafe filename %q", name)
	}
	return nil
}

func copySnapshotStorage(cwd string, destinationStorageRoot string, collections map[string]snapshotCollection) (copyStats, error) {
	sourceStorageRoot := filepath.Join(cwd, "pb_data", "storage")
	stats := copyStats{}
	for _, ref := range storageRefsFromCollections(collections) {
		source := filepath.Join(sourceStorageRoot, ref.CollectionID, ref.RecordID)
		if _, err := os.Stat(source); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return stats, err
		}
		destination := filepath.Join(destinationStorageRoot, ref.CollectionID, ref.RecordID)
		copied, err := copyTree(source, destination)
		if err != nil {
			return stats, err
		}
		stats.Files += copied.Files
		stats.Bytes += copied.Bytes
	}
	return stats, nil
}

func scanSnapshotStorageByCollection(cwd string, collections map[string]snapshotCollection) (map[string]copyStats, error) {
	return scanStorageRootByCollection(filepath.Join(cwd, "pb_data", "storage"), collections)
}

func scanSnapshotStorageRootByCollection(storageRoot string, collections map[string]snapshotCollection) (map[string]copyStats, error) {
	return scanStorageRootByCollection(storageRoot, collections)
}

func scanStorageRootByCollection(sourceStorageRoot string, collections map[string]snapshotCollection) (map[string]copyStats, error) {
	statsByCollection := map[string]copyStats{}
	for _, ref := range storageRefsFromCollections(collections) {
		source := filepath.Join(sourceStorageRoot, ref.CollectionID, ref.RecordID)
		copied, err := scanTree(source, true)
		if err != nil {
			return nil, err
		}
		stats := statsByCollection[ref.CollectionKey]
		stats.Files += copied.Files
		stats.Bytes += copied.Bytes
		statsByCollection[ref.CollectionKey] = stats
	}
	return statsByCollection, nil
}

func sumStorageStats(statsByCollection map[string]copyStats) copyStats {
	total := copyStats{}
	for _, stats := range statsByCollection {
		total.Files += stats.Files
		total.Bytes += stats.Bytes
	}
	return total
}

func scanSnapshotStorage(cwd string, collections map[string]snapshotCollection) (copyStats, error) {
	sourceStorageRoot := filepath.Join(cwd, "pb_data", "storage")
	stats := copyStats{}
	for _, ref := range storageRefsFromCollections(collections) {
		source := filepath.Join(sourceStorageRoot, ref.CollectionID, ref.RecordID)
		copied, err := scanTree(source, true)
		if err != nil {
			return stats, err
		}
		stats.Files += copied.Files
		stats.Bytes += copied.Bytes
	}
	return stats, nil
}

func storageRefsFromCollections(collections map[string]snapshotCollection) []storageRef {
	refs := []storageRef{}
	seen := map[string]bool{}
	for _, key := range snapshotOrder {
		collection := collections[key]
		spec := specsByKey[key]
		if spec.CollectionID == "" {
			continue
		}
		for _, record := range collection.Records {
			recordID := strings.TrimSpace(toString(record["id"]))
			if recordID == "" {
				continue
			}
			seenKey := spec.CollectionID + "/" + recordID
			if seen[seenKey] {
				continue
			}
			seen[seenKey] = true
			refs = append(refs, storageRef{CollectionKey: key, CollectionID: spec.CollectionID, RecordID: recordID})
		}
	}
	return refs
}

func loadSnapshot(snapshotRoot string) (snapshotManifest, map[string]snapshotCollection, error) {
	manifestPath := filepath.Join(snapshotRoot, "manifest.json")
	var manifest snapshotManifest
	if err := readJSONFile(manifestPath, &manifest); err != nil {
		return snapshotManifest{}, nil, fmt.Errorf("failed to read snapshot manifest: %w", err)
	}
	collections := map[string]snapshotCollection{}
	for _, key := range snapshotOrder {
		path := filepath.Join(snapshotRoot, "records", key+".json")
		var collection snapshotCollection
		if err := readJSONFile(path, &collection); err != nil {
			return snapshotManifest{}, nil, fmt.Errorf("failed to read snapshot records %s: %w", key, err)
		}
		collections[key] = collection
	}
	return manifest, collections, nil
}

func collectCurrentStorageRefs(db *sql.DB, websiteID string, assetsMode string, snapshotCollections map[string]snapshotCollection) ([]storageRef, error) {
	refs := storageRefsFromCollections(snapshotCollections)
	seen := map[string]bool{}
	for _, ref := range refs {
		seen[ref.CollectionID+"/"+ref.RecordID] = true
	}
	addRefs := func(key string, ids []string) {
		spec := specsByKey[key]
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			seenKey := spec.CollectionID + "/" + id
			if seen[seenKey] {
				continue
			}
			seen[seenKey] = true
			refs = append(refs, storageRef{CollectionKey: key, CollectionID: spec.CollectionID, RecordID: id})
		}
	}
	currentPageIDs, err := queryIDs(db, specsByKey["pages"], "website = ?", []interface{}{websiteID})
	if err != nil {
		return nil, err
	}
	currentBlockIDs := []string{}
	if len(currentPageIDs) > 0 {
		where, args := inClause("page", currentPageIDs)
		currentBlockIDs, err = queryIDs(db, specsByKey["blocks"], where, args)
		if err != nil {
			return nil, err
		}
	}
	currentAssetIDs := []string{}
	if assetsMode == assetModeAll {
		currentAssetIDs, err = queryIDs(db, specsByKey["assets"], "", nil)
	} else {
		currentAssetIDs, err = queryIDs(db, specsByKey["assets"], "website = ?", []interface{}{websiteID})
	}
	if err != nil {
		return nil, err
	}
	addRefs("websites", []string{websiteID})
	addRefs("pages", currentPageIDs)
	addRefs("blocks", currentBlockIDs)
	addRefs("assets", currentAssetIDs)
	return refs, nil
}

func restoreDatabase(db *sql.DB, websiteID string, assetsMode string, collections map[string]snapshotCollection) error {
	currentPageIDs, err := queryIDs(db, specsByKey["pages"], "website = ?", []interface{}{websiteID})
	if err != nil {
		return err
	}
	currentBlockIDs := []string{}
	if len(currentPageIDs) > 0 {
		where, args := inClause("page", currentPageIDs)
		currentBlockIDs, err = queryIDs(db, specsByKey["blocks"], where, args)
		if err != nil {
			return err
		}
	}
	currentAssetIDs := []string{}
	if assetsMode == assetModeAll {
		currentAssetIDs, err = queryIDs(db, specsByKey["assets"], "", nil)
	} else {
		currentAssetIDs, err = queryIDs(db, specsByKey["assets"], "website = ?", []interface{}{websiteID})
	}
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, key := range restoreDeleteOrder {
		spec := specsByKey[key]
		snapshotIDs := recordIDs(collections[key].Records)
		switch key {
		case "blocks":
			if err := deleteByIDs(tx, spec.Table, append(currentBlockIDs, snapshotIDs...)); err != nil {
				return err
			}
		case "pages":
			if _, err := tx.Exec("DELETE FROM "+quoteIdent(spec.Table)+" WHERE website = ?", websiteID); err != nil {
				return err
			}
			if err := deleteByIDs(tx, spec.Table, snapshotIDs); err != nil {
				return err
			}
		case "assets":
			if assetsMode == assetModeAll {
				if _, err := tx.Exec("DELETE FROM " + quoteIdent(spec.Table)); err != nil {
					return err
				}
			} else {
				if _, err := tx.Exec("DELETE FROM "+quoteIdent(spec.Table)+" WHERE website = ?", websiteID); err != nil {
					return err
				}
				if err := deleteByIDs(tx, spec.Table, append(currentAssetIDs, snapshotIDs...)); err != nil {
					return err
				}
			}
		case "websites":
			if _, err := tx.Exec("DELETE FROM "+quoteIdent(spec.Table)+" WHERE id = ?", websiteID); err != nil {
				return err
			}
		case "components":
			if _, err := tx.Exec("DELETE FROM " + quoteIdent(spec.Table)); err != nil {
				return err
			}
		}
	}

	for _, key := range restoreInsertOrder {
		if err := insertSnapshotRecords(tx, collections[key]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func deleteByIDs(tx *sql.Tx, table string, ids []string) error {
	ids = uniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	where, args := inClause("id", ids)
	_, err := tx.Exec("DELETE FROM "+quoteIdent(table)+" WHERE "+where, args...)
	return err
}

func insertSnapshotRecords(tx *sql.Tx, collection snapshotCollection) error {
	if len(collection.Records) == 0 {
		return nil
	}
	if len(collection.Columns) == 0 {
		return fmt.Errorf("snapshot collection %s has records but no columns", collection.Key)
	}
	placeholders := make([]string, len(collection.Columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := "INSERT INTO " + quoteIdent(collection.Table) + " (" + quoteIdentList(collection.Columns) + ") VALUES (" + strings.Join(placeholders, ",") + ")"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, record := range collection.Records {
		values := make([]interface{}, len(collection.Columns))
		for i, column := range collection.Columns {
			values[i] = record[column]
		}
		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("failed to insert %s record %s: %w", collection.Label, toString(record["id"]), err)
		}
	}
	return nil
}

func restoreStorage(cwd string, preparedStorageRoot string, refsToRemove []storageRef, collections map[string]snapshotCollection) error {
	targetStorageRoot := filepath.Join(cwd, "pb_data", "storage")
	for _, ref := range refsToRemove {
		if strings.TrimSpace(ref.CollectionID) == "" || strings.TrimSpace(ref.RecordID) == "" {
			continue
		}
		target := filepath.Join(targetStorageRoot, ref.CollectionID, ref.RecordID)
		if err := ensurePathInsideWorkspace(cwd, target); err != nil {
			return err
		}
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	return copySnapshotStorageRoot(preparedStorageRoot, targetStorageRoot)
}

func copySnapshotStorageRoot(sourceRoot string, destinationRoot string) error {
	if _, err := os.Stat(sourceRoot); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	_, err := copyTree(sourceRoot, destinationRoot)
	return err
}

func buildManifestCollections(collections map[string]snapshotCollection) []manifestCollection {
	items := []manifestCollection{}
	for _, key := range snapshotOrder {
		collection := collections[key]
		items = append(items, manifestCollection{
			Key:          collection.Key,
			Label:        collection.Label,
			CollectionID: collection.CollectionID,
			Table:        collection.Table,
			Records:      len(collection.Records),
			FileFields:   append([]string{}, collection.FileFields...),
		})
	}
	return items
}

func buildManifestStorageCollections(collections map[string]snapshotCollection, statsByCollection map[string]copyStats) []manifestStorageCollection {
	items := []manifestStorageCollection{}
	for _, key := range snapshotOrder {
		collection := collections[key]
		stats := statsByCollection[key]
		items = append(items, manifestStorageCollection{
			Key:          collection.Key,
			Label:        collection.Label,
			CollectionID: collection.CollectionID,
			Files:        stats.Files,
			Bytes:        stats.Bytes,
		})
	}
	return items
}

func buildNativeFileFieldList(collections map[string]snapshotCollection) []string {
	fields := []string{}
	for _, key := range snapshotOrder {
		collection := collections[key]
		for _, field := range collection.FileFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			fields = append(fields, collection.Label+"."+field)
		}
	}
	return fields
}

func buildSkippedCMSCollections(db *sql.DB) []manifestSkippedCollection {
	existing := currentCollectionNameMap(db)
	skipped := []manifestSkippedCollection{}
	for _, candidate := range optionalCMSCollectionCandidates {
		if existing[strings.ToLower(candidate.Name)] {
			continue
		}
		skipped = append(skipped, candidate)
	}
	return skipped
}

func buildOperationalExclusions(db *sql.DB) []string {
	existing := currentCollectionNameMap(db)
	excluded := []string{}
	for _, name := range operationalCollectionCandidates {
		if existing[strings.ToLower(name)] {
			excluded = append(excluded, name)
		}
	}
	sort.Strings(excluded)
	return excluded
}

func currentCollectionNameMap(db *sql.DB) map[string]bool {
	rows, err := db.Query("SELECT name FROM _collections")
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	result := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return result
		}
		result[strings.ToLower(strings.TrimSpace(name))] = true
	}
	return result
}

func recordIDs(records []map[string]interface{}) []string {
	ids := []string{}
	for _, record := range records {
		id := strings.TrimSpace(toString(record["id"]))
		if id != "" {
			ids = append(ids, id)
		}
	}
	return uniqueStrings(ids)
}

func countSnapshotRecords(collections map[string]snapshotCollection) int {
	total := 0
	for _, key := range snapshotOrder {
		total += len(collections[key].Records)
	}
	return total
}

func inClause(field string, ids []string) (string, []interface{}) {
	ids = uniqueStrings(ids)
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return quoteIdent(field) + " IN (" + strings.Join(placeholders, ",") + ")", args
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func uniqueStringsPreserveOrder(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func sortedKeys(values map[string]bool) []string {
	result := []string{}
	for value, ok := range values {
		if ok && strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func quoteIdent(identifier string) string {
	return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
}

func quoteIdentList(columns []string) string {
	quoted := make([]string, len(columns))
	for i, column := range columns {
		quoted[i] = quoteIdent(column)
	}
	return strings.Join(quoted, ",")
}

func ensureSnapshotDestination(snapshotRoot string, overwrite bool) error {
	info, err := os.Stat(snapshotRoot)
	if err == nil && info != nil {
		if !overwrite {
			return errors.New("snapshot destination already exists. Use --overwrite to replace it")
		}
		if err := os.RemoveAll(snapshotRoot); err != nil {
			return fmt.Errorf("failed to overwrite existing snapshot destination: %w", err)
		}
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check snapshot destination: %w", err)
	}
	return nil
}

func ensureExistingPBDataIsSafeDirectory(destinationPBData string) error {
	info, err := os.Stat(destinationPBData)
	if err != nil {
		return fmt.Errorf("destination pb_data not found: %w", err)
	}
	if !info.IsDir() {
		return errors.New("destination pb_data path is not a directory")
	}
	if err := ensurePathIsNotSymlink(destinationPBData); err != nil {
		return fmt.Errorf("refusing to restore into symlinked pb_data path: %w", err)
	}
	return nil
}

func ensureSafetyBackupDestination(path string) error {
	info, err := os.Stat(path)
	if err == nil && info != nil {
		return fmt.Errorf("safety backup destination already exists: %s", path)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect safety backup destination: %w", err)
	}
	return nil
}

func ensurePathIsNotSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("unsupported symlink path: %s", path)
	}
	return nil
}

func ensurePathInsideWorkspace(workspaceRoot string, targetPath string) error {
	rootAbs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace path: %w", err)
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}
	relativePath, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return fmt.Errorf("failed to verify target path location: %w", err)
	}
	if relativePath == "." {
		return nil
	}
	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return fmt.Errorf("refusing path outside workspace: %s", targetAbs)
	}
	return nil
}

func copyTree(sourceRoot string, destinationRoot string) (copyStats, error) {
	stats := copyStats{}
	info, err := os.Stat(sourceRoot)
	if err != nil {
		return stats, err
	}
	if !info.IsDir() {
		return stats, errors.New("source path is not a directory")
	}
	err = filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		destinationPath := filepath.Join(destinationRoot, relativePath)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsupported symlink in source tree: %s", path)
		}
		if d.IsDir() {
			entryInfo, err := d.Info()
			if err != nil {
				return err
			}
			mode := fs.FileMode(0o755)
			if entryInfo != nil {
				mode = entryInfo.Mode().Perm()
			}
			return os.MkdirAll(destinationPath, mode)
		}
		if err := copyFile(path, destinationPath); err != nil {
			return err
		}
		entryInfo, err := d.Info()
		if err != nil {
			return err
		}
		stats.Files++
		stats.Bytes += entryInfo.Size()
		return nil
	})
	if err != nil {
		return copyStats{}, err
	}
	return stats, nil
}

func scanTree(sourceRoot string, missingOK bool) (copyStats, error) {
	stats := copyStats{}
	info, err := os.Stat(sourceRoot)
	if err != nil {
		if missingOK && os.IsNotExist(err) {
			return stats, nil
		}
		if os.IsNotExist(err) {
			return stats, nil
		}
		return stats, err
	}
	if !info.IsDir() {
		return stats, errors.New("source path is not a directory")
	}
	err = filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsupported symlink in source tree: %s", path)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		stats.Files++
		stats.Bytes += info.Size()
		return nil
	})
	return stats, err
}

func copyFile(sourcePath string, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(destinationPath), err)
	}
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, sourceInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destinationPath, err)
	}
	defer destinationFile.Close()
	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", sourcePath, destinationPath, err)
	}
	return nil
}

func writeJSONFile(path string, value interface{}) error {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(bytes, '\n'), 0o644)
}

func readJSONFile(path string, out interface{}) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}

func printCreateSummary(mode string, snapshotName string, data snapshotData, sourcePath string, destinationPath string, storageStats copyStats, storageByCollection map[string]copyStats) {
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Snapshot: %s\n", snapshotName)
	fmt.Printf("Website slug: %s\n", data.Website.Slug)
	fmt.Printf("Website id: %s\n", data.Website.ID)
	fmt.Printf("Assets mode: %s\n", data.AssetsMode)
	fmt.Printf("Source: %s\n", sourcePath)
	fmt.Printf("Destination: %s\n", destinationPath)
	fmt.Println("Collections:")
	for _, item := range buildManifestCollections(data.Collections) {
		fmt.Printf("- %s: %d records", item.Label, item.Records)
		if len(item.FileFields) > 0 {
			fmt.Printf(" (file fields: %s)", strings.Join(item.FileFields, ", "))
		}
		fmt.Println()
	}
	printStorageByCollection(storageByCollection, data.Collections)
	printList("Native file fields included", buildNativeFileFieldList(data.Collections))
	printSkippedCollections(data.SkippedCollections)
	printList("Operational collections excluded", data.OperationalCollectionsExcluded)
	fmt.Printf("Storage files: %d\n", storageStats.Files)
	fmt.Printf("Storage bytes: %d\n", storageStats.Bytes)
	fmt.Println("Scope: CMS-owned records, full Assets library by default, and native file storage. Operational Leads/Newsletter/Booking records are not included.")
	fmt.Println("Reminder: dev_qa_snapshots may contain local DB/file data and must not be committed.")
}

func printRestoreDryRunSummary(snapshotName string, manifest snapshotManifest, collections map[string]snapshotCollection, destinationPath string, plannedSafetyBackupPath string, storageStats copyStats, storageByCollection map[string]copyStats, assetsMode string) {
	fmt.Println("Mode: DRY-RUN")
	fmt.Printf("Snapshot: %s\n", snapshotName)
	fmt.Printf("Website slug: %s\n", manifest.WebsiteSlug)
	fmt.Printf("Website id: %s\n", manifest.WebsiteID)
	fmt.Printf("Assets mode: %s\n", assetsMode)
	fmt.Printf("Destination pb_data path: %s\n", destinationPath)
	fmt.Printf("Planned safety backup path: %s\n", plannedSafetyBackupPath)
	fmt.Println("Collections to restore:")
	for _, item := range buildManifestCollections(collections) {
		fmt.Printf("- %s: %d records", item.Label, item.Records)
		if len(item.FileFields) > 0 {
			fmt.Printf(" (file fields: %s)", strings.Join(item.FileFields, ", "))
		}
		fmt.Println()
	}
	printStorageByCollection(storageByCollection, collections)
	printList("Native file fields included", buildNativeFileFieldList(collections))
	printSkippedCollections(manifest.SkippedCollections)
	printList("Operational collections excluded", manifest.OperationalCollectionsExcluded)
	fmt.Printf("Storage files: %d\n", storageStats.Files)
	fmt.Printf("Storage bytes: %d\n", storageStats.Bytes)
	fmt.Println("Reminder: Stop the Nuvio/PocketBase backend before restoring a CMS snapshot.")
	fmt.Println("Reminder: pb_data and dev_qa_snapshots may contain local DB/file data and must not be committed.")
}

func printRestoreWriteSummary(snapshotName string, manifest snapshotManifest, destinationPath string, safetyBackupPath string, restoreLogPath string, recordCount int, storageStats copyStats, storageByCollection map[string]copyStats, assetsMode string) {
	fmt.Println("Mode: WRITE")
	fmt.Printf("Restored snapshot: %s\n", snapshotName)
	fmt.Printf("Website slug: %s\n", manifest.WebsiteSlug)
	fmt.Printf("Website id: %s\n", manifest.WebsiteID)
	fmt.Printf("Assets mode: %s\n", assetsMode)
	fmt.Printf("Destination pb_data path: %s\n", destinationPath)
	fmt.Printf("Safety backup path: %s\n", safetyBackupPath)
	fmt.Printf("Restored records: %d\n", recordCount)
	printStorageByCollection(storageByCollection, nil)
	fmt.Printf("Storage files: %d\n", storageStats.Files)
	fmt.Printf("Storage bytes: %d\n", storageStats.Bytes)
	fmt.Printf("Restore log path: %s\n", restoreLogPath)
	fmt.Println("Reminder: pb_data and dev_qa_snapshots may contain local DB/file data and must not be committed.")
}

func printStorageByCollection(storageByCollection map[string]copyStats, collections map[string]snapshotCollection) {
	if len(storageByCollection) == 0 {
		return
	}
	fmt.Println("Storage by collection:")
	for _, key := range snapshotOrder {
		stats := storageByCollection[key]
		if stats.Files == 0 && stats.Bytes == 0 {
			continue
		}
		label := key
		if collection, ok := collections[key]; ok && strings.TrimSpace(collection.Label) != "" {
			label = collection.Label
		} else if spec, ok := specsByKey[key]; ok {
			label = spec.Label
		}
		fmt.Printf("- %s: %d files, %d bytes\n", label, stats.Files, stats.Bytes)
	}
}

func printList(label string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Println(label + ":")
	for _, value := range values {
		fmt.Printf("- %s\n", value)
	}
}

func printSkippedCollections(items []manifestSkippedCollection) {
	if len(items) == 0 {
		return
	}
	fmt.Println("CMS collection candidates not present or stored in included records:")
	for _, item := range items {
		fmt.Printf("- %s: %s\n", item.Name, item.Reason)
	}
}

func toString(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	case fmt.Stringer:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}
