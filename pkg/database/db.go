package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/models"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Initialize() error {
	dbPath, err := config.GetDBPath()
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		fmt.Println("Warning: Could not initialize SQLite database, falling back to file-based storage.")
		fmt.Println("Error was:", err.Error())
		return InitializeFileDB()
	}

	err = db.AutoMigrate(&models.Table{}, &models.Entry{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func CreateTable(table *models.Table) error {
	return CreateTableWrapper(table)
}

func GetTables() ([]models.Table, error) {
	return GetTablesWrapper()
}

func GetTable(id uint) (models.Table, error) {
	return GetTableWrapper(id)
}

func GetTableWithEntries(id uint) (models.Table, error) {
	return GetTableWithEntriesWrapper(id)
}

func DeleteTable(id uint) error {
	return DeleteTableWrapper(id)
}

func CreateEntry(entry *models.Entry) error {
	return CreateEntryWrapper(entry)
}

func GetEntries(tableID uint) ([]models.Entry, error) {
	return GetEntriesWrapper(tableID)
}

func GetEntry(id uint) (models.Entry, error) {
	return GetEntryWrapper(id)
}

func UpdateEntry(entry *models.Entry) error {
	return UpdateEntryWrapper(entry)
}

func DeleteEntry(id uint) error {
	return DeleteEntryWrapper(id)
}

func SearchEntries(tableID uint, query string) ([]models.Entry, error) {
	return SearchEntriesWrapper(tableID, query)
}
