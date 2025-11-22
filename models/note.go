package models

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/ssh-notes/terminal-notes/logger"
	"github.com/ssh-notes/terminal-notes/utils"
)

type Note struct {
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Path       string    `json:"path"`
	Encrypted  bool      `json:"encrypted"`
}

func (m *MainModel) loadNotes() {
	m.notes = []NoteItem{}
	
	// Ensure data directory exists
	if err := os.MkdirAll(m.dataDir, 0700); err != nil {
		return
	}
	
	// Load folders and notes
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			m.notes = append(m.notes, NoteItem{
				title:    entry.Name(),
				path:     filepath.Join(m.dataDir, entry.Name()),
				isFolder: true,
			})
		} else if strings.HasSuffix(entry.Name(), ".json") {
			note, err := m.loadNoteFromFile(filepath.Join(m.dataDir, entry.Name()))
			if err == nil {
				m.notes = append(m.notes, NoteItem{
					title: note.Title,
					path:  filepath.Join(m.dataDir, entry.Name()),
					tags:  note.Tags,
				})
			}
		}
	}
	
	// Apply sorting
	m.SortNotes(m.sortMode)
	
	// Apply filtering if active
	displayNotes := m.notes
	if m.showFiltered && len(m.filteredNotes) > 0 {
		displayNotes = m.filteredNotes
	}
	
	// Update list items
	items := make([]list.Item, len(displayNotes))
	for i, note := range displayNotes {
		items[i] = note
	}
	m.browserList.SetItems(items)
}

func (m *MainModel) loadNoteFromFile(path string) (*Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var note Note
	if err := json.Unmarshal(data, &note); err != nil {
		return nil, err
	}
	
	// Decrypt if needed
	if note.Encrypted {
		decrypted, err := decryptNote(note.Content)
		if err == nil {
			note.Content = decrypted
		}
	}
	
	note.Path = path
	return &note, nil
}

func (m *MainModel) createNewNote() {
	// Generate filename from timestamp
	filename := fmt.Sprintf("note_%d.json", time.Now().Unix())
	
	// Validate filename
	if err := utils.ValidateFilename(filename); err != nil {
		logger.Error("Invalid filename: %v", err)
		return
	}
	
	path := filepath.Join(m.dataDir, filename)
	
	// Sanitize path
	safePath, err := utils.SanitizePath(path)
	if err != nil {
		logger.Error("Invalid path: %v", err)
		return
	}
	path = safePath
	
	note := &Note{
		Title:     "Untitled Note",
		Content:   "",
		Tags:      []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Path:      path,
		Encrypted: false,
	}
	
	// Validate note data
	if err := utils.ValidateTitle(note.Title); err != nil {
		logger.Error("Invalid title: %v", err)
		return
	}
	if err := utils.ValidateContent(note.Content); err != nil {
		logger.Error("Invalid content: %v", err)
		return
	}
	if err := utils.ValidateTags(note.Tags); err != nil {
		logger.Error("Invalid tags: %v", err)
		return
	}
	
	// Save the note first using safe write
	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal note: %v", err)
		return
	}
	
	if err := utils.SafeWriteFile(path, data, 0600); err != nil {
		logger.Error("Failed to save note: %v", err)
		return
	}
	
	logger.LogRequest(m.username, "create_note", nil)
	
	m.currentNote = note
	m.editor.SetValue("")
	m.editorText = ""
	m.editingTitleInEditor = false
	m.currentView = "editor"
	m.loadNotes() // Refresh sidebar
}

func (m *MainModel) openNote(path string) {
	note, err := m.loadNoteFromFile(path)
	if err != nil {
		return
	}
	
	m.currentNote = note
	m.editor.SetValue(note.Content)
	m.editorText = note.Content
	m.editingTitleInEditor = false
	m.currentView = "editor"
	// Don't change view - stay in main view to show in two-pane
	// User can press 'e' or 'enter' to edit
}

