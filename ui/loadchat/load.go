package loadchat

import (
	"fmt"
	"path/filepath"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type model struct {
	filepicker filepicker.Model
	help       help.Model
	title      string
	width      int
	services   *service.Services
}

// New creates a new instance of the UI.
func NewLoadChatModel(pc *service.Services) model {
	fp := filepicker.New()
	fp.CurrentDirectory = filepath.Dir(viper.GetViper().ConfigFileUsed())
	fp.ShowHidden = true
	fp.AutoHeight = true
	fp.AllowedTypes = []string{"yml", "yaml"}

	return model{
		filepicker: fp,
		services:   pc,
		help:       help.New(),
		title:      "Chat Loader",
	}
}

// Init intializes the UI.
func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update handles all UI interactions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.filepicker, cmd = m.filepicker.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.filepicker.Height = msg.Height - lipgloss.Height(m.GetTitleView())
		m.width = msg.Width
	case tea.KeyMsg:
	}

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		if err := m.services.ChatMessages.LoadFromFile(path); err != nil {
			return m, event.Error(err)
		} else {
			return m, event.RemoveStack(m)
		}
	}

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.GetTitleView(), m.filepicker.View())
}

func (m model) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}
