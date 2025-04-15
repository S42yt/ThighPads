package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
)

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
