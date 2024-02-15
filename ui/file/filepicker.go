package file

import (
	"fmt"
	"os"

	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/samber/lo"
)

type model struct {
	filepicker    filepicker.Model
	multiMode     bool
	selectedFiles []string
	keys          *keyMap
	help          help.Model
	title         string
	width         int
}

// New creates a new instance of the UI.
func NewFilePicker(multipleMode bool) model {
	startDir, _ := os.Getwd()
	fp := filepicker.New()
	fp.CurrentDirectory = startDir
	fp.ShowHidden = true
	fp.AutoHeight = true

	return model{
		filepicker:    fp,
		multiMode:     multipleMode,
		selectedFiles: []string{},
		keys:          newKeyMap(),
		help:          help.New(),
		title:         "File Picker",
	}
}

// Init intializes the UI.
func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update handles all UI interactions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.filepicker, cmd = m.filepicker.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.filepicker.Height = msg.Height - lipgloss.Height(m.help.View(m.keys)) - lipgloss.Height(m.GetTitleView())
		m.width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.submit):
			if len(m.selectedFiles) != 0 {
				return m, tea.Sequence(event.RemoveStack(m), event.FileSelection(m.selectedFiles, m.multiMode))
			}
		}
	}

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		if _, ok := lo.Find(m.selectedFiles, func(item string) bool {
			return item == path
		}); ok {
			m.selectedFiles = lo.Filter(m.selectedFiles, func(item string, _ int) bool {
				return item != path
			})
			return m, nil
		}
		m.selectedFiles = append(m.selectedFiles, path)
	}

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m model) View() string {
	if len(m.selectedFiles) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			fmt.Sprintf("%s\n%s\n%s", m.GetTitleView(), m.filepicker.View(), m.help.View(m.keys)),
			wordwrap.String(lo.Reduce(m.selectedFiles,
				func(agg string, item string, i int) string { return fmt.Sprintf("%s\n[%d] %s", agg, i, item) }, ""), (m.width/2)-5),
		)
	}
	return fmt.Sprintf("%s\n%s\n%s", m.GetTitleView(), m.filepicker.View(), m.help.View(m.keys))
}

func (m model) GetTitleView() string {
	numberOfItems := ""
	if len(m.selectedFiles) > 0 {
		numberOfItems = fmt.Sprintf(" (%d)", len(m.selectedFiles))
	}
	return style.TitleStyle.Render(m.title + numberOfItems)
}
