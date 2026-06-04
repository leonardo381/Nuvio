package cmsqasnapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCreateOptionsDefaultsAssetsModeToAll(t *testing.T) {
	opts, err := parseCreateOptions([]string{
		"--name", "cms_test",
		"--websiteId", "abc123def456",
	})
	if err != nil {
		t.Fatalf("parseCreateOptions failed: %v", err)
	}
	if opts.AssetsMode != assetModeAll {
		t.Fatalf("expected default assets mode %q, got %q", assetModeAll, opts.AssetsMode)
	}
}

func TestParseCreateOptionsAcceptsWebsiteAssetsMode(t *testing.T) {
	opts, err := parseCreateOptions([]string{
		"--name", "cms_test",
		"--websiteId", "abc123def456",
		"--assetsMode", "website",
	})
	if err != nil {
		t.Fatalf("parseCreateOptions failed: %v", err)
	}
	if opts.AssetsMode != assetModeWebsite {
		t.Fatalf("expected assets mode %q, got %q", assetModeWebsite, opts.AssetsMode)
	}
}

func TestParseCreateOptionsRejectsInvalidAssetsMode(t *testing.T) {
	if _, err := parseCreateOptions([]string{
		"--name", "cms_test",
		"--websiteId", "abc123def456",
		"--assetsMode", "everything",
	}); err == nil {
		t.Fatal("expected invalid assetsMode to fail")
	}
}

func TestNormalizeAssetsModeForRestoreUsesLegacyWebsiteModeForMissingManifestValue(t *testing.T) {
	mode, err := normalizeAssetsModeForRestore("")
	if err != nil {
		t.Fatalf("normalizeAssetsModeForRestore failed: %v", err)
	}
	if mode != assetModeWebsite {
		t.Fatalf("expected missing manifest assets mode to use %q, got %q", assetModeWebsite, mode)
	}
}

func TestManifestHelpersIncludeFileFieldsAndStorage(t *testing.T) {
	collections := testRequiredNativeFileCollections()

	manifestCollections := buildManifestCollections(collections)
	if len(manifestCollections) != len(snapshotOrder) {
		t.Fatalf("expected %d manifest collections, got %d", len(snapshotOrder), len(manifestCollections))
	}
	if got := manifestCollections[0].FileFields; len(got) != 2 || got[0] != "logo" || got[1] != "seoImage" {
		t.Fatalf("unexpected website file fields: %#v", got)
	}
	if fields := buildNativeFileFieldList(collections); !containsAllStrings(fields, []string{
		"Websites.logo",
		"Websites.seoImage",
		"Assets.file",
		"Pages.seo_social_image",
		"Blocks.image",
	}) {
		t.Fatalf("expected 5 native file fields, got %#v", fields)
	}

	stats := map[string]copyStats{
		"websites": {Files: 2, Bytes: 100},
		"assets":   {Files: 3, Bytes: 200},
	}
	total := sumStorageStats(stats)
	if total.Files != 5 || total.Bytes != 300 {
		t.Fatalf("unexpected total storage stats: %#v", total)
	}
	storageCollections := buildManifestStorageCollections(collections, stats)
	if storageCollections[0].Files != 2 || storageCollections[2].Files != 3 {
		t.Fatalf("unexpected storage collection stats: %#v", storageCollections)
	}
}

func TestNativeFileFieldStorageIsCopiedAndCounted(t *testing.T) {
	collections := testRequiredNativeFileCollections()
	cwd := t.TempDir()
	sourceStorageRoot := filepath.Join(cwd, "pb_data", "storage")
	writeRequiredNativeFiles(t, sourceStorageRoot)

	if err := validateSourceStorage(cwd, collections); err != nil {
		t.Fatalf("validateSourceStorage failed: %v", err)
	}

	statsByCollection, err := scanSnapshotStorageByCollection(cwd, collections)
	if err != nil {
		t.Fatalf("scanSnapshotStorageByCollection failed: %v", err)
	}
	if statsByCollection["websites"].Files != 2 {
		t.Fatalf("expected Websites storage stats for logo and seoImage, got %#v", statsByCollection["websites"])
	}
	if statsByCollection["assets"].Files != 1 {
		t.Fatalf("expected Assets.file storage stats, got %#v", statsByCollection["assets"])
	}
	if statsByCollection["pages"].Files != 1 {
		t.Fatalf("expected Pages.seo_social_image storage stats, got %#v", statsByCollection["pages"])
	}
	if statsByCollection["blocks"].Files != 1 {
		t.Fatalf("expected Blocks.image storage stats, got %#v", statsByCollection["blocks"])
	}
	total := sumStorageStats(statsByCollection)
	if total.Files != 5 {
		t.Fatalf("expected 5 copied native file storage files, got %#v", total)
	}

	destinationStorageRoot := filepath.Join(t.TempDir(), "snapshot", "storage")
	copied, err := copySnapshotStorage(cwd, destinationStorageRoot, collections)
	if err != nil {
		t.Fatalf("copySnapshotStorage failed: %v", err)
	}
	if copied.Files != 5 {
		t.Fatalf("expected 5 copied files, got %#v", copied)
	}

	for _, required := range testRequiredNativeFilePaths(destinationStorageRoot) {
		if _, err := os.Stat(required); err != nil {
			t.Fatalf("expected copied storage file %s: %v", required, err)
		}
	}
}

