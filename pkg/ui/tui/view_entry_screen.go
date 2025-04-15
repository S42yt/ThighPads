package tui

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateViewEntryScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			a.screen = EditEntryScreen
			a.entryTitleInput = TextInputField(a.currentEntry.Title)
			a.entryTagsInput = TextInputField(a.currentEntry.Tags)
			a.entryContent = textarea.New()
			a.entryContent.SetValue(a.currentEntry.Content)
			a.entryContent.SetWidth(a.width - 10)
			a.entryContent.SetHeight(a.height - 20)
			a.entryContent.Focus()
			return a, nil
		case "c":

			err := clipboard.WriteAll(a.currentEntry.Content)
			if err != nil {
				a.errorMsg = "Failed to copy to clipboard: " + err.Error()
			} else {
				a.successMsg = "Entry content copied to clipboard."
			}
			return a, nil
		case "b":
			a.screen = TableScreen
			return a, nil
		case "q", "ctrl+c", "esc":
			return a, tea.Quit
		}
	}

	return a, nil
}

func (a *App) viewViewEntryScreen() string {
	title := Title.Render(a.currentEntry.Title)
	tags := Subtitle.Render("Tags: " + a.currentEntry.Tags)
	date := Subtle.Render("Created on " + a.currentEntry.CreatedAt.Format("Jan 02, 2006"))

	content := BoxStyle.Render(Normal.Render(a.currentEntry.Content))

	help := HelpView(map[string]string{
		"e": "Edit",
		"c": "Copy to clipboard",
		"b": "Back",
		"q": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n%s\n\n%s\n\n%s",
		title,
		tags,
		date,
		content,
		help,
	)
}
