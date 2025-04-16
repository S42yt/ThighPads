package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
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

			a.entryContent = textarea.New()
			a.entryContent.Placeholder = "Enter your content here..."
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
		case "home", "g":
			a.entryViewport.GotoTop()
			return a, nil
		case "end", "G":
			a.entryViewport.GotoBottom()
			return a, nil
		case "pageup":
			a.entryViewport.HalfViewUp()
			return a, nil
		case "pagedown":
			a.entryViewport.HalfViewDown()
			return a, nil
		}
	}

	// Handle mouse wheel events
	a.entryViewport, cmd = a.entryViewport.Update(msg)
	return a, cmd
}

func (a *App) viewViewEntryScreen() string {
	// Create new viewport with appropriate dimensions
	viewportHeight := a.height - 16
	if viewportHeight < 5 {
		viewportHeight = 5 // Minimum reasonable height
	}

	a.entryViewport = viewport.New(a.width-6, viewportHeight)
	a.entryViewport.Style = BoxStyle

	// Process content with syntax highlighting if enabled
	content := a.currentEntry.Content
	if a.config.SyntaxHighlighting && len(a.config.EnabledSyntaxThemes) > 0 {
		// Split tags and clean them
		tags := []string{}
		if a.currentEntry.Tags != "" {
			for _, tag := range strings.Split(a.currentEntry.Tags, ",") {
				cleanTag := strings.ToLower(strings.TrimSpace(tag))
				if cleanTag != "" {
					tags = append(tags, cleanTag)
				}
			}
		}

		// Apply syntax highlighting based on entry tags
		if len(tags) > 0 {
			content = ApplyHighlighting(content, tags)
		}
	}

	// Set the processed content
	a.entryViewport.SetContent(content)

	title := Title.Copy().Width(a.width - 4).Render(a.currentEntry.Title)

	// Format tags for display
	tagDisplay := "Tags: "
	if a.currentEntry.Tags != "" {
		tagDisplay += a.currentEntry.Tags
	} else {
		tagDisplay += "(none)"
	}

	tags := Subtitle.Copy().Width(a.width - 4).Render(tagDisplay)
	date := Subtle.Copy().Width(a.width - 4).Render(
		"Created on " + a.currentEntry.CreatedAt.Format("Jan 02, 2006"))

	// Content with border
	contentView := a.entryViewport.View()

	// Show scroll indicators when needed
	scrollInfo := ""
	if a.entryViewport.TotalLineCount() > a.entryViewport.Height {
		scrollPercent := 0
		if a.entryViewport.TotalLineCount()-a.entryViewport.Height > 0 {
			scrollPercent = int(float64(a.entryViewport.YOffset) /
				float64(a.entryViewport.TotalLineCount()-a.entryViewport.Height) * 100)
		}

		// Add line numbers for reference
		scrollInfo = Subtle.Render(fmt.Sprintf("Scroll: %d%% (Line %d of %d lines)",
			scrollPercent, a.entryViewport.YOffset+1, a.entryViewport.TotalLineCount()))
	}

	navigationTips := Normal.Render("▲/▼: Scroll  PgUp/PgDn: Page up/down  g/G: Top/Bottom")

	return fmt.Sprintf(
		"%s\n%s\n%s\n\n%s\n\n%s\n%s",
		title,
		tags,
		date,
		contentView,
		scrollInfo,
		navigationTips,
	)
}
