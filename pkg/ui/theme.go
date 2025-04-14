package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme and styles for the application
type Theme struct {
	// Colors
	PrimaryColor   lipgloss.Color
	SecondaryColor lipgloss.Color
	AccentColor    lipgloss.Color
	TextColor      lipgloss.Color
	DimTextColor   lipgloss.Color
	ErrorColor     lipgloss.Color
	SuccessColor   lipgloss.Color

	// Styles
	AppTitle     lipgloss.Style
	Title        lipgloss.Style
	Label        lipgloss.Style
	ListItem     lipgloss.Style
	SelectedItem lipgloss.Style
	InfoText     lipgloss.Style
	ErrorText    lipgloss.Style
	SuccessText  lipgloss.Style
	BoxStyle     lipgloss.Style
	StatusBar    lipgloss.Style
	HelpStyle    lipgloss.Style
}

// NewTheme creates a new theme with default colors
func NewTheme() *Theme {
	// Purple theme colors
	primaryColor := lipgloss.Color("#9370DB")   // Medium Purple
	secondaryColor := lipgloss.Color("#B19CD9") // Light Purple
	accentColor := lipgloss.Color("#7B68EE")    // Medium Slate Blue
	textColor := lipgloss.Color("#E6E6FA")      // Lavender
	dimTextColor := lipgloss.Color("#D8BFD8")   // Thistle
	errorColor := lipgloss.Color("#FF6347")     // Tomato
	successColor := lipgloss.Color("#98FB98")   // Pale Green

	theme := &Theme{
		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		AccentColor:    accentColor,
		TextColor:      textColor,
		DimTextColor:   dimTextColor,
		ErrorColor:     errorColor,
		SuccessColor:   successColor,
	}

	// Base text style
	baseStyle := lipgloss.NewStyle().Foreground(textColor)

	// Define styles
	theme.AppTitle = baseStyle.Copy().
		Foreground(accentColor).
		Bold(true).
		Underline(true).
		MarginBottom(1)

	theme.Title = baseStyle.Copy().
		Foreground(primaryColor).
		Bold(true).
		MarginBottom(1)

	theme.Label = baseStyle.Copy().
		Foreground(secondaryColor).
		Bold(true)

	theme.ListItem = baseStyle.Copy().
		Foreground(textColor)

	theme.SelectedItem = baseStyle.Copy().
		Foreground(primaryColor).
		Bold(true)

	theme.InfoText = baseStyle.Copy().
		Foreground(dimTextColor).
		Italic(true).
		MarginTop(1)

	theme.ErrorText = baseStyle.Copy().
		Foreground(errorColor).
		Bold(true).
		MarginTop(1)

	theme.SuccessText = baseStyle.Copy().
		Foreground(successColor).
		Bold(true).
		MarginTop(1)

	theme.BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	theme.StatusBar = lipgloss.NewStyle().
		Foreground(textColor).
		Background(primaryColor).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	theme.HelpStyle = lipgloss.NewStyle().
		Foreground(dimTextColor).
		MarginTop(1)

	return theme
}
