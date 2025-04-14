package models

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Entry struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags"`
	Content     string            `json:"content"`
	Fields      map[string]string `json:"fields,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}


type Table struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Author    string           `json:"author"`
	Entries   map[string]Entry `json:"entries"` 
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
	Path      string           `json:"-"` 
}


type SearchResult struct {
	EntryID       string `json:"entry_id"`
	TableName     string `json:"table_name"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	Context       string `json:"context,omitempty"`
	MatchingField string `json:"matching_field,omitempty"`
}


func NewTable(name, author string) *Table {
	now := time.Now().Format(time.RFC3339)
	return &Table{
		ID:        GenerateID(),
		Name:      name,
		Author:    author,
		Entries:   make(map[string]Entry),
		CreatedAt: now,
		UpdatedAt: now,
	}
}


func (t *Table) AddEntry(entry Entry) {
	now := time.Now().Format(time.RFC3339)
	if entry.ID == "" {
		entry.ID = GenerateID()
	}
	if entry.CreatedAt == "" {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	t.Entries[entry.ID] = entry
	t.UpdatedAt = now
}


func (t *Table) RemoveEntry(id string) bool {
	if _, exists := t.Entries[id]; exists {
		delete(t.Entries, id)
		t.UpdatedAt = time.Now().Format(time.RFC3339)
		return true
	}
	return false
}


func (t *Table) GetEntry(id string) (Entry, bool) {
	entry, ok := t.Entries[id]
	return entry, ok
}


func (t *Table) UpdateEntry(id string, updatedEntry Entry) bool {
	if _, exists := t.Entries[id]; exists {
		updatedEntry.ID = id                             
		updatedEntry.CreatedAt = t.Entries[id].CreatedAt 
		updatedEntry.UpdatedAt = time.Now().Format(time.RFC3339)
		t.Entries[id] = updatedEntry
		t.UpdatedAt = updatedEntry.UpdatedAt
		return true
	}
	return false
}


func (t *Table) GetEntrySlice() []Entry {
	entries := make([]Entry, 0, len(t.Entries))
	for _, entry := range t.Entries {
		entries = append(entries, entry)
	}
	return entries
}


func (t *Table) Save() error {
	if t.Path == "" {
		return errors.New("table has no path defined")
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.Path, data, 0644)
}


func LoadTable(path string) (*Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var table Table
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, err
	}

	
	if table.Entries == nil {
		table.Entries = make(map[string]Entry)
	}

	
	if table.ID == "" {
		table.ID = GenerateID()
	}

	table.Path = path
	return &table, nil
}


func GenerateID() string {
	return time.Now().Format("20060102150405.000")
}
