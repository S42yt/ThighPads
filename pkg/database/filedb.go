package database

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/models"
)

type FileDB struct {
	Tables  []models.Table
	Entries []models.Entry
	mu      sync.RWMutex
	dbPath  string
	nextID  uint
}

var fileDB *FileDB

func InitializeFileDB() error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(configPath, "thighpads.json")

	fileDB = &FileDB{
		Tables:  []models.Table{},
		Entries: []models.Entry{},
		dbPath:  dbPath,
		nextID:  1,
	}

	if _, err := os.Stat(dbPath); err == nil {
		data, err := os.ReadFile(dbPath)
		if err != nil {
			return err
		}

		var db FileDB
		if err := json.Unmarshal(data, &db); err != nil {
			return err
		}

		fileDB.Tables = db.Tables
		fileDB.Entries = db.Entries

		for _, table := range fileDB.Tables {
			if table.ID >= fileDB.nextID {
				fileDB.nextID = table.ID + 1
			}
		}

		for _, entry := range fileDB.Entries {
			if entry.ID >= fileDB.nextID {
				fileDB.nextID = entry.ID + 1
			}
		}
	}

	DB = nil

	return nil
}

func (db *FileDB) Save() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.dbPath, data, 0644)
}

func (db *FileDB) CreateTable(table *models.Table) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	table.ID = db.nextID
	db.nextID++
	table.CreatedAt = time.Now()

	db.Tables = append(db.Tables, *table)
	return db.Save()
}

func (db *FileDB) GetTables() ([]models.Table, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.Tables, nil
}

func (db *FileDB) GetTable(id uint) (models.Table, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, table := range db.Tables {
		if table.ID == id {
			return table, nil
		}
	}

	return models.Table{}, errors.New("table not found")
}

func (db *FileDB) GetTableWithEntries(id uint) (models.Table, error) {
	table, err := db.GetTable(id)
	if err != nil {
		return table, err
	}

	entries, err := db.GetEntries(id)
	if err != nil {
		return table, err
	}

	table.Entries = entries
	return table, nil
}

func (db *FileDB) DeleteTable(id uint) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tableIndex := -1
	for i, table := range db.Tables {
		if table.ID == id {
			tableIndex = i
			break
		}
	}

	if tableIndex == -1 {
		return errors.New("table not found")
	}

	db.Tables = append(db.Tables[:tableIndex], db.Tables[tableIndex+1:]...)

	var remainingEntries []models.Entry
	for _, entry := range db.Entries {
		if entry.TableID != id {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	db.Entries = remainingEntries
	return db.Save()
}

func (db *FileDB) CreateEntry(entry *models.Entry) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tableExists := false
	for _, table := range db.Tables {
		if table.ID == entry.TableID {
			tableExists = true
			break
		}
	}

	if !tableExists {
		return errors.New("table not found")
	}

	entry.ID = db.nextID
	db.nextID++
	entry.CreatedAt = time.Now()

	db.Entries = append(db.Entries, *entry)
	return db.Save()
}

func (db *FileDB) GetEntries(tableID uint) ([]models.Entry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entries []models.Entry
	for _, entry := range db.Entries {
		if entry.TableID == tableID {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (db *FileDB) GetEntry(id uint) (models.Entry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, entry := range db.Entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return models.Entry{}, errors.New("entry not found")
}

func (db *FileDB) UpdateEntry(entry *models.Entry) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entryIndex := -1
	for i, e := range db.Entries {
		if e.ID == entry.ID {
			entryIndex = i
			break
		}
	}

	if entryIndex == -1 {
		return errors.New("entry not found")
	}

	db.Entries[entryIndex] = *entry
	return db.Save()
}

func (db *FileDB) DeleteEntry(id uint) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entryIndex := -1
	for i, entry := range db.Entries {
		if entry.ID == id {
			entryIndex = i
			break
		}
	}

	if entryIndex == -1 {
		return errors.New("entry not found")
	}

	db.Entries = append(db.Entries[:entryIndex], db.Entries[entryIndex+1:]...)
	return db.Save()
}

func (db *FileDB) SearchEntries(tableID uint, query string) ([]models.Entry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []models.Entry
	for _, entry := range db.Entries {
		if entry.TableID == tableID {

			if containsIgnoreCase(entry.Title, query) || containsIgnoreCase(entry.Tags, query) {
				results = append(results, entry)
			}
		}
	}

	return results, nil
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func CreateTableWrapper(table *models.Table) error {
	if DB != nil {
		return DB.Create(table).Error
	}
	return fileDB.CreateTable(table)
}

func GetTablesWrapper() ([]models.Table, error) {
	if DB != nil {
		var tables []models.Table
		err := DB.Find(&tables).Error
		return tables, err
	}
	return fileDB.GetTables()
}

func GetTableWrapper(id uint) (models.Table, error) {
	if DB != nil {
		var table models.Table
		err := DB.First(&table, id).Error
		return table, err
	}
	return fileDB.GetTable(id)
}

func GetTableWithEntriesWrapper(id uint) (models.Table, error) {
	if DB != nil {
		var table models.Table
		err := DB.First(&table, id).Error
		if err != nil {
			return table, err
		}

		var entries []models.Entry
		err = DB.Where("table_id = ?", id).Find(&entries).Error
		if err != nil {
			return table, err
		}

		table.Entries = entries
		return table, nil
	}
	return fileDB.GetTableWithEntries(id)
}

func DeleteTableWrapper(id uint) error {
	if DB != nil {
		err := DB.Where("table_id = ?", id).Delete(&models.Entry{}).Error
		if err != nil {
			return err
		}
		return DB.Delete(&models.Table{}, id).Error
	}
	return fileDB.DeleteTable(id)
}

func CreateEntryWrapper(entry *models.Entry) error {
	if DB != nil {
		return DB.Create(entry).Error
	}
	return fileDB.CreateEntry(entry)
}

func GetEntriesWrapper(tableID uint) ([]models.Entry, error) {
	if DB != nil {
		var entries []models.Entry
		err := DB.Where("table_id = ?", tableID).Find(&entries).Error
		return entries, err
	}
	return fileDB.GetEntries(tableID)
}

func GetEntryWrapper(id uint) (models.Entry, error) {
	if DB != nil {
		var entry models.Entry
		err := DB.First(&entry, id).Error
		return entry, err
	}
	return fileDB.GetEntry(id)
}

func UpdateEntryWrapper(entry *models.Entry) error {
	if DB != nil {
		return DB.Save(entry).Error
	}
	return fileDB.UpdateEntry(entry)
}

func DeleteEntryWrapper(id uint) error {
	if DB != nil {
		return DB.Delete(&models.Entry{}, id).Error
	}
	return fileDB.DeleteEntry(id)
}

func SearchEntriesWrapper(tableID uint, query string) ([]models.Entry, error) {
	if DB != nil {
		var entries []models.Entry
		err := DB.Where("table_id = ? AND (title LIKE ? OR tags LIKE ?)",
			tableID, "%"+query+"%", "%"+query+"%").Find(&entries).Error
		return entries, err
	}
	return fileDB.SearchEntries(tableID, query)
}
