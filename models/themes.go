package models

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name        string
	Background  string
	Foreground  string
	Primary     string
	Secondary   string
	Accent      string
	Border      string
	Selected    string
	Title       string
	StatusBar   string
}

var Themes = map[string]Theme{
	"default": {
		Name:       "Default",
		Background: "236",
		Foreground: "255",
		Primary:    "62",
		Secondary:  "240",
		Accent:     "205",
		Border:     "62",
		Selected:   "205",
		Title:      "62",
		StatusBar:  "236",
	},
	"dark": {
		Name:       "Dark",
		Background: "0",
		Foreground: "255",
		Primary:    "33",
		Secondary:  "238",
		Accent:     "51",
		Border:     "33",
		Selected:   "51",
		Title:      "33",
		StatusBar:  "0",
	},
	"light": {
		Name:       "Light",
		Background: "255",
		Foreground: "0",
		Primary:    "62",
		Secondary:  "240",
		Accent:     "205",
		Border:     "62",
		Selected:   "205",
		Title:      "62",
		StatusBar:  "252",
	},
	"monokai": {
		Name:       "Monokai",
		Background: "235",
		Foreground: "252",
		Primary:    "141",
		Secondary:  "59",
		Accent:     "197",
		Border:     "141",
		Selected:   "197",
		Title:      "141",
		StatusBar:  "235",
	},
	"nord": {
		Name:       "Nord",
		Background: "236",
		Foreground: "255",
		Primary:    "109",
		Secondary:  "240",
		Accent:     "103",
		Border:     "109",
		Selected:   "103",
		Title:      "109",
		StatusBar:  "236",
	},
}

func (m *MainModel) ApplyTheme(themeName string) {
	_, exists := Themes[themeName]
	if !exists {
		themeName = "default"
	}
	
	// Update theme
	m.currentTheme = themeName
}

func GetThemeStyles(themeName string) map[string]lipgloss.Style {
	theme, exists := Themes[themeName]
	if !exists {
		theme = Themes["default"]
	}
	
	return map[string]lipgloss.Style{
		"sidebar": lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(0, 1),
		"main": lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(0, 1),
		"title": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Title)).
			Padding(0, 1),
		"selected": lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Selected)).
			Bold(true),
		"normal": lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Secondary)),
		"status": lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Secondary)).
			Background(lipgloss.Color(theme.StatusBar)),
	}
}

// Get styles based on current theme (method for MainModel)
func (m *MainModel) getStyles() map[string]lipgloss.Style {
	return GetThemeStyles(m.currentTheme)
}

