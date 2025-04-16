package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s42yt/thighpads/pkg/config"
	"github.com/s42yt/thighpads/pkg/models"
)

func (a *App) updateSettingsScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1": 
			a.settingsSubScreen = "theme_selector"
			return a, nil
		case "2": 
			a.config.AutoCheckUpdate = !a.config.AutoCheckUpdate
			return a, nil
		case "3": 
			switch a.config.DefaultExport {
			case "config":
				a.config.DefaultExport = "desktop"
			case "desktop":
				a.config.DefaultExport = "both"
			default:
				a.config.DefaultExport = "config"
			}
			return a, nil
		case "4": 
			a.settingsSubScreen = "import_theme"
			a.filePathInput.SetValue("")
			a.filePathInput.Focus()
			return a, nil
		case "5": 
			a.settingsSubScreen = "import_syntax"
			a.filePathInput.SetValue("")
			a.filePathInput.Focus()
			return a, nil
		case "6": 
			a.config.SyntaxHighlighting = !a.config.SyntaxHighlighting
			return a, nil
		case "7": 
			a.settingsSubScreen = "syntax_manager"
			return a, nil
		case "s": 
			err := config.SaveConfig(a.config)
			if err != nil {
				a.errorMsg = err.Error()
				return a, nil
			}
			a.screen = HomeScreen
			a.successMsg = "Settings saved successfully."
			return a, nil
		case "esc":
			if a.settingsSubScreen != "" {
				a.settingsSubScreen = ""
				return a, nil
			}
			a.screen = HomeScreen
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	
	switch a.settingsSubScreen {
	case "theme_selector":
		return a.handleThemeSelector(msg)
	case "import_theme":
		return a.handleImportTheme(msg)
	case "import_syntax":
		return a.handleImportSyntax(msg)
	case "syntax_manager":
		return a.handleSyntaxManager(msg)
	}

	
	if a.filePathInput.Focused() {
		a.filePathInput, cmd = a.filePathInput.Update(msg)
		return a, cmd
	}

	return a, cmd
}

func (a *App) handleThemeSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			a.settingsSubScreen = ""
			return a, nil
		}


		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			numKey := msg.Runes[0] - '0'
			if numKey >= 1 && int(numKey) <= len(a.config.AvailableThemes) {
				a.config.Theme = a.config.AvailableThemes[numKey-1]

				
				theme, err := config.LoadTheme(a.config.Theme)
				if err == nil {
					ApplyCustomTheme(theme)
				}

				a.settingsSubScreen = ""
				a.successMsg = "Theme applied: " + a.config.Theme
				return a, nil
			}
		}
	}
	return a, nil
}

func (a *App) handleImportTheme(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.filePathInput.Value() != "" {
				err := config.ImportThemeFromFile(a.filePathInput.Value())
				if err != nil {
					a.errorMsg = "Failed to import theme: " + err.Error()
				} else {
					
					themes, _ := config.DiscoverThemes()
					a.config.AvailableThemes = themes
					a.settingsSubScreen = ""
					a.successMsg = "Theme imported successfully."
				}
				return a, nil
			}
		case tea.KeyEsc:
			a.settingsSubScreen = ""
			return a, nil
		}
	}

	a.filePathInput, cmd = a.filePathInput.Update(msg)
	return a, cmd
}

func (a *App) handleImportSyntax(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if a.filePathInput.Value() != "" {
				err := config.ImportSyntaxFromFile(a.filePathInput.Value())
				if err != nil {
					a.errorMsg = "Failed to import syntax: " + err.Error()
				} else {
					
					syntaxes, _ := config.DiscoverSyntaxThemes()
					a.config.AvailableSyntaxes = syntaxes
					a.settingsSubScreen = ""
					a.successMsg = "Syntax highlighting imported successfully."
				}
				return a, nil
			}
		case tea.KeyEsc:
			a.settingsSubScreen = ""
			return a, nil
		}
	}

	a.filePathInput, cmd = a.filePathInput.Update(msg)
	return a, cmd
}

func (a *App) handleSyntaxManager(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			a.settingsSubScreen = ""
			return a, nil
		}

		
		if msg.Type == tea.KeyRunes {
			numKey := msg.Runes[0] - '0' 
			if numKey >= 1 && int(numKey) <= len(a.config.AvailableSyntaxes) {
				syntaxName := a.config.AvailableSyntaxes[numKey-1]

				
				found := false
				for i, name := range a.config.EnabledSyntaxThemes {
					if name == syntaxName {
						
						a.config.EnabledSyntaxThemes = append(
							a.config.EnabledSyntaxThemes[:i],
							a.config.EnabledSyntaxThemes[i+1:]...,
						)
						found = true
						break
					}
				}

				if !found {
					
					a.config.EnabledSyntaxThemes = append(a.config.EnabledSyntaxThemes, syntaxName)

					
					syntaxPath, _ := config.GetSyntaxPath()
					syntax, err := models.LoadSyntaxFromFile(filepath.Join(syntaxPath, syntaxName+".json"))
					if err == nil {
						
						for _, tag := range syntax.Tags {
							a.config.TagSyntaxMap[strings.ToLower(tag)] = syntaxName
						}
					}
				}

				return a, nil
			}
		}
	}
	return a, nil
}

