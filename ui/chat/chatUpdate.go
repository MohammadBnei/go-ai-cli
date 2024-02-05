package chat

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/pager"
	"github.com/MohammadBnei/go-openai-cli/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"moul.io/banner"
)

type ChatUpdateFunc func(m *chatModel) (tea.Model, tea.Cmd)

func reset(m chatModel) (chatModel, tea.Cmd) {
	m.textarea.Reset()
	m.aiResponse = getInfoContent(m)

	m.userPrompt = "Infos"
	m.currentChatIndices = &currentChatIndexes{
		user:      -1,
		assistant: -1,
	}

	err := viper.ReadInConfig()
	if err != nil {
		m.err = err
	}

	return m, event.UpdateContent
}

func getInfoContent(m chatModel) string {
	smallTitleStyle := style.TitleStyle.Margin(0).Padding(0, 2)
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
	err := m.promptConfig.CloseContextById(m.currentChatIndices.user)
	if err != nil {
		m.err = err
	}
	return m, nil
}

func quit(m chatModel) (chatModel, tea.Cmd) {
	return m, tea.Quit
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
		return m, event.AddStack(pager.NewPagerModel(m.userPrompt, m.aiResponse, m.promptConfig))
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
	return m, event.UpdateContent
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
	return m, event.UpdateContent
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
	userMsg, err := pc.ChatMessages.AddMessage(pc.UserPrompt, service.RoleUser)
	if err != nil {
		return err
	}
	assistantMessage, err := pc.ChatMessages.AddMessage("", service.RoleAssistant)
	if err != nil {
		return err
	}

	currentChatIds.user = userMsg.Id
	currentChatIds.assistant = assistantMessage.Id

	pc.ChatMessages.SetAssociatedId(userMsg.Id, assistantMessage.Id)

	generate, err := api.GetGenerateFunction()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, userMsg.Id)
	defer pc.DeleteContextById(userMsg.Id)

	_, err = generate(ctx, pc.ChatMessages.ToLangchainMessage(), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		if err := ctx.Err(); err != nil {
			pc.DeleteContextById(userMsg.Id)
			if err == io.EOF {
				return nil
			}
			return err
		}
		previous := pc.ChatMessages.FindById(assistantMessage.Id)
		if previous == nil {
			pc.DeleteContextById(userMsg.Id)
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
