package tui

import (
	"fmt"
	"io"
	"sort"
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

	width := m.Width() - 4
	if width < 10 {
		width = 10
	}

	var title, desc string
	if index == m.Index() {
		title = Selected.Copy().Width(width).Render(truncateString(i.Title, width-4))
		desc = Selected.Copy().Width(width).Render(truncateString(i.Description, width-4))
	} else {
		title = Unselected.Copy().Width(width).Render(truncateString(i.Title, width-4))
		desc = Subtle.Copy().Width(width).Render(truncateString(i.Description, width-4))
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func truncateString(s string, max int) string {
	if max <= 3 {
		return s
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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

	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.Styles.PaginationStyle = Subtle

	l.Styles.HelpStyle = Subtle

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
	return BoxStyle.Copy().Border(lipgloss.NormalBorder()).Render(helpText)
}

func ErrorView(message string) string {
	return Error.Copy().Width(40).Render("Error: " + message)
}

func SuccessView(message string) string {
	return Success.Copy().Width(40).Render(message)
}

func CenterView(content string, width int) string {
	if width < 10 {
		width = 10
	}
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, content)
}

func FixedToolbarView(keys map[string]string, width int) string {
	var helpEntries []string

	priorityKeys := []string{"Ctrl+C", "Esc", "Enter", "Tab", "↑/↓"}
	remainingKeys := make(map[string]string)

	for _, key := range priorityKeys {
		if description, exists := keys[key]; exists {
			entry := fmt.Sprintf("%s: %s",
				Subtle.Render(key),
				Normal.Render(description))
			helpEntries = append(helpEntries, entry)
			delete(keys, key)
		}
	}

	for key, description := range keys {
		remainingKeys[key] = description
	}

	availableWidth := width - 10
	currentWidth := 0

	for _, entry := range helpEntries {
		currentWidth += lipgloss.Width(entry) + 3
	}

	for key, description := range remainingKeys {
		entry := fmt.Sprintf("%s: %s",
			Subtle.Render(key),
			Normal.Render(description))
		entryWidth := lipgloss.Width(entry) + 3

		if currentWidth+entryWidth < availableWidth {
			helpEntries = append(helpEntries, entry)
			currentWidth += entryWidth
		}
	}

	if len(remainingKeys) > 0 && len(helpEntries) < len(priorityKeys)+len(remainingKeys) {
		helpEntries = append(helpEntries, Subtle.Render("...more"))
	}

	helpText := strings.Join(helpEntries, " \u2022 ")
	return BoxStyle.Copy().Width(width - 4).Border(lipgloss.NormalBorder()).Render(helpText)
}

// FullHelpView renders a comprehensive help screen showing all key bindings
func FullHelpView(keys map[string]string, width, height int, inTextInput bool) string {
	var sections []string
	sections = append(sections, Subtitle.Render("Available Commands:"))

	// Sort keys for consistent display
	var keyList []string
	for key := range keys {
		keyList = append(keyList, key)
	}
	sort.Strings(keyList)

	// Calculate columns
	numColumns := 1
	if width >= 100 {
		numColumns = 3
	} else if width >= 60 {
		numColumns = 2
	}

	columnWidth := (width - 10) / numColumns
	columnItems := (len(keys) + numColumns - 1) / numColumns

	// Organize keys into columns
	columns := make([][]string, numColumns)
	columnCount := 0
	itemCount := 0

	for _, key := range keyList {
		keyDisplay := fmt.Sprintf("  %s: %s",
			Selected.Copy().Padding(0, 0).Render(key),
			Normal.Render(keys[key]))

		columns[columnCount] = append(columns[columnCount], keyDisplay)
		itemCount++

		if itemCount >= columnItems {
			columnCount++
			itemCount = 0
			if columnCount >= numColumns {
				break
			}
		}
	}

	// Build rows with proper spacing
	var rows []string
	maxRows := 0
	for _, col := range columns {
		if len(col) > maxRows {
			maxRows = len(col)
		}
	}

	for i := 0; i < maxRows; i++ {
		var rowItems []string
		for j := 0; j < numColumns; j++ {
			if j < len(columns) && i < len(columns[j]) {
				// Pad each entry to column width
				rowItems = append(rowItems,
					lipgloss.NewStyle().Width(columnWidth).Render(columns[j][i]))
			} else {
				rowItems = append(rowItems, "")
			}
		}
		rows = append(rows, strings.Join(rowItems, " "))
	}

	sections = append(sections, strings.Join(rows, "\n"))

	// Show appropriate key to toggle help view based on mode
	if inTextInput {
		sections = append(sections, Subtle.Render("\nPress 'Ctrl+h' to toggle this help screen"))
	} else {
		sections = append(sections, Subtle.Render("\nPress 'h' to toggle this help screen"))
	}

	return BoxStyle.Copy().Render(strings.Join(sections, "\n"))
}
