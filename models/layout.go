package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout styles for two-pane design
var (
	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	mainStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("236"))
)

// RenderTwoPane renders the two-pane layout
func (m *MainModel) RenderTwoPane() string {
	// Ensure minimum dimensions
	if m.width < 40 {
		m.width = 40
	}
	if m.height < 10 {
		m.height = 10
	}
	
	sidebarWidth := m.width / 3
	mainWidth := m.width - sidebarWidth - 4 // Account for borders and padding
	
	// Ensure minimum widths
	if sidebarWidth < 15 {
		sidebarWidth = 15
		mainWidth = m.width - 19
	}
	if mainWidth < 20 {
		mainWidth = 20
		sidebarWidth = m.width - 24
	}
	
	// Final safety check
	if sidebarWidth < 10 {
		sidebarWidth = 10
	}
	if mainWidth < 10 {
		mainWidth = 10
	}

	// Get theme styles
	styles := m.getStyles()
	
	// Render sidebar
	sidebar := m.renderSidebar(sidebarWidth, m.height-3)
	
	// Render main content
	mainContent := m.renderMainContent(mainWidth, m.height-3)
	
	// Combine side by side with theme styles
	combined := lipgloss.JoinHorizontal(lipgloss.Left,
		styles["sidebar"].Width(sidebarWidth).Height(m.height-3).Render(sidebar),
		styles["main"].Width(mainWidth).Height(m.height-3).Render(mainContent),
	)
	
	// Add status bar at bottom
	statusBar := m.renderStatusBar()
	
	return lipgloss.JoinVertical(lipgloss.Left, combined, statusBar)
}

func (m *MainModel) renderSidebar(width, height int) string {
	var s strings.Builder
	styles := m.getStyles()
	
	// Ensure minimum width
	if width < 10 {
		width = 10
	}
	
	// Title
	title := styles["title"].Render("Notes")
	s.WriteString(title + "\n")
	separatorLen := width - 2
	if separatorLen > 0 {
		s.WriteString(strings.Repeat("─", separatorLen) + "\n")
	} else {
		s.WriteString("─\n")
	}
	
	// Notes list
	items := m.getSidebarItems()
	
	// Calculate how many items fit
	maxItems := height - 3
	startIdx := 0
	if m.sidebarCursor >= maxItems {
		startIdx = m.sidebarCursor - maxItems + 1
	}
	
	for i := startIdx; i < len(items) && i < startIdx+maxItems; i++ {
		item := items[i]
		line := ""
		
		if i == m.sidebarCursor {
			line = styles["selected"].Render("▶ " + item)
		} else {
			line = styles["normal"].Render("  " + item)
		}
		
		// Truncate if too long
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		
		s.WriteString(line + "\n")
	}
	
	return s.String()
}

func (m *MainModel) getSidebarItems() []string {
	items := []string{}
	
	// Use filtered notes if filtering is active
	notesToShow := m.notes
	if m.showFiltered && len(m.filteredNotes) > 0 {
		notesToShow = m.filteredNotes
	}
	
	// Add folders and notes
	for _, note := range notesToShow {
		if note.isFolder {
			expanded := m.sidebarExpanded[note.title]
			if expanded {
				items = append(items, "▼ "+note.title)
			} else {
				items = append(items, "▶ "+note.title)
			}
		} else {
			items = append(items, "  "+note.title)
		}
	}
	
	return items
}

func (m *MainModel) renderMainContent(width, height int) string {
	var s strings.Builder
	
	if m.currentNote == nil {
		// Show welcome or empty state
		s.WriteString(titleStyle.Render("Welcome to Terminal Notes") + "\n\n")
		s.WriteString("Select a note from the sidebar or create a new one.\n")
		s.WriteString("\nPress 'n' to create a new note")
		return s.String()
	}
	
	// Note title
	title := titleStyle.Render(m.currentNote.Title)
	s.WriteString(title + "\n")
	separatorLen := width - 2
	if separatorLen > 0 {
		s.WriteString(strings.Repeat("─", separatorLen) + "\n\n")
	} else {
		s.WriteString("─\n\n")
	}
	
	// Note content
	content := m.currentNote.Content
	if content == "" {
		content = "No content yet. Start typing..."
	}
	
	// Wrap content to fit width
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line) > width-4 {
			// Word wrap
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word)+1 > width-4 {
					if currentLine != "" {
						s.WriteString(currentLine + "\n")
					}
					currentLine = word
				} else {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += word
				}
			}
			if currentLine != "" {
				s.WriteString(currentLine + "\n")
			}
		} else {
			s.WriteString(line + "\n")
		}
	}
	
	// Show metadata
	separatorLen = width - 2
	if separatorLen > 0 {
		s.WriteString("\n" + strings.Repeat("─", separatorLen) + "\n")
	} else {
		s.WriteString("\n─\n")
	}
	if !m.currentNote.CreatedAt.IsZero() {
		s.WriteString(fmt.Sprintf("Created: %s\n", m.currentNote.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	if len(m.currentNote.Tags) > 0 {
		s.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(m.currentNote.Tags, ", ")))
	}
	
	return s.String()
}

func (m *MainModel) renderStatusBar() string {
	var left, right strings.Builder
	
	// Left side: current note info
	if m.currentNote != nil {
		left.WriteString(fmt.Sprintf("%s • %d notes", m.currentNote.Title, len(m.notes)))
	} else {
		left.WriteString(fmt.Sprintf("%d notes", len(m.notes)))
	}
	
	// Right side: keyboard shortcuts (shortened for smaller terminals)
	shortcuts := []string{
		"↑/k,↓/j",
		"enter: edit",
		"n: new",
		"s: sort",
		"f: filter",
		"q: quit",
	}
	right.WriteString(strings.Join(shortcuts, " • "))
	
	// Calculate spacing safely
	leftStr := left.String()
	rightStr := right.String()
	spacing := m.width - len(leftStr) - len(rightStr)
	
	// Ensure spacing is non-negative
	if spacing < 0 {
		// If too narrow, just show left side
		if m.width < len(leftStr) {
			leftStr = leftStr[:m.width]
		}
		spacing = 0
	}
	
	// Get theme styles
	styles := m.getStyles()
	
	// Combine with proper spacing
	statusBar := styles["status"].Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			leftStr,
			strings.Repeat(" ", spacing),
			rightStr,
		),
	)
	
	return statusBar
}

