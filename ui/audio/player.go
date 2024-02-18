package audio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type AudioPlayerModel struct {
	promptConfig *service.PromptConfig
	chatMsgId    int64
	playing      bool

	keys *keyMap

	title string

	streamer beep.StreamSeeker
	format   *beep.Format
	ctrl     *beep.Ctrl
	speed    *beep.Resampler
}

func NewPlayerModel(pc *service.PromptConfig) *AudioPlayerModel {

	return &AudioPlayerModel{
		title:        "Audio controller",
		chatMsgId:    -1,
		playing:      false,
		promptConfig: pc,
		keys:         newKeyMap(),
	}
}

func (m AudioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

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
			speaker.Lock()
			m.speed.SetRatio(m.speed.Ratio() + 0.2)
			speaker.Unlock()

		case key.Matches(msg, m.keys.speedDown):
			speaker.Lock()
			m.speed.SetRatio(m.speed.Ratio() - 0.2)
			speaker.Unlock()

		}
	}
	return m, nil
}

func (m AudioPlayerModel) Init() tea.Cmd {
	return nil
}

func (m AudioPlayerModel) View() string {
	title := m.GetTitleView()
	if m.chatMsgId == -1 {
		return fmt.Sprintf("%\nsNo message selected", title)
	}
	speaker.Lock()
	position := m.format.SampleRate.D(m.streamer.Position())
	length := m.format.SampleRate.D(m.streamer.Len())
	speed := m.speed.Ratio()
	speaker.Unlock()
	return fmt.Sprintf("%s\n%v / %v\n%.3fx", title, position.Round(time.Second), length.Round(time.Second), speed)
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

func (m *AudioPlayerModel) InitSpeaker(chatMsgId int64) error {
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

	if err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60)); err != nil {
		return err
	}

	m.format = &format
	m.streamer = streamer

	m.ctrl = &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	m.speed = beep.ResampleRatio(4, 1, m.ctrl)

	return nil
}
