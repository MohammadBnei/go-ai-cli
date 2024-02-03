package prompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"golang.org/x/term"
)

type ChatUpdateFunc func(m *chatModel) (tea.Model, tea.Cmd)

func reset(m *chatModel) (tea.Model, tea.Cmd) {
	m.textarea.Reset()
	paragraphStyle := lipgloss.NewStyle().Padding(2)
	m.aiResponse = lipgloss.JoinHorizontal(lipgloss.Center,
		paragraphStyle.Render("Tokens : "+fmt.Sprintf("%d", m.promptConfig.ChatMessages.TotalTokens)),
		paragraphStyle.Render("Messages : "+fmt.Sprintf("%d", len(m.promptConfig.ChatMessages.Messages))),
	)
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

func (m *chatModel) resize() tea.Msg {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	return tea.WindowSizeMsg{
		Width:  w,
		Height: h,
	}
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

func addPagerToStack(m *chatModel) (tea.Model, tea.Cmd) {
	if m.aiResponse == "" {
		return m, nil
	}

	_, index, ok := lo.FindIndexOf[tea.Model](m.stack, func(item tea.Model) bool {
		_, ok := item.(pagerModel)
		return ok
	})
	if !ok {
		pager := pagerModel{
			title:   m.userPrompt,
			content: m.aiResponse,
			pc:      m.promptConfig,
		}
		p, cmd := pager.Update(m.size)
		pager = p.(pagerModel)

		m.stack = append(m.stack, pager)

		return m, tea.Batch(m.stack[len(m.stack)-1].Init(), cmd)
	} else {
		m.stack = lo.Slice[tea.Model](m.stack, index-1, index)
	}
	return m, nil
}

func changeResponseUp(m *chatModel) (tea.Model, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}
	minIndex := lo.Min(lo.Filter[int]([]int{m.currentChatIndices.user, m.currentChatIndices.assistant}, func(i int, _ int) bool { return i >= 0 }))
	previous := minIndex - 1
	c := m.promptConfig.ChatMessages.FindById(previous)
	if c == nil {
		c = &m.promptConfig.ChatMessages.Messages[len(m.promptConfig.ChatMessages.Messages)-1]
	}
	m.changeCurrentChatHelper(c)
	m.viewport.GotoTop()
	return m, event.UpdateContent
}

func changeResponseDown(m *chatModel) (tea.Model, tea.Cmd) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		return m, nil
	}
	maxIndex := lo.Max([]int{m.currentChatIndices.assistant, m.currentChatIndices.user})
	next := maxIndex + 1
	c := m.promptConfig.ChatMessages.FindById(next)
	if c == nil {
		c = &m.promptConfig.ChatMessages.Messages[0]
	}
	m.changeCurrentChatHelper(c)
	m.viewport.GotoTop()
	return m, event.UpdateContent
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
	userMsg, _ := pc.ChatMessages.AddMessage(pc.UserPrompt, service.RoleUser)
	assistantMessage, _ := pc.ChatMessages.AddMessage("", service.RoleAssistant)

	currentChatIds.user = userMsg.Id
	currentChatIds.assistant = assistantMessage.Id

	pc.ChatMessages.SetAssociatedId(userMsg.Id, assistantMessage.Id)

	var generateContent func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)
	var err error
	llmModel := llms.WithModel(viper.GetString("model"))

	// TODO: inverse this condition
	if !pc.OllamaMode {
		llama, e := ollama.New()
		err = e
		generateContent = llama.GenerateContent
		llmModel = llms.WithModel("llama2-uncensored")
	} else {
		llm, e := openai.New(openai.WithToken(viper.GetString("OPENAI_KEY")))
		err = e
		generateContent = llm.GenerateContent
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, userMsg.Id)
	defer pc.DeleteContextById(userMsg.Id)

	_, err = generateContent(ctx, pc.ChatMessages.ToLangchainMessage(), llmModel, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
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
