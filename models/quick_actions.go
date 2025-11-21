package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/ssh-notes/terminal-notes/utils"
)

// QuickActions handles quick action commands
type QuickActions struct {
	commandInput textinput.Model
	showCommand  bool
	command      string
}

func (m *MainModel) handleQuickAction(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "g": // Quick jump/go to note
		return m.quickJump()
	case "r": // Recent notes
		return m.showRecentNotes()
	case "ctrl+n": // Quick new note
		m.createNewNote()
		return m, nil
	case "ctrl+d": // Duplicate note
		return m.duplicateNote()
	case "ctrl+l": // Copy note link
		return m.copyNoteLink()
	case "p": // Pin/unpin note
		return m.togglePin()
	}
	return m, nil
}

func (m *MainModel) quickJump() (tea.Model, tea.Cmd) {
	// Show a quick input to jump to note by name
	m.searchMode = true
	m.currentView = "search"
	m.searchInput.Focus()
	return m, textinput.Blink
}

func (m *MainModel) showRecentNotes() (tea.Model, tea.Cmd) {
	// Sort by modified date and show top 10
	m.sortMode = SortByModified
	m.SortNotes(m.sortMode)
	
	// Limit to 10 most recent
	if len(m.notes) > 10 {
		m.notes = m.notes[:10]
	}
	
	return m, nil
}

func (m *MainModel) duplicateNote() (tea.Model, tea.Cmd) {
	if m.currentNote == nil {
		return m, nil
	}
	
	// Create a copy with new timestamp
	newFilename := fmt.Sprintf("note_%d.json", time.Now().Unix())
	newPath := filepath.Join(m.dataDir, newFilename)
	
	duplicate := &Note{
		Title:     m.currentNote.Title + " (Copy)",
		Content:   m.currentNote.Content,
		Tags:      append([]string{}, m.currentNote.Tags...),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Path:      newPath,
		Encrypted: m.currentNote.Encrypted,
	}
	
	// Save duplicate
	data, err := json.MarshalIndent(duplicate, "", "  ")
	if err == nil {
		if err := utils.SafeWriteFile(newPath, data, 0600); err == nil {
			m.loadNotes()
			m.currentNote = duplicate
			return m, nil
		}
	}
	
	return m, nil
}

func (m *MainModel) copyNoteLink() (tea.Model, tea.Cmd) {
	if m.currentNote == nil {
		return m, nil
	}
	
	// Generate internal link format: [[Note Title]]
	link := fmt.Sprintf("[[%s]]", m.currentNote.Title)
	
	// In a real implementation, this would copy to clipboard
	// For now, we'll just show it in a message
	m.currentNote.Content += "\n\n" + link
	m.saveCurrentNote()
	
	return m, nil
}

func (m *MainModel) togglePin() (tea.Model, tea.Cmd) {
	if m.currentNote == nil {
		return m, nil
	}
	
	// Add/remove "pinned" tag
	pinned := false
	newTags := []string{}
	for _, tag := range m.currentNote.Tags {
		if tag == "pinned" {
			pinned = true
		} else {
			newTags = append(newTags, tag)
		}
	}
	
	if pinned {
		m.currentNote.Tags = newTags
	} else {
		m.currentNote.Tags = append(newTags, "pinned")
	}
	
	m.saveCurrentNote()
	m.loadNotes()
	
	return m, nil
}

