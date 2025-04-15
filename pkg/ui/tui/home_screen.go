package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateHomeScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			a.screen = NewTableScreen
			a.tableNameInput = TextInputField("Enter table name")
			return a, nil
		case "i":
			a.screen = ImportScreen
			a.importPathInput = TextInputField("Enter path to .thighpad file")
			return a, nil
		case "q", "ctrl+c":
			return a, tea.Quit
		case "enter":
			if len(a.tables) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {
					for _, table := range a.tables {
						if table.ID == selected.ID {
							a.currentTable = table
							a.screen = TableScreen
							a.loadEntries()
							return a, nil
						}
					}
				}
			}
		}
	}

	if len(a.tables) > 0 {
		a.list, cmd = a.list.Update(msg)
	}

	return a, cmd
}

func (a *App) viewHomeScreen() string {
	title := Title.Copy().Width(a.width - 4).Render("ThighPads")
	subtitle := Subtitle.Copy().Width(a.width - 4).Render(fmt.Sprintf("Welcome, %s", a.config.Username))

	var content string
	if len(a.tables) == 0 {
		content = BoxStyle.Copy().Width(a.width - 4).Render(Normal.Render("You don't have any tables yet. Press 'n' to create your first one."))
	} else {
		content = BoxStyle.Copy().Width(a.width - 4).Render(a.list.View())
	}

	help := HelpView(map[string]string{
		"↑/↓":   "Navigate",
		"Enter": "Select table",
		"n":     "New table",
		"i":     "Import table",
		"q":     "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		content,
		help,
	)
}
