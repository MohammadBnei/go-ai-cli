package audio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type AudioPlayerModel struct {
	promptConfig *service.PromptConfig
	chatMsgId    int64

	keys *keyMap

	title string

	streamer beep.StreamSeeker
	format   *beep.Format
	ctrl     *beep.Ctrl
	speed    *beep.Resampler

	help help.Model

	viewport viewport.Model
}

func NewPlayerModel(pc *service.PromptConfig) *AudioPlayerModel {

	return &AudioPlayerModel{
		title:        "Audio controller",
		chatMsgId:    -1,
		promptConfig: pc,
		keys:         newKeyMap(),
		help:         help.New(),
		viewport:     viewport.New(0, 0),
	}
}

func (m AudioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - lipgloss.Height(m.help.View(m.keys)) - lipgloss.Height(m.GetTitleView())

	case Tick:
		cmds = append(cmds, tea.Tick(time.Second, func(time.Time) tea.Msg { return Tick{} }))

	case tea.KeyMsg:
		if m.chatMsgId == -1 {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.play):
			speaker.Lock()
			m.ctrl.Paused = !m.ctrl.Paused
			speaker.Unlock()

		case key.Matches(msg, m.keys.speedUp):
			newRatio := m.speed.Ratio() + 0.2
			if newRatio <= 2 {
				speaker.Lock()
				m.speed.SetRatio(newRatio)
				speaker.Unlock()
			}

		case key.Matches(msg, m.keys.speedDown):
			newRatio := m.speed.Ratio() - 0.2
			if newRatio > 0.2 {
				speaker.Lock()
				m.speed.SetRatio(newRatio)
				speaker.Unlock()
			}

		case key.Matches(msg, m.keys.forward):
			position := m.format.SampleRate.D(m.streamer.Position())
			position += 5 * time.Second
			speaker.Lock()
			err := m.streamer.Seek(m.format.SampleRate.N(position))
			if err != nil {
				cmds = append(cmds, event.Error(err))
			}
			speaker.Unlock()

		case key.Matches(msg, m.keys.back):
			position := m.format.SampleRate.D(m.streamer.Position())
			position -= 5 * time.Second
			if position >= 0 {
				speaker.Lock()
				err := m.streamer.Seek(m.format.SampleRate.N(position))
				if err != nil {
					cmds = append(cmds, event.Error(err))
				}
				speaker.Unlock()
			}

		}
	}

	if m.chatMsgId != -1 {
		speaker.Lock()
		position := m.format.SampleRate.D(m.streamer.Position())
		length := m.format.SampleRate.D(m.streamer.Len())
		speed := m.speed.Ratio()
		speaker.Unlock()

		m.viewport.SetContent(fmt.Sprintf("%v / %v\n%.1fx", position.Round(time.Second), length.Round(time.Second), speed))
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AudioPlayerModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return Tick{} })
}

func (m AudioPlayerModel) View() string {
	title := m.GetTitleView()
	if m.chatMsgId == -1 {
		return fmt.Sprintf("%s\nNo message selected", title)
	}

	return fmt.Sprintf("%s\n%s\n%s", title, m.viewport.View(), m.help.View(m.keys))
}

func (m AudioPlayerModel) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}

func (m *AudioPlayerModel) Clear() {
	m.chatMsgId = -1
	m.ctrl = nil
	m.streamer = nil
	m.format = nil
	m.speed = nil
}

func (m *AudioPlayerModel) InitSpeaker(chatMsgId int64) any {
	m.chatMsgId = chatMsgId
	msg := m.promptConfig.ChatMessages.FindById(chatMsgId)
	if msg == nil {
		return errors.New("message not found")
	}
	if msg.Audio == nil {
		ctx, cancelFn := service.LoadContext(context.Background())
		m.promptConfig.AddContextWithId(ctx, cancelFn, chatMsgId)
		defer m.promptConfig.DeleteContext(ctx)
		err := msg.FetchAudio(ctx)
		if err != nil {
			return err
		}
	}
	streamer, format, err := mp3.Decode(msg.Audio)
	if err != nil {
		return err
	}

	if err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/30)); err != nil {
		return err
	}

	m.format = &format
	m.streamer = streamer

	m.ctrl = &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	m.speed = beep.ResampleRatio(4, 1, m.ctrl)

	speaker.Play(m.speed)

	return StartPlayingEvent{}
}

type StartPlayingEvent struct{}

type Tick struct{}
