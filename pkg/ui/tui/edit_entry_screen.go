package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/database"
)

func (a *App) updateEditEntryScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if a.entryTitleInput.Focused() {
				a.entryTitleInput.Blur()
				a.entryTagsInput.Focus()
				a.entryContent.Blur()
			} else if a.entryTagsInput.Focused() {
				a.entryTitleInput.Blur()
				a.entryTagsInput.Blur()
				a.entryContent.Focus()
			} else {
				a.entryTitleInput.Focus()
				a.entryTagsInput.Blur()
				a.entryContent.Blur()
			}
			return a, nil
		case tea.KeyCtrlS:
			if a.entryTitleInput.Value() != "" {
				updatedEntry := a.currentEntry
				updatedEntry.Title = a.entryTitleInput.Value()
				updatedEntry.Tags = a.entryTagsInput.Value()
				updatedEntry.Content = a.entryContent.Value()

				err := database.UpdateEntry(&updatedEntry)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.currentEntry = updatedEntry
				a.screen = TableScreen
				a.loadEntries()
				a.successMsg = "Entry updated successfully."
				return a, nil
			} else {
				a.errorMsg = "Title cannot be empty"
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = ViewEntryScreen

			
			a.entryViewport.SetContent(a.currentEntry.Content)
			a.entryViewport.GotoTop()
			return a, nil
		case tea.KeyCtrlC:
			return a, tea.Quit
		}
	}

	
	if a.entryTitleInput.Focused() {
		a.entryTitleInput, cmd = a.entryTitleInput.Update(msg)
	} else if a.entryTagsInput.Focused() {
		a.entryTagsInput, cmd = a.entryTagsInput.Update(msg)
	} else {
		a.entryContent, cmd = a.entryContent.Update(msg)
	}

	return a, cmd
}

func (a *App) viewEditEntryScreen() string {
	title := Title.Copy().Width(a.width - 4).Render("Edit Entry")
	subtitle := Subtitle.Copy().Width(a.width - 4).Render(a.currentTable.Name)

	availWidth := a.width - 6 

	titleInput := Subtitle.Render("Title:") + "\n" + a.entryTitleInput.View()
	tagsInput := Subtitle.Render("Tags:") + "\n" + a.entryTagsInput.View()

	focusIndicator := ""
	if a.entryTitleInput.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing title...")
	} else if a.entryTagsInput.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing tags...")
	} else if a.entryContent.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing content... (Use arrow keys to navigate)")
	}

	content := Subtitle.Render("Content:") + "\n" + a.entryContent.View()

	form := BoxStyle.Copy().Width(availWidth).Render(
		fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
			titleInput,
			tagsInput,
			content,
			focusIndicator,
		),
	)

	help := HelpView(map[string]string{
		"Tab":    "Next field",
		"Ctrl+S": "Save changes",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		form,
		help,
	)
}
