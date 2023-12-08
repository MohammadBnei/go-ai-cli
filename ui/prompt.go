package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func BasicPrompt(label, previousPrompt string) (string, error) {
	p := tea.NewProgram(initialModel(label, "Ask AI..."))
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	err = p.ReleaseTerminal()
	if err != nil {
		return "", err
	}

	return m.(model).Message, nil
}

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	Message   string
	label     string

	err error
}

func initialModel(label, placeholder string) model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()
	ti.Prompt = ": "

	return model{
		textInput: ti,
		label:     label,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.Message = "\\quit"
			return m, tea.Quit
		case tea.KeyEnter:
			m.Message = m.textInput.Value()
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, tea.Batch(cmd)
}

func (m model) View() string {
	return fmt.Sprint(
		m.label,
		m.textInput.View(),
	)
}
