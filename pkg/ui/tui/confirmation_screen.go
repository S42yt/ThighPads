package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/database"
)

func (a *App) updateConfirmationScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":

			switch a.confirmationScreen {
			case DeleteTableConfirmation:
				err := database.DeleteTable(a.confirmDeleteTable)
				if err != nil {
					a.errorMsg = err.Error()
				} else {
					a.successMsg = "Table deleted successfully."
					a.loadTables()
				}
				a.screen = HomeScreen
				return a, nil
			}
		case "n", "esc":

			switch a.confirmationScreen {
			case DeleteTableConfirmation:
				a.screen = HomeScreen
			}
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) viewConfirmationScreen() string {
	var title, message string

	switch a.confirmationScreen {
	case DeleteTableConfirmation:
		title = Title.Render("Delete Table")

		for _, table := range a.tables {
			if table.ID == a.confirmDeleteTable {
				message = Warning.Render(fmt.Sprintf("Are you sure you want to delete table \"%s\"?\nThis will delete all entries and cannot be undone.", table.Name))
				break
			}
		}
	}

	help := HelpView(map[string]string{
		"y": "Yes, delete",
		"n": "No, cancel",
	})

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		BoxStyle.Render(message),
		help,
	)
}
