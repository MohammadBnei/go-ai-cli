package quit

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.quit},
	}
}

func newKeyMap() *keyMap {
	return &keyMap{
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}
