package list

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type Item struct {
	ItemId          string
	ItemTitle       string
	ItemDescription string
}

func (i Item) Title() string       { return i.ItemTitle }
func (i Item) Description() string { return i.ItemDescription }
func (i Item) Id() string          { return i.ItemId }
func (i Item) FilterValue() string { return i.ItemTitle }

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type DelegateFunctions struct {
	ChooseFn func(string) tea.Cmd
	RemoveFn func(string) tea.Cmd
	EditFn   func(string) tea.Cmd
	AddFn    func(string) tea.Cmd
}

type Model struct {
	List         list.Model
	Keys         *listKeyMap
	DelegateKeys *delegateKeyMap
	DelegateFn   *DelegateFunctions
}

func NewFancyListModel(title string, items []Item, delegateFn *DelegateFunctions) *Model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)
	if delegateFn == nil {
		delegateFn = &DelegateFunctions{}
	}

	delegate := newItemDelegate(delegateKeys, delegateFn)
	itemList := lo.Map[Item, list.Item](items, func(i Item, _ int) list.Item { return i })
	groceryList := list.New(itemList, delegate, 0, 0)
	groceryList.Title = title

	groceryList.KeyMap.Quit.Unbind()
	groceryList.KeyMap.ForceQuit.Unbind()
	groceryList.KeyMap.CursorUp.SetKeys("up")
	groceryList.KeyMap.CursorUp.SetHelp("↑", "up")
	groceryList.KeyMap.CursorDown.SetKeys("down")
	groceryList.KeyMap.CursorDown.SetHelp("↓", "down")

	groceryList.Styles.Title = titleStyle
	groceryList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			listKeys.insertItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return &Model{
		List:         groceryList,
		Keys:         listKeys,
		DelegateKeys: delegateKeys,
		DelegateFn:   delegateFn,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width, msg.Height)

	case Item:
		items := lo.Map[list.Item, Item](m.List.Items(), func(item list.Item, index int) Item {
			return item.(Item)
		})
		_, index, ok := lo.FindIndexOf[Item](items, func(item Item) bool {
			return item.Id() == msg.Id()
		})
		if ok {
			m.List.SetItem(index, msg)
		} else {
			m.List.InsertItem(0, msg)
		}

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.List.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.Keys.toggleSpinner):
			cmd := m.List.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.Keys.toggleTitleBar):
			v := !m.List.ShowTitle()
			m.List.SetShowTitle(v)
			m.List.SetShowFilter(v)
			m.List.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.Keys.toggleStatusBar):
			m.List.SetShowStatusBar(!m.List.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.Keys.togglePagination):
			m.List.SetShowPagination(!m.List.ShowPagination())
			return m, nil

		case key.Matches(msg, m.Keys.toggleHelpMenu):
			m.List.SetShowHelp(!m.List.ShowHelp())
			return m, nil

		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.List.Update(msg)
	m.List = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.List.View()
}

func (m *Model) WithChooseFn(chooseFn func(string) tea.Cmd) *Model {
	m.DelegateFn.ChooseFn = chooseFn
	return nil
}

func (m *Model) WithEditFn(editFn func(string) tea.Cmd) *Model {
	m.DelegateFn.EditFn = editFn
	return nil
}

func (m *Model) WithAddFn(addFn func(string) tea.Cmd) *Model {
	m.DelegateFn.AddFn = addFn
	return nil
}

func (m *Model) WithRemoveFn(removeFn func(string) tea.Cmd) *Model {
	m.DelegateFn.RemoveFn = removeFn
	return nil
}
