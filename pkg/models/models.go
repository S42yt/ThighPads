package models

import (
	"encoding/json"
	"os"
	"time"
)

type Table struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Author    string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Entries   []Entry   `gorm:"-"`
}

type Entry struct {
	ID        uint      `gorm:"primaryKey"`
	TableID   uint      `gorm:"not null"`
	Title     string    `gorm:"not null"`
	Tags      string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type ThemeColors struct {
	Name       string `json:"name"`
	Author     string `json:"author"`
	Version    string `json:"version"`
	Accent     string `json:"accent"`
	Secondary  string `json:"secondary"`
	Text       string `json:"text"`
	Subtle     string `json:"subtle"`
	Error      string `json:"error"`
	Success    string `json:"success"`
	Warning    string `json:"warning"`
	Background string `json:"background"`
}

type SyntaxHighlight struct {
	Name        string              `json:"name"`
	Author      string              `json:"author"`
	Version     string              `json:"version"`
	Language    string              `json:"language"`
	Tags        []string            `json:"tags"`
	TokenColors map[string]string   `json:"tokenColors"`
	Rules       []SyntaxRulePattern `json:"rules"`
}

type SyntaxRulePattern struct {
	Pattern string `json:"pattern"`
	Token   string `json:"token"`
}

func LoadThemeFromFile(filePath string) (*ThemeColors, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var theme ThemeColors
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, err
	}

	return &theme, nil
}

func LoadSyntaxFromFile(filePath string) (*SyntaxHighlight, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var syntax SyntaxHighlight
	if err := json.Unmarshal(data, &syntax); err != nil {
		return nil, err
	}

	return &syntax, nil
}

type Config struct {
	Username            string            `json:"username"`
	Theme               string            `json:"theme"`
	CustomTheme         ThemeColors       `json:"customTheme"`
	EnabledSyntaxThemes []string          `json:"enabledSyntaxThemes"`
	DefaultExport       string            `json:"defaultExport"`
	AutoCheckUpdate     bool              `json:"autoCheckUpdate"`
	EditorWidth         int               `json:"editorWidth"`
	EditorHeight        int               `json:"editorHeight"`
	SyntaxHighlighting  bool              `json:"syntaxHighlighting"`
	AvailableThemes     []string          `json:"availableThemes"`
	AvailableSyntaxes   []string          `json:"availableSyntaxes"`
	TagSyntaxMap        map[string]string `json:"tagSyntaxMap"`
}

func GetDefaultConfig() *Config {
	return &Config{
		Username: "",
		Theme:    "default",
		CustomTheme: ThemeColors{
			Name:       "Default",
			Author:     "ThighPads",
			Version:    "1.0",
			Accent:     "#7D56F4",
			Secondary:  "#AE88FF",
			Text:       "#FFFFFF",
			Subtle:     "#888888",
			Error:      "#FF5555",
			Success:    "#55FF55",
			Warning:    "#FFAA55",
			Background: "#222222",
		},
		DefaultExport:       "config",
		AutoCheckUpdate:     true,
		EditorWidth:         80,
		EditorHeight:        20,
		SyntaxHighlighting:  true,
		EnabledSyntaxThemes: []string{},
		AvailableThemes:     []string{"default", "dark", "light"},
		AvailableSyntaxes:   []string{},
		TagSyntaxMap:        make(map[string]string),
	}
}
