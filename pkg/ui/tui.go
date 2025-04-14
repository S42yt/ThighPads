package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/app"
	"github.com/s42yt/thighpads/pkg/models"
)

type Screen int

const (
	ScreenMain Screen = iota
	ScreenTableDetails
	ScreenEntryDetails
	ScreenCreateTable
	ScreenCreateEntry
	ScreenEditEntry
	ScreenImportTable
	ScreenExportTable
	ScreenConfirm
	ScreenSearch
	ScreenSearchResults
)

type TUI struct {
	app            *app.App
	theme          *Theme
	screen         Screen
	width          int
	height         int
	tableList      list.Model
	entryList      list.Model
	searchList     list.Model
	activeTable    *models.Table
	activeEntry    *models.Entry
	form           InputForm
	textarea       textarea.Model
	textinput      textinput.Model
	searchInput    textinput.Model
	searchResults  []models.SearchResult
	status         StatusMsg
	confirmMsg     string
	confirmFunc    func(bool) tea.Cmd
	showHelp       bool
	quitting       bool
	importPath     string
	exportPath     string
	previousScreen Screen
}

// NewTUI creates a new TUI instance
func NewTUI(app *app.App) *TUI {
	theme := NewTheme()

	// Default sizes
	width := 80
	height := 24

	// Create custom delegate
	delegate := NewCustomItemDelegate(theme)

	// Initialize empty lists with custom delegate
	tableList := list.New([]list.Item{}, delegate, width, height-6)
	tableList.Title = "Tables"
	tableList.Styles.Title = theme.Title
	tableList.Styles.PaginationStyle = theme.InfoText
	tableList.SetShowStatusBar(false)
	tableList.SetShowHelp(false)

	entryList := list.New([]list.Item{}, delegate, width, height-6)
	entryList.Title = "Entries"
	entryList.Styles.Title = theme.Title
	entryList.Styles.PaginationStyle = theme.InfoText
	entryList.SetShowStatusBar(false)
	entryList.SetShowHelp(false)

	searchList := list.New([]list.Item{}, delegate, width, height-6)
	searchList.Title = "Search Results"
	searchList.Styles.Title = theme.Title
	searchList.Styles.PaginationStyle = theme.InfoText
	searchList.SetShowStatusBar(false)
	searchList.SetShowHelp(false)

	// Default text area for content
	ta := textarea.New()
	ta.Placeholder = "Enter content here..."
	ta.ShowLineNumbers = true
	ta.Focus()

	// Default text input
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 100
	ti.Width = width - 4

	// Search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search for entries..."
	searchInput.CharLimit = 100
	searchInput.Width = width - 4
	searchInput.Focus()

	return &TUI{
		app:         app,
		theme:       theme,
		screen:      ScreenMain,
		width:       width,
		height:      height,
		tableList:   tableList,
		entryList:   entryList,
		searchList:  searchList,
		textarea:    ta,
		textinput:   ti,
		searchInput: searchInput,
		showHelp:    false,
	}
}

func (t *TUI) Init() tea.Cmd {
	return tea.Batch(
		t.loadTables(),
	)
}

func (t *TUI) loadTables() tea.Cmd {
	return func() tea.Msg {
		tables := t.app.GetAllTables()
		items := make([]list.Item, len(tables))
		for i, table := range tables {
			items[i] = NewTableItem(table)
		}
		return tea.Batch(
			func() tea.Msg {
				t.tableList.SetItems(items)
				return nil
			},
		)
	}
}

func (t *TUI) loadEntries() tea.Cmd {
	if t.activeTable == nil {
		return nil
	}

	return func() tea.Msg {
		items := make([]list.Item, len(t.activeTable.Entries))
		for i, entry := range t.activeTable.Entries {
			items[i] = NewEntryItem(entry)
		}
		return tea.Batch(
			func() tea.Msg {
				t.entryList.SetItems(items)
				t.entryList.Title = fmt.Sprintf("Entries in '%s'", t.activeTable.Name)
				return nil
			},
		)
	}
}

func (t *TUI) createTableForm() tea.Cmd {
	form := NewInputForm(t.theme, "Create New Table", []string{"Table Name"}, []string{})
	return func() tea.Msg {
		t.form = form
		t.screen = ScreenCreateTable
		return nil
	}
}

