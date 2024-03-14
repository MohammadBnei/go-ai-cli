package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"golang.org/x/tools/godoc/util"
	"golang.org/x/tools/godoc/vfs"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/MohammadBnei/go-ai-cli/ui/style"
)

type PickFileModel struct {
	filepicker   Model
	multiMode    bool
	keys         *keyMap
	help         help.Model
	title        string
	width        int
	selectedList *list.Model
	fileFocus    bool
}

// NewFilesPicker New creates a new instance of the UI.
func NewFilesPicker(allowedTypes []string) PickFileModel {
	startDir, _ := os.Getwd()
	fp := New()
	fp.CurrentDirectory = startDir
	fp.ShowHidden = true
	fp.AutoHeight = true
	if len(allowedTypes) > 0 {
		fp.AllowedTypes = allowedTypes
	}

	fileList := list.NewFancyListModel("Selected File", []list.Item{}, &list.DelegateFunctions{
		RemoveFn: func(s string) tea.Cmd { return nil },
	})

	return PickFileModel{
		filepicker:   fp,
		keys:         newKeyMap(),
		help:         help.New(),
		title:        "File Picker",
		selectedList: fileList,
		fileFocus:    true,
	}
}

// Init intializes the UI.
func (m PickFileModel) Init() tea.Cmd {
	return tea.Batch(m.filepicker.Init(), m.selectedList.Init())
}

// Update handles all UI interactions.
func (m PickFileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		newSize := tea.WindowSizeMsg{
			Width:  msg.Width / 2,
			Height: msg.Height - lipgloss.Height(m.GetTitleView()) - lipgloss.Height(m.help.View(m.keys)),
		}
		m.filepicker.Height = newSize.Height
		m.selectedList.List.SetHeight(newSize.Height)
		m.selectedList.List.SetWidth(msg.Width/2 - 1)
		m.width = msg.Width

		m.filepicker, cmd = m.filepicker.Update(newSize)
		cmds = append(cmds, cmd)
		// Update the selected list.
		l, cmd := m.selectedList.Update(newSize)
		if fL, ok := l.(*list.Model); ok {
			m.selectedList = fL
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.submit):
			if len(m.selectedList.Items()) != 0 {
				filePaths := lo.Map(m.selectedList.Items(), func(item list.Item, _ int) string {
					return item.ItemId
				})
				return m, tea.Sequence(event.RemoveStack(m), event.FileSelection(filePaths, m.multiMode))
			}
		case key.Matches(msg, m.keys.toggleHidden):
			m.filepicker.ShowHidden = !m.filepicker.ShowHidden
			return m, m.filepicker.readDir(m.filepicker.CurrentDirectory, m.filepicker.ShowHidden)
		case key.Matches(msg, m.keys.addDir):
			var addFileSequence []tea.Cmd
			if err := filepath.Walk(m.filepicker.CurrentDirectory, func(path string, info os.FileInfo, err error) error {
				// Skip hidden files and directories
				if strings.HasPrefix(info.Name(), ".") && !m.filepicker.ShowHidden {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if path == m.filepicker.CurrentDirectory || info.IsDir() {
					return nil
				}
				if err != nil {
					return err
				}
				addFileSequence = append(addFileSequence, addFile(path))
				return nil
			}); err != nil {
				return m, event.Error(err)
			}
			cmds = append(cmds, tea.Sequence(addFileSequence...))
		case key.Matches(msg, m.keys.changeFocus):
			m.fileFocus = !m.fileFocus
		}

	case addFileEvent:
		items := m.selectedList.Items()
		if _, ok := lo.Find(items, func(item list.Item) bool {
			return item.ItemId == msg.file
		}); ok {
			cmds = append(cmds, removeFile(msg.file))
		} else {
			f, err := os.ReadFile(msg.file)
			if err != nil {
				return m, event.Error(err)
			}
			if !util.IsTextFile(vfs.OS("/"), msg.file) {
				return m, event.Error(fmt.Errorf("file %s is not a text file", msg.file))
			}
			tokens, err := service.CountTokens(string(f))
			if err != nil {
				return m, event.Error(err)
			}
			m.selectedList.List.InsertItem(10000, list.Item{
				ItemId:          msg.file,
				ItemTitle:       msg.file,
				ItemDescription: fmt.Sprintf("%d tokens", tokens),
			})
		}

	case removeFileEvent:
		cmd := m.selectedList.RemoveItemById(msg.file)
		return m, cmd
	}

	if m.fileFocus {
		m.filepicker, cmd = m.filepicker.Update(msg)
		cmds = append(cmds, cmd)
		// Did the user select a file?
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			cmds = append(cmds, addFile(path))
		}
	} else {
		// Update the selected list.
		l, cmd := m.selectedList.Update(msg)
		if fL, ok := l.(list.Model); ok {
			m.selectedList = &fL
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m PickFileModel) View() string {
	selectedString := "> "
	m.selectedList.List.Title = strings.ReplaceAll(m.selectedList.List.Title, selectedString, "")
	if m.fileFocus {
		m.title = selectedString + m.title
	} else {
		m.selectedList.List.Title = selectedString + m.selectedList.List.Title
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		fmt.Sprintf("%s\n%s\n%s", m.GetTitleView(), m.filepicker.View(), m.help.View(m.keys)),
		m.selectedList.View(),
	)
}

func (m PickFileModel) GetTitleView() string {
	return style.TitleStyle.Render(m.title)
}