func (m *MainModel) saveCurrentNote() {
	if m.currentNote == nil {
		return
	}
	
	// Validate content before saving
	if err := utils.ValidateContent(m.editorText); err != nil {
		logger.Error("Invalid content: %v", err)
		return
	}
	
	m.currentNote.Content = m.editorText
	m.currentNote.UpdatedAt = time.Now()
	
	// Extract title from first line if it's markdown (only if title wasn't manually set)
	if m.currentNote.Title == "Untitled Note" || strings.TrimSpace(m.currentNote.Title) == "" {
		content := strings.TrimSpace(m.editorText)
		if strings.HasPrefix(content, "# ") {
			lines := strings.Split(content, "\n")
			if len(lines) > 0 {
				newTitle := strings.TrimPrefix(lines[0], "# ")
				if err := utils.ValidateTitle(newTitle); err == nil {
					m.currentNote.Title = newTitle
				}
			}
		}
	}
	
	// Validate title
	if err := utils.ValidateTitle(m.currentNote.Title); err != nil {
		logger.Error("Invalid title: %v", err)
		return
	}
	
	// Validate tags
	if err := utils.ValidateTags(m.currentNote.Tags); err != nil {
		logger.Error("Invalid tags: %v", err)
		return
	}
	
	// Create backup before saving
	if err := utils.BackupFile(m.currentNote.Path); err != nil {
		logger.Warn("Failed to create backup: %v", err)
	}
	
	// Save version history before overwriting
	if err := m.SaveVersion(m.currentNote); err != nil {
		logger.Warn("Failed to save version: %v", err)
	}
	
	// Save to file using safe write
	data, err := json.MarshalIndent(m.currentNote, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal note: %v", err)
		return
	}
	
	if err := utils.SafeWriteFile(m.currentNote.Path, data, 0600); err != nil {
		logger.Error("Failed to save note: %v", err)
		return
	}
	
	logger.LogRequest(m.username, "save_note", nil)
}

func (m *MainModel) deleteNote(path string) {
	os.Remove(path)
	m.loadNotes()
}

func (m *MainModel) previewNote(path string) {
	note, err := m.loadNoteFromFile(path)
	if err != nil {
		return
	}
	
	m.currentNote = note
	m.previewContent = renderMarkdown(note.Content)
	m.previewViewport.SetContent(m.previewContent)
	m.currentView = "preview"
}

func (m *MainModel) previewCurrentNote() {
	if m.currentNote == nil {
		return
	}
	
	// Update content from editor
	m.currentNote.Content = m.editorText
	
	// Render markdown
	m.previewContent = renderMarkdown(m.currentNote.Content)
	
	// Set viewport content and ensure proper sizing
	m.previewViewport.SetContent(m.previewContent)
	m.previewViewport.Width = m.width - 4
	previewHeight := m.height - 6
	if previewHeight < 1 {
		previewHeight = 1
	}
	m.previewViewport.Height = previewHeight
	
	// Switch to preview view
	m.currentView = "preview"
}

func (m *MainModel) performSearch() {
	m.searchResults = []NoteItem{}
	query := strings.ToLower(m.searchQuery)
	
	if query == "" {
		return
	}
	
	// Search through all notes
	err := filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			note, err := m.loadNoteFromFile(path)
			if err != nil {
				return nil
			}
			
			// Search in title, content, and tags
			content := strings.ToLower(note.Content)
			title := strings.ToLower(note.Title)
			tags := strings.ToLower(strings.Join(note.Tags, " "))
			
			if strings.Contains(title, query) ||
				strings.Contains(content, query) ||
				strings.Contains(tags, query) {
				m.searchResults = append(m.searchResults, NoteItem{
					title: note.Title,
					path:  path,
					tags:  note.Tags,
				})
			}
		}
		
		return nil
	})
	
	if err != nil {
		return
	}
}

func (m *MainModel) addTagsToNote() {
	if m.currentNote == nil {
		return
	}
	
	tagsStr := m.tagsInput.Value()
	tags := strings.Split(tagsStr, ",")
	
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	
	m.currentNote.Tags = append(m.currentNote.Tags, tags...)
	m.saveCurrentNote()
	m.tagsInput.SetValue("")
}

