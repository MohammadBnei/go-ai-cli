package file

import tea "github.com/charmbracelet/bubbletea"

type toggleFocusEvent struct{}

func toggleFocus() tea.Msg {
	return toggleFocusEvent{}
}

type addFileEvent struct {
	file string
}

func addFile(filename string) tea.Cmd {
	return func() tea.Msg {
		return addFileEvent{
			file: filename,
		}
	}
}

type removeFileEvent struct {
	file string
}

func removeFile(filename string) tea.Cmd {
	return func() tea.Msg {
		return removeFileEvent{
			file: filename,
		}
	}
}
