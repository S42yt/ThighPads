package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/models"
)

type DB struct {
	db  *sql.DB
	cfg *config.Config
}

func New(cfg *config.Config) (*DB, error) {
	dbPath := filepath.Join(cfg.DataDir, "thighpads.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &DB{
		db:  db,
		cfg: cfg,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func initDatabase(db *sql.DB) error {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tables (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			author TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS entries (
			id TEXT PRIMARY KEY,
			table_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create entries table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS entry_fields (
			entry_id TEXT NOT NULL,
			field_name TEXT NOT NULL,
			field_value TEXT,
			PRIMARY KEY (entry_id, field_name),
			FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create entry_fields table: %w", err)
	}

	return nil
}

func (d *DB) GetAllTables() ([]*models.Table, error) {
	rows, err := d.db.Query(`
		SELECT id, name, author, created_at, updated_at
		FROM tables
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []*models.Table
	for rows.Next() {
		table := &models.Table{
			Entries: make(map[string]models.Entry),
		}
		var createdAt, updatedAt string
		err := rows.Scan(&table.ID, &table.Name, &table.Author, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		entries, err := d.getEntriesForTable(table.ID)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			table.Entries[entry.ID] = entry
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (d *DB) GetTable(name string) (*models.Table, error) {
	row := d.db.QueryRow(`
		SELECT id, name, author, created_at, updated_at
		FROM tables
		WHERE name = ?
	`, name)

	table := &models.Table{
		Entries: make(map[string]models.Entry),
	}
	var createdAt, updatedAt string
	err := row.Scan(&table.ID, &table.Name, &table.Author, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("table not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan table: %w", err)
	}

	entries, err := d.getEntriesForTable(table.ID)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		table.Entries[entry.ID] = entry
	}

	return table, nil
}

func (d *DB) CreateTable(name, author string) (*models.Table, error) {

	tableID := models.GenerateID()
	now := time.Now().Format(time.RFC3339)

	_, err := d.db.Exec(`
		INSERT INTO tables (id, name, author, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, tableID, name, author, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &models.Table{
		ID:      tableID,
		Name:    name,
		Author:  author,
		Path:    "",
		Entries: make(map[string]models.Entry),
	}, nil
}

func (d *DB) DeleteTable(name string) error {

	var tableID string
	row := d.db.QueryRow("SELECT id FROM tables WHERE name = ?", name)
	err := row.Scan(&tableID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("table not found: %s", name)
		}
		return fmt.Errorf("failed to get table ID: %w", err)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM entry_fields
		WHERE entry_id IN (
			SELECT id FROM entries WHERE table_id = ?
		)
	`, tableID)
	if err != nil {
		return fmt.Errorf("failed to delete entry fields: %w", err)
	}

	_, err = tx.Exec("DELETE FROM entries WHERE table_id = ?", tableID)
	if err != nil {
		return fmt.Errorf("failed to delete entries: %w", err)
	}

	_, err = tx.Exec("DELETE FROM tables WHERE id = ?", tableID)
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}

	return tx.Commit()
}

func (d *DB) AddEntry(tableName string, entry models.Entry) error {

	var tableID string
	row := d.db.QueryRow("SELECT id FROM tables WHERE name = ?", tableName)
	err := row.Scan(&tableID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("table not found: %s", tableName)
		}
		return fmt.Errorf("failed to get table ID: %w", err)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if entry.ID == "" {
		entry.ID = models.GenerateID()
	}
	now := time.Now().Format(time.RFC3339)

	_, err = tx.Exec(`
		INSERT INTO entries (id, table_id, title, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, entry.ID, tableID, entry.Title, entry.Description, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	for key, value := range entry.Fields {
		_, err = tx.Exec(`
			INSERT INTO entry_fields (entry_id, field_name, field_value)
			VALUES (?, ?, ?)
		`, entry.ID, key, value)
		if err != nil {
			return fmt.Errorf("failed to insert entry field: %w", err)
		}
	}

	return tx.Commit()
}

func (d *DB) UpdateEntry(tableName string, entryID string, entry models.Entry) error {

	var tableID string
	row := d.db.QueryRow("SELECT id FROM tables WHERE name = ?", tableName)
	err := row.Scan(&tableID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("table not found: %s", tableName)
		}
		return fmt.Errorf("failed to get table ID: %w", err)
	}

	var existingID string
	row = d.db.QueryRow("SELECT id FROM entries WHERE id = ? AND table_id = ?", entryID, tableID)
	err = row.Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("entry not found: %s", entryID)
		}
		return fmt.Errorf("failed to get entry: %w", err)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	_, err = tx.Exec(`
		UPDATE entries
		SET title = ?, description = ?, updated_at = ?
		WHERE id = ?
	`, entry.Title, entry.Description, now, entryID)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	_, err = tx.Exec("DELETE FROM entry_fields WHERE entry_id = ?", entryID)
	if err != nil {
		return fmt.Errorf("failed to delete old fields: %w", err)
	}

	for key, value := range entry.Fields {
		_, err = tx.Exec(`
			INSERT INTO entry_fields (entry_id, field_name, field_value)
			VALUES (?, ?, ?)
		`, entryID, key, value)
		if err != nil {
			return fmt.Errorf("failed to insert entry field: %w", err)
		}
	}

	return tx.Commit()
}

func (d *DB) DeleteEntry(tableName string, entryID string) error {

	var tableID string
	row := d.db.QueryRow("SELECT id FROM tables WHERE name = ?", tableName)
	err := row.Scan(&tableID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("table not found: %s", tableName)
		}
		return fmt.Errorf("failed to get table ID: %w", err)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM entry_fields WHERE entry_id = ?", entryID)
	if err != nil {
		return fmt.Errorf("failed to delete entry fields: %w", err)
	}

	result, err := tx.Exec("DELETE FROM entries WHERE id = ? AND table_id = ?", entryID, tableID)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("entry not found: %s", entryID)
	}

	return tx.Commit()
}

func (d *DB) getEntriesForTable(tableID string) ([]models.Entry, error) {

	rows, err := d.db.Query(`
		SELECT id, title, description
		FROM entries
		WHERE table_id = ?
	`, tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		var entry models.Entry
		entry.Fields = make(map[string]string)
		err := rows.Scan(&entry.ID, &entry.Title, &entry.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		fields, err := d.getEntryFields(entry.ID)
		if err != nil {
			return nil, err
		}
		entry.Fields = fields

		entries = append(entries, entry)
	}

	return entries, nil
}

func (d *DB) getEntryFields(entryID string) (map[string]string, error) {
	rows, err := d.db.Query(`
		SELECT field_name, field_value
		FROM entry_fields
		WHERE entry_id = ?
	`, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entry fields: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]string)
	for rows.Next() {
		var name, value string
		err := rows.Scan(&name, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field: %w", err)
		}
		fields[name] = value
	}

	return fields, nil
}

func (d *DB) SearchEntries(query string) ([]models.SearchResult, error) {

	searchParam := "%" + query + "%"

	rows, err := d.db.Query(`
		SELECT e.id, e.title, e.description, t.name as table_name
		FROM entries e
		JOIN tables t ON e.table_id = t.id
		WHERE e.title LIKE ? OR e.description LIKE ?
		UNION
		SELECT e.id, e.title, e.description, t.name as table_name
		FROM entries e
		JOIN tables t ON e.table_id = t.id
		JOIN entry_fields ef ON e.id = ef.entry_id
		WHERE ef.field_value LIKE ?
		ORDER BY table_name, title
	`, searchParam, searchParam, searchParam)
	if err != nil {
		return nil, fmt.Errorf("failed to search entries: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		err := rows.Scan(&result.EntryID, &result.Title, &result.Description, &result.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		fields, err := d.getEntryFields(result.EntryID)
		if err != nil {
			return nil, err
		}

		query = strings.ToLower(query)
		for name, value := range fields {
			if strings.Contains(strings.ToLower(value), query) {
				maxPreview := 50
				index := strings.Index(strings.ToLower(value), query)
				start := index - 10
				if start < 0 {
					start = 0
				}
				end := index + len(query) + 10
				if end > len(value) {
					end = len(value)
				}
				if len(value) > maxPreview {
					if start > 0 {
						result.Context = "..." + value[start:end]
					} else {
						result.Context = value[start:end] + "..."
					}
					if end < len(value) {
						result.Context += "..."
					}
				} else {
					result.Context = value
				}
				result.MatchingField = name
				break
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (d *DB) ExportTable(tableName string) (*models.Table, error) {

	table, err := d.GetTable(tableName)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (d *DB) ImportTable(table *models.Table) error {

	var count int
	row := d.db.QueryRow("SELECT COUNT(*) FROM tables WHERE name = ?", table.Name)
	err := row.Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("table already exists: %s", table.Name)
	}

	_, err = d.CreateTable(table.Name, table.Author)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	for _, entry := range table.Entries {
		if err := d.AddEntry(table.Name, entry); err != nil {
			return fmt.Errorf("failed to add entry: %w", err)
		}
	}

	return nil
}
