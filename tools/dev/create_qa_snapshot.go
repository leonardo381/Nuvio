package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	qaSnapshotConfirmToken = "CREATE_QA_SNAPSHOT"
	qaSnapshotOptInEnv     = "NUVIO_ALLOW_DEV_RESET"
	qaSnapshotNote         = "Local dev QA snapshot. Do not use in production."
)

var qaSnapshotNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type qaSnapshotOptions struct {
	Name           string
	Confirm        string
	Overwrite      bool
	BackendStopped bool
}

type qaSnapshotManifest struct {
	SnapshotName    string `json:"snapshotName"`
	CreatedAt       string `json:"createdAt"`
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
	CopiedFiles     int64  `json:"copiedFiles"`
	CopiedBytes     int64  `json:"copiedBytes"`
	Note            string `json:"note"`
}

type qaCopyStats struct {
	Files int64
	Bytes int64
}

func main() {
	if err := runQASnapshotTool(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func runQASnapshotTool() error {
	opts, err := parseQASnapshotOptions()
	if err != nil {
		return err
	}

	if strings.TrimSpace(os.Getenv(qaSnapshotOptInEnv)) != "1" {
		return fmt.Errorf("%s must be set to 1", qaSnapshotOptInEnv)
	}

	if !qaSnapshotNamePattern.MatchString(opts.Name) {
		return errors.New("invalid --name. Use only letters, numbers, dash, and underscore")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve current workspace: %w", err)
	}

	if err := ensureQASnapshotWorkspace(cwd); err != nil {
		return err
	}

	sourcePBData := filepath.Join(cwd, "pb_data")
	snapshotsRoot := filepath.Join(cwd, "dev_qa_snapshots")
	snapshotRoot := filepath.Join(snapshotsRoot, opts.Name)
	destinationPBData := filepath.Join(snapshotRoot, "pb_data")
	manifestPath := filepath.Join(snapshotRoot, "manifest.json")

	writeMode := strings.TrimSpace(opts.Confirm) == qaSnapshotConfirmToken
	if strings.TrimSpace(opts.Confirm) != "" && !writeMode {
		return fmt.Errorf("invalid --confirm token; expected %s", qaSnapshotConfirmToken)
	}

	if writeMode && !opts.BackendStopped {
		return errors.New("Stop the Nuvio/PocketBase backend before creating a snapshot.")
	}

	stats, err := scanQASnapshotSource(sourcePBData)
	if err != nil {
		return err
	}

	if !writeMode {
		printQASnapshotSummary(
			"DRY-RUN",
			opts.Name,
			sourcePBData,
			destinationPBData,
			stats,
		)
		return nil
	}

	if err := ensureSnapshotDestination(snapshotRoot, opts.Overwrite); err != nil {
		return err
	}

	if err := os.MkdirAll(destinationPBData, 0o755); err != nil {
		return fmt.Errorf("failed to prepare destination directory: %w", err)
	}

	copyStats, err := copyQASnapshotTree(sourcePBData, destinationPBData)
	if err != nil {
		return err
	}

	manifest := qaSnapshotManifest{
		SnapshotName:    opts.Name,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		SourcePath:      sourcePBData,
		DestinationPath: destinationPBData,
		CopiedFiles:     copyStats.Files,
		CopiedBytes:     copyStats.Bytes,
		Note:            qaSnapshotNote,
	}

	if err := writeQASnapshotManifest(manifestPath, manifest); err != nil {
		return err
	}

	printQASnapshotSummary(
		"WRITE",
		opts.Name,
		sourcePBData,
		destinationPBData,
		copyStats,
	)

	return nil
}

func parseQASnapshotOptions() (qaSnapshotOptions, error) {
	var opts qaSnapshotOptions
	flag.StringVar(&opts.Name, "name", "", "Snapshot name")
	flag.StringVar(&opts.Confirm, "confirm", "", "Confirmation token for write mode")
	flag.BoolVar(&opts.Overwrite, "overwrite", false, "Overwrite existing snapshot if present")
	flag.BoolVar(&opts.BackendStopped, "backendStopped", false, "Confirm backend is stopped before copying")
	flag.Parse()

	opts.Name = strings.TrimSpace(opts.Name)
	if opts.Name == "" {
		return opts, errors.New("--name is required")
	}

	return opts, nil
}

func ensureQASnapshotWorkspace(cwd string) error {
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

func scanQASnapshotSource(sourcePBData string) (qaCopyStats, error) {
	info, err := os.Stat(sourcePBData)
	if err != nil {
		return qaCopyStats{}, fmt.Errorf("source pb_data not found: %w", err)
	}
	if !info.IsDir() {
		return qaCopyStats{}, errors.New("source pb_data path is not a directory")
	}

	stats := qaCopyStats{}
	err = filepath.WalkDir(sourcePBData, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsupported symlink in source tree: %s", path)
		}
		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		stats.Files++
		stats.Bytes += fileInfo.Size()
		return nil
	})
	if err != nil {
		return qaCopyStats{}, fmt.Errorf("failed to scan source pb_data: %w", err)
	}

	return stats, nil
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

func copyQASnapshotTree(sourceRoot string, destinationRoot string) (qaCopyStats, error) {
	stats := qaCopyStats{}

	err := filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, walkErr error) error {
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
			info, err := d.Info()
			if err != nil {
				return err
			}
			mode := fs.FileMode(0o755)
			if info != nil {
				mode = info.Mode().Perm()
			}
			return os.MkdirAll(destinationPath, mode)
		}

		if err := copyQASnapshotFile(path, destinationPath); err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		stats.Files++
		stats.Bytes += info.Size()
		return nil
	})
	if err != nil {
		return qaCopyStats{}, fmt.Errorf("failed to copy snapshot source tree: %w", err)
	}

	return stats, nil
}

func copyQASnapshotFile(sourcePath string, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}

	destinationDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destinationDir, err)
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

func writeQASnapshotManifest(path string, manifest qaSnapshotManifest) error {
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode snapshot manifest: %w", err)
	}

	if err := os.WriteFile(path, append(manifestBytes, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot manifest: %w", err)
	}

	return nil
}

func printQASnapshotSummary(
	mode string,
	snapshotName string,
	sourcePath string,
	destinationPath string,
	stats qaCopyStats,
) {
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Snapshot: %s\n", snapshotName)
	fmt.Printf("Source: %s\n", sourcePath)
	fmt.Printf("Destination: %s\n", destinationPath)
	fmt.Printf("Files: %d\n", stats.Files)
	fmt.Printf("Bytes: %d\n", stats.Bytes)
	fmt.Println("Reminder: dev_qa_snapshots may contain local DB data and must not be committed.")
}
