//go:build !portaudio

package audio

import (
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/transition"
	tea "github.com/charmbracelet/bubbletea"
)

type AudioPlayerModel struct {
	transitionModel *transition.Model
}

func NewPlayerModel(pc *service.PromptConfig) *AudioPlayerModel {

	return &AudioPlayerModel{
		transitionModel: transition.NewTransitionModel("Portaudio not found"),
	}
}

func (m AudioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m AudioPlayerModel) Init() tea.Cmd {
	return nil
}

func (m AudioPlayerModel) View() string {
	return m.transitionModel.View()
}

func (m *AudioPlayerModel) Clear() {

}

func (m *AudioPlayerModel) InitSpeaker(chatMsgId int64) any {
	return nil
}

type StartPlayingEvent struct{}

type Tick struct{}
