package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/models"
)

type ThighpadFile struct {
	Table   models.Table     `json:"table"`
	Entries []models.Entry   `json:"entries"`
	Meta    ThighpadFileMeta `json:"meta"`
}

type ThighpadFileMeta struct {
	ExportedAt time.Time `json:"exportedAt"`
	ExportedBy string    `json:"exportedBy"`
	Version    string    `json:"version"`
}

type ExportLocation int

const (
	DefaultLocation ExportLocation = iota
	DesktopLocation
	BothLocations
)

const (
	FileExtension = ".thighpad"
	FileVersion   = "1.0"
)

func ExportTable(tableID uint, exportedBy string, filename string) (string, error) {
	return ExportTableToLocation(tableID, exportedBy, DefaultLocation, filename)
}

func ExportTableToDesktop(tableID uint, exportedBy string, filename string) (string, error) {
	return ExportTableToLocation(tableID, exportedBy, DesktopLocation, filename)
}

func ExportTableToLocation(tableID uint, exportedBy string, location ExportLocation, filename string) (string, error) {
	table, err := database.GetTableWithEntries(tableID)
	if err != nil {
		return "", err
	}

	thighpadFile := ThighpadFile{
		Table:   table,
		Entries: table.Entries,
		Meta: ThighpadFileMeta{
			ExportedAt: time.Now(),
			ExportedBy: exportedBy,
			Version:    FileVersion,
		},
	}

	data, err := json.MarshalIndent(thighpadFile, "", "  ")
	if err != nil {
		return "", err
	}

	safeFilename := filename
	if safeFilename == "" {
		safeFilename = sanitizeFilename(table.Name)
	} else {
		safeFilename = sanitizeFilename(safeFilename)
	}

	if !strings.HasSuffix(strings.ToLower(safeFilename), strings.ToLower(FileExtension)) {
		safeFilename += FileExtension
	}

	paths := []string{}

	if location == DefaultLocation || location == BothLocations {
		defaultPath, err := config.GetExportPath()
		if err != nil {
			return "", err
		}
		paths = append(paths, defaultPath)
	}

	if location == DesktopLocation || location == BothLocations {
		desktopPath, err := config.GetDesktopExportPath()
		if err != nil {
			fmt.Printf("Warning: could not get desktop export path: %v\n", err)
		} else {
			paths = append(paths, desktopPath)
		}
	}

	if len(paths) == 0 {
		return "", errors.New("no valid export paths available")
	}

	var lastExportedPath string

	for _, path := range paths {

		path = config.NormalizePath(path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				fmt.Printf("Warning: could not create export directory %s: %v\n", path, err)
				continue
			}
		}

		filename := filepath.Join(path, safeFilename)

		counter := 1
		originalFilename := filename
		for {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				break
			}
			filename = fmt.Sprintf("%s_%d%s",
				originalFilename[:len(originalFilename)-len(FileExtension)],
				counter,
				FileExtension)
			counter++

			if counter > 1000 {
				return "", errors.New("failed to find a unique filename after 1000 attempts")
			}
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			fmt.Printf("Warning: could not write to export file %s: %v\n", filename, err)
			continue
		}

		lastExportedPath = filename
	}

	if lastExportedPath == "" {
		return "", errors.New("failed to export to any location")
	}

	return lastExportedPath, nil
}

func ImportFile(filePath string, newAuthor string) error {

	normalizedPath := config.NormalizePath(filePath)

	if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", normalizedPath)
	}

	if !strings.HasSuffix(strings.ToLower(normalizedPath), strings.ToLower(FileExtension)) {
		return fmt.Errorf("file must have %s extension", FileExtension)
	}

	data, err := os.ReadFile(normalizedPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var thighpadFile ThighpadFile
	err = json.Unmarshal(data, &thighpadFile)
	if err != nil {
		return fmt.Errorf("invalid file format: %w", err)
	}

	if thighpadFile.Table.Name == "" {
		return errors.New("invalid file: missing table name")
	}

	if thighpadFile.Meta.Version == "" {

		thighpadFile.Meta.Version = FileVersion
	} else if thighpadFile.Meta.Version != FileVersion {
		return fmt.Errorf("unsupported file version: %s (expected %s)",
			thighpadFile.Meta.Version, FileVersion)
	}

	newTable := models.Table{
		Name:      thighpadFile.Table.Name,
		Author:    newAuthor,
		CreatedAt: time.Now(),
	}

	err = database.CreateTable(&newTable)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	entriesCount := 0
	for _, entry := range thighpadFile.Entries {
		newEntry := models.Entry{
			TableID:   newTable.ID,
			Title:     entry.Title,
			Tags:      entry.Tags,
			Content:   entry.Content,
			CreatedAt: time.Now(),
		}

		err = database.CreateEntry(&newEntry)
		if err != nil {

			fmt.Printf("Warning: failed to import entry '%s': %v\n", entry.Title, err)
			continue
		}
		entriesCount++
	}

	if entriesCount == 0 && len(thighpadFile.Entries) > 0 {

		return errors.New("failed to import any entries")
	}

	return nil
}

func sanitizeFilename(name string) string {

	invalidChars := []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*'}
	result := []rune(name)

	for i, ch := range result {
		for _, invalid := range invalidChars {
			if ch == invalid {
				result[i] = '_'
				break
			}
		}
	}

	resultStr := string(result)
	resultStr = filepath.Clean(resultStr)

	if resultStr == "" || resultStr == "." || resultStr == ".." {
		resultStr = "ThighPads_Export"
	}

	return resultStr
}
