//go:build portaudio

package audio

import (
	"fmt"
	"time"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
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

	selectMode bool

	fileList *list.Model

	help help.Model

	viewport viewport.Model

	fileId string

	ticking bool

	size tea.WindowSizeMsg
}

func NewPlayerModel(pc *service.PromptConfig) (*AudioPlayerModel, error) {
	m := &AudioPlayerModel{
		title:        "Audio controller",
		promptConfig: pc,
		keys:         newKeyMap(),
		help:         help.New(),
		viewport:     viewport.New(0, 0),
		fileList:     list.NewFancyListModel("Audio files", []list.Item{}, getDelegateFn(pc)),
	}

	m.RefreshFileList()

	return m, nil
}

func (m AudioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	m.ticking = m.streamer != nil && m.streamer.Position() != m.streamer.Len() && !m.ctrl.Paused && m.fileId != ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - lipgloss.Height(m.help.View(m.keys)) - lipgloss.Height(m.GetTitleView())
		m.fileList.List.SetHeight(msg.Height)
		m.fileList.List.SetWidth(msg.Width)

	case Tick:
		if m.ticking {
			cmds = append(cmds, m.DoTick())
		}

	case SelectAudioFileEvent:
		cmds = append(cmds, m.RefreshFileList())
		if !m.ticking {
			cmds = append(cmds, m.DoTick())
			m.ticking = true
		}
		return m, m.InitSpeaker(msg.Id)

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.toggleSelect) {
			cmds = append(cmds, m.RefreshFileList())
			m.selectMode = !m.selectMode
		}
		if m.streamer != nil && m.fileId != "" {
			switch {
			case key.Matches(msg, m.keys.play):
				speaker.Lock()
				if m.streamer.Position() == m.streamer.Len() {
					speaker.Unlock()
					speaker.Clear()
					cmd := m.InitSpeaker(m.fileId)
					return m, cmd
				} else {
					m.ctrl.Paused = !m.ctrl.Paused
					if !m.ticking {
						cmds = append(cmds, m.DoTick())
						m.ticking = true
					}
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
	}

	fileList, cmd := m.fileList.Update(msg)
	if fl, ok := fileList.(list.Model); ok {
		m.fileList = &fl
		cmds = append(cmds, cmd)
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
		cmds = append(cmds, cmd)

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
	if m.selectMode || m.fileId == "" {
		return m.fileList.View()
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

func (m *AudioPlayerModel) InitSpeaker(id string) tea.Cmd {
	file, _, err := m.promptConfig.FileService.Get(id)
	if err != nil {
		return event.Error(err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return event.Error(err)
	}
	m.streamer = streamer

	if err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/30)); err != nil {
		return event.Error(err)
	}

	m.fileId = id

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

func (m *AudioPlayerModel) RefreshFileList() tea.Cmd {
	files, err := m.promptConfig.FileService.List(service.Audio)
	if err != nil {
		return event.Error(err)
	}
	items := getFilesAsItem(files, m.promptConfig)
	return m.fileList.List.SetItems(items)
}

type StartPlayingEvent struct {
	PlayerModel *AudioPlayerModel
}

type Tick struct{}