func (t *TUI) createEntryForm() tea.Cmd {
	form := NewInputForm(t.theme, "Create New Entry", []string{"Title", "Tags (comma separated)"}, []string{})
	t.textarea = textarea.New()
	t.textarea.Placeholder = "Enter content here..."
	t.textarea.ShowLineNumbers = true
	t.textarea.Focus()

	return func() tea.Msg {
		t.form = form
		t.screen = ScreenCreateEntry
		return nil
	}
}

func (t *TUI) editEntryForm() tea.Cmd {
	if t.activeEntry == nil {
		return nil
	}

	entry := *t.activeEntry
	form := NewInputForm(t.theme, "Edit Entry", []string{"Title", "Tags (comma separated)"},
		[]string{entry.Title, strings.Join(entry.Tags, ", ")})

	t.textarea = textarea.New()
	t.textarea.SetValue(entry.Content)
	t.textarea.ShowLineNumbers = true
	t.textarea.Focus()

	return func() tea.Msg {
		t.form = form
		t.screen = ScreenEditEntry
		return nil
	}
}

func (t *TUI) importTableForm() tea.Cmd {
	form := NewInputForm(t.theme, "Import Table", []string{"Path to .thighpad file"}, []string{})
	return func() tea.Msg {
		t.form = form
		t.screen = ScreenImportTable
		return nil
	}
}

func (t *TUI) exportTableForm() tea.Cmd {
	if t.activeTable == nil {
		return showStatus("No table selected", true)
	}

	form := NewInputForm(t.theme, "Export Table",
		[]string{fmt.Sprintf("Export '%s' to path:", t.activeTable.Name)},
		[]string{t.app.Config.GetTablePath(t.activeTable.Name)})

	return func() tea.Msg {
		t.form = form
		t.screen = ScreenExportTable
		return nil
	}
}

func (t *TUI) showSearchScreen() tea.Cmd {
	t.previousScreen = t.screen
	t.searchInput.Focus()
	t.searchInput.SetValue("")

	return func() tea.Msg {
		t.screen = ScreenSearch
		return nil
	}
}

func (t *TUI) performSearch() tea.Cmd {
	query := t.searchInput.Value()
	if strings.TrimSpace(query) == "" {
		return showStatus("Please enter a search query", true)
	}

	return func() tea.Msg {
		results, err := t.app.SearchEntries(query)
		if err != nil {
			return errMsg(err)
		}

		t.searchResults = results
		items := make([]list.Item, len(results))
		for i, result := range results {
			items[i] = NewSearchResultItem(result)
		}

		t.searchList.SetItems(items)
		t.searchList.Title = fmt.Sprintf("Search Results for '%s'", query)
		t.screen = ScreenSearchResults

		if len(results) == 0 {
			return statusMsg{
				Text:    fmt.Sprintf("No results found for '%s'", query),
				IsError: false,
			}
		}

		return statusMsg{
			Text:    fmt.Sprintf("Found %d results for '%s'", len(results), query),
			IsError: false,
		}
	}
}

func (t *TUI) showConfirm(msg string, action func(bool) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		t.confirmMsg = msg
		t.confirmFunc = action
		t.screen = ScreenConfirm
		return nil
	}
}

func (t *TUI) deleteTable() tea.Cmd {
	if t.activeTable == nil {
		return showStatus("No table selected", true)
	}

	tableName := t.activeTable.Name

	return t.showConfirm(
		fmt.Sprintf("Delete table '%s'? This cannot be undone. (y/n)", tableName),
		func(confirmed bool) tea.Cmd {
			if !confirmed {
				return backToMainCmd()
			}

			err := t.app.DeleteTable(tableName)
			if err != nil {
				return showError(err)
			}

			t.activeTable = nil
			return tea.Batch(
				showStatus(fmt.Sprintf("Table '%s' deleted", tableName), false),
				backToMainCmd(),
				t.loadTables(),
			)
		},
	)
}

