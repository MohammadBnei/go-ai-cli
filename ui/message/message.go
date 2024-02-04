package message

import (
	"fmt"
	"strconv"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/form"
	"github.com/MohammadBnei/go-openai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

var (
	titleStyle     = lipgloss.NewStyle()
	systemColor    = titleStyle.Foreground(lipgloss.Color("#FFFFD9"))
	userColor      = titleStyle.Foreground(lipgloss.Color("#C2CCB8"))
	assistantColor = titleStyle.Foreground(lipgloss.Color("#6C8073"))
)

func NewMessageModel(promptConfig *service.PromptConfig) tea.Model {
	items := getItemsAslist(promptConfig)

	delegateFn := getDelegateFn(promptConfig)

	return list.NewFancyListModel("system", items, delegateFn)
}

func getItemsAslist(promptConfig *service.PromptConfig) []list.Item {
	messages := promptConfig.ChatMessages.FilterByOpenAIRoles()

	res := lo.Map(messages, func(m service.ChatMessage, _ int) list.Item {
		return toItem(m)
	})

	return res
}

func toItem(message service.ChatMessage) list.Item {
	title := message.Content
	switch message.Role {
	case service.RoleSystem:
		title = systemColor.Render(title)
	case service.RoleAssistant:
		title = assistantColor.Render(title)
	case service.RoleUser:
		title = userColor.Render(title)

	}
	return list.Item{
		ItemId:          fmt.Sprintf("%d", message.Id),
		ItemTitle:       title,
		ItemDescription: string(message.Role),
	}
}

func getDelegateFn(promptConfig *service.PromptConfig) *list.DelegateFunctions {
	return &list.DelegateFunctions{
		ChooseFn: func(s string) tea.Cmd {
			message, err := getMessage(promptConfig, s)
			if err != nil {
				return event.Error(err)
			}

			editModel := form.NewEditModel("Editing message ["+s+"]", huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Content").Key(s).Value(&message.Content).Lines(10),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString(s)
				message.Content = content
				err := promptConfig.ChatMessages.UpdateMessage(*message)
				if err != nil {
					return event.Error(err)
				}
				return func() tea.Msg {
					return toItem(*message)
				}
			})

			return event.AddStack(editModel)
		},
		RemoveFn: func(s string) tea.Cmd {
			id, err := strconv.Atoi(s)
			if err == nil {
				return event.Error(err)
			}
			err = promptConfig.ChatMessages.DeleteMessage(id)
			if err != nil {
				return event.Error(err)
			}

			return nil
		},
	}
}

func getMessage(promptConfig *service.PromptConfig, id string) (*service.ChatMessage, error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	message := promptConfig.ChatMessages.FindById(intId)
	if message == nil {
		return nil, err
	}

	return message, nil
}
