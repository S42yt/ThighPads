package tui

import "github.com/charmbracelet/lipgloss"

var (
	accentColor     = lipgloss.Color("#7D56F4")
	secondaryColor  = lipgloss.Color("#AE88FF")
	textColor       = lipgloss.Color("#FFFFFF")
	subtleColor     = lipgloss.Color("#888888")
	errorColor      = lipgloss.Color("#FF5555")
	successColor    = lipgloss.Color("#55FF55")
	warningColor    = lipgloss.Color("#FFAA55")
	backgroundColor = lipgloss.Color("#222222")

	Title = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(0, 2)

	Subtitle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(textColor)

	Subtle = lipgloss.NewStyle().
		Foreground(subtleColor)

	Success = lipgloss.NewStyle().
		Foreground(successColor)

	Error = lipgloss.NewStyle().
		Foreground(errorColor)

	Warning = lipgloss.NewStyle().
		Foreground(warningColor)

	Selected = lipgloss.NewStyle().
			Foreground(textColor).
			Background(accentColor).
			Bold(true).
			Padding(0, 1)

	Unselected = lipgloss.NewStyle().
			Foreground(subtleColor).
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2)

	AppStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Padding(1, 2)
)
