package info

import (
	"fmt"

	"github.com/MohammadBnei/go-openai-cli/ui/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	title, content string
	viewport       viewport.Model
}

func NewInfoModel(title, content string) tea.Model {
	return model{
		title:   style.TitleStyle.Render(title),
		content: content,
		viewport: viewport.Model{
			Width:  80,
			Height: 10,
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - lipgloss.Height(m.title)

	}

	m.viewport.SetContent(m.content)
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.title, m.viewport.View())
}
