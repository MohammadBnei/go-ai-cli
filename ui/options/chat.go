package options

import (
	"errors"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/MohammadBnei/go-ai-cli/ui/loadchat"
	"github.com/MohammadBnei/go-ai-cli/ui/savechat"
	tea "github.com/charmbracelet/bubbletea"
)

type chatModel struct {
	list list.Model

	title string
}

const (
	SAVE  = "save"
	LOAD  = "load"
	CLEAR = "clear"
)

func NewChatOptionsModel(pc *service.PromptConfig) tea.Model {
	items := getCOItemsAsUiList(pc)

	return list.NewFancyListModel("Options > Chat", items, &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			switch id {
			case SAVE:
				return event.AddStack(savechat.NewSaveChatModel(pc), "Loading Save chat...")
			case LOAD:
				return event.AddStack(loadchat.NewLoadChatModel(pc), "Loading Load chat...")
			case CLEAR:
				pc.ChatMessages.ClearMessages()
				return event.RemoveStack(list.Model{})
			}

			return event.Error(errors.New("unknown option: " + id))

		},
	})

}

func getCOItemsAsUiList(pc *service.PromptConfig) []list.Item {
	return []list.Item{
		{ItemId: SAVE, ItemTitle: "Save"},
		{ItemId: LOAD, ItemTitle: "Load"},
		{ItemId: CLEAR, ItemTitle: "Clear"},
	}
}
