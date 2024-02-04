package file

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/filetree"
)

type model struct {
	filetree filetree.Model
}

// New creates a new instance of the UI.
func NewFilePicker() model {
	startDir, _ := os.Getwd()
	filetreeModel := filetree.New(
		true,
		true,
		startDir,
		"",
		lipgloss.AdaptiveColor{Light: "#000000", Dark: "63"},
		lipgloss.AdaptiveColor{Light: "#000000", Dark: "63"},
		lipgloss.AdaptiveColor{Light: "63", Dark: "63"},
		lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
	)

	return model{
		filetree: filetreeModel,
	}
}

// Init intializes the UI.
func (m model) Init() tea.Cmd {
	return m.filetree.Init()
}

// Update handles all UI interactions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.filetree.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		}
	}

	m.filetree.GetSelectedItem()

	m.filetree, cmd = m.filetree.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m model) View() string {
	return m.filetree.View()
}
