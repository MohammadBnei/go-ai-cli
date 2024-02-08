package chat

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/MohammadBnei/go-ai-cli/command"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/file"
	"github.com/MohammadBnei/go-ai-cli/ui/helper"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/MohammadBnei/go-ai-cli/ui/transition"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"moul.io/banner"
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

	mdRenderer *glamour.TermRenderer
	keys       *listKeyMap
	help       help.Model

	transition      bool
	transitionModel *transition.Model
}

func initialChatModel(pc *service.PromptConfig) chatModel {
	var err error

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

	smallTitleStyle := style.TitleStyle.Copy().Margin(0).Padding(0, 2)

	vp := viewport.New(w, 0)
	vp.SetContent(banner.Inline("go ai cli") + "\n" +
		banner.Inline("bnei") + "\n\n" +
		"Api : " + smallTitleStyle.Render(viper.GetString("API_TYPE")) + "\n" +
		"Model : " + smallTitleStyle.Render(viper.GetString("model")) + "\n",
	)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	mdRenderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(80))

	modelStruct := chatModel{
		textarea:     ta,
		promptConfig: pc,
		viewport:     vp,
		err:          err,
		spinner:      spinner.New(),
		aiResponse:   "",
		userPrompt:   "",
		currentChatIndices: &currentChatIndexes{
			user:      -1,
			assistant: -1,
		},

		keys: newListKeyMap(),

		errorList: []error{},
		history:   helper.NewHistoryManager(),

		mdRenderer: mdRenderer,

		help: help.New(),

		transitionModel: transition.NewTransitionModel(""),
	}

	modelStruct.LoadLastChat()

	return modelStruct
}

func (m chatModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Go AI cli")
}

var commandSelectionFn = CommandSelectionFactory()

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	// Prevent the cursor of the textarea from blinking
	// Prevents screen blinking issues
	case cursor.BlinkMsg:
		return m, nil
	case tea.WindowSizeMsg:
		m.size = msg

		m.mdRenderer, _ = glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(msg.Width))

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - lipgloss.Height(m.GetTitleView()) - m.textarea.Height() - lipgloss.Height(m.help.View(m.keys))

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		var cmd tea.Cmd
		m, cmd = keyMapUpdate(msg, m)
		if cmd != nil {
			return m, cmd
		}

		switch msg.Type {

		case tea.KeyCtrlU:
			if len(m.stack) == 0 && (m.textarea.Value() == "" || m.textarea.Value() == m.history.Current()) {
				m.textarea.SetValue(m.history.Previous())
			}

		case tea.KeyCtrlJ:
			if len(m.stack) == 0 && (m.textarea.Value() == "" || m.textarea.Value() == m.history.Current()) {
				m.textarea.SetValue(m.history.Next())
			}

		case tea.KeyCtrlF:
			if len(m.stack) == 0 {
				return m, event.AddStack(file.NewFilePicker(true), "Loading Filepicker...")
			}

		case tea.KeyCtrlE:
			if len(m.stack) == 0 {
				return m, event.AddStack(list.NewFancyListModel("Errors", lo.Map(m.errorList, func(e error, index int) list.Item {
					return list.Item{
						ItemId:          fmt.Sprintf("%d", index),
						ItemTitle:       e.Error(),
						ItemDescription: "",
					}
				}), nil), "Loading Errors...")
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

	case event.ClearScreenEvent:
		m.viewport.SetContent("")
		return m, nil

	case event.SetChatTextviewEvent:
		m.textarea.SetValue(msg.Content)

	case event.UpdateChatContentEvent:
		if msg.UserPrompt != "" {
			m.userPrompt = msg.UserPrompt
		}

		if msg.Content != "" {
			m.aiResponse = msg.Content
			cmds = append(cmds, event.UpdateAiResponse(m.aiResponse))
		}
		if m.userPrompt != "" {
			aiRes := m.aiResponse
			if viper.GetBool("md") && m.userPrompt != "Infos" {
				str, err := m.mdRenderer.Render(aiRes)
				if err != nil {
					return m, event.Error(err)
				}
				aiRes = str
			}

			if m.transition && m.transitionModel.Title == m.userPrompt {
				m.transition = false
			}

			m.viewport.SetContent(aiRes)
			m.viewport.Height = m.size.Height - lipgloss.Height(m.GetTitleView()) - m.textarea.Height() - lipgloss.Height(m.help.View(m.keys))

		}

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
		if len(m.stack) == 0 {
			return m, tea.Sequence(event.Transition("..."), m.Init(), event.Transition(""), m.resize)
		}
		return m, nil
	case event.AddStackEvent:
		m.stack = append(m.stack, msg.Stack)
		return m, tea.Sequence(event.Transition(msg.Title), m.stack[len(m.stack)-1].Init(), event.Transition(""), m.resize, event.UpdateChatContent(m.userPrompt, m.aiResponse))

	case event.TransitionEvent:
		m.transition = msg.Title != ""
		m.transitionModel.Title = msg.Title

	case service.ChatMessage:
		if msg.Id == m.currentChatIndices.user {
			m.userPrompt = msg.Content
		}

		if msg.Id == m.currentChatIndices.assistant {
			m.aiResponse = msg.Content
		}

		cmds = append(cmds, tea.Sequence(event.UpdateChatContent(m.userPrompt, m.aiResponse), waitForUpdate(m.promptConfig.UpdateChan)))

	case event.StartSpinnerEvent:
		return m, nil

	case event.FileSelectionEvent:
		if len(m.stack) == 0 {
			for _, item := range msg.Files {
				_, err := m.promptConfig.ChatMessages.AddMessageFromFile(item.FileName())
				if err != nil {
					return m, event.Error(err)
				}
			}
		}

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

	return m, tea.Batch(cmds...)
}

func (m chatModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %s", style.StatusMessageStyle(m.err.Error()))
	}

	if m.transition {
		return AppStyle.Render(m.transitionModel.View())
	}

	if len(m.stack) > 0 {
		return AppStyle.Render(m.stack[len(m.stack)-1].View())
	}

	helpView := m.help.View(m.keys)
	return AppStyle.Render(fmt.Sprintf("%s\n%s\n%s\n%s",
		m.GetTitleView(),
		m.viewport.View(),
		m.textarea.View(),
		helpView,
	))
}

