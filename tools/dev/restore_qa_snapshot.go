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
	qaRestoreConfirmToken         = "RESTORE_QA_SNAPSHOT"
	qaRestoreOptInEnv             = "NUVIO_ALLOW_DEV_RESET"
	qaRestoreNote                 = "Local dev QA restore. Do not use in production."
	qaRestoreBackupNote           = "Local dev QA pre-restore safety backup. Do not use in production."
	qaRestoreBackendStoppedNotice = "Stop the Nuvio/PocketBase backend before restoring a snapshot."
)

var qaRestoreSnapshotNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type qaRestoreOptions struct {
	Name           string
	Confirm        string
	BackendStopped bool
}

type qaRestoreManifest struct {
	SnapshotName    string `json:"snapshotName"`
	CreatedAt       string `json:"createdAt"`
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
	CopiedFiles     int64  `json:"copiedFiles"`
	CopiedBytes     int64  `json:"copiedBytes"`
	Note            string `json:"note"`
}

type qaRestoreLog struct {
	RestoredSnapshotName string `json:"restoredSnapshotName"`
	RestoredAt           string `json:"restoredAt"`
	SourceSnapshotPath   string `json:"sourceSnapshotPath"`
	DestinationPath      string `json:"destinationPath"`
	SafetyBackupPath     string `json:"safetyBackupPath"`
	CopiedFiles          int64  `json:"copiedFiles"`
	CopiedBytes          int64  `json:"copiedBytes"`
	Note                 string `json:"note"`
}

type qaRestoreCopyStats struct {
	Files int64
	Bytes int64
}

type qaRestoreSnapshotValidation struct {
	HasDataDB   bool
	HasManifest bool
}

