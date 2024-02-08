package file

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	selectFile key.Binding
	submit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.selectFile, k.submit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.selectFile, k.submit},
	}
}

func newKeyMap() *keyMap {
	return &keyMap{
		selectFile: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add file"),
		),
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
	}
}
