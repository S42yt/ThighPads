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

	osName := runtime.GOOS
	arch := runtime.GOARCH

	var assetPattern string
	if osName == "windows" {
		assetPattern = fmt.Sprintf("thighpads_%s_%s.exe", osName, arch)
	} else {
		assetPattern = fmt.Sprintf("thighpads_%s_%s", osName, arch)
	}

	var downloadURL string

	versionedPattern := strings.Replace(assetPattern, "thighpads", fmt.Sprintf("thighpads_%s", release.TagName), 1)

	for _, asset := range release.Assets {
		if asset.Name == versionedPattern {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		for _, asset := range release.Assets {
			if asset.Name == assetPattern {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}
	}

	if downloadURL == "" {
		return false, "", "", fmt.Errorf("no suitable download found for your platform (%s/%s)", osName, arch)
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

	tempDir := os.TempDir()
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	tempFile, err := os.CreateTemp(tempDir, "thighpads_update_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	fmt.Println("Downloading update... (this may take a moment)")

	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}

	fmt.Printf("Downloaded %d bytes\n", written)
	tempFile.Close()

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	absCurrentExe, err := filepath.Abs(currentExe)
	if err == nil {
		currentExe = absCurrentExe
	}

	switch runtime.GOOS {
	case "windows":

		updaterBat := filepath.Join(os.TempDir(), "thighpads_updater.bat")
		batContent := fmt.Sprintf(
			`@echo off
echo Applying ThighPads update...
timeout /t 2 /nobreak > NUL
copy /Y "%s" "%s"
if errorlevel 1 (
  echo Failed to update ThighPads. Please try again or update manually.
  pause
  exit /b 1
)
del "%s"
echo Update completed successfully!
start "" "%s"
exit
`, tempFile.Name(), currentExe, tempFile.Name(), currentExe)

		if err := os.WriteFile(updaterBat, []byte(batContent), 0644); err != nil {
			return fmt.Errorf("failed to create updater script: %w", err)
		}

		cmd := exec.Command("cmd", "/c", "start", "/b", "cmd", "/c", updaterBat)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start updater: %w", err)
		}
	case "darwin", "linux":

		if err := os.Chmod(tempFile.Name(), 0755); err != nil {
			return fmt.Errorf("failed to make update executable: %w", err)
		}

		updaterScript := filepath.Join(os.TempDir(), "thighpads_updater.sh")
		scriptContent := fmt.Sprintf(
			`#!/bin/bash
echo "Applying ThighPads update..."
sleep 2
cp -f "%s" "%s"
if [ $? -ne 0 ]; then
  echo "Failed to update ThighPads. Please try again or update manually."
  exit 1
fi
chmod +x "%s"
rm "%s"
echo "Update completed successfully!"
exec "%s"
`, tempFile.Name(), currentExe, currentExe, tempFile.Name(), currentExe)

		if err := os.WriteFile(updaterScript, []byte(scriptContent), 0755); err != nil {
			return fmt.Errorf("failed to create updater script: %w", err)
		}

		cmd := exec.Command("/bin/bash", updaterScript)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start updater: %w", err)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	fmt.Println("Update will be applied when ThighPads restarts.")
	os.Exit(0)
	return nil
}
