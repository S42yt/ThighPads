package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/models"
)

type Screen int

const (
	SetupScreen Screen = iota
	HomeScreen
	TableScreen
	NewTableScreen
	NewEntryScreen
	ViewEntryScreen
	EditEntryScreen
	ImportScreen
	ExportScreen
	SettingsScreen
)

const (
	DefaultExport = iota
	DesktopExport
	BothExport
)

type HelpMsg bool

func ToggleHelp() tea.Cmd {
	return func() tea.Msg {
		return HelpMsg(true)
	}
}

type App struct {
	screen             Screen
	width              int
	height             int
	config             *models.Config
	tables             []models.Table
	currentTable       models.Table
	entries            []models.Entry
	currentEntry       models.Entry
	list               list.Model
	usernameInput      textinput.Model
	tableNameInput     textinput.Model
	entryTitleInput    textinput.Model
	entryTagsInput     textinput.Model
	entryContent       textarea.Model
	entryViewport      viewport.Model
	importPathInput    textinput.Model
	exportName         textinput.Model
	errorMsg           string
	successMsg         string
	exportLocation     int
	bottomGap          int
	showFullHelp       bool
	filePathInput      textinput.Model
	settingsSubScreen  string
	syntaxHighlighters map[string]*SyntaxHighlighter
}

func Initialize() (*tea.Program, error) {
	isFirstRun, err := config.IsFirstRun()
	if err != nil {
		return nil, err
	}

	_, err = config.EnsureConfigFolderExists()
	if err != nil {
		return nil, err
	}

	err = database.Initialize()
	if err != nil {
		return nil, err
	}

	var cfg *models.Config
	var initialScreen Screen

	if isFirstRun {
		initialScreen = SetupScreen
		cfg = &models.Config{Username: ""}
	} else {
		initialScreen = HomeScreen
		cfg, err = config.LoadConfig()
		if err != nil {
			return nil, err
		}
	}

	app := &App{
		screen:             initialScreen,
		config:             cfg,
		bottomGap:          4,
		showFullHelp:       false,
		syntaxHighlighters: make(map[string]*SyntaxHighlighter),
	}

	if app.screen == SetupScreen {
		app.usernameInput = TextInputField("Enter your username")
	} else {
		app.loadTables()
		
		if app.config.Theme != "" {
			theme, err := config.LoadTheme(app.config.Theme)
			if err == nil {
				ApplyCustomTheme(theme)
			}
		}
		
		app.loadSyntaxHighlighters()
	}

	
	app.filePathInput = TextInputField("Enter file path")

	p := tea.NewProgram(app, tea.WithAltScreen())
	return p, nil
}


