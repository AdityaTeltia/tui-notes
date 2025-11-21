package main

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
	"github.com/ssh-notes/terminal-notes/models"
)

type App struct {
	username  string
	dataDir   string
	model     tea.Model
	program   *tea.Program
	session   ssh.Session
}

func NewApp(username, dataDir string) *App {
	return &App{
		username: username,
		dataDir:  dataDir,
	}
}

func (a *App) Run(s ssh.Session, width, height int) error {
	a.session = s
	
	// Ensure minimum size
	if width < 40 {
		width = 80
	}
	if height < 10 {
		height = 24
	}
	
	// Initialize the main model with default size
	initialModel := models.NewMainModel(a.username, a.dataDir)
	
	// Pre-set the window size on the model
	initialModel.Update(tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	})
	
	// Create program with SSH session I/O
	opts := []tea.ProgramOption{
		tea.WithInput(s),
		tea.WithOutput(s),
	}
	
	a.program = tea.NewProgram(initialModel, opts...)
	
	// Run the program - this blocks until exit
	// Window size was already set on the model, so it should render immediately
	_, err := a.program.Run()
	return err
}

func (a *App) Quit() {
	if a.program != nil {
		a.program.Quit()
	}
}

