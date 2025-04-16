package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/s42yt/thighpads/pkg/app"
)

func main() {

	wipe := flag.Bool("wipe", false, "Wipe all ThighPads data and start fresh")
	showVersion := flag.Bool("version", false, "Show version information")
	skipInstall := flag.Bool("skip-install", false, "Skip global installation")
	forceInstall := flag.Bool("install", false, "Force global installation")
	uninstall := flag.Bool("uninstall", false, "Uninstall ThighPads from your system")
	checkUpdate := flag.Bool("check-update", false, "Check for updates")
	update := flag.Bool("update", false, "Update ThighPads to the latest version")
	showHelp := flag.Bool("help", false, "Show detailed help information")
	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	if *uninstall {
		fmt.Println("Uninstalling ThighPads...")
		if err := uninstallGlobal(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

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

	if !*skipInstall && !isInstalledGlobally() {
		if *forceInstall {
			fmt.Println("Installing ThighPads as a global command...")
			if err := installGlobal(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to install globally: %v\n", err)
			}
		} else {
			installGlobalSilently()
		}
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║ ThighPads Help                                     ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  thighpads [options]")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("  --help           Show this help message")
	fmt.Println("  --version        Show version information")
	fmt.Println("  --check-update   Check for updates")
	fmt.Println("  --update         Update ThighPads to the latest version")
	fmt.Println("  --wipe           Wipe all ThighPads data and start fresh")
	fmt.Println("  --install        Force global installation")
	fmt.Println("  --skip-install   Skip global installation")
	fmt.Println("  --uninstall      Uninstall ThighPads from your system")
	fmt.Println("")
	fmt.Println("KEYBOARD SHORTCUTS:")
	fmt.Println("  Global:")
	fmt.Println("    Ctrl+C    Quit")
	fmt.Println("    Esc       Go back or cancel")
	fmt.Println("")
	fmt.Println("  Home Screen:")
	fmt.Println("    Enter     Select table")
	fmt.Println("    n         New table")
	fmt.Println("    i         Import table")
	fmt.Println("    q         Quit")
	fmt.Println("")
	fmt.Println("  Table Screen:")
	fmt.Println("    Enter     View entry")
	fmt.Println("    n         New entry")
	fmt.Println("    d         Delete entry")
	fmt.Println("    e         Export table")
	fmt.Println("    b         Back to home")
	fmt.Println("    q         Quit")
	fmt.Println("")
	fmt.Println("FILES:")
	fmt.Println("  Configuration: ~/.config/thighpads/config.json")
	fmt.Println("  Database:      ~/.config/thighpads/thighpads.db")
	fmt.Println("  Exports:       ~/.config/thighpads/exports/")
	fmt.Println("")
}
