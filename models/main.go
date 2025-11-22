package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainModel struct {
	dataDir     string
	username    string
	currentView string // "main", "editor", "title_edit", "search", "tags"
	
	// Two-pane layout
	sidebarCursor int
	sidebarExpanded map[string]bool // Track expanded folders
	showSidebar   bool
	
	// Browser
	browserList list.Model
	notes       []NoteItem
	
	// Editor
	editor      textinput.Model
	editorText  string
	editorMode  string // "normal", "insert", "vim"
	currentNote *Note
	editingTitleInEditor bool // Whether we're editing title in the editor view
	
	// Title editing
	titleInput  textinput.Model
	editingTitle bool
	
	// Preview
	previewViewport viewport.Model
	previewContent  string
	
	// Search
	searchInput textinput.Model
	searchResults []NoteItem
	searchQuery   string
	searchMode    bool
	
	// Tags
	tagsInput textinput.Model
	showTags  bool
	
	// Sorting and filtering
	sortMode    SortMode
	filterTag   string
	filteredNotes []NoteItem
	showFiltered bool
	
	// Quick actions
	recentNotes []NoteItem
	pinnedNotes []NoteItem
	
	// Templates
	templates      []Template
	selectedTemplate int
	
	// Version history
	versions       []Version
	showVersions   bool
	selectedVersion int
	
	// Theme
	currentTheme   string
	
	width  int
	height int
}

type NoteItem struct {
	title    string
	path     string
	isFolder bool
	tags     []string
}

func (i NoteItem) FilterValue() string { return i.title }
func (i NoteItem) Title() string {
	if i.isFolder {
		return "📁 " + i.title
	}
	return "📄 " + i.title
}
func (i NoteItem) Description() string {
	if len(i.tags) > 0 {
		return strings.Join(i.tags, ", ")
	}
	return i.path
}

func NewMainModel(username, dataDir string) *MainModel {
	m := &MainModel{
		username:      username,
		dataDir:       dataDir,
		currentView:   "main",
		sidebarCursor: 0,
		sidebarExpanded: make(map[string]bool),
		showSidebar:   true,
		width:              80,  // Default width
		height:             24,  // Default height
		editorMode:         "insert", // Start in insert mode
		editingTitleInEditor: false,
		sortMode:           SortByModified, // Default sort by modified date
		currentTheme:       "default",
	}
	
	// Initialize editor
	m.editor = textinput.New()
	m.editor.Placeholder = "Start typing..."
	m.editor.Focus()
	m.editorMode = "insert" // Start in insert mode
	
	// Initialize search
	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search notes..."
	m.searchInput.Focus()
	
	// Initialize tags input
	m.tagsInput = textinput.New()
	m.tagsInput.Placeholder = "Enter tags (comma-separated)..."
	
	// Initialize title input
	m.titleInput = textinput.New()
	m.titleInput.Placeholder = "Note title..."
	m.titleInput.Focus()
	
	// Initialize browser list
	items := []list.Item{}
	m.browserList = list.New(items, list.NewDefaultDelegate(), 80, 20)
	m.browserList.Title = "Notes"
	m.browserList.SetShowStatusBar(true)
	m.browserList.SetFilteringEnabled(true)
	
	// Initialize preview viewport
	m.previewViewport = viewport.New(80, 20)
	
	// Load initial notes
	m.loadNotes()
	
	return m
}

func (m *MainModel) Init() tea.Cmd {
	// Enter alt screen for proper TUI rendering
	return tea.EnterAltScreen
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width > 4 {
			m.browserList.SetWidth(m.width - 4)
			m.editor.Width = m.width - 4
			m.previewViewport.Width = m.width - 4
		}
		if m.height > 6 {
			m.browserList.SetHeight(m.height - 6)
			m.previewViewport.Height = m.height - 4
		}
		return m, nil
		
	case tea.KeyMsg:
		switch m.currentView {
		case "main":
			return m.handleMainKey(msg)
		case "editor":
			return m.handleEditorKey(msg)
		case "title_edit":
			return m.handleTitleEditKey(msg)
		case "preview":
			return m.handlePreviewKey(msg)
		case "search":
			return m.handleSearchKey(msg)
		case "tags":
			return m.handleTagsKey(msg)
		case "template_select":
			return m.handleTemplateSelectKey(msg)
		case "versions":
			return m.handleVersionsKey(msg)
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *MainModel) View() string {
	// Ensure we have minimum dimensions
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}
	
	switch m.currentView {
	case "main":
		return m.RenderTwoPane()
	case "editor":
		return m.renderEditor()
	case "title_edit":
		return m.renderTitleEdit()
	case "preview":
		return m.renderPreview()
	case "search":
		return m.renderSearch()
	case "tags":
		return m.renderTags()
	case "template_select":
		return m.renderTemplateSelect()
	case "versions":
		return m.renderVersions()
	default:
		return m.RenderTwoPane()
	}
}

