package audio

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	play      key.Binding
	speedUp   key.Binding
	speedDown key.Binding
	back      key.Binding
	forward   key.Binding

	toggleSelect key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.play, k.speedUp, k.speedDown, k.back, k.forward}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.play, k.speedUp, k.speedDown}}
}

func newKeyMap() *keyMap {
	return &keyMap{
		play: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "play/pause"),
		),
		speedUp: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "accelerate"),
		),
		speedDown: key.NewBinding(
			key.WithKeys("ctrl+z"),
			key.WithHelp("ctrl+z", "slow down"),
		),
		back: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "-5sec"),
		),
		forward: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "+5sec"),
		),

		toggleSelect: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "select mode"),
		),
	}
}
