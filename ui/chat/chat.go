package chat

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"os"
	"reflect"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/chains"
	"golang.org/x/term"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/audio"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/helper"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/MohammadBnei/go-ai-cli/ui/transition"
)

var (
	AppStyle = lipgloss.NewStyle().Margin(0, 2, 0)
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

var ChatProgram *tea.Program

type currentChatMessages struct {
	user, assistant *service.ChatMessage
}
type chatModel struct {
	viewport            viewport.Model
	textarea            textarea.Model
	promptConfig        *service.PromptConfig
	err                 error
	spinner             spinner.Model
	userPrompt          string
	aiResponse          string
	currentChatMessages *currentChatMessages
	size                tea.WindowSizeMsg

	history *helper.HistoryManager

	stack     []tea.Model
	errorList []error

	mdRenderer *glamour.TermRenderer
	keys       *listKeyMap
	help       help.Model

	transition      bool
	transitionModel *transition.Model

	loading bool

	audioPlayer *audio.AudioPlayerModel

	chain     chains.Chain
	chainName string
}

func NewChatModel(pc *service.PromptConfig) (*chatModel, error) {
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
	ta.Cursor.Blink = false

	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.ShowLineNumbers = false

	vp := viewport.New(w, 0)
	vp.MouseWheelDelta = 1

	mdRenderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		return nil, err
	}

	audioPlayer, err := audio.NewPlayerModel(pc)
	if err != nil {
		return nil, err
	}

	modelStruct := chatModel{
		textarea:     ta,
		promptConfig: pc,
		viewport:     vp,
		err:          err,
		spinner:      spinner.New(),
		aiResponse:   "",
		userPrompt:   "",
		currentChatMessages: &currentChatMessages{
			user:      nil,
			assistant: nil,
		},

		keys: newListKeyMap(),

		errorList: []error{},
		history:   helper.NewHistoryManager(),

		mdRenderer: mdRenderer,

		help: help.New(),

		transitionModel: transition.NewTransitionModel(""),

		audioPlayer: audioPlayer,
	}

	return &modelStruct, nil
}

func (m chatModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Go AI cli")
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.LoadingTitle()
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size.Height = msg.Height
		m.size.Width = msg.Width
		m.Resize()

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
		case tea.KeyCtrlW:
			if len(m.stack) == 0 {
				switch {
				case m.currentChatMessages.user != nil && m.currentChatMessages.user.Role == service.RoleUser:
					m.userPrompt = m.currentChatMessages.user.Content
				case m.currentChatMessages.assistant != nil:
					m.aiResponse = m.currentChatMessages.assistant.Content
				}
			}
			cmds = append(cmds, tea.Sequence(event.Transition("clear"), event.UpdateChatContent("", ""), event.Transition("")))

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

			if len(m.stack) == 0 && m.textarea.Value() != "" {
				m.history.Add(m.textarea.Value())

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
			if viper.GetBool(config.UI_MARKDOWN_MODE) && m.userPrompt != "Infos" {
				str, err := m.mdRenderer.Render(aiRes)
				if err != nil {
					return m, event.Error(err)
				}
				aiRes = str
			}
			if !viper.GetBool(config.UI_MARKDOWN_MODE) {
				aiRes = wordwrap.String(aiRes, m.viewport.Width)
			}

			m.viewport.SetContent(aiRes)
			m.Resize()
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
			switch {
			case m.currentChatMessages.user != nil:
				m.currentChatMessages.user = m.promptConfig.ChatMessages.FindById(m.currentChatMessages.user.Id.Int64())
			case m.currentChatMessages.assistant != nil:
				m.currentChatMessages.assistant = m.promptConfig.ChatMessages.FindById(m.currentChatMessages.assistant.Id.Int64())
			}
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
		if m.currentChatMessages.user != nil && msg.Id == m.currentChatMessages.user.Id {
			m.userPrompt = msg.Content
		}

		if m.currentChatMessages.assistant != nil && msg.Id == m.currentChatMessages.assistant.Id {
			m.aiResponse = msg.Content
		}

		cmds = append(cmds, tea.Sequence(event.UpdateChatContent(m.userPrompt, m.aiResponse), waitForUpdate(m.promptConfig.UpdateChan)))

	case event.StartSpinnerEvent:
		return m, nil

	case event.FileSelectionEvent:
		if len(m.stack) == 0 {
			for _, item := range msg.Files {
				_, err := m.promptConfig.ChatMessages.AddMessageFromFile(item)
				if err != nil {
					return m, event.Error(err)
				}
			}
		}

	case event.AgentSelectionEvent:
		m.chain = msg.Executor
		m.chainName = msg.Name

	case audio.StartPlayingEvent:
		m.audioPlayer = msg.PlayerModel
		if apModel, ok := lo.Find(m.stack, func(item tea.Model) bool {
			_, ok := item.(audio.AudioPlayerModel)
			return ok
		}); ok {
			apModel = msg.PlayerModel
			_ = apModel
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

		m.promptConfig.UserPrompt = m.textarea.Value()

		cmds = append(cmds, tiCmd, vpCmd)
	}

	if m.currentChatMessages.assistant == nil &&
		m.currentChatMessages.user == nil &&
		m.userPrompt == "" &&
		m.aiResponse == "" {
		m.Intro()
	}

	m.LoadingTitle()

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

func (m chatModel) LoadingTitle() {
	m.loading = len(m.promptConfig.Contexts) != 0
	if m.loading {
		style.TitleStyle = style.TitleStyle.Background(style.LoadingBackgroundColor)
	} else {
		style.TitleStyle = style.TitleStyle.Background(style.NormalBackgroundColor)
	}
}

func (m chatModel) GetTitleView() string {
	userPrompt := m.userPrompt
	if m.currentChatMessages.user != nil && m.currentChatMessages.user.Order != 0 {
		userPrompt = fmt.Sprintf("[%d] %s", m.currentChatMessages.user.Order, userPrompt)
	}
	if userPrompt == "" {
		userPrompt = "Chat"
	}
	if m.chain != nil {
		userPrompt = fmt.Sprintf("%s | %s", m.chainName, userPrompt)
	}
	return style.TitleStyle.Render(wordwrap.String(userPrompt, m.size.Width-8))
}

func waitForUpdate(updateChan chan service.ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-updateChan
	}
}

func (m *chatModel) Intro() {
	m.viewport.SetContent(getInfoContent(*m))
}

func (m *chatModel) Resize() {
	w, h := AppStyle.GetFrameSize()
	style.TitleStyle.MaxWidth(m.size.Width - w)

	m.viewport.Width = m.size.Width - w
	m.viewport.Height = m.size.Height - lipgloss.Height(m.GetTitleView()) - m.textarea.Height() - lipgloss.Height(m.help.View(m.keys)) - h

	m.mdRenderer, _ = glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(m.viewport.Width-2))
}
