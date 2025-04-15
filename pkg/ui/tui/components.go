package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Selectable struct {
	Title       string
	Description string
	ID          uint
}

func (i Selectable) FilterValue() string { return i.Title }

type ItemDelegate struct{}

func (d ItemDelegate) Height() int { return 2 }

func (d ItemDelegate) Spacing() int { return 1 }

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Selectable)
	if !ok {
		return
	}

	var title, desc string
	if index == m.Index() {
		title = Selected.Render(i.Title)
		desc = Selected.Render(i.Description)
	} else {
		title = Unselected.Render(i.Title)
		desc = Subtle.Render(i.Description)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func TextInputField(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.PromptStyle = Subtitle
	ti.TextStyle = Normal
	ti.Focus()
	return ti
}

func SelectableList(title string, items []list.Item, width, height int) list.Model {
	delegate := ItemDelegate{}

	l := list.New(items, delegate, width, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = Title
	l.Styles.FilterPrompt = Subtitle
	l.Styles.FilterCursor = Subtitle

	return l
}

func HelpView(keys map[string]string) string {
	var helpEntries []string

	for key, description := range keys {
		entry := fmt.Sprintf("%s: %s",
			Subtle.Render(key),
			Normal.Render(description))
		helpEntries = append(helpEntries, entry)
	}

	helpText := strings.Join(helpEntries, " • ")
	return BoxStyle.Render(helpText)
}

func ErrorView(message string) string {
	return Error.Render("Error: " + message)
}

func SuccessView(message string) string {
	return Success.Render(message)
}

func CenterView(content string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, content)
}
