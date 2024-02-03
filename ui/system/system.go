package system

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/form"
	"github.com/MohammadBnei/go-openai-cli/ui/helper"
	uiList "github.com/MohammadBnei/go-openai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-module/carbon/v2"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tigergraph/promptui"
)

func NewSystemModel(promptConfig *service.PromptConfig) tea.Model {
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
		viper.Set("default-systems", savedDefaultSystemPrompt)
	}

	items := getItemsAsUiList(promptConfig)

	delegateFn := getDelegateFn(promptConfig)

	return uiList.NewFancyListModel("system", items, delegateFn)
}

func getItemsAsUiList(promptConfig *service.PromptConfig) []uiList.Item {
	savedSystemPrompt := viper.GetStringMapString("systems")
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")

	res := lo.MapToSlice[string, string, uiList.Item](savedSystemPrompt, func(k string, v string) uiList.Item {
		_, isDefault := savedDefaultSystemPrompt[k]
		found := true
		if _, err := promptConfig.ChatMessages.FindMessageByContent(v); err != nil {
			if errors.Is(err, service.ErrNotFound) {
				found = false
			}
		}
		return uiList.Item{
			ItemId:          k,
			ItemTitle:       v,
			ItemDescription: lipgloss.JoinHorizontal(lipgloss.Center, "Added: "+helper.CheckedStringHelper(found), " | Default: "+helper.CheckedStringHelper(isDefault), " | Date: "+k),
		}
	})
	sort.Slice(res, func(i, j int) bool {
		return carbon.Parse(res[i].ItemId).Gt(carbon.Parse(res[j].ItemId))
	})

	return res
}

func getDelegateFn(promptConfig *service.PromptConfig) *uiList.DelegateFunctions {
	return &uiList.DelegateFunctions{
		ChooseFn: func(s string) tea.Cmd {
			savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")

			v, ok := viper.GetStringMapString("systems")[s]
			if !ok {
				return event.Error(errors.New(s + " not found in systems"))
			}
			newItem := uiList.Item{ItemId: s, ItemTitle: v}
			_, isDefault := savedDefaultSystemPrompt[s]
			exists, err := promptConfig.ChatMessages.AddMessage(v, service.RoleSystem)
			if err != nil {
				if errors.Is(err, service.ErrAlreadyExist) {
					promptConfig.ChatMessages.DeleteMessage(exists.Id)
					newItem.ItemDescription = lipgloss.JoinHorizontal(lipgloss.Center, "Added: "+helper.CheckedStringHelper(false), " | Default: "+helper.CheckedStringHelper(isDefault), " | Date: "+s)
					return func() tea.Msg {
						return newItem
					}
				}
				return event.Error(err)
			}

			newItem.ItemDescription = lipgloss.JoinHorizontal(lipgloss.Center, "Added: "+helper.CheckedStringHelper(true), " | Default: "+helper.CheckedStringHelper(isDefault), " | Date: "+s)
			return func() tea.Msg {
				return newItem
			}
		},
		EditFn: func(s string) tea.Cmd {
			savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")

			v, ok := viper.GetStringMapString("systems")[s]
			if !ok {
				return func() tea.Msg {
					return errors.New(s + " not found in systems")
				}
			}
			_, isDefault := savedDefaultSystemPrompt[s]

			editModel := form.NewEditModel("Editing system ["+s+"]", huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Content").Key(s).Value(&v).Lines(10),
				huh.NewSelect[bool]().Key("default").Title("Added by default").Value(&isDefault).Options(huh.NewOptions[bool](true, false)...),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString(s)
				UpdateFromSystemList(s, content)

				isDefault := form.GetBool("default")
				var err error
				if isDefault {
					err = SetDefaultSystem(s)
				} else {
					err = UnsetDefaultSystem(s)
				}

				if err != nil {
					return event.Error(err)
				}

				return func() tea.Msg {
					found := true
					if _, err := promptConfig.ChatMessages.FindMessageByContent(v); err != nil {
						if errors.Is(err, service.ErrNotFound) {
							found = false
						}
					}
					dft := "❌"
					if isDefault {
						dft = "✅"
					}
					return uiList.Item{ItemId: s, ItemTitle: content, ItemDescription: lipgloss.JoinHorizontal(lipgloss.Center, "Added: "+helper.CheckedStringHelper(found), "| Default: "+dft, " | Date: "+s)}
				}
			})

			return event.AddStack(editModel)
		},
		AddFn: func(s string) tea.Cmd {
			addModel := form.NewEditModel("New system", huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Content").Key(s).Lines(10),
				huh.NewSelect[bool]().Key("default").Title("Added by default").Options(huh.NewOptions[bool](true, false)...),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString(s)
				UpdateFromSystemList(s, content)

				isDefault := form.GetBool("default")
				var err error
				if isDefault {
					err = SetDefaultSystem(s)
				} else {
					err = UnsetDefaultSystem(s)
				}

				if err != nil {
					return event.Error(err)
				}

				return func() tea.Msg {
					dft := "❌"
					if isDefault {
						dft = "✅"
					}
					return uiList.Item{ItemId: s, ItemTitle: content, ItemDescription: lipgloss.JoinHorizontal(lipgloss.Center, "Added: ❌", "| Default: "+dft, " | Date: "+s)}
				}
			})

			return event.AddStack(addModel)
		},
		RemoveFn: func(s string) tea.Cmd {
			RemoveFromSystemList(s)
			return nil
		},
	}
}

