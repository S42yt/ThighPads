package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

func toUserFriendlyPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	normalizedPath := filepath.ToSlash(path)
	normalizedHome := filepath.ToSlash(homeDir)

	if strings.HasPrefix(normalizedPath, normalizedHome) {
		return "~" + normalizedPath[len(normalizedHome):]
	}

	return normalizedPath
}

func fromUserFriendlyPath(path string) string {

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

	if runtime.GOOS == "windows" {
		return filepath.FromSlash(path)
	}

	return path
}

func ensureExtension(path, ext string) string {
	if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(ext)) {
		return path + ext
	}
	return path
}
