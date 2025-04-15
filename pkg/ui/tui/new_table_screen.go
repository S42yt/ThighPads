package tui

import (
    "fmt"
    "time"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/s42yt/thighpads/pkg/database"
    "github.com/s42yt/thighpads/pkg/models"
)

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