package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/models"
)

// Storage represents a data storage backend
type Storage struct {
	db  *database.DB
	cfg *config.Config
}

// New creates a new storage instance
func New(cfg *config.Config) (*Storage, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create database connection
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Storage{
		db:  db,
		cfg: cfg,
	}, nil
}

// Close closes the storage backend
func (s *Storage) Close() error {
	return s.db.Close()
}

// GetTables returns all available tables
func (s *Storage) GetTables() ([]*models.Table, error) {
	return s.db.GetAllTables()
}

// GetTable retrieves a table by name
func (s *Storage) GetTable(name string) (*models.Table, error) {
	return s.db.GetTable(name)
}

// CreateTable creates a new table
func (s *Storage) CreateTable(name, author string) (*models.Table, error) {
	return s.db.CreateTable(name, author)
}

// DeleteTable deletes a table
func (s *Storage) DeleteTable(name string) error {
	return s.db.DeleteTable(name)
}

// AddEntry adds a new entry to a table
func (s *Storage) AddEntry(tableName string, entry models.Entry) error {
	return s.db.AddEntry(tableName, entry)
}

// UpdateEntry updates an existing entry
func (s *Storage) UpdateEntry(tableName string, entryID string, entry models.Entry) error {
	return s.db.UpdateEntry(tableName, entryID, entry)
}

// DeleteEntry removes an entry from a table
func (s *Storage) DeleteEntry(tableName string, entryID string) error {
	return s.db.DeleteEntry(tableName, entryID)
}

// SearchEntries searches for entries across all tables
func (s *Storage) SearchEntries(query string) ([]models.SearchResult, error) {
	return s.db.SearchEntries(query)
}

// ExportTable exports a table to a file
func (s *Storage) ExportTable(tableName, outputPath string) error {
	// Get the table
	table, err := s.db.ExportTable(tableName)
	if err != nil {
		return fmt.Errorf("failed to export table: %w", err)
	}

	// Save it to a file
	table.Path = outputPath
	if err := table.Save(); err != nil {
		return fmt.Errorf("failed to save exported table: %w", err)
	}

	return nil
}

// ImportTable imports a table from a file
func (s *Storage) ImportTable(filePath string) error {
	// Load the table from file
	table, err := models.LoadTable(filePath)
	if err != nil {
		return fmt.Errorf("failed to load table: %w", err)
	}

	// Import it to the database
	return s.db.ImportTable(table)
}

// GetBackupDir returns the backup directory path
func (s *Storage) GetBackupDir() string {
	backupDir := filepath.Join(s.cfg.DataDir, "backups")
	os.MkdirAll(backupDir, 0755) // Ensure the backup directory exists
	return backupDir
}
