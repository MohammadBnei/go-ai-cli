package chat

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/command"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/pager"
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
	if err := m.promptConfig.CloseContextById(m.currentChatIndices.user); err != nil {
		m.err = err
	}
	return m, nil
}

func addPagerToStack(m chatModel) (chatModel, tea.Cmd) {
	if m.aiResponse == "" {
		return m, nil
	}

	_, index, ok := lo.FindIndexOf[tea.Model](m.stack, func(item tea.Model) bool {
		_, ok := item.(pager.PagerModel)
		return ok
	})
	if !ok {
		return m, event.AddStack(pager.NewPagerModel(m.userPrompt, m.aiResponse, m.promptConfig), "Loading Pager...")
	} else {
		m.stack = lo.Slice[tea.Model](m.stack, index-1, index)
	}
	return m, nil
}

func changeResponseUp(m chatModel) (chatModel, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}
	currentIndexes := lo.Filter[int]([]int{m.currentChatIndices.user, m.currentChatIndices.assistant}, func(i int, _ int) bool { return i >= 0 })
	minIndex := lo.Min(currentIndexes)
	previous := minIndex - 1
	if len(currentIndexes) == 0 {
		previous = len(m.promptConfig.ChatMessages.Messages) - 1
	}
	c := m.promptConfig.ChatMessages.FindById(previous)
	if c == nil {
		return m, event.Error(errors.New("no previous message"))
	}
	m.changeCurrentChatHelper(c)
	m.viewport.GotoTop()
	return m, event.UpdateChatContent("", "")
}

func changeResponseDown(m chatModel) (chatModel, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}
	maxIndex := lo.Max([]int{m.currentChatIndices.assistant, m.currentChatIndices.user})
	next := maxIndex + 1
	c := m.promptConfig.ChatMessages.FindById(next)
	if c == nil {
		return m, event.Error(errors.New("no next message"))
	}
	m.changeCurrentChatHelper(c)
	m.viewport.GotoTop()
	return m, event.UpdateChatContent("", "")
}

func callFunction(m *chatModel) (*chatModel, tea.Cmd) {
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

	userMsg, err := m.promptConfig.ChatMessages.AddMessage(m.promptConfig.UserPrompt, service.RoleUser)
	if err != nil {
		return m, event.Error(err)
	}
	assistantMessage, err := m.promptConfig.ChatMessages.AddMessage("", service.RoleAssistant)
	if err != nil {
		return m, event.Error(err)
	}

	m.currentChatIndices.user = userMsg.Id
	m.currentChatIndices.assistant = assistantMessage.Id

	m.promptConfig.ChatMessages.SetAssociatedId(userMsg.Id, assistantMessage.Id)

	go sendPrompt(m.promptConfig, m.currentChatIndices)
	// if err != nil {
	// 	m.err = err

	// 	m.promptConfig.ChatMessages.DeleteMessage(m.currentChatIndices.user)
	// 	m.promptConfig.ChatMessages.DeleteMessage(m.currentChatIndices.assistant)
	// 	m.currentChatIndices.assistant = -1
	// 	m.currentChatIndices.user = -1

	// 	m.userPrompt = ""
	// }

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
			m.currentChatIndices.user = previous.Id
			m.currentChatIndices.assistant = previous.AssociatedMessageId
		case service.RoleAssistant:
			m.currentChatIndices.assistant = previous.Id
			m.currentChatIndices.user = previous.AssociatedMessageId
		}
	} else {
		m.currentChatIndices.assistant = -1
		m.currentChatIndices.user = previous.Id
	}

	if m.currentChatIndices.assistant >= 0 {
		m.aiResponse = m.promptConfig.ChatMessages.FindById(m.currentChatIndices.assistant).Content
		m.userPrompt = m.promptConfig.ChatMessages.FindById(m.currentChatIndices.user).Content
	} else {
		m.aiResponse = previous.Content
		m.userPrompt = "System / File | " + previous.Date.String()
	}

}

func sendPrompt(pc *service.PromptConfig, currentChatIds *currentChatIndexes) error {
	generate, err := api.GetGenerateFunction()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, currentChatIds.user)
	defer pc.DeleteContextById(currentChatIds.user)

	_, err = generate(ctx, pc.ChatMessages.ToLangchainMessage(), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		if err := ctx.Err(); err != nil {
			pc.DeleteContextById(currentChatIds.user)
			if err == io.EOF {
				return nil
			}
			return err
		}
		previous := pc.ChatMessages.FindById(currentChatIds.assistant)
		if previous == nil {
			pc.DeleteContextById(currentChatIds.user)
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
		chatProgram.Send(err)
		return err
	}

	return nil
}
