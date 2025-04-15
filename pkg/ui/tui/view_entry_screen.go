package tui

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateViewEntryScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			a.screen = EditEntryScreen
			a.entryTitleInput = TextInputField(a.currentEntry.Title)
			a.entryTagsInput = TextInputField(a.currentEntry.Tags)
			a.entryContent.SetValue(a.currentEntry.Content)
			a.entryContent.SetWidth(a.width - 6) 
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

	
	a.entryViewport, cmd = a.entryViewport.Update(msg)
	return a, cmd
}

func (a *App) viewViewEntryScreen() string {
	
	a.entryViewport = viewport.New(a.width-6, a.height-16)
	a.entryViewport.SetContent(a.currentEntry.Content)

	title := Title.Copy().Width(a.width - 4).Render(a.currentEntry.Title)
	tags := Subtitle.Copy().Width(a.width - 4).Render("Tags: " + a.currentEntry.Tags)
	date := Subtle.Copy().Width(a.width - 4).Render("Created on " + a.currentEntry.CreatedAt.Format("Jan 02, 2006"))

	content := BoxStyle.Width(a.width - 4).Render(a.entryViewport.View())

	scrollInfo := ""
	if a.entryViewport.TotalLineCount() > a.entryViewport.Height {
		scrollPercent := 0
		if a.entryViewport.TotalLineCount()-a.entryViewport.Height > 0 {
			scrollPercent = int(float64(a.entryViewport.YOffset) / float64(a.entryViewport.TotalLineCount()-a.entryViewport.Height) * 100)
		}
		scrollInfo = Subtle.Render(fmt.Sprintf("Scroll: %d%% (%d of %d lines)",
			scrollPercent, a.entryViewport.YOffset+1, a.entryViewport.TotalLineCount()))
	}

	help := HelpView(map[string]string{
		"↑/↓": "Scroll",
		"e":   "Edit",
		"c":   "Copy to clipboard",
		"b":   "Back",
		"q":   "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n%s\n\n%s\n%s\n\n%s",
		title,
		tags,
		date,
		content,
		scrollInfo,
		help,
	)
}
