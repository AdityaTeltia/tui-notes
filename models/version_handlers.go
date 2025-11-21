package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *MainModel) handleVersionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = "main"
		m.showVersions = false
		return m, nil
	case "up", "k":
		if m.selectedVersion > 0 {
			m.selectedVersion--
		}
		return m, nil
	case "down", "j":
		if m.selectedVersion < len(m.versions)-1 {
			m.selectedVersion++
		}
		return m, nil
	case "enter", "r":
		// Restore selected version
		if m.selectedVersion < len(m.versions) && m.currentNote != nil {
			version := m.versions[m.selectedVersion]
			m.RestoreVersion(m.currentNote, version.ID)
			m.currentView = "main"
			m.loadNotes()
		}
		return m, nil
	}
	return m, nil
}

func (m *MainModel) renderVersions() string {
	var s strings.Builder
	
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Version History")
	
	s.WriteString(title + "\n\n")
	
	if len(m.versions) == 0 {
		s.WriteString("No versions available.\n")
	} else {
		for i, version := range m.versions {
			cursor := " "
			if i == m.selectedVersion {
				cursor = ">"
			}
			
			style := lipgloss.NewStyle().PaddingLeft(2)
			if i == m.selectedVersion {
				style = style.Foreground(lipgloss.Color("205")).Bold(true)
			}
			
			timeStr := version.CreatedAt.Format("2006-01-02 15:04:05")
			preview := version.Content
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			
			line := fmt.Sprintf("%s %s - %s", cursor, timeStr, preview)
			s.WriteString(style.Render(line) + "\n")
		}
	}
	
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(2).
		Render("↑/↓: Navigate | Enter: Restore | Esc: Back")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

