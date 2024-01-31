package prompt

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"context"
	"fmt"
	"log"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"moul.io/banner"
)

type ChatChan struct {
	UpdateChan chan service.ChatMessage
	UserPrompt chan string
	Done       chan bool
}

func Chat(chatChannels *ChatChan, pc *command.PromptConfig) {
	p := tea.NewProgram(initialChatModel(chatChannels, pc), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type currentChatIndexes struct {
	user      int
	assistant int
}
type chatModel struct {
	viewport           viewport.Model
	promptConfig       *command.PromptConfig
	textarea           textarea.Model
	err                error
	spinner            spinner.Model
	userPrompt         string
	userStyle          lipgloss.Style
	assistantStyle     lipgloss.Style
	aiResponse         string
	currentChatIndices *currentChatIndexes
}

var terminalWidth, terminalHeight, _ = ui.GetTerminalSize()

func initialChatModel(chatChannels *ChatChan, pc *command.PromptConfig) chatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 0

	ta.SetWidth(terminalWidth)
	ta.SetHeight(2)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(terminalWidth, terminalHeight-3)

	vp.SetContent(banner.Inline("go ai cli - prompt"))

	ta.KeyMap.InsertNewline.SetEnabled(false)

	modelStruct := chatModel{
		textarea:     ta,
		promptConfig: pc,
		viewport:     vp,
		err:          nil,
		spinner:      spinner.New(),
		aiResponse:   "",
		userPrompt:   "",
		currentChatIndices: &currentChatIndexes{
			user:      -1,
			assistant: -1,
		},
		userStyle: lipgloss.NewStyle().Background(lipgloss.Color("#595302")).Foreground(lipgloss.Color("#8947C8")).Bold(true).Padding(1).Margin(1).Width(terminalWidth),
	}

	return modelStruct
}

func (m chatModel) Init() tea.Cmd {
	return nil
}

var commandSelectionFn = CommandSelectionFactory()

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlR:
			m.textarea.Reset()
			m.initViewport()
			return m, tea.EnterAltScreen

		case tea.KeyCtrlC:
			if m.err != nil {
				m.err = nil
				return m, nil
			}
			err := m.promptConfig.CloseContextById(m.currentChatIndices.user)
			if err != nil {
				m.err = err
			}

		case tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyShiftUp:
			previous := m.currentChatIndices.assistant - 2
			m.changeCurrentChatHelper(previous)
			m.viewport.GotoTop()
			return m, nil

		case tea.KeyShiftDown:
			previous := m.currentChatIndices.assistant + 2
			m.changeCurrentChatHelper(previous)
			m.viewport.GotoTop()
			return m, nil

		case tea.KeyCtrlP:
			Pager(m.userPrompt, m.aiResponse)

		case tea.KeyEnter:
			if m.err != nil {
				m.err = nil
				return m, nil
			}
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

			m.userPrompt = v
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
	case service.ChatMessage:
		if msg.Id == m.currentChatIndices.user {
			m.userPrompt = msg.Content
		}

		if msg.Id == m.currentChatIndices.assistant {
			m.aiResponse = msg.Content
		}

		return m, waitForUpdate(m.promptConfig.UpdateChan)

	// We handle errors just like any other message

	case errMsg:
		m.err = msg
		return m, nil

	default:
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m chatModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %s", m.err)
	}

	if m.userPrompt != "" {
		aiRes := m.aiResponse
		if m.promptConfig.MdMode {
			aiRes = string(markdown.Render(m.aiResponse, terminalWidth, 3))
		}
		m.viewport.SetContent(fmt.Sprintf("%s\n%s", m.userStyle.Render(m.userPrompt), aiRes))
	}

	return fmt.Sprintf(
		"%s\n%s\n",
		m.viewport.View(),
		m.textarea.View(),
	)
}

func waitForUpdate(updateChan chan service.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-updateChan
	}
}

func (m *chatModel) changeCurrentChatHelper(previous int) {
	if len(m.promptConfig.ChatMessages.Messages) == 0 {
		m.currentChatIndices.assistant = -1
		m.currentChatIndices.user = -1
		return
	}
	if previous < 0 {
		previous = len(m.promptConfig.ChatMessages.Messages) - 1
	}
	prev := m.promptConfig.ChatMessages.FindById(previous)
	if prev == nil {
		prev = &m.promptConfig.ChatMessages.Messages[0]
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

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, userMsg.Id)

	stream, err := api.SendPromptToOpenAi(ctx, &api.GPTChanRequest{
		Messages: pc.ChatMessages.FilterByOpenAIRoles(),
	})
	if err != nil {
		return err
	}

	go func(stream <-chan *api.GPTChanResponse) {
		defer pc.DeleteContext(ctx)
		for v := range stream {
			previous := pc.ChatMessages.FindById(assistantMessage.Id)
			if previous == nil {
				log.Fatalln("previous message not found")
			}
			previous.Content += string(v.Content)
			pc.ChatMessages.UpdateMessage(*previous)
			if pc.UpdateChan != nil {
				pc.UpdateChan <- *previous
			}
		}
	}(stream)

	return nil
}

func (m *chatModel) initViewport() (tea.Model, tea.Cmd) {
	w, h, err := ui.GetTerminalSize()
	terminalHeight = h
	terminalWidth = w

	if err != nil {
		m.err = err
		return m, nil
	}

	vp := viewport.New(terminalWidth, terminalHeight-3)

	vp.SetContent(banner.Inline("go ai cli - prompt"))

	m.viewport = vp

	return m, nil
}
