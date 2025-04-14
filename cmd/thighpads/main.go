package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/app"
	"github.com/s42yt/thighpads/pkg/ui"
)

func main() {
	
	thighpads, err := app.New()
	if err != nil {
		fmt.Printf("Error initializing ThighPads: %v\n", err)
		os.Exit(1)
	}

	
	tui := ui.NewTUI(thighpads)

	
	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ThighPads: %v\n", err)
		os.Exit(1)
	}
}
