package main

import (
	"fmt"
	"os"

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
	_, err = os.Stat(configPath)
	return os.IsNotExist(err)
}
