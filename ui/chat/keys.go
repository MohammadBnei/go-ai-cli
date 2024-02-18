package chat

import (
	"context"

	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/file"
	"github.com/MohammadBnei/go-ai-cli/ui/info"
	"github.com/MohammadBnei/go-ai-cli/ui/options"
	"github.com/MohammadBnei/go-ai-cli/ui/quit"
	"github.com/MohammadBnei/go-ai-cli/ui/speech"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type listKeyMap struct {
	copy                 key.Binding
	changeCurMessageUp   key.Binding
	changeCurMessageDown key.Binding
	quit                 key.Binding
	options              key.Binding
	cancel               key.Binding
	toggleHelpMenu       key.Binding
	showInfo             key.Binding
	speechToText         key.Binding
	textToSpeech         key.Binding
	addFile              key.Binding
	audiPlayer           key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		copy: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "copy"),
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

		options: key.NewBinding(
			key.WithKeys("ctrl+g", "esc"),
			key.WithHelp("ctrl+g, ", "options"),
		),

		quit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "quit"),
		),
		cancel: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel"),
		),

		showInfo: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "show info"),
		),

		speechToText: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "speech to text"),
		),
		textToSpeech: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "text to speech"),
		),
		addFile: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "add file(s)"),
		),

		audiPlayer: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "audio player"),
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
				return m, tea.Sequence(event.Cancel, event.RemoveStack(m.stack[len(m.stack)-1]))
			case msg.String() == "esc":
				return m, event.AddStack(options.NewOptionsModel(m.promptConfig), "Loading Options...")

			case m.help.ShowAll:
				m.help.ShowAll = false
				return m, func() tea.Msg { return m.size }

			case m.currentChatMessages.user != nil && m.promptConfig.FindContextWithId(m.currentChatMessages.user.Id.Int64()) != nil:
				return closeContext(m)

			}

		case key.Matches(msg, m.keys.quit):
			return m, event.AddStack(quit.NewQuitModel(m.promptConfig), "Quitting...")

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.help.ShowAll = !m.help.ShowAll
			return m, func() tea.Msg { return m.size }

		case key.Matches(msg, m.keys.copy):
			if len(m.stack) == 0 && m.currentChatMessages.assistant != nil {
				err := clipboard.WriteAll(m.currentChatMessages.assistant.Content)
				if err != nil {
					return m, event.Error(err)
				}
				return m, nil
			}

		case key.Matches(msg, m.keys.changeCurMessageUp):
			if len(m.stack) == 0 {
				return changeResponseUp(m)
			}
		case key.Matches(msg, m.keys.changeCurMessageDown):
			if len(m.stack) == 0 {
				return changeResponseDown(m)
			}

		case key.Matches(msg, m.keys.options):
			if len(m.stack) == 0 {
				return m, event.AddStack(options.NewOptionsModel(m.promptConfig), "Loading Options...")
			}

		case key.Matches(msg, m.keys.showInfo):
			if len(m.stack) == 0 {
				return m, event.AddStack(info.NewInfoModel("Info", getInfoContent(m)), "Loading info...")
			}

		case key.Matches(msg, m.keys.speechToText):
			if len(m.stack) == 0 {
				return m, event.AddStack(speech.NewSpeechModel(m.promptConfig, m.textarea.Value()), "Loading speech...")
			}

		case key.Matches(msg, m.keys.textToSpeech):
			if m.aiResponse != "" && m.currentChatMessages.assistant != nil && len(m.stack) == 0 {
				ctx, cancel := context.WithCancel(context.Background())
				m.promptConfig.AddContextWithId(ctx, cancel, m.currentChatMessages.assistant.Id.Int64())
				return m, tea.Sequence(func() tea.Msg {
					err := m.promptConfig.ChatMessages.FindById(m.currentChatMessages.assistant.Id.Int64()).FetchAudio(ctx)
					if err != nil {
						return err
					}
					return m.audioPlayer.InitSpeaker(m.currentChatMessages.assistant.Id.Int64())
				}, func() tea.Msg {
					m.promptConfig.DeleteContextById(m.currentChatMessages.assistant.Id.Int64())
					return event.Transition("")
				})
			}

		case key.Matches(msg, m.keys.addFile):
			if len(m.stack) == 0 {
				return m, event.AddStack(file.NewFilePicker(true), "Loading filepicker...")
			}

		case key.Matches(msg, m.keys.audiPlayer):
			if len(m.stack) == 0 {
				return m, event.AddStack(m.audioPlayer, "Loading audio player...")
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
		{k.changeCurMessageDown, k.changeCurMessageUp, k.copy},
		{k.cancel, k.quit, k.toggleHelpMenu},
		{k.showInfo, k.speechToText, k.textToSpeech},
		{k.addFile, k.options},
	}
}