func TestValidateSnapshotStorageFailsWhenRequiredNativeFileIsMissing(t *testing.T) {
	collections := testRequiredNativeFileCollections()
	snapshotRoot := t.TempDir()
	snapshotStorageRoot := filepath.Join(snapshotRoot, "storage")
	writeRequiredNativeFiles(t, snapshotStorageRoot)

	missingPageImage := filepath.Join(snapshotStorageRoot, specsByKey["pages"].CollectionID, "page1", "page-share.png")
	if err := os.Remove(missingPageImage); err != nil {
		t.Fatalf("failed to remove required snapshot file: %v", err)
	}

	err := validateSnapshotStorage(snapshotRoot, collections)
	if err == nil {
		t.Fatal("expected validateSnapshotStorage to fail for missing native file")
	}
	if !strings.Contains(err.Error(), "missing storage file") || !strings.Contains(err.Error(), "seo_social_image") {
		t.Fatalf("expected missing seo_social_image storage error, got %v", err)
	}
}

func testRequiredNativeFileCollections() map[string]snapshotCollection {
	return map[string]snapshotCollection{
		"websites": {
			Key:          "websites",
			Label:        "Websites",
			CollectionID: specsByKey["websites"].CollectionID,
			Table:        specsByKey["websites"].Table,
			FileFields:   []string{"logo", "seoImage"},
			Records: []map[string]interface{}{{
				"id":       "site1",
				"logo":     "logo.png",
				"seoImage": "seo.jpg",
			}},
		},
		"components": {
			Key:          "components",
			Label:        "Components",
			CollectionID: specsByKey["components"].CollectionID,
			Table:        specsByKey["components"].Table,
			Records:      []map[string]interface{}{},
		},
		"assets": {
			Key:          "assets",
			Label:        "Assets",
			CollectionID: specsByKey["assets"].CollectionID,
			Table:        specsByKey["assets"].Table,
			FileFields:   []string{"file"},
			Records: []map[string]interface{}{{
				"id":   "asset1",
				"file": "asset.webp",
			}},
		},
		"pages": {
			Key:          "pages",
			Label:        "Pages",
			CollectionID: specsByKey["pages"].CollectionID,
			Table:        specsByKey["pages"].Table,
			FileFields:   []string{"seo_social_image"},
			Records: []map[string]interface{}{{
				"id":               "page1",
				"seo_social_image": "page-share.png",
			}},
		},
		"blocks": {
			Key:          "blocks",
			Label:        "Blocks",
			CollectionID: specsByKey["blocks"].CollectionID,
			Table:        specsByKey["blocks"].Table,
			FileFields:   []string{"image"},
			Records: []map[string]interface{}{{
				"id":    "block1",
				"image": "block.gif",
			}},
		},
	}
}

func writeRequiredNativeFiles(t *testing.T, storageRoot string) {
	t.Helper()
	files := map[string]string{
		filepath.Join(storageRoot, specsByKey["websites"].CollectionID, "site1", "logo.png"):    "logo",
		filepath.Join(storageRoot, specsByKey["websites"].CollectionID, "site1", "seo.jpg"):     "seo",
		filepath.Join(storageRoot, specsByKey["assets"].CollectionID, "asset1", "asset.webp"):   "asset",
		filepath.Join(storageRoot, specsByKey["pages"].CollectionID, "page1", "page-share.png"): "page",
		filepath.Join(storageRoot, specsByKey["blocks"].CollectionID, "block1", "block.gif"):    "block",
	}
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to prepare storage dir for %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write storage file %s: %v", path, err)
		}
	}
}

func testRequiredNativeFilePaths(storageRoot string) []string {
	return []string{
		filepath.Join(storageRoot, specsByKey["websites"].CollectionID, "site1", "logo.png"),
		filepath.Join(storageRoot, specsByKey["websites"].CollectionID, "site1", "seo.jpg"),
		filepath.Join(storageRoot, specsByKey["assets"].CollectionID, "asset1", "asset.webp"),
		filepath.Join(storageRoot, specsByKey["pages"].CollectionID, "page1", "page-share.png"),
		filepath.Join(storageRoot, specsByKey["blocks"].CollectionID, "block1", "block.gif"),
	}
}

func containsAllStrings(values []string, expected []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range expected {
		if !seen[value] {
			return false
		}
	}
	return true
}
