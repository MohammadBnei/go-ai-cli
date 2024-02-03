package prompt

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"moul.io/banner"
)

type AncestorChan struct {
	UpdateChan chan service.ChatMessage
	UserPrompt chan string
	Done       chan bool
}

func Ancestor(ancestorChannels *AncestorChan, pc *service.PromptConfig) {
	p := tea.NewProgram(initialAncestorModel(ancestorChannels, pc), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type ancestorModel struct {
	viewport         viewport.Model
	promptConfig     *service.PromptConfig
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	assistStyle      lipgloss.Style
	err              error
	spinner          spinner.Model
	ancestorChannels *AncestorChan
}

func initialAncestorModel(ancestorChannels *AncestorChan, pc *service.PromptConfig) ancestorModel {
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

	modelStruct := ancestorModel{
		textarea:         ta,
		promptConfig:     pc,
		ancestorChannels: ancestorChannels,
		viewport:         vp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		assistStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		err:              nil,
		spinner:          spinner.New(),
	}

	return modelStruct
}

func (m ancestorModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		waitForUpdate(m.ancestorChannels.UpdateChan),
	)
}

func (m ancestorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			terminalWidth, terminalHeight, _ = ui.GetTerminalSize()
			m.viewport = viewport.New(terminalWidth, terminalHeight-3)
			return m, tea.EnterAltScreen

		case tea.KeyCtrlC:
			m.promptConfig.CloseLastContext()

		case tea.KeyCtrlD:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit

		case tea.KeyEnter:
			m.ancestorChannels.UserPrompt <- m.textarea.Value()
			<-m.ancestorChannels.Done
			m.textarea.Reset()
			m.viewport.GotoBottom()
			return m, nil
		}
	case service.ChatMessage:
		m.viewport.SetContent(PrintAncestorMessages(m.promptConfig.ChatMessages.Messages, viper.GetBool("md")))
		return m, waitForUpdate(m.ancestorChannels.UpdateChan)

	// We handle errors just like any other message

	case errMsg:
		m.err = msg
		return m, nil

	default:
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m ancestorModel) View() string {
	return fmt.Sprintf(
		"%s\n%s\n",
		m.viewport.View(),
		m.textarea.View(),
	)
}

func PrintAncestorMessages(messages []service.ChatMessage, markdownMode bool) (str string) {
	for _, msg := range messages {
		switch msg.Role {
		case service.RoleUser:
			str += fmt.Sprintf("🧠: %s\n", msg.Content)
		case service.RoleAssistant:
			content := msg.Content
			if markdownMode {
				content = string(markdown.Render(string(content), terminalWidth, 6))
			}
			str += fmt.Sprintf("%s\n\n", content)
		// case service.RoleSystem:
		// 	str += fmt.Sprintf("🤖: %s", lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(msg.Content))
		case service.RoleApp:
			str += fmt.Sprintf("🧌: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(msg.Content))
		}
	}

	return
}
