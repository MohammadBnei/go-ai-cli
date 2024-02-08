package quit

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type model struct {
	promptConfig *service.PromptConfig
	form         *huh.Form
	title        string
	keys         *keyMap
}

func NewQuitModel(promptConfig *service.PromptConfig) tea.Model {
	return model{promptConfig: promptConfig, form: constructForm(), title: "Quitting", keys: newKeyMap()}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.form.WithWidth(msg.Width)
		m.form.WithHeight(msg.Height - lipgloss.Height(m.GetTitleView()))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		}
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		if !m.form.GetBool("confirm") {
			return m, event.RemoveStack(m)
		}
		if m.form.GetBool("save") {
			err := saveChat(*m.promptConfig)
			if err != nil {
				fmt.Println(err)
			}
		}
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.GetTitleView(), m.form.View())
}

func (m model) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}

func constructForm() *huh.Form {
	tRue := true
	group := huh.NewGroup(
		huh.NewSelect[bool]().Key("save").Title("Save current chat ?").Options(huh.NewOptions[bool](true, false)...),
		huh.NewConfirm().Key("confirm").Title("Confirm").Value(&tRue),
	)
	return huh.NewForm(group)
}

func saveChat(pc service.PromptConfig) error {
	chatMessages := pc.ChatMessages
	chatMessages.Id = "last-chat"
	chatMessages.Description = "Saved at : " + time.Now().Format("2006-01-02 15:04:05")

	err := chatMessages.SaveToFile(filepath.Dir(viper.GetViper().ConfigFileUsed()) + "/last-chat.yml")
	if err != nil {
		return err
	}

	return nil
}
