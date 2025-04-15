package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/data"
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

func (a *App) updateSetupScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.usernameInput.Value() != "" {
				a.config.Username = a.usernameInput.Value()
				err := config.SaveConfig(a.config)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = HomeScreen
				a.loadTables()
				a.successMsg = "Setup complete! Welcome to ThighPads."
				return a, nil
			}
		case tea.KeyCtrlC:
			return a, tea.Quit
		}
	}

	a.usernameInput, cmd = a.usernameInput.Update(msg)
	return a, cmd
}

func (a *App) viewSetupScreen() string {
	title := Title.Render("Welcome to ThighPads")
	subtitle := BoxStyle.Render(Subtitle.Render("First-time Setup"))

	usernameInput := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Please enter your username:"),
			a.usernameInput.View(),
		),
	)

	help := HelpView(map[string]string{
		"Enter":  "Save username",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		usernameInput,
		help,
	)
}

func (a *App) updateHomeScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			a.screen = NewTableScreen
			a.tableNameInput = TextInputField("Enter table name")
			return a, nil
		case "i":
			a.screen = ImportScreen
			a.importPathInput = TextInputField("Enter path to .thighpad file")
			return a, nil
		case "q", "ctrl+c":
			return a, tea.Quit
		case "enter":
			if len(a.tables) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {
					for _, table := range a.tables {
						if table.ID == selected.ID {
							a.currentTable = table
							a.screen = TableScreen
							a.loadEntries()
							return a, nil
						}
					}
				}
			}
		}
	}

	if len(a.tables) > 0 {
		a.list, cmd = a.list.Update(msg)
	}

	return a, cmd
}

func (a *App) viewHomeScreen() string {
	title := Title.Render("ThighPads")
	subtitle := Subtitle.Render(fmt.Sprintf("Welcome, %s", a.config.Username))

	var content string
	if len(a.tables) == 0 {
		content = BoxStyle.Render(Normal.Render("You don't have any tables yet. Press 'n' to create your first one."))
	} else {
		content = BoxStyle.Render(a.list.View())
	}

	help := HelpView(map[string]string{
		"Enter": "Select table",
		"n":     "New table",
		"i":     "Import table",
		"q":     "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		content,
		help,
	)
}

func (a *App) updateTableScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			a.screen = NewEntryScreen
			a.entryTitleInput = TextInputField("Enter title")
			a.entryTagsInput = TextInputField("Enter tags (comma-separated)")
			a.entryContent = textarea.New()
			a.entryContent.Placeholder = "Enter your content here..."
			a.entryContent.SetWidth(a.width - 10)
			a.entryContent.SetHeight(a.height - 20)
			a.entryContent.Focus()
			return a, nil
		case "e":
			a.screen = ExportScreen
			a.exportName = TextInputField(a.currentTable.Name)
			return a, nil
		case "b":
			a.screen = HomeScreen
			a.loadTables()
			return a, nil
		case "d":
			if len(a.entries) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {
					err := database.DeleteEntry(selected.ID)
					if err != nil {
						a.errorMsg = err.Error()
					} else {
						a.successMsg = "Entry deleted successfully."
						a.loadEntries()
					}
					return a, nil
				}
			}
		case "q", "ctrl+c":
			return a, tea.Quit
		case "enter":
			if len(a.entries) > 0 {
				selected, ok := a.list.SelectedItem().(Selectable)
				if ok {
					for _, entry := range a.entries {
						if entry.ID == selected.ID {
							a.currentEntry = entry
							a.screen = ViewEntryScreen
							return a, nil
						}
					}
				}
			}
		}
	}

	if len(a.entries) > 0 {
		a.list, cmd = a.list.Update(msg)
	}

	return a, cmd
}

func (a *App) viewTableScreen() string {
	title := Title.Render(a.currentTable.Name)
	subtitle := Subtitle.Render(fmt.Sprintf("Created by %s on %s",
		a.currentTable.Author,
		a.currentTable.CreatedAt.Format("Jan 02, 2006")))

	var content string
	if len(a.entries) == 0 {
		content = BoxStyle.Render(Normal.Render("This table is empty. Press 'n' to create your first entry."))
	} else {
		content = BoxStyle.Render(a.list.View())
	}

	help := HelpView(map[string]string{
		"Enter": "View entry",
		"n":     "New entry",
		"d":     "Delete entry",
		"e":     "Export table",
		"b":     "Back to home",
		"q":     "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		content,
		help,
	)
}

func (a *App) updateNewTableScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.tableNameInput.Value() != "" {
				newTable := models.Table{
					Name:      a.tableNameInput.Value(),
					Author:    a.config.Username,
					CreatedAt: time.Now(),
				}

				err := database.CreateTable(&newTable)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = HomeScreen
				a.loadTables()
				a.successMsg = "Table created successfully."
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = HomeScreen
			return a, nil
		case tea.KeyCtrlC:
			return a, tea.Quit
		}
	}

	a.tableNameInput, cmd = a.tableNameInput.Update(msg)
	return a, cmd
}

