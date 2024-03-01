package options

import (
	"errors"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/config"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

type configModel struct {
	list list.Model

	title string
}

const (
	SAVE_CONFIG = "save"
	SETTINGS    = "settings"
)

func NewConfigOptionsModel(pc *service.PromptConfig) tea.Model {
	items := getConfOItemsAsUiList(pc)

	return list.NewFancyListModel("Options > Config", items, &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			switch id {
			case SAVE_CONFIG:
				return tea.Sequence(event.RemoveStack(nil), event.Error(viper.WriteConfig()))
			case SETTINGS:
				return event.AddStack(config.NewConfigModel(pc), "Loading Config...")
			case CLEAR:
				pc.ChatMessages.ClearMessages()
				return event.RemoveStack(list.Model{})
			}

			return event.Error(errors.New("unknown option: " + id))

		},
	})

}

func getConfOItemsAsUiList(pc *service.PromptConfig) []list.Item {
	return []list.Item{
		{ItemId: SETTINGS, ItemTitle: "Settings"},
		{ItemId: SAVE_CONFIG, ItemTitle: "Save Config"},
	}
}
