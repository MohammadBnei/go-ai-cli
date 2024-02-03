package prompt

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"reflect"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/MohammadBnei/go-openai-cli/ui/config"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/list"
	"github.com/MohammadBnei/go-openai-cli/ui/system"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"moul.io/banner"
)

var (
	AppStyle = lipgloss.NewStyle().Margin(1, 2, 0)

	userStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8947C8")).Bold(true).Margin(0, 2, 2)
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

func DefaultStyle() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.NormalBorder()).Padding(1)
	return s
}

func Chat(pc *service.PromptConfig) {
	p := tea.NewProgram(initialChatModel(pc),
		tea.WithAltScreen())

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
	textarea           textarea.Model
	promptConfig       *service.PromptConfig
	err                error
	spinner            spinner.Model
	userPrompt         string
	aiResponse         string
	currentChatIndices *currentChatIndexes
	size               tea.WindowSizeMsg

	stack     []tea.Model
	errorList []error
}

var terminalWidth, terminalHeight, _ = ui.GetTerminalSize()

func initialChatModel(pc *service.PromptConfig) chatModel {
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

	vp := viewport.New(0, 0)

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

		errorList: []error{},
	}

	return modelStruct
}

func (m chatModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

var commandSelectionFn = CommandSelectionFactory()

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := AppStyle.GetFrameSize()
		msg.Width = msg.Width - h
		msg.Height = msg.Height - v
		m.size = msg

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.textarea.Height()

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlR:
			return reset(&m)

		case tea.KeyCtrlK:
			m.promptConfig.ChatMessages.DeleteMessage(m.currentChatIndices.user)
			m.promptConfig.ChatMessages.DeleteMessage(m.currentChatIndices.assistant)
			return reset(&m)

		case tea.KeyCtrlC:
			if m.err != nil {
				m.err = nil
				return m, nil
			}
			if len(m.stack) > 0 {
				return m, event.RemoveStack(m.stack[len(m.stack)-1])
			}
			if m.promptConfig.FindContextWithId(m.currentChatIndices.user) != nil {
				return closeContext(&m)
			}

		case tea.KeyCtrlD:
			return quit(&m)

		case tea.KeyShiftUp:
			return changeResponseUp(&m)

		case tea.KeyShiftDown:
			return changeResponseDown(&m)

		case tea.KeyCtrlP:
			if len(m.stack) == 0 {
				return addPagerToStack(&m)
			}

		case tea.KeyCtrlI:
			if len(m.stack) == 0 {
				return m, event.AddStack(config.NewConfigModel(m.promptConfig))
			}

		case tea.KeyCtrlL:
			if len(m.stack) == 0 {
				return m, event.AddStack(system.NewSystemModel(m.promptConfig))
			}

		case tea.KeyCtrlE:
			if len(m.stack) == 0 {
				return m, event.AddStack(list.NewFancyListModel("Errors", lo.Map(m.errorList, func(e error, _ int) list.Item {
					return list.Item{
						ItemId:          e.Error(),
						ItemTitle:       e.Error(),
						ItemDescription: e.Error(),
					}
				}), nil))
			}

		case tea.KeyEnter:
			if m.err != nil {
				m.err = nil
				return m, nil
			}

			if len(m.stack) == 0 {
				if e, c := callFunction(&m); e != nil {
					return e, c
				}

				return promptSend(&m)
			}
		}

	case event.UpdateContentEvent:
		cmds = append(cmds, event.UpdateAiResponse(m.aiResponse), event.UpdateUserPrompt(m.userPrompt))

	case event.RemoveStackEvent:
		if msg.Stack != nil {
			_, index, ok := lo.FindIndexOf[tea.Model](m.stack, func(item tea.Model) bool {
				return reflect.TypeOf(item) == reflect.TypeOf(msg.Stack)
			})
			if !ok || index == len(m.stack) {
				return m, nil
			}
		}

		// TODO: find better solutions, direct comparison provokes panic
		m.stack = m.stack[:len(m.stack)-1]
		return m, m.resize

	case event.AddStackEvent:
		m.stack = append(m.stack, msg.Stack)
		return m, tea.Sequence(m.stack[len(m.stack)-1].Init(), m.resize, event.UpdateContent)

	case service.ChatMessage:
		if msg.Id == m.currentChatIndices.user {
			m.userPrompt = msg.Content
		}

		if msg.Id == m.currentChatIndices.assistant {
			m.aiResponse = msg.Content
		}

		cmds = append(cmds, waitForUpdate(m.promptConfig.UpdateChan), event.UpdateContent)

	case error:
		m.err = msg
		m.errorList = append(m.errorList, msg)
		return m, nil

	}

	if len(m.stack) > 0 {
		var cmd tea.Cmd
		m.stack[len(m.stack)-1], cmd = m.stack[len(m.stack)-1].Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)

		cmds = append(cmds, tiCmd, vpCmd)
	}

	if m.userPrompt != "" {
		aiRes := m.aiResponse
		if m.promptConfig.MdMode {
			aiRes = string(markdown.Render(m.aiResponse, m.size.Width, 0))
		}
		userPrompt := m.userPrompt
		if m.currentChatIndices.user >= 0 {
			_, index, _ := lo.FindIndexOf[service.ChatMessage](m.promptConfig.ChatMessages.Messages, func(c service.ChatMessage) bool { return c.Id == m.currentChatIndices.user })
			userPrompt = fmt.Sprintf("[%d] %s", index+1, userPrompt)
		}
		m.viewport.SetContent(fmt.Sprintf("%s\n%s", userStyle.Render(userPrompt), aiRes))
	}

	return m, tea.Batch(cmds...)
}

func (m chatModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %s", m.err)
	}

	if len(m.stack) > 0 {
		return m.stack[len(m.stack)-1].View()
	}
	return AppStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		m.textarea.View(),
	))
}

func waitForUpdate(updateChan chan service.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-updateChan
	}
}
