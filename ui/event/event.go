package event

import tea "github.com/charmbracelet/bubbletea"

type AddStackEvent struct {
	Stack tea.Model
}

func AddStack(stack tea.Model) tea.Cmd {
	return func() tea.Msg {
		return AddStackEvent{Stack: stack}
	}
}

type RemoveStackEvent struct {
	Stack tea.Model
}

func RemoveStack(stack tea.Model) tea.Cmd {
	return func() tea.Msg {
		return RemoveStackEvent{Stack: stack}
	}
}

func Error(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

type UpdateContentEvent struct {
	Content string
}

type UpdateAiResponseEvent UpdateContentEvent
type UpdateUserPromptEvent UpdateContentEvent

func UpdateContent() tea.Msg {
	return UpdateContentEvent{}
}

func UpdateAiResponse(content string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAiResponseEvent{Content: content}
	}
}
func UpdateUserPrompt(content string) tea.Cmd {
	return func() tea.Msg {
		return UpdateUserPromptEvent{Content: content}
	}
}
