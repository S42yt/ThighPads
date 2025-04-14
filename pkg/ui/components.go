package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/s42yt/thighpads/pkg/models"
)

type Item struct {
	title    string
	Desc     string
	Data     interface{}
	MetaText string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.Desc }
func (i Item) FilterValue() string { return i.title }

func NewTableItem(table *models.Table) Item {
	return Item{
		title:    table.Name,
		Desc:     fmt.Sprintf("Entries: %d - Author: %s", len(table.Entries), table.Author),
		Data:     table,
		MetaText: shortTimeString(table.UpdatedAt),
	}
}

func NewEntryItem(entry models.Entry) Item {
	desc := entry.Content
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	tags := ""
	if len(entry.Tags) > 0 {
		tags = "[" + strings.Join(entry.Tags, ", ") + "]"
	}
	return Item{
		title:    entry.Title,
		Desc:     desc,
		Data:     entry,
		MetaText: tags,
	}
}

func NewSearchResultItem(result models.SearchResult) Item {
	desc := result.Entry.Content
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	tags := ""
	if len(result.Entry.Tags) > 0 {
		tags = "[" + strings.Join(result.Entry.Tags, ", ") + "]"
	}
	return Item{
		title:    fmt.Sprintf("%s - %s", result.TableName, result.Entry.Title),
		Desc:     desc,
		Data:     result,
		MetaText: tags,
	}
}

type InputForm struct {
	Title     string
	Labels    []string
	Inputs    []textinput.Model
	FocusedID int
	theme     *Theme
}

func NewInputForm(theme *Theme, title string, labels []string, initialValues []string) InputForm {
	inputs := make([]textinput.Model, len(labels))
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = "Enter " + strings.ToLower(labels[i])
		inputs[i].Width = 40

		if i < len(initialValues) && initialValues[i] != "" {
			inputs[i].SetValue(initialValues[i])
		}
	}

	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return InputForm{
		Title:     title,
		Labels:    labels,
		Inputs:    inputs,
		FocusedID: 0,
		theme:     theme,
	}
}

func (f *InputForm) NextInput() {
	if f.FocusedID >= 0 {
		f.Inputs[f.FocusedID].Blur()
	}
	f.FocusedID = (f.FocusedID + 1) % len(f.Inputs)
	f.Inputs[f.FocusedID].Focus()
}

func (f *InputForm) PrevInput() {
	if f.FocusedID >= 0 {
		f.Inputs[f.FocusedID].Blur()
	}
	f.FocusedID = (f.FocusedID - 1 + len(f.Inputs)) % len(f.Inputs)
	f.Inputs[f.FocusedID].Focus()
}

func (f *InputForm) GetValues() []string {
	values := make([]string, len(f.Inputs))
	for i, input := range f.Inputs {
		values[i] = input.Value()
	}
	return values
}

func (f *InputForm) View() string {
	var sb strings.Builder

	sb.WriteString(f.theme.AppTitle.Render(f.Title) + "\n\n")

	for i, input := range f.Inputs {
		sb.WriteString(f.theme.Label.Render(f.Labels[i]) + "\n")
		sb.WriteString(input.View() + "\n\n")
	}

	sb.WriteString(f.theme.InfoText.Render("Tab: Next field • Enter: Submit • Esc: Cancel"))

	return f.theme.BoxStyle.Render(sb.String())
}

type StatusMsg struct {
	Text    string
	IsError bool
}

type CustomItemDelegate struct {
	theme *Theme
}

func NewCustomItemDelegate(theme *Theme) list.ItemDelegate {
	return &CustomItemDelegate{theme: theme}
}

func (d CustomItemDelegate) Height() int {
	return 2
}

func (d CustomItemDelegate) Spacing() int {
	return 1
}

func (d CustomItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	var cmd tea.Cmd
	return cmd
}

func (d CustomItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(Item)
	if !ok {
		_, _ = fmt.Fprint(w, "Error: item is not a valid Item")
		return
	}

	title := i.Title()
	desc := i.Description()
	metaText := i.MetaText

	var titleStyle, descStyle lipgloss.Style
	if index == m.Index() {
		titleStyle = d.theme.SelectedItem.Copy().Width(m.Width())
		descStyle = d.theme.SelectedItem.Copy().Width(m.Width())
	} else {
		titleStyle = d.theme.ListItem.Copy().Width(m.Width())
		descStyle = d.theme.ListItem.Copy().Width(m.Width())
	}

	metaTextStyle := d.theme.InfoText

	if metaText != "" {
		titleWidth := lipgloss.Width(title)
		metaWidth := lipgloss.Width(metaText)

		// Only show meta text if there's enough space
		if titleWidth+metaWidth+2 < m.Width() {
			titleLine := titleStyle.Render(title)
			metaLine := metaTextStyle.Render(metaText)
			padding := m.Width() - titleWidth - metaWidth
			titleLine = titleLine + strings.Repeat(" ", padding) + metaLine
			_, _ = fmt.Fprint(w, titleLine+"\n"+descStyle.Render(desc))
			return
		}
	}

	_, _ = fmt.Fprint(w, titleStyle.Render(title)+"\n"+descStyle.Render(desc))
}

type (
	errMsg struct{ err error }

	statusMsg struct {
		Text    string
		IsError bool
	}

	tableSelected struct{ name string }

	entrySelected struct{ ID string }

	formSubmitted struct{ values []string }

	formCancelled struct{}

	confirmAction struct{ confirmed bool }
)

func showError(err error) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			Text:    err.Error(),
			IsError: true,
		}
	}
}

func showStatus(text string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			Text:    text,
			IsError: isError,
		}
	}
}

func backToMainCmd() tea.Cmd {
	return func() tea.Msg {
		return tableSelected{name: ""}
	}
}

func shortTimeString(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("Jan _2")
}

var keys = struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Help     key.Binding
	Quit     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Create   key.Binding
	NewEntry key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Import   key.Binding
	Export   key.Binding
	Search   key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
}{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h", "backspace", "esc"),
		key.WithHelp("←/h/esc", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "forward"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Create: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new table"),
	),
	NewEntry: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add entry"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit entry"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Import: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "import"),
	),
	Export: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "export"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
}
