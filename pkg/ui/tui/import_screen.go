package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/data"
)

func (a *App) updateImportScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.importPathInput.Value() != "" {

				rawPath := a.importPathInput.Value()
				path := (rawPath)

				if _, err := os.Stat(path); os.IsNotExist(err) {

					if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(data.FileExtension)) {
						pathWithExt := path + data.FileExtension
						if _, err := os.Stat(pathWithExt); err == nil {
							path = pathWithExt
						} else {
							a.errorMsg = fmt.Sprintf("File does not exist: %s", (path))
							return a, nil
						}
					} else {
						a.errorMsg = fmt.Sprintf("File does not exist: %s", (path))
						return a, nil
					}
				}

				if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(data.FileExtension)) {
					a.errorMsg = "File must have " + data.FileExtension + " extension"
					return a, nil
				}

				err := data.ImportFile(path, a.config.Username)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

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
		}
	}

	a.importPathInput, cmd = a.importPathInput.Update(msg)
	return a, cmd
}

func (a *App) viewImportScreen() string {
	title := Title.Copy().Width(a.width - 4).Render("Import Table")

	importHelp := Subtle.Render("Examples: ~/Documents/my_table.thighpad, C:\\Users\\files\\table.thighpad")

	importInput := BoxStyle.Copy().Width(a.width - 6).Render(
		fmt.Sprintf("%s\n%s\n\n%s",
			Normal.Render("Enter path to .thighpad file:"),
			importHelp,
			a.importPathInput.View(),
		),
	)

	return fmt.Sprintf(
		"%s\n\n%s",
		title,
		importInput,
	)
}
