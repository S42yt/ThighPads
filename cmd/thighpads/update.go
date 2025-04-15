package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/s42yt/thighpads/pkg/config"
)

func checkForUpdates(forceCheck bool) (bool, string, string, error) {
	if !forceCheck {
		lastCheck, err := getLastUpdateCheck()
		if err == nil && time.Since(lastCheck) < updateCheckPeriod {
			return false, "", "", nil
		}
	}

	saveLastUpdateCheck()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releasesURL)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", "", fmt.Errorf("failed to parse update info: %w", err)
	}

	if release.PreRelease {
		return false, "", "", nil
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if latestVersion == appVersion {
		return false, "", "", nil
	}

	downloadFileName := fmt.Sprintf("thighpads_%s.exe", release.TagName)
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == downloadFileName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return false, "", "", fmt.Errorf("no suitable download found for your platform")
	}

	return true, latestVersion, downloadURL, nil
}

func getLastUpdateCheck() (time.Time, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return time.Time{}, err
	}

	updateFile := filepath.Join(configPath, "lastupdate")
	data, err := os.ReadFile(updateFile)
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, err
	}

	return timestamp, nil
}

func saveLastUpdateCheck() error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	updateFile := filepath.Join(configPath, "lastupdate")
	timestamp := time.Now().Format(time.RFC3339)

	return os.WriteFile(updateFile, []byte(timestamp), 0644)
}

func updateThighPads(downloadURL string) error {
	fmt.Println("Downloading update...")

	tempFile, err := os.CreateTemp("", "thighpads_update_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}

	tempFile.Close()

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	if runtime.GOOS == "windows" {
		updaterBat := filepath.Join(os.TempDir(), "thighpads_updater.bat")
		batContent := fmt.Sprintf(
			`@echo off
timeout /t 1 /nobreak > NUL
copy /Y "%s" "%s"
del "%s"
start "" "%s"
exit
`, tempFile.Name(), currentExe, tempFile.Name(), currentExe)

		if err := os.WriteFile(updaterBat, []byte(batContent), 0644); err != nil {
			return fmt.Errorf("failed to create updater script: %w", err)
		}

		cmd := exec.Command("cmd", "/c", updaterBat)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start updater: %w", err)
		}
	} else {
		input, err := os.ReadFile(tempFile.Name())
		if err != nil {
			return fmt.Errorf("failed to read new executable: %w", err)
		}

		if err := os.WriteFile(currentExe, input, 0755); err != nil {
			return fmt.Errorf("failed to update executable: %w", err)
		}
	}

	fmt.Println("Update will be applied when ThighPads restarts.")
	os.Exit(0)
	return nil
}
