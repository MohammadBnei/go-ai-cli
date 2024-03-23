package chat

import (
	"context"
	"errors"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-module/carbon"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"moul.io/banner"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	godcontext "github.com/MohammadBnei/go-ai-cli/service/godcontext"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
)

type ChatUpdateFunc func(m *chatModel) (tea.Model, tea.Cmd)

func getInfoContent(m chatModel) string {
	smallTitleStyle := style.TitleStyle.Copy().Margin(0).Padding(0, 2)
	return banner.Inline("go ai cli") + "\n" +
		lipgloss.NewStyle().AlignVertical(lipgloss.Center).Height(m.viewport.Height).Render(
			"Api : "+smallTitleStyle.Render(viper.GetString(config.AI_API_TYPE))+"\n"+
				"Model : "+smallTitleStyle.Render(viper.GetString(config.AI_MODEL_NAME))+"\n"+
				"Messages : "+smallTitleStyle.Render(fmt.Sprintf("%d", len(m.services.ChatMessages.Messages)))+"\n"+
				"Tokens : "+smallTitleStyle.Render(fmt.Sprintf("%d", m.services.ChatMessages.TotalTokens))+"\n",
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
	if err := m.services.Contexts.CloseContextById(m.currentChatMessages.user.Id.Int64()); err != nil {
		m.err = err
	}
	return m, nil
}

func changeResponseUp(m chatModel) (chatModel, tea.Cmd) {
	if len(m.services.ChatMessages.Messages) == 0 {
		return m, nil
	}

	var msg *service.ChatMessage
	if m.currentChatMessages.user == nil && m.currentChatMessages.assistant == nil {
		msg = &m.services.ChatMessages.Messages[0]
	} else {
		order := 0
		if m.currentChatMessages.user == nil {
			order = int(m.currentChatMessages.assistant.Order)
		} else {
			order = int(m.currentChatMessages.user.Order)
		}
		if m.currentChatMessages.assistant != nil && int(m.currentChatMessages.assistant.Order) < order {
			order = int(m.currentChatMessages.assistant.Order)
		}

		if order == 1 {
			msg = &m.services.ChatMessages.Messages[len(m.services.ChatMessages.Messages)-1]
		} else {
			msg = m.services.ChatMessages.FindByOrder(uint(order - 1))
		}
	}

	m.changeCurrentChatHelper(msg)
	m.viewport.GotoTop()
	return m, tea.Sequence(event.Transition("clear"), event.UpdateChatContent("", ""), event.Transition(""))
}

func changeResponseDown(m chatModel) (chatModel, tea.Cmd) {
	if len(m.services.ChatMessages.Messages) == 0 {
		return m, nil
	}

	var msg *service.ChatMessage
	if m.currentChatMessages.user == nil && m.currentChatMessages.assistant == nil {
		msg = &m.services.ChatMessages.Messages[0]
	} else {
		order := 0
		if m.currentChatMessages.user == nil {
			order = int(m.currentChatMessages.assistant.Order)
		} else {
			order = int(m.currentChatMessages.user.Order)
		}
		if m.currentChatMessages.assistant != nil && int(m.currentChatMessages.assistant.Order) > order {
			order = int(m.currentChatMessages.assistant.Order)
		}

		if order >= len(m.services.ChatMessages.Messages) {
			msg = &m.services.ChatMessages.Messages[0]
		} else {
			msg = m.services.ChatMessages.FindByOrder(uint(order + 1))
		}
	}

	m.changeCurrentChatHelper(msg)
	m.viewport.GotoTop()
	return m, tea.Sequence(event.Transition("clear"), event.UpdateChatContent("", ""), event.Transition(""))
}

func promptSend(m *chatModel) (tea.Model, tea.Cmd) {
	m.title = m.textarea.Value()

	userMsg, err := m.services.ChatMessages.AddMessage(m.title, service.RoleUser)
	if err != nil {
		return m, event.Error(err)
	}
	assistantMessage, err := m.services.ChatMessages.AddMessage("", service.RoleAssistant)
	if err != nil {
		return m, event.Error(err)
	}

	m.currentChatMessages.user = userMsg
	m.currentChatMessages.assistant = assistantMessage

	err = m.services.ChatMessages.SetAssociatedId(userMsg.Id.Int64(), assistantMessage.Id.Int64())
	if err != nil {
		return m, event.Error(err)
	}

	switch {
	case m.chain != nil:
		go sendAgentPrompt(*m, *m.currentChatMessages)
		m.chain = nil
	default:
		go sendPrompt(m.services, *m.currentChatMessages)
	}

	m.textarea.Reset()
	m.content = ""

	m.viewport.SetContent("")

	m.viewport.GotoBottom()
	return m, tea.Sequence(event.Transition(m.title), waitForUpdate(m.services.UpdateChan), event.Transition(""))
}

func (m *chatModel) changeCurrentChatHelper(msg *service.ChatMessage) {
	if msg == nil {
		m.err = errors.New("msg is nil, (changeCurrentChatHelper)")
		return
	}
	if msg.AssociatedMessageId != 0 || msg.Role == service.RoleSystem {
		switch msg.Role {
		case service.RoleUser:
			m.currentChatMessages.user = msg
			m.currentChatMessages.assistant = m.services.ChatMessages.FindById(msg.AssociatedMessageId)
		case service.RoleAssistant:
			m.currentChatMessages.assistant = msg
			m.currentChatMessages.user = m.services.ChatMessages.FindById(msg.AssociatedMessageId)
		case service.RoleSystem:
			m.title = "System / File | " + msg.Date.String()
			m.currentChatMessages.user = msg
			m.content = msg.Content
			m.currentChatMessages.assistant = nil
			return
		}
	} else {
		m.currentChatMessages.user = msg
	}

	if m.currentChatMessages.assistant != nil && m.currentChatMessages.user != nil {
		m.content = m.currentChatMessages.assistant.Content
		m.title = m.currentChatMessages.user.Content
	} else {
		m.content = msg.Content
		m.title = carbon.FromStdTime(msg.Date).String()
	}

}

func sendPrompt(pc *service.Services, currentChatMsgs currentChatMessages) error {
	llm, err := api.GetLlmModel()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(godcontext.GodContext)
	pc.Contexts.AddContextWithId(ctx, cancel, currentChatMsgs.user.Id.Int64())
	defer pc.Contexts.CloseContextById(currentChatMsgs.user.Id.Int64())

	options := []llms.CallOption{
		llms.WithStreamingFunc(MakeStreamingFunc(currentChatMsgs.user.Id.Int64(), currentChatMsgs.assistant.Id.Int64(), pc)),
	}

	if v := viper.GetFloat64(config.AI_TEMPERATURE); v >= 0 {
		options = append(options, llms.WithTemperature(v))
	}
	if v := viper.GetInt(config.AI_TOP_K); v >= 0 {
		options = append(options, llms.WithTopK(v))

	}
	if v := viper.GetFloat64(config.AI_TOP_P); v >= 0 {
		options = append(options, llms.WithTopP(v))
	}

	if pc.UpdateChan != nil {
		pc.UpdateChan <- *pc.ChatMessages.FindById(currentChatMsgs.assistant.Id.Int64())
	}

	if viper.GetBool(config.C_COMPLETION_MODE) {
		_, err = llms.GenerateFromSinglePrompt(ctx, llm, pc.UserPrompt, options...)
	} else {
		_, err = llm.GenerateContent(ctx, pc.ChatMessages.ToLangchainMessage(),
			options...,
		)
	}

	if err != nil {
		ChatProgram.Send(err)
		return err
	}

	ChatProgram.Send(event.DoneGenerating(currentChatMsgs.user.Id.Int64(), currentChatMsgs.assistant.Id.Int64()))

	return nil
}

func sendAgentPrompt(m chatModel, currentChatMsgs currentChatMessages) error {
	ctx, cancel := context.WithCancel(godcontext.GodContext)
	m.services.Contexts.AddContextWithId(ctx, cancel, currentChatMsgs.user.Id.Int64())
	defer m.services.Contexts.CloseContext(ctx)

	assitantMsgID := currentChatMsgs.assistant.Id.Int64()
	userID := currentChatMsgs.user.Id.Int64()

	if m.services.UpdateChan != nil {
		m.services.UpdateChan <- *m.services.ChatMessages.FindById(assitantMsgID)
	}

	userMessages, _ := m.services.ChatMessages.FilterMessages(service.RoleUser)
	last, err := lo.Last(userMessages)
	if err != nil {
		return err
	}

	output, err := chains.Run(ctx, m.chain, last.Content, chains.WithStreamingFunc(MakeStreamingFunc(userID, assitantMsgID, m.services)))

	if err != nil {
		ChatProgram.Send(err)
		return err
	}

	currentChatMsgs.assistant.Content = output
	currentChatMsgs.assistant.Meta.Agent = m.chainName
	currentChatMsgs.user.Meta.Agent = m.chainName
	m.services.ChatMessages.UpdateMessage(*currentChatMsgs.assistant)
	if m.services.UpdateChan != nil {
		m.services.UpdateChan <- *currentChatMsgs.assistant
	}

	return nil
}

func MakeStreamingFunc(userMsgID, assistantMsgID int64, services *service.Services) func(ctx context.Context, chunk []byte) error {
	handleErr := func(err error) error {
		services.Contexts.CloseContextById(userMsgID)
		if err == io.EOF {
			return nil
		}
		return err
	}
	return func(ctx context.Context, chunk []byte) error {

		if err := ctx.Err(); err != nil {
			return handleErr(err)
		}
		updated, err := services.ChatMessages.AppendToMessageContent(assistantMsgID, string(chunk))
		if err != nil {
			return handleErr(err)
		}
		if services.UpdateChan != nil {
			services.UpdateChan <- updated
		}
		return nil
	}
}
