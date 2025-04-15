package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/s42yt/thighpads/pkg/config"
)

func version() error {
	fmt.Printf("ThighPads v%s\n", appVersion)
	return nil
}

func wipeData() error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	return os.RemoveAll(configPath)
}

func isFirstRun() bool {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return true
	}

	configFile := filepath.Join(configPath, "config.json")
	_, err = os.Stat(configFile)
	return os.IsNotExist(err)
}
