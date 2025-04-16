package tui

import (
	"fmt"
	"strings"

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

				exportName := a.exportName.Value()

				if !strings.HasSuffix(strings.ToLower(exportName), strings.ToLower(data.FileExtension)) {
					exportName += data.FileExtension
				}

				switch a.exportLocation {
				case DefaultExport:
					filename, err = data.ExportTable(a.currentTable.ID, a.config.Username, exportName)
				case DesktopExport:
					filename, err = data.ExportTableToDesktop(a.currentTable.ID, a.config.Username, exportName)
				case BothExport:
					filename, err = data.ExportTableToLocation(a.currentTable.ID, a.config.Username, data.BothLocations, exportName)
				default:

					switch a.config.DefaultExport {
					case "desktop":
						filename, err = data.ExportTableToDesktop(a.currentTable.ID, a.config.Username, exportName)
					case "both":
						filename, err = data.ExportTableToLocation(a.currentTable.ID, a.config.Username, data.BothLocations, exportName)
					default:
						filename, err = data.ExportTable(a.currentTable.ID, a.config.Username, exportName)
					}
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
	title := Title.Copy().Width(a.width - 4).Render("Export Table")
	subtitle := Subtitle.Copy().Width(a.width - 4).Render(a.currentTable.Name)

	exportFileInput := Subtitle.Render("Export filename:") + "\n" + a.exportName.View()

	exportInfo := BoxStyle.Copy().Width(a.width - 6).Render(
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

	form := BoxStyle.Copy().Width(a.width - 6).Render(
		fmt.Sprintf("%s\n\n%s",
			exportFileInput,
			locationInfo,
		),
	)

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		exportInfo,
		form,
	)
}
