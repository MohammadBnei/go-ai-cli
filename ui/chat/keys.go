package chat

import (
	"github.com/MohammadBnei/go-openai-cli/ui/config"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/info"
	"github.com/MohammadBnei/go-openai-cli/ui/message"
	"github.com/MohammadBnei/go-openai-cli/ui/system"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type listKeyMap struct {
	systemMessages       key.Binding
	curMessages          key.Binding
	globalConfig         key.Binding
	pager                key.Binding
	changeCurMessageUp   key.Binding
	changeCurMessageDown key.Binding
	quit                 key.Binding
	cancel               key.Binding
	toggleHelpMenu       key.Binding
	showInfo             key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		systemMessages: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "system messages"),
		),
		curMessages: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "current messages"),
		),
		globalConfig: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "global config"),
		),
		pager: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "pager"),
		),
		changeCurMessageUp: key.NewBinding(
			key.WithKeys("ctrl+j"),
			key.WithHelp("ctrl+j", "previous message"),
		),
		changeCurMessageDown: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "next message"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "toggle help"),
		),

		quit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "quit"),
		),
		cancel: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c/esc", "cancel"),
		),

		showInfo: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "show info"),
		),
	}
}

func keyMapUpdate(msg tea.Msg, m chatModel) (chatModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.cancel):
			switch {
			case m.err != nil:
				m.err = nil
				return m, nil

			case len(m.stack) > 0:
				return m, event.RemoveStack(m.stack[len(m.stack)-1])

			case m.help.ShowAll:
				m.help.ShowAll = false
				return m, func() tea.Msg { return m.size }

			case m.promptConfig.FindContextWithId(m.currentChatIndices.user) != nil && msg.String() == "ctrl+c":
				return closeContext(m)

			}

		case key.Matches(msg, m.keys.quit):
			return quit(m)

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.help.ShowAll = !m.help.ShowAll
			return m, func() tea.Msg { return m.size }

		case key.Matches(msg, m.keys.systemMessages):
			if len(m.stack) == 0 {
				return m, event.AddStack(system.NewSystemModel(m.promptConfig))
			}

		case key.Matches(msg, m.keys.globalConfig):
			if len(m.stack) == 0 {
				return m, event.AddStack(config.NewConfigModel(m.promptConfig))
			}

		case key.Matches(msg, m.keys.curMessages):
			if len(m.stack) == 0 {
				return m, event.AddStack(message.NewMessageModel(m.promptConfig))
			}

		case key.Matches(msg, m.keys.changeCurMessageUp):
			return changeResponseUp(m)

		case key.Matches(msg, m.keys.changeCurMessageDown):
			return changeResponseDown(m)

		case key.Matches(msg, m.keys.pager):
			if len(m.stack) == 0 {
				return addPagerToStack(m)
			}

		case key.Matches(msg, m.keys.showInfo):
			if len(m.stack) == 0 {
				return m, event.AddStack(info.NewInfoModel("Info", getInfoContent(m)))
			}

		}
	}

	return m, nil
}

func (k *listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.cancel, k.quit, k.toggleHelpMenu}
}

func (k *listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.systemMessages, k.curMessages, k.globalConfig},
		{k.changeCurMessageDown, k.changeCurMessageUp, k.pager},
		{k.cancel, k.quit, k.toggleHelpMenu},
		{k.showInfo},
	}
}
