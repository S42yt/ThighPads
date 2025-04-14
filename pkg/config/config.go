package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	AppName     string
	ConfigDir   string
	TablesDir   string
	DataDir     string
	Username    string
	DefaultPath string
	UseDatabase bool
}

func New() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	tablesDir := filepath.Join(configDir, "tables")
	dataDir := filepath.Join(configDir, "data")

	if err := ensureDir(configDir); err != nil {
		return nil, err
	}

	if err := ensureDir(tablesDir); err != nil {
		return nil, err
	}

	if err := ensureDir(dataDir); err != nil {
		return nil, err
	}

	username, err := getUsername()
	if err != nil {
		username = "thighpads_user"
	}

	return &Config{
		AppName:     "ThighPads",
		ConfigDir:   configDir,
		TablesDir:   tablesDir,
		DataDir:     dataDir,
		Username:    username,
		DefaultPath: filepath.Join(tablesDir, "default.thighpad"),
		UseDatabase: true,
	}, nil
}

func (c *Config) GetTablePath(name string) string {
	return filepath.Join(c.TablesDir, name+".thighpad")
}

func (c *Config) GetAllTablePaths() ([]string, error) {
	pattern := filepath.Join(c.TablesDir, "*.thighpad")
	return filepath.Glob(pattern)
}

func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "thighpads"), nil
}

func getUsername() (string, error) {
	if username := os.Getenv("USER"); username != "" {
		return username, nil
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return username, nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}
