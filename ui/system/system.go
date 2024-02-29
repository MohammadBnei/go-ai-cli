package system

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/form"
	"github.com/MohammadBnei/go-ai-cli/ui/helper"
	uiList "github.com/MohammadBnei/go-ai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-module/carbon"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func NewSystemModel(promptConfig *service.PromptConfig) tea.Model {
	savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
		viper.Set(config.PR_SYSTEM_DEFAULT, savedDefaultSystemPrompt)
	}

	items := getItemsAsUIList(promptConfig)

	delegateFn := getDelegateFn(promptConfig)

	return uiList.NewFancyListModel("system", items, delegateFn)
}

func getItemsAsUIList(promptConfig *service.PromptConfig) []uiList.Item {
	savedSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM)
	savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)

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
			ItemTitle:       strings.ReplaceAll(v, "\n\n", "\n"),
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
			savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)

			v, ok := viper.GetStringMapString(config.PR_SYSTEM)[s]
			if !ok {
				return event.Error(errors.New(s + " not found in systems"))
			}
			newItem := uiList.Item{ItemId: s, ItemTitle: v}
			_, isDefault := savedDefaultSystemPrompt[s]
			exists, err := promptConfig.ChatMessages.AddMessage(v, service.RoleSystem)
			if err != nil {
				if errors.Is(err, service.ErrAlreadyExist) {
					promptConfig.ChatMessages.DeleteMessage(exists.Id.Int64())
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
			savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)

			v, ok := viper.GetStringMapString(config.PR_SYSTEM)[s]
			if !ok {
				return func() tea.Msg {
					return errors.New(s + " not found in systems")
				}
			}
			_, isDefault := savedDefaultSystemPrompt[s]

			tRue := true

			editModel := form.NewEditModel("Editing system ["+s+"]", huh.NewForm(huh.NewGroup(
				huh.NewText().Editor("nvim").CharLimit(0).Title("Content").Key(s).Value(&v).Lines(10),
				huh.NewSelect[bool]().Key("default").Title("Added by default").Value(&isDefault).Options(huh.NewOptions[bool](true, false)...),
				huh.NewSelect[bool]().Key("add").Title("Add it ?").Options(huh.NewOptions[bool](true, false)...).Value(&tRue),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString(s)
				addIt := form.GetBool("add")

				if addIt {
					promptConfig.ChatMessages.AddMessage(content, service.RoleSystem)
				}

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

			return event.AddStack(editModel, "Editing "+s+"...")
		},
		AddFn: func(_ string) tea.Cmd {
			tRue := true

			addModel := form.NewEditModel("New system", huh.NewForm(huh.NewGroup(
				huh.NewText().Editor("nvim").CharLimit(0).Title("Content").Key("content").Lines(10).Validate(func(s string) error {
					if s == "" {
						return errors.New("content cannot be empty")
					}
					return nil
				}),
				huh.NewSelect[bool]().Key("default").Title("Added by default").Options(huh.NewOptions[bool](true, false)...),
				huh.NewSelect[bool]().Key("save").Title("Save it ?").Options(huh.NewOptions[bool](true, false)...).Value(&tRue),
				huh.NewSelect[bool]().Key("add").Title("Add it ?").Options(huh.NewOptions[bool](true, false)...).Value(&tRue),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString("content")
				saveIt := form.GetBool("save")
				addIt := form.GetBool("add")

				if addIt {
					promptConfig.ChatMessages.AddMessage(content, service.RoleSystem)
				}

				if !saveIt {
					return nil
				}
				title := time.Now().Format("2006-01-02 15:04:05")
				UpdateFromSystemList(title, content)

				isDefault := form.GetBool("default")
				var err error
				if isDefault {
					err = SetDefaultSystem(title)
				} else {
					err = UnsetDefaultSystem(title)
				}

				if err != nil {
					return event.Error(err)
				}

				return func() tea.Msg {
					dft := "❌"
					if isDefault {
						dft = "✅"
					}
					return uiList.Item{ItemId: title, ItemTitle: content, ItemDescription: lipgloss.JoinHorizontal(lipgloss.Center, "Added: ❌", "| Default: "+dft, " | Date: "+title)}
				}
			})

			return event.AddStack(addModel, "Adding new system...")
		},
		RemoveFn: func(s string) tea.Cmd {
			RemoveFromSystemList(s)
			return nil
		},
	}
}

func SetDefaultSystem(id string) error {
	savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
	savedDefaultSystemPrompt[id] = ""
	viper.Set(config.PR_SYSTEM_DEFAULT, savedDefaultSystemPrompt)

	return viper.WriteConfig()
}

func UnsetDefaultSystem(id string) error {
	savedDefaultSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
	delete(savedDefaultSystemPrompt, id)
	viper.Set(config.PR_SYSTEM_DEFAULT, savedDefaultSystemPrompt)
	return viper.WriteConfig()
}

func AddToSystemList(content string, key string) {
	if key == "" {
		key = time.Now().Format("2006-01-02 15:04:05")
	}
	systems := viper.GetStringMapString(config.PR_SYSTEM)
	systems[key] = content
	viper.Set("systems", systems)
	viper.WriteConfig()
}
func RemoveFromSystemList(time string) {
	systems := viper.GetStringMapString(config.PR_SYSTEM)
	delete(systems, time)
	viper.Set("systems", systems)
	viper.WriteConfig()
}

func UpdateFromSystemList(time string, content string) {
	systems := viper.GetStringMapString(config.PR_SYSTEM)
	systems[time] = content
	viper.Set("systems", systems)
	viper.WriteConfig()
}
