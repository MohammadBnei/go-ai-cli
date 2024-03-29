package speech

import (
	"errors"

	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type listKeyMap struct {
	toggleRecord key.Binding
	end          key.Binding
	clear        key.Binding
	copy         key.Binding
}

func newKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleRecord: key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "toggle record")),
		end:          key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		copy:         key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "copy to clipboard")),
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
			if m.recording || m.aiCancelCtx != nil {
				m.directReturn = true
				return m, StopRecording
			}
			return m, tea.Sequence(event.SetChatTextview(m.textarea.Value()), event.RemoveStack(m))

		case key.Matches(msg, m.keys.clear):
			m.textarea.Reset()

		case key.Matches(msg, m.keys.copy):
			if clipboard.Unsupported {
				return m, event.Error(errors.New("clipboard is not supported on this platform"))
			}
			if m.textarea.Value() != "" {
				err := clipboard.WriteAll(m.textarea.Value())
				if err != nil {
					return m, event.Error(err)
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (k *listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.toggleRecord, k.clear, k.copy, k.end}
}

func (k *listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
