package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/data"
)

func (a *App) updateExportScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			a.exportLocation = DefaultExport
			return a, nil
		case "2":
			a.exportLocation = DesktopExport
			return a, nil
		case "3":
			a.exportLocation = BothExport
			return a, nil
		case "enter":
			if a.exportName.Value() != "" {
				var filename string
				var err error

				switch a.exportLocation {
				case DefaultExport:
					filename, err = data.ExportTable(a.currentTable.ID, a.config.Username)
				case DesktopExport:
					filename, err = data.ExportTableToDesktop(a.currentTable.ID, a.config.Username)
				case BothExport:
					filename, err = data.ExportTableToLocation(a.currentTable.ID, a.config.Username, data.BothLocations)
				}

				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = TableScreen
				a.successMsg = "Table exported successfully to: " + filename
				return a, nil
			}
		case "esc":
			a.screen = TableScreen
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	a.exportName, cmd = a.exportName.Update(msg)
	return a, cmd
}

func (a *App) viewExportScreen() string {
	title := Title.Render("Export Table")
	subtitle := Subtitle.Render(a.currentTable.Name)

	exportInfo := BoxStyle.Render(
		Normal.Render(fmt.Sprintf("Exporting table with %d entries", len(a.entries))),
	)

	locationInfo := ""
	switch a.exportLocation {
	case DefaultExport:
		locationInfo = Success.Render("Export location: Default (config folder)")
	case DesktopExport:
		locationInfo = Success.Render("Export location: Desktop")
	case BothExport:
		locationInfo = Success.Render("Export location: Both config folder and desktop")
	default:
		locationInfo = Normal.Render("Select export location: [1] Default  [2] Desktop  [3] Both")
	}

	help := HelpView(map[string]string{
		"1-3":    "Select location",
		"Enter":  "Export table",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		exportInfo,
		locationInfo,
		help,
	)
}
