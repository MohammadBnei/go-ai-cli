package list

import (
	"github.com/MohammadBnei/go-openai-cli/ui/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newItemDelegate(keys *delegateKeyMap, delegateFn *DelegateFunctions) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string
		var id string

		help := []key.Binding{}

		if delegateFn.AddFn != nil {
			help = append(help, keys.add)
		}

		itemSelected := true
		if i, ok := m.SelectedItem().(Item); ok {
			title = i.Title()
			id = i.Id()

			switch {
			case delegateFn.ChooseFn != nil:
				help = append(help, keys.choose)
				fallthrough
			case delegateFn.EditFn != nil:
				help = append(help, keys.edit)
				fallthrough
			case delegateFn.RemoveFn != nil:
				help = append(help, keys.remove)
			}
		} else {
			itemSelected = false
		}

		d.ShortHelpFunc = func() []key.Binding {
			return help
		}

		d.FullHelpFunc = func() [][]key.Binding {
			return [][]key.Binding{help}
		}

		cmds := []tea.Cmd{}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				if !itemSelected || delegateFn.ChooseFn == nil {
					return nil
				}
				cmds = append(cmds, delegateFn.ChooseFn(id))
				cmds = append(cmds, m.NewStatusMessage(style.StatusMessageStyle("Chose "+title)))
				return tea.Batch(cmds...)

			case key.Matches(msg, keys.remove):
				if !itemSelected || delegateFn.RemoveFn == nil {
					return nil
				}
				cmds = append(cmds, delegateFn.RemoveFn(id))
				index := m.Index()
				m.RemoveItem(index)
				if len(m.Items()) == 0 {
					keys.remove.SetEnabled(false)
				}
				cmds = append(cmds, delegateFn.RemoveFn(id), m.NewStatusMessage(style.StatusMessageStyle("Deleted "+title)))
				return tea.Batch(cmds...)

			case key.Matches(msg, keys.edit):
				if !itemSelected || delegateFn.EditFn == nil {
					return nil
				}
				cmds = append(cmds, delegateFn.EditFn(id), m.NewStatusMessage(style.StatusMessageStyle("Edited "+title)))
				return tea.Batch(cmds...)

			case key.Matches(msg, keys.add):
				if delegateFn.AddFn == nil {
					return nil
				}
				cmds = append(cmds, delegateFn.AddFn(id), m.NewStatusMessage(style.StatusMessageStyle("Created "+title)))
				return tea.Batch(cmds...)
			}
		}

		return nil
	}

	help := []key.Binding{}
	if delegateFn.ChooseFn != nil {
		help = append(help, keys.choose)
	}
	if delegateFn.EditFn != nil {
		help = append(help, keys.edit)
	}
	if delegateFn.AddFn != nil {
		help = append(help, keys.add)
	}
	if delegateFn.RemoveFn != nil {
		help = append(help, keys.remove)
	}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
	edit   key.Binding
	add    key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
		d.edit,
		d.add,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
			d.edit,
			d.add,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "create"),
		),
		remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}
