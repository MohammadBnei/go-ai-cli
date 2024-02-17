package savechat

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
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

func NewSaveChatModel(promptConfig *service.PromptConfig) tea.Model {
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
			err := saveChat(*m.promptConfig, m.form.GetString("name"))
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

var filenamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_. -]*(\.[a-zA-Z]{1,4})?$`)

func constructForm() *huh.Form {
	tRue := true
	group := huh.NewGroup(
		huh.NewInput().Key("name").Title("Saved chat name (leave blank for auto-load)").Validate(func(s string) error {
			if s == "" {
				return nil
			}
			if !filenamePattern.MatchString(s) {
				return errors.New("the file name provided is not valid (alphanumerical character only)")
			}
			return nil
		}),
		huh.NewConfirm().Key("confirm").Title("Confirm").Value(&tRue),
	)
	return huh.NewForm(group)
}

func saveChat(pc service.PromptConfig, filename string) error {
	if filename == "" {
		filename = "last-chat"
	}
	chatMessages := pc.ChatMessages
	chatMessages.Id = filename
	if chatMessages.Description == "" {
		chatMessages.Description = "Saved at : " + time.Now().Format("2006-01-02 15:04:05")
	} else {
		chatMessages.Description += "\nUpdated at : " + time.Now().Format("2006-01-02 15:04:05")
	}

	err := chatMessages.SaveToFile(filepath.Dir(viper.GetViper().ConfigFileUsed()) + fmt.Sprintf("/%s.yml", filename))
	if err != nil {
		return err
	}

	return nil
}
