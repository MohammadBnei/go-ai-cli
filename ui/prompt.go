package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func StringPrompt(label string) (string, error) {
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
	textInput textarea.Model
	Message   string
	label     string

	err error
}

func initialModel(label, placeholder string) model {
	ti := textarea.New()
	ti.Placeholder = placeholder
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()
	ti.Prompt = ": "
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		width = 80
	}
	ti.SetWidth(width)
	ti.SetHeight(1)

	return model{
		textInput: ti,
		label:     label,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
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
