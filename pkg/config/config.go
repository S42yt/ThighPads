package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/s42yt/thighpads/pkg/models"
)

const (
	ConfigFolderName      = ".config/thighpads"
	ConfigFileName        = "config.json"
	DBFileName            = "thighpads.db"
	ExportFolderName      = "exports"
	ExportsConfigFileName = "exports_config.json"
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
