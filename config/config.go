package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Security SecurityConfig `json:"security"`
	Data     DataConfig     `json:"data"`
	Logging  LoggingConfig  `json:"logging"`
}

type ServerConfig struct {
	Port    string `json:"port"`
	HostKey string `json:"host_key"`
	DataDir string `json:"data_dir"`
}

type SecurityConfig struct {
	AuthMode          string `json:"auth_mode"` // "password", "key", "both", "none"
	RequirePassword   bool   `json:"require_password"`
	PasswordFile      string `json:"password_file"`
	AuthorizedKeysFile string `json:"authorized_keys_file"`
	MaxLoginAttempts  int    `json:"max_login_attempts"`
	SessionTimeout    int    `json:"session_timeout"` // seconds
}

type DataConfig struct {
	BaseDir        string `json:"base_dir"`
	EnableEncryption bool  `json:"enable_encryption"`
	MaxNoteSize    int64  `json:"max_note_size"` // bytes
	BackupEnabled  bool   `json:"backup_enabled"`
	BackupInterval int    `json:"backup_interval"` // minutes
}

type LoggingConfig struct {
	Level      string `json:"level"` // "debug", "info", "warn", "error"
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`    // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`     // days
	Compress   bool   `json:"compress"`
}

var DefaultConfig = Config{
	Server: ServerConfig{
		Port:    "2222",
		HostKey: "host_key",
		DataDir: "./data",
	},
	Security: SecurityConfig{
		AuthMode:         "none",
		RequirePassword:  false,
		MaxLoginAttempts: 5,
		SessionTimeout:   3600,
	},
	Data: DataConfig{
		BaseDir:         "./data",
		EnableEncryption: false,
		MaxNoteSize:     10 * 1024 * 1024, // 10MB
		BackupEnabled:   false,
		BackupInterval:  60,
	},
	Logging: LoggingConfig{
		Level:      "info",
		File:       "",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	},
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig
	
	if path == "" {
		// Try default locations
		paths := []string{
			"./config.json",
			"/etc/ssh-notes/config.json",
			filepath.Join(os.Getenv("HOME"), ".ssh-notes", "config.json"),
		}
		
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config: %w", err)
			}
		}
	}
	
	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		path = "./config.json"
	}
	
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

