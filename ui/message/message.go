package message

import (
	"strings"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/form"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/bwmarrin/snowflake"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

var (
	titleStyle     = lipgloss.NewStyle()
	systemColor    = titleStyle.Background(lipgloss.Color("#FFFFD9"))
	userColor      = titleStyle.Background(lipgloss.Color("#B32D00"))
	assistantColor = titleStyle.Background(lipgloss.Color("#243333"))
)

func NewMessageModel(promptConfig *service.PromptConfig) tea.Model {
	items := getItemsAslist(promptConfig)

	delegateFn := getDelegateFn(promptConfig)

	return list.NewFancyListModel("message", items, delegateFn)
}

func getDelegateFn(promptConfig *service.PromptConfig) *list.DelegateFunctions {
	return &list.DelegateFunctions{
		AddFn: func(s string) tea.Cmd {
			editModel := form.NewEditModel("Creating message", huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Content").Key("content").Lines(10),
				huh.NewSelect[service.ROLES]().Key("role").Title("Role").Options(huh.NewOptions[service.ROLES]([]service.ROLES{service.RoleAssistant, service.RoleUser, service.RoleSystem}...)...),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString("content")
				role := form.Get("role").(service.ROLES)
				msg, err := promptConfig.ChatMessages.AddMessage(content, role)
				if err != nil {
					return event.Error(err)
				}
				return func() tea.Msg {
					return toItem(*msg)
				}
			})

			return event.AddStack(editModel, "Creating Message...")
		},
		ChooseFn: func(s string) tea.Cmd {
			message, err := getMessage(promptConfig, s)
			if err != nil {
				return event.Error(err)
			}

			editModel := form.NewEditModel("Editing message ["+s+"]", huh.NewForm(huh.NewGroup(
				huh.NewText().Title("Content").Key(s).Value(&message.Content).Lines(10),
				huh.NewSelect[service.ROLES]().Key("role").Title("Role").Options(huh.NewOptions[service.ROLES]([]service.ROLES{service.RoleAssistant, service.RoleUser, service.RoleSystem}...)...),
			)), func(form *huh.Form) tea.Cmd {
				content := form.GetString(s)
				role := form.Get("role").(service.ROLES)
				message.Content = content
				message.Role = role
				err := promptConfig.ChatMessages.UpdateMessage(*message)
				if err != nil {
					return event.Error(err)
				}
				return func() tea.Msg {
					return toItem(*message)
				}
			})

			return event.AddStack(editModel, "Editing Message...")
		},
		RemoveFn: func(s string) tea.Cmd {
			snowflake.ParseBase64(s)
			id, err := snowflake.ParseBase64(s)
			if err != nil {
				return event.Error(err)
			}
			err = promptConfig.ChatMessages.DeleteMessage(id.Int64())
			if err != nil {
				return event.Error(err)
			}

			return nil
		},
	}
}

func getMessage(promptConfig *service.PromptConfig, id string) (*service.ChatMessage, error) {
	intId, err := snowflake.ParseBase64(id)
	if err != nil {
		return nil, err
	}
	message := promptConfig.ChatMessages.FindById(intId.Int64())
	if message == nil {
		return nil, err
	}

	return message, nil
}

func getItemsAslist(promptConfig *service.PromptConfig) []list.Item {
	messages := promptConfig.ChatMessages.FilterByOpenAIRoles()

	res := lo.Map(messages, func(m service.ChatMessage, _ int) list.Item {
		return toItem(m)
	})

	return res
}

func toItem(message service.ChatMessage) list.Item {
	choosenStyle := systemColor
	switch message.Role {
	case service.RoleSystem:
		choosenStyle = systemColor
	case service.RoleAssistant:
		choosenStyle = assistantColor
	case service.RoleUser:
		choosenStyle = userColor
	}

	splitted := strings.Split(strings.TrimSpace(message.Content), "\n")
	title := splitted[0]
	if len(splitted) > 1 {
		title += "..."
	}

	return list.Item{
		ItemId:          message.Id.Base64(),
		ItemTitle:       choosenStyle.Render(title),
		ItemDescription: string(message.Role),
	}
}
