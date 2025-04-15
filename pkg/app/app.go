package app

import (
	"fmt"

	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/database"
	"github.com/s42yt/thighpads/pkg/ui/tui"
)

func Run() error {

	_, err := config.EnsureConfigFolderExists()
	if err != nil {
		return fmt.Errorf("failed to create config folder: %w", err)
	}

	err = database.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	program, err := tui.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}

	_, err = program.Run()
	return err
}
