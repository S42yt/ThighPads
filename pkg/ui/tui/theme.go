package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/s42yt/thighpads/pkg/models"
)

var (
	accentColor     = lipgloss.Color("#7D56F4")
	secondaryColor  = lipgloss.Color("#AE88FF")
	textColor       = lipgloss.Color("#FFFFFF")
	subtleColor     = lipgloss.Color("#888888")
	errorColor      = lipgloss.Color("#FF5555")
	successColor    = lipgloss.Color("#55FF55")
	warningColor    = lipgloss.Color("#FFAA55")
	backgroundColor = lipgloss.Color("#222222")

	// Syntax highlighting token colors
	keywordColor  = lipgloss.Color("#569CD6")
	stringColor   = lipgloss.Color("#CE9178")
	numberColor   = lipgloss.Color("#B5CEA8")
	commentColor  = lipgloss.Color("#6A9955")
	functionColor = lipgloss.Color("#DCDCAA")
	typeColor     = lipgloss.Color("#4EC9B0")
	variableColor = lipgloss.Color("#9CDCFE")
	operatorColor = lipgloss.Color("#D4D4D4")

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

	// Syntax highlighting styles
	KeywordStyle  = lipgloss.NewStyle().Foreground(keywordColor)
	StringStyle   = lipgloss.NewStyle().Foreground(stringColor)
	NumberStyle   = lipgloss.NewStyle().Foreground(numberColor)
	CommentStyle  = lipgloss.NewStyle().Foreground(commentColor)
	FunctionStyle = lipgloss.NewStyle().Foreground(functionColor)
	TypeStyle     = lipgloss.NewStyle().Foreground(typeColor)
	VariableStyle = lipgloss.NewStyle().Foreground(variableColor)
	OperatorStyle = lipgloss.NewStyle().Foreground(operatorColor)
)

// Active syntax highlighters
var activeSyntaxHighlighters []*SyntaxHighlighter

// SyntaxHighlighter represents a syntax highlighter with its regex patterns
type SyntaxHighlighter struct {
	Name  string
	Rules []SyntaxRule
	Tags  []string
}

// SyntaxRule represents a syntax highlighting rule with a compiled regex pattern
type SyntaxRule struct {
	Pattern *regexp.Regexp
	Token   string
	Style   lipgloss.Style
}

func WithWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.Copy().Width(width)
}

func ApplyCustomTheme(theme *models.ThemeColors) {
	accentColor = lipgloss.Color(theme.Accent)
	secondaryColor = lipgloss.Color(theme.Secondary)
	textColor = lipgloss.Color(theme.Text)
	subtleColor = lipgloss.Color(theme.Subtle)
	errorColor = lipgloss.Color(theme.Error)
	successColor = lipgloss.Color(theme.Success)
	warningColor = lipgloss.Color(theme.Warning)
	backgroundColor = lipgloss.Color(theme.Background)

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
}

// LoadSyntaxHighlighter loads a syntax highlighter from a models.SyntaxHighlight
func LoadSyntaxHighlighter(syntax *models.SyntaxHighlight) *SyntaxHighlighter {
	highlighter := &SyntaxHighlighter{
		Name: syntax.Name,
		Tags: syntax.Tags,
	}

	for _, rule := range syntax.Rules {
		// Compile the regex pattern
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue // Skip invalid patterns
		}

		// Get the appropriate style for the token
		var style lipgloss.Style
		switch rule.Token {
		case "keyword":
			style = KeywordStyle
		case "string":
			style = StringStyle
		case "number":
			style = NumberStyle
		case "comment":
			style = CommentStyle
		case "function":
			style = FunctionStyle
		case "type":
			style = TypeStyle
		case "variable":
			style = VariableStyle
		case "operator":
			style = OperatorStyle
		default:
			style = Normal
		}

		// Apply custom color if provided
		if color, ok := syntax.TokenColors[rule.Token]; ok {
			style = style.Copy().Foreground(lipgloss.Color(color))
		}

		highlighter.Rules = append(highlighter.Rules, SyntaxRule{
			Pattern: pattern,
			Token:   rule.Token,
			Style:   style,
		})
	}

	return highlighter
}

// SetActiveSyntaxHighlighters sets the active syntax highlighters
func SetActiveSyntaxHighlighters(highlighters []*SyntaxHighlighter) {
	activeSyntaxHighlighters = highlighters
}

// AddSyntaxHighlighter adds a syntax highlighter to the active list
func AddSyntaxHighlighter(highlighter *SyntaxHighlighter) {
	activeSyntaxHighlighters = append(activeSyntaxHighlighters, highlighter)
}

// ClearSyntaxHighlighters removes all active syntax highlighters
func ClearSyntaxHighlighters() {
	activeSyntaxHighlighters = nil
}

// ApplyHighlighting applies syntax highlighting to text based on tags
func ApplyHighlighting(text string, tags []string) string {
	if len(activeSyntaxHighlighters) == 0 || len(tags) == 0 {
		return text
	}

	// Convert tags to lowercase set for easier matching
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[strings.ToLower(tag)] = true
	}

	// Find applicable highlighters
	var applicableHighlighters []*SyntaxHighlighter
	for _, highlighter := range activeSyntaxHighlighters {
		for _, tag := range highlighter.Tags {
			if tagSet[strings.ToLower(tag)] {
				applicableHighlighters = append(applicableHighlighters, highlighter)
				break
			}
		}
	}

	if len(applicableHighlighters) == 0 {
		return text
	}

	// Apply highlighting
	highlightedText := text

	// Create a slice to track which parts of the text have been highlighted
	type segment struct {
		start, end int
		style      lipgloss.Style
	}

	var segments []segment

	// Apply all rules from applicable highlighters
	for _, highlighter := range applicableHighlighters {
		for _, rule := range highlighter.Rules {
			matches := rule.Pattern.FindAllStringIndex(text, -1)
			for _, match := range matches {
				// Check if this segment overlaps with any existing segments
				overlap := false
				for _, seg := range segments {
					if match[0] < seg.end && match[1] > seg.start {
						overlap = true
						break
					}
				}

				if !overlap {
					segments = append(segments, segment{
						start: match[0],
						end:   match[1],
						style: rule.Style,
					})
				}
			}
		}
	}

	// Sort segments by start position in reverse order to apply styles without affecting positions
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[i].start < segments[j].start {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}

	// Apply styles from last to first to maintain correct positions
	for _, seg := range segments {
		part := highlightedText[seg.start:seg.end]
		styledPart := seg.style.Render(part)
		highlightedText = highlightedText[:seg.start] + styledPart + highlightedText[seg.end:]
	}

	return highlightedText
}
