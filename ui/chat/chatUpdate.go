package chat

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"moul.io/banner"
)

type ChatUpdateFunc func(m *chatModel) (tea.Model, tea.Cmd)

func getInfoContent(m chatModel) string {
	smallTitleStyle := style.TitleStyle.Copy().Margin(0).Padding(0, 2)
	return banner.Inline("go ai cli") + "\n" +
		lipgloss.NewStyle().AlignVertical(lipgloss.Center).Height(m.viewport.Height).Render(
			"Api : "+smallTitleStyle.Render(viper.GetString("API_TYPE"))+"\n"+
				"Model : "+smallTitleStyle.Render(viper.GetString("model"))+"\n"+
				"Messages : "+smallTitleStyle.Render(fmt.Sprintf("%d", len(m.promptConfig.ChatMessages.Messages)))+"\n"+
				"Tokens : "+smallTitleStyle.Render(fmt.Sprintf("%d", m.promptConfig.ChatMessages.TotalTokens))+"\n",
		)
}

func (m chatModel) resize() tea.Msg {
	return tea.WindowSizeMsg{Width: m.size.Width, Height: m.size.Height}
}

func closeContext(m chatModel) (chatModel, tea.Cmd) {
	if m.err != nil {
		m.err = nil
		return m, nil
	}
	if err := m.promptConfig.CloseContextById(m.currentChatMessages.user.Id.Int64()); err != nil {
		m.err = err
	}
	return m, nil
}

func changeResponseUp(m chatModel) (chatModel, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}

	var previous *service.ChatMessage
	if m.currentChatMessages.user == nil {
		previous = &m.promptConfig.ChatMessages.Messages[len(m.promptConfig.ChatMessages.Messages)-1]
	}

	if previous == nil {

		_, idx, _ := lo.FindIndexOf(m.promptConfig.ChatMessages.Messages, func(c service.ChatMessage) bool {
			return c.Id == m.currentChatMessages.user.Id
		})
		switch idx {
		case -1:
			return m, event.Error(errors.New("current message not found"))
		case 0:
			previous = &m.promptConfig.ChatMessages.Messages[len(m.promptConfig.ChatMessages.Messages)-1]
		default:
			previous = &m.promptConfig.ChatMessages.Messages[idx-1]
		}
	}
	m.changeCurrentChatHelper(previous)
	m.viewport.GotoTop()
	return m, tea.Sequence(event.Transition("clear"), event.UpdateChatContent("", ""), event.Transition(""))
}

func changeResponseDown(m chatModel) (chatModel, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}

	var previous *service.ChatMessage
	currentUserMsg := m.currentChatMessages.user

	if currentUserMsg == nil {
		previous = &m.promptConfig.ChatMessages.Messages[0]
	}

	if previous == nil {
		_, idx, _ := lo.FindIndexOf(m.promptConfig.ChatMessages.Messages, func(c service.ChatMessage) bool {
			return c.Id == currentUserMsg.Id
		})

		switch idx {
		case -1:
			return m, event.Error(errors.New("current message not found"))
		case len(m.promptConfig.ChatMessages.Messages) - 1:
			previous = &m.promptConfig.ChatMessages.Messages[0]
		case len(m.promptConfig.ChatMessages.Messages) - 2:
			if m.promptConfig.ChatMessages.Messages[idx+1].Id.Int64() == currentUserMsg.AssociatedMessageId {
				previous = &m.promptConfig.ChatMessages.Messages[0]
			}
		default:
			if currentUserMsg.Role == service.RoleUser &&
				m.promptConfig.ChatMessages.Messages[idx+1].Id.Int64() == currentUserMsg.AssociatedMessageId &&
				idx+2 < len(m.promptConfig.ChatMessages.Messages) {
				previous = &m.promptConfig.ChatMessages.Messages[idx+2]
			} else {
				previous = &m.promptConfig.ChatMessages.Messages[idx+1]
			}
		}
	}

	m.changeCurrentChatHelper(previous)
	m.viewport.GotoTop()
	return m, tea.Sequence(event.Transition("clear"), event.UpdateChatContent("", ""), event.Transition(""))
}