func (t *TUI) deleteEntry() tea.Cmd {
	if t.activeTable == nil || t.activeEntry == nil {
		return showStatus("No entry selected", true)
	}

	tableName := t.activeTable.Name
	entryID := t.activeEntry.ID
	entryTitle := t.activeEntry.Title

	return t.showConfirm(
		fmt.Sprintf("Delete entry '%s'? This cannot be undone. (y/n)", entryTitle),
		func(confirmed bool) tea.Cmd {
			if !confirmed {
				return nil
			}

			err := t.app.DeleteEntry(tableName, entryID)
			if err != nil {
				return showError(err)
			}

			table, err := t.app.GetTable(tableName)
			if err != nil {
				return showError(err)
			}

			t.activeTable = table
			t.activeEntry = nil

			return tea.Batch(
				showStatus(fmt.Sprintf("Entry '%s' deleted", entryTitle), false),
				t.loadEntries(),
			)
		},
	)
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Global shortcuts
		case key.Matches(msg, keys.Quit):
			t.quitting = true
			return t, tea.Quit

		case key.Matches(msg, keys.Help):
			t.showHelp = !t.showHelp
			return t, nil

		case key.Matches(msg, keys.Search):
			// Don't trigger search if we're in a form or other input mode
			if t.screen != ScreenCreateTable &&
				t.screen != ScreenCreateEntry &&
				t.screen != ScreenEditEntry &&
				t.screen != ScreenImportTable &&
				t.screen != ScreenExportTable &&
				t.screen != ScreenConfirm &&
				t.screen != ScreenSearch {
				return t, t.showSearchScreen()
			}
		}

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.tableList.SetSize(msg.Width, msg.Height-6)
		t.entryList.SetSize(msg.Width, msg.Height-6)
		t.searchList.SetSize(msg.Width, msg.Height-6)

	case statusMsg:
		t.status = StatusMsg(msg)

	case tableSelected:
		if msg.name == "" {
			t.activeTable = nil
			t.screen = ScreenMain
		} else {
			table, err := t.app.GetTable(msg.name)
			if err != nil {
				return t, showError(err)
			}
			t.activeTable = table
			t.screen = ScreenTableDetails
			cmds = append(cmds, t.loadEntries())
		}
		return t, tea.Batch(cmds...)

	case entrySelected:
		if t.activeTable != nil {
			entry, found := t.activeTable.GetEntry(msg.ID)
			if found {
				t.activeEntry = &entry
				t.screen = ScreenEntryDetails
			}
		}
		return t, nil

	case searchResultSelected:
		// Get the table from the search result
		table, err := t.app.GetTable(msg.tableName)
		if err != nil {
			return t, showError(err)
		}

		// Set active table and load entries
		t.activeTable = table

		// Find the entry within the table
		entry, found := t.activeTable.GetEntry(msg.entryID)
		if found {
			t.activeEntry = &entry
			t.screen = ScreenEntryDetails
		} else {
			t.screen = ScreenTableDetails
			cmds = append(cmds, t.loadEntries())
		}
		return t, tea.Batch(cmds...)

	case formSubmitted:
		switch t.screen {
		case ScreenCreateTable:
			if len(msg.values) > 0 && msg.values[0] != "" {
				table, err := t.app.CreateTable(msg.values[0])
				if err != nil {
					return t, showError(err)
				}
				t.activeTable = table
				t.screen = ScreenTableDetails
				return t, tea.Batch(
					showStatus(fmt.Sprintf("Table '%s' created", table.Name), false),
					t.loadTables(),
					t.loadEntries(),
				)
			}

		case ScreenCreateEntry:
			if t.activeTable != nil && len(msg.values) >= 2 {
				title := msg.values[0]
				tagStr := msg.values[1]
				tags := []string{}

				if tagStr != "" {
					for _, tag := range strings.Split(tagStr, ",") {
						tags = append(tags, strings.TrimSpace(tag))
					}
				}

				content := t.textarea.Value()

				err := t.app.AddEntry(t.activeTable.Name, title, content, tags)
				if err != nil {
					return t, showError(err)
				}

				table, err := t.app.GetTable(t.activeTable.Name)
				if err != nil {
					return t, showError(err)
				}

				t.activeTable = table
				t.screen = ScreenTableDetails
				return t, tea.Batch(
					showStatus("Entry created", false),
					t.loadEntries(),
				)
			}

		case ScreenEditEntry:
			if t.activeTable != nil && t.activeEntry != nil && len(msg.values) >= 2 {
				title := msg.values[0]
				tagStr := msg.values[1]
				tags := []string{}

				if tagStr != "" {
					for _, tag := range strings.Split(tagStr, ",") {
						tags = append(tags, strings.TrimSpace(tag))
					}
				}

				content := t.textarea.Value()

				err := t.app.UpdateEntry(
					t.activeTable.Name,
					t.activeEntry.ID,
					title,
					content,
					tags,
				)
				if err != nil {
					return t, showError(err)
				}

				table, err := t.app.GetTable(t.activeTable.Name)
				if err != nil {
					return t, showError(err)
				}

				entry, found := table.GetEntry(t.activeEntry.ID)
				if found {
					t.activeEntry = &entry
				}

				t.activeTable = table
				t.screen = ScreenEntryDetails
				return t, tea.Batch(
					showStatus("Entry updated", false),
					t.loadEntries(),
				)
			}

		case ScreenImportTable:
			if len(msg.values) > 0 && msg.values[0] != "" {
				path := msg.values[0]
				table, err := t.app.ImportTable(path)
				if err != nil {
					return t, showError(err)
				}

				t.activeTable = table
				t.screen = ScreenTableDetails
				return t, tea.Batch(
					showStatus(fmt.Sprintf("Table '%s' imported", table.Name), false),
					t.loadTables(),
					t.loadEntries(),
				)
			}

		case ScreenExportTable:
			if t.activeTable != nil && len(msg.values) > 0 && msg.values[0] != "" {
				path := msg.values[0]
				err := t.app.ExportTable(t.activeTable.Name, path)
				if err != nil {
					return t, showError(err)
				}

				t.screen = ScreenTableDetails
				return t, showStatus(fmt.Sprintf("Table '%s' exported to %s", t.activeTable.Name, path), false)
			}
		}

	case formCancelled:
		switch t.screen {
		case ScreenCreateTable, ScreenImportTable, ScreenExportTable:
			t.screen = ScreenMain
		case ScreenCreateEntry, ScreenEditEntry:
			if t.activeTable != nil {
				t.screen = ScreenTableDetails
			} else {
				t.screen = ScreenMain
			}
		case ScreenSearch:
			t.screen = t.previousScreen
		}
		return t, nil

	case confirmAction:
		if t.confirmFunc != nil {
			cmd = t.confirmFunc(msg.confirmed)
			t.confirmFunc = nil
			return t, cmd
		}
	}

	switch t.screen {
	case ScreenMain:
		t.tableList, cmd = t.tableList.Update(msg)
		cmds = append(cmds, cmd)

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Enter) || key.Matches(msg, keys.Right):
				if i, ok := t.tableList.SelectedItem().(Item); ok {
					if table, ok := i.Data.(*models.Table); ok {
						t.activeTable = table
						t.screen = ScreenTableDetails
						cmds = append(cmds, t.loadEntries())
					}
				}

			case key.Matches(msg, keys.Create):
				cmds = append(cmds, t.createTableForm())

			case key.Matches(msg, keys.Import):
				cmds = append(cmds, t.importTableForm())
			}
		}

	case ScreenTableDetails:
		t.entryList, cmd = t.entryList.Update(msg)
		cmds = append(cmds, cmd)

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Enter) || key.Matches(msg, keys.Right):
				if i, ok := t.entryList.SelectedItem().(Item); ok {
					if entry, ok := i.Data.(models.Entry); ok {
						t.activeEntry = &entry
						t.screen = ScreenEntryDetails
					}
				}

			case key.Matches(msg, keys.Back) || key.Matches(msg, keys.Left):
				t.activeTable = nil
				t.screen = ScreenMain

			case key.Matches(msg, keys.NewEntry):
				cmds = append(cmds, t.createEntryForm())

			case key.Matches(msg, keys.Delete):
				cmds = append(cmds, t.deleteTable())

			case key.Matches(msg, keys.Export):
				cmds = append(cmds, t.exportTableForm())
			}
		}

	case ScreenEntryDetails:

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Back) || key.Matches(msg, keys.Left):
				t.activeEntry = nil
				t.screen = ScreenTableDetails

			case key.Matches(msg, keys.Edit):
				cmds = append(cmds, t.editEntryForm())

			case key.Matches(msg, keys.Delete):
				cmds = append(cmds, t.deleteEntry())
			}
		}

	case ScreenCreateTable, ScreenImportTable, ScreenExportTable:

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				values := t.form.GetValues()
				return t, func() tea.Msg {
					return formSubmitted{values: values}
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				return t, func() tea.Msg {
					return formCancelled{}
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
				t.form.NextInput()

			case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
				t.form.PrevInput()

			default:
				for i := range t.form.Inputs {
					if i == t.form.FocusedID {
						t.form.Inputs[i], cmd = t.form.Inputs[i].Update(msg)
						cmds = append(cmds, cmd)
					}
				}
			}
		}

	case ScreenCreateEntry, ScreenEditEntry:

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+enter"))):
				values := t.form.GetValues()
				return t, func() tea.Msg {
					return formSubmitted{values: values}
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				return t, func() tea.Msg {
					return formCancelled{}
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+t"))):

				if t.form.FocusedID >= 0 {
					for i := range t.form.Inputs {
						t.form.Inputs[i].Blur()
					}
					t.form.FocusedID = -1
					t.textarea.Focus()
				} else {
					t.textarea.Blur()
					t.form.FocusedID = 0
					t.form.Inputs[0].Focus()
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
				if t.form.FocusedID >= 0 {
					t.form.NextInput()
				}

			default:
				if t.form.FocusedID >= 0 {
					for i := range t.form.Inputs {
						if i == t.form.FocusedID {
							t.form.Inputs[i], cmd = t.form.Inputs[i].Update(msg)
							cmds = append(cmds, cmd)
						}
					}
				} else {
					t.textarea, cmd = t.textarea.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}

	case ScreenSearch:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Enter):
				return t, t.performSearch()

			case key.Matches(msg, keys.Back):
				t.screen = t.previousScreen
				return t, nil

			default:
				t.searchInput, cmd = t.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenSearchResults:
		t.searchList, cmd = t.searchList.Update(msg)
		cmds = append(cmds, cmd)

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Enter) || key.Matches(msg, keys.Right):
				if i, ok := t.searchList.SelectedItem().(Item); ok {
					if result, ok := i.Data.(models.SearchResult); ok {
						return t, func() tea.Msg {
							return searchResultSelected{
								tableName: result.TableName,
								entryID:   result.Entry.ID,
							}
						}
					}
				}

			case key.Matches(msg, keys.Back) || key.Matches(msg, keys.Left):
				return t, t.showSearchScreen()
			}
		}

	case ScreenConfirm:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, keys.Confirm):
				return t, func() tea.Msg {
					return confirmAction{confirmed: true}
				}

			case key.Matches(msg, keys.Cancel):
				return t, func() tea.Msg {
					return confirmAction{confirmed: false}
				}
			}
		}
	}

	return t, tea.Batch(cmds...)
}

