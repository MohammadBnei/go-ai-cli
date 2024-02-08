package config

import (
	"errors"
	"fmt"
	"sort"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/form"
	"github.com/MohammadBnei/go-ai-cli/ui/helper"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func NewConfigModel(promptConfig *service.PromptConfig) tea.Model {

	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
		viper.Set("default-systems", savedDefaultSystemPrompt)
	}

	items := getItemsAsUiList(promptConfig)

	delegateFn := getDelegateFn(promptConfig)

	return list.NewFancyListModel("config", items, delegateFn)

}

func getDelegateFn(promptConfig *service.PromptConfig) *list.DelegateFunctions {
	return &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			value := viper.Get(id)
			switch value := value.(type) {
			case string:
				editModel, err := getEditModel(id)
				if err != nil {
					return event.Error(err)
				}
				return event.AddStack(editModel, "Loading Editing "+id+"...")
			case bool:
				viper.Set(id, !value)
				err := viper.WriteConfig()
				if err != nil {
					return event.Error(err)
				}
				return func() tea.Msg {
					return list.Item{
						ItemId:          id,
						ItemTitle:       id,
						ItemDescription: helper.CheckedStringHelper(!value),
					}
				}
			default:
				return nil

			}
		},
		EditFn: func(id string) tea.Cmd {
			editModel, err := getEditModel(id)
			if err != nil {
				return event.Error(err)
			}

			return event.AddStack(editModel, "Loading Editing "+id+"...")
		},
	}
}

func getEditModel(id string) (tea.Model, error) {
	value := viper.Get(id)
	switch value := value.(type) {
	case string:
		var editModel *huh.Form
		var afterCmd tea.Cmd
		switch id {
		case "model":
			modelSelectForm, err := newModelSelectForm(value)
			if err != nil {
				return nil, err
			}
			editModel = modelSelectForm

		case "api_type":
			editModel = newApiTypeSelectForm(value)
			afterCmd = func() tea.Msg {
				modelSelectForm, err := newModelSelectForm(viper.GetString("model"))
				if err != nil {
					return err
				}
				return event.AddStackEvent{Stack: form.NewEditModel("Editing config model after updating the api type", modelSelectForm, func(form *huh.Form) tea.Cmd {
					result := form.GetString("model")
					return UpdateConfigValue("model", result, result)
				})}
			}

		case "configfile":
			afterCmd = func() tea.Msg {
				path := viper.GetString("configfile")
				viper.SetConfigFile(path)
				err := viper.ReadInConfig()
				if err != nil {
					if errors.Is(err, viper.ConfigFileNotFoundError{}) {
						return viper.WriteConfig()
					}
					return err
				}
				return nil
			}

		default:
			editModel = huh.NewForm(huh.NewGroup(
				huh.NewText().Title(id).Key(id).Value(&value).Lines(10)),
			)
		}
		return form.NewEditModel("Editing config ["+id+"]", editModel, func(form *huh.Form) tea.Cmd {
			result := form.GetString(id)
			return tea.Sequence(UpdateConfigValue(id, result, result), afterCmd)
		}), nil

	case bool:
		return form.NewEditModel("Editing config ["+id+"]", huh.NewForm(huh.NewGroup(
			huh.NewSelect[bool]().Key(id).Title(id).Options(huh.NewOptions[bool](true, false)...)),
		), func(form *huh.Form) tea.Cmd {
			result := form.GetBool(id)
			updateEvent := UpdateConfigValue(id, result, helper.CheckedStringHelper(result))
			if id == "md" {
				return tea.Sequence(updateEvent, event.UpdateChatContent("", ""))
			}
			return updateEvent
		},
		), nil

	default:
		return nil, fmt.Errorf("unknown type, : %T", value)
	}
}
func UpdateConfigValue(id string, value any, strValue string) tea.Cmd {
	viper.Set(id, value)
	err := viper.WriteConfig()
	if err != nil {
		return event.Error(err)
	}

	return func() tea.Msg {
		return list.Item{
			ItemId:          id,
			ItemTitle:       id,
			ItemDescription: strValue,
		}
	}
}

func getBoolItem(key string, value bool) list.Item {
	return list.Item{ItemId: key, ItemTitle: key, ItemDescription: helper.CheckedStringHelper(value)}
}

func getStringItem(key, value string) list.Item {
	return list.Item{ItemId: key, ItemTitle: key, ItemDescription: value}
}

func getItemsAsUiList(promptConfig *service.PromptConfig) []list.Item {
	items := []list.Item{}
	for k, v := range viper.AllSettings() {
		switch t := v.(type) {
		case string:
			switch t {
			default:
				items = append(items, getStringItem(k, t))
			}
		case bool:
			items = append(items, getBoolItem(k, t))

		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ItemId < items[j].ItemId
	})

	return items
}
