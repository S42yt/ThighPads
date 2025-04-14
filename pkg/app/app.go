package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/data"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/models"
)

// App represents the main application state
type App struct {
	Config  *config.Config
	Storage *data.Storage
	DB      *database.DB
}

// New creates a new App instance
func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	app := &App{
		Config: cfg,
	}

	// Initialize database if enabled
	if cfg.UseDatabase {
		db, err := database.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		app.DB = db

		// We still initialize storage for compatibility and migration
		storage, err := data.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %w", err)
		}
		app.Storage = storage

		// Optional: Migrate any existing tables from files to DB
		// This can be removed once migration is complete
		if err := app.migrateFromFilesToDB(); err != nil {
			return nil, fmt.Errorf("migration error: %w", err)
		}
	} else {
		// Initialize file storage if database is disabled
		storage, err := data.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %w", err)
		}
		app.Storage = storage
	}

	return app, nil
}

// migrateFromFilesToDB migrates tables from files to the database
func (a *App) migrateFromFilesToDB() error {
	// Skip if we have database tables already
	tables, err := a.DB.GetAllTables()
	if err != nil {
		return err
	}
	if len(tables) > 0 {
		return nil // Already have tables in DB
	}

	// Get tables from file storage
	fileTables := a.Storage.ListTableNames()
	for _, name := range fileTables {
		table, ok := a.Storage.GetTable(name)
		if !ok {
			continue
		}

		// Import the table into the database
		dbTable, err := a.DB.CreateTable(table.Name, table.Author)
		if err != nil {
			return err
		}

		// Add entries
		for _, entry := range table.GetEntrySlice() {
			// Convert content to description in new DB schema
			if entry.Description == "" && entry.Content != "" {
				entry.Description = entry.Content
			}

			// Convert tags to fields
			if entry.Fields == nil {
				entry.Fields = make(map[string]string)
			}

			if len(entry.Tags) > 0 {
				entry.Fields["tags"] = strings.Join(entry.Tags, ", ")
			}

			if entry.Content != "" {
				entry.Fields["content"] = entry.Content
			}

			// Add the entry to the database
			if err := a.DB.AddEntry(dbTable.Name, entry); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateTable creates a new table
func (a *App) CreateTable(name string) (*models.Table, error) {
	// Validate table name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("table name cannot be empty")
	}

	// Only allow alphanumeric characters and underscores
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return nil, errors.New("table name can only contain letters, numbers, and underscores")
		}
	}

	if a.Config.UseDatabase {
		return a.DB.CreateTable(name, a.Config.Username)
	}

	return a.Storage.CreateTable(name)
}

// GetTable returns a table by name
func (a *App) GetTable(name string) (*models.Table, error) {
	if a.Config.UseDatabase {
		return a.DB.GetTable(name)
	}

	table, ok := a.Storage.GetTable(name)
	if !ok {
		return nil, errors.New("table not found")
	}
	return table, nil
}

// GetAllTables returns all tables
func (a *App) GetAllTables() []*models.Table {
	if a.Config.UseDatabase {
		tables, err := a.DB.GetAllTables()
		if err != nil {
			return []*models.Table{}
		}
		return tables
	}

	names := a.Storage.ListTableNames()
	tables := make([]*models.Table, 0, len(names))
	for _, name := range names {
		if table, ok := a.Storage.GetTable(name); ok {
			tables = append(tables, table)
		}
	}
	return tables
}

// DeleteTable deletes a table
func (a *App) DeleteTable(name string) error {
	if a.Config.UseDatabase {
		return a.DB.DeleteTable(name)
	}
	return a.Storage.DeleteTable(name)
}

// AddEntry adds an entry to a table
func (a *App) AddEntry(tableName string, title, content string, tags []string) error {
	if title == "" {
		return errors.New("entry title cannot be empty")
	}

	entry := models.Entry{
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	// For database mode, convert to proper structure
	if a.Config.UseDatabase {
		entry.Description = content
		entry.Fields = make(map[string]string)
		entry.Fields["content"] = content
		if len(tags) > 0 {
			entry.Fields["tags"] = strings.Join(tags, ", ")
		}
		return a.DB.AddEntry(tableName, entry)
	}

	_, err := a.Storage.AddEntry(tableName, entry)
	return err
}

// UpdateEntry updates an existing entry
func (a *App) UpdateEntry(tableName, entryID, title, content string, tags []string) error {
	if a.Config.UseDatabase {
		// Get current entry to preserve fields
		table, err := a.DB.GetTable(tableName)
		if err != nil {
			return err
		}

		entry, found := table.GetEntry(entryID)
		if !found {
			return errors.New("entry not found")
		}

		// Update fields
		entry.Title = title
		entry.Description = content

		if entry.Fields == nil {
			entry.Fields = make(map[string]string)
		}

		entry.Fields["content"] = content
		if len(tags) > 0 {
			entry.Fields["tags"] = strings.Join(tags, ", ")
		}

		return a.DB.UpdateEntry(tableName, entryID, entry)
	}

	table, err := a.GetTable(tableName)
	if err != nil {
		return err
	}

	updatedEntry := models.Entry{
		ID:      entryID,
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	if !table.UpdateEntry(entryID, updatedEntry) {
		return errors.New("failed to update entry")
	}

	return a.Storage.SaveTable(table)
}

// DeleteEntry deletes an entry from a table
func (a *App) DeleteEntry(tableName, entryID string) error {
	if a.Config.UseDatabase {
		return a.DB.DeleteEntry(tableName, entryID)
	}

	_, err := a.Storage.RemoveEntry(tableName, entryID)
	return err
}

// ImportTable imports a table from a file
func (a *App) ImportTable(path string) (*models.Table, error) {
	// Check if file exists and has the right extension
	if !strings.HasSuffix(path, ".thighpad") {
		return nil, errors.New("file must have .thighpad extension")
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, errors.New("path is a directory, not a file")
	}

	// Load the table from the file
	table, err := models.LoadTable(path)
	if err != nil {
		return nil, err
	}

	if a.Config.UseDatabase {
		// Import into database
		if err := a.DB.ImportTable(table); err != nil {
			return nil, err
		}
		return a.DB.GetTable(table.Name)
	}

	return a.Storage.ImportTable(path)
}

// ExportTable exports a table to a file
func (a *App) ExportTable(tableName, path string) error {
	// Ensure path has the right extension
	if !strings.HasSuffix(path, ".thighpad") {
		path += ".thighpad"
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if a.Config.UseDatabase {
		// Export from database
		table, err := a.DB.ExportTable(tableName)
		if err != nil {
			return err
		}

		// Save to file
		table.Path = path
		return table.Save()
	}

	return a.Storage.ExportTable(tableName, path)
}

// SearchEntries searches for entries across all tables
func (a *App) SearchEntries(query string) ([]models.SearchResult, error) {
	if a.Config.UseDatabase {
		return a.DB.SearchEntries(query)
	}

	// Simple search for file storage (could be improved)
	var results []models.SearchResult
	tables := a.GetAllTables()

	for _, table := range tables {
		for _, entry := range table.GetEntrySlice() {
			query = strings.ToLower(query)
			found := false

			if strings.Contains(strings.ToLower(entry.Title), query) {
				found = true
			}

			if strings.Contains(strings.ToLower(entry.Content), query) {
				found = true
			}

			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					found = true
					break
				}
			}

			if found {
				results = append(results, models.SearchResult{
					EntryID:   entry.ID,
					TableName: table.Name,
					Title:     entry.Title,
					Context:   truncateString(entry.Content, 50),
				})
			}
		}
	}

	return results, nil
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Close performs cleanup operations when the app is shutting down
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
