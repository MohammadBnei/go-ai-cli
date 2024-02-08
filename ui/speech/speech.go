package speech

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxDuration = 3 * time.Minute

type model struct {
	recording       bool
	viewport        viewport.Model
	textarea        textarea.Model
	keys            *listKeyMap
	title           string
	progress        progress.Model
	maxDuration     time.Duration
	currentDuration int
	timer           timer.Model
	recordCancelCtx func()
	aiCancelCtx     func()
	help            help.Model
	directReturn    bool
	lang            string
	langSelect      *huh.Form
}

func NewSpeechModel(promptConfig *service.PromptConfig, content string) tea.Model {
	ta := textarea.New()
	ta.Placeholder = "Recort a message"
	ta.SetValue(content)
	ta.Focus()
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	m := model{
		recording:       false,
		viewport:        viewport.New(10, 10),
		textarea:        ta,
		keys:            newKeyMap(),
		title:           "Speech mode",
		progress:        progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C")),
		maxDuration:     maxDuration,
		currentDuration: 0,
		help:            help.New(),
		lang:            "en",
	}
	m.langSelect = constructSelectLangForm(&m.lang)

	m.timer = timer.New(m.maxDuration)

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m, cmd = keyMapUpdate(msg, m)

	if cmd != nil {
		return m, cmd
	}

	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(m.GetTitleView()) - lipgloss.Height(m.help.View(m.keys))

		m.textarea.SetWidth(msg.Width)

	case tea.KeyMsg:
		if msg.String() == "up" || msg.String() == "down" {
			langSelect, cmd := m.langSelect.Update(msg)
			if form, ok := langSelect.(*huh.Form); ok {
				m.langSelect = form
				cmds = append(cmds, cmd, func() tea.Msg {
					return tea.KeyMsg{
						Type: tea.KeyCtrlN,
					}
				})
			}

			return m, tea.Batch(cmds...)
		}

	case appendText:
		m.textarea.SetValue(strings.TrimSpace(fmt.Sprintf("%s\n%s", m.textarea.Value(), msg.content)))
		if m.directReturn {
			return m, tea.Sequence(event.SetChatTextview(m.textarea.Value()), event.RemoveStack(m))
		}
		m.timer = timer.New(m.maxDuration)

	case setText:
		m.textarea.SetValue(msg.content)

	case startRecordingEvent:
		ctx, cancelFn := context.WithCancel(context.Background())
		m.recordCancelCtx = cancelFn

		aiCtx, cancelFn := context.WithCancel(context.Background())
		m.aiCancelCtx = cancelFn

		return m, tea.Batch(func() tea.Msg {
			res, err := SpeechToText(ctx, aiCtx, &SpeechConfig{Duration: m.maxDuration, Lang: m.lang})
			if err != nil {
				return err
			}

			return appendText{content: res}
		}, m.timer.Start())

	case stopRecordingEvent:
		if m.recordCancelCtx != nil {
			m.recordCancelCtx()
		}
		return m, m.timer.Stop()

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		return m, StopRecording

	case event.CancelEvent:
		if m.recordCancelCtx != nil {
			m.recordCancelCtx()
		}
		if m.aiCancelCtx != nil {
			m.aiCancelCtx()
		}

		return m, nil

	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	langSelect, cmd := m.langSelect.Update(msg)
	if form, ok := langSelect.(*huh.Form); ok {
		if lang := m.langSelect.GetString("lang"); lang != "" {
			m.lang = lang
			m.langSelect = constructSelectLangForm(&m.lang)
		} else {
			m.langSelect = form
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

var timerStyle = lipgloss.NewStyle().Padding(3)

func (m model) View() string {
	timerView := m.timer.View()
	if m.recording {
		timerView = timerStyle.Background(lipgloss.Color("#FF7CCB")).MarginLeft(3).Render(timerView)
	} else {
		timerView = timerStyle.Background(lipgloss.Color("#FDFF8C")).MarginLeft(3).Render(timerView)
	}
	m.viewport.SetContent(lipgloss.JoinHorizontal(lipgloss.Center, m.langSelect.View(), timerView))
	return fmt.Sprintf("%s\n%s\n%s\n%s", m.GetTitleView(), m.textarea.View(), m.viewport.View(), m.help.View(m.keys))
}

func (m model) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}

type appendText struct {
	content string
}
type setText struct {
	content string
}

type startRecordingEvent struct{}
type stopRecordingEvent struct{}

func StartRecording() tea.Msg {
	return startRecordingEvent{}
}
func StopRecording() tea.Msg {
	return stopRecordingEvent{}
}

func constructSelectLangForm(defaultLang *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("lang").Title("Language").Options(huh.NewOptions[string]("en", "fr", "es", "it", "fa", "ar")...).Value(defaultLang),
		),
	).WithShowHelp(false).
		WithShowErrors(false).
		WithKeyMap(&huh.KeyMap{
			Select: huh.SelectKeyMap{
				Submit: key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next")),
				Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
				Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
				Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
			},
		}).WithWidth(10)
}
