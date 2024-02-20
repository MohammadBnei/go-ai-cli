//go:build portaudio

package audio

import (
	"fmt"
	"io"
	"os"
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

	keys *keyMap

	title string

	streamer beep.StreamSeeker
	format   *beep.Format
	ctrl     *beep.Ctrl
	speed    *beep.Resampler

	help help.Model

	viewport viewport.Model

	audio io.ReadCloser
}

func NewPlayerModel(pc *service.PromptConfig) *AudioPlayerModel {

	return &AudioPlayerModel{
		title:        "Audio controller",
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
		if m.streamer != nil {
			cmds = append(cmds, m.DoTick())
		}

	case tea.KeyMsg:
		if m.streamer == nil {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.play):
			speaker.Lock()
			if m.streamer.Position() == m.streamer.Len() {
				speaker.Unlock()
				speaker.Clear()
				cmd := m.InitSpeaker(m.audio)
				return m, cmd
			} else {
				m.ctrl.Paused = !m.ctrl.Paused
			}
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
			speaker.Lock()
			if position >= 0 {
				err := m.streamer.Seek(m.format.SampleRate.N(position))
				if err != nil {
					cmds = append(cmds, event.Error(err))
				}
			} else {
				err := m.streamer.Seek(0)
				if err != nil {
					cmds = append(cmds, event.Error(err))
				}
			}
			speaker.Unlock()

		}
	}

	if m.streamer != nil {
		speaker.Lock()
		position := m.format.SampleRate.D(m.streamer.Position())
		length := m.format.SampleRate.D(m.streamer.Len())
		speed := m.speed.Ratio()
		speaker.Unlock()

		m.viewport.SetContent(fmt.Sprintf("%v / %v\n%.1fx", position.Round(time.Second), length.Round(time.Second), speed))
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd, func() tea.Msg {
			return StartPlayingEvent{&m}
		})

	}

	return m, tea.Batch(cmds...)
}

func (m AudioPlayerModel) Init() tea.Cmd {
	return nil
}

func (m AudioPlayerModel) DoTick() tea.Cmd {
	return tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
		return Tick{}
	})
}

func (m AudioPlayerModel) View() string {
	title := m.GetTitleView()
	if m.streamer == nil {
		return fmt.Sprintf("%s\nNo audio selected", title)
	}

	return fmt.Sprintf("%s\n%s\n%s", title, m.viewport.View(), m.help.View(m.keys))
}

func (m AudioPlayerModel) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}

func (m *AudioPlayerModel) Clear() {
	speaker.Clear()
	m.ctrl = nil
	m.streamer = nil
	m.format = nil
	m.speed = nil
}

func (m *AudioPlayerModel) InitSpeaker(audio io.ReadCloser) tea.Cmd {
	data, err := io.ReadAll(audio)
	if err != nil {
		return event.Error(err)
	}

	// TODO : switch to virtual fs
	err = os.WriteFile("audio.mp3", data, 0666)
	if err != nil {
		return event.Error(err)
	}
	defer os.Remove("audio.mp3")

	audioFile, err := os.Open("audio.mp3")
	if err != nil {
		return event.Error(err)
	}
	m.audio, _ = os.Open("audio.mp3")
	streamer, format, err := mp3.Decode(audioFile)
	if err != nil {
		return event.Error(err)
	}
	m.streamer = streamer

	if err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/30)); err != nil {
		return event.Error(err)
	}

	m.format = &format

	m.ctrl = &beep.Ctrl{Streamer: m.streamer, Paused: false}
	m.speed = beep.ResampleRatio(4, 1, m.ctrl)

	speaker.Play(m.speed)

	return func() tea.Msg {
		return StartPlayingEvent{
			PlayerModel: m,
		}
	}
}

type StartPlayingEvent struct {
	PlayerModel *AudioPlayerModel
}

type Tick struct{}
