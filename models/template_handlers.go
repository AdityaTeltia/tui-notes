package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *MainModel) handleTemplateSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = "main"
		return m, nil
	case "up", "k":
		if m.selectedTemplate > 0 {
			m.selectedTemplate--
		}
		return m, nil
	case "down", "j":
		if m.selectedTemplate < len(m.templates)-1 {
			m.selectedTemplate++
		}
		return m, nil
	case "enter":
		if m.selectedTemplate < len(m.templates) {
			m.CreateNoteFromTemplate(m.templates[m.selectedTemplate])
		}
		return m, nil
	}
	return m, nil
}

func (m *MainModel) renderTemplateSelect() string {
	var s strings.Builder
	
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Select Template")
	
	s.WriteString(title + "\n\n")
	
	for i, tmpl := range m.templates {
		cursor := " "
		if i == m.selectedTemplate {
			cursor = ">"
		}
		
		style := lipgloss.NewStyle().PaddingLeft(2)
		if i == m.selectedTemplate {
			style = style.Foreground(lipgloss.Color("205")).Bold(true)
		}
		
		line := fmt.Sprintf("%s %s - %s", cursor, tmpl.Name, tmpl.Description)
		s.WriteString(style.Render(line) + "\n")
	}
	
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(2).
		Render("↑/↓: Navigate | Enter: Select | Esc: Cancel")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

