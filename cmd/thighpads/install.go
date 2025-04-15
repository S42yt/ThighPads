package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func isInstalledGlobally() bool {
	switch runtime.GOOS {
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			return false
		}
		destPath := filepath.Join(userProfile, "AppData", "Local", "Programs", "ThighPads", "thighpads.exe")
		_, err := os.Stat(destPath)
		return err == nil
	case "darwin", "linux":
		possibleLocations := []string{
			"/usr/local/bin/thighpads",
			"/usr/bin/thighpads",
		}

		homeDir, err := os.UserHomeDir()
		if err == nil {
			possibleLocations = append(possibleLocations,
				filepath.Join(homeDir, "bin", "thighpads"),
				filepath.Join(homeDir, ".local", "bin", "thighpads"))
		}

		for _, location := range possibleLocations {
			if _, err := os.Stat(location); err == nil {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func installGlobalSilently() {
	go func() {
		switch runtime.GOOS {
		case "windows":
			installWindowsGlobalSilently()
		case "darwin", "linux":
			installUnixGlobalSilently()
		}
	}()
}

func installWindowsGlobalSilently() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("USERPROFILE not set")
	}

	destDir := filepath.Join(userProfile, "AppData", "Local", "Programs", "ThighPads")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, "thighpads.exe")

	input, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return err
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;%s", [EnvironmentVariableTarget]::User)`, destDir))
	return cmd.Run()
}

func installUnixGlobalSilently() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	var destDir string

	if os.Getuid() == 0 {
		destDir = "/usr/local/bin"
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
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
			return err
		}
	}

	destPath := filepath.Join(destDir, "thighpads")

	input, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, input, 0755)
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

func uninstallGlobal() error {
	switch runtime.GOOS {
	case "windows":
		return uninstallWindowsGlobal()
	case "darwin", "linux":
		return uninstallUnixGlobal()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func uninstallWindowsGlobal() error {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("USERPROFILE environment variable not set")
	}

	destDir := filepath.Join(userProfile, "AppData", "Local", "Programs", "ThighPads")
	destPath := filepath.Join(destDir, "thighpads.exe")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return fmt.Errorf("ThighPads is not installed globally")
	}

	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to remove installation directory: %w", err)
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("PATH", ($env:PATH -replace [regex]::Escape(";%s"), ""), [EnvironmentVariableTarget]::User)`, destDir))
	_ = cmd.Run()

	fmt.Println("ThighPads has been uninstalled successfully.")
	fmt.Println("You may need to restart your terminal for PATH changes to take effect.")
	return nil
}

func uninstallUnixGlobal() error {
	possibleLocations := []string{
		"/usr/local/bin/thighpads",
		"/usr/bin/thighpads",
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		possibleLocations = append(possibleLocations,
			filepath.Join(homeDir, "bin", "thighpads"),
			filepath.Join(homeDir, ".local", "bin", "thighpads"))
	}

	uninstalled := false
	for _, location := range possibleLocations {
		if _, err := os.Stat(location); err == nil {
			if err := os.Remove(location); err != nil {
				fmt.Printf("Failed to remove %s: %v\n", location, err)
			} else {
				fmt.Printf("Removed %s\n", location)
				uninstalled = true
			}
		}
	}

	if !uninstalled {
		return fmt.Errorf("ThighPads is not installed globally or couldn't be found")
	}

	fmt.Println("ThighPads has been uninstalled successfully.")
	return nil
}
