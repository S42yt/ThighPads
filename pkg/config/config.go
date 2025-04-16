package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/s42yt/thighpads/pkg/models"
)

const (
	ConfigFolderName      = ".config/thighpads"
	ConfigFileName        = "config.json"
	DBFileName            = "thighpads.db"
	ExportFolderName      = "exports"
	ExportsConfigFileName = "exports_config.json"
	ThemesFolderName      = "themes"
	SyntaxFolderName      = "syntax"
)

type ExportsConfig struct {
	DesktopPath string `json:"desktopPath"`
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ConfigFolderName), nil
}

func EnsureConfigFolderExists() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(configPath, 0755)
		if err != nil {
			return "", err
		}

		exportPath := filepath.Join(configPath, ExportFolderName)
		err = os.MkdirAll(exportPath, 0755)
		if err != nil {
			return "", err
		}
	}

	themesPath := filepath.Join(configPath, ThemesFolderName)
	if _, err := os.Stat(themesPath); os.IsNotExist(err) {
		err = os.MkdirAll(themesPath, 0755)
		if err != nil {
			return "", err
		}
	}

	syntaxPath := filepath.Join(configPath, SyntaxFolderName)
	if _, err := os.Stat(syntaxPath); os.IsNotExist(err) {
		err = os.MkdirAll(syntaxPath, 0755)
		if err != nil {
			return "", err
		}
	}

	return configPath, nil
}

func LoadConfig() (*models.Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configPath, ConfigFileName)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.New("config file not found")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config models.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	themes, err := DiscoverThemes()
	if err == nil {
		config.AvailableThemes = themes
	}

	syntaxes, err := DiscoverSyntaxThemes()
	if err == nil {
		config.AvailableSyntaxes = syntaxes
	}

	return &config, nil
}

func SaveConfig(config *models.Config) error {
	configPath, err := EnsureConfigFolderExists()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configPath, ConfigFileName)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func GetDBPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, DBFileName), nil
}

func GetExportPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, ExportFolderName), nil
}

func GetThemesPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, ThemesFolderName), nil
}

func GetSyntaxPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, SyntaxFolderName), nil
}

func IsFirstRun() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}

	configFile := filepath.Join(configPath, ConfigFileName)

	_, err = os.Stat(configFile)
	return os.IsNotExist(err), nil
}

func GetDesktopExportPath() (string, error) {

	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	exportsConfigFile := filepath.Join(configPath, ExportsConfigFileName)
	if _, err := os.Stat(exportsConfigFile); err == nil {
		data, err := os.ReadFile(exportsConfigFile)
		if err == nil {
			var exportsConfig ExportsConfig
			if err := json.Unmarshal(data, &exportsConfig); err == nil && exportsConfig.DesktopPath != "" {

				if _, err := os.Stat(exportsConfig.DesktopPath); err == nil {
					return exportsConfig.DesktopPath, nil
				}
			}
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	desktopDir := filepath.Join(homeDir, "Desktop")
	desktopExportsDir := filepath.Join(desktopDir, "ThighPads Exports")

	if _, err := os.Stat(desktopExportsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(desktopExportsDir, 0755); err != nil {
			return "", err
		}
	}

	return desktopExportsDir, nil
}

func DiscoverThemes() ([]string, error) {
	themesPath, err := GetThemesPath()
	if err != nil {
		return nil, err
	}

	themes := []string{"default", "dark", "light"}

	files, err := os.ReadDir(themesPath)
	if err != nil {

		return themes, nil
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {

			themeName := strings.TrimSuffix(file.Name(), ".json")
			themes = append(themes, themeName)
		}
	}

	return themes, nil
}

func DiscoverSyntaxThemes() ([]string, error) {
	syntaxPath, err := GetSyntaxPath()
	if err != nil {
		return nil, err
	}

	var syntaxes []string

	files, err := os.ReadDir(syntaxPath)
	if err != nil {
		return syntaxes, nil
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {

			syntaxName := strings.TrimSuffix(file.Name(), ".json")
			syntaxes = append(syntaxes, syntaxName)
		}
	}

	return syntaxes, nil
}

func LoadTheme(themeName string) (*models.ThemeColors, error) {
	if themeName == "default" || themeName == "dark" || themeName == "light" {

		return GetBuiltInTheme(themeName), nil
	}

	themesPath, err := GetThemesPath()
	if err != nil {
		return nil, err
	}

	themeFile := filepath.Join(themesPath, themeName+".json")
	return models.LoadThemeFromFile(themeFile)
}

func LoadSyntaxHighlighting(syntaxName string) (*models.SyntaxHighlight, error) {
	syntaxPath, err := GetSyntaxPath()
	if err != nil {
		return nil, err
	}

	syntaxFile := filepath.Join(syntaxPath, syntaxName+".json")
	return models.LoadSyntaxFromFile(syntaxFile)
}

func ImportThemeFromFile(srcPath string) error {
	themesPath, err := GetThemesPath()
	if err != nil {
		return err
	}

	theme, err := models.LoadThemeFromFile(srcPath)
	if err != nil {
		return err
	}

	var destFilename string
	if theme.Name != "" {
		destFilename = theme.Name + ".json"
	} else {
		destFilename = filepath.Base(srcPath)
	}

	if !strings.HasSuffix(destFilename, ".json") {
		destFilename += ".json"
	}

	destPath := filepath.Join(themesPath, destFilename)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, data, 0644)
}

func ImportSyntaxFromFile(srcPath string) error {
	syntaxPath, err := GetSyntaxPath()
	if err != nil {
		return err
	}

	syntax, err := models.LoadSyntaxFromFile(srcPath)
	if err != nil {
		return err
	}

	var destFilename string
	if syntax.Name != "" {
		destFilename = syntax.Name + ".json"
	} else {
		destFilename = filepath.Base(srcPath)
	}

	if !strings.HasSuffix(destFilename, ".json") {
		destFilename += ".json"
	}

	destPath := filepath.Join(syntaxPath, destFilename)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, data, 0644)
}

func GetBuiltInTheme(name string) *models.ThemeColors {
	switch name {
	case "dark":
		return &models.ThemeColors{
			Name:       "Dark",
			Author:     "ThighPads",
			Version:    "1.0",
			Accent:     "#7D56F4",
			Secondary:  "#AE88FF",
			Text:       "#FFFFFF",
			Subtle:     "#888888",
			Error:      "#FF5555",
			Success:    "#55FF55",
			Warning:    "#FFAA55",
			Background: "#121212",
		}
	case "light":
		return &models.ThemeColors{
			Name:       "Light",
			Author:     "ThighPads",
			Version:    "1.0",
			Accent:     "#7D56F4",
			Secondary:  "#9D66FF",
			Text:       "#333333",
			Subtle:     "#777777",
			Error:      "#CC0000",
			Success:    "#00CC00",
			Warning:    "#CC7700",
			Background: "#F5F5F5",
		}
	default:
		return &models.ThemeColors{
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
		}
	}
}

func NormalizePath(path string) string {

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}

	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}

	return filepath.Clean(path)
}
