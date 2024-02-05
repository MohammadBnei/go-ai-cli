package pager

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui/event"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type pagerContentUpdate string
type pagerTitleUpdate string

type PagerModel struct {
	content    string
	title      string
	ready      bool
	viewport   viewport.Model
	pc         *service.PromptConfig
	mdRenderer *glamour.TermRenderer
}

func NewPagerModel(title, content string, pc *service.PromptConfig) tea.Model {
	mdRenderer, _ := glamour.NewTermRenderer()
	pager := PagerModel{
		title:      title,
		content:    content,
		pc:         pc,
		mdRenderer: mdRenderer,
	}

	return pager
}

func (m PagerModel) Init() tea.Cmd {
	return nil
}

func (m PagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case event.UpdateUserPromptEvent:
		m.title = msg.Content

	case event.UpdateAiResponseEvent:
		m.content = msg.Content

	case *service.PromptConfig:
		m.pc = msg
		return m, nil

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.mdRenderer, _ = glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(msg.Width))

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.MouseWheelEnabled = true
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	content := m.content
	if viper.GetBool("md") {
		str, err := m.mdRenderer.Render(content)
		if err != nil {
			return m, event.Error(err)
		}
		content = str
	}
	m.viewport.SetContent(content)

	return m, tea.Batch(cmds...)
}

func (m PagerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m PagerModel) headerView() string {
	title := titleStyle.Render(m.title)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m PagerModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Pager(userPrompt, content string) {
	p := tea.NewProgram(
		PagerModel{title: userPrompt, content: content},
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
