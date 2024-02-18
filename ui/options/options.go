package options

import (
	"errors"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/MohammadBnei/go-ai-cli/ui/message"
	"github.com/MohammadBnei/go-ai-cli/ui/quit"
	"github.com/MohammadBnei/go-ai-cli/ui/system"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	list list.Model

	title string
}

const (
	CONFIG         = "config"
	MESSAGES       = "messages"
	SYSTEM_PROMPTS = "system_prompts"
	CHAT           = "chat"
	ERRORS         = "errors"
	REFRESH        = "refresh"
	EXIT           = "exit"
)

func NewOptionsModel(pc *service.PromptConfig) tea.Model {
	items := getItemsAsUiList(pc)

	return list.NewFancyListModel("Options", items, &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			switch id {
			case CONFIG:
				return event.AddStack(NewConfigOptionsModel(pc), "Loading Config...")
			case MESSAGES:
				return event.AddStack(message.NewMessageModel(pc), "Loading Messages...")
			case SYSTEM_PROMPTS:
				return event.AddStack(system.NewSystemModel(pc), "Loading System Prompts...")
			case CHAT:
				return event.AddStack(NewChatOptionsModel(pc), "Loading Chat...")
			case EXIT:
				return event.AddStack(quit.NewQuitModel(pc), "Quitting...")
			}
			return event.Error(errors.New("unknown option: " + id))
		},
	})

}

func getItemsAsUiList(pc *service.PromptConfig) []list.Item {
	return []list.Item{
		{ItemId: CONFIG, ItemTitle: "Config"},
		{ItemId: MESSAGES, ItemTitle: "Messages"},
		{ItemId: SYSTEM_PROMPTS, ItemTitle: "System Prompts"},
		{ItemId: CHAT, ItemTitle: "Chat"},
		{ItemId: ERRORS, ItemTitle: "Errors"},
		{ItemId: REFRESH, ItemTitle: "Refresh"},
		{ItemId: EXIT, ItemTitle: "Exit"},
	}
}
