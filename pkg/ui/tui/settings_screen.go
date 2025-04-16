package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
)

func (a *App) updateSettingsScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":

			switch a.config.Theme {
			case "default":
				a.config.Theme = "dark"
			case "dark":
				a.config.Theme = "light"
			case "light":
				a.config.Theme = "custom"
			default:
				a.config.Theme = "default"
			}
			return a, nil
		case "2":

			a.config.AutoCheckUpdate = !a.config.AutoCheckUpdate
			return a, nil
		case "3":

			switch a.config.DefaultExport {
			case "config":
				a.config.DefaultExport = "desktop"
			case "desktop":
				a.config.DefaultExport = "both"
			default:
				a.config.DefaultExport = "config"
			}
			return a, nil
		case "s":

			err := config.SaveConfig(a.config)
			if err != nil {
				a.errorMsg = err.Error()
				return a, nil
			}
			a.screen = HomeScreen
			a.successMsg = "Settings saved successfully."
			return a, nil
		case "esc":
			a.screen = HomeScreen
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	return a, cmd
}

func (a *App) viewSettingsScreen() string {
	title := Title.Render("Settings")

	themeStatus := ""
	switch a.config.Theme {
	case "default":
		themeStatus = "Default"
	case "dark":
		themeStatus = "Dark"
	case "light":
		themeStatus = "Light"
	case "custom":
		themeStatus = "Custom"
	}

	autoUpdateStatus := "On"
	if !a.config.AutoCheckUpdate {
		autoUpdateStatus = "Off"
	}

	exportStatus := ""
	switch a.config.DefaultExport {
	case "config":
		exportStatus = "Config folder"
	case "desktop":
		exportStatus = "Desktop"
	case "both":
		exportStatus = "Both"
	}

	settings := BoxStyle.Render(
		fmt.Sprintf("%s: %s\n\n%s: %s\n\n%s: %s",
			Subtitle.Render("1. Theme"),
			Normal.Render(themeStatus),
			Subtitle.Render("2. Auto-check updates"),
			Normal.Render(autoUpdateStatus),
			Subtitle.Render("3. Default export location"),
			Normal.Render(exportStatus),
		),
	)

	help := HelpView(map[string]string{
		"1-3": "Change setting",
		"s":   "Save settings",
		"Esc": "Cancel",
	})

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		settings,
		help,
	)
}
