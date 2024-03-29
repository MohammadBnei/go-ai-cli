package config

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/MohammadBnei/go-ai-cli/config"
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

	savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
		viper.Set(config.PR_SYSTEM_DEFAULT, savedDefaultSystemPrompt)
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
		case config.AI_MODEL_NAME:
			modelSelectForm, err := newModelSelectForm(value)
			if err != nil {
				return nil, err
			}
			editModel = modelSelectForm

		case config.AI_OPENAI_IMAGE_MODEL:
			modelSelectForm, err := newImageModelSelectForm(value)
			if err != nil {
				return nil, err
			}
			editModel = modelSelectForm

		case config.AI_API_TYPE:
			editModel = newApiTypeSelectForm(value)
			afterCmd = func() tea.Msg {
				modelSelectForm, err := newModelSelectForm(viper.GetString(config.AI_MODEL_NAME))
				if err != nil {
					return err
				}
				return event.AddStackEvent{Stack: form.NewEditModel("Editing config model after updating the api type", modelSelectForm, func(form *huh.Form) tea.Cmd {
					result := form.GetString(config.AI_MODEL_NAME)
					return UpdateConfigValue(config.AI_MODEL_NAME, result, result)
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
				huh.NewText().Editor("nvim").CharLimit(0).Title(id).Key(id).Value(&value).Lines(10)),
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
			if id == config.UI_MARKDOWN_MODE {
				return tea.Sequence(updateEvent, event.UpdateChatContent("", ""))
			}
			return updateEvent
		},
		), nil

	case int:
		strVal := fmt.Sprintf("%d", value)
		return form.NewEditModel("Editing config ["+id+"]", huh.NewForm(huh.NewGroup(
			huh.NewInput().Key(id).Title(id).
				Validate(func(s string) error { _, err := strconv.Atoi(s); return err }).
				Value(&strVal).CharLimit(3),
		),
		), func(form *huh.Form) tea.Cmd {
			strValue := form.GetString(id)
			result, err := strconv.Atoi(strValue)
			if err != nil {
				return event.Error(err)
			}
			return UpdateConfigValue(id, result, strValue)
		},
		), nil

	case float64:
		strVal := fmt.Sprintf("%.2f", value)
		return form.NewEditModel("Editing config ["+id+"]", huh.NewForm(huh.NewGroup(
			huh.NewInput().Key(id).Title(id).
				Validate(func(s string) error { _, err := strconv.ParseFloat(s, 64); return err }).
				Value(&strVal).CharLimit(3),
		),
		), func(form *huh.Form) tea.Cmd {
			strValue := form.GetString(id)
			result, err := strconv.ParseFloat(strValue, 64)
			if err != nil {
				return event.Error(err)
			}
			return UpdateConfigValue(id, result, strValue)
		},
		), nil

	default:
		return nil, fmt.Errorf("unknown type, : %T", value)
	}
}
func UpdateConfigValue(id string, value any, strValue string) tea.Cmd {
	viper.Set(id, value)

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
func getIntItem(key string, value int) list.Item {
	return list.Item{ItemId: key, ItemTitle: key, ItemDescription: fmt.Sprintf("%d", value)}
}
func getFloatItem(key string, value float64) list.Item {
	return list.Item{ItemId: key, ItemTitle: key, ItemDescription: fmt.Sprintf("%.2f", value)}
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
		case int:
			items = append(items, getIntItem(k, t))
		case float64:
			items = append(items, getFloatItem(k, t))

		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ItemId < items[j].ItemId
	})

	return items
}
