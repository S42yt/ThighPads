package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// ExportTable exports a table to the default location
func ExportTable(tableID uint, exportedBy string) (string, error) {
	return ExportTableToLocation(tableID, exportedBy, DefaultLocation)
}

// ExportTableToDesktop exports a table to the desktop location
func ExportTableToDesktop(tableID uint, exportedBy string) (string, error) {
	return ExportTableToLocation(tableID, exportedBy, DesktopLocation)
}

// ExportTableToLocation exports a table to the specified location
func ExportTableToLocation(tableID uint, exportedBy string, location ExportLocation) (string, error) {
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

	// Create a safe filename without special characters
	safeFilename := sanitizeFilename(table.Name)

	// Get export paths based on location
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
			// If desktop path fails, just log it but continue with default path
			fmt.Printf("Warning: could not get desktop export path: %v\n", err)
		} else {
			paths = append(paths, desktopPath)
		}
	}

	if len(paths) == 0 {
		return "", errors.New("no valid export paths available")
	}

	// Keep track of the last path exported to
	var lastExportedPath string

	// Export to all paths
	for _, path := range paths {
		// Ensure directory exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				fmt.Printf("Warning: could not create export directory %s: %v\n", path, err)
				continue
			}
		}

		filename := filepath.Join(path, safeFilename+FileExtension)

		// Check if file exists and append a number to make it unique
		counter := 1
		originalFilename := filename
		for {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				break
			}
			filename = fmt.Sprintf("%s_%d%s", originalFilename[:len(originalFilename)-len(FileExtension)], counter, FileExtension)
			counter++
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var thighpadFile ThighpadFile
	err = json.Unmarshal(data, &thighpadFile)
	if err != nil {
		return err
	}

	if thighpadFile.Meta.Version != FileVersion {
		return errors.New("unsupported file version")
	}

	newTable := models.Table{
		Name:      thighpadFile.Table.Name,
		Author:    newAuthor,
		CreatedAt: time.Now(),
	}

	err = database.CreateTable(&newTable)
	if err != nil {
		return err
	}

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
			return err
		}
	}

	return nil
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(name string) string {
	// Replace invalid filename characters with underscores
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

	// Ensure filename doesn't start or end with spaces or periods
	resultStr := string(result)
	resultStr = filepath.Clean(resultStr)

	// If after cleaning we have an empty string, use a default name
	if resultStr == "" || resultStr == "." || resultStr == ".." {
		resultStr = "ThighPads_Export"
	}

	return resultStr
}