func (a *App) viewNewTableScreen() string {
	title := Title.Render("Create New Table")

	nameInput := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Enter a name for your new table:"),
			a.tableNameInput.View(),
		),
	)

	help := HelpView(map[string]string{
		"Enter":  "Create table",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		nameInput,
		help,
	)
}

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
	title := Title.Render("New Entry")
	subtitle := Subtitle.Render(a.currentTable.Name)

	titleInput := Subtitle.Render("Title:") + "\n" + a.entryTitleInput.View()
	tagsInput := Subtitle.Render("Tags:") + "\n" + a.entryTagsInput.View()
	content := Subtitle.Render("Content:") + "\n" + a.entryContent.View()

	form := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s\n\n%s",
			titleInput,
			tagsInput,
			content,
		),
	)

	help := HelpView(map[string]string{
		"Tab":    "Next field",
		"Ctrl+S": "Save entry",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		form,
		help,
	)
}

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

func (a *App) updateEditEntryScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				updatedEntry := a.currentEntry
				updatedEntry.Title = a.entryTitleInput.Value()
				updatedEntry.Tags = a.entryTagsInput.Value()
				updatedEntry.Content = a.entryContent.Value()

				err := database.UpdateEntry(&updatedEntry)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.currentEntry = updatedEntry
				a.screen = TableScreen
				a.loadEntries()
				a.successMsg = "Entry updated successfully."
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = ViewEntryScreen
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

func (a *App) viewEditEntryScreen() string {
	title := Title.Render("Edit Entry")
	subtitle := Subtitle.Render(a.currentTable.Name)

	titleInput := Subtitle.Render("Title:") + "\n" + a.entryTitleInput.View()
	tagsInput := Subtitle.Render("Tags:") + "\n" + a.entryTagsInput.View()
	content := Subtitle.Render("Content:") + "\n" + a.entryContent.View()

	form := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s\n\n%s",
			titleInput,
			tagsInput,
			content,
		),
	)

	help := HelpView(map[string]string{
		"Tab":    "Next field",
		"Ctrl+S": "Save changes",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		form,
		help,
	)
}

func (a *App) updateImportScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.importPathInput.Value() != "" {
				path := a.importPathInput.Value()

				if _, err := os.Stat(path); os.IsNotExist(err) {
					a.errorMsg = "File does not exist"
					return a, nil
				}

				if !strings.HasSuffix(path, data.FileExtension) {
					a.errorMsg = "File must have " + data.FileExtension + " extension"
					return a, nil
				}

				err := data.ImportFile(path, a.config.Username)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = HomeScreen
				a.loadTables()
				a.successMsg = "Table imported successfully."
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = HomeScreen
			return a, nil
		case tea.KeyCtrlC:
			return a, tea.Quit
		}
	}

	a.importPathInput, cmd = a.importPathInput.Update(msg)
	return a, cmd
}

func (a *App) viewImportScreen() string {
	title := Title.Render("Import Table")

	importInput := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Enter path to .thighpad file:"),
			a.importPathInput.View(),
		),
	)

	help := HelpView(map[string]string{
		"Enter":  "Import table",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		importInput,
		help,
	)
}

func (a *App) updateExportScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.exportName.Value() != "" {
				filename, err := data.ExportTable(a.currentTable.ID, a.config.Username)
				if err != nil {
					a.errorMsg = err.Error()
					return a, nil
				}

				a.screen = TableScreen
				a.successMsg = "Table exported successfully to: " + filename
				return a, nil
			}
		case tea.KeyEsc:
			a.screen = TableScreen
			return a, nil
		case tea.KeyCtrlC:
			return a, tea.Quit
		}
	}

	a.exportName, cmd = a.exportName.Update(msg)
	return a, cmd
}

func (a *App) viewExportScreen() string {
	title := Title.Render("Export Table")
	subtitle := Subtitle.Render(a.currentTable.Name)

	exportInfo := BoxStyle.Render(
		Normal.Render(fmt.Sprintf("Exporting table with %d entries", len(a.entries))),
	)

	help := HelpView(map[string]string{
		"Enter":  "Export table",
		"Esc":    "Cancel",
		"Ctrl+C": "Quit",
	})

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		title,
		subtitle,
		exportInfo,
		help,
	)
}
