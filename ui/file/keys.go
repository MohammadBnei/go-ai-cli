package file

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	submit       key.Binding
	changeCwd    key.Binding
	changeFocus  key.Binding
	toggleHidden key.Binding
	addDir       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.submit, k.changeFocus, k.toggleHidden, k.addDir}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.submit, k.toggleHidden, k.changeFocus, k.changeCwd, k.addDir},
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
		toggleHidden: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "toggle hidden"),
		),
		changeFocus: key.NewBinding(
			key.WithKeys("tab", "ctrl+i"),
			key.WithHelp("tab/ctrl+i", "change focus"),
		),
		addDir: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "add dir"),
		),
	}
}