func (m *MainModel) showExportImport() {
	// This would show a menu for export/import
	// For now, export/import can be done via CLI commands
}

func (m *MainModel) showSettings() {
	// TODO: Implement settings
}

// renderMarkdown provides an enhanced markdown renderer for terminal
func renderMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	var result strings.Builder
	inCodeBlock := false
	codeBlockLang := ""
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				result.WriteString("\n" + strings.Repeat("─", 60) + "\n")
				if codeBlockLang != "" {
					result.WriteString("Code: " + codeBlockLang + "\n")
					result.WriteString(strings.Repeat("─", 60) + "\n")
				}
			} else {
				inCodeBlock = false
				result.WriteString(strings.Repeat("─", 60) + "\n\n")
			}
			continue
		}
		
		if inCodeBlock {
			// In code block - preserve formatting
			result.WriteString(line + "\n")
			continue
		}
		
		// Headings with figlet-style rendering
		if strings.HasPrefix(line, "# ") {
			// H1 - Use figlet if available, otherwise bold with underline
			text := strings.TrimPrefix(line, "# ")
			figletText := renderFiglet(text, "standard")
			if figletText != "" {
				result.WriteString("\n" + figletText + "\n")
			} else {
				// Fallback: bold with underline
				result.WriteString("\n" + strings.ToUpper(text) + "\n")
				result.WriteString(strings.Repeat("═", len(text)) + "\n")
			}
		} else if strings.HasPrefix(line, "## ") {
			// H2 - Use smaller figlet font
			text := strings.TrimPrefix(line, "## ")
			figletText := renderFiglet(text, "small")
			if figletText != "" {
				result.WriteString("\n" + figletText + "\n")
			} else {
				// Fallback
				result.WriteString("\n" + text + "\n")
				result.WriteString(strings.Repeat("─", len(text)) + "\n")
			}
		} else if strings.HasPrefix(line, "### ") {
			// H3 - Bold text
			text := strings.TrimPrefix(line, "### ")
			result.WriteString("\n" + strings.ToUpper(text) + "\n")
		} else if strings.HasPrefix(line, "#### ") {
			// H4
			text := strings.TrimPrefix(line, "#### ")
			result.WriteString("\n" + text + "\n")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// List item
			item := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			// Check for checkbox
			if strings.HasPrefix(item, "[ ]") {
				result.WriteString("  ☐ " + strings.TrimPrefix(item, "[ ]") + "\n")
			} else if strings.HasPrefix(item, "[x]") || strings.HasPrefix(item, "[X]") {
				result.WriteString("  ☑ " + strings.TrimPrefix(strings.TrimPrefix(item, "[x]"), "[X]") + "\n")
			} else {
				result.WriteString("  • " + item + "\n")
			}
		} else if strings.HasPrefix(line, "> ") {
			// Blockquote
			quote := strings.TrimPrefix(line, "> ")
			result.WriteString("  │ " + quote + "\n")
		} else if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			// Bold text
			text := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(line, "**"), "**"), " ")
			result.WriteString(strings.ToUpper(text) + "\n")
		} else if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") && !strings.HasPrefix(line, "**") {
			// Italic text
			text := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(line, "*"), "*"), " ")
			result.WriteString(text + "\n")
		} else if strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`") {
			// Inline code
			code := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(line, "`"), "`"), " ")
			result.WriteString("[" + code + "]\n")
		} else if trimmed == "" {
			// Empty line
			result.WriteString("\n")
		} else if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "***") {
			// Horizontal rule
			result.WriteString(strings.Repeat("─", 60) + "\n")
		} else {
			// Regular text
			result.WriteString(line + "\n")
		}
	}
	
	return result.String()
}

// renderFiglet attempts to render text using figlet (if available)
// Returns empty string if figlet is not available
func renderFiglet(text, font string) string {
	// Try to execute figlet command
	cmd := exec.Command("figlet", "-f", font, "-w", "60", text)
	output, err := cmd.Output()
	if err != nil {
		// Figlet not available, return empty to use fallback
		return ""
	}
	return string(output)
}