func (a *App) viewSettingsScreen() string {
	if a.settingsSubScreen != "" {
		switch a.settingsSubScreen {
		case "theme_selector":
			return a.viewThemeSelector()
		case "import_theme":
			return a.viewImportTheme()
		case "import_syntax":
			return a.viewImportSyntax()
		case "syntax_manager":
			return a.viewSyntaxManager()
		}
	}

	title := Title.Render("Settings")

	themeStatus := "Unknown"
	if theme, ok := find(a.config.AvailableThemes, a.config.Theme); ok {
		themeStatus = theme
	}

	autoUpdateStatus := "On"
	if !a.config.AutoCheckUpdate {
		autoUpdateStatus = "Off"
	}

	exportStatus := ""
	switch a.config.DefaultExport {
	case "config":
		exportStatus = "Config folder"
	case "desktop":
		exportStatus = "Desktop"
	case "both":
		exportStatus = "Both"
	}

	syntaxHighlightingStatus := "On"
	if !a.config.SyntaxHighlighting {
		syntaxHighlightingStatus = "Off"
	}

	settings := BoxStyle.Render(
		fmt.Sprintf("%s: %s\n\n%s: %s\n\n%s: %s\n\n%s\n\n%s\n\n%s: %s\n\n%s",
			Subtitle.Render("1. Theme"),
			Normal.Render(themeStatus),
			Subtitle.Render("2. Auto-check updates"),
			Normal.Render(autoUpdateStatus),
			Subtitle.Render("3. Default export location"),
			Normal.Render(exportStatus),
			Subtitle.Render("4. Import theme file"),
			Subtitle.Render("5. Import syntax highlighting"),
			Subtitle.Render("6. Syntax highlighting"),
			Normal.Render(syntaxHighlightingStatus),
			Subtitle.Render("7. Manage syntax highlighting"),
		),
	)

	footer := Normal.Render("\nPress 's' to save, 'esc' to return")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		settings,
		footer,
	)
}

func (a *App) viewThemeSelector() string {
	title := Title.Render("Select Theme")

	var content string
	for i, theme := range a.config.AvailableThemes {
		selected := ""
		if theme == a.config.Theme {
			selected = " " + Selected.Render("ACTIVE")
		}
		content += fmt.Sprintf("%d. %s%s\n", i+1, theme, selected)
	}

	themeList := BoxStyle.Render(content)

	footer := Normal.Render("\nPress a number to select, 'esc' to cancel")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		themeList,
		footer,
	)
}

func (a *App) viewImportTheme() string {
	title := Title.Render("Import Theme")

	content := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Enter the path to your theme JSON file:"),
			a.filePathInput.View(),
		),
	)

	footer := Normal.Render("\nPress Enter to import, 'esc' to cancel")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		content,
		footer,
	)
}

func (a *App) viewImportSyntax() string {
	title := Title.Render("Import Syntax Highlighting")

	content := BoxStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			Normal.Render("Enter the path to your syntax highlighting JSON file:"),
			a.filePathInput.View(),
		),
	)

	footer := Normal.Render("\nPress Enter to import, 'esc' to cancel")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		content,
		footer,
	)
}

func (a *App) viewSyntaxManager() string {
	title := Title.Render("Manage Syntax Highlighting")

	var content string
	for i, syntax := range a.config.AvailableSyntaxes {
		selected := ""
		if contains(a.config.EnabledSyntaxThemes, syntax) {
			selected = " " + Selected.Render("ENABLED")
		}
		content += fmt.Sprintf("%d. %s%s\n", i+1, syntax, selected)
	}

	if len(a.config.AvailableSyntaxes) == 0 {
		content = "No syntax highlighting themes available.\nImport some using option 5 from the main settings."
	}

	syntaxList := BoxStyle.Render(content)

	footer := Normal.Render("\nPress a number to toggle, 'esc' to return")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		title,
		syntaxList,
		footer,
	)
}


func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func find(slice []string, item string) (string, bool) {
	for _, s := range slice {
		if s == item {
			return s, true
		}
	}
	return "", false
}
