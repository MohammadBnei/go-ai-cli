package transition

import (
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Title string
}

func NewTransitionModel(title string) *Model {
	return &Model{Title: title}
}

func (m Model) Init() tea.Cmd {
	return nil
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
func (m Model) View() string {
	return m.GetTitleView()
}

func (m Model) GetTitleView() string {
	return style.TitleStyle.Render(m.Title)
}
