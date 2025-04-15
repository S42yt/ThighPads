package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/data"
)

func (a *App) updateImportScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.importPathInput.Value() != "" {
				path := a.importPathInput.Value()

				if _, err := os.Stat(path); os.IsNotExist(err) {
					a.errorMsg = "File does not exist"
					return a, nil
				}

				if !strings.HasSuffix(path, data.FileExtension) {
					a.errorMsg = "File must have " + data.FileExtension + " extension"
					return a, nil
				}

				a.successMsg = "Importing table. Please wait..."

				err := data.ImportFile(path, a.config.Username)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.saveRecentImport(path)

				a.screen = HomeScreen
				a.loadTables()
				a.successMsg = "Table imported successfully."
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = HomeScreen
			return a, nil
		case tea.KeyCtrlC:
			return a, tea.Quit

		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] >= '1' && msg.Runes[0] <= '9' {
				idx := int(msg.Runes[0] - '1')
				if idx < len(a.recentImports) {
					a.importPathInput.SetValue(a.recentImports[idx])
					return a, nil
				}
			}
		}
	}

	a.importPathInput, cmd = a.importPathInput.Update(msg)
	return a, cmd
}

func (a *App) viewImportScreen() string {
	title := Title.Render("Import Table")

	importInput := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Enter path to .thighpad file:"),
			a.importPathInput.View(),
		),
	)

	recentImportsView := ""
	if len(a.recentImports) > 0 {
		recentImportsList := []string{"Recent imports (press number to select):"}
		for i, path := range a.recentImports {
			if i >= 9 {
				break
			}
			baseName := filepath.Base(path)
			recentImportsList = append(recentImportsList, fmt.Sprintf("%d. %s", i+1, baseName))
		}
		recentImportsView = "\n\n" + BoxStyle.Render(Normal.Render(strings.Join(recentImportsList, "\n")))
	}

	desktopTip := Subtle.Render("\nTip: Check your Desktop for ThighPads Exports folder")

	help := HelpView(map[string]string{
		"Enter":  "Import table",
		"Esc":    "Cancel",
		"1-9":    "Select recent import",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n\n%s%s%s\n\n%s",
		title,
		importInput,
		recentImportsView,
		desktopTip,
		help,
	)
}

func (a *App) loadRecentImports() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return
	}

	recentImportsFile := filepath.Join(configPath, "recent_imports.txt")
	data, err := os.ReadFile(recentImportsFile)
	if err == nil {
		a.recentImports = strings.Split(strings.TrimSpace(string(data)), "\n")

		var validPaths []string
		for _, path := range a.recentImports {
			if _, err := os.Stat(path); err == nil {
				validPaths = append(validPaths, path)
			}
		}
		a.recentImports = validPaths
	}
}

func (a *App) saveRecentImport(path string) {

	filteredImports := []string{path}
	for _, existingPath := range a.recentImports {
		if existingPath != path && len(filteredImports) < 10 {
			filteredImports = append(filteredImports, existingPath)
		}
	}
	a.recentImports = filteredImports

	configPath, err := config.GetConfigPath()
	if err != nil {
		return
	}

	recentImportsFile := filepath.Join(configPath, "recent_imports.txt")
	_ = os.WriteFile(recentImportsFile, []byte(strings.Join(a.recentImports, "\n")), 0644)
}
