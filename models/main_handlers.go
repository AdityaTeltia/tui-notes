package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMainKey handles keyboard input in the main two-pane view
func (m *MainModel) handleMainKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)
	
	// Navigation
	case "up", "k":
		if m.sidebarCursor > 0 {
			m.sidebarCursor--
		}
		m.selectNote()
		return m, nil
	case "down", "j":
		notesToUse := m.notes
		if m.showFiltered && len(m.filteredNotes) > 0 {
			notesToUse = m.filteredNotes
		}
		if m.sidebarCursor < len(notesToUse)-1 {
			m.sidebarCursor++
		}
		m.selectNote()
		return m, nil
	
	// Expand/collapse folders
	case "left":
		if m.sidebarCursor < len(m.notes) {
			note := m.notes[m.sidebarCursor]
			if note.isFolder {
				m.sidebarExpanded[note.title] = !m.sidebarExpanded[note.title]
			}
		}
		return m, nil
	case "l", "right":
		if m.sidebarCursor < len(m.notes) {
			note := m.notes[m.sidebarCursor]
			if note.isFolder {
				m.sidebarExpanded[note.title] = !m.sidebarExpanded[note.title]
			}
		}
		return m, nil
	
	// Edit note
	case "enter", "e":
		notesToUse := m.notes
		if m.showFiltered && len(m.filteredNotes) > 0 {
			notesToUse = m.filteredNotes
		}
		
		if m.sidebarCursor < len(notesToUse) {
			note := notesToUse[m.sidebarCursor]
			if !note.isFolder {
				loadedNote, err := m.loadNoteFromFile(note.path)
				if err == nil {
					m.currentNote = loadedNote
					m.editor.SetValue(loadedNote.Content)
					m.editorText = loadedNote.Content
					m.currentView = "editor"
				}
			}
		}
		return m, nil
	
	// New note
	case "n":
		m.createNewNote()
		return m, nil
	
	// New note from template
	case "ctrl+n":
		m.templates = m.LoadTemplates()
		if len(m.templates) > 0 {
			m.currentView = "template_select"
		} else {
			m.createNewNote()
		}
		return m, nil
	
	// New folder
	case "N":
		// Create new folder - for now, just create a note with folder-like name
		// In a full implementation, you'd create a directory
		return m, nil
	
	// Edit title
	case "t":
		if m.currentNote != nil {
			m.editingTitle = true
			m.titleInput.SetValue(m.currentNote.Title)
			m.titleInput.Focus()
			m.currentView = "title_edit"
			return m, textinput.Blink
		}
		return m, nil
	
	// Version history
	case "ctrl+h":
		if m.currentNote != nil {
			versions, err := m.LoadVersions(m.currentNote.Path)
			if err == nil {
				m.versions = versions
				m.selectedVersion = 0
				m.currentView = "versions"
			}
		}
		return m, nil
	
	// Theme selector
	case "ctrl+t":
		m.cycleTheme()
		return m, nil
	
	// Archive/Delete
	case "backspace", "d":
		if m.sidebarCursor < len(m.notes) {
			note := m.notes[m.sidebarCursor]
			m.deleteNote(note.path)
		}
		return m, nil
	
	// Toggle sidebar
	case "tab":
		m.showSidebar = !m.showSidebar
		return m, nil
	
	// Search
	case "/":
		m.searchMode = true
		m.currentView = "search"
		return m, textinput.Blink
	
	// Tags
	case "#":
		m.currentView = "tags"
		return m, nil
	
	// Sorting
	case "s":
		m.cycleSortMode()
		return m, nil
	
	// Filter by tag
	case "f":
		m.showFilterMenu()
		return m, nil
	
	// Clear filter
	case "ctrl+f":
		m.clearFilter()
		return m, nil
	
	// Quick actions
	case "g":
		return m.quickJump()
	case "r":
		return m.showRecentNotes()
	case "ctrl+d":
		return m.duplicateNote()
	case "ctrl+l":
		return m.copyNoteLink()
	}
	
	return m, nil
}

func (m *MainModel) cycleSortMode() {
	modes := []SortMode{
		SortByModified,
		SortByDateNewest,
		SortByDateOldest,
		SortByTitleAsc,
		SortByTitleDesc,
	}
	
	currentIdx := 0
	for i, mode := range modes {
		if mode == m.sortMode {
			currentIdx = i
			break
		}
	}
	
	nextIdx := (currentIdx + 1) % len(modes)
	m.sortMode = modes[nextIdx]
	m.SortNotes(m.sortMode)
}

func (m *MainModel) showFilterMenu() {
	// For now, just toggle filter mode
	// In full implementation, show a menu to select tag
	m.showFiltered = !m.showFiltered
	if m.showFiltered && m.filterTag != "" {
		m.filteredNotes = m.FilterNotesByTag(m.filterTag)
	}
}

func (m *MainModel) clearFilter() {
	m.showFiltered = false
	m.filterTag = ""
	m.filteredNotes = nil
}

// selectNote selects the note at the current cursor position
func (m *MainModel) selectNote() {
	notesToUse := m.notes
	if m.showFiltered && len(m.filteredNotes) > 0 {
		notesToUse = m.filteredNotes
	}
	
	if m.sidebarCursor < len(notesToUse) {
		note := notesToUse[m.sidebarCursor]
		if !note.isFolder {
			loadedNote, err := m.loadNoteFromFile(note.path)
			if err == nil {
				m.currentNote = loadedNote
			}
		}
	}
}

// handleTitleEditKey handles keyboard input when editing title
func (m *MainModel) handleTitleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Save title
		if m.currentNote != nil {
			m.currentNote.Title = m.titleInput.Value()
			m.saveCurrentNote()
			m.loadNotes() // Refresh sidebar to show updated title
		}
		m.editingTitle = false
		m.currentView = "main"
		return m, nil
	case "esc":
		m.editingTitle = false
		m.currentView = "main"
		return m, nil
	}
	
	var cmd tea.Cmd
	m.titleInput, cmd = m.titleInput.Update(msg)
	return m, cmd
}

func (m *MainModel) renderTitleEdit() string {
	var s strings.Builder
	s.WriteString("Edit Title\n\n")
	s.WriteString(m.titleInput.View())
	s.WriteString("\n\nPress Enter to save, Esc to cancel")
	return s.String()
}

