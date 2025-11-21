package models

import (
	"fmt"
	"regexp"
	"strings"
)

var linkRegex = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// ExtractLinks extracts all [[link]] references from note content
func (n *Note) ExtractLinks() []string {
	matches := linkRegex.FindAllStringSubmatch(n.Content, -1)
	links := make([]string, 0, len(matches))
	
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, strings.TrimSpace(match[1]))
		}
	}
	
	return links
}

// FindBacklinks finds all notes that link to this note
func (m *MainModel) FindBacklinks(noteTitle string) []*Note {
	backlinks := []*Note{}
	
	for _, noteItem := range m.notes {
		if noteItem.isFolder {
			continue
		}
		
		note, err := m.loadNoteFromFile(noteItem.path)
		if err != nil {
			continue
		}
		
		links := note.ExtractLinks()
		for _, link := range links {
			if strings.EqualFold(link, noteTitle) {
				backlinks = append(backlinks, note)
				break
			}
		}
	}
	
	return backlinks
}

// RenderLinks renders [[links]] as clickable (highlighted) text
func RenderLinks(content string) string {
	return linkRegex.ReplaceAllStringFunc(content, func(match string) string {
		linkText := linkRegex.FindStringSubmatch(match)
		if len(linkText) > 1 {
			// Highlight link text
			return fmt.Sprintf("[[%s]]", linkText[1])
		}
		return match
	})
}

// ResolveLink finds a note by title (fuzzy match)
func (m *MainModel) ResolveLink(linkTitle string) *Note {
	// Exact match first
	for _, noteItem := range m.notes {
		if noteItem.isFolder {
			continue
		}
		
		if strings.EqualFold(noteItem.title, linkTitle) {
			note, err := m.loadNoteFromFile(noteItem.path)
			if err == nil {
				return note
			}
		}
	}
	
	// Fuzzy match
	linkLower := strings.ToLower(linkTitle)
	for _, noteItem := range m.notes {
		if noteItem.isFolder {
			continue
		}
		
		if strings.Contains(strings.ToLower(noteItem.title), linkLower) {
			note, err := m.loadNoteFromFile(noteItem.path)
			if err == nil {
				return note
			}
		}
	}
	
	return nil
}

