package models

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportNotes exports all notes to a directory or archive
func (m *MainModel) ExportNotes(format, outputPath string) error {
	switch format {
	case "markdown", "md":
		return m.exportToMarkdown(outputPath)
	case "json":
		return m.exportToJSON(outputPath)
	case "tar":
		return m.exportToTar(outputPath)
	case "zip":
		return m.exportToZip(outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (m *MainModel) exportToMarkdown(outputPath string) error {
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}
	
	return filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			note, err := m.loadNoteFromFile(path)
			if err != nil {
				return nil
			}
			
			// Create markdown file
			relPath, _ := filepath.Rel(m.dataDir, path)
			mdPath := strings.TrimSuffix(relPath, ".json") + ".md"
			mdPath = filepath.Join(outputPath, mdPath)
			
			// Create directory if needed
			if err := os.MkdirAll(filepath.Dir(mdPath), 0755); err != nil {
				return err
			}
			
			// Write markdown
			content := fmt.Sprintf("# %s\n\n", note.Title)
			if len(note.Tags) > 0 {
				content += fmt.Sprintf("Tags: %s\n\n", strings.Join(note.Tags, ", "))
			}
			content += fmt.Sprintf("Created: %s\nUpdated: %s\n\n", 
				note.CreatedAt.Format(time.RFC3339),
				note.UpdatedAt.Format(time.RFC3339))
			content += note.Content
			
			return os.WriteFile(mdPath, []byte(content), 0644)
		}
		
		return nil
	})
}

func (m *MainModel) exportToJSON(outputPath string) error {
	var notes []*Note
	
	err := filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			note, err := m.loadNoteFromFile(path)
			if err != nil {
				return nil
			}
			notes = append(notes, note)
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputPath, data, 0644)
}

func (m *MainModel) exportToTar(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	var tw *tar.Writer
	if strings.HasSuffix(outputPath, ".gz") {
		gzw := gzip.NewWriter(file)
		defer gzw.Close()
		tw = tar.NewWriter(gzw)
	} else {
		tw = tar.NewWriter(file)
	}
	defer tw.Close()
	
	return filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			note, err := m.loadNoteFromFile(path)
			if err != nil {
				return nil
			}
			
			relPath, _ := filepath.Rel(m.dataDir, path)
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath
			
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			
			data, err := json.Marshal(note)
			if err != nil {
				return err
			}
			
			_, err = tw.Write(data)
			return err
		}
		
		return nil
	})
}

func (m *MainModel) exportToZip(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	zw := zip.NewWriter(file)
	defer zw.Close()
	
	return filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			note, err := m.loadNoteFromFile(path)
			if err != nil {
				return nil
			}
			
			relPath, _ := filepath.Rel(m.dataDir, path)
			f, err := zw.Create(relPath)
			if err != nil {
				return err
			}
			
			data, err := json.Marshal(note)
			if err != nil {
				return err
			}
			
			_, err = f.Write(data)
			return err
		}
		
		return nil
	})
}

// ImportNotes imports notes from a directory or archive
func (m *MainModel) ImportNotes(format, inputPath string) error {
	switch format {
	case "markdown", "md":
		return m.importFromMarkdown(inputPath)
	case "json":
		return m.importFromJSON(inputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (m *MainModel) importFromMarkdown(inputPath string) error {
	return filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			
			// Parse markdown to extract title and content
			lines := strings.Split(string(content), "\n")
			title := info.Name()
			noteContent := string(content)
			
			if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
				title = strings.TrimPrefix(lines[0], "# ")
				noteContent = strings.Join(lines[1:], "\n")
			}
			
			// Create note
			note := &Note{
				Title:     strings.TrimSuffix(title, ".md"),
				Content:   noteContent,
				Tags:      []string{},
				CreatedAt: info.ModTime(),
				UpdatedAt: time.Now(),
				Path:      filepath.Join(m.dataDir, strings.TrimSuffix(info.Name(), ".md")+".json"),
				Encrypted: false,
			}
			
			// Save note
			data, err := json.MarshalIndent(note, "", "  ")
			if err != nil {
				return err
			}
			
			return os.WriteFile(note.Path, data, 0600)
		}
		
		return nil
	})
}

func (m *MainModel) importFromJSON(inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	
	var notes []*Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return err
	}
	
	for _, note := range notes {
		// Generate new path
		filename := fmt.Sprintf("imported_%d.json", time.Now().UnixNano())
		note.Path = filepath.Join(m.dataDir, filename)
		note.UpdatedAt = time.Now()
		
		// Save note
		data, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			continue
		}
		
		os.WriteFile(note.Path, data, 0600)
	}
	
	return nil
}

