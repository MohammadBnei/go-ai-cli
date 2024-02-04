package prompt

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui/config"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/MohammadBnei/go-openai-cli/ui/helper"
	"github.com/MohammadBnei/go-openai-cli/ui/list"
	"github.com/MohammadBnei/go-openai-cli/ui/style"
	"github.com/MohammadBnei/go-openai-cli/ui/system"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	AppStyle = lipgloss.NewStyle().Margin(1, 2, 0)
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

	history *helper.HistoryManager

	stack     []tea.Model
	errorList []error
}

func initialChatModel(pc *service.PromptConfig) chatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 0

	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	ta.SetWidth(w)
	ta.SetHeight(2)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(0, 0)

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
		history:   helper.NewHistoryManager(),
	}

	reset(&modelStruct)

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

	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg

		style.TitleStyle.Width(msg.Width)

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

		case tea.KeyCtrlH:
			if len(m.stack) == 0 && (m.textarea.Value() == "" || m.textarea.Value() == m.history.Current()) {
				m.textarea.SetValue(m.history.Previous())
			}

		case tea.KeyCtrlJ:
			if len(m.stack) == 0 && (m.textarea.Value() == "" || m.textarea.Value() == m.history.Current()) {
				m.textarea.SetValue(m.history.Next())
			}

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
				m.history.Add(m.textarea.Value())
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
		if viper.GetBool("md") && m.userPrompt != "Infos" {
			str, err := glamour.Render(aiRes, "dark")
			if err != nil {
				cmds = append(cmds, event.Error(err))
			}
			aiRes = str
		}
		userPrompt := m.userPrompt
		if m.currentChatIndices.user >= 0 {
			_, index, _ := lo.FindIndexOf[service.ChatMessage](m.promptConfig.ChatMessages.Messages, func(c service.ChatMessage) bool { return c.Id == m.currentChatIndices.user })
			userPrompt = fmt.Sprintf("[%d] %s", index+1, userPrompt)
		}
		m.viewport.SetContent(fmt.Sprintf("%s\n%s", style.TitleStyle.Render(userPrompt), aiRes))
	}

	return m, tea.Batch(cmds...)
}

func (m chatModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %s", style.StatusMessageStyle(m.err.Error()))
	}

	if len(m.stack) > 0 {
		return AppStyle.Render(m.stack[len(m.stack)-1].View())
	}
	return AppStyle.Render(fmt.Sprintf("%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
	))
}

func waitForUpdate(updateChan chan service.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-updateChan
	}
}
