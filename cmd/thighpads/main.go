package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/s42yt/thighpads/pkg/app"
	"github.com/s42yt/thighpads/pkg/config"
)

const (
	appVersion        = "1.0.0"
	releasesURL       = "https://api.github.com/repos/s42yt/thighpads/releases/latest"
	updateCheckPeriod = 7 * 24 * time.Hour
)

type GithubRelease struct {
	TagName    string  `json:"tag_name"`
	Assets     []Asset `json:"assets"`
	PreRelease bool    `json:"prerelease"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

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

	var downloadURL string
	assetName := fmt.Sprintf("thighpads_%s_%s_%s", latestVersion, runtime.GOOS, runtime.GOARCH)

	switch runtime.GOOS {
	case "windows":
		assetName += ".exe"
	case "darwin":
		assetName += ".tar.gz"
	case "linux":
		assetName += ".tar.gz"
	}

	for _, asset := range release.Assets {
		if asset.Name == assetName {
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

	if strings.HasSuffix(downloadURL, ".tar.gz") {

		extractDir, err := os.MkdirTemp("", "thighpads_extract_*")
		if err != nil {
			return fmt.Errorf("failed to create extraction directory: %w", err)
		}
		defer os.RemoveAll(extractDir)

		cmd := exec.Command("tar", "-xzf", tempFile.Name(), "-C", extractDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract update: %w", err)
		}

		var executable string
		err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (info.Name() == "thighpads" || info.Name() == "thighpads.exe") {
				executable = path
				return filepath.SkipDir
			}
			return nil
		})

		if executable == "" {
			return fmt.Errorf("executable not found in update package")
		}

		if err := os.Chmod(executable, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}

		currentExe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get current executable path: %w", err)
		}

		input, err := os.ReadFile(executable)
		if err != nil {
			return fmt.Errorf("failed to read new executable: %w", err)
		}

		return os.WriteFile(currentExe, input, 0755)
	} else {

		currentExe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get current executable path: %w", err)
		}

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

		fmt.Println("Update will be applied when ThighPads restarts.")
		os.Exit(0)
		return nil
	}
}

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

func installGlobal() error {
	switch runtime.GOOS {
	case "windows":
		return installWindowsGlobal()
	case "darwin", "linux":
		return installUnixGlobal()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func installWindowsGlobal() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("USERPROFILE environment variable not set")
	}

	destDir := filepath.Join(userProfile, "AppData", "Local", "Programs", "ThighPads")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	destPath := filepath.Join(destDir, "thighpads.exe")

	input, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("failed to read executable: %w", err)
	}

	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return fmt.Errorf("failed to install executable: %w", err)
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;%s", [EnvironmentVariableTarget]::User)`, destDir))

	if err := cmd.Run(); err != nil {
		fmt.Printf("Installed to %s but couldn't add to PATH automatically.\n", destPath)
		fmt.Printf("Please add %s to your PATH manually.\n", destDir)
		return nil
	}

	fmt.Printf("ThighPads installed successfully to %s and added to PATH.\n", destPath)
	fmt.Println("You may need to restart your terminal to use the 'thighpads' command.")
	return nil
}

func installUnixGlobal() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	var destDir string

	if os.Getuid() == 0 {

		destDir = "/usr/local/bin"
	} else {

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		binDir := filepath.Join(homeDir, "bin")
		localBinDir := filepath.Join(homeDir, ".local", "bin")

		path := os.Getenv("PATH")
		if strings.Contains(path, localBinDir) {
			destDir = localBinDir
		} else if strings.Contains(path, binDir) {
			destDir = binDir
		} else {

			destDir = localBinDir
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create bin directory: %w", err)
		}
	}

	destPath := filepath.Join(destDir, "thighpads")

	input, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("failed to read executable: %w", err)
	}

	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return fmt.Errorf("failed to write executable: %w", err)
	}

	fmt.Printf("ThighPads installed successfully to %s\n", destPath)

	path := os.Getenv("PATH")
	if !strings.Contains(path, destDir) {
		fmt.Printf("NOTE: %s is not in your PATH. Add it with:\n", destDir)
		fmt.Printf("echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc && source ~/.bashrc\n", destDir)
	}

	return nil
}

func main() {
	wipe := flag.Bool("wipe", false, "Wipe all ThighPads data and start fresh")
	showVersion := flag.Bool("version", false, "Show version information")
	skipInstall := flag.Bool("skip-install", false, "Skip global installation")
	forceInstall := flag.Bool("install", false, "Force global installation")
	checkUpdate := flag.Bool("check-update", false, "Check for updates")
	update := flag.Bool("update", false, "Update ThighPads to the latest version")
	flag.Parse()

	if *wipe {
		redBlink := "\033[5;31m"
		reset := "\033[0m"

		fmt.Printf("%sWARNING: You are about to delete all ThighPads data!%s\n", redBlink, reset)
		fmt.Print("This will permanently erase all your tables and entries. Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			fmt.Println("Wiping all ThighPads data...")
			if err := wipeData(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to wipe data: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All data wiped. Starting fresh.")
		} else {
			fmt.Println("Wipe operation cancelled.")
			os.Exit(0)
		}
	}

	if *showVersion {
		if err := version(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get version: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *checkUpdate || *update {
		fmt.Println("Checking for updates...")
		hasUpdate, newVersion, downloadURL, err := checkForUpdates(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
			if !*update {
				os.Exit(1)
			}
		} else if hasUpdate {
			fmt.Printf("New version available: v%s (current: v%s)\n", newVersion, appVersion)
			if *update {
				if err := updateThighPads(downloadURL); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Update completed successfully!")
				os.Exit(0)
			}
		} else {
			fmt.Println("You're already running the latest version.")
			if *update {
				os.Exit(0)
			}
		}

		if *checkUpdate {
			os.Exit(0)
		}
	} else {

		go func() {
			hasUpdate, newVersion, _, err := checkForUpdates(false)
			if err == nil && hasUpdate {
				fmt.Printf("\nNew version v%s available! Run 'thighpads --update' to update.\n", newVersion)
			}
		}()
	}

	if *forceInstall || (!*skipInstall && isFirstRun()) {
		fmt.Println("Installing ThighPads as a global command...")
		if err := installGlobal(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install globally: %v\n", err)

		}
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
