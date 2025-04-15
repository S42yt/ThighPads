package tui

import (
    "fmt"
    "os"
    "strings"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/s42yt/thighpads/pkg/data"
)

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