package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	if isFirstRun() {
		fmt.Println()
		fmt.Println("╔════════════════════════════════════════════════════╗")
		fmt.Println("║ Welcome to ThighPads!                              ║")
		fmt.Println("║                                                    ║")
		fmt.Println("║ This appears to be your first time running the     ║")
		fmt.Println("║ application. Would you like to install ThighPads   ║")
		fmt.Println("║ globally so you can run it from anywhere?          ║")
		fmt.Println("╚════════════════════════════════════════════════════╝")
		fmt.Print("Install globally? [Y/n]: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(response)

		if response == "" || response == "y" || response == "yes" {
			fmt.Println("\nInstalling ThighPads globally...")

			doneChannel := make(chan bool)
			errChannel := make(chan error)

			switch runtime.GOOS {
			case "windows":
				go func() {
					err := installWindowsGlobalSilently()
					if err != nil {
						errChannel <- err
					} else {
						doneChannel <- true
					}
				}()
			case "darwin", "linux":
				go func() {
					err := installUnixGlobalSilently()
					if err != nil {
						errChannel <- err
					} else {
						doneChannel <- true
					}
				}()
			}

			progressDone := make(chan bool)
			go PrintIndeterminateProgress("Installing ThighPads globally", progressDone)

			select {
			case <-doneChannel:
				progressDone <- true
				time.Sleep(500 * time.Millisecond)
				fmt.Println("\n✅ ThighPads installed successfully!")
				fmt.Println("   You can now run 'thighpads' from any directory.")
			case err := <-errChannel:
				progressDone <- true
				fmt.Printf("\n❌ Installation failed: %v\n", err)
			}

			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("Skipping global installation.")
			fmt.Println("You can install it later with 'thighpads --install'")
			time.Sleep(1 * time.Second)
		}
	} else {
		go func() {
			switch runtime.GOOS {
			case "windows":
				installWindowsGlobalSilently()
			case "darwin", "linux":
				installUnixGlobalSilently()
			}
		}()
	}
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	possibleDirs := []string{
		"/usr/local/bin",
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(homeDir, "bin"),
	}

	var destDir string
	for _, dir := range possibleDirs {

		dirExists := true
		if _, err := os.Stat(dir); os.IsNotExist(err) {

			if err := os.MkdirAll(dir, 0755); err != nil {
				continue
			}
			dirExists = false
		}

		testFile := filepath.Join(dir, ".thighpads_write_test")
		if err := os.WriteFile(testFile, []byte{}, 0644); err == nil {
			os.Remove(testFile)
			destDir = dir
			break
		}

		if !dirExists {
			os.Remove(dir)
		}
	}

	if destDir == "" {
		destDir = filepath.Join(homeDir, ".local", "bin")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("couldn't create installation directory: %w", err)
		}
	}

	destPath := filepath.Join(destDir, "thighpads")
	input, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return err
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return err
	}

	path := os.Getenv("PATH")
	if !strings.Contains(path, destDir) {

		profiles := []string{
			filepath.Join(homeDir, ".bashrc"),
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".zshrc"),
			filepath.Join(homeDir, ".profile"),
		}

		for _, profile := range profiles {
			if _, err := os.Stat(profile); err == nil {
				appendCmd := fmt.Sprintf("\n# Added by ThighPads\nexport PATH=\"%s:$PATH\"\n", destDir)
				profileContent, err := os.ReadFile(profile)
				if err == nil && !strings.Contains(string(profileContent), destDir) {
					os.WriteFile(profile, append(profileContent, []byte(appendCmd)...), 0644)
				}
			}
		}

		os.Setenv("PATH", destDir+":"+os.Getenv("PATH"))
	}

	if destDir != "/usr/local/bin" {

		lnCmd := exec.Command("sudo", "ln", "-sf", destPath, "/usr/local/bin/thighpads")
		if err := lnCmd.Run(); err != nil {

			exec.Command("ln", "-sf", destPath, "/usr/local/bin/thighpads").Run()
		}
	}

	return nil
}