// Old menu handler removed - using two-pane layout now

func (m *MainModel) handleBrowserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "main"
		return m, nil
	case "enter":
		selected := m.browserList.SelectedItem()
		if selected != nil {
			item := selected.(NoteItem)
			if item.isFolder {
				// Navigate into folder
				m.dataDir = filepath.Join(m.dataDir, item.title)
				m.loadNotes()
			} else {
				// Open note
				m.openNote(item.path)
			}
		}
		return m, nil
	case "n":
		m.createNewNote()
		return m, nil
	case "d":
		selected := m.browserList.SelectedItem()
		if selected != nil {
			item := selected.(NoteItem)
			m.deleteNote(item.path)
		}
		return m, nil
	case "e":
		selected := m.browserList.SelectedItem()
		if selected != nil {
			item := selected.(NoteItem)
			m.openNote(item.path)
		}
		return m, nil
	case "p":
		selected := m.browserList.SelectedItem()
		if selected != nil {
			item := selected.(NoteItem)
			m.previewNote(item.path)
		}
		return m, nil
	}
	
	// Delegate to list
	var cmd tea.Cmd
	m.browserList, cmd = m.browserList.Update(msg)
	return m, cmd
}

func (m *MainModel) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editorMode == "vim" {
		return m.handleVimKey(msg)
	}
	
	// Handle title editing in editor
	if m.editingTitleInEditor {
		switch msg.String() {
		case "enter":
			// Save title and switch to content editing
			if m.currentNote != nil {
				m.currentNote.Title = m.titleInput.Value()
				m.saveCurrentNote()
				m.loadNotes() // Refresh sidebar
			}
			m.editingTitleInEditor = false
			m.editor.Focus()
			return m, textinput.Blink
		case "esc":
			m.editingTitleInEditor = false
			m.editor.Focus()
			return m, textinput.Blink
		}
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(msg)
		return m, cmd
	}
	
	// Handle navigation/command keys
	switch msg.String() {
	case "esc":
		if m.editorMode == "insert" {
			m.editorMode = "normal"
			return m, nil
		}
		m.saveCurrentNote()
		m.currentView = "main"
		m.loadNotes()
		return m, nil
	case "ctrl+c":
		m.saveCurrentNote()
		m.currentView = "main"
		m.loadNotes()
		return m, nil
	case "ctrl+s":
		m.saveCurrentNote()
		m.loadNotes() // Refresh sidebar after save
		return m, nil
	case "ctrl+p":
		m.previewCurrentNote()
		return m, nil
	case "ctrl+t":
		// Edit title from editor (only Ctrl+T, not just 't')
		if m.currentNote != nil {
			m.editingTitleInEditor = true
			m.titleInput.SetValue(m.currentNote.Title)
			m.titleInput.Focus()
			return m, textinput.Blink
		}
		return m, nil
	case "i":
		if m.editorMode == "normal" {
			m.editorMode = "insert"
			m.editor.Focus()
			return m, textinput.Blink
		}
	case "v":
		if m.editorMode == "normal" {
			m.editorMode = "vim"
			return m, nil
		}
	}
	
	// Handle text input in insert mode
	var cmd tea.Cmd
	if m.editorMode == "insert" {
		// Use the textinput component for proper input handling
		m.editor, cmd = m.editor.Update(msg)
		m.editorText = m.editor.Value()
	}
	return m, cmd
}

func (m *MainModel) handleVimKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		m.editorMode = "insert"
		m.editor.Focus()
		return m, textinput.Blink
	case "esc":
		m.editorMode = "normal"
		return m, nil
	case "h", "left":
		// Move cursor left
		return m, nil
	case "l", "right":
		// Move cursor right
		return m, nil
	case "j", "down":
		// Move cursor down
		return m, nil
	case "k", "up":
		// Move cursor up
		return m, nil
	case "w":
		// Save
		m.saveCurrentNote()
		return m, nil
	case "q":
		m.saveCurrentNote()
		m.currentView = "main"
		m.loadNotes()
		return m, nil
	}
	return m, nil
}

func (m *MainModel) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "main"
		return m, nil
	case "e":
		if m.currentNote != nil {
			m.currentView = "editor"
			m.editor.SetValue(m.currentNote.Content)
			m.editorText = m.currentNote.Content
		}
		return m, nil
	}
	
	var cmd tea.Cmd
	m.previewViewport, cmd = m.previewViewport.Update(msg)
	return m, cmd
}

