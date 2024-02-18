package audio

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	play      key.Binding
	speedUp   key.Binding
	speedDown key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.play, k.speedUp, k.speedDown}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.play, k.speedUp, k.speedDown}}
}

func newKeyMap() *keyMap {
	return &keyMap{
		play: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "add file"),
		),
		speedUp: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "submit"),
		),
		speedDown: key.NewBinding(
			key.WithKeys("ctrl+z"),
			key.WithHelp("ctrl+z", "change cwd"),
		),
	}
}