func (t *TUI) View() string {
	if t.quitting {
		return "Thanks for using ThighPads!\n"
	}

	var s strings.Builder

	s.WriteString(t.theme.AppTitle.Render("ThighPads - Terminal Snippet Manager"))
	s.WriteString("\n\n")

	switch t.screen {
	case ScreenMain:
		s.WriteString(t.tableList.View())

	case ScreenTableDetails:
		if t.activeTable == nil {
			s.WriteString("No table selected")
		} else {
			s.WriteString(t.entryList.View())
		}

	case ScreenEntryDetails:
		if t.activeEntry == nil {
			s.WriteString("No entry selected")
		} else {
			var sb strings.Builder
			sb.WriteString(t.theme.Title.Render(t.activeEntry.Title))
			sb.WriteString("\n\n")

			if len(t.activeEntry.Tags) > 0 {
				for _, tag := range t.activeEntry.Tags {
					sb.WriteString(t.theme.Tag.Render(tag) + " ")
				}
				sb.WriteString("\n\n")
			}

			sb.WriteString(t.theme.Code.Render(t.activeEntry.Content))
			sb.WriteString("\n\n")

			sb.WriteString(t.theme.InfoText.Render(
				fmt.Sprintf("Created: %s • Updated: %s",
					formatTime(t.activeEntry.CreatedAt),
					formatTime(t.activeEntry.UpdatedAt)),
			))

			s.WriteString(t.theme.BoxStyle.Render(sb.String()))
		}

	case ScreenSearch:
		var sb strings.Builder

		sb.WriteString(t.theme.Title.Render("Search"))
		sb.WriteString("\n\n")

		sb.WriteString(t.theme.Label.Render("Enter search term:") + "\n")
		sb.WriteString(t.searchInput.View() + "\n\n")

		sb.WriteString(t.theme.InfoText.Render("Enter: Search • Esc: Cancel"))

		s.WriteString(t.theme.BoxStyle.Render(sb.String()))

	case ScreenSearchResults:
		s.WriteString(t.searchList.View())

	case ScreenCreateTable, ScreenImportTable, ScreenExportTable:
		s.WriteString(t.form.View())

	case ScreenCreateEntry, ScreenEditEntry:
		var sb strings.Builder

		for i, input := range t.form.Inputs {
			sb.WriteString(t.theme.Label.Render(t.form.Labels[i]) + "\n")
			sb.WriteString(input.View() + "\n\n")
		}

		sb.WriteString(t.theme.Label.Render("Content") + "\n")
		sb.WriteString(t.textarea.View() + "\n\n")

		sb.WriteString(t.theme.InfoText.Render(
			"Ctrl+T: Toggle between fields and content • Ctrl+Enter: Submit • Esc: Cancel",
		))

		s.WriteString(t.theme.BoxStyle.Render(
			t.theme.AppTitle.Render(t.form.Title) + "\n\n" + sb.String(),
		))

	case ScreenConfirm:
		s.WriteString(t.theme.BoxStyle.Render(
			t.confirmMsg + "\n\n" +
				t.theme.InfoText.Render("Press Y to confirm or N to cancel"),
		))
	}

	if t.status.Text != "" {
		style := t.theme.InfoText
		if t.status.IsError {
			style = t.theme.ErrorText
		}
		s.WriteString("\n" + style.Render(t.status.Text))
	}

	s.WriteString("\n\n" + t.renderContextualHelp())

	return s.String()
}