func installGlobal() error {
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║ ThighPads Global Installation                      ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println("Installing ThighPads globally on your system...")

	progressDone := make(chan bool)
	errorChan := make(chan error)

	go func() {
		var err error
		switch runtime.GOOS {
		case "windows":
			err = installWindowsGlobal()
		case "darwin", "linux":
			err = installUnixGlobal()
		default:
			err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}

		if err != nil {
			errorChan <- err
		} else {
			progressDone <- true
		}
	}()

	doneChannel := make(chan bool)
	go PrintIndeterminateProgress("Installing ThighPads globally", doneChannel)

	select {
	case <-progressDone:
		doneChannel <- true
		time.Sleep(500 * time.Millisecond)
		fmt.Println("\n✅ ThighPads installation complete!")
		fmt.Println("You can now run 'thighpads' from any directory.")
		return nil
	case err := <-errorChan:
		doneChannel <- true
		fmt.Println("\n❌ Installation failed.")
		return err
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
		return fmt.Errorf("installed to %s but couldn't add to PATH: %w", destPath, err)
	}

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

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	path := os.Getenv("PATH")
	if !strings.Contains(path, destDir) {
		os.Setenv("PATH", destDir+":"+os.Getenv("PATH"))

		homeDir, _ := os.UserHomeDir()
		profiles := []string{
			filepath.Join(homeDir, ".bashrc"),
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".zshrc"),
			filepath.Join(homeDir, ".profile"),
		}

		for _, profile := range profiles {
			if _, err := os.Stat(profile); err == nil {
				appendCmd := fmt.Sprintf("\n# Added by ThighPads\nexport PATH=\"%s:$PATH\"\n", destDir)
				profileContent, err := os.ReadFile(profile)
				if err == nil && !strings.Contains(string(profileContent), destDir) {
					os.WriteFile(profile, append(profileContent, []byte(appendCmd)...), 0644)
				}
			}
		}
	}

	if destDir != "/usr/local/bin" {
		lnCmd := exec.Command("sudo", "ln", "-sf", destPath, "/usr/local/bin/thighpads")
		if err := lnCmd.Run(); err != nil {
			exec.Command("ln", "-sf", destPath, "/usr/local/bin/thighpads").Run()
		}
	}

	return nil
}

func uninstallGlobal() error {
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║ ThighPads Uninstallation                           ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Print("Are you sure you want to uninstall ThighPads? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(response)

	if response != "y" && response != "yes" {
		fmt.Println("Uninstallation cancelled.")
		return nil
	}

	fmt.Println("Uninstalling ThighPads...")

	progressDone := make(chan bool)
	errorChan := make(chan error)

	go func() {
		var err error
		switch runtime.GOOS {
		case "windows":
			err = uninstallWindowsGlobal()
		case "darwin", "linux":
			err = uninstallUnixGlobal()
		default:
			err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}

		if err != nil {
			errorChan <- err
		} else {
			progressDone <- true
		}
	}()

	doneChannel := make(chan bool)
	go PrintIndeterminateProgress("Uninstalling ThighPads", doneChannel)

	select {
	case <-progressDone:
		doneChannel <- true
		time.Sleep(500 * time.Millisecond)
		fmt.Println("\n✅ ThighPads has been uninstalled successfully.")
		fmt.Println("   You may need to restart your terminal for PATH changes to take effect.")
		return nil
	case err := <-errorChan:
		doneChannel <- true
		fmt.Println("\n❌ Uninstallation failed.")
		return err
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
				uninstalled = true
			}
		}
	}

	if !uninstalled {
		return fmt.Errorf("ThighPads is not installed globally or couldn't be found")
	}

	profiles := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	for _, profile := range profiles {
		if _, err := os.Stat(profile); err == nil {
			content, err := os.ReadFile(profile)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				var newLines []string

				for _, line := range lines {
					if strings.Contains(line, "# Added by ThighPads") ||
						(strings.Contains(line, "export PATH=") &&
							(strings.Contains(line, "/bin/thighpads") ||
								strings.Contains(line, "ThighPads"))) {
						continue
					}
					newLines = append(newLines, line)
				}

				os.WriteFile(profile, []byte(strings.Join(newLines, "\n")), 0644)
			}
		}
	}

	return nil
}