func SendAsSystem() (string, error) {
	systemPrompt := promptui.Prompt{
		Label: "specify model behavior",
	}
	command, err := systemPrompt.Run()
	if err != nil {
		return "", err
	}

	if helper.YesNoPrompt("save prompt ?") {
		AddToSystemList(command, time.Now().Format("2006-01-02 15:04:05"))
	}

	return command, nil
}

func SetDefaultSystem(id string) error {
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	savedDefaultSystemPrompt[id] = ""
	viper.Set("default-systems", savedDefaultSystemPrompt)

	return viper.WriteConfig()
}

func UnsetDefaultSystem(id string) error {
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	delete(savedDefaultSystemPrompt, id)
	viper.Set("default-systems", savedDefaultSystemPrompt)
	return viper.WriteConfig()
}

func SetSystemDefault(unset bool) (commandToAdd []string, err error) {
	savedSystemPrompt := viper.GetStringMapString("systems")
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	keyStringFromSP := lo.Keys[string](savedSystemPrompt)
	sort.Slice(keyStringFromSP, func(i, j int) bool {
		return carbon.Parse(keyStringFromSP[i]).Gt(carbon.Parse(keyStringFromSP[j]))
	})
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
	}
	keys, err := SystemPrompt(savedSystemPrompt, func(i, w, h int) string {
		defaultStr := "❌"
		_, ok := savedDefaultSystemPrompt[keyStringFromSP[i]]

		switch {
		case i == -1:
			return ""
		case ok:
			defaultStr = "✅"
		}

		return fmt.Sprintf("Date: %s\nDefault: %s\n%s", keyStringFromSP[i], defaultStr, ui.AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromSP[i]]))
	})
	if err != nil {
		return
	}

	sendCommands := false
	if !unset {
		sendCommands = helper.YesNoPrompt("send commands ?")
	}

	for _, id := range keys {
		if unset {
			delete(savedDefaultSystemPrompt, id)
		} else {
			savedDefaultSystemPrompt[id] = ""
			if sendCommands {
				commandToAdd = append(commandToAdd, savedSystemPrompt[id])
			}
		}

	}

	viper.Set("default-systems", savedDefaultSystemPrompt)

	err = viper.WriteConfig()

	return
}

func SelectSystemCommand() ([]string, error) {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keys, err := SystemPrompt(savedSystemPrompt, nil)
	if err != nil {
		return nil, err
	}

	commandToSend := []string{}
	for _, id := range keys {
		commandToSend = append(commandToSend, savedSystemPrompt[id])
	}
	return commandToSend, nil
}

func SystemPrompt(savedSystemPrompt map[string]string, previewWindowFunc func(int, int, int) string) ([]string, error) {
	keyStringFromMap := lo.Keys[string](savedSystemPrompt)
	if len(keyStringFromMap) == 0 {
		return nil, errors.New("no saved systems")
	}
	sort.Slice(keyStringFromMap, func(i, j int) bool {
		return carbon.Parse(keyStringFromMap[i]).Gt(carbon.Parse(keyStringFromMap[j]))
	})
	if previewWindowFunc == nil {
		previewWindowFunc = func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Date: %s\n%s", keyStringFromMap[i], ui.AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromMap[i]]))
		}
	}

	idx, err := fuzzyfinder.FindMulti(
		keyStringFromMap,
		func(i int) string {
			return savedSystemPrompt[keyStringFromMap[i]]
		},
		fuzzyfinder.WithPreviewWindow(previewWindowFunc),
	)
	if err != nil {
		return nil, err
	}

	return lo.Map[int, string](idx, func(i int, _ int) string {
		return keyStringFromMap[i]
	}), nil
}

func DeleteSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keys, err := SystemPrompt(savedSystemPrompt, nil)
	if err != nil {
		return err
	}
	for _, id := range keys {
		RemoveFromSystemList(id)
		fmt.Printf("removed %s \n", id)
	}
	return nil
}

func AddToSystemList(content string, key string) {
	if key == "" {
		key = time.Now().Format("2006-01-02 15:04:05")
	}
	systems := viper.GetStringMapString("systems")
	systems[key] = content
	viper.Set("systems", systems)
	viper.WriteConfig()
}
func RemoveFromSystemList(time string) {
	systems := viper.GetStringMapString("systems")
	delete(systems, time)
	viper.Set("systems", systems)
	viper.WriteConfig()
}

func UpdateFromSystemList(time string, content string) {
	systems := viper.GetStringMapString("systems")
	systems[time] = content
	viper.Set("systems", systems)
	viper.WriteConfig()
}