func (t *TUI) renderContextualHelp() string {
	var shortcuts []struct {
		key  string
		desc string
	}

	if t.screen != ScreenConfirm &&
		t.screen != ScreenCreateEntry &&
		t.screen != ScreenEditEntry &&
		t.screen != ScreenCreateTable &&
		t.screen != ScreenImportTable &&
		t.screen != ScreenExportTable &&
		t.screen != ScreenSearch {
		shortcuts = append(shortcuts, struct{ key, desc string }{"/", "search"})
	}

	if t.screen != ScreenConfirm {
		shortcuts = append(shortcuts,
			struct{ key, desc string }{"?", "help"},
			struct{ key, desc string }{"ctrl+c", "quit"},
		)
	}

	switch t.screen {
	case ScreenMain:
		shortcuts = append([]struct{ key, desc string }{
			{"↑/↓", "navigate"},
			{"enter/→", "open table"},
			{"n", "new table"},
			{"i", "import"},
		}, shortcuts...)

	case ScreenTableDetails:
		shortcuts = append([]struct{ key, desc string }{
			{"↑/↓", "navigate"},
			{"enter/→", "view entry"},
			{"a", "add entry"},
			{"d", "delete table"},
			{"x", "export"},
			{"esc/←", "back"},
		}, shortcuts...)

	case ScreenEntryDetails:
		shortcuts = append([]struct{ key, desc string }{
			{"e", "edit"},
			{"d", "delete"},
			{"esc/←", "back"},
		}, shortcuts...)

	case ScreenSearch:
		shortcuts = append([]struct{ key, desc string }{
			{"enter", "search"},
			{"esc", "cancel"},
		}, shortcuts...)

	case ScreenSearchResults:
		shortcuts = append([]struct{ key, desc string }{
			{"↑/↓", "navigate"},
			{"enter/→", "view entry"},
			{"esc/←", "back to search"},
		}, shortcuts...)

	case ScreenCreateTable, ScreenImportTable, ScreenExportTable:
		shortcuts = append([]struct{ key, desc string }{
			{"tab", "next field"},
			{"enter", "submit"},
			{"esc", "cancel"},
		}, shortcuts...)

	case ScreenCreateEntry, ScreenEditEntry:
		shortcuts = append([]struct{ key, desc string }{
			{"ctrl+t", "toggle fields/content"},
			{"tab", "next field"},
			{"ctrl+enter", "submit"},
			{"esc", "cancel"},
		}, shortcuts...)

	case ScreenConfirm:
		shortcuts = []struct{ key, desc string }{
			{"y", "confirm"},
			{"n", "cancel"},
		}
	}

	var b strings.Builder
	for i, sc := range shortcuts {
		b.WriteString(t.theme.Button.Render(" " + sc.key + " "))
		b.WriteString(" " + sc.desc)

		if i < len(shortcuts)-1 {
			b.WriteString("   ")
		}
	}

	return t.theme.Footer.Render(b.String())
}

func formatTime(timeStr string) string {
	if len(timeStr) > 10 {
		return timeStr[:10]
	}
	return timeStr
}

type searchResultSelected struct {
	tableName string
	entryID   string
}