func (m chatModel) GetTitleView() string {
	userPrompt := m.userPrompt
	if m.currentChatIndices.user >= 0 {
		_, index, _ := lo.FindIndexOf[service.ChatMessage](m.promptConfig.ChatMessages.Messages, func(c service.ChatMessage) bool { return c.Id == m.currentChatIndices.user })
		userPrompt = fmt.Sprintf("[%d] %s", index+1, userPrompt)
	}
	if userPrompt == "" {
		userPrompt = "Chat"
	}
	return style.TitleStyle.Render(userPrompt)
}

func waitForUpdate(updateChan chan service.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-updateChan
	}
}

func CommandSelectionFactory() func(cmd string, pc *service.PromptConfig) error {
	commandMap := make(map[string]func(*service.PromptConfig) error)

	command.AddAllCommand(commandMap)
	keys := lo.Keys[string](commandMap)

	return func(cmd string, pc *service.PromptConfig) error {

		var err error

		switch {
		case cmd == "":
			commandMap["help"](pc)
		case cmd == "\\":
			selection, err2 := fuzzyfinder.Find(keys, func(i int) string {
				return keys[i]
			})
			if err2 != nil {
				return err2
			}

			err = commandMap[keys[selection]](pc)
		case strings.HasPrefix(cmd, "\\"):
			command, ok := commandMap[cmd[1:]]
			if !ok {
				return errors.New("command not found")
			}
			err = command(pc)
		}

		return err
	}
}

func (m *chatModel) LoadLastChat() {
	lastChatPath := filepath.Dir(viper.GetViper().ConfigFileUsed()) + "/last-chat.yml"
	if stat, err := os.Stat(lastChatPath); err == nil && !stat.IsDir() {
		m.err = m.promptConfig.ChatMessages.LoadFromFile(lastChatPath)
	}
}