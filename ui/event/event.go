package event

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmc/langchaingo/agents"
)

type AddStackEvent struct {
	Stack tea.Model
	Title string
}

func AddStack(stack tea.Model, title string) tea.Cmd {
	return func() tea.Msg {
		return AddStackEvent{Stack: stack, Title: title}
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

type UpdateChatContentEvent struct {
	UserPrompt string
	Content    string
}

type UpdateAiResponseEvent UpdateChatContentEvent
type UpdateUserPromptEvent UpdateChatContentEvent

type SetChatTextviewEvent struct {
	Content string
}

func UpdateChatContent(userPromt, content string) tea.Cmd {
	return func() tea.Msg {
		return UpdateChatContentEvent{
			UserPrompt: userPromt,
			Content:    content,
		}
	}
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

func SetChatTextview(content string) tea.Cmd {
	return func() tea.Msg {
		return SetChatTextviewEvent{Content: content}
	}
}

type ClearScreenEvent struct{}

func ClearScreen() tea.Msg {
	return ClearScreenEvent{}
}

type StartSpinnerEvent struct{}
type StopSpinnerEvent struct{}

func StartSpinner() tea.Msg {
	return StartSpinnerEvent{}
}
func StopSpinner() tea.Msg {
	return StopSpinnerEvent{}
}

type CancelEvent struct{}

func Cancel() tea.Msg {
	return CancelEvent{}
}

type TransitionEvent struct {
	Title string
}

func Transition(title string) tea.Cmd {

	return func() tea.Msg {
		return TransitionEvent{Title: title}
	}
}

type FileSelectionEvent struct {
	Files     []string
	MultiMode bool
}

func FileSelection(files []string, multiMode bool) tea.Cmd {
	return func() tea.Msg {
		return FileSelectionEvent{Files: files, MultiMode: multiMode}
	}
}

type AgentSelectionEvent struct {
	Executor *agents.Executor
	Name     string
}

func AgentSelection(executor *agents.Executor, name string) tea.Cmd {
	return func() tea.Msg {
		return AgentSelectionEvent{
			Executor: executor,
			Name:     name,
		}
	}
}
