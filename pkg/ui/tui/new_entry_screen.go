package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/models"
)

func (a *App) updateNewEntryScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				newEntry := models.Entry{
					TableID:   a.currentTable.ID,
					Title:     a.entryTitleInput.Value(),
					Tags:      a.entryTagsInput.Value(),
					Content:   a.entryContent.Value(),
					CreatedAt: time.Now(),
				}

				err := database.CreateEntry(&newEntry)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = TableScreen
				a.loadEntries()
				a.successMsg = "Entry created successfully."
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = TableScreen
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

func (a *App) viewNewEntryScreen() string {
	title := Title.Copy().Width(a.width - 4).Render("New Entry")
	subtitle := Subtitle.Copy().Width(a.width - 4).Render(a.currentTable.Name)

	availWidth := a.width - 6

	titleInput := Subtitle.Render("Title:") + "\n" + a.entryTitleInput.View()
	tagsInput := Subtitle.Render("Tags:") + "\n" + a.entryTagsInput.View()

	// Add focus indicator to show which field is currently active
	focusIndicator := ""
	if a.entryTitleInput.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing title...")
	} else if a.entryTagsInput.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing tags...")
	} else if a.entryContent.Focused() {
		focusIndicator = Subtitle.Foreground(accentColor).Render("Editing content... (Use Tab to switch fields, Ctrl+S to save)")
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

	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		title,
		subtitle,
		form,
	)
}
