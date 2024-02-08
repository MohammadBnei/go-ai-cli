package style

import "github.com/charmbracelet/lipgloss"


var (
	LoadingBackgroundColor = lipgloss.Color("#77777")
	NormalBackgroundColor = lipgloss.Color("#25A065")

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(NormalBackgroundColor).
			Padding(0, 4).
			MarginBottom(1).
			Bold(true)

	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)
