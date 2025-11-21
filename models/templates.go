package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ssh-notes/terminal-notes/utils"
)

type Template struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
}

func (m *MainModel) LoadTemplates() []Template {
	templatesDir := filepath.Join(m.dataDir, ".templates")
	templates := []Template{}
	
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Create default templates
		m.createDefaultTemplates(templatesDir)
	}
	
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return templates
	}
	
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(templatesDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			
			var template Template
			if err := json.Unmarshal(data, &template); err == nil {
				templates = append(templates, template)
			}
		}
	}
	
	return templates
}

func (m *MainModel) createDefaultTemplates(templatesDir string) {
	os.MkdirAll(templatesDir, 0700)
	
	defaultTemplates := []Template{
		{
			Name:        "meeting",
			Description: "Meeting notes template",
			Title:       "Meeting Notes - {{date}}",
			Content:     "# Meeting Notes\n\n**Date:** {{date}}\n**Attendees:** \n**Agenda:**\n\n## Notes\n\n## Action Items\n\n- [ ] \n",
			Tags:        []string{"meeting"},
		},
		{
			Name:        "journal",
			Description: "Daily journal template",
			Title:       "Journal - {{date}}",
			Content:     "# Journal Entry\n\n**Date:** {{date}}\n\n## Today's Highlights\n\n\n## Thoughts\n\n\n## Tomorrow's Goals\n\n- [ ] \n",
			Tags:        []string{"journal"},
		},
		{
			Name:        "code",
			Description: "Code snippet template",
			Title:       "Code: {{title}}",
			Content:     "# {{title}}\n\n**Language:** \n**Description:**\n\n```\n\n```\n",
			Tags:        []string{"code"},
		},
		{
			Name:        "todo",
			Description: "To-do list template",
			Title:       "Todo List - {{date}}",
			Content:     "# Todo List\n\n**Date:** {{date}}\n\n## Tasks\n\n- [ ] \n- [ ] \n- [ ] \n",
			Tags:        []string{"todo"},
		},
	}
	
	for _, tmpl := range defaultTemplates {
		path := filepath.Join(templatesDir, tmpl.Name+".json")
		data, _ := json.MarshalIndent(tmpl, "", "  ")
		os.WriteFile(path, data, 0644)
	}
}

func (m *MainModel) CreateNoteFromTemplate(template Template) {
	// Replace template variables
	now := time.Now()
	title := strings.ReplaceAll(template.Title, "{{date}}", now.Format("2006-01-02"))
	title = strings.ReplaceAll(title, "{{title}}", "Untitled")
	content := strings.ReplaceAll(template.Content, "{{date}}", now.Format("2006-01-02"))
	content = strings.ReplaceAll(content, "{{title}}", title)
	
	// Create note
	filename := fmt.Sprintf("note_%d.json", time.Now().Unix())
	path := filepath.Join(m.dataDir, filename)
	
	note := &Note{
		Title:     title,
		Content:   content,
		Tags:      append([]string{}, template.Tags...),
		CreatedAt: now,
		UpdatedAt: now,
		Path:      path,
		Encrypted: false,
	}
	
	// Save note
	data, err := json.MarshalIndent(note, "", "  ")
	if err == nil {
		if err := utils.SafeWriteFile(path, data, 0600); err == nil {
			m.currentNote = note
			m.editor.SetValue(note.Content)
			m.editorText = note.Content
			m.currentView = "editor"
			m.loadNotes()
		}
	}
}

