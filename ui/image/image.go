package image

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/file"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

const (
	PROMPT = iota
	GENERATING
	DONE
)

type model struct {
	promptConfig *service.PromptConfig
	title        string

	editForm *huh.Form

	state int

	prompt string
	size   string
	path   string

	data *[]byte

	spinner    spinner.Model
	filepicker file.FilepickerModel
}

func NewImageModel(pc *service.PromptConfig) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Moon
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		promptConfig: pc,
		title:        "Image",
		filepicker:   file.NewFilePicker(false, []string{"jpg", "jpeg"}),
		editForm:     constructPromptForm(pc.UserPrompt),
		state:        PROMPT,
		spinner:      s,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	editForm, cmd := m.editForm.Update(msg)
	if form, ok := editForm.(*huh.Form); ok {
		m.editForm = form
	}
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.editForm.WithHeight(msg.Height - lipgloss.Height(m.GetTitleView()) - lipgloss.Height(m.spinner.View()))

	case generateErrorEvent:
		return m, tea.Sequence(event.RemoveStack(m), event.Error(msg.error))

	case generateImageEvent:
		m.state = GENERATING
		m.editForm = constructFilepickerForm()
		return m, tea.Sequence(m.editForm.Init(), m.spinner.Tick, func() tea.Msg {
			ctx, cancel := context.WithCancel(context.Background())
			m.promptConfig.AddContext(ctx, cancel)
			defer m.promptConfig.DeleteContext(ctx)
			data, err := api.GenerateImage(ctx, m.prompt, m.size)
			if err != nil {
				return generateError(err)()
			}
			return pickfile(&data)()
		})

	case pickfileEvent:
		m.data = msg.Data

		if m.path != "" {
			return m, writeFile(m.path)
		}

	case writeFileEvent:
		m.state = DONE
		return m, func() tea.Msg {
			var f *os.File
			var err error
			m.path = msg.Path
			m.path += ".png"
			if strings.Contains(msg.Path, "/") {
				f, err = os.Create(m.path)
			} else {
				wd, _ := os.Getwd()
				m.path = fmt.Sprintf("%s/%s", wd, m.path)
				f, err = os.Create(m.path)
			}
			if err != nil {
				return generateError(err)()
			}
			defer f.Close()

			if _, err = f.Write(*m.data); err != nil {
				return generateError(err)()
			}
			if err := OpenOnSystem(m.path); err != nil {
				return generateError(err)()
			}
			return event.RemoveStack(m)()
		}
	}

	if m.editForm.State == huh.StateCompleted {
		switch {
		case PROMPT == m.state:
			m.prompt = m.editForm.GetString("content")
			m.size = m.editForm.GetString("size")
			confirmed := m.editForm.GetBool("confirm")

			errs := m.editForm.Errors()
			if len(errs) > 0 {
				return m, tea.Sequence(event.Error(errors.Join(errs...)), m.editForm.PrevGroup())
			}

			if !confirmed {
				return m, event.RemoveStack(m)
			}
			return m, generateImage(m.prompt, m.size)

		case GENERATING == m.state:
			fallthrough
		case DONE == m.state && m.path == "":
			path := m.editForm.GetString("path")
			confirm := m.editForm.GetBool("confirmPath")

			if !confirm {
				m.editForm = constructFilepickerForm()
				return m, m.editForm.Init()
			}
			m.path = path
			if m.data != nil {
				return m, writeFile(path)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	spinner := ""
	if m.state == GENERATING {
		spinner = m.spinner.View()
	}

	return fmt.Sprintf("%s\n%s\n%s", m.GetTitleView(), spinner, m.editForm.View())
}

func (m model) Init() tea.Cmd {
	return m.editForm.Init()
}

func (m model) GetTitleView() string {
	return style.TitleStyle.Render(fmt.Sprintf("%s (%s)", m.title, viper.GetString(config.AI_OPENAI_IMAGE_MODEL)))
}

func constructPromptForm(content string) *huh.Form {
	tRue := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().CharLimit(0).Title("Content").Key("content").Lines(3).Value(&content),
			huh.NewSelect[string]().Title("Size").Key("size").Options(huh.NewOptions[string](
				openai.CreateImageSize1792x1024,
				openai.CreateImageSize1024x1792,
				openai.CreateImageSize1024x1024,
				openai.CreateImageSize512x512,
				openai.CreateImageSize256x256,
			)...),
			huh.NewConfirm().Affirmative("Generate").Negative("Cancel").Key("confirm").Value(&tRue),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return form
}
func constructFilepickerForm() *huh.Form {
	tRue := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().CharLimit(0).Title("Path").Key("path").Lines(1).Validate(func(s string) error {
				if f, err := os.Create(s); err != nil {
					return err
				} else {
					os.Remove(s)
					f.Close()
					return nil

				}
			}),
			huh.NewConfirm().Affirmative("Save").Negative("Cancel").Key("confirmPath").Value(&tRue),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return form
}

type generateErrorEvent struct {
	error
}

type generateImageEvent struct {
	Prompt string
	Size   string
}
type pickfileEvent struct {
	Data *[]byte
}

type writeFileEvent struct {
	Path string
}

func generateError(e error) tea.Cmd {
	return func() tea.Msg {
		return generateErrorEvent{
			error: e,
		}
	}
}

func generateImage(prompt, size string) tea.Cmd {
	return func() tea.Msg {
		return generateImageEvent{
			Prompt: prompt,
			Size:   size,
		}
	}
}
func pickfile(data *[]byte) tea.Cmd {
	return func() tea.Msg {
		return pickfileEvent{
			Data: data,
		}
	}
}

func writeFile(path string) tea.Cmd {
	return func() tea.Msg {
		return writeFileEvent{
			Path: path,
		}
	}
}

func OpenOnSystem(path string) error {
	// Use the `open` command on macOS and Linux, or `start` on Windows
	var cmd *exec.Cmd
	switch os := runtime.GOOS; os {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("start", path)
	default:
		log.Fatalf("unsupported operating system: %s", os)
	}

	// Execute the command
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
