package tui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
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
)

const (
    DefaultExport = iota
    DesktopExport
    BothExport
)

type App struct {
    screen          Screen
    width           int
    height          int
    config          *models.Config
    tables          []models.Table
    currentTable    models.Table
    entries         []models.Entry
    currentEntry    models.Entry
    list            list.Model
    usernameInput   textinput.Model
    tableNameInput  textinput.Model
    entryTitleInput textinput.Model
    entryTagsInput  textinput.Model
    entryContent    textarea.Model
    importPathInput textinput.Model
    exportName      textinput.Model
    errorMsg        string
    successMsg      string
    exportLocation  int
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
        screen: initialScreen,
        config: cfg,
    }

    if app.screen == SetupScreen {
        app.usernameInput = TextInputField("Enter your username")
    } else {
        app.loadTables()
    }

    p := tea.NewProgram(app, tea.WithAltScreen())
    return p, nil
}

func (a *App) Init() tea.Cmd {
    return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    if msg, ok := msg.(tea.WindowSizeMsg); ok {
        a.width = msg.Width
        a.height = msg.Height

        if a.list.Items() != nil {
            a.list.SetWidth(msg.Width - 10)
            a.list.SetHeight(msg.Height - 10)
        }

        if a.screen == NewEntryScreen || a.screen == EditEntryScreen {
            a.entryContent.SetWidth(msg.Width - 10)
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
    }

    return a, cmd
}

func (a *App) View() string {
    var view string

    switch a.screen {
    case SetupScreen:
        view = a.viewSetupScreen()
    case HomeScreen:
        view = a.viewHomeScreen()
    case TableScreen:
        view = a.viewTableScreen()
    case NewTableScreen:
        view = a.viewNewTableScreen()
    case NewEntryScreen:
        view = a.viewNewEntryScreen()
    case ViewEntryScreen:
        view = a.viewViewEntryScreen()
    case EditEntryScreen:
        view = a.viewEditEntryScreen()
    case ImportScreen:
        view = a.viewImportScreen()
    case ExportScreen:
        view = a.viewExportScreen()
    }

    if a.errorMsg != "" {
        view += "\n" + ErrorView(a.errorMsg)
    }

    if a.successMsg != "" {
        view += "\n" + SuccessView(a.successMsg)
    }

    return AppStyle.Render(view)
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

        a.list = SelectableList("Your Tables", items, a.width-10, a.height-10)
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

        a.list = SelectableList(a.currentTable.Name, items, a.width-10, a.height-10)
    }
}