func (a *App) loadSyntaxHighlighters() {
	if !a.config.SyntaxHighlighting {
		return
	}

	ClearSyntaxHighlighters()

	
	for _, name := range a.config.EnabledSyntaxThemes {
		syntax, err := config.LoadSyntaxHighlighting(name)
		if err != nil {
			continue
		}

		highlighter := LoadSyntaxHighlighter(syntax)
		a.syntaxHighlighters[name] = highlighter
		AddSyntaxHighlighter(highlighter)
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if _, ok := msg.(HelpMsg); ok {
		a.showFullHelp = !a.showFullHelp
		return a, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok {

		inTextInput := (a.screen == NewEntryScreen || a.screen == EditEntryScreen ||
			a.screen == ImportScreen || a.screen == SetupScreen || a.screen == NewTableScreen)

		if (inTextInput && k.String() == "ctrl+h") || (!inTextInput && k.String() == "h") {
			a.showFullHelp = !a.showFullHelp
			return a, nil
		}

		a.errorMsg = ""
		a.successMsg = ""
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		a.width = msg.Width
		a.height = msg.Height

		if a.list.Items() != nil {
			a.list.SetWidth(msg.Width - 4)
			a.list.SetHeight(msg.Height - 12)
		}

		if a.screen == ViewEntryScreen {
			a.entryViewport.Width = msg.Width - 6
			a.entryViewport.Height = msg.Height - 16
		}

		if a.screen == NewEntryScreen || a.screen == EditEntryScreen {
			a.entryContent.SetWidth(msg.Width - 6)
			a.entryContent.SetHeight(msg.Height - 20)
		}

		return a, nil
	}

	switch a.screen {
	case SetupScreen:
		return a.updateSetupScreen(msg)
	case HomeScreen:
		return a.updateHomeScreen(msg)
	case TableScreen:
		return a.updateTableScreen(msg)
	case NewTableScreen:
		return a.updateNewTableScreen(msg)
	case NewEntryScreen:
		return a.updateNewEntryScreen(msg)
	case ViewEntryScreen:
		return a.updateViewEntryScreen(msg)
	case EditEntryScreen:
		return a.updateEditEntryScreen(msg)
	case ImportScreen:
		return a.updateImportScreen(msg)
	case ExportScreen:
		return a.updateExportScreen(msg)
	case SettingsScreen:
		return a.updateSettingsScreen(msg)
	}

	return a, cmd
}

func (a *App) View() string {
	var view string
	var toolbarKeys map[string]string

	inTextInput := (a.screen == NewEntryScreen || a.screen == EditEntryScreen ||
		a.screen == ImportScreen || a.screen == SetupScreen || a.screen == NewTableScreen)

	helpKey := "h"
	helpDesc := "Show help"
	if inTextInput {
		helpKey = "Ctrl+h"
	}

	if a.showFullHelp {
		helpDesc = "Hide help"
	}

	switch a.screen {
	case SetupScreen:
		view = a.viewSetupScreen()
		toolbarKeys = map[string]string{
			"Enter":  "Save username",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case HomeScreen:
		view = a.viewHomeScreen()
		toolbarKeys = map[string]string{
			"↑/↓":   "Navigate",
			"Enter": "Select table",
			"n":     "New table",
			"s":     "Settings",
			"i":     "Import table",
			helpKey: helpDesc,
			"q":     "Quit",
		}
	case TableScreen:
		view = a.viewTableScreen()
		toolbarKeys = map[string]string{
			"↑/↓":   "Navigate",
			"Enter": "View entry",
			"n":     "New entry",
			"d":     "Delete entry",
			"e":     "Export table",
			"b":     "Back to home",
			helpKey: helpDesc,
			"q":     "Quit",
		}
	case NewTableScreen:
		view = a.viewNewTableScreen()
		toolbarKeys = map[string]string{
			"Enter":  "Create table",
			"Esc":    "Cancel",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case NewEntryScreen:
		view = a.viewNewEntryScreen()
		toolbarKeys = map[string]string{
			"Tab":    "Next field",
			"Ctrl+S": "Save entry",
			"Esc":    "Cancel",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case ViewEntryScreen:
		view = a.viewViewEntryScreen()
		toolbarKeys = map[string]string{
			"↑/↓":    "Scroll",
			"Ctrl+e": "Edit",
			"c":      "Copy to clipboard",
			"Ctrl+b": "Back",
			helpKey:  helpDesc,
			"Ctrl+q": "Quit",
		}
	case EditEntryScreen:
		view = a.viewEditEntryScreen()
		toolbarKeys = map[string]string{
			"Tab":    "Next field",
			"Ctrl+S": "Save changes",
			"Esc":    "Cancel",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case ImportScreen:
		view = a.viewImportScreen()
		toolbarKeys = map[string]string{
			"Enter":  "Import table",
			"Esc":    "Cancel",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case ExportScreen:
		view = a.viewExportScreen()
		toolbarKeys = map[string]string{
			"1-3":    "Select location",
			"Enter":  "Export table",
			"Esc":    "Cancel",
			helpKey:  helpDesc,
			"Ctrl+C": "Quit",
		}
	case SettingsScreen:
		view = a.viewSettingsScreen()
		toolbarKeys = map[string]string{
			"1-4":   "Change setting",
			"s":     "Save settings",
			helpKey: helpDesc,
			"Esc":   "Cancel",
		}
	}

	statusView := ""
	if a.errorMsg != "" {
		statusView = "\n" + ErrorView(a.errorMsg)
	}
	if a.successMsg != "" {
		statusView = "\n" + SuccessView(a.successMsg)
	}

	contentView := view

	if a.showFullHelp {

		helpHeight := a.height / 2
		if helpHeight > 20 {
			helpHeight = 20
		}

		inTextInput := (a.screen == NewEntryScreen || a.screen == EditEntryScreen ||
			a.screen == ImportScreen || a.screen == SetupScreen || a.screen == NewTableScreen)

		helpView := FullHelpView(toolbarKeys, a.width, helpHeight, inTextInput)
		contentView = contentView + "\n\n" + helpView
	}

	if !a.showFullHelp {

	} else {

	}

	help := FixedToolbarView(toolbarKeys, a.width)

	return AppStyle.Render(contentView + statusView + "\n" + help)
}

func (a *App) loadTables() {
	tables, err := database.GetTables()
	if err == nil {
		a.tables = tables

		items := make([]list.Item, len(tables))
		for i, table := range tables {
			desc := fmt.Sprintf("Created by %s on %s",
				table.Author,
				table.CreatedAt.Format("Jan 02, 2006"))

			items[i] = Selectable{
				Title:       table.Name,
				Description: desc,
				ID:          table.ID,
			}
		}

		a.list = SelectableList("Your Tables", items, a.width-4, a.height-12)
	}
}

func (a *App) loadEntries() {
	entries, err := database.GetEntries(a.currentTable.ID)
	if err == nil {
		a.entries = entries

		items := make([]list.Item, len(entries))
		for i, entry := range entries {
			desc := fmt.Sprintf("Tags: %s", entry.Tags)

			items[i] = Selectable{
				Title:       entry.Title,
				Description: desc,
				ID:          entry.ID,
			}
		}

		a.list = SelectableList(a.currentTable.Name, items, a.width-4, a.height-12)
	}
}
