package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RecoverPanic recovers from panics and logs them
func RecoverPanic() {
	if r := recover(); r != nil {
		// Log the panic
		fmt.Fprintf(os.Stderr, "PANIC: %v\n", r)
		
		// Optionally save to recovery file
		recoveryFile := filepath.Join(os.TempDir(), fmt.Sprintf("ssh-notes-recovery-%d.log", time.Now().Unix()))
		if f, err := os.Create(recoveryFile); err == nil {
			fmt.Fprintf(f, "Panic recovered at %s\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(f, "Error: %v\n", r)
			f.Close()
		}
	}
}

// SafeWriteFile writes data to a file with atomic operation
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	// Write to temporary file first
	tmpPath := path + ".tmp"
	
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up on failure
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

// BackupFile creates a backup of a file
func BackupFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Nothing to backup
	}
	
	backupPath := path + ".backup." + time.Now().Format("20060102-150405")
	
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}
	
	return nil
}

