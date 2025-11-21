package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Version struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *MainModel) SaveVersion(note *Note) error {
	if note == nil {
		return fmt.Errorf("note is nil")
	}
	
	versionsDir := filepath.Join(m.dataDir, ".versions")
	if err := os.MkdirAll(versionsDir, 0700); err != nil {
		return err
	}
	
	versionID := fmt.Sprintf("%d", time.Now().Unix())
	versionFile := filepath.Join(versionsDir, fmt.Sprintf("%s_%s.json", 
		filepath.Base(note.Path), versionID))
	
	version := Version{
		ID:        versionID,
		Content:   note.Content,
		Title:     note.Title,
		CreatedAt: time.Now(),
	}
	
	data, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(versionFile, data, 0600)
}

func (m *MainModel) LoadVersions(notePath string) ([]Version, error) {
	versionsDir := filepath.Join(m.dataDir, ".versions")
	baseName := filepath.Base(notePath)
	prefix := baseName + "_"
	
	versions := []Version{}
	
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return versions, nil
	}
	
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) && 
		   strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(versionsDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			
			var version Version
			if err := json.Unmarshal(data, &version); err == nil {
				versions = append(versions, version)
			}
		}
	}
	
	// Sort by creation date (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.After(versions[j].CreatedAt)
	})
	
	return versions, nil
}

func (m *MainModel) RestoreVersion(note *Note, versionID string) error {
	versions, err := m.LoadVersions(note.Path)
	if err != nil {
		return err
	}
	
	for _, version := range versions {
		if version.ID == versionID {
			note.Content = version.Content
			note.Title = version.Title
			note.UpdatedAt = time.Now()
			m.saveCurrentNote()
			return nil
		}
	}
	
	return fmt.Errorf("version not found")
}