func (m *MainModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = "main"
		m.searchMode = false
		return m, nil
	case "enter":
		m.performSearch()
		return m, nil
	}
	
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = m.searchInput.Value()
	return m, cmd
}

func (m *MainModel) handleTagsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = "main"
		return m, nil
	case "enter":
		m.addTagsToNote()
		return m, nil
	}
	
	var cmd tea.Cmd
	m.tagsInput, cmd = m.tagsInput.Update(msg)
	return m, cmd
}

// Old renderMenu removed - using two-pane layout now

func (m *MainModel) renderBrowser() string {
	return m.browserList.View()
}

func (m *MainModel) renderEditor() string {
	var s strings.Builder
	
	// If editing title, show title input
	if m.editingTitleInEditor {
		titleLabel := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Render("Edit Title:")
		
		s.WriteString(titleLabel + "\n\n")
		s.WriteString(m.titleInput.View())
		s.WriteString("\n\n")
		help := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Enter: Save title and edit content | Esc: Cancel")
		s.WriteString(help)
		return s.String()
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Editor")
	
	if m.currentNote != nil {
		title += " - " + m.currentNote.Title
	}
	
	s.WriteString(title + "\n\n")
	
	// Show title field (editable)
	if m.currentNote != nil {
		titleLabel := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Title: " + m.currentNote.Title)
		s.WriteString(titleLabel)
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(" (Press Ctrl+T to edit)"))
		s.WriteString("\n")
		separatorLen := m.width - 4
		if separatorLen > 0 {
			s.WriteString(strings.Repeat("─", separatorLen) + "\n\n")
		} else {
			s.WriteString("─\n\n")
		}
	}
	
	if m.editorMode == "vim" {
		mode := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("-- %s MODE --", strings.ToUpper(m.editorMode)))
		s.WriteString(mode + "\n")
	} else if m.editorMode == "normal" {
		mode := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("-- NORMAL MODE -- Press 'i' to insert")
		s.WriteString(mode + "\n")
	}
	
	// Show editor content (multi-line)
	editorHeight := m.height - 12
	if editorHeight < 1 {
		editorHeight = 1
	}
	editorStyle := lipgloss.NewStyle().
		Width(m.width - 4).
		Height(editorHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	
	content := m.editorText
	if content == "" {
		content = m.editor.Placeholder
	}
	
	// Split content into lines for display
	lines := strings.Split(content, "\n")
	if len(lines) > editorHeight {
		lines = lines[:editorHeight]
	}
	displayContent := strings.Join(lines, "\n")
	
	s.WriteString(editorStyle.Render(displayContent))
	
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(1).
		Render("Ctrl+S: Save | Ctrl+T: Edit title | Ctrl+P: Preview | Esc: Back | i: Insert")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

func (m *MainModel) renderPreview() string {
	var s strings.Builder
	
	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Preview")
	
	if m.currentNote != nil {
		title += " - " + m.currentNote.Title
	}
	
	s.WriteString(title + "\n")
	separatorLen := m.width - 2
	if separatorLen > 0 {
		s.WriteString(strings.Repeat("─", separatorLen) + "\n\n")
	} else {
		s.WriteString("─\n\n")
	}
	
	// Preview content in viewport
	previewHeight := m.height - 6 // Reserve space for title, separator, and help
	if previewHeight < 1 {
		previewHeight = 1
	}
	
	// Update viewport size if needed
	if m.previewViewport.Height != previewHeight {
		m.previewViewport.Height = previewHeight
	}
	if m.previewViewport.Width != m.width-4 {
		m.previewViewport.Width = m.width - 4
	}
	
	s.WriteString(m.previewViewport.View())
	
	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(1).
		Render("↑/↓: Scroll | Esc/q: Back | e: Edit")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

func (m *MainModel) renderSearch() string {
	var s strings.Builder
	
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Search Notes")
	
	s.WriteString(title + "\n\n")
	s.WriteString(m.searchInput.View() + "\n\n")
	
	if len(m.searchResults) > 0 {
		s.WriteString(fmt.Sprintf("Found %d results:\n\n", len(m.searchResults)))
		for _, result := range m.searchResults {
			s.WriteString(fmt.Sprintf("  • %s\n", result.Title()))
		}
	}
	
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(1).
		Render("Enter: Search | Esc: Back")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

func (m *MainModel) renderTags() string {
	var s strings.Builder
	
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Manage Tags")
	
	s.WriteString(title + "\n\n")
	s.WriteString(m.tagsInput.View())
	
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingTop(1).
		Render("Enter: Add tags | Esc: Back")
	
	s.WriteString("\n" + help)
	
	return s.String()
}

