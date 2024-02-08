package speech

import (
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type listKeyMap struct {
	toggleRecord key.Binding
	end          key.Binding
	clear        key.Binding
}

func newKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleRecord: key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "toggle record")),
		end:          key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		clear:        key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+c", "clear recording")),
	}
}

func keyMapUpdate(msg tea.Msg, m model) (model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.keys.toggleRecord):
			m.recording = !m.recording
			if m.recording {
				return m, StartRecording
			} else {
				return m, StopRecording
			}

		case key.Matches(msg, m.keys.end):
			if m.recording {
				m.directReturn = true
				return m, StopRecording
			} else {
				return m, tea.Sequence(event.SetChatTextview(m.textarea.Value()), event.RemoveStack(m))
			}

		case key.Matches(msg, m.keys.clear):
			m.textarea.Reset()
		}
	}

	return m, nil
}

func (k *listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.toggleRecord, k.clear, k.end}
}

func (k *listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