func main() {
	if err := runQARestoreTool(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func runQARestoreTool() error {
	opts, err := parseQARestoreOptions()
	if err != nil {
		return err
	}

	if strings.TrimSpace(os.Getenv(qaRestoreOptInEnv)) != "1" {
		return fmt.Errorf("%s must be set to 1", qaRestoreOptInEnv)
	}

	if !qaRestoreSnapshotNamePattern.MatchString(opts.Name) {
		return errors.New("invalid --name. Use only letters, numbers, dash, and underscore")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve current workspace: %w", err)
	}

	if err := ensureQASnapshotWorkspace(cwd); err != nil {
		return err
	}

	snapshotsRoot := filepath.Join(cwd, "dev_qa_snapshots")
	sourceSnapshotRoot := filepath.Join(snapshotsRoot, opts.Name)
	sourceSnapshotPBData := filepath.Join(sourceSnapshotRoot, "pb_data")
	sourceSnapshotManifest := filepath.Join(sourceSnapshotRoot, "manifest.json")
	destinationPBData := filepath.Join(cwd, "pb_data")

	utcNow := time.Now().UTC()
	timestamp := utcNow.Format("20060102T150405Z")
	safetyBackupName := fmt.Sprintf("pre_restore_backup_%s", timestamp)
	safetyBackupRoot := filepath.Join(snapshotsRoot, safetyBackupName)
	safetyBackupPBData := filepath.Join(safetyBackupRoot, "pb_data")
	safetyBackupManifestPath := filepath.Join(safetyBackupRoot, "manifest.json")
	tempRestoreRoot := filepath.Join(snapshotsRoot, fmt.Sprintf("restore_tmp_%s", timestamp))
	tempRestorePBData := filepath.Join(tempRestoreRoot, "pb_data")
	restoreLogsRoot := filepath.Join(snapshotsRoot, "restore_logs")
	restoreLogPath := filepath.Join(restoreLogsRoot, fmt.Sprintf("restore_%s.json", timestamp))

	if err := ensurePathInsideWorkspace(cwd, sourceSnapshotRoot); err != nil {
		return err
	}
	if err := ensurePathInsideWorkspace(cwd, destinationPBData); err != nil {
		return err
	}
	if err := ensurePathInsideWorkspace(cwd, safetyBackupRoot); err != nil {
		return err
	}
	if err := ensurePathInsideWorkspace(cwd, tempRestoreRoot); err != nil {
		return err
	}
	if err := ensurePathInsideWorkspace(cwd, restoreLogPath); err != nil {
		return err
	}

	writeMode := strings.TrimSpace(opts.Confirm) == qaRestoreConfirmToken
	if strings.TrimSpace(opts.Confirm) != "" && !writeMode {
		return fmt.Errorf("invalid --confirm token; expected %s", qaRestoreConfirmToken)
	}

	if writeMode {
		if !opts.BackendStopped {
			return errors.New(qaRestoreBackendStoppedNotice)
		}
		if err := ensureLocalDevWorkspacePath(cwd); err != nil {
			return err
		}
	}

	validation, err := validateQARestoreSnapshot(sourceSnapshotPBData, sourceSnapshotManifest)
	if err != nil {
		return err
	}

	snapshotStats, err := scanQASnapshotSource(sourceSnapshotPBData)
	if err != nil {
		return err
	}

	if !writeMode {
		printQARestoreDryRunSummary(
			opts.Name,
			sourceSnapshotPBData,
			destinationPBData,
			safetyBackupPBData,
			validation,
			snapshotStats,
		)
		return nil
	}

	if err := ensureExistingPBDataIsSafeDirectory(destinationPBData); err != nil {
		return err
	}

	if err := ensureSafetyBackupDestination(safetyBackupRoot); err != nil {
		return err
	}

	if err := os.MkdirAll(safetyBackupPBData, 0o755); err != nil {
		return fmt.Errorf("failed to prepare safety backup directory: %w", err)
	}

	backupStats, err := copyQASnapshotTree(destinationPBData, safetyBackupPBData)
	if err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}

	backupManifest := qaRestoreManifest{
		SnapshotName:    safetyBackupName,
		CreatedAt:       utcNow.Format(time.RFC3339),
		SourcePath:      destinationPBData,
		DestinationPath: safetyBackupPBData,
		CopiedFiles:     backupStats.Files,
		CopiedBytes:     backupStats.Bytes,
		Note:            qaRestoreBackupNote,
	}
	if err := writeQARestoreManifest(safetyBackupManifestPath, backupManifest); err != nil {
		return fmt.Errorf("failed to write safety backup manifest: %w", err)
	}

	if err := prepareQARestoreTempRoot(tempRestoreRoot, tempRestorePBData); err != nil {
		return err
	}

	restoreStats, err := copyQASnapshotTree(sourceSnapshotPBData, tempRestorePBData)
	if err != nil {
		return fmt.Errorf("failed to prepare restore copy. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	if err := os.RemoveAll(destinationPBData); err != nil {
		return fmt.Errorf("failed to replace pb_data. Safety backup preserved at %s: %w", safetyBackupRoot, err)
	}

	if err := os.Rename(tempRestorePBData, destinationPBData); err != nil {
		return fmt.Errorf(
			"failed to move restored pb_data into place. Safety backup preserved at %s. Prepared restore copy at %s: %w",
			safetyBackupRoot,
			tempRestorePBData,
			err,
		)
	}

	if err := os.RemoveAll(tempRestoreRoot); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(
			"restore succeeded but failed to clean temporary restore directory %s. Safety backup preserved at %s: %w",
			tempRestoreRoot,
			safetyBackupRoot,
			err,
		)
	}

	restoreLog := qaRestoreLog{
		RestoredSnapshotName: opts.Name,
		RestoredAt:           time.Now().UTC().Format(time.RFC3339),
		SourceSnapshotPath:   sourceSnapshotPBData,
		DestinationPath:      destinationPBData,
		SafetyBackupPath:     safetyBackupPBData,
		CopiedFiles:          restoreStats.Files,
		CopiedBytes:          restoreStats.Bytes,
		Note:                 qaRestoreNote,
	}

	if err := os.MkdirAll(restoreLogsRoot, 0o755); err != nil {
		return fmt.Errorf("restore succeeded but failed to prepare restore logs directory: %w", err)
	}
	if err := writeQARestoreLog(restoreLogPath, restoreLog); err != nil {
		return fmt.Errorf("restore succeeded but failed to write restore log: %w", err)
	}

	printQARestoreWriteSummary(
		opts.Name,
		destinationPBData,
		safetyBackupPBData,
		restoreStats,
		restoreLogPath,
	)

	return nil
}

func parseQARestoreOptions() (qaRestoreOptions, error) {
	var opts qaRestoreOptions
	flag.StringVar(&opts.Name, "name", "", "Snapshot name")
	flag.StringVar(&opts.Confirm, "confirm", "", "Confirmation token for write mode")
	flag.BoolVar(&opts.BackendStopped, "backendStopped", false, "Confirm backend is stopped before restoring")
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

func ensureLocalDevWorkspacePath(cwd string) error {
	lowerPath := strings.ToLower(filepath.Clean(cwd))
	if strings.HasPrefix(lowerPath, `\\`) {
		return errors.New("write mode is allowed only for local/dev workspace paths")
	}

	localMarkers := []string{
		`\users\`,
		`/users/`,
		`\home\`,
		`/home/`,
		`\tmp\`,
		`/tmp/`,
		`\documents\`,
		`/documents/`,
	}

	for _, marker := range localMarkers {
		if strings.Contains(lowerPath, marker) {
			return nil
		}
	}

	return errors.New("write mode is allowed only for local/dev workspace paths")
}

func validateQARestoreSnapshot(snapshotPBDataPath string, snapshotManifestPath string) (qaRestoreSnapshotValidation, error) {
	snapshotInfo, err := os.Stat(snapshotPBDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return qaRestoreSnapshotValidation{}, fmt.Errorf("snapshot not found: %s", snapshotPBDataPath)
		}
		return qaRestoreSnapshotValidation{}, fmt.Errorf("failed to inspect snapshot pb_data: %w", err)
	}
	if !snapshotInfo.IsDir() {
		return qaRestoreSnapshotValidation{}, errors.New("snapshot pb_data path is not a directory")
	}

	if err := ensurePathIsNotSymlink(snapshotPBDataPath); err != nil {
		return qaRestoreSnapshotValidation{}, err
	}

	dataDBPath := filepath.Join(snapshotPBDataPath, "data.db")
	dataInfo, err := os.Stat(dataDBPath)
	if err != nil {
		if os.IsNotExist(err) {
			return qaRestoreSnapshotValidation{}, errors.New("snapshot is missing required file: pb_data/data.db")
		}
		return qaRestoreSnapshotValidation{}, fmt.Errorf("failed to inspect snapshot data.db: %w", err)
	}
	if dataInfo.IsDir() {
		return qaRestoreSnapshotValidation{}, errors.New("snapshot data.db path is not a file")
	}

	manifestInfo, err := os.Stat(snapshotManifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return qaRestoreSnapshotValidation{}, errors.New("snapshot is missing required file: manifest.json")
		}
		return qaRestoreSnapshotValidation{}, fmt.Errorf("failed to inspect snapshot manifest: %w", err)
	}
	if manifestInfo.IsDir() {
		return qaRestoreSnapshotValidation{}, errors.New("snapshot manifest path is not a file")
	}

	return qaRestoreSnapshotValidation{
		HasDataDB:   true,
		HasManifest: true,
	}, nil
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

func ensureExistingPBDataIsSafeDirectory(destinationPBData string) error {
	destinationInfo, err := os.Stat(destinationPBData)
	if err != nil {
		return fmt.Errorf("destination pb_data not found: %w", err)
	}
	if !destinationInfo.IsDir() {
		return errors.New("destination pb_data path is not a directory")
	}
	if err := ensurePathIsNotSymlink(destinationPBData); err != nil {
		return fmt.Errorf("refusing to restore into symlinked pb_data path: %w", err)
	}
	return nil
}

func ensureSafetyBackupDestination(safetyBackupRoot string) error {
	info, err := os.Stat(safetyBackupRoot)
	if err == nil && info != nil {
		return fmt.Errorf("safety backup destination already exists: %s", safetyBackupRoot)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect safety backup destination: %w", err)
	}
	return nil
}

func prepareQARestoreTempRoot(tempRestoreRoot string, tempRestorePBData string) error {
	if info, err := os.Stat(tempRestoreRoot); err == nil && info != nil {
		return fmt.Errorf("temporary restore directory already exists: %s", tempRestoreRoot)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect temporary restore directory: %w", err)
	}

	if err := os.MkdirAll(tempRestorePBData, 0o755); err != nil {
		return fmt.Errorf("failed to prepare temporary restore directory: %w", err)
	}
	return nil
}

func scanQASnapshotSource(sourcePBData string) (qaRestoreCopyStats, error) {
	info, err := os.Stat(sourcePBData)
	if err != nil {
		return qaRestoreCopyStats{}, fmt.Errorf("snapshot pb_data not found: %w", err)
	}
	if !info.IsDir() {
		return qaRestoreCopyStats{}, errors.New("snapshot pb_data path is not a directory")
	}

	stats := qaRestoreCopyStats{}
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
		return qaRestoreCopyStats{}, fmt.Errorf("failed to scan snapshot pb_data: %w", err)
	}

	return stats, nil
}

func copyQASnapshotTree(sourceRoot string, destinationRoot string) (qaRestoreCopyStats, error) {
	stats := qaRestoreCopyStats{}

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
		return qaRestoreCopyStats{}, fmt.Errorf("failed to copy source tree: %w", err)
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

func writeQARestoreManifest(path string, manifest qaRestoreManifest) error {
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode manifest: %w", err)
	}

	if err := os.WriteFile(path, append(manifestBytes, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

func writeQARestoreLog(path string, restoreLog qaRestoreLog) error {
	restoreLogBytes, err := json.MarshalIndent(restoreLog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode restore log: %w", err)
	}

	if err := os.WriteFile(path, append(restoreLogBytes, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write restore log: %w", err)
	}

	return nil
}

func printQARestoreDryRunSummary(
	snapshotName string,
	sourcePath string,
	destinationPath string,
	plannedSafetyBackupPath string,
	validation qaRestoreSnapshotValidation,
	stats qaRestoreCopyStats,
) {
	fmt.Println("Mode: DRY-RUN")
	fmt.Printf("Snapshot: %s\n", snapshotName)
	fmt.Printf("Source snapshot path: %s\n", sourcePath)
	fmt.Printf("Destination pb_data path: %s\n", destinationPath)
	fmt.Printf("Snapshot data.db exists: %t\n", validation.HasDataDB)
	fmt.Printf("Snapshot manifest exists: %t\n", validation.HasManifest)
	fmt.Printf("Planned safety backup path: %s\n", plannedSafetyBackupPath)
	fmt.Printf("Planned restore files: %d\n", stats.Files)
	fmt.Printf("Planned restore bytes: %d\n", stats.Bytes)
	fmt.Printf("Reminder: %s\n", qaRestoreBackendStoppedNotice)
	fmt.Println("Reminder: pb_data and dev_qa_snapshots may contain local DB data and must not be committed.")
}

func printQARestoreWriteSummary(
	snapshotName string,
	destinationPath string,
	safetyBackupPath string,
	stats qaRestoreCopyStats,
	restoreLogPath string,
) {
	fmt.Println("Mode: WRITE")
	fmt.Printf("Restored snapshot: %s\n", snapshotName)
	fmt.Printf("Destination pb_data path: %s\n", destinationPath)
	fmt.Printf("Safety backup path: %s\n", safetyBackupPath)
	fmt.Printf("Copied files: %d\n", stats.Files)
	fmt.Printf("Copied bytes: %d\n", stats.Bytes)
	fmt.Printf("Restore log path: %s\n", restoreLogPath)
	fmt.Println("Reminder: pb_data and dev_qa_snapshots may contain local DB data and must not be committed.")
}
