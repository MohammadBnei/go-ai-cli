package prompt

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"moul.io/banner"
)

type ChatChan struct {
	UpdateChan chan service.ChatMessage
	UserPrompt chan string
	Done       chan bool
}

func Chat(chatChannels *ChatChan, pc *command.PromptConfig) {
	p := tea.NewProgram(initialChatModel(chatChannels, pc),
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion())

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
	promptConfig       *command.PromptConfig
	err                error
	spinner            spinner.Model
	userPrompt         string
	userStyle          lipgloss.Style
	assistantStyle     lipgloss.Style
	aiResponse         string
	currentChatIndices *currentChatIndexes
	size               tea.WindowSizeMsg

	stack []tea.Model
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
	return tea.EnterAltScreen
}

var commandSelectionFn = CommandSelectionFactory()

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	cmds := []tea.Cmd{}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	cmds = append(cmds, tiCmd, vpCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEsc:
			if len(m.stack) > 0 {
				m.stack = m.stack[:len(m.stack)-1]
				return m, nil
			}

		case tea.KeyCtrlR:
			return reset(&m)

		case tea.KeyCtrlC:
			return closeContext(&m)

		case tea.KeyCtrlD:
			return quit(&m)

		case tea.KeyShiftUp:
			return changeResponseUp(&m)

		case tea.KeyShiftDown:
			return changeResponseDown(&m)

		case tea.KeyCtrlP:
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

				cmds = append(cmds, m.stack[len(m.stack)-1].Init(), cmd)
			} else {
				m.stack = lo.Slice[tea.Model](m.stack, index-1, index)
			}

		case tea.KeyEnter:
			if m.err != nil {
				m.err = nil
				return m, nil
			}

			if e, c := callFunction(&m); e != nil {
				return e, c
			}

			return promptSend(&m)
		}

	case service.ChatMessage:
		if msg.Id == m.currentChatIndices.user {
			m.userPrompt = msg.Content
		}

		if msg.Id == m.currentChatIndices.assistant {
			m.aiResponse = msg.Content
		}

		return m, tea.Batch(waitForUpdate(m.promptConfig.UpdateChan), func() tea.Msg {
			return pagerContentUpdate(m.aiResponse)
		})

	// We handle errors just like any other message

	case errMsg:
		m.err = msg
		return m, nil

	}

	if len(m.stack) > 0 {
		var cmd tea.Cmd
		m.stack[len(m.stack)-1], cmd = m.stack[len(m.stack)-1].Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.userPrompt != "" {
		aiRes := m.aiResponse
		if m.promptConfig.MdMode {
			aiRes = string(markdown.Render(m.aiResponse, terminalWidth, 3))
		}
		m.viewport.SetContent(fmt.Sprintf("%s\n%s", m.userStyle.Render(m.userPrompt), aiRes))
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
