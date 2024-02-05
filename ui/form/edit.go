package form

import (
	"fmt"

	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type editModel struct {
	form      *huh.Form
	title     string
	submitted bool

	onSubmit func(form *huh.Form) tea.Cmd
}

func (m editModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cmds = append(cmds, event.RemoveStack(m))
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted && !m.submitted {
		m.submitted = true
		if m.onSubmit != nil {
			cmds = append(cmds, event.RemoveStack(m), m.onSubmit(m.form))
			return m, tea.Sequence(cmds...)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m editModel) View() string {
	return fmt.Sprintf("%s\n%s", style.TitleStyle.Render(m.title), m.form.View())
}

func NewEditModel(title string, form *huh.Form, onSubmit func(form *huh.Form) tea.Cmd) *editModel {
	m := editModel{form: form}
	m.onSubmit = onSubmit
	m.title = title
	return &m
}

func (m *editModel) WithExitOnSubmit() *editModel {
	m.onSubmit = func(form *huh.Form) tea.Cmd {
		return tea.Sequence(m.onSubmit(m.form), event.RemoveStack(m))
	}
	return m
}
