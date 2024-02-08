package file

import (
	"fmt"
	"os"

	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/filetree"
	"github.com/samber/lo"
)

type model struct {
	filetree      filetree.Model
	multiMode     bool
	selectedFiles []filetree.Item
	keys          *keyMap
	help          help.Model
}

// New creates a new instance of the UI.
func NewFilePicker(multipleMode bool) model {
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
		filetree:      filetreeModel,
		multiMode:     multipleMode,
		selectedFiles: []filetree.Item{},
		keys:          newKeyMap(),
		help:          help.New(),
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
		m.filetree.SetSize(msg.Width, msg.Height-lipgloss.Height(m.help.View(m.keys)))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.selectFile):
			if _, ok := lo.Find(m.selectedFiles, func(item filetree.Item) bool {
				return item.FileName() == m.filetree.GetSelectedItem().FileName()
			}); ok {
				m.selectedFiles = lo.Filter(m.selectedFiles, func(item filetree.Item, _ int) bool {
					return item.FileName() != m.filetree.GetSelectedItem().FileName()
				})
				return m, nil
			}
			m.selectedFiles = append(m.selectedFiles, m.filetree.GetSelectedItem())
		case key.Matches(msg, m.keys.submit):
			return m, tea.Sequence(event.RemoveStack(m), event.FileSelection(m.selectedFiles, m.multiMode))
		}
	}

	m.filetree, cmd = m.filetree.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.filetree.View(), m.help.View(m.keys))
}
