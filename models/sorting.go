package models

import (
	"sort"
	"strings"
	"time"
)

type SortMode string

const (
	SortByDateNewest SortMode = "date_newest"
	SortByDateOldest SortMode = "date_oldest"
	SortByTitleAsc   SortMode = "title_asc"
	SortByTitleDesc  SortMode = "title_desc"
	SortByModified   SortMode = "modified"
)

func (m *MainModel) SortNotes(mode SortMode) {
	switch mode {
	case SortByDateNewest:
		sort.Slice(m.notes, func(i, j int) bool {
			noteI, _ := m.loadNoteFromFile(m.notes[i].path)
			noteJ, _ := m.loadNoteFromFile(m.notes[j].path)
			if noteI == nil || noteJ == nil {
				return false
			}
			return noteI.CreatedAt.After(noteJ.CreatedAt)
		})
	case SortByDateOldest:
		sort.Slice(m.notes, func(i, j int) bool {
			noteI, _ := m.loadNoteFromFile(m.notes[i].path)
			noteJ, _ := m.loadNoteFromFile(m.notes[j].path)
			if noteI == nil || noteJ == nil {
				return false
			}
			return noteI.CreatedAt.Before(noteJ.CreatedAt)
		})
	case SortByTitleAsc:
		sort.Slice(m.notes, func(i, j int) bool {
			return strings.ToLower(m.notes[i].title) < strings.ToLower(m.notes[j].title)
		})
	case SortByTitleDesc:
		sort.Slice(m.notes, func(i, j int) bool {
			return strings.ToLower(m.notes[i].title) > strings.ToLower(m.notes[j].title)
		})
	case SortByModified:
		sort.Slice(m.notes, func(i, j int) bool {
			noteI, _ := m.loadNoteFromFile(m.notes[i].path)
			noteJ, _ := m.loadNoteFromFile(m.notes[j].path)
			if noteI == nil || noteJ == nil {
				return false
			}
			return noteI.UpdatedAt.After(noteJ.UpdatedAt)
		})
	}
}

func (m *MainModel) FilterNotesByTag(tag string) []NoteItem {
	if tag == "" {
		return m.notes
	}
	
	filtered := []NoteItem{}
	for _, note := range m.notes {
		if note.isFolder {
			continue
		}
		
		loadedNote, err := m.loadNoteFromFile(note.path)
		if err != nil {
			continue
		}
		
		for _, noteTag := range loadedNote.Tags {
			if strings.EqualFold(noteTag, tag) {
				filtered = append(filtered, note)
				break
			}
		}
	}
	
	return filtered
}

func (m *MainModel) FilterNotesByDateRange(start, end time.Time) []NoteItem {
	filtered := []NoteItem{}
	for _, note := range m.notes {
		if note.isFolder {
			continue
		}
		
		loadedNote, err := m.loadNoteFromFile(note.path)
		if err != nil {
			continue
		}
		
		if (loadedNote.CreatedAt.After(start) || loadedNote.CreatedAt.Equal(start)) &&
			(loadedNote.CreatedAt.Before(end) || loadedNote.CreatedAt.Equal(end)) {
			filtered = append(filtered, note)
		}
	}
	
	return filtered
}

