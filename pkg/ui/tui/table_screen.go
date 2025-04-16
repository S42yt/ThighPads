package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/database"
)

func (a *App) updateTableScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			a.screen = NewEntryScreen
			a.entryTitleInput = TextInputField("Enter title")
			a.entryTagsInput = TextInputField("Enter tags (comma-separated)")
			a.entryContent = textarea.New()
			a.entryContent.Placeholder = "Enter your content here..."
			a.entryContent.SetWidth(a.width - 6)
			a.entryContent.SetHeight(a.height - 20)
			a.entryContent.Focus()
			return a, nil
		case "e":
			a.screen = ExportScreen
			a.exportName = TextInputField(a.currentTable.Name)
			a.exportLocation = -1
			return a, nil
		case "b":
			a.screen = HomeScreen
			a.loadTables()
			return a, nil
		case "d":
			if len(a.entries) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {

					if a.errorMsg == "confirm_delete" {
						err := database.DeleteEntry(selected.ID)
						if err != nil {
							a.errorMsg = err.Error()
						} else {
							a.successMsg = "Entry deleted successfully."
							a.loadEntries()
						}
						a.errorMsg = ""
						return a, nil
					} else {
						a.errorMsg = "confirm_delete"
						a.successMsg = "Press 'd' again to confirm deletion"
						return a, nil
					}
				}
			}
		case "q", "ctrl+c":
			return a, tea.Quit
		case "enter":
			if len(a.entries) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {
					for _, entry := range a.entries {
						if entry.ID == selected.ID {
							a.currentEntry = entry

							a.entryViewport.Width = a.width - 6
							a.entryViewport.Height = a.height - 16
							a.entryViewport.SetContent(entry.Content)
							a.entryViewport.GotoTop()

							a.screen = ViewEntryScreen
							return a, nil
						}
					}
				}
			}
		}
	}

	if len(a.entries) > 0 {
		a.list, cmd = a.list.Update(msg)
	}

	return a, cmd
}

func (a *App) viewTableScreen() string {
	title := Title.Copy().Width(a.width - 4).Render(a.currentTable.Name)
	subtitle := Subtitle.Copy().Width(a.width - 4).Render(fmt.Sprintf("Created by %s on %s",
		a.currentTable.Author,
		a.currentTable.CreatedAt.Format("Jan 02, 2006")))

	var content string
	if len(a.entries) == 0 {
		content = BoxStyle.Copy().Width(a.width - 4).Render(
			Normal.Render("This table is empty. Press 'n' to create your first entry."))
	} else {

		a.list.SetWidth(a.width - 6)
		a.list.SetHeight(a.height - 12)
		content = BoxStyle.Copy().Width(a.width - 4).Render(a.list.View())
	}

	if a.errorMsg == "confirm_delete" {
		warningBox := Warning.Copy().Width(a.width - 6).Render("Press 'd' again to confirm deletion")
		content = warningBox + "\n\n" + content
		a.errorMsg = "confirm_delete"
	}

	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		title,
		subtitle,
		content,
	)
}
