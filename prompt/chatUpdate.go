package prompt

import (
	"context"
	"errors"
	"io"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type ChatUpdateFunc func(m *chatModel) (tea.Model, tea.Cmd)

func reset(m *chatModel) (tea.Model, tea.Cmd) {
	m.textarea.Reset()
	return m, tea.EnterAltScreen
}

func closeContext(m *chatModel) (tea.Model, tea.Cmd) {
	if m.err != nil {
		m.err = nil
		return m, nil
	}
	err := m.promptConfig.CloseContextById(m.currentChatIndices.user)
	if err != nil {
		m.err = err
	}
	return m, nil
}

func quit(m *chatModel) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func changeResponseUp(m *chatModel) (tea.Model, tea.Cmd) {
	minIndex := lo.Min([]int{m.currentChatIndices.assistant, m.currentChatIndices.user})
	previous := minIndex - 1
	m.changeCurrentChatHelper(previous)
	m.viewport.GotoTop()
	return m, nil
}

func changeResponseDown(m *chatModel) (tea.Model, tea.Cmd) {
	maxIndex := lo.Max([]int{m.currentChatIndices.assistant, m.currentChatIndices.user})
	next := maxIndex + 1
	m.changeCurrentChatHelper(next)
	m.viewport.GotoTop()
	return m, nil
}

func callFunction(m *chatModel) (tea.Model, tea.Cmd) {
	v := m.textarea.Value()
	switch v {
	case "":
		m.viewport.SetContent(command.HELP)
		return m, nil
	case "\\quit":
		return m, tea.Quit
	case "\\help":
		m.viewport.SetContent(command.HELP)
		m.textarea.Reset()
		return m, nil
	}

	if v[0] == '\\' {
		m.textarea.Blur()
		err := commandSelectionFn(v, m.promptConfig)
		m.textarea.Reset()
		m.textarea.Focus()
		if err != nil {
			m.err = err
		}
		return m, tea.ClearScreen
	}
	return nil, nil
}

func promptSend(m *chatModel) (tea.Model, tea.Cmd) {
	m.userPrompt = m.textarea.Value()
	m.promptConfig.UserPrompt = m.userPrompt

	go func() {
		err := sendPrompt(m.promptConfig, m.currentChatIndices)
		if err != nil {
			m.err = err
		}
	}()

	m.textarea.Reset()
	m.aiResponse = ""

	m.viewport.GotoBottom()
	return m, waitForUpdate(m.promptConfig.UpdateChan)
}

func (m *chatModel) changeCurrentChatHelper(previous int) {
	messages := m.promptConfig.ChatMessages.FilterByOpenAIRoles()
	if len(messages) == 0 {
		m.currentChatIndices.assistant = -1
		m.currentChatIndices.user = -1
		return
	}
	if previous < 0 {
		previous = len(messages) - 1
	}
	prev := m.promptConfig.ChatMessages.FindById(previous)
	if prev == nil {
		prev = &messages[0]
	}

	switch prev.Role {
	case service.RoleAssistant:
		m.currentChatIndices.assistant = prev.Id
		m.currentChatIndices.user = prev.AssociatedMessageId
	case service.RoleUser:
		m.currentChatIndices.user = prev.Id
		m.currentChatIndices.assistant = prev.AssociatedMessageId
	}

	m.userPrompt = m.promptConfig.ChatMessages.FindById(m.currentChatIndices.user).Content
	m.aiResponse = m.promptConfig.ChatMessages.FindById(m.currentChatIndices.assistant).Content
}

func sendPrompt(pc *command.PromptConfig, currentChatIds *currentChatIndexes) error {
	userMsg, _ := pc.ChatMessages.AddMessage(pc.UserPrompt, service.RoleUser)
	assistantMessage, _ := pc.ChatMessages.AddMessage("", service.RoleAssistant)

	currentChatIds.user = userMsg.Id
	currentChatIds.assistant = assistantMessage.Id

	pc.ChatMessages.SetAssociatedId(userMsg.Id, assistantMessage.Id)

	llm, err := openai.New(openai.WithToken(viper.GetString("OPENAI_KEY")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, userMsg.Id)

	go llm.GenerateContent(ctx, pc.ChatMessages.ToLangchainMessage(), llms.WithModel(viper.GetString("model")), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		if err := ctx.Err(); err != nil {
			pc.DeleteContext(ctx)
			if err == io.EOF {
				return nil
			}
			return err
		}
		previous := pc.ChatMessages.FindById(assistantMessage.Id)
		if previous == nil {
			pc.DeleteContext(ctx)
			return errors.New("previous message not found")
		}
		previous.Content += string(chunk)
		pc.ChatMessages.UpdateMessage(*previous)
		if pc.UpdateChan != nil {
			pc.UpdateChan <- *previous
		}
		return nil
	}))

	if err != nil {
		return err
	}

	return nil
}
