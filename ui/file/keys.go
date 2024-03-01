package file

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	submit    key.Binding
	changeCwd key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.submit, k.changeCwd}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.submit, k.changeCwd},
	}
}

func newKeyMap() *keyMap {
	return &keyMap{
		submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "submit"),
		),
		changeCwd: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "change cwd"),
		),
	}
}
