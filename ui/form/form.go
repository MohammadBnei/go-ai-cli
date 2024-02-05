package form

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(0, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.Copy().
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type Model struct {
	lg     *lipgloss.Renderer
	styles *Styles
	form   *huh.Form
	width  int

	title string

	onSubmit func(form *huh.Form) tea.Cmd
}

func NewModel(form *huh.Form, title string, onComplete func(form *huh.Form) tea.Cmd) Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	m.form = form.WithShowErrors(false).WithShowHelp(false)
	m.title = title

	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - m.styles.Base.GetHorizontalFrameSize()
		m.styles.Base = m.styles.Base.Height(msg.Height).Width(msg.Width)
		m.form.WithWidth(m.width)
		m.form.WithHeight(msg.Height - m.styles.Base.GetVerticalFrameSize() - 2*lipgloss.Height(m.appBoundaryView("t")))

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+enter":
			m.form.State = huh.StateCompleted
		}
	}

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		if m.onSubmit != nil {
			cmds = append(cmds, m.onSubmit(m.form))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) GetBody() string {
	v := strings.TrimSuffix(m.form.View(), "\n\n")

	return m.lg.NewStyle().Render(v)
}

func (m Model) View() string {
	s := m.styles

	switch m.form.State {
	case huh.StateCompleted:
		var b strings.Builder
		fmt.Fprintf(&b, "Your Email is:\n\n%s\n\n", m.form.GetString("email"))
		return s.Status.Copy().Margin(0, 1).Padding(1, 2).Width(m.width).Render(b.String()) + "\n\n"
	default:

		errors := m.form.Errors()
		header := m.appBoundaryView(m.title)
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		body := m.GetBody()

		return s.Base.Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
	}
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
