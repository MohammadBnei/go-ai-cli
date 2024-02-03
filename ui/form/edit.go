package form

import (
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type editModel struct {
	form *huh.Form

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

	if m.form.State == huh.StateCompleted {
		if m.onSubmit != nil {
			cmds = append(cmds, tea.Sequence(m.onSubmit(m.form), event.RemoveStack(m)))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m editModel) View() string {
	return m.form.View()
}

func NewEditModel(form *huh.Form, onSubmit func(form *huh.Form) tea.Cmd) *editModel {
	m := editModel{form: form}
	m.onSubmit = onSubmit
	return &m
}

func (m *editModel) WithExitOnSubmit() *editModel {
	m.onSubmit = func(form *huh.Form) tea.Cmd {
		return tea.Sequence(m.onSubmit(m.form), event.RemoveStack(m))
	}
	return m
}
