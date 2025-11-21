package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
	"github.com/ssh-notes/terminal-notes/config"
	"github.com/ssh-notes/terminal-notes/logger"
	"github.com/ssh-notes/terminal-notes/utils"
)

var (
	cfgPath    = flag.String("config", "", "Path to configuration file")
	hostKeyPath = flag.String("hostkey", "", "Path to SSH host key (overrides config)")
	port        = flag.String("port", "", "SSH server port (overrides config)")
	dataDir     = flag.String("data", "", "Data directory (overrides config)")
	version     = flag.Bool("version", false, "Show version information")
)

const Version = "1.0.0"

func main() {
	defer utils.RecoverPanic()
	
	// Check if running as CLI command
	if len(os.Args) > 1 && (os.Args[1] == "export" || os.Args[1] == "import") {
		runCLI()
		return
	}

	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("SSH Notes Server v%s\n", Version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command-line flags
	if *port != "" {
		cfg.Server.Port = *port
	}
	if *dataDir != "" {
		cfg.Server.DataDir = *dataDir
	}
	if *hostKeyPath != "" {
		cfg.Server.HostKey = *hostKeyPath
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.File); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.Server.DataDir, 0700); err != nil {
		logger.Fatal("Failed to create data directory: %v", err)
	}

	// Load or generate host key
	hostKeyFile := cfg.Server.HostKey
	if _, err := os.Stat(hostKeyFile); os.IsNotExist(err) {
		logger.Info("Generating host key at %s", hostKeyFile)
		if err := generateHostKey(hostKeyFile); err != nil {
			logger.Fatal("Failed to generate host key: %v", err)
		}
	}

	// Setup SSH server
	server := &ssh.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handleSSHSession,
	}

	// Configure authentication based on config
	setupAuth(server, cfg)

	// Load host key
	if err := server.SetOption(ssh.HostKeyFile(hostKeyFile)); err != nil {
		logger.Fatal("Failed to load host key: %v", err)
	}

	logger.Info("Starting SSH notes server v%s on port %s", Version, cfg.Server.Port)
	logger.Info("Data directory: %s", cfg.Server.DataDir)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down server...")
		server.Close()
		logger.Close()
		os.Exit(0)
	}()

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server error: %v", err)
	}
}

func setupAuth(server *ssh.Server, cfg *config.Config) {
	switch cfg.Security.AuthMode {
	case "password":
		server.PasswordHandler = setupPasswordAuth(cfg)
	case "key":
		server.PublicKeyHandler = setupKeyAuth(cfg)
	case "both":
		server.PasswordHandler = setupPasswordAuth(cfg)
		server.PublicKeyHandler = setupKeyAuth(cfg)
	default: // "none"
		server.PasswordHandler = func(ctx ssh.Context, password string) bool { return true }
		server.PublicKeyHandler = func(ctx ssh.Context, key ssh.PublicKey) bool { return true }
	}
}

func setupPasswordAuth(cfg *config.Config) func(ssh.Context, string) bool {
	if !cfg.Security.RequirePassword {
		return func(ctx ssh.Context, password string) bool { return true }
	}
	
	// Load password file if specified
	auth := NewUserAuth()
	if cfg.Security.PasswordFile != "" {
		if err := auth.LoadUsers(cfg.Security.PasswordFile); err != nil {
			logger.Warn("Failed to load password file: %v", err)
		}
	}
	
	return func(ctx ssh.Context, password string) bool {
		username := ctx.User()
		if err := utils.ValidateUsername(username); err != nil {
			logger.Warn("Invalid username: %s", username)
			return false
		}
		return auth.VerifyPassword(username, password)
	}
}

func setupKeyAuth(cfg *config.Config) func(ssh.Context, ssh.PublicKey) bool {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		// In production, check against authorized_keys file
		// For now, accept all keys if no file specified
		if cfg.Security.AuthorizedKeysFile == "" {
			return true
		}
		// TODO: Implement authorized_keys checking
		return true
	}
}

func handleSSHSession(s ssh.Session) {
	defer utils.RecoverPanic()
	defer s.Close()
	
	startTime := time.Now()
	username := s.User()
	remoteAddr := s.RemoteAddr().String()
	
	if username == "" {
		username = "guest"
	}
	
	// Validate username
	if err := utils.ValidateUsername(username); err != nil {
		logger.Warn("Invalid username attempt: %s from %s", username, remoteAddr)
		fmt.Fprintf(s, "Error: Invalid username\r\n")
		return
	}
	
	logger.LogConnection(username, remoteAddr)
	defer func() {
		logger.LogDisconnection(username, time.Since(startTime))
	}()

	// Load config to get data directory
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		cfg = &config.DefaultConfig
	}
	
	userDataDir := filepath.Join(cfg.Server.DataDir, username)

	// Create user directory if it doesn't exist
	if err := os.MkdirAll(userDataDir, 0700); err != nil {
		logger.Error("Failed to create user directory for %s: %v", username, err)
		fmt.Fprintf(s, "Error: Failed to create user directory\r\n")
		return
	}

	// Request PTY for proper terminal handling
	pty, winCh, isPty := s.Pty()
	if !isPty {
		// PTY not allocated - send message and use defaults
		fmt.Fprintf(s, "PTY required. Please connect with: ssh -t -p %s %s@localhost\r\n", *port, username)
		fmt.Fprintf(s, "Using default terminal size...\r\n")
		width, height := 80, 24
		app := NewApp(username, userDataDir)
		if err := app.Run(s, width, height); err != nil {
			fmt.Fprintf(s, "Error: %v\r\n", err)
		}
		return
	}

	// Clear screen and move cursor to top, then hide cursor
	fmt.Fprintf(s, "\x1b[2J\x1b[H\x1b[?25l")

	// Initialize and run TUI
	app := NewApp(username, userDataDir)
	
	// Handle window resize
	go func() {
		for win := range winCh {
			if app.program != nil {
				app.program.Send(tea.WindowSizeMsg{
					Width:  win.Width,
					Height: win.Height,
				})
			}
		}
	}()
	
	// Run the app (it will send initial window size)
	if err := app.Run(s, pty.Window.Width, pty.Window.Height); err != nil {
		logger.Error("App error for user %s: %v", username, err)
		fmt.Fprintf(s, "Error: %v\r\n", err)
	}
}

// Authentication removed - using NoClientAuth for terminal.shop-like experience