func promptSend(m *chatModel) (tea.Model, tea.Cmd) {
	m.userPrompt = m.textarea.Value()
	m.promptConfig.UserPrompt = m.userPrompt

	userMsg, err := m.promptConfig.ChatMessages.AddMessage(m.promptConfig.UserPrompt, service.RoleUser)
	if err != nil {
		return m, event.Error(err)
	}
	assistantMessage, err := m.promptConfig.ChatMessages.AddMessage("", service.RoleAssistant)
	if err != nil {
		return m, event.Error(err)
	}

	m.currentChatMessages.user = userMsg
	m.currentChatMessages.assistant = assistantMessage

	err = m.promptConfig.ChatMessages.SetAssociatedId(userMsg.Id.Int64(), assistantMessage.Id.Int64())
	if err != nil {
		return m, event.Error(err)
	}

	go sendPrompt(m.promptConfig, *m.currentChatMessages)

	m.textarea.Reset()
	m.aiResponse = ""

	m.viewport.SetContent("")

	m.viewport.GotoBottom()
	return m, tea.Sequence(event.Transition(m.userPrompt), waitForUpdate(m.promptConfig.UpdateChan), event.Transition(""))
}

func (m *chatModel) changeCurrentChatHelper(previous *service.ChatMessage) {
	if previous.AssociatedMessageId >= 0 {
		switch previous.Role {
		case service.RoleUser:
			m.currentChatMessages.user = previous
			m.currentChatMessages.assistant = m.promptConfig.ChatMessages.FindById(previous.AssociatedMessageId)
		case service.RoleAssistant:
			m.currentChatMessages.assistant = previous
			m.currentChatMessages.user = m.promptConfig.ChatMessages.FindById(previous.AssociatedMessageId)
		}
	} else {
		m.currentChatMessages.assistant = nil
		m.currentChatMessages.user = previous
	}

	if m.currentChatMessages.assistant != nil {
		m.aiResponse = m.currentChatMessages.assistant.Content
		m.userPrompt = m.currentChatMessages.user.Content
	} else {
		m.aiResponse = previous.Content
		m.userPrompt = "System / File | " + previous.Date.String()
	}

}

func sendPrompt(pc *service.PromptConfig, currentChatMsgs currentChatMessages) error {
	generate, err := api.GetGenerateFunction()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, currentChatMsgs.user.Id.Int64())
	defer pc.DeleteContextById(currentChatMsgs.user.Id.Int64())

	options := []llms.CallOption{
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			if err := ctx.Err(); err != nil {
				pc.DeleteContextById(currentChatMsgs.user.Id.Int64())
				if err == io.EOF {
					return nil
				}
				return err
			}
			previous := currentChatMsgs.assistant
			if previous == nil {
				pc.DeleteContextById(currentChatMsgs.user.Id.Int64())
				return errors.New("previous message not found")
			}
			previous.Content += string(chunk)
			pc.ChatMessages.UpdateMessage(*previous)
			if pc.UpdateChan != nil {
				pc.UpdateChan <- *previous
			}
			return nil
		}),
	}

	if v := viper.GetFloat64("temperature"); v >= 0 {
		options = append(options, llms.WithTemperature(v))
	}
	if v := viper.GetInt("topK"); v >= 0 {
		options = append(options, llms.WithTopK(v))

	}
	if v := viper.GetFloat64("topP"); v >= 0 {
		options = append(options, llms.WithTopP(v))
	}

	if pc.UpdateChan != nil {
		pc.UpdateChan <- *pc.ChatMessages.FindById(currentChatMsgs.assistant.Id.Int64())
	}

	_, err = generate(ctx, pc.ChatMessages.ToLangchainMessage(),
		options...,
	)

	if err != nil {
		chatProgram.Send(err)
		return err
	}

	return nil
}